// Package installer installs the published osimage image onto a disk
// with bootc (composefs backend, systemd-boot, systemd-repart layout)
// and optionally seeds per-machine state onto it before the first boot.
package installer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/greboid/osimage-install/internal/blockdev"
	"github.com/greboid/osimage-install/internal/shadow"
	"github.com/greboid/osimage-install/internal/tailscale"
	"github.com/greboid/osimage-install/internal/target"
)

const (
	// DefaultRef is the published image; also the default upgrade origin.
	DefaultRef = "git.mouse-lake.ts.net/osimage/minimal:latest"

	// rootPartType is the GPT type of the repart root partition
	// (root-x86-64; this installer is x86_64-only).
	rootPartType = "4f68bce3-e8cd-4db1-96e7-fbcaf984b709"
	// minDiskBytes covers the repart profile: 1 GiB ESP + 4 GiB root.
	minDiskBytes = 5 << 30
)

// Paths are the host locations the installer reads and writes;
// overridable for tests.
type Paths struct {
	Sys         string
	Designators string
	RepartDir   string
	ApkDB       string
	StorageConf string
	StoreDir    string
	RunDir      string
	// TailscaleState is the installer tailscaled's state directory.
	TailscaleState string
}

// DefaultPaths are the live installer's paths.
var DefaultPaths = Paths{
	Sys:         "/sys",
	Designators: "/dev/disk/by-designator",
	RepartDir:   "/usr/lib/repart.d",
	ApkDB:       "/lib/apk/db/installed",
	StorageConf: "/etc/containers/storage.conf",
	StoreDir:    "/run/osimage-store",
	RunDir:      "/run",

	TailscaleState: tailscale.DefaultStateDir,
}

// Password is a precomputed crypt(3) hash for one account.
type Password struct {
	User string
	Hash string
}

// Config is one installation.
type Config struct {
	Disk      string
	SourceRef string
	TargetRef string
	// BootcArgs are extra `bootc install to-disk` options.
	BootcArgs []string

	Hostname  string
	Passwords []Password
	// Tailscale, when non-nil, carries the installer's Tailscale node
	// over to the target. The caller logs the node in first
	// (tailscale.Prepare); the installer copies its state.
	Tailscale *tailscale.Options

	DryRun bool
	// Confirm, when set, is asked after the plan is printed and before
	// the disk is touched; false aborts with ErrAborted.
	Confirm func() (bool, error)
	Stdout  io.Writer
	Stderr  io.Writer
	Paths   Paths
}

// ErrAborted is returned when Confirm declines the install.
var ErrAborted = errors.New("aborted; nothing was changed")

// Target is the resolved, preflighted install target.
type Target struct {
	Path           string // /dev/<name>
	Name           string
	Ident          string
	Content        string // what is already on the disk, empty when blank
	ProfileVersion string
}

// Candidate is one whole disk found on the system. Reason is empty
// when preflight would accept the disk; otherwise it says why not.
type Candidate struct {
	Path    string // /dev/<name>
	Name    string
	Size    uint64
	Ident   string
	Content string
	Reason  string
}

// Usable reports whether the installer may target the disk.
func (d Candidate) Usable() bool { return d.Reason == "" }

// Candidates inspects every whole disk and reports, per disk, whether
// the preflight checks would accept it. The interactive disk picker
// offers the usable ones and shows the reasons for the rest.
func Candidates(ctx context.Context, c *Config) ([]Candidate, error) {
	disks, err := blockdev.Disks(ctx)
	if err != nil {
		return nil, err
	}
	return candidates(c, disks)
}

// candidates is Candidates with the disk list injected.
func candidates(c *Config, disks []*blockdev.Device) ([]Candidate, error) {
	sys := blockdev.SysFS{Root: c.Paths.Sys}
	live, err := liveMediaDisks(c.Paths.Designators, sys)
	if err != nil {
		return nil, err
	}
	var out []Candidate
	for _, dev := range disks {
		cand := Candidate{
			Path:    dev.Path,
			Name:    filepath.Base(dev.Path),
			Size:    uint64(dev.Size),
			Ident:   dev.Ident(),
			Content: dev.Describe(),
		}
		// RAM-backed devices pass the size check on big-memory hosts
		// and survive until reboot only; never offer them.
		if strings.HasPrefix(cand.Name, "zram") || strings.HasPrefix(cand.Name, "ram") {
			cand.Reason = fmt.Sprintf("%s is RAM-backed, not a fixed disk", cand.Path)
		} else if err := diskChecks(c, sys, live, cand.Name, dev); err != nil {
			cand.Reason = err.Error()
		}
		out = append(out, cand)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// unsupportedOptions are bootc options the repart/UKI layout owns.
var unsupportedOptions = []string{"--karg", "--karg-delete", "--filesystem", "--root-size", "--block-setup"}

// CheckBootcArgs refuses options that would fight the repart layout or
// the image-baked kernel command line.
func CheckBootcArgs(args []string) error {
	for _, a := range args {
		name, _, _ := strings.Cut(a, "=")
		if slices.Contains(unsupportedOptions, name) {
			return fmt.Errorf("%s is not supported: the kernel command line is baked into the image and the disk layout is defined by the repart profile (machine-specific arguments belong in image variants)", name)
		}
	}
	return nil
}

// transports bootc accepts as an explicit source ref prefix.
var transports = []string{"docker://", "registry:", "oci:", "oci-archive:", "containers-storage:", "docker-archive:", "dir:"}

// SourceImgref adds the registry transport unless one is present.
func SourceImgref(ref string) string {
	for _, t := range transports {
		if strings.HasPrefix(ref, t) {
			return ref
		}
	}
	return "docker://" + ref
}

// Validate checks everything that does not touch the disk.
func (c *Config) Validate() error {
	if c.Disk == "" {
		return errors.New("no target disk given")
	}
	if c.SourceRef == "" || c.TargetRef == "" {
		return errors.New("empty image reference")
	}
	if err := CheckBootcArgs(c.BootcArgs); err != nil {
		return err
	}
	if c.Hostname != "" {
		if err := target.ValidateHostname(c.Hostname); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, p := range c.Passwords {
		if err := shadow.ValidateUser(p.User); err != nil {
			return err
		}
		if err := shadow.ValidateHash(p.Hash); err != nil {
			return fmt.Errorf("password for %s: %w", p.User, err)
		}
		if seen[p.User] {
			return fmt.Errorf("password for %s given more than once", p.User)
		}
		seen[p.User] = true
	}
	if c.Tailscale != nil {
		if c.Tailscale.Hostname != "" {
			if err := target.ValidateHostname(c.Tailscale.Hostname); err != nil {
				return fmt.Errorf("tailscale: %w", err)
			}
		}
		if err := c.Tailscale.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// diskChecks is the per-disk part of preflight: boot media, mounted
// filesystems and size. name is dev's kernel block device name.
func diskChecks(c *Config, sys blockdev.SysFS, live []string, name string, dev *blockdev.Device) error {
	if slices.Contains(live, name) {
		return fmt.Errorf("%s is the installer media this system booted from", dev.Path)
	}
	if m := dev.MountedAt(); len(m) > 0 {
		return fmt.Errorf("%s has a mounted filesystem at %s; unmount it first", dev.Path, m[0])
	}
	if dev.Size < minDiskBytes {
		return fmt.Errorf("%s is %d MiB; the repart profile needs at least %d MiB", dev.Path, dev.Size>>20, minDiskBytes>>20)
	}
	return nil
}

// Preflight resolves the target disk and refuses unsafe targets: a
// partition, the installer's own boot media, a disk with anything
// mounted, a disk too small for the repart profile, or a missing or
// unexpected repart profile.
func Preflight(ctx context.Context, c *Config) (*Target, error) {
	real, err := filepath.EvalSymlinks(c.Disk)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.Disk, err)
	}
	fi, err := os.Stat(real)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeDevice == 0 || fi.Mode()&os.ModeCharDevice != 0 {
		return nil, fmt.Errorf("%s is not a block device", c.Disk)
	}
	sys := blockdev.SysFS{Root: c.Paths.Sys}
	name := filepath.Base(real)
	if sys.IsPartition(name) {
		return nil, fmt.Errorf("%s is a partition; osimage-install installs to a whole disk", c.Disk)
	}

	live, err := liveMediaDisks(c.Paths.Designators, sys)
	if err != nil {
		return nil, err
	}

	dev, err := blockdev.Lsblk(ctx, real)
	if err != nil {
		return nil, err
	}
	if err := diskChecks(c, sys, live, name, dev); err != nil {
		return nil, err
	}

	if _, err := exec.LookPath("systemd-repart"); err != nil {
		return nil, errors.New("systemd-repart is not installed; refusing to install with bootc's fallback partitioner")
	}
	if err := CheckRepartProfile(c.Paths.RepartDir); err != nil {
		return nil, err
	}
	version, err := PackageVersion(c.Paths.ApkDB, "osimage-installer")
	if err != nil || version == "" {
		version = "unknown"
	}
	return &Target{Path: real, Name: name, Ident: dev.Ident(), Content: dev.Describe(), ProfileVersion: version}, nil
}

// liveMediaDisks returns the whole disks backing the live system, found
// through udev's boot-disk designator links (root and ESP).
func liveMediaDisks(designators string, sys blockdev.SysFS) ([]string, error) {
	var out []string
	for _, d := range []string{"root", "esp"} {
		real, err := filepath.EvalSymlinks(filepath.Join(designators, d))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		disk, err := sys.DiskOf(filepath.Base(real))
		if err != nil {
			return nil, err
		}
		if !slices.Contains(out, disk) {
			out = append(out, disk)
		}
	}
	return out, nil
}

// repartProfile is what the installer expects the shipped repart.d
// definitions to say; bootc reads them from the installer root, not
// from the target image.
var repartProfile = []struct {
	file  string
	lines []string
}{
	{"10-esp.conf", []string{"Type=esp", "Format=vfat"}},
	{"20-root.conf", []string{"Type=root-x86-64", "Format=ext4", "SizeMinBytes=4G"}},
}

// CheckRepartProfile verifies the repart definitions in dir.
func CheckRepartProfile(dir string) error {
	for _, def := range repartProfile {
		path := filepath.Join(dir, def.file)
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("repart definition %s missing from installer root", path)
		}
		have := map[string]bool{}
		for line := range strings.SplitSeq(string(b), "\n") {
			have[strings.TrimSpace(line)] = true
		}
		for _, want := range def.lines {
			if !have[want] {
				return fmt.Errorf("%s is not the expected profile (missing %s)", path, want)
			}
		}
	}
	return nil
}

// PackageVersion reads a package's version from the apk database.
func PackageVersion(db, pkg string) (string, error) {
	f, err := os.Open(db)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	current := ""
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			current = ""
			continue
		}
		switch k {
		case "P":
			current = v
		case "V":
			if current == pkg {
				return v, nil
			}
		}
	}
	return "", sc.Err()
}

// BootcCommand is the `bootc install to-disk` argument list.
func BootcCommand(c *Config, t *Target) []string {
	return append([]string{
		"install", "to-disk",
		"--source-imgref", SourceImgref(c.SourceRef),
		"--target-imgref", c.TargetRef,
		"--composefs-backend",
		"--bootloader", "systemd",
		"--run-repart",
		"--wipe", t.Path,
	}, c.BootcArgs...)
}

// Install runs the installation and the post-install seeding.
func Install(ctx context.Context, c *Config, t *Target) error {
	out := c.Stdout
	fmt.Fprintf(out, "== installing %s onto %s (origin: %s)\n", SourceImgref(c.SourceRef), c.Disk, c.TargetRef)
	fmt.Fprintf(out, "== target: %s\n", t.Ident)
	fmt.Fprintf(out, "== repart profile: osimage-installer %s (%s: 1 GiB esp, ext4 root >= 4 GiB)\n", t.ProfileVersion, c.Paths.RepartDir)
	if c.Hostname != "" {
		fmt.Fprintf(out, "== hostname: %s\n", c.Hostname)
	}
	for _, p := range c.Passwords {
		fmt.Fprintf(out, "== password: %s\n", p.User)
	}
	if c.Tailscale != nil {
		fmt.Fprintln(out, "== tailscale: the installer's node moves to the installed system")
	}

	args := BootcCommand(c, t)
	if c.DryRun {
		fmt.Fprintf(out, "== dry run: bootc %s\n", strings.Join(args, " "))
		return nil
	}
	if c.Confirm != nil {
		ok, err := c.Confirm()
		if err != nil {
			return err
		}
		if !ok {
			return ErrAborted
		}
	}

	if err := ensureStorageConf(c.Paths); err != nil {
		return err
	}
	// Modules bootc's composefs staging needs; udev may not have
	// coldplugged them on a live system with no matching root device.
	for _, mod := range []string{"erofs", "overlay"} {
		_ = exec.CommandContext(ctx, "modprobe", mod).Run()
	}

	bootc := exec.CommandContext(ctx, "bootc", args...)
	bootc.Stdin, bootc.Stdout, bootc.Stderr = os.Stdin, c.Stdout, c.Stderr
	if err := bootc.Run(); err != nil {
		return fmt.Errorf("bootc install: %w", err)
	}
	// Always seeded: at least the machine-id is written.
	if err := seed(ctx, c, t); err != nil {
		return fmt.Errorf("the image is installed, but seeding per-machine state failed (unseeded state keeps the image defaults on first boot): %w", err)
	}
	return nil
}

// ensureStorageConf points podman at plain tmpfs storage: its overlay
// driver cannot stack on the live system's overlayfs root. The pull
// lands there, so RAM must cover the uncompressed image.
func ensureStorageConf(p Paths) error {
	if b, err := os.ReadFile(p.StorageConf); err == nil && strings.Contains(string(b), "osimage-store") {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.StorageConf), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.StoreDir, 0o700); err != nil {
		return err
	}
	conf := fmt.Sprintf("[storage]\ndriver = \"overlay\"\nrunroot = %q\ngraphroot = %q\n",
		filepath.Join(p.StoreDir, "runroot"), filepath.Join(p.StoreDir, "graphroot"))
	return os.WriteFile(p.StorageConf, []byte(conf), 0o644)
}

// seed mounts the new root partition and writes the requested state.
func seed(ctx context.Context, c *Config, t *Target) (err error) {
	part, err := rootPartition(ctx, t.Path)
	if err != nil {
		return err
	}
	mnt, err := os.MkdirTemp(c.Paths.RunDir, "osimage-install-root-")
	if err != nil {
		return err
	}
	if err := syscall.Mount(part, mnt, "ext4", 0, ""); err != nil {
		os.Remove(mnt)
		return fmt.Errorf("mounting %s: %w", part, err)
	}
	defer func() {
		syscall.Sync()
		if uerr := syscall.Unmount(mnt, 0); uerr != nil {
			if err == nil {
				err = fmt.Errorf("unmounting %s: %w", mnt, uerr)
			}
			return
		}
		os.Remove(mnt)
	}()

	root := target.Root{Path: mnt}
	varDir, err := root.Var()
	if err != nil {
		return err
	}

	id, err := target.NewMachineID()
	if err != nil {
		return err
	}
	if err := target.WriteMachineID(varDir, id); err != nil {
		return fmt.Errorf("machine-id: %w", err)
	}
	fmt.Fprintf(c.Stdout, "== seeded machine-id %s\n", id)

	if c.Hostname != "" {
		if err := target.WriteHostname(varDir, c.Hostname); err != nil {
			return fmt.Errorf("hostname: %w", err)
		}
		fmt.Fprintf(c.Stdout, "== seeded hostname %s\n", c.Hostname)
	}

	if len(c.Passwords) > 0 {
		etc, err := root.DeploymentEtc()
		if err != nil {
			return err
		}
		content, uid, gid, err := target.ReadShadow(etc)
		if err != nil {
			return err
		}
		day := time.Now().Unix() / 86400
		for _, p := range c.Passwords {
			if content, err = shadow.SetHash(content, p.User, p.Hash, day); err != nil {
				if errors.Is(err, shadow.ErrNoUser) {
					return fmt.Errorf("%w (accounts come from the image's apko accounts: list)", err)
				}
				return err
			}
		}
		if err := target.WriteShadow(varDir, content, uid, gid); err != nil {
			return fmt.Errorf("shadow: %w", err)
		}
		fmt.Fprintf(c.Stdout, "== seeded passwords for %d account(s)\n", len(c.Passwords))
	}

	if c.Tailscale != nil {
		// Copied last, so the snapshot is as fresh as possible.
		if err := tailscale.CopyState(c.Paths.TailscaleState, filepath.Join(varDir, "lib", "tailscale")); err != nil {
			return err
		}
		fmt.Fprintln(c.Stdout, "== seeded tailscale node state")
	}
	return nil
}

// rootPartition finds the repart root partition on the freshly
// partitioned disk, waiting briefly for udev to catch up.
func rootPartition(ctx context.Context, disk string) (string, error) {
	_ = exec.CommandContext(ctx, "udevadm", "settle", "--timeout=30").Run()
	var lastErr error
	for range 10 {
		dev, err := blockdev.Lsblk(ctx, disk)
		if err == nil {
			parts := dev.PartitionsOfType(rootPartType)
			switch len(parts) {
			case 1:
				return parts[0].Path, nil
			case 0:
				err = fmt.Errorf("no root-x86-64 partition on %s", disk)
			default:
				return "", fmt.Errorf("%d root-x86-64 partitions on %s", len(parts), disk)
			}
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return "", lastErr
}
