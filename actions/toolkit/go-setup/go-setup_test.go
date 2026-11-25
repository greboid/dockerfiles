package gosetup

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testReleases() []goRelease {
	file := func(name string) []goFile {
		return []goFile{{
			Filename: name,
			OS:       "linux",
			Arch:     runtime.GOARCH,
			SHA256:   "abc",
		}}
	}
	return []goRelease{
		{Version: "go1.23.4", Stable: true, Files: file("go1.23.4.linux-" + runtime.GOARCH + ".tar.gz")},
		{Version: "go1.23.3", Stable: true, Files: file("go1.23.3.linux-" + runtime.GOARCH + ".tar.gz")},
		{Version: "go1.23.0-rc.1", Stable: false, Files: file("go1.23.0-rc.1.linux-" + runtime.GOARCH + ".tar.gz")},
		{Version: "go1.22.10", Stable: true, Files: file("go1.22.10.linux-" + runtime.GOARCH + ".tar.gz")},
		{Version: "go1.22.9", Stable: true, Files: file("go1.22.9.linux-" + runtime.GOARCH + ".tar.gz")},
	}
}

func TestResolveGo(t *testing.T) {
	cases := []struct {
		pattern string
		want    string
		wantErr bool
	}{
		{"", "go1.23.4", false},
		{"stable", "go1.23.4", false},
		{"latest", "go1.23.4", false},
		{"1.x", "go1.23.4", false},
		{"1", "go1.23.4", false},
		{"1.22", "go1.22.10", false},
		{"1.22.x", "go1.22.10", false},
		{"1.22.9", "go1.22.9", false},
		{"1.23.0-rc.1", "go1.23.0-rc.1", false}, // exact match may be a pre-release
		{"1.99", "", true},
		{"garbage", "", true},
	}
	for _, c := range cases {
		r, f, err := resolveGo(testReleases(), c.pattern)
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveGo(%q): expected error", c.pattern)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveGo(%q): %v", c.pattern, err)
			continue
		}
		if r.Version != c.want {
			t.Errorf("resolveGo(%q) = %s want %s", c.pattern, r.Version, c.want)
		}
		if f == nil || f.OS != "linux" || f.Arch != runtime.GOARCH {
			t.Errorf("resolveGo(%q) picked wrong file %+v", c.pattern, f)
		}
	}
}

func TestGoVersionFromMod(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	os.WriteFile(path, []byte("module example.com/x\n\ngo 1.22\n\ntoolchain go1.23.1\n"), 0o644)
	if v, err := goVersionFromMod(path); err != nil || v != "1.23.1" {
		t.Errorf("toolchain directive should win: %q %v", v, err)
	}

	os.WriteFile(path, []byte("module example.com/x\n\ngo 1.22\n"), 0o644)
	if v, err := goVersionFromMod(path); err != nil || v != "1.22" {
		t.Errorf("go directive: %q %v", v, err)
	}

	os.WriteFile(path, []byte("module example.com/x\n"), 0o644)
	if _, err := goVersionFromMod(path); err == nil {
		t.Error("expected error without directives")
	}
}
