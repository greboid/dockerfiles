package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/docker-build"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	tags        = flag.String("tags", "", "Full image references to tag the image with, separated by commas or whitespace")
	contextPath = flag.String("context", "", "Build context directory, relative to the workspace")
	file        = flag.String("file", "", "Containerfile/Dockerfile path, relative to the context")
	target      = flag.String("target", "", "Build target stage")
	platforms   = flag.String("platforms", "", "Target platforms, separated by commas or whitespace (e.g. linux/amd64,linux/arm64); more than one builds each platform but loads nothing into the runner engine")
	buildArgs   = flag.String("build-args", "", "Build-time variables, one KEY=VALUE per line")
	noCache     = flag.String("no-cache", "false", "Do not use the docker build cache")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("docker-build", *debug)

	in := dockerbuild.Inputs{
		Tags:      *tags,
		Context:   *contextPath,
		File:      *file,
		Target:    *target,
		Platforms: *platforms,
		BuildArgs: *buildArgs,
		NoCache:   *noCache,
	}
	if err := dockerbuild.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[docker-build] ERROR: %v\n", err)
		os.Exit(1)
	}
}
