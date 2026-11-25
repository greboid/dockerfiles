// Package govet runs go vet over the packages of the module checked
// out in the workspace — ./... by default — in a container whose image
// ships the Go toolchain. Caches are kept in the workspace via
// internal/gotool.
package govet

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
}

// Run executes go vet.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	if err := gotool.RequiresGoMod(c, "go-vet checks"); err != nil {
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
	log.Info("go vet " + strings.Join(packages, " "))

	cmd := exec.CommandContext(ctx, goBin, append([]string{"vet"}, packages...)...)
	cmd.Dir = c.Workspace
	cmd.Env = gotool.CommandEnv(c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
