package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"package-order/internal/build"
	"package-order/internal/command"
	"package-order/internal/container"
	"package-order/internal/registry"
	"package-order/internal/spec"
	"package-order/internal/update"

	"github.com/csmith/envflag/v2"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "packages":
		packagesCommand()
	case "containers":
		containersCommand()
	case "update":
		updateCommand()
	case "ci":
		ciCommand()
	case "help", "-h", "--help":
		printHelp()
		os.Exit(0)

	default:
		if _, err := fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand); err != nil {
			log.Printf("Warning: failed to write to stderr: %v", err)
		}
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: dockerfiles <command> [subcommand] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  packages <subcommand>    Manage packages")
	fmt.Println("  containers <subcommand>  Manage containers")
	fmt.Println("  update <subcommand>      Update package versions")
	fmt.Println("  ci [options]             CI build mode")
	fmt.Println("  help                     Show this help message")
	fmt.Println()
	fmt.Println("Use 'dockerfiles <command> help' for more information about a command")
	fmt.Println()
}

// dispatchCommand handles subcommand dispatching with help
func dispatchCommand(commandName string, subcommands map[string]func(), printHelp func()) {
	if len(os.Args) < 3 {
		printHelp()
		os.Exit(1)
	}

	subcommand := os.Args[2]

	// Check for help first
	if subcommand == "help" || subcommand == "-h" || subcommand == "--help" {
		printHelp()
		os.Exit(0)
	}

	// Dispatch to the appropriate handler
	if handler, exists := subcommands[subcommand]; exists {
		handler()
	} else {
		if _, err := fmt.Fprintf(os.Stderr, "Unknown %s subcommand: %s\n\n", commandName, subcommand); err != nil {
			log.Printf("Warning: failed to write to stderr: %v", err)
		}
		printHelp()
		os.Exit(1)
	}
}

// packagesCommand handles package subcommands
func packagesCommand() {
	dispatchCommand("packages", map[string]func(){
		"build": packagesBuildCommand,
	}, printPackagesHelp)
}

func printPackagesHelp() {
	fmt.Println("Usage: dockerfiles packages <subcommand> [options] [package]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  build [package]    Build packages in dependency order")
	fmt.Println("                     If package is specified, only build that package and its dependencies")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Build options:")
	fmt.Println("  -force             When package specified: force rebuild of only that package")
	fmt.Println("                     When no package specified: force rebuild of all packages")
	fmt.Println("  -forceall          Force rebuild of specified package and all its dependencies")
	fmt.Println()
}

func packagesBuildCommand() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	force := fs.Bool("force", false, "force rebuild of specified package only (when package is specified)")
	forceAll := fs.Bool("forceall", false, "force rebuild of specified package and all its dependencies")
	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	var packageName string
	if fs.NArg() > 0 {
		packageName = fs.Arg(0)
	}

	if packageName == "" && *force {
		*forceAll = true
		*force = false
	}

	pkgGraph, packageSpecs, err := spec.LoadPackageGraph("packages")
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	buildOrder, err := command.ResolveBuildOrder(pkgGraph, packageName)
	if err != nil {
		log.Fatalf("Failed to resolve build order: %v", err)
	}

	// Clean up any stale bubblewrap directories before starting
	log.Println("Cleaning up stale bubblewrap directories before build...")
	if err := build.CleanupBubblewrapDirs(); err != nil {
		log.Printf("Warning: cleanup failed: %v", err)
	}

	config := command.BuildConfig{
		ItemName:   packageName,
		ItemType:   "package",
		Force:      *force,
		ForceAll:   *forceAll,
		SourceDir:  "packages",
		BuildOrder: buildOrder,
		Graph:      pkgGraph,
	}

	checkNeedsBuild := func(name string) bool {
		s := packageSpecs[name]
		return build.CheckPackageNeedsBuild(s, "./repo")
	}

	buildItem := func(name string) error {
		if err := build.Package(name, "packages", "./repo", "melange.rsa", "melange.rsa.pub"); err != nil {
			// Clean up bubblewrap dirs after a failed build
			if cleanupErr := build.CleanupBubblewrapDirs(); cleanupErr != nil {
				log.Printf("Warning: cleanup after failed build failed: %v", cleanupErr)
			}
			return err
		}
		return nil
	}

	cleanup := func(name string) error {
		// Clean up after each successful build to prevent accumulation
		return build.CleanupBubblewrapDirs()
	}

	if err := command.ExecuteBuild(config, checkNeedsBuild, buildItem, cleanup); err != nil {
		log.Fatalf("%v", err)
	}
}

// containersCommand handles container subcommands
func containersCommand() {
	dispatchCommand("containers", map[string]func(){
		"build": containersBuildCommand,
	}, printContainersHelp)
}

func printContainersHelp() {
	fmt.Println("Usage: dockerfiles containers <subcommand> [options] [container]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  build [container]  Build containers in dependency order")
	fmt.Println("                     If container is specified, only build that container and its dependencies")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Build options:")
	fmt.Println("  -force             When container specified: force rebuild of only that container")
	fmt.Println("                     When no container specified: force rebuild of all containers")
	fmt.Println("  -forceall          Force rebuild of specified container and all its dependencies")
	fmt.Println()
}

func containersBuildCommand() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	force := fs.Bool("force", false, "force rebuild of specified container only (when container is specified)")
	forceAll := fs.Bool("forceall", false, "force rebuild of specified container and all its dependencies")
	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	var containerName string
	if fs.NArg() > 0 {
		containerName = fs.Arg(0)
	}

	if containerName == "" && *force {
		*forceAll = true
		*force = false
	}

	containerGraph, _, err := spec.LoadContainerGraph("containers")
	if err != nil {
		log.Fatalf("Failed to load containers: %v", err)
	}

	buildOrder, err := command.ResolveBuildOrder(containerGraph, containerName)
	if err != nil {
		log.Fatalf("Failed to resolve build order: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	config := command.BuildConfig{
		ItemName:   containerName,
		ItemType:   "container",
		Force:      *force,
		ForceAll:   *forceAll,
		SourceDir:  "containers",
		BuildOrder: buildOrder,
		Graph:      containerGraph,
	}

	checkNeedsBuild := func(name string) bool {
		return container.NeedsBuild(name, "output")
	}

	buildItem := func(name string) error {
		return build.Container(name, "containers", "repo", "melange.rsa.pub", "", "output", false)
	}

	cleanup := func(name string) error {
		// Clean up SBOM files
		return registry.CleanupSBOMs(name, "output")
	}

	if err := command.ExecuteBuild(config, checkNeedsBuild, buildItem, cleanup); err != nil {
		log.Fatalf("%v", err)
	}
}

// updateCommand handles update subcommands
func updateCommand() {
	dispatchCommand("update", map[string]func(){
		"check":  updateCheckCommand,
		"update": updateUpdateCommand,
		"list":   updateListCommand,
	}, printUpdateHelp)
}

func printUpdateHelp() {
	fmt.Println("Usage: dockerfiles update <subcommand> [options] [package]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  check [package]    Check for available updates (dry-run)")
	fmt.Println("                     If package is specified, only check that package")
	fmt.Println("  update [package]   Update packages to latest versions")
	fmt.Println("                     If package is specified, only update that package")
	fmt.Println("  list               List all packages and their update status")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Update options:")
	fmt.Println("  --dry-run          Check for updates without making changes")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dockerfiles update check               # Check all packages for updates")
	fmt.Println("  dockerfiles update check go            # Check only the go package")
	fmt.Println("  dockerfiles update update --dry-run    # Dry-run update of all packages")
	fmt.Println("  dockerfiles update update redis        # Update only the redis package")
	fmt.Println("  dockerfiles update list                # List all packages and versions")
	fmt.Println()
}

func checkMelangeInstalled() error {
	cmd := exec.Command("melange", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("melange is not installed or not in PATH")
	}
	return nil
}

func updatePackages(dir string, dryRun bool, packageName string) error {
	ctx := context.Background()

	var files []string
	if packageName != "" {
		// Update specific package
		files = []string{filepath.Join(dir, fmt.Sprintf("%s.yaml", packageName))}
	} else {
		// Get all packages
		matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no YAML files found in %s directory", dir)
		}
		files = matches
	}

	updatedCount := 0
	skippedCount := 0
	errorCount := 0

	fmt.Println("Checking for package updates...")
	if dryRun {
		fmt.Println("(Dry run mode - no changes will be made)")
	}
	fmt.Println()

	for _, yamlFile := range files {
		// Check if file exists
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			log.Printf("Warning: package file not found: %s", yamlFile)
			continue
		}

		s, err := spec.ReadPackageSpec(yamlFile)
		if err != nil {
			log.Printf("Error reading %s: %v", yamlFile, err)
			errorCount++
			continue
		}

		if !s.Update.Enabled {
			skippedCount++
			continue
		}

		oldVersion := s.Package.Version

		updated, err := update.PackageVersion(ctx, s, yamlFile)
		if err != nil {
			errorCount++
			log.Printf("Error updating %s: %v", s.Package.Name, err)
			continue
		}

		if updated {
			updatedCount++
			fmt.Printf("✓ %s: %s -> %s\n", s.Package.Name, oldVersion, s.Package.Version)
		} else {
			fmt.Printf("  %s: %s (already up to date)\n", s.Package.Name, s.Package.Version)
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total packages checked: %d\n", len(files))
	if updatedCount > 0 {
		fmt.Printf("  Updated: %d\n", updatedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("  Skipped (updates not enabled): %d\n", skippedCount)
	}
	if errorCount > 0 {
		fmt.Printf("  Errors: %d\n", errorCount)
	}

	return nil
}

func updateCheckCommand() {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	var packageName string
	if fs.NArg() > 0 {
		packageName = fs.Arg(0)
	}

	if err := checkMelangeInstalled(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use dry-run mode for check (don't actually update)
	err := updatePackages("packages", true, packageName)
	if err != nil {
		log.Fatalf("Failed to check for updates: %v", err)
	}
}

func updateUpdateCommand() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "check for updates without making changes")
	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	var packageName string
	if fs.NArg() > 0 {
		packageName = fs.Arg(0)
	}

	if err := checkMelangeInstalled(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	err := updatePackages("packages", *dryRun, packageName)
	if err != nil {
		log.Fatalf("Failed to update packages: %v", err)
	}
}

func updateListCommand() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	files, err := filepath.Glob(filepath.Join("packages", "*.yaml"))
	if err != nil {
		log.Fatalf("Failed to list packages: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No YAML files found in packages directory")
	}

	// Sort by name
	packages := make([]string, 0, len(files))
	for _, file := range files {
		packages = append(packages, strings.TrimSuffix(filepath.Base(file), ".yaml"))
	}

	fmt.Println("Packages with update configuration:")
	fmt.Println()

	for _, yamlFile := range files {
		s, err := spec.ReadPackageSpec(yamlFile)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlFile, err)
			continue
		}

		status := "disabled"
		if s.Update.Enabled {
			status = "enabled"
		}

		fmt.Printf("  %-25s v%-12s [%s]\n", s.Package.Name, s.Package.Version, status)
	}
}

// ciCommand handles CI build mode
func ciCommand() {
	registryFlag := flag.String("registry", "reg.g5d.dev", "Docker registry to check and push to")
	push := flag.Bool("push", false, "Push images to registry after building")
	checkOnly := flag.Bool("check-only", false, "Only check what needs to be built, don't build")
	signingKey := flag.String("signing-key", "melange.rsa", "Path to melange signing key")
	keyring := flag.String("keyring", "melange.rsa.pub", "Path to keyring")
	packagesDir := flag.String("packages-dir", "packages", "Directory containing package YAML files")
	containersDir := flag.String("containers-dir", "containers", "Directory containing container YAML files")
	repoDir := flag.String("repo-dir", "repo", "Directory for APK repository")
	outputDir := flag.String("output-dir", "output", "Directory for build output and SBOMs")
	updatePackagesFlag := flag.Bool("update", false, "Check for and update packages to latest versions before building")
	containerFlag := flag.String("container", "", "Build only the specified container and its dependencies")
	forcePackages := flag.Bool("force-packages", false, "Force rebuild of package dependencies for the specified container")
	forceAllPackages := flag.Bool("force-all-packages", false, "Force rebuild of all package dependencies (direct and transitive)")

	envflag.Parse()

	ctx := context.Background()

	log.Println("Loading package specifications...")
	packageGraph, packageSpecs, err := spec.LoadPackageGraph(*packagesDir)
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	// Check for updates and create hypothetical specs (without modifying YAMLs yet)
	hypotheticalPackageSpecs := make(map[string]*spec.PackageSpec)
	packagesWithUpdates := make(map[string]string) // name -> new version
	updatesAvailable := false

	if *updatePackagesFlag {
		log.Println("Checking for package updates...")
		for name, s := range packageSpecs {
			latestVersion, hasUpdate, err := update.CheckPackageUpdates(ctx, s)
			if err != nil {
				log.Printf("Warning: failed to check updates for %s: %v", name, err)
				hypotheticalPackageSpecs[name] = s
				continue
			}

			if hasUpdate {
				log.Printf("  - %s: %s -> %s (update available)", name, s.Package.Version, latestVersion)
				updatesAvailable = true
				packagesWithUpdates[name] = latestVersion

				// Create hypothetical spec with new version
				hypotheticalSpec := *s
				hypotheticalSpec.Package.Version = latestVersion
				hypotheticalPackageSpecs[name] = &hypotheticalSpec
			} else {
				hypotheticalPackageSpecs[name] = s
			}
		}
	} else {
		// No update check requested, use current specs
		hypotheticalPackageSpecs = packageSpecs
	}

	log.Println("Loading container specifications...")
	packageNames := make(map[string]bool)
	for name := range packageSpecs {
		packageNames[name] = true
	}
	containerGraph, containerSpecs, err := spec.LoadContainerGraph(*containersDir)
	if err != nil {
		log.Fatalf("Failed to load containers: %v", err)
	}

	// Validate container flag if provided
	if *containerFlag != "" {
		if _, exists := containerSpecs[*containerFlag]; !exists {
			log.Fatalf("Container %s not found", *containerFlag)
		}
	}

	// Collect package dependencies to force rebuild
	var packageDepsToForce map[string]bool

	if *forceAllPackages && *containerFlag == "" {
		// Force rebuild ALL packages when no container is specified
		log.Printf("Forcing rebuild of all packages...")
		packageDepsToForce = make(map[string]bool)
		for name := range packageSpecs {
			packageDepsToForce[name] = true
		}
		log.Printf("Will force rebuild %d package(s)", len(packageDepsToForce))
	} else if *containerFlag != "" && (*forcePackages || *forceAllPackages) {
		// Force rebuild packages related to specific container
		containerSpec := containerSpecs[*containerFlag]

		if *forceAllPackages {
			// Force rebuild all package dependencies (direct and transitive)
			log.Printf("Collecting all package dependencies for container %s...", *containerFlag)
			packageDepsToForce = spec.CollectContainerPackageDeps(containerSpec, packageGraph, packageNames)
		} else if *forcePackages {
			// Force rebuild only direct package dependencies
			log.Printf("Collecting direct package dependencies for container %s...", *containerFlag)
			packageDepsToForce = make(map[string]bool)
			for _, pkg := range containerSpec.Contents.Packages {
				if packageNames[pkg] {
					packageDepsToForce[pkg] = true
				}
			}
		}

		if len(packageDepsToForce) > 0 {
			log.Printf("Will force rebuild %d package(s)", len(packageDepsToForce))
		}
	}

	// Determine which packages need to be built
	log.Println("Determining which packages need to be built...")
	var packagesToBuild []string
	packageBuildOrder, err := packageGraph.TopologicalSort()
	if err != nil {
		log.Fatalf("Failed to sort packages: %v", err)
	}

	for _, name := range packageBuildOrder {
		s := packageSpecs[name]
		shouldBuild := build.CheckPackageNeedsBuild(s, *repoDir)

		// Force rebuild if this package is in the force list
		if packageDepsToForce != nil && packageDepsToForce[name] {
			shouldBuild = true
			log.Printf("  - %s (forced rebuild)", name)
		} else if shouldBuild {
			log.Printf("  - %s (version %s-r%d needs building)", name, s.Package.Version, s.Package.Epoch)
		}

		if shouldBuild {
			packagesToBuild = append(packagesToBuild, name)
		}
	}

	// Determine which containers need to be built
	log.Println("Determining which containers need to be built...")
	var containersToBuild []string
	containerBuildOrder, err := containerGraph.TopologicalSort()
	if err != nil {
		log.Fatalf("Failed to sort containers: %v", err)
	}

	for _, name := range containerBuildOrder {
		// If targeting a specific container, only build that one
		if *containerFlag != "" && name != *containerFlag {
			continue
		}

		s := containerSpecs[name]
		// Use hypothetical package specs to check if container would need rebuilding
		// This allows us to detect if updates would trigger a rebuild before actually updating
		needsBuild, err := registry.CheckContainerNeedsBuild(name, hypotheticalPackageSpecs, s, *registryFlag)
		if err != nil {
			log.Printf("Warning: error checking %s: %v (assuming needs build)", name, err)
			needsBuild = true
		}

		// Force rebuild container if any of its package dependencies were force-rebuilt
		wasForced := false
		if !needsBuild && packageDepsToForce != nil {
			for _, pkg := range s.Contents.Packages {
				if packageDepsToForce[pkg] {
					needsBuild = true
					wasForced = true
					log.Printf("  - %s (forced due to package dependency %s)", name, pkg)
					break
				}
			}
		}

		if needsBuild {
			if !wasForced {
				log.Printf("  - %s", name)
			}
			containersToBuild = append(containersToBuild, name)
		}
	}

	// If updates are available and containers would need rebuilding, apply the updates
	if updatesAvailable && len(containersToBuild) > 0 {
		log.Println("\nApplying package updates that would trigger container rebuilds...")
		for name, newVersion := range packagesWithUpdates {
			yamlFile := filepath.Join(*packagesDir, fmt.Sprintf("%s.yaml", name))
			log.Printf("  - Updating %s to %s", name, newVersion)

			// Use melange bump to update
			if err := update.BumpPackageVersion(yamlFile, newVersion); err != nil {
				log.Printf("Warning: failed to update %s: %v", name, err)
				continue
			}
		}

		// Reload package specs after updates
		log.Println("Reloading package specifications after updates...")
		packageGraph, packageSpecs, err = spec.LoadPackageGraph(*packagesDir)
		if err != nil {
			log.Fatalf("Failed to reload packages after updates: %v", err)
		}

		// Re-determine which packages need to be built with updated specs
		log.Println("Re-determining which packages need to be built with updated versions...")
		packagesToBuild = []string{}
		packageBuildOrder, err = packageGraph.TopologicalSort()
		if err != nil {
			log.Fatalf("Failed to sort packages: %v", err)
		}

		for _, name := range packageBuildOrder {
			s := packageSpecs[name]
			shouldBuild := build.CheckPackageNeedsBuild(s, *repoDir)

			// Force rebuild if this package is in the force list
			if packageDepsToForce != nil && packageDepsToForce[name] {
				shouldBuild = true
			}

			// Force rebuild if this package was just updated
			if _, wasUpdated := packagesWithUpdates[name]; wasUpdated {
				shouldBuild = true
			}

			if shouldBuild {
				packagesToBuild = append(packagesToBuild, name)
			}
		}
	} else if updatesAvailable && len(containersToBuild) == 0 {
		log.Println("\nPackage updates available, but no containers would need rebuilding. Skipping updates.")
	}

	// Print summary
	log.Printf("\nSummary:")
	log.Printf("  Packages to build: %d", len(packagesToBuild))
	log.Printf("  Containers to build: %d", len(containersToBuild))

	if *checkOnly {
		log.Println("\nCheck-only mode: exiting without building")
		return
	}

	if len(packagesToBuild) == 0 && len(containersToBuild) == 0 {
		log.Println("\nNothing to build!")
		return
	}

	// Create repo directory if it doesn't exist
	if err := os.MkdirAll(*repoDir, 0755); err != nil {
		log.Fatalf("Failed to create repo directory: %v", err)
	}

	// Build packages
	if len(packagesToBuild) > 0 {
		log.Printf("\nBuilding %d packages...", len(packagesToBuild))
		for _, name := range packagesToBuild {
			if err := build.Package(name, *packagesDir, *repoDir, *signingKey, *keyring); err != nil {
				log.Fatalf("Failed to build package %s: %v", name, err)
			}
		}
	}

	// Build containers
	if len(containersToBuild) > 0 {
		log.Printf("\nBuilding %d containers...", len(containersToBuild))
		for _, name := range containersToBuild {
			if err := build.Container(name, *containersDir, *repoDir, *keyring, *registryFlag, *outputDir, *push); err != nil {
				log.Fatalf("Failed to build container %s: %v", name, err)
			}

			// If pushing, attach SBOM to the image
			if *push {
				if err := registry.AttachSBOM(name, *registryFlag, *outputDir); err != nil {
					log.Printf("Warning: failed to attach SBOM for %s: %v", name, err)
				}
			}

			// Clean up SBOM files
			if err := registry.CleanupSBOMs(name, *outputDir); err != nil {
				log.Printf("Warning: SBOM cleanup failed: %v", err)
			}
		}
	}

	log.Println("\nAll builds completed successfully!")
}
