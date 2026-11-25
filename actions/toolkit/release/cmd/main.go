package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/release"
)

var (
	tag        = flag.String("tag", "", "Tag to release; defaults to the pushed tag name on tag events")
	name       = flag.String("name", "", "Release title; defaults to the tag")
	body       = flag.String("body", "", "Release notes")
	commitish  = flag.String("commitish", "", "Commit to point the tag at; defaults to the commit the workflow runs on")
	assets     = flag.String("assets", "", "Whitespace/newline-separated files to upload, each as path or name=path")
	replace    = flag.String("replace", "true", "Replace an existing release for the tag, its assets included")
	draft      = flag.String("draft", "false", "Mark the release as a draft")
	prerelease = flag.String("prerelease", "false", "Mark the release as a pre-release")
	repository = flag.String("repository", "", "owner/name to release into; defaults to the workflow's repository")
	audience   = flag.String("audience", "", "Audience of a Forgejo Authorized Integration")
	debug      = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("release", *debug)

	in := release.Inputs{
		Tag:        *tag,
		Name:       *name,
		Body:       *body,
		Commitish:  *commitish,
		Assets:     *assets,
		Replace:    *replace,
		Draft:      *draft,
		Prerelease: *prerelease,
		Repository: *repository,
		Token:      os.Getenv("TOKEN"),
		Audience:   *audience,
	}
	if err := release.Run(context.Background(), c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[release] ERROR: %v\n", err)
		os.Exit(1)
	}
}
