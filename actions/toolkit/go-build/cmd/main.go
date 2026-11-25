package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/go-build"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	cgo      = flag.String("cgo", "0", "Enable cgo (sets CGO_ENABLED=1)")
	ldflags  = flag.String("ldflags", "", "Flags passed to -ldflags verbatim")
	packages = flag.String("packages", "", "Space-separated package patterns to build")
	output   = flag.String("output", "", "Output path or binary name, relative to the workspace")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("go-build", *debug)

	in := gobuild.Inputs{
		Cgo:      *cgo,
		Ldflags:  *ldflags,
		Packages: *packages,
		Output:   *output,
	}
	if err := gobuild.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[go-build] ERROR: %v\n", err)
		os.Exit(1)
	}
}
