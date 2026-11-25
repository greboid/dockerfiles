package gotool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

func TestModuleName(t *testing.T) {
	dir := t.TempDir()
	mod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(mod, []byte("module git.example.com/owner/deep/name\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	name, err := ModuleName(mod)
	if err != nil || name != "git.example.com/owner/deep/name" {
		t.Errorf("got %q, %v", name, err)
	}

	if err := os.WriteFile(mod, []byte("go 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ModuleName(mod); err == nil {
		t.Error("missing module directive should error")
	}
}

func TestCacheEnv(t *testing.T) {
	t.Setenv("GOMODCACHE", "")
	t.Setenv("GOCACHE", "")
	c := &gha.Context{Workspace: "/ws"}

	env := CacheEnv(c)
	joined := strings.Join(env, " ")
	if !strings.Contains(joined, "GOMODCACHE=") || !strings.Contains(joined, ".actions") {
		t.Errorf("cache env should default into the workspace: %v", env)
	}

	// An explicitly configured cache wins.
	t.Setenv("GOMODCACHE", "/custom/mods")
	env = CacheEnv(c)
	joined = strings.Join(env, " ")
	if strings.Contains(joined, "GOMODCACHE=") {
		t.Errorf("explicit GOMODCACHE should be left alone: %v", env)
	}
	if !strings.Contains(joined, "GOCACHE=") {
		t.Errorf("GOCACHE should still default: %v", env)
	}
}

func TestCommandEnv(t *testing.T) {
	t.Setenv("GOROOT", "/elsewhere/.actions/go/go1.25.0")
	t.Setenv("GOMODCACHE", "")
	t.Setenv("GOCACHE", "")
	t.Setenv("PATH", "/usr/local/go/bin")
	c := &gha.Context{Workspace: "/ws"}

	env := CommandEnv(c)
	joined := strings.Join(env, " ")
	if strings.Contains(joined, "GOROOT=") {
		t.Errorf("a GOROOT exported for other steps must not reach go invocations: %v", env)
	}
	if !strings.Contains(joined, "PATH=/usr/local/go/bin") {
		t.Errorf("the rest of the environment should be preserved: %v", env)
	}
	if !strings.Contains(joined, "GOMODCACHE=/ws/.actions/gomodcache") ||
		!strings.Contains(joined, "GOCACHE=/ws/.actions/gobuildcache") {
		t.Errorf("workspace cache defaults should be included: %v", env)
	}
}

func TestWithoutEnvKey(t *testing.T) {
	env := WithoutEnvKey([]string{"A=1", "CGO_ENABLED=0", "B=2"}, "CGO_ENABLED")
	if len(env) != 2 || env[0] != "A=1" || env[1] != "B=2" {
		t.Errorf("key should be dropped, got %v", env)
	}

	// Only whole keys match: CGO_ENABLED=0 must not shadow a variable
	// whose name merely extends it.
	env = WithoutEnvKey([]string{"CGO_ENABLED_FOO=1"}, "CGO_ENABLED")
	if len(env) != 1 {
		t.Errorf("longer keys should be kept, got %v", env)
	}
}

func TestRequiresGoMod(t *testing.T) {
	dir := t.TempDir()
	c := &gha.Context{Workspace: dir}
	if err := RequiresGoMod(c, "go-test tests"); err == nil ||
		!strings.Contains(err.Error(), "go-test tests the module at the workspace root") {
		t.Errorf("missing go.mod should error naming the action, got %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RequiresGoMod(c, "go-test tests"); err != nil {
		t.Errorf("present go.mod should pass, got %v", err)
	}
}
