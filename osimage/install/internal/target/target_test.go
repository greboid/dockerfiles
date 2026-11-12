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
		must(t, os.WriteFile(filepath.Join(etc, "shadow"), []byte("root:::0:::::\n"), 0o640))
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
	etc, err := r.DeploymentEtc()
	if err != nil || etc != filepath.Join(r.Path, deployDir, "abc", "etc") {
		t.Errorf("DeploymentEtc() = %q, %v", etc, err)
	}
}

func TestLocateErrors(t *testing.T) {
	if _, err := (Root{Path: t.TempDir()}).Var(); err == nil {
		t.Error("Var() on an empty root: want error")
	}
	if _, err := fakeInstall(t).DeploymentEtc(); err == nil {
		t.Error("no deployments: want error")
	}
	if _, err := fakeInstall(t, "a", "b").DeploymentEtc(); err == nil {
		t.Error("two deployments: want error")
	}
}

func TestWriteHostnameAndShadow(t *testing.T) {
	r := fakeInstall(t, "abc")
	v, _ := r.Var()
	etc, _ := r.DeploymentEtc()

	must(t, WriteHostname(v, "box1"))
	assertFile(t, filepath.Join(v, stateDir, "hostname"), "box1\n", 0o644)

	content, uid, gid, err := ReadShadow(etc)
	must(t, err)
	if string(content) != "root:::0:::::\n" || uid != os.Getuid() || gid != os.Getgid() {
		t.Errorf("ReadShadow() = %q, owner %d:%d", content, uid, gid)
	}
	must(t, WriteShadow(v, []byte("root:$6$s$h:1:0:::::\n"), uid, gid))
	assertFile(t, filepath.Join(v, stateDir, "shadow"), "root:$6$s$h:1:0:::::\n", 0o640)

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
