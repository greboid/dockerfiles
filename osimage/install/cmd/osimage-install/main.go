// Command osimage-install installs the published osimage image onto a
// disk from the live installer, optionally seeding the hostname, account
// passwords and the installer's Tailscale node before the first boot.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/greboid/osimage-install/internal/installer"
	"github.com/greboid/osimage-install/internal/sha512crypt"
	"github.com/greboid/osimage-install/internal/shadow"
	"github.com/greboid/osimage-install/internal/tailscale"
)

// seedHostname is the image default the initramfs seeds from
// (osimage-config); the live installer ships the same file.
const seedHostname = "/usr/lib/osimage/etc-seed/hostname"

const usage = `usage: osimage-install [options] [<disk>] [bootc install to-disk options]

Installs the published osimage image onto <disk> (whole disk, wiped)
with bootc: composefs backend, systemd-boot, systemd-repart layout
(1 GiB ESP + ext4 root grown to fill the disk).

On a terminal, <disk> may be omitted: the installer lists the disks
(usable ones numbered, refused ones with the reason) and asks which
to install to. It then asks for the hostname, root's password and
whether to move the installer's Tailscale node over, using the options
below as the defaults, then shows the plan and asks before erasing
the disk.

Image:
  --source-ref REF        image to install (env OSIMAGE_SOURCE_REF)
  --target-ref REF        upgrade origin recorded on the target
                          (env OSIMAGE_TARGET_REF)
                          both default to OSIMAGE_DEFAULT_REF, else
                          ` + installer.DefaultRef + `

Per-machine state, written to the target before its first boot:
  --hostname NAME         seed /etc/hostname
  --password USER         prompt for USER's password (repeatable)
  --password-file USER=PATH
                          read USER's password from the first line of
                          PATH (repeatable)
  --password-hash USER=HASH
                          set USER's crypt(3) hash as-is (repeatable)

Tailscale: the installer's own node moves to the installed system, so
it is logged in once and the installed machine boots as that node:
  --tailscale             carry the installer's node over; logs it in
                          first if needed, showing the login URL as a
                          QR code on the terminal
  --tailscale-authkey-file PATH
                          log in with this auth key instead of the QR
                          code (implies --tailscale; only used when the
                          installer is not logged in yet)
  --tailscale-hostname NAME
                          node name (default: the hostname)
  --tailscale-ssh         enable Tailscale SSH on the node
  --tailscale-tags TAGS   comma-separated tags to request (tag:x,...);
                          only at login

Other:
  --non-interactive       ask nothing: use the options as given and do
                          not confirm before erasing the disk (always
                          the case when stdin is not a terminal)
  --dry-run               check everything, print the bootc command,
                          change nothing

Arguments after <disk> go to bootc install to-disk; options owned by
the image or the repart profile (--karg, --filesystem, ...) are
refused. REGISTRY_AUTH_FILE is honoured for private registries.
`

// pairs collects repeatable flags.
type pairs []string

func (p *pairs) String() string     { return strings.Join(*p, ",") }
func (p *pairs) Set(v string) error { *p = append(*p, v); return nil }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "osimage-install: %v\n", err)
		stop()
		os.Exit(1)
	}
}

// flags are the parsed command line.
type flags struct {
	hostname       string
	tailscale      bool
	tsKey          string
	tsHostname     string
	tsSSH          bool
	tsTags         string
	nonInteractive bool
	prompt         pairs
}

func run(ctx context.Context, args []string, stdin *os.File, stdout, stderr io.Writer) error {
	defRef := envOr("OSIMAGE_DEFAULT_REF", installer.DefaultRef)

	var f flags
	var pwFiles, pwHashes pairs
	fs := flag.NewFlagSet("osimage-install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	sourceRef := fs.String("source-ref", envOr("OSIMAGE_SOURCE_REF", defRef), "")
	targetRef := fs.String("target-ref", envOr("OSIMAGE_TARGET_REF", defRef), "")
	dryRun := fs.Bool("dry-run", false, "")
	fs.StringVar(&f.hostname, "hostname", "", "")
	fs.BoolVar(&f.tailscale, "tailscale", false, "")
	fs.StringVar(&f.tsKey, "tailscale-authkey-file", "", "")
	fs.StringVar(&f.tsHostname, "tailscale-hostname", "", "")
	fs.BoolVar(&f.tsSSH, "tailscale-ssh", false, "")
	fs.StringVar(&f.tsTags, "tailscale-tags", "", "")
	fs.BoolVar(&f.nonInteractive, "non-interactive", false, "")
	fs.Var(&f.prompt, "password", "")
	fs.Var(&pwFiles, "password-file", "")
	fs.Var(&pwHashes, "password-hash", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	isTerm := term.IsTerminal(int(stdin.Fd()))
	interactive := isTerm && !f.nonInteractive
	// Without a terminal (or with --non-interactive) the disk must be
	// named; on a terminal it is asked for below.
	if fs.NArg() < 1 && !interactive {
		fs.Usage()
		return errors.New("no target disk given")
	}

	disk, bootcArgs := "", []string(nil)
	if fs.NArg() > 0 {
		disk, bootcArgs = fs.Arg(0), fs.Args()[1:]
	}

	cfg := &installer.Config{
		Disk:      disk,
		SourceRef: *sourceRef,
		TargetRef: *targetRef,
		BootcArgs: bootcArgs,
		Hostname:  f.hostname,
		DryRun:    *dryRun,
		Stdout:    stdout,
		Stderr:    stderr,
		Paths:     installer.DefaultPaths,
	}

	for _, p := range pwHashes {
		user, hash, ok := strings.Cut(p, "=")
		if !ok {
			return fmt.Errorf("--password-hash %q: want USER=HASH", p)
		}
		cfg.Passwords = append(cfg.Passwords, installer.Password{User: user, Hash: hash})
	}
	for _, p := range pwFiles {
		user, path, ok := strings.Cut(p, "=")
		if !ok {
			return fmt.Errorf("--password-file %q: want USER=PATH", p)
		}
		pw, err := readPasswordFile(path)
		if err != nil {
			return fmt.Errorf("password for %s: %w", user, err)
		}
		hash, err := sha512crypt.Generate(pw)
		if err != nil {
			return err
		}
		cfg.Passwords = append(cfg.Passwords, installer.Password{User: user, Hash: hash})
	}

	if f.tailscale || f.tsKey != "" {
		cfg.Tailscale = f.tailscaleOptions()
	} else if f.tsHostname != "" || f.tsSSH || f.tsTags != "" {
		return errors.New("--tailscale-* options need --tailscale")
	}

	p := newTermPrompter(stdin, stderr)
	if fs.NArg() < 1 {
		// Interactive with no disk on the command line: ask which one
		// to install to.
		if err := chooseDisk(ctx, p, cfg); err != nil {
			return err
		}
	}

	if err := cfg.Validate(); err != nil {
		return err
	}
	// Explicitly prompted users are checked now too, so a typo fails
	// before any prompt.
	given := map[string]bool{}
	for _, p := range cfg.Passwords {
		given[p.User] = true
	}
	for _, user := range f.prompt {
		if err := shadow.ValidateUser(user); err != nil {
			return err
		}
		if given[user] {
			return fmt.Errorf("password for %s given more than once", user)
		}
		given[user] = true
	}

	t, err := installer.Preflight(ctx, cfg)
	if err != nil {
		return err
	}

	if interactive {
		if err := ask(ctx, p, cfg, &f); err != nil {
			return err
		}
	}

	for _, user := range f.prompt {
		if !isTerm {
			return fmt.Errorf("a terminal is required to prompt for %s's password (use --password-file or --password-hash)", user)
		}
		pw, err := p.password(user, "")
		if err != nil {
			return err
		}
		if err := addPassword(cfg, user, pw); err != nil {
			return err
		}
	}

	if cfg.Tailscale != nil {
		if err := prepareTailscale(ctx, cfg, isTerm, stdout, stderr); err != nil {
			return err
		}
	}

	if interactive {
		q := wipeQuestion(cfg.Disk, t)
		cfg.Confirm = func() (bool, error) {
			return p.yesNo(q, false)
		}
	}
	return installer.Install(ctx, cfg, t)
}

// wipeQuestion is the erase confirmation. A disk that already holds
// data gets it named, so the wipe is explicit.
func wipeQuestion(disk string, t *installer.Target) string {
	if t.Content == "" {
		return fmt.Sprintf("Erase %s (%s) and install?", disk, t.Ident)
	}
	return fmt.Sprintf("Erase %s (%s, currently %s) and install?", disk, t.Ident, t.Content)
}

// ask fills in the hostname, root's password and the Tailscale choice
// on the terminal, with the command line values as defaults.
func ask(ctx context.Context, p *prompter, cfg *installer.Config, f *flags) error {
	def := cfg.Hostname
	if def == "" {
		def = imageHostname()
	}
	name, err := p.hostname(def)
	if err != nil {
		return err
	}
	if cfg.Hostname != "" || name != imageHostname() {
		cfg.Hostname = name
	}

	// root, unless --password root asks for it explicitly later
	if !slices.Contains(f.prompt, "root") {
		keep := "no password (image default)"
		for _, pw := range cfg.Passwords {
			if pw.User == "root" {
				keep = "the password from the command line"
			}
		}
		pw, err := p.password("root", keep)
		if err != nil {
			return err
		}
		if pw != nil {
			if err := addPassword(cfg, "root", pw); err != nil {
				return err
			}
		}
	}

	// Tailscale: offered when the installer has it; the default is yes
	// when asked for on the command line or already logged in.
	if _, err := exec.LookPath("tailscale"); err != nil && cfg.Tailscale == nil {
		return nil
	}
	defTS := cfg.Tailscale != nil
	if st, err := tailscale.GetStatus(ctx); err == nil && st.Running() {
		defTS = true
		fmt.Fprintf(p.out, "Tailscale: this installer is %s\n", st.Describe())
	}
	useTS, err := p.yesNo("Move this installer's Tailscale node to the installed system?", defTS)
	if err != nil {
		return err
	}
	if !useTS {
		cfg.Tailscale = nil
		return nil
	}
	if cfg.Tailscale == nil {
		cfg.Tailscale = f.tailscaleOptions()
	}
	if f.tsHostname == "" {
		cfg.Tailscale.Hostname = cfg.Hostname
	}
	if cfg.Tailscale.SSH, err = p.yesNo("Enable Tailscale SSH?", cfg.Tailscale.SSH); err != nil {
		return err
	}
	return cfg.Tailscale.Validate()
}

func (f *flags) tailscaleOptions() *tailscale.Options {
	o := &tailscale.Options{AuthKeyFile: f.tsKey, Hostname: f.tsHostname, SSH: f.tsSSH}
	if o.Hostname == "" {
		o.Hostname = f.hostname
	}
	for tag := range strings.SplitSeq(f.tsTags, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			o.Tags = append(o.Tags, tag)
		}
	}
	return o
}

// addPassword hashes pw and sets it for user, replacing an earlier
// entry (a prompt answer overrides the command line default).
func addPassword(cfg *installer.Config, user string, pw []byte) error {
	hash, err := sha512crypt.Generate(pw)
	if err != nil {
		return err
	}
	for i := range cfg.Passwords {
		if cfg.Passwords[i].User == user {
			cfg.Passwords[i].Hash = hash
			return nil
		}
	}
	cfg.Passwords = append(cfg.Passwords, installer.Password{User: user, Hash: hash})
	return nil
}

// prepareTailscale logs the installer's node in (QR code or auth key)
// and applies the requested name and SSH setting. A dry run only
// reports the node's state.
func prepareTailscale(ctx context.Context, cfg *installer.Config, isTerm bool, stdout, stderr io.Writer) error {
	if cfg.DryRun {
		st, err := tailscale.GetStatus(ctx)
		if err != nil {
			return err
		}
		if st.Running() {
			fmt.Fprintf(stdout, "== tailscale: installer node %s\n", st.Describe())
		} else {
			fmt.Fprintf(stdout, "== tailscale: installer is %s; a real run logs it in first\n", st.BackendState)
		}
		return nil
	}
	st, err := tailscale.Prepare(ctx, *cfg.Tailscale, isTerm, stdout, stderr)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "== tailscale: installer node %s\n", st.Describe())
	return nil
}

func imageHostname() string {
	if b, err := os.ReadFile(seedHostname); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return s
		}
	}
	return "osimage"
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func readPasswordFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	line, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return nil, fmt.Errorf("%s: empty password", path)
	}
	return []byte(line), nil
}
