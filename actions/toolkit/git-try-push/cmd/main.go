package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/git-try-push"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	branch      = flag.String("branch", "", "Branch to push HEAD to; defaults to the checked-out branch")
	tags        = flag.String("tags", "", "Whitespace-separated tag names to push alongside the branch")
	remote      = flag.String("remote", "origin", "Remote to push to")
	retries     = flag.String("retries", "10", "How many times to retry after a failed push")
	initialWait = flag.String("initial-wait", "2", "Seconds to wait before the first retry")
	waitFactor  = flag.String("wait-factor", "2", "Multiplier applied to the wait after each retry")
	waitMax     = flag.String("wait-max", "60", "Cap, in seconds, of the growing wait")
	rebase      = flag.String("rebase", "true", "Catch up with the remote branch (fetch + rebase --autostash) between attempts")
	pushFlags   = flag.String("push-flags", "", "Extra whitespace-separated git push flags, e.g. --force-with-lease")
	audience    = flag.String("audience", "", "Audience of a Forgejo Authorized Integration")
	username    = flag.String("username", "", "Username for token basic auth")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("git-try-push", *debug)

	in := gittrypush.Inputs{
		Branch:      *branch,
		Tags:        *tags,
		Remote:      *remote,
		Retries:     *retries,
		InitialWait: *initialWait,
		WaitFactor:  *waitFactor,
		WaitMax:     *waitMax,
		Rebase:      *rebase,
		PushFlags:   *pushFlags,
		Token:       os.Getenv("TOKEN"),
		Audience:    *audience,
		Username:    *username,
	}
	if err := gittrypush.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[git-try-push] ERROR: %v\n", err)
		os.Exit(1)
	}
}
