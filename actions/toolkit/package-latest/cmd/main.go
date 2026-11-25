package main

import (
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/package-latest"
)

var (
	serverURL = flag.String("server-url", "", "Instance base URL; defaults to the workflow's server URL")
	owner     = flag.String("owner", "", "Owner (user or org) whose packages to list; defaults to the workflow's repository owner")
	typ       = flag.String("type", "", "Package type to list (container, generic, ...)")
	pkg       = flag.String("package", "", "Package name to match (substring)")
	prefix    = flag.String("prefix", "", "Only versions starting with this prefix count, e.g. 1.20261007.")
	debug     = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("package-latest", *debug)

	in := packagelatest.Inputs{
		ServerURL: *serverURL,
		Owner:     *owner,
		Type:      *typ,
		Package:   *pkg,
		Prefix:    *prefix,
		Token:     os.Getenv("TOKEN"),
	}
	if err := packagelatest.Run(c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[package-latest] ERROR: %v\n", err)
		os.Exit(1)
	}
}
