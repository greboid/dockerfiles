// Package target seeds per-machine state onto a freshly installed
// osimage root partition.
//
// The layout is bootc's composefs backend (bootc 1.17.3,
// composefs_consts.rs / bootc_composefs/state.rs):
//
//	state/os/default/var       the /var shared by every deployment
//	state/deploy/<id>/etc/     a pristine copy of the image's /etc
//	state/deploy/<id>/var  ->  ../../os/default/var
//
// The osimage initramfs (90osimage-state in osimage-kernel) seeds
// /var/lib/osimage/etc/{machine-id,hostname,shadow} only when they do
// not exist yet, then binds them read-only over /etc. Writing them here
// before the first boot therefore replaces the image defaults; the
// machine-id in particular is generated here, so the read-only bind
// never needs PID 1 to fill it in.
package target

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	sharedVar = "state/os/default/var"
	deployDir = "state/deploy"
	stateDir  = "lib/osimage/etc"
)

// Root is a mounted, freshly installed root partition.
type Root struct {
	Path string
}

// Var returns the deployment-shared /var directory.
func (r Root) Var() (string, error) {
	p := filepath.Join(r.Path, sharedVar)
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("%s not found on the installed root; not a bootc composefs install?", sharedVar)
	}
	return p, nil
}

// DeploymentEtc returns the /etc copy of the only deployment. A fresh
// install has exactly one; anything else means this is not the disk
// that was just installed.
func (r Root) DeploymentEtc() (string, error) {
	entries, err := os.ReadDir(filepath.Join(r.Path, deployDir))
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", deployDir, err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	if len(ids) != 1 {
		return "", fmt.Errorf("found %d deployments under %s, want exactly 1 on a fresh install", len(ids), deployDir)
	}
	return filepath.Join(r.Path, deployDir, ids[0], "etc"), nil
}

// ValidateHostname accepts an RFC 1123 host name of at most 64 bytes
// (HOST_NAME_MAX), optionally dotted.
func ValidateHostname(name string) error {
	if name == "" || len(name) > 64 {
		return fmt.Errorf("hostname %q must be 1-64 characters", name)
	}
	for label := range strings.SplitSeq(name, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("hostname %q has an invalid label %q", name, label)
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return fmt.Errorf("hostname %q contains %q (letters, digits and '-' only)", name, c)
			}
		}
	}
	return nil
}

func ensureStateDir(varDir string) (string, error) {
	dir := filepath.Join(varDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	// Same mode the initramfs gives it.
	return dir, os.Chmod(dir, 0o755)
}

// WriteHostname seeds /var/lib/osimage/etc/hostname.
func WriteHostname(varDir, name string) error {
	if err := ValidateHostname(name); err != nil {
		return err
	}
	dir, err := ensureStateDir(varDir)
	if err != nil {
		return err
	}
	// Owner left as-is (-1): the installer runs as root.
	return writeFile(filepath.Join(dir, "hostname"), []byte(name+"\n"), 0o644, -1, -1)
}

// NewMachineID returns a random machine ID in systemd's format: 32
// lowercase hex digits.
func NewMachineID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating machine-id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// WriteMachineID seeds /var/lib/osimage/etc/machine-id (mode 0444,
// like systemd's own) with id.
func WriteMachineID(varDir, id string) error {
	if len(id) != 32 || strings.Trim(id, "0123456789abcdef") != "" {
		return fmt.Errorf("invalid machine-id %q", id)
	}
	dir, err := ensureStateDir(varDir)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "machine-id"), []byte(id+"\n"), 0o444, -1, -1)
}

// ReadShadow returns the image's shadow file and its owner.
func ReadShadow(etcDir string) (content []byte, uid, gid int, err error) {
	p := filepath.Join(etcDir, "shadow")
	content, err = os.ReadFile(p)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("reading the image's shadow: %w", err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		return nil, 0, 0, err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, 0, 0, errors.New("cannot read shadow ownership")
	}
	return content, int(st.Uid), int(st.Gid), nil
}

// WriteShadow seeds /var/lib/osimage/etc/shadow (mode 0640, like the
// initramfs seed).
func WriteShadow(varDir string, content []byte, uid, gid int) error {
	dir, err := ensureStateDir(varDir)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "shadow"), content, 0o640, uid, gid)
}

// writeFile writes atomically: nothing binds these paths yet, so a
// rename is safe here (unlike on the running system). uid/gid -1
// keep the creating process's identity.
func writeFile(path string, data []byte, mode os.FileMode, uid, gid int) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if (uid >= 0 && uid != os.Getuid()) || (gid >= 0 && gid != os.Getgid()) {
		if err := tmp.Chown(uid, gid); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
