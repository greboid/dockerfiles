// Package golint checks that the .go files of the module checked out
// in the workspace are gofmt-formatted, failing the step with the list
// of offenders otherwise. It uses the gofmt shipping with the
// toolchain its image provides. Only the module's own files are
// checked: the trees the go tool ignores for ./... patterns —
// dot- and underscore-prefixed directories (which also keeps the
// go-setup installs under .actions out) and vendor directories — are
// pruned, so the check flags first-party code only.
package golint

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gotool"
)

// Inputs carries the action inputs (go-lint has none but debug).
type Inputs struct{}

// Run checks every module .go file and fails with the offenders.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	if err := gotool.RequiresGoMod(c, "go-lint checks"); err != nil {
		return err
	}

	gofmt, err := gotool.Gofmt()
	if err != nil {
		return err
	}
	if goBin, err := gotool.Find(); err == nil {
		gotool.Version(log, goBin)
	}

	files, err := moduleGoFiles(c.Workspace)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		log.Info("no Go files found; nothing to check")
		return nil
	}

	cmd := exec.CommandContext(ctx, gofmt, append([]string{"-l"}, files...)...)
	cmd.Dir = c.Workspace
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	unformatted := strings.TrimSpace(string(out))
	if unformatted != "" {
		bad := len(strings.Split(unformatted, "\n"))
		return fmt.Errorf("%d file(s) not gofmt-formatted; run gofmt -w on them:\n%s", bad, unformatted)
	}
	log.Info(fmt.Sprintf("all %d file(s) gofmt-formatted", len(files)))
	return nil
}

// moduleGoFiles lists the module's .go files as paths relative to the
// workspace, skipping the trees the go tool ignores for ./... patterns
// (see the package comment), in lexical order.
func moduleGoFiles(ws string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(ws, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != ws && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "vendor") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") {
			rel, err := filepath.Rel(ws, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
