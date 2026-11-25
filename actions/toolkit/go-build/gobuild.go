// Package gobuild compiles the Go module checked out in the workspace
// with the static-build flags shared by the other actions: netgo and
// osusergo build tags, -a, -trimpath, CGO disabled by default and
// ldflags that strip and statically link the result. It runs in a
// container whose image ships the Go toolchain, so no setup step is
// needed; caches are kept in the workspace via internal/gotool.
package gobuild

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gotool"
)

// Inputs carries the action inputs.
type Inputs struct {
	Cgo      string
	Ldflags  string
	Packages string
	Output   string
}

// Run builds the workspace module and reports the binary path.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	if err := gotool.RequiresGoMod(c, "go-build compiles"); err != nil {
		return err
	}

	goBin, err := gotool.Find()
	if err != nil {
		return err
	}
	gotool.Version(log, goBin)

	output := in.Output
	if output == "" {
		// default to the last path element of the module path in go.mod
		module, err := gotool.ModuleName(filepath.Join(c.Workspace, "go.mod"))
		if err != nil {
			return err
		}
		output = module[strings.LastIndexByte(module, '/')+1:]
		log.Info(fmt.Sprintf("no output given; deriving %q from go.mod", output))
	}
	if !filepath.IsAbs(output) {
		output = filepath.Join(c.Workspace, output)
	}
	if dir := filepath.Dir(output); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	ldflags := in.Ldflags
	if ldflags == "" {
		ldflags = `-w -extldflags "-static" -buildid=.`
	}
	cgo := 0
	if gha.Truthy(in.Cgo, false) {
		cgo = 1
	}

	packages := strings.Fields(in.Packages)
	if len(packages) == 0 {
		packages = []string{"."}
		log.Info("no packages given; building the workspace root package")
	}

	env := gotool.WithoutEnvKey(gotool.CommandEnv(c), "CGO_ENABLED")
	env = append(env, fmt.Sprintf("CGO_ENABLED=%d", cgo))

	args := append([]string{"build",
		"-tags", "netgo,osusergo",
		"-a",
		"-trimpath",
		"-ldflags=" + ldflags,
		"-o", output,
	}, packages...)
	log.Info(fmt.Sprintf("go build -tags netgo,osusergo -a -trimpath -ldflags=%s -o %s %s (CGO_ENABLED=%d)", ldflags, output, strings.Join(packages, " "), cgo))

	cmd := exec.CommandContext(ctx, goBin, args...)
	cmd.Dir = c.Workspace
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := c.Output("binary", output); err != nil {
		return err
	}
	log.Info("built " + output)
	return nil
}
