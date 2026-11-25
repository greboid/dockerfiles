// Package gotool is the shared toolchain plumbing of the go-* actions:
// locating the go (and gofmt) binaries the action image ships, logging
// their versions, and pointing the module and build caches at the
// workspace so repeated go-* steps reuse them.
package gotool

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

// Find returns the path of the go binary on PATH.
func Find() (string, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("no Go toolchain found; use an action image with Go")
	}
	return goBin, nil
}

// Gofmt returns the path of the gofmt shipping with the toolchain: it
// sits next to the go binary, with a PATH fallback.
func Gofmt() (string, error) {
	if goBin, err := Find(); err == nil {
		gofmt := filepath.Join(filepath.Dir(goBin), "gofmt")
		if info, err := os.Stat(gofmt); err == nil && !info.IsDir() {
			return gofmt, nil
		}
	}
	gofmt, err := exec.LookPath("gofmt")
	if err != nil {
		return "", fmt.Errorf("no gofmt found next to the go toolchain; use an action image with Go")
	}
	return gofmt, nil
}

// Version logs the toolchain version.
func Version(log *slog.Logger, goBin string) {
	if version, err := exec.Command(goBin, "version").Output(); err == nil {
		log.Info(strings.TrimSpace(string(version)) + " (" + goBin + ")")
	}
}

// CacheEnv returns the environment additions keeping the module and
// build caches inside the workspace, in dot-prefixed directories (which
// the go tool and go-lint both skip), so repeated go-* steps reuse
// them. An explicitly configured cache wins.
func CacheEnv(c *gha.Context) []string {
	var env []string
	if os.Getenv("GOMODCACHE") == "" {
		env = append(env, "GOMODCACHE="+c.ResolvePath(filepath.Join(".actions", "gomodcache")))
	}
	if os.Getenv("GOCACHE") == "" {
		env = append(env, "GOCACHE="+c.ResolvePath(filepath.Join(".actions", "gobuildcache")))
	}
	return env
}

// CommandEnv returns the environment for a `go` child process: the
// current environment without GOROOT, plus the CacheEnv workspace
// cache defaults. The go-* action images each ship their own
// toolchain, and a GOROOT exported into the job for other steps (by
// go-setup, say) would pair this go with that other version's
// compile/asm/vet tools — vet and builds fail with tool version
// mismatches. Without GOROOT the go binary uses the tree it ships in.
func CommandEnv(c *gha.Context) []string {
	return append(WithoutEnvKey(os.Environ(), "GOROOT"), CacheEnv(c)...)
}

// WithoutEnvKey returns env without any KEY= entry, so an appended
// value cannot end up duplicated (with the winner depending on who
// reads the environment first).
func WithoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, prefix) {
			out = append(out, kv)
		}
	}
	return out
}

// RequiresGoMod errors when the workspace has no go.mod, with a
// message naming what the action would have done with the module.
func RequiresGoMod(c *gha.Context, what string) error {
	if info, err := os.Stat(filepath.Join(c.Workspace, "go.mod")); err == nil && !info.IsDir() {
		return nil
	}
	return fmt.Errorf("no go.mod in the workspace; %s the module at the workspace root", what)
}

// ModuleName returns the module path declared in the go.mod at path.
func ModuleName(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("no module directive in %s", path)
}
