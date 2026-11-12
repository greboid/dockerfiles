package tailscale

import (
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoginArgs(t *testing.T) {
	o := Options{Hostname: "box1", SSH: true, Tags: []string{"tag:server", "tag:home"}}
	if got, want := o.LoginArgs(), []string{
		"up", "--reset", "--qr", "--hostname=box1", "--ssh", "--advertise-tags=tag:server,tag:home",
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("LoginArgs() = %v, want %v", got, want)
	}
	o = Options{AuthKeyFile: "/tmp/key"}
	if got, want := o.LoginArgs(), []string{
		"up", "--reset", "--auth-key=file:/tmp/key", "--timeout=2m",
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("LoginArgs() with key = %v, want %v", got, want)
	}
}

func TestSetArgs(t *testing.T) {
	if got := (Options{}).SetArgs(); got != nil {
		t.Errorf("SetArgs() with nothing to change = %v, want nil", got)
	}
	if got, want := (Options{Hostname: "box1"}).SetArgs(), []string{"set", "--hostname=box1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SetArgs() = %v, want %v", got, want)
	}
	if got, want := (Options{SSH: true}).SetArgs(), []string{"set", "--ssh"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SetArgs() = %v, want %v", got, want)
	}
}

func TestParseStatus(t *testing.T) {
	st, err := ParseStatus([]byte(`{"BackendState":"Running","Self":{"HostName":"osimage","DNSName":"box1.tail1234.ts.net.","TailscaleIPs":["100.64.0.5","fd7a:115c:a1e0::5"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !st.Running() || st.Describe() != "box1.tail1234.ts.net (100.64.0.5, fd7a:115c:a1e0::5)" {
		t.Errorf("status = %+v, %q", st, st.Describe())
	}
	st, err = ParseStatus([]byte(`{"BackendState":"NeedsLogin","Self":null}`))
	if err != nil || st.Running() || st.Describe() != "NeedsLogin" {
		t.Errorf("NeedsLogin status = %+v, %v", st, err)
	}
	for _, in := range []string{`nope`, `{}`} {
		if _, err := ParseStatus([]byte(in)); err == nil {
			t.Errorf("ParseStatus(%q): want error", in)
		}
	}
}

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string]Options{
		"missing key": {AuthKeyFile: filepath.Join(dir, "nope")},
		"empty key":   {AuthKeyFile: empty},
		"bad tag":     {Tags: []string{"server"}},
		"bare tag":    {Tags: []string{"tag:"}},
	}
	for name, o := range cases {
		if err := o.Validate(); err == nil {
			t.Errorf("%s: Validate() = nil, want error", name)
		}
	}
}

func TestCopyState(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "var", "lib", "tailscale")
	must(t, os.WriteFile(filepath.Join(src, "tailscaled.state"), []byte(`{"_profiles":"x"}`), 0o600))
	must(t, os.MkdirAll(filepath.Join(src, "ssh"), 0o700))
	must(t, os.WriteFile(filepath.Join(src, "ssh", "ssh_host_ed25519_key"), []byte("key"), 0o600))
	// a socket in the state dir must be skipped, not copied or fatal
	l, err := net.Listen("unix", filepath.Join(src, "sock"))
	must(t, err)
	defer l.Close()

	must(t, CopyState(src, dst))
	for path, want := range map[string]string{
		"tailscaled.state":         `{"_profiles":"x"}`,
		"ssh/ssh_host_ed25519_key": "key",
	} {
		b, err := os.ReadFile(filepath.Join(dst, path))
		if err != nil || string(b) != want {
			t.Errorf("%s = %q, %v; want %q", path, b, err, want)
		}
		fi, _ := os.Stat(filepath.Join(dst, path))
		if fi.Mode().Perm() != 0o600 {
			t.Errorf("%s mode = %v", path, fi.Mode().Perm())
		}
	}
	if _, err := os.Lstat(filepath.Join(dst, "sock")); err == nil {
		t.Error("socket was copied")
	}
	if fi, _ := os.Stat(dst); fi.Mode().Perm() != 0o700 {
		t.Errorf("state dir mode = %v", fi.Mode().Perm())
	}
	if err := CopyState(t.TempDir(), dst); err == nil {
		t.Error("source without tailscaled.state: want error")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
