package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/go-test"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	packages = flag.String("packages", "./...", "Space-separated package patterns to test")
	flags    = flag.String("flags", "", "Extra whitespace-separated go test flags, e.g. -race -count=1")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("go-test", *debug)

	in := gotest.Inputs{
		Packages: *packages,
		Flags:    *flags,
	}
	if err := gotest.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[go-test] ERROR: %v\n", err)
		os.Exit(1)
	}
}
