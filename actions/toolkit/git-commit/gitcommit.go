// Package gitcommit commits the changes present in the workspace —
// all of them, or only the files matching the given pathspecs — and
// optionally pushes the commit to its branch: a compact replacement
// for stefanzweifel/git-auto-commit-action. When nothing ends up
// staged the step succeeds with changes-detected=false and pushes
// nothing.
package gitcommit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gitcmd"
)

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN) so it never shows in argv.
type Inputs struct {
	Message  string
	Files    string
	Name     string
	Email    string
	Push     string
	Branch   string
	Token    string
	Audience string
	Username string
}

// Run commits and optionally pushes, reporting the commit SHA.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	ws := c.Workspace
	if !fileExists(filepath.Join(ws, ".git")) {
		return fmt.Errorf("no .git in the workspace; run git-checkout first")
	}
	message := strings.TrimSpace(in.Message)
	if message == "" {
		return fmt.Errorf("the message input is required")
	}

	var auth *gitcmd.Auth
	defer func() { auth.Cleanup() }()

	// Stage. The runner delivers the pathspecs verbatim, so git's own
	// globbing applies.
	files := strings.Fields(in.Files)
	if len(files) > 0 {
		args := []string{"add"}
		for _, f := range files {
			args = append(args, "--", f)
		}
		if err := gitcmd.Run(ws, nil, args...); err != nil {
			return err
		}
	} else {
		if err := gitcmd.Run(ws, nil, "add", "-A"); err != nil {
			return err
		}
	}

	// Nothing staged: report and succeed without committing or
	// pushing.
	if err := gitcmd.Run(ws, nil, "diff", "--cached", "--quiet"); err == nil {
		head, err := revParse(ws)
		if err != nil {
			return err
		}
		if err := c.Output("changes-detected", "false"); err != nil {
			return err
		}
		if err := c.Output("commit", head); err != nil {
			return err
		}
		log.Info("no changes to commit")
		return nil
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = c.Actor
	}
	if name == "" {
		name = "git-commit"
	}
	email := strings.TrimSpace(in.Email)
	if email == "" {
		host := c.Host()
		if host == "" {
			host = "forgejo.invalid"
		}
		email = fmt.Sprintf("%s@users.noreply.%s", name, host)
	}

	if err := gitcmd.Run(ws, nil,
		"-c", "user.name="+name, "-c", "user.email="+email,
		"commit", "-m", message,
	); err != nil {
		return err
	}
	sha, err := revParse(ws)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("committed %s as %s <%s>", sha, name, email))

	if err := c.Output("changes-detected", "true"); err != nil {
		return err
	}
	if err := c.Output("commit", sha); err != nil {
		return err
	}

	if !gha.Truthy(in.Push, true) {
		return nil
	}
	branch := strings.TrimSpace(in.Branch)
	if branch == "" {
		branch = gitcmd.CurrentBranch(c, ws)
	}
	if branch == "" {
		return fmt.Errorf("cannot push: HEAD is detached and no branch input was given")
	}
	auth, err = gitcmd.Setup(ctx, c, log, in.Token, in.Audience, in.Username)
	if err != nil {
		return err
	}
	log.Info("pushing " + branch)
	if err := gitcmd.Run(ws, auth.Environ(), "push", "origin", "HEAD:refs/heads/"+branch); err != nil {
		return err
	}
	log.Info("pushed " + branch)
	return nil
}

// revParse returns the SHA HEAD points at.
func revParse(ws string) (string, error) {
	out, err := gitcmd.Output(ws, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
