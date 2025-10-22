package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/csmith/latest"
	"gopkg.in/yaml.v3"
)

type PackageSpec struct {
	Package struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Epoch   int    `yaml:"epoch"`
	} `yaml:"package"`
	Update struct {
		Enabled bool `yaml:"enabled"`
		Git     *struct {
			StripPrefix     string `yaml:"strip-prefix,omitempty"`
			TagFilterPrefix string `yaml:"tag-filter-prefix,omitempty"`
		} `yaml:"git,omitempty"`
		Latest *struct {
			Identifier  string `yaml:"identifier"`
			StripPrefix string `yaml:"strip-prefix"`
		} `yaml:"latest,omitempty"`
	} `yaml:"update"`
	Pipeline []map[string]interface{} `yaml:"pipeline"`
}

type UpdateResult struct {
	Package    string
	OldVersion string
	NewVersion string
	Updated    bool
	Error      error
}

func checkMelangeInstalled() error {
	cmd := exec.Command("melange", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("melange is not installed or not in PATH")
	}
	return nil
}

func getLatestVersion(ctx context.Context, identifier string) (string, error) {
	// Parse identifier format
	// Supported formats:
	//   github:org/repo
	//   go
	//   postgres:15
	//   alpine:package-name
	//   image:registry/image

	parts := strings.SplitN(identifier, ":", 2)
	source := parts[0]

	switch source {
	case "github":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid github identifier format, expected github:org/repo")
		}
		// GitTag expects a full URL
		repoURL := "https://github.com/" + parts[1]
		// Use IgnoreErrors to skip tags that can't be parsed (like those with ^{} suffix from annotated tags)
		// Use IgnorePreRelease to skip pre-release versions (alpha, beta, rc, etc.)
		return latest.GitTag(ctx, repoURL, &latest.TagOptions{
			IgnoreErrors:     true,
			IgnorePreRelease: true,
		})

	case "go":
		version, _, _, err := latest.GoRelease(ctx, nil)
		if err != nil {
			return "", err
		}
		// GoRelease returns versions like "go1.25.3", strip the "go" prefix
		return strings.TrimPrefix(version, "go"), nil

	case "postgres":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid postgres identifier format, expected postgres:version")
		}
		// Parse major version from identifier (e.g., "15" from "postgres:15")
		majorVersionStr := parts[1]
		majorVersion, err := strconv.Atoi(majorVersionStr)
		if err != nil {
			return "", fmt.Errorf("invalid postgres major version %s: %w", majorVersionStr, err)
		}

		// Use MajorVersionMax to filter to specific major version
		version, _, _, err := latest.PostgresRelease(ctx, &latest.TagOptions{
			MajorVersionMax: majorVersion,
		})
		if err != nil {
			return "", err
		}
		return version, nil

	case "alpine":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid alpine identifier format, expected alpine:package-name")
		}
		version, _, _, err := latest.AlpinePackage(ctx, parts[1], nil)
		return version, err

	case "image":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid image identifier format, expected image:registry/image")
		}
		return latest.ImageTag(ctx, parts[1], nil)

	default:
		return "", fmt.Errorf("unsupported identifier source: %s", source)
	}
}

func getPackagesFromDirectory(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML files found in %s directory", dir)
	}

	var packages []string
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		packages = append(packages, name)
	}

	sort.Strings(packages)
	return packages, nil
}

func readPackageSpec(yamlFile string) (*PackageSpec, error) {
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", yamlFile, err)
	}

	var spec PackageSpec
	if err = yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", yamlFile, err)
	}

	return &spec, nil
}

func extractRepositoryFromPipeline(spec *PackageSpec) (string, error) {
	// Look for git-checkout in pipeline
	for _, step := range spec.Pipeline {
		if uses, ok := step["uses"].(string); ok && uses == "git-checkout" {
			if with, ok := step["with"].(map[string]interface{}); ok {
				if repo, ok := with["repository"].(string); ok {
					return repo, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no git-checkout step found in pipeline")
}

func updatePackage(ctx context.Context, yamlFile string, dryRun bool) (*UpdateResult, error) {
	// Read package spec to get current version
	spec, err := readPackageSpec(yamlFile)
	if err != nil {
		return nil, err
	}

	result := &UpdateResult{
		Package:    spec.Package.Name,
		OldVersion: spec.Package.Version,
	}

	// Skip if update is not enabled
	if !spec.Update.Enabled {
		result.Error = fmt.Errorf("updates not enabled")
		return result, nil
	}

	var latestVersion string
	var stripPrefix string

	// Determine update source and get latest version
	// If Latest is not configured, assume Git update method
	if spec.Update.Latest == nil {
		// Get strip prefix if Git section exists
		if spec.Update.Git != nil {
			stripPrefix = spec.Update.Git.StripPrefix
		}

		var identifier string

		// Special case for Go - use GoRelease function directly
		if spec.Package.Name == "go" {
			identifier = "go"
		} else if strings.HasPrefix(spec.Package.Name, "postgres-") {
			// Special case for Postgres - extract major version from package name
			majorVer := strings.TrimPrefix(spec.Package.Name, "postgres-")
			identifier = "postgres:" + majorVer
		} else {
			// Standard git repository - extract repository from pipeline
			repo, err := extractRepositoryFromPipeline(spec)
			if err != nil {
				result.Error = fmt.Errorf("failed to extract repository: %w", err)
				return result, nil
			}

			if strings.HasPrefix(repo, "https://github.com/") {
				// GitHub repository - convert to github: identifier
				identifier = "github:" + strings.TrimPrefix(repo, "https://github.com/")
			} else {
				// Non-GitHub git repository - use as-is
				identifier = "github:" + strings.TrimPrefix(repo, "https://github.com/")
			}
		}

		var err error
		latestVersion, err = getLatestVersion(ctx, identifier)
		if err != nil {
			result.Error = fmt.Errorf("failed to get latest version: %w", err)
			return result, nil
		}
	} else if spec.Update.Latest != nil {
		// Direct latest identifier
		if spec.Update.Latest.Identifier == "" {
			result.Error = fmt.Errorf("no latest identifier configured")
			return result, nil
		}

		stripPrefix = spec.Update.Latest.StripPrefix

		var err error
		latestVersion, err = getLatestVersion(ctx, spec.Update.Latest.Identifier)
		if err != nil {
			result.Error = fmt.Errorf("failed to get latest version: %w", err)
			return result, nil
		}
	} else {
		result.Error = fmt.Errorf("no update source configured (git or latest)")
		return result, nil
	}

	// Strip prefix if configured
	if stripPrefix != "" {
		latestVersion = strings.TrimPrefix(latestVersion, stripPrefix)
	}

	result.NewVersion = latestVersion

	// Check if already up to date
	if latestVersion == spec.Package.Version {
		result.Updated = false
		return result, nil
	}

	// If dry-run, don't actually update
	if dryRun {
		result.Updated = true
		return result, nil
	}

	// Use melange bump to update the package
	cmd := exec.Command("melange", "bump", yamlFile, latestVersion)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = fmt.Errorf("melange bump failed: %w\nOutput: %s", err, string(output))
		return result, nil
	}

	result.Updated = true
	return result, nil
}

func updatePackages(dir string, dryRun bool, packageName string) error {
	ctx := context.Background()

	var packages []string
	var err error

	if packageName != "" {
		// Update specific package
		packages = []string{packageName}
	} else {
		// Get all packages
		packages, err = getPackagesFromDirectory(dir)
		if err != nil {
			return err
		}
	}

	results := make([]*UpdateResult, 0, len(packages))
	updatedCount := 0
	skippedCount := 0
	errorCount := 0

	fmt.Println("Checking for package updates...")
	if dryRun {
		fmt.Println("(Dry run mode - no changes will be made)")
	}
	fmt.Println()

	for _, pkg := range packages {
		yamlFile := filepath.Join(dir, fmt.Sprintf("%s.yaml", pkg))

		// Check if file exists
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			log.Printf("Warning: package file not found: %s", yamlFile)
			continue
		}

		result, err := updatePackage(ctx, yamlFile, dryRun)
		if err != nil {
			log.Printf("Error processing %s: %v", pkg, err)
			errorCount++
			continue
		}

		results = append(results, result)

		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "updates not enabled") {
				skippedCount++
			} else {
				errorCount++
				log.Printf("Error updating %s: %v", pkg, result.Error)
			}
		} else if result.Updated {
			updatedCount++
			fmt.Printf("✓ %s: %s -> %s\n", result.Package, result.OldVersion, result.NewVersion)
		} else {
			fmt.Printf("  %s: %s (already up to date)\n", result.Package, result.OldVersion)
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total packages checked: %d\n", len(results))
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

func checkCommand() {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	var packageName string
	if fs.NArg() > 0 {
		packageName = fs.Arg(0)
	}

	if err := checkMelangeInstalled(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use dry-run mode for check
	err := updatePackages("packages", true, packageName)
	if err != nil {
		log.Fatalf("Failed to check for updates: %v", err)
	}
}

func updateCommand() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "check for updates without making changes")
	fs.Parse(os.Args[2:])

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

func listCommand() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	packages, err := getPackagesFromDirectory("packages")
	if err != nil {
		log.Fatalf("Failed to list packages: %v", err)
	}

	fmt.Println("Packages with update configuration:")
	fmt.Println()

	for _, pkg := range packages {
		yamlFile := filepath.Join("packages", fmt.Sprintf("%s.yaml", pkg))
		spec, err := readPackageSpec(yamlFile)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", pkg, err)
			continue
		}

		status := "disabled"
		if spec.Update.Enabled {
			status = "enabled"
		}

		fmt.Printf("  %-25s v%-12s [%s]\n", spec.Package.Name, spec.Package.Version, status)
	}
}

func printHelp() {
	fmt.Println("Usage: update <command> [options] [package]")
	fmt.Println()
	fmt.Println("Commands:")
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
	fmt.Println("  update check               # Check all packages for updates")
	fmt.Println("  update check go            # Check only the go package")
	fmt.Println("  update update --dry-run    # Dry-run update of all packages")
	fmt.Println("  update update redis        # Update only the redis package")
	fmt.Println("  update list                # List all packages and versions")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "check":
		checkCommand()
	case "update":
		updateCommand()
	case "list":
		listCommand()
	case "help", "-h", "--help":
		printHelp()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown update subcommand: %s\n\n", subcommand)
		printHelp()
		os.Exit(1)
	}
}
