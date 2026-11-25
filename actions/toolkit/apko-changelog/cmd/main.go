package main

import (
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/apko-changelog"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	newLockfile = flag.String("new-lockfile", "", "Path (relative to the workspace) of the new apko lock file")
	oldLockfile = flag.String("old-lockfile", "", "Path (relative to the workspace) of the previous apko lock file")
	version     = flag.String("version", "", "Version string to mention in the message; defaults to the VERSTRING environment variable")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("apko-changelog", *debug)

	in := apkochangelog.Inputs{
		NewLockfile: *newLockfile,
		OldLockfile: *oldLockfile,
		Version:     *version,
	}
	if err := apkochangelog.Run(c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[apko-changelog] ERROR: %v\n", err)
		os.Exit(1)
	}
}
