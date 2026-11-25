package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/go-vet"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	packages = flag.String("packages", "./...", "Space-separated package patterns to vet")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("go-vet", *debug)

	in := govet.Inputs{Packages: *packages}
	if err := govet.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[go-vet] ERROR: %v\n", err)
		os.Exit(1)
	}
}
