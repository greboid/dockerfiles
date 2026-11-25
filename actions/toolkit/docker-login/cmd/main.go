package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/docker-login"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	registry = flag.String("registry", "", "Registry to log in to")
	username = flag.String("username", "", "Registry username")
	audience = flag.String("audience", "", "Audience of a Forgejo Authorized Integration")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("docker-login", *debug)

	in := dockerlogin.Inputs{
		Registry: *registry,
		Username: *username,
		Audience: *audience,
		Password: os.Getenv("PASSWORD"),
		GHCR:     os.Getenv("GHCR"),
	}
	if err := dockerlogin.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[docker-login] ERROR: %v\n", err)
		os.Exit(1)
	}
}
