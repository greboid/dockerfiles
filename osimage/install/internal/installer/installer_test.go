package installer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/greboid/osimage-install/internal/blockdev"
	"github.com/greboid/osimage-install/internal/tailscale"
)

func TestSourceImgref(t *testing.T) {
	tests := map[string]string{
		"git.mouse-lake.ts.net/osimage/minimal:latest":       "docker://git.mouse-lake.ts.net/osimage/minimal:latest",
		"docker://quay.io/x:1":                               "docker://quay.io/x:1",
		"oci:/run/media/minimal":                             "oci:/run/media/minimal",
		"oci-archive:/run/media/minimal.tar":                 "oci-archive:/run/media/minimal.tar",
		"containers-storage:localhost/minimal:latest":        "containers-storage:localhost/minimal:latest",
		"localhost:5000/osimage/minimal@sha256:0123456789ab": "docker://localhost:5000/osimage/minimal@sha256:0123456789ab",
	}
	for in, want := range tests {
		if got := SourceImgref(in); got != want {
			t.Errorf("SourceImgref(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCheckBootcArgs(t *testing.T) {
	if err := CheckBootcArgs([]string{"--skip-fetch-check", "--target-transport=registry"}); err != nil {
		t.Error(err)
	}
	for _, bad := range []string{"--karg", "--karg=quiet", "--filesystem=xfs", "--root-size", "--block-setup=tpm2-luks", "--karg-delete=x"} {
		if err := CheckBootcArgs([]string{"--x", bad}); err == nil {
			t.Errorf("CheckBootcArgs(%q) = nil, want error", bad)
		}
	}
}

func TestBootcCommand(t *testing.T) {
	c := &Config{SourceRef: "r/img:1", TargetRef: "r/img:1", BootcArgs: []string{"--skip-fetch-check"}}
	got := BootcCommand(c, &Target{Path: "/dev/vda"})
	want := []string{
		"install", "to-disk",
		"--source-imgref", "docker://r/img:1",
		"--target-imgref", "r/img:1",
		"--composefs-backend", "--bootloader", "systemd", "--run-repart",
		"--wipe", "/dev/vda",
		"--skip-fetch-check",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BootcCommand() = %v, want %v", got, want)
	}
}

func writeProfile(t *testing.T, esp, root string) string {
	t.Helper()
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "10-esp.conf"), []byte(esp), 0o644))
	must(t, os.WriteFile(filepath.Join(dir, "20-root.conf"), []byte(root), 0o644))
	return dir
}

func TestCheckRepartProfile(t *testing.T) {
	good := writeProfile(t,
		"# comment\n[Partition]\nType=esp\nFormat=vfat\nSizeMinBytes=1G\n",
		"[Partition]\nType=root-x86-64\nFormat=ext4\nSizeMinBytes=4G\nGrowFileSystem=yes\n")
	if err := CheckRepartProfile(good); err != nil {
		t.Error(err)
	}
	bad := writeProfile(t, "Type=esp\nFormat=vfat\n", "Type=root-x86-64\nFormat=xfs\nSizeMinBytes=4G\n")
	if err := CheckRepartProfile(bad); err == nil || !strings.Contains(err.Error(), "Format=ext4") {
		t.Errorf("wrong root format: err = %v", err)
	}
	if err := CheckRepartProfile(t.TempDir()); err == nil {
		t.Error("missing profile: want error")
	}
}

func TestPackageVersion(t *testing.T) {
	db := filepath.Join(t.TempDir(), "installed")
	must(t, os.WriteFile(db, []byte(
		"C:Q1abc=\nP:osimage-config\nV:1.0.0-r16\nA:x86_64\n\n"+
			"C:Q1def=\nP:osimage-installer\nV:1.0.0-r3\nA:x86_64\n\n"), 0o644))
	if v, err := PackageVersion(db, "osimage-installer"); err != nil || v != "1.0.0-r3" {
		t.Errorf("PackageVersion() = %q, %v", v, err)
	}
	if v, _ := PackageVersion(db, "nope"); v != "" {
		t.Errorf("missing package: got %q", v)
	}
}

func TestCandidates(t *testing.T) {
	mi := filepath.Join(t.TempDir(), "mountinfo")
	must(t, os.WriteFile(mi, []byte("30 22 8:32 / /run/initramfs/live ro - iso9660 /dev/sdc ro\n"), 0o644))
	c := &Config{Paths: Paths{Mountinfo: mi, Sys: t.TempDir()}}
	disks := []*blockdev.Device{
		{Path: "/dev/vda", Type: "disk", Size: 20 << 30},
		{Path: "/dev/sda", Type: "disk", Size: 120 << 30, Model: "Samsung SSD", Children: []blockdev.Device{
			{Path: "/dev/sda1", Type: "part", FSType: "vfat"},
			{Path: "/dev/sda2", Type: "part", FSType: "ext4"},
		}},
		{Path: "/dev/sdb", Type: "disk", Size: 2 << 30},
		{Path: "/dev/sdc", Type: "disk", Size: 8 << 30},
		{Path: "/dev/sdd", Type: "disk", Size: 8 << 30, Mountpoints: []*string{new("/mnt/data")}},
		{Path: "/dev/zram0", Type: "disk", Size: 16 << 30},
	}
	got, err := candidates(c, disks)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, d := range got {
		paths = append(paths, d.Path)
	}
	want := []string{"/dev/sda", "/dev/sdb", "/dev/sdc", "/dev/sdd", "/dev/vda", "/dev/zram0"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("candidates = %v, want %v", paths, want)
	}
	if got[0].Reason != "" || got[0].Content != "2 partitions (vfat, ext4)" || got[0].Ident != "Samsung SSD" {
		t.Errorf("sda = %+v", got[0])
	}
	if got[4].Reason != "" || got[4].Content != "" {
		t.Errorf("vda = %+v", got[4])
	}
	for _, want := range []struct{ path, reason string }{
		{"/dev/sdb", "needs at least 5120 MiB"},
		{"/dev/sdc", "booted from"},
		{"/dev/sdd", "mounted filesystem at /mnt/data"},
		{"/dev/zram0", "RAM-backed"},
	} {
		var d Candidate
		for _, g := range got {
			if g.Path == want.path {
				d = g
			}
		}
		if d.Usable() || !strings.Contains(d.Reason, want.reason) {
			t.Errorf("%s = %+v; want refused with %q", want.path, d, want.reason)
		}
	}
}

//go:fix inline
func strp(s string) *string { return new(s) }

func TestValidate(t *testing.T) {
	key := filepath.Join(t.TempDir(), "key")
	must(t, os.WriteFile(key, []byte("tskey\n"), 0o600))
	base := func() *Config {
		return &Config{Disk: "/dev/vda", SourceRef: DefaultRef, TargetRef: DefaultRef}
	}
	if err := base().Validate(); err != nil {
		t.Errorf("minimal config: %v", err)
	}
	cases := map[string]func(c *Config){
		"no disk":       func(c *Config) { c.Disk = "" },
		"bad bootc arg": func(c *Config) { c.BootcArgs = []string{"--karg=x"} },
		"bad hostname":  func(c *Config) { c.Hostname = "a_b" },
		"plaintext":     func(c *Config) { c.Passwords = []Password{{"root", "hunter2"}} },
		"duplicate user": func(c *Config) {
			c.Passwords = []Password{{"root", "$6$s$h"}, {"root", "$6$s$i"}}
		},
		"tailscale bad hostname": func(c *Config) {
			c.Tailscale = &tailscale.Options{AuthKeyFile: key, Hostname: "a_b"}
		},
	}
	for name, mutate := range cases {
		c := base()
		mutate(c)
		if err := c.Validate(); err == nil {
			t.Errorf("%s: Validate() = nil, want error", name)
		}
	}
}

func TestDryRun(t *testing.T) {
	var out bytes.Buffer
	c := &Config{
		Disk: "/dev/vda", SourceRef: DefaultRef, TargetRef: DefaultRef,
		Hostname: "box1", Passwords: []Password{{"root", "$6$s$h"}},
		DryRun: true, Stdout: &out, Stderr: &out, Paths: DefaultPaths,
	}
	if err := Install(context.Background(), c, &Target{Path: "/dev/vda", Ident: "QEMU", ProfileVersion: "1.0.0-r3"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"== hostname: box1", "== password: root", "== dry run: bootc install to-disk"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("dry-run output missing %q:\n%s", want, out.String())
		}
	}
}

func TestEnsureStorageConf(t *testing.T) {
	dir := t.TempDir()
	p := Paths{StorageConf: filepath.Join(dir, "etc", "containers", "storage.conf"), StoreDir: filepath.Join(dir, "store")}
	must(t, ensureStorageConf(p))
	b, err := os.ReadFile(p.StorageConf)
	must(t, err)
	if !strings.Contains(string(b), `graphroot = "`+filepath.Join(dir, "store", "graphroot")+`"`) {
		t.Errorf("storage.conf = %q", b)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfirmDeclined(t *testing.T) {
	var out bytes.Buffer
	asked := false
	c := &Config{
		Disk: "/dev/vda", SourceRef: DefaultRef, TargetRef: DefaultRef,
		Stdout: &out, Stderr: &out, Paths: Paths{StorageConf: filepath.Join(t.TempDir(), "must-not-exist")},
		Confirm: func() (bool, error) { asked = true; return false, nil },
	}
	err := Install(context.Background(), c, &Target{Path: "/dev/vda"})
	if !asked || !errors.Is(err, ErrAborted) {
		t.Fatalf("Install() = %v (asked %v), want ErrAborted after asking", err, asked)
	}
	if _, err := os.Stat(c.Paths.StorageConf); err == nil {
		t.Error("declined install still wrote storage.conf")
	}
	if !strings.Contains(out.String(), "== installing") {
		t.Errorf("plan not printed before confirming: %q", out.String())
	}
}
