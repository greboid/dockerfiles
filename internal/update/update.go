package update

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/csmith/latest"

	"package-order/internal/spec"
)

// BumpPackageVersion updates a package YAML file to a new version using melange bump
func BumpPackageVersion(yamlFile, version string) error {
	cmd := exec.Command("melange", "bump", yamlFile, version)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("melange bump failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetLatestVersion checks for the latest version of a package
func GetLatestVersion(ctx context.Context, identifier string) (string, error) {
	// Split only on the first colon to get the source type
	colonIndex := strings.Index(identifier, ":")

	var source, remainder string
	if colonIndex == -1 {
		// No colon means it's a simple identifier like "go" or "postgres"
		source = identifier
		remainder = ""
	} else {
		source = identifier[:colonIndex]
		remainder = identifier[colonIndex+1:]
	}

	switch source {
	case "github":
		repoURL := "https://github.com/" + remainder
		return latest.GitTag(ctx, repoURL, &latest.TagOptions{
			IgnoreErrors:     true,
			IgnorePreRelease: true,
		})

	case "git":
		// Format: git:repo_url or git:repo_url:tag_prefix
		// Find the last colon to separate tag prefix (if present)
		var repoURL string
		var trimPrefixes []string

		// Check if there's a tag prefix by looking for the pattern after the repo URL
		// Repo URLs end with .git or are just https://host/path
		lastColonIndex := strings.LastIndex(remainder, ":")
		if lastColonIndex > strings.Index(remainder, "://") {
			// There's a colon after the protocol, might be a tag prefix
			repoURL = remainder[:lastColonIndex]
			tagPrefix := remainder[lastColonIndex+1:]
			if tagPrefix != "" {
				trimPrefixes = []string{tagPrefix}
			}
		} else {
			repoURL = remainder
		}

		return latest.GitTag(ctx, repoURL, &latest.TagOptions{
			IgnoreErrors:     true,
			IgnorePreRelease: true,
			TrimPrefixes:     trimPrefixes,
		})

	case "go":
		version, _, _, err := latest.GoRelease(ctx, nil)
		if err != nil {
			return "", err
		}
		return strings.TrimPrefix(version, "go"), nil

	case "postgres":
		// Format: postgres or postgres:majorVersion
		majorVersion := 0
		if remainder != "" {
			_, _ = fmt.Sscanf(remainder, "%d", &majorVersion)
		}
		version, _, _, err := latest.PostgresRelease(ctx, &latest.TagOptions{
			MajorVersionMax: majorVersion,
		})
		if err != nil {
			return "", err
		}
		return version, nil

	case "alpine":
		if remainder == "" {
			return "", fmt.Errorf("alpine identifier requires package name (alpine:package-name)")
		}
		version, _, _, err := latest.AlpinePackage(ctx, remainder, nil)
		return version, err

	case "image":
		if remainder == "" {
			return "", fmt.Errorf("image identifier requires registry/image (image:registry/image)")
		}
		return latest.ImageTag(ctx, remainder, nil)

	default:
		return "", fmt.Errorf("unsupported identifier source: %s", source)
	}
}

// CheckPackageUpdates checks if a package has an update available without modifying the YAML
// Returns: (latestVersion string, hasUpdate bool, error)
func CheckPackageUpdates(ctx context.Context, s *spec.PackageSpec) (string, bool, error) {
	if !s.Update.Enabled {
		return s.Package.Version, false, nil
	}

	var identifier string
	var stripPrefix string
	var tagFilterPrefix string

	// Determine update source
	if s.Update.Latest != nil {
		identifier = s.Update.Latest.Identifier
		stripPrefix = s.Update.Latest.StripPrefix
	} else {
		// Use Git method
		if s.Update.Git != nil {
			stripPrefix = s.Update.Git.StripPrefix
			tagFilterPrefix = s.Update.Git.TagFilterPrefix
		}

		if s.Package.Name == "go" {
			identifier = "go"
		} else if strings.HasPrefix(s.Package.Name, "postgres") {
			// Use special postgres release checker with major version
			identifier = "postgres"
			// Extract major version from package name (e.g., postgres-15 -> 15)
			if len(s.Package.Name) > len("postgres-") {
				majorVersion := strings.TrimPrefix(s.Package.Name, "postgres-")
				if majorVersion != s.Package.Name {
					identifier = "postgres:" + majorVersion
				}
			}
		} else {
			repo, err := spec.ExtractRepositoryFromPipeline(s)
			if err != nil {
				return "", false, err
			}
			// Use git: prefix for any git repository, optionally with tag filter
			identifier = "git:" + repo
			if tagFilterPrefix != "" {
				identifier += ":" + tagFilterPrefix
			}
		}
	}

	latestVersion, err := GetLatestVersion(ctx, identifier)
	if err != nil {
		return "", false, err
	}

	if stripPrefix != "" {
		latestVersion = strings.TrimPrefix(latestVersion, stripPrefix)
	}

	hasUpdate := latestVersion != s.Package.Version
	return latestVersion, hasUpdate, nil
}

// PackageVersion updates a package to the latest version if available
func PackageVersion(ctx context.Context, s *spec.PackageSpec, yamlFile string) (bool, error) {
	latestVersion, hasUpdate, err := CheckPackageUpdates(ctx, s)
	if err != nil {
		return false, err
	}

	if !hasUpdate {
		return false, nil
	}

	log.Printf("Updating %s: %s -> %s", s.Package.Name, s.Package.Version, latestVersion)

	// Use melange bump to update
	if err := BumpPackageVersion(yamlFile, latestVersion); err != nil {
		return false, err
	}

	// Reload the spec after update
	updated, err := spec.ReadPackageSpec(yamlFile)
	if err != nil {
		return false, err
	}
	*s = *updated

	return true, nil
}
