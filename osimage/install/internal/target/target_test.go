package target

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeInstall lays out a root partition the way bootc's composefs
// install leaves it.
func fakeInstall(t *testing.T, ids ...string) Root {
	t.Helper()
	root := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(root, sharedVar, "lib"), 0o755))
	for _, id := range ids {
		etc := filepath.Join(root, deployDir, id, "etc")
		must(t, os.MkdirAll(etc, 0o755))
		must(t, os.Symlink("../../os/default/var", filepath.Join(root, deployDir, id, "var")))
	}
	return Root{Path: root}
}

func TestLocate(t *testing.T) {
	r := fakeInstall(t, "abc")
	v, err := r.Var()
	if err != nil || v != filepath.Join(r.Path, sharedVar) {
		t.Errorf("Var() = %q, %v", v, err)
	}
}

func TestLocateErrors(t *testing.T) {
	if _, err := (Root{Path: t.TempDir()}).Var(); err == nil {
		t.Error("Var() on an empty root: want error")
	}

}

func TestWriteHostname(t *testing.T) {
	r := fakeInstall(t, "abc")
	v, _ := r.Var()

	must(t, WriteHostname(v, "box1"))
	assertFile(t, filepath.Join(v, stateDir, "hostname"), "box1\n", 0o644)

	fi, err := os.Stat(filepath.Join(v, stateDir))
	must(t, err)
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("state dir mode = %v", fi.Mode().Perm())
	}
	if err := WriteHostname(v, "-bad"); err == nil {
		t.Error("invalid hostname: want error")
	}
}

func TestMachineID(t *testing.T) {
	a, err := NewMachineID()
	must(t, err)
	b, err := NewMachineID()
	must(t, err)
	if a == b || len(a) != 32 {
		t.Fatalf("NewMachineID() = %q, %q", a, b)
	}
	v, _ := fakeInstall(t, "abc").Var()
	must(t, WriteMachineID(v, a))
	assertFile(t, filepath.Join(v, stateDir, "machine-id"), a+"\n", 0o444)
	for _, bad := range []string{"", "0123", strings.ToUpper(a), "0123456789abcdef0123456789abcdeg"} {
		if err := WriteMachineID(v, bad); err == nil {
			t.Errorf("WriteMachineID(%q) = nil, want error", bad)
		}
	}
}

func TestValidateHostname(t *testing.T) {
	for _, h := range []string{"osimage", "box-1", "a.b.c", "X1"} {
		if err := ValidateHostname(h); err != nil {
			t.Errorf("ValidateHostname(%q) = %v", h, err)
		}
	}
	long := ""
	for range 65 {
		long += "a"
	}
	for _, h := range []string{"", "-a", "a-", "a..b", "a_b", "a b", long} {
		if err := ValidateHostname(h); err == nil {
			t.Errorf("ValidateHostname(%q) = nil, want error", h)
		}
	}
}

func assertFile(t *testing.T, path, want string, mode os.FileMode) {
	t.Helper()
	b, err := os.ReadFile(path)
	must(t, err)
	if string(b) != want {
		t.Errorf("%s = %q, want %q", path, b, want)
	}
	fi, err := os.Stat(path)
	must(t, err)
	if fi.Mode().Perm() != mode {
		t.Errorf("%s mode = %v, want %v", path, fi.Mode().Perm(), mode)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestStaticIPv6(t *testing.T) {
	s, err := ParseStaticIPv6("2a03:4000:1:2::1/64", "fe80::1")
	must(t, err)
	v, _ := fakeInstall(t, "abc").Var()
	must(t, WriteStaticIPv6(v, *s))
	assertFile(t, filepath.Join(v, networkDir, ipv6DropIn),
		"[Network]\nAddress=2a03:4000:1:2::1/64\nGateway=fe80::1\n", 0o644)

	for _, bad := range [][2]string{
		{"2a03:4000:1:2::1", "fe80::1"},
		{"192.0.2.1/24", "fe80::1"},
		{"fe80::2/64", "fe80::1"},
		{"2a03:4000:1:2::1/64", "192.0.2.1"},
		{"2a03:4000:1:2::1/64", "fe80::1%eth0"},
	} {
		if _, err := ParseStaticIPv6(bad[0], bad[1]); err == nil {
			t.Errorf("ParseStaticIPv6(%q, %q) = nil, want error", bad[0], bad[1])
		}
	}
}

func TestStaticIPv4(t *testing.T) {
	s, err := ParseStaticIPv4("192.0.2.10/24", "192.0.2.1")
	must(t, err)
	v, _ := fakeInstall(t, "abc").Var()
	must(t, WriteStaticIPv4(v, *s))
	assertFile(t, filepath.Join(v, networkDir, ipv4DropIn),
		"[Network]\nDHCP=ipv6\nAddress=192.0.2.10/24\nGateway=192.0.2.1\n", 0o644)

	for _, bad := range [][2]string{
		{"192.0.2.10", "192.0.2.1"},
		{"2001:db8::1/64", "192.0.2.1"},
		{"0.0.0.0/0", "192.0.2.1"},
		{"192.0.2.10/24", "fe80::1"},
		{"192.0.2.10/24", ""},
	} {
		if _, err := ParseStaticIPv4(bad[0], bad[1]); err == nil {
			t.Errorf("ParseStaticIPv4(%q, %q) = nil, want error", bad[0], bad[1])
		}
	}
}

func TestDNS(t *testing.T) {
	d, err := ParseDNS("1.1.1.1, 2606:4700::1111", "no")
	must(t, err)
	v, _ := fakeInstall(t, "abc").Var()
	must(t, WriteDNS(v, *d))
	assertFile(t, filepath.Join(v, networkDir, dnsDropIn),
		"[Network]\nDNS=1.1.1.1\nDNS=2606:4700::1111\nDNSSEC=no\n", 0o644)

	if d, err := ParseDNS("", ""); d != nil || err != nil {
		t.Errorf("ParseDNS empty = %v, %v; want nil, nil", d, err)
	}
	if d, err := ParseDNS("", "allow-downgrade"); err != nil || d.String() != "dnssec allow-downgrade" {
		t.Errorf("ParseDNS dnssec only = %v, %v", d, err)
	}
	for _, bad := range [][2]string{{"x", ""}, {"0.0.0.0", ""}, {"", "off"}} {
		if _, err := ParseDNS(bad[0], bad[1]); err == nil {
			t.Errorf("ParseDNS(%q, %q) = nil, want error", bad[0], bad[1])
		}
	}
}

func TestCopyNetwork(t *testing.T) {
	src := t.TempDir()
	must(t, os.WriteFile(filepath.Join(src, "10-eth0.network"), []byte("[Network]\nDNSSEC=no\n"), 0o640))
	must(t, os.MkdirAll(filepath.Join(src, "80-dhcp.network.d"), 0o755))
	must(t, os.WriteFile(filepath.Join(src, "80-dhcp.network.d", "x.conf"), []byte("[Network]\n"), 0o644))
	must(t, os.Symlink("/dev/null", filepath.Join(src, "99-mask.network")))

	files, err := NetworkFiles(src)
	must(t, err)
	if strings.Join(files, " ") != "10-eth0.network 80-dhcp.network.d/x.conf 99-mask.network" {
		t.Errorf("NetworkFiles() = %v", files)
	}
	if files, err := NetworkFiles(filepath.Join(src, "missing")); err != nil || len(files) != 0 {
		t.Errorf("NetworkFiles(missing) = %v, %v", files, err)
	}

	v, _ := fakeInstall(t, "abc").Var()
	must(t, CopyNetwork(src, v))
	dst := filepath.Join(v, networkDir)
	assertFile(t, filepath.Join(dst, "10-eth0.network"), "[Network]\nDNSSEC=no\n", 0o640)
	assertFile(t, filepath.Join(dst, "80-dhcp.network.d", "x.conf"), "[Network]\n", 0o644)
	if l, err := os.Readlink(filepath.Join(dst, "99-mask.network")); err != nil || l != "/dev/null" {
		t.Errorf("mask symlink = %q, %v", l, err)
	}
}
