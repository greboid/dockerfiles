// Package gotest runs go test over the module checked out in the
// workspace — ./... by default, so the whole module including the
// tests of internal packages — in a container whose image ships the Go
// toolchain. Caches are kept in the workspace via internal/gotool.
package gotest

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gotool"
)

// Inputs carries the action inputs.
type Inputs struct {
	Packages string
	Flags    string
}

// Run executes go test.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	if err := gotool.RequiresGoMod(c, "go-test tests"); err != nil {
		return err
	}

	goBin, err := gotool.Find()
	if err != nil {
		return err
	}
	gotool.Version(log, goBin)

	packages := strings.Fields(in.Packages)
	if len(packages) == 0 {
		packages = []string{"./..."}
	}
	args := append(strings.Fields(in.Flags), packages...)
	log.Info("go test " + strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, goBin, append([]string{"test"}, args...)...)
	cmd.Dir = c.Workspace
	cmd.Env = gotool.CommandEnv(c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
