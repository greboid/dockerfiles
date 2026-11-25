package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/go-lint"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var debug = flag.Bool("debug", false, "Enable debug logging")

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("go-lint", *debug)

	if err := golint.Run(context.Background(), c, log, golint.Inputs{}); err != nil {
		fmt.Fprintf(os.Stderr, "[go-lint] ERROR: %v\n", err)
		os.Exit(1)
	}
}
