package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/docker-push"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	tags           = flag.String("tags", "", "Full image references to tag and push, separated by commas or whitespace")
	registries     = flag.String("registries", "", "Comma separated registries for derived tags")
	repository     = flag.String("repository", "", "Image name for derived tags")
	contextPath    = flag.String("context", "", "Build context directory, relative to the workspace")
	file           = flag.String("file", "", "Containerfile/Dockerfile path, relative to the context")
	target         = flag.String("target", "", "Build target stage")
	platforms      = flag.String("platforms", "", "Target platforms, separated by commas or whitespace (e.g. linux/amd64,linux/arm64); more than one builds a multi-platform manifest list")
	buildArgs      = flag.String("build-args", "", "Build-time variables, one KEY=VALUE per line")
	push           = flag.String("push", "true", "Push after building")
	noCache        = flag.String("no-cache", "false", "Do not use the docker build cache")
	registry       = flag.String("registry", "", "Registry for the explicit credentials below")
	username       = flag.String("username", "", "Registry username")
	ghcr           = flag.String("ghcr", "false", "Mirror the built image to the GitHub Container Registry")
	ghcrRepository = flag.String("ghcr-repository", "", "GHCR repository to mirror to, owner/name")
	ghcrUsername   = flag.String("ghcr-username", "", "Username for the ghcr.io login")
	debug          = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("docker-push", *debug)

	in := dockerpush.Inputs{
		Tags:           *tags,
		Registries:     *registries,
		Repository:     *repository,
		Token:          os.Getenv("TOKEN"),
		Context:        *contextPath,
		File:           *file,
		Target:         *target,
		Platforms:      *platforms,
		BuildArgs:      *buildArgs,
		Push:           *push,
		NoCache:        *noCache,
		Registry:       *registry,
		Username:       *username,
		Password:       os.Getenv("PASSWORD"),
		GHCR:           *ghcr,
		GHCRRepository: *ghcrRepository,
		GHCRToken:      os.Getenv("GHCR_TOKEN"),
		GHCRUsername:   *ghcrUsername,
	}
	if err := dockerpush.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[docker-push] ERROR: %v\n", err)
		os.Exit(1)
	}
}
