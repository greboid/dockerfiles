// Package tailscale carries the live installer's Tailscale node over to
// the installed system.
//
// The installer's own tailscaled (the tailscale package's tailscaled.service,
// state in /var/lib/tailscale) is logged in once: interactively, with
// the login URL shown as a QR code on the console, or headlessly with
// an auth key. That node is usually also what you SSH into the
// installer over. After the install its state directory is copied to
// the target's /var/lib/tailscale, the path the installed system's
// tailscaled.service uses, so the installed machine boots as that same
// node.
//
// This moves the node rather than cloning it: the live installer runs
// from RAM and is gone once the machine reboots into the installed
// system, so the identity is never online twice. Do not boot the
// installed disk in a different machine while the installer is still
// running.
package tailscale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultStateDir is the installer tailscaled's state directory.
const DefaultStateDir = "/var/lib/tailscale"

// Options describe how the node is prepared before the install.
type Options struct {
	// Hostname renames the node; empty keeps its current name.
	Hostname string
	// SSH enables Tailscale SSH (never disabled: that could cut the
	// session the install is driven over).
	SSH bool
	// Tags are requested at login; the CLI cannot change them on a
	// node that is already logged in.
	Tags []string
	// AuthKeyFile logs in headlessly when the node needs a login;
	// without it the login URL is shown as a QR code.
	AuthKeyFile string
}

// Validate checks the options that do not depend on the daemon.
func (o Options) Validate() error {
	if o.AuthKeyFile != "" {
		b, err := os.ReadFile(o.AuthKeyFile)
		if err != nil {
			return fmt.Errorf("tailscale: reading auth key: %w", err)
		}
		if strings.TrimSpace(string(b)) == "" {
			return fmt.Errorf("tailscale: auth key file %s is empty", o.AuthKeyFile)
		}
	}
	for _, tag := range o.Tags {
		if !strings.HasPrefix(tag, "tag:") || len(tag) == len("tag:") {
			return fmt.Errorf("tailscale: tag %q must look like tag:<name>", tag)
		}
	}
	if _, err := exec.LookPath("tailscale"); err != nil {
		return errors.New("tailscale: the tailscale CLI is not in the installer")
	}
	return nil
}

// Status is the part of `tailscale status --json` the installer uses.
type Status struct {
	BackendState string
	Self         *struct {
		HostName     string
		DNSName      string
		TailscaleIPs []string
	}
}

// Running reports whether the node is logged in and connected.
func (s *Status) Running() bool { return s.BackendState == "Running" }

// Describe is a short "name (ip, ...)" label for the node.
func (s *Status) Describe() string {
	if s.Self == nil {
		return s.BackendState
	}
	name := strings.TrimSuffix(s.Self.DNSName, ".")
	if name == "" {
		name = s.Self.HostName
	}
	return fmt.Sprintf("%s (%s)", name, strings.Join(s.Self.TailscaleIPs, ", "))
}

// ParseStatus parses `tailscale status --json` output.
func ParseStatus(b []byte) (*Status, error) {
	var s Status
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parsing tailscale status: %w", err)
	}
	if s.BackendState == "" {
		return nil, errors.New("tailscale status has no BackendState; is tailscaled running?")
	}
	return &s, nil
}

// GetStatus asks the installer's tailscaled for its state.
func GetStatus(ctx context.Context) (*Status, error) {
	// status exits non-zero when not logged in but still prints JSON.
	out, err := exec.CommandContext(ctx, "tailscale", "status", "--json").Output()
	if len(out) == 0 && err != nil {
		return nil, fmt.Errorf("tailscale status: %w (is tailscaled running?)", err)
	}
	return ParseStatus(out)
}

// LoginArgs are the `tailscale up` arguments for logging the
// installer in. --reset keeps up from refusing over settings left from
// an earlier session on this live system.
func (o Options) LoginArgs() []string {
	args := []string{"up", "--reset"}
	if o.AuthKeyFile != "" {
		args = append(args, "--auth-key=file:"+o.AuthKeyFile, "--timeout=2m")
	} else {
		args = append(args, "--qr")
	}
	if o.Hostname != "" {
		args = append(args, "--hostname="+o.Hostname)
	}
	if o.SSH {
		args = append(args, "--ssh")
	}
	if len(o.Tags) > 0 {
		args = append(args, "--advertise-tags="+strings.Join(o.Tags, ","))
	}
	return args
}

// SetArgs are the `tailscale set` arguments for a node that is already
// logged in, or nil when nothing needs changing.
func (o Options) SetArgs() []string {
	var args []string
	if o.Hostname != "" {
		args = append(args, "--hostname="+o.Hostname)
	}
	if o.SSH {
		args = append(args, "--ssh")
	}
	if args == nil {
		return nil
	}
	return append([]string{"set"}, args...)
}

// Prepare makes sure the installer's node is logged in and configured,
// logging it in (QR code or auth key) when needed. interactive reports
// whether a person can scan a QR code.
func Prepare(ctx context.Context, o Options, interactive bool, stdout, stderr io.Writer) (*Status, error) {
	st, err := GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	if st.Running() {
		if len(o.Tags) > 0 {
			return nil, fmt.Errorf("tailscale: the installer is already logged in as %s and tags can only be requested at login; drop --tailscale-tags, or run `tailscale logout` first", st.Describe())
		}
		if args := o.SetArgs(); args != nil {
			if err := run(ctx, args, stdout, stderr); err != nil {
				return nil, err
			}
		}
	} else {
		if o.AuthKeyFile == "" && !interactive {
			return nil, errors.New("tailscale: the installer is not logged in; run on a terminal to scan the login QR code, or pass --tailscale-authkey-file")
		}
		if o.AuthKeyFile == "" {
			fmt.Fprintln(stdout, "== tailscale: scan the QR code (or open the URL) to log this machine in")
		}
		if err := run(ctx, o.LoginArgs(), stdout, stderr); err != nil {
			return nil, err
		}
	}
	if st, err = GetStatus(ctx); err != nil {
		return nil, err
	}
	if !st.Running() {
		return nil, fmt.Errorf("tailscale: node is %s after login, not Running", st.BackendState)
	}
	return st, nil
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "tailscale", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tailscale %s: %w", args[0], err)
	}
	return nil
}

// CopyState copies the installer's state directory (regular files and
// directories; sockets and the like are skipped) to dst, which is
// created with mode 0700 like tailscaled's own.
func CopyState(src, dst string) error {
	if _, err := os.Stat(filepath.Join(src, "tailscaled.state")); err != nil {
		return fmt.Errorf("tailscale: no node state to copy: %w", err)
	}
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dst, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return err
		}
		out := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(out, info.Mode().Perm())
		case info.Mode().IsRegular():
			return copyFile(path, out, info.Mode().Perm())
		}
		return nil
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp := dst + ".osimage-install"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}
