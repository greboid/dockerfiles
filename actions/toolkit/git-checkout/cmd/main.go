package main

import (
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/git-checkout"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	repository = flag.String("repository", "", "Repository to check out (owner/name)")
	ref        = flag.String("ref", "", "Ref to check out")
	username   = flag.String("username", "", "Username for token basic auth")
	path       = flag.String("path", "", "Directory, relative to the workspace, to check out into")
	fetchDepth = flag.String("fetch-depth", "1", "Number of commits to fetch; 0 fetches all history and tags")
	submodules = flag.String("submodules", "false", "false, true or recursive")
	debug      = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("git-checkout", *debug)

	in := gitcheckout.Inputs{
		Repository: *repository,
		Ref:        *ref,
		Token:      os.Getenv("TOKEN"),
		Username:   *username,
		Path:       *path,
		FetchDepth: *fetchDepth,
		Submodules: *submodules,
	}
	if err := gitcheckout.Run(c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[git-checkout] ERROR: %v\n", err)
		os.Exit(1)
	}
}
