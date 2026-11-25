// Package gitcmd wraps the git CLI for the git-* actions: running
// commands in the workspace — with authentication from Setup when the
// operation talks to the remote — plus the small lookups they share
// (current branch, fallback author identity).
package gitcmd

import (
	"os"
	"os/exec"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

// Run executes git in dir (empty: the working directory) with env
// additions (typically Auth.Environ plus DefaultIdentity), streaming
// its output into the workflow log.
func Run(dir string, env []string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Output executes git in dir and returns its stdout.
func Output(dir string, args ...string) (string, error) {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

// DefaultIdentity returns environment additions with a fallback
// author/committer identity for git operations that create commits
// without one — a rebase replaying commits in a fresh container, say.
// Any configured identity wins, in which case nothing is returned.
// The fallback matches git-commit's defaults: the given name (the
// workflow actor by default) at the instance's noreply domain.
func DefaultIdentity(c *gha.Context, name string) ([]string, error) {
	if out, err := Output("", "config", "user.email"); err == nil && strings.TrimSpace(out) != "" {
		return nil, nil
	}
	if name == "" {
		name = c.Actor
	}
	if name == "" {
		name = "forgejo-runner"
	}
	host := c.Host()
	if host == "" {
		host = "forgejo.invalid"
	}
	email := name + "@users.noreply." + host
	return []string{
		"GIT_AUTHOR_NAME=" + name,
		"GIT_AUTHOR_EMAIL=" + email,
		"GIT_COMMITTER_NAME=" + name,
		"GIT_COMMITTER_EMAIL=" + email,
	}, nil
}

// CurrentBranch returns the name of the branch checked out in dir,
// falling back to the workflow ref; empty when HEAD is detached and
// the ref is not a branch. (Branch checkouts by git-checkout have no
// upstream configured, so there is no @{upstream} to ask instead.)
func CurrentBranch(c *gha.Context, dir string) string {
	if out, err := Output(dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		if b := strings.TrimSpace(out); b != "" && b != "HEAD" {
			return b
		}
	}
	if b, ok := strings.CutPrefix(c.Ref, "refs/heads/"); ok {
		return b
	}
	return ""
}
