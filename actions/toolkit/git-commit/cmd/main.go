package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/git-commit"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	message  = flag.String("message", "", "Commit message (required)")
	files    = flag.String("files", "", "Whitespace-separated pathspecs to stage, with git's own globbing")
	name     = flag.String("name", "", "Committer name; defaults to GITHUB_ACTOR")
	email    = flag.String("email", "", "Committer email; defaults to <name>@users.noreply.<instance host>")
	push     = flag.String("push", "true", "Push the commit to origin after committing")
	branch   = flag.String("branch", "", "Branch to push to; defaults to the checked-out branch")
	audience = flag.String("audience", "", "Audience of a Forgejo Authorized Integration")
	username = flag.String("username", "", "Username for token basic auth")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("git-commit", *debug)

	in := gitcommit.Inputs{
		Message:  *message,
		Files:    *files,
		Name:     *name,
		Email:    *email,
		Push:     *push,
		Branch:   *branch,
		Token:    os.Getenv("TOKEN"),
		Audience: *audience,
		Username: *username,
	}
	if err := gitcommit.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[git-commit] ERROR: %v\n", err)
		os.Exit(1)
	}
}
