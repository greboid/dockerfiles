package main

import (
	"flag"
	"fmt"
	"os"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/node-setup"
)

var (
	nodeVersion     = flag.String("node-version", "", "Node.js version to install (22, 22.x, 22.13.1, lts/jod; empty derives from node-version-file, .nvmrc or package.json)")
	nodeVersionFile = flag.String("node-version-file", "", "File to derive the version from (.nvmrc or package.json engines)")
	workingDir      = flag.String("working-dir", "", "Project directory, relative to the workspace")
	installDir      = flag.String("install-dir", "", "Install root, relative to the workspace")
	force           = flag.String("force", "false", "Re-download even when the toolchain is already installed")
	cache           = flag.String("cache", "", "Accepted for setup-node familiarity; ignored")
	packageManager  = flag.String("package-manager", "pnpm", "Package manager to install alongside Node: pnpm, yarn or npm")
	registryURL     = flag.String("registry-url", "", "Registry to configure in the project .npmrc")
	debug           = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	c := gha.ContextFromEnv()
	log := gha.ConfigureLogging("node-setup", *debug)

	in := nodesetup.Inputs{
		NodeVersion:     *nodeVersion,
		NodeVersionFile: *nodeVersionFile,
		WorkingDir:      *workingDir,
		InstallDir:      *installDir,
		Force:           *force,
		Cache:           *cache,
		PackageManager:  *packageManager,
		RegistryURL:     *registryURL,
		Token:           os.Getenv("TOKEN"),
	}
	if err := nodesetup.Run(c, log, in); err != nil {
		fmt.Fprintf(os.Stderr, "[node-setup] ERROR: %v\n", err)
		os.Exit(1)
	}
}
