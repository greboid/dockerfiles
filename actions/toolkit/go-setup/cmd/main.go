package main

import (
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/go-setup"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

var (
	goVersion     = flag.String("go-version", "", "Go version to install (1.22, 1.22.x, 1.22.5; empty derives from go-version-file or go.mod)")
	goVersionFile = flag.String("go-version-file", "", "File to derive the version from (toolchain/go directive)")
	installDir    = flag.String("install-dir", "", "Install root, relative to the workspace")
	cache         = flag.String("cache", "true", "Use the runner cache for toolchain archives")
	force         = flag.String("force", "false", "Re-download even when the toolchain is already installed")
	debug         = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("go-setup", *debug)

	in := gosetup.Inputs{
		GoVersion:     *goVersion,
		GoVersionFile: *goVersionFile,
		InstallDir:    *installDir,
		Cache:         *cache,
		Force:         *force,
	}
	if err := gosetup.Run(c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[go-setup] ERROR: %v\n", err)
		os.Exit(1)
	}
}
