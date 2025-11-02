package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"package-order/internal/spec"
)

// CheckPackageNeedsBuild determines if a package needs to be built
func CheckPackageNeedsBuild(s *spec.PackageSpec, repoDir string) bool {
	archList := s.Package.TargetArch
	if len(archList) == 0 {
		archList = []string{"x86_64"}
	}

	for _, arch := range archList {
		apkFile := filepath.Join(repoDir, arch,
			fmt.Sprintf("%s-%s-r%d.apk", s.Package.Name, s.Package.Version, s.Package.Epoch))
		if _, err := os.Stat(apkFile); os.IsNotExist(err) {
			return true
		}
	}

	return false
}

// Package builds a single package using melange
func Package(name, packagesDir, repoDir, signingKey, keyring string) error {
	yamlFile := filepath.Join(packagesDir, fmt.Sprintf("%s.yaml", name))

	log.Printf("Building package: %s", name)
	cmd := exec.Command("melange", "build", yamlFile,
		"--signing-key", signingKey,
		"--keyring-append", keyring,
		"--out-dir", repoDir,
		"--repository-append", repoDir,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build package %s: %w", name, err)
	}

	log.Printf("Successfully built package: %s\n", name)
	return nil
}

// Container builds a single container using apko
func Container(name, containersDir, repoDir, keyring, registry, outputDir string, push bool) error {
	yamlFile := filepath.Join(containersDir, fmt.Sprintf("%s.yaml", name))

	log.Printf("Building container: %s", name)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var args []string
	if push {
		// Publish directly to registry (apko will use Docker credential helpers)
		log.Printf("Push enabled: publishing to %s/%s:latest", registry, name)
		args = []string{
			"publish",
			"--repository-append", repoDir,
			"--keyring-append", keyring,
			"--arch", "amd64",
			"--sbom-path", outputDir,
			yamlFile,
			fmt.Sprintf("%s/%s:latest", registry, name),
		}
	} else {
		// Build to tarball
		args = []string{
			"build",
			"--repository-append", repoDir,
			"--keyring-append", keyring,
			"--arch", "amd64",
			"--sbom-path", outputDir,
			yamlFile,
			name,
			filepath.Join(outputDir, fmt.Sprintf("%s.tar.gz", name)),
		}
	}

	cmd := exec.Command("apko", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build container %s: %w", name, err)
	}

	log.Printf("Successfully built container: %s\n", name)
	return nil
}

// CleanupBubblewrapDirs removes stale bubblewrap-guest* directories from /tmp
// Uses chmod -R to fix permissions before removal since bubblewrap creates
// directories with restricted permissions
func CleanupBubblewrapDirs() error {
	tmpDir := "/tmp"
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read /tmp directory: %w", err)
	}

	cleaned := 0
	failed := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "bubblewrap-guest") {
			dirPath := filepath.Join(tmpDir, entry.Name())

			// First try to fix permissions recursively using chmod
			chmodCmd := exec.Command("chmod", "-R", "u+rwX", dirPath)
			if err := chmodCmd.Run(); err != nil {
				log.Printf("Warning: failed to chmod %s: %v", dirPath, err)
			}

			// Now try to remove
			if err := os.RemoveAll(dirPath); err != nil {
				// If os.RemoveAll fails, try using rm -rf as last resort
				rmCmd := exec.Command("rm", "-rf", dirPath)
				if rmErr := rmCmd.Run(); rmErr != nil {
					log.Printf("Warning: failed to remove %s: %v", dirPath, err)
					failed++
				} else {
					cleaned++
				}
			} else {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		log.Printf("Cleaned up %d stale bubblewrap-guest directories", cleaned)
	}
	if failed > 0 {
		log.Printf("Failed to clean up %d bubblewrap-guest directories", failed)
	}

	return nil
}
