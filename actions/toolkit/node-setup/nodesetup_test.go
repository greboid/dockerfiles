package nodesetup

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testIndex() []nodeIndexEntry {
	mk := func(version string, lts any) nodeIndexEntry {
		return nodeIndexEntry{
			Version: version,
			Lts:     lts,
			Files:   []string{"linux-" + nodeArch(runtime.GOARCH)},
		}
	}
	return []nodeIndexEntry{
		mk("v23.0.0-rc.2", false),
		mk("v22.13.1", "Jod"),
		mk("v22.13.0", "Jod"),
		mk("v20.19.0", "Iron"),
		mk("v18.20.5", "Hydrogen"),
		mk("v18.19.0", false),
	}
}

func TestResolveNode(t *testing.T) {
	cases := []struct {
		pattern string
		want    string
		wantErr bool
	}{
		{"latest", "v22.13.1", false}, // rc skipped
		{"current", "v22.13.1", false},
		{"lts/*", "v22.13.1", false},
		{"lts/jod", "v22.13.1", false},
		{"lts/Iron", "v20.19.0", false},
		{"18", "v18.20.5", false}, // Hydrogen lts is newest v18
		{"18.19", "v18.19.0", false},
		{"18.19.0", "v18.19.0", false},
		{"v22.13.1", "v22.13.1", false},
		{"23.0.0-rc.2", "v23.0.0-rc.2", false}, // exact may match pre-release
		{"19", "", true},
		{"junk!!", "", true},
	}
	for _, c := range cases {
		e, filename, err := resolveNode(testIndex(), c.pattern)
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveNode(%q): expected error", c.pattern)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveNode(%q): %v", c.pattern, err)
			continue
		}
		if e.Version != c.want {
			t.Errorf("resolveNode(%q) = %s want %s", c.pattern, e.Version, c.want)
		}
		wantFile := fmt.Sprintf("node-%s-linux-%s.tar.gz", c.want, nodeArch(runtime.GOARCH))
		if filename != wantFile {
			t.Errorf("resolveNode(%q) file %q want %q", c.pattern, filename, wantFile)
		}
	}

	noLinux := []nodeIndexEntry{{Version: "v22.13.1", Lts: "Jod", Files: []string{"win-x64-zip"}}}
	if _, _, err := resolveNode(noLinux, "22"); err == nil {
		t.Error("expected error when no linux tarball is available")
	}
}

func TestNpmAuthKey(t *testing.T) {
	cases := map[string]string{
		"https://registry.npmjs.org/":              "//registry.npmjs.org/:_authToken",
		"https://registry.npmjs.org":               "//registry.npmjs.org/:_authToken",
		"https://ex.com/api/packages/greboid/npm/": "//ex.com/api/packages/greboid/npm/:_authToken",
	}
	for in, want := range cases {
		got, err := npmAuthKey(in)
		if err != nil {
			t.Fatalf("npmAuthKey(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("npmAuthKey(%q) = %q want %q", in, got, want)
		}
	}
}

func TestUpsertNpmrc(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".npmrc")
	os.WriteFile(path, []byte("registry=https://old.example\nsave-exact=true\n"), 0o644)

	entries := []struct{ k, v string }{
		{"registry", "https://new.example"},
		{"//new.example/:_authToken", "tok"},
	}
	if err := upsertNpmrc(path, entries); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	want := "registry=https://new.example\nsave-exact=true\n//new.example/:_authToken=tok\n"
	if string(b) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(b), want)
	}
}

func TestResolveWorkingDir(t *testing.T) {
	ws := t.TempDir()
	sub := filepath.Join(ws, "app")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got, err := resolveWorkingDir(ws, ""); err != nil || got != ws {
		t.Errorf("empty input: got %q, %v; want workspace %q", got, err, ws)
	}
	if got, err := resolveWorkingDir(ws, "app"); err != nil || got != sub {
		t.Errorf("relative input: got %q, %v; want %q", got, err, sub)
	}
	if got, err := resolveWorkingDir(ws, sub); err != nil || got != sub {
		t.Errorf("absolute input: got %q, %v; want %q", got, err, sub)
	}
	if _, err := resolveWorkingDir(ws, "missing"); err == nil {
		t.Error("expected error for a missing directory")
	}
	if _, err := resolveWorkingDir(ws, "package.json"); err == nil {
		t.Error("expected error when working-dir names a file")
	}
}

func TestNodeVersionFromFile(t *testing.T) {
	dir := t.TempDir()

	nvmrc := filepath.Join(dir, ".nvmrc")
	os.WriteFile(nvmrc, []byte("v18.19.0\n"), 0o644)
	if v, err := nodeVersionFromFile(nvmrc); err != nil || v != "v18.19.0" {
		t.Errorf("nvmrc: %q %v", v, err)
	}

	pkg := filepath.Join(dir, "package.json")
	os.WriteFile(pkg, []byte(`{"engines":{"node":">= 18.12"}}`), 0o644)
	if v, err := nodeVersionFromFile(pkg); err != nil || v != "18.12" {
		t.Errorf("package.json: %q %v", v, err)
	}

	os.WriteFile(pkg, []byte(`{}`), 0o644)
	if v, err := nodeVersionFromFile(pkg); err != nil || v != "latest" {
		t.Errorf("missing engines.node: %q %v; want latest", v, err)
	}

	os.WriteFile(pkg, []byte(`{"engines":{"node":"*"}}`), 0o644)
	if _, err := nodeVersionFromFile(pkg); err == nil {
		t.Error("expected error when engines.node has no parseable version")
	}
}

func TestPackageManagerVersion(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")

	write := func(body string) {
		if err := os.WriteFile(pkg, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// No package.json at all.
	if v, err := packageManagerVersion(dir, "pnpm", log); err != nil || v != "" {
		t.Errorf("no package.json: got %q %v; want \"\"", v, err)
	}

	write(`{"packageManager":"pnpm@9.15.0+sha512-abc123"}`)
	if v, err := packageManagerVersion(dir, "pnpm", log); err != nil || v != "9.15.0" {
		t.Errorf("pnpm pin: got %q %v; want 9.15.0", v, err)
	}

	write(`{"packageManager":"yarn@4.1.1"}`)
	if v, err := packageManagerVersion(dir, "yarn", log); err != nil || v != "4.1.1" {
		t.Errorf("yarn pin: got %q %v; want 4.1.1", v, err)
	}
	if v, err := packageManagerVersion(dir, "pnpm", log); err != nil || v != "" {
		t.Errorf("mismatched manager: got %q %v; want \"\"", v, err)
	}

	write(`{}`)
	if v, err := packageManagerVersion(dir, "pnpm", log); err != nil || v != "" {
		t.Errorf("no packageManager field: got %q %v; want \"\"", v, err)
	}

	write(`not json`)
	if _, err := packageManagerVersion(dir, "pnpm", log); err == nil {
		t.Error("expected error for unparseable package.json")
	}
}

func TestSetupPackageManager(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	binDir := t.TempDir()
	workDir := t.TempDir()

	if err := setupPackageManager("npm", binDir, workDir, log); err != nil {
		t.Errorf("npm: %v", err)
	}
	if err := setupPackageManager("bun", binDir, workDir, log); err == nil {
		t.Error("expected error for an unsupported package manager")
	}
	if err := setupPackageManager("", binDir, workDir, log); err == nil {
		t.Error("expected error for an empty package manager")
	}

	// A manager already present in the toolchain is reused without
	// touching npm.
	if err := os.WriteFile(filepath.Join(binDir, "pnpm"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := setupPackageManager("pnpm", binDir, workDir, log); err != nil {
		t.Errorf("cached pnpm: %v", err)
	}
}

func TestWithoutEnv(t *testing.T) {
	t.Setenv("PATH", "/original")
	env := withoutEnv("PATH")
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			t.Errorf("withoutEnv left a PATH entry behind: %q", e)
		}
	}
	// Appending a fresh value must not create a duplicate.
	env = append(env, "PATH=/new")
	count := 0
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("got %d PATH entries, want 1", count)
	}
}

func TestMuslFilename(t *testing.T) {
	got := muslFilename("node-v22.13.1-linux-x64.tar.gz")
	want := "node-v22.13.1-linux-x64-musl.tar.gz"
	if got != want {
		t.Errorf("muslFilename = %q want %q", got, want)
	}
}

func TestDetectMusl(t *testing.T) {
	stat := func(paths ...string) func(string) (fs.FileInfo, error) {
		set := map[string]bool{}
		for _, p := range paths {
			set[p] = true
		}
		return func(p string) (fs.FileInfo, error) {
			if set[p] {
				return nil, nil
			}
			return nil, os.ErrNotExist
		}
	}
	cases := []struct {
		name   string
		goarch string
		paths  []string
		want   bool
	}{
		{"alpine amd64", "amd64", []string{"/lib/ld-musl-x86_64.so.1"}, true},
		{"alpine arm64", "arm64", []string{"/lib/ld-musl-aarch64.so.1"}, true},
		{"glibc amd64", "amd64", []string{"/lib64/ld-linux-x86-64.so.2"}, false},
		{"glibc arm64", "arm64", []string{"/lib/ld-linux-aarch64.so.1"}, false},
		{"gcompat shim", "amd64", []string{"/lib/ld-musl-x86_64.so.1", "/lib64/ld-linux-x86-64.so.2"}, false},
		{"musl loader for another arch", "amd64", []string{"/lib/ld-musl-aarch64.so.1"}, false},
		{"no loaders", "amd64", nil, false},
		{"unsupported arch", "ppc64le", []string{"/lib/ld-musl-powerpc64le.so.1"}, false},
	}
	for _, c := range cases {
		if got := detectMusl(c.goarch, stat(c.paths...)); got != c.want {
			t.Errorf("%s: detectMusl = %v want %v", c.name, got, c.want)
		}
	}
}

func TestCheckNodeRuns(t *testing.T) {
	dir := t.TempDir()
	if err := checkNodeRuns(dir); err == nil {
		t.Error("expected an error when node is missing from the toolchain")
	}
	if err := os.WriteFile(filepath.Join(dir, "node"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := checkNodeRuns(dir); err != nil {
		t.Errorf("node present: %v", err)
	}
}

func TestParseAPKINDEX(t *testing.T) {
	index := []byte("P:libgcc\nV:15.2.0-r5\nA:x86_64\nC:Q1LSywFilrZ7CEu/f6MfIEV5TB+sY=\n\nP:libstdc++\nV:15.2.0-r5\nC:Q16OyN4lX+7wXQ0QH7GOJr46LbZ7U=\n")
	pkgs := parseAPKINDEX(index)
	if pkgs["libgcc"].version != "15.2.0-r5" {
		t.Errorf("libgcc version = %q", pkgs["libgcc"].version)
	}
	if pkgs["libstdc++"].version != "15.2.0-r5" {
		t.Errorf("libstdc++ version = %q", pkgs["libstdc++"].version)
	}
	if _, ok := pkgs["missing"]; ok {
		t.Error("unexpected package parsed")
	}
}

func TestChildEnv(t *testing.T) {
	t.Setenv("LD_LIBRARY_PATH", "")
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")

	env := childEnv(bin)
	if paths := grepEnv(env, "PATH"); len(paths) != 1 || !strings.HasPrefix(paths[0], bin) {
		t.Errorf("PATH entries = %q, want exactly one starting with %s", paths, bin)
	}
	if lds := grepEnv(env, "LD_LIBRARY_PATH"); len(lds) != 0 {
		t.Errorf("LD_LIBRARY_PATH entries = %q, want none without a compat dir", lds)
	}

	compat := compatLibDir(bin)
	if err := os.MkdirAll(compat, 0o755); err != nil {
		t.Fatal(err)
	}
	env = childEnv(bin)
	if lds := grepEnv(env, "LD_LIBRARY_PATH"); len(lds) != 1 || lds[0] != compat {
		t.Errorf("LD_LIBRARY_PATH entries = %q, want exactly %s", lds, compat)
	}
}

func grepEnv(env []string, key string) []string {
	var out []string
	for _, e := range env {
		if after, ok := strings.CutPrefix(e, key+"="); ok {
			out = append(out, after)
		}
	}
	return out
}

func TestFindChecksum(t *testing.T) {
	body := []byte("abc123  node-v18.19.0-linux-x64.tar.gz\ndef456  node-v18.19.0-linux-x64.tar.xz\n")
	if sum, err := findChecksum(body, "node-v18.19.0-linux-x64.tar.gz"); err != nil || sum != "abc123" {
		t.Errorf("got %q %v", sum, err)
	}
	if _, err := findChecksum(body, "missing"); err == nil {
		t.Error("expected error for missing file")
	}
}
