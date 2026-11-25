// Package gittrypush pushes HEAD to the remote branch — and any tags
// given — retrying when a concurrent push won the race: each failed
// attempt fetches the remote branch and rebases the local commits onto
// it before waiting (with linear backoff) and pushing again. A compact
// replacement for cutlerydrawer/action-git-try-push for matrix jobs
// and pipelines that all push to the same branch.
package gittrypush

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gitcmd"
)

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN) so it never shows in argv.
type Inputs struct {
	Branch      string
	Tags        string
	Remote      string
	Retries     string
	InitialWait string
	WaitFactor  string
	WaitMax     string
	Rebase      string
	PushFlags   string
	Token       string
	Audience    string
	Username    string
}

// Run pushes, retrying with fetch + rebase between attempts.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	ws := c.Workspace
	if !fileExists(filepath.Join(ws, ".git")) {
		return fmt.Errorf("no .git in the workspace; run git-checkout first")
	}

	remote := strings.TrimSpace(in.Remote)
	if remote == "" {
		remote = "origin"
	}
	branch := strings.TrimSpace(in.Branch)
	if branch == "" {
		branch = gitcmd.CurrentBranch(c, ws)
	}
	tags := strings.Fields(in.Tags)

	retries, err := gha.IntValue("retries", in.Retries, 10)
	if err != nil {
		return err
	}
	initialWait, err := gha.IntValue("initial_wait", in.InitialWait, 2)
	if err != nil {
		return err
	}
	waitFactor, err := gha.IntValue("wait_factor", in.WaitFactor, 2)
	if err != nil {
		return err
	}
	waitMax, err := gha.IntValue("wait_max", in.WaitMax, 60)
	if err != nil {
		return err
	}

	if branch == "" && len(tags) == 0 {
		return fmt.Errorf("nothing to push: HEAD is detached and no branch or tags input was given")
	}

	// The push argv, assembled once: the verb, flags, then remote,
	// then refspecs — HEAD for the branch, one refs/tags/<t> per tag.
	args := append([]string{"push"}, strings.Fields(in.PushFlags)...)
	args = append(args, remote)
	if branch != "" {
		args = append(args, "HEAD:refs/heads/"+branch)
	}
	for _, t := range tags {
		args = append(args, "refs/tags/"+t)
	}

	auth, err := gitcmd.Setup(ctx, c, log, in.Token, in.Audience, in.Username)
	if err != nil {
		return err
	}
	defer auth.Cleanup()
	identity, err := gitcmd.DefaultIdentity(c, "")
	if err != nil {
		return err
	}
	env := append(auth.Environ(), identity...)

	what := branch
	if len(tags) > 0 {
		if what != "" {
			what += " "
		}
		what += "+ tags: " + strings.Join(tags, " ")
	}
	log.Info("pushing to " + what)

	attempt := 0
	for {
		err := gitcmd.Run(ws, env, args...)
		if err == nil {
			break
		}
		attempt++
		if attempt > retries {
			return fmt.Errorf("push still failing after %d retries", retries)
		}

		// Catch up with whoever won the race. The fetch updates the
		// remote-tracking ref (so a user-given --force-with-lease has
		// a current expectation); the rebase replays the local
		// commits on top. A fetch failure is not fatal — the push may
		// still succeed on retry — but a conflicted rebase needs a
		// human.
		if branch != "" && gha.Truthy(in.Rebase, true) {
			refspec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, remote, branch)
			if err := gitcmd.Run(ws, env, "fetch", remote, refspec); err != nil {
				log.Warn(fmt.Sprintf("fetch of %s/%s failed; retrying the push as-is", remote, branch))
			} else {
				if err := gitcmd.Run(ws, env, "rebase", "--autostash", fmt.Sprintf("refs/remotes/%s/%s", remote, branch)); err != nil {
					// best-effort cleanup of the conflicted rebase
					_ = gitcmd.Run(ws, env, "rebase", "--abort")
					return fmt.Errorf("rebase onto %s/%s hit conflicts; resolve them manually", remote, branch)
				}
				log.Info("rebased onto " + remote + "/" + branch)
			}
		}

		log.Info(fmt.Sprintf("push rejected (attempt %d/%d); waiting %ds before retrying", attempt, retries, initialWait))
		time.Sleep(time.Duration(initialWait) * time.Second)
		initialWait *= waitFactor
		if initialWait > waitMax {
			initialWait = waitMax
		}
	}

	out, err := gitcmd.Output(ws, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	head := strings.TrimSpace(out)
	if err := c.Output("head", head); err != nil {
		return err
	}
	log.Info("pushed " + head)
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
