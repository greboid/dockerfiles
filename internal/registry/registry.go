package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"package-order/internal/spec"
)

// Image represents image metadata from the registry
type Image struct {
	Name    string
	Version string
	Epoch   int
	Exists  bool
}

// QueryImage checks if an image exists in the registry and gets its metadata
func QueryImage(registry, imageName string) (*Image, error) {
	ctx := context.Background()

	// Create system context - will use Docker credential helpers automatically
	sys := &types.SystemContext{}

	// Parse the image reference
	ref, err := docker.ParseReference("//" + registry + "/" + imageName + ":latest")
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Get the image source
	src, err := ref.NewImageSource(ctx, sys)
	if err != nil {
		// Check if the error is because the image doesn't exist
		errStr := err.Error()
		if strings.Contains(errStr, "manifest unknown") ||
			strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "NAME_UNKNOWN") {
			return &Image{
				Name:   imageName,
				Exists: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to create image source: %w", err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Printf("Warning: failed to close image source: %v", err)
		}
	}()

	// Get the image
	img, err := image.FromUnparsedImage(ctx, sys, image.UnparsedInstance(src, nil))
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	// Get the image config to access labels
	configBlob, err := img.ConfigBlob(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config blob: %w", err)
	}

	var config ocispec.Image
	if err := json.Unmarshal(configBlob, &config); err != nil {
		return nil, fmt.Errorf("failed to parse image config: %w", err)
	}

	// Extract version and epoch from labels (apko sets these)
	version := config.Config.Labels["org.opencontainers.image.version"]
	epochStr := config.Config.Labels["org.opencontainers.image.revision"]

	epoch := 0
	if epochStr != "" {
		_, _ = fmt.Sscanf(epochStr, "%d", &epoch)
	}

	return &Image{
		Name:    imageName,
		Version: version,
		Epoch:   epoch,
		Exists:  true,
	}, nil
}

// PullSBOM pulls the SBOM artifact from the registry using ORAS Go library
func PullSBOM(name, registry string) (map[string]interface{}, error) {
	ctx := context.Background()

	repo, err := setupRepository(registry, name)
	if err != nil {
		return nil, err
	}

	// Resolve the manifest for the latest tag to get its digest
	tag := "latest"
	manifestDescriptor, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve manifest: %w", err)
	}

	// Fetch referrers (artifacts attached to this image)
	var sbomDescriptor *ocispec.Descriptor
	err = repo.Referrers(ctx, manifestDescriptor, "", func(referrers []ocispec.Descriptor) error {
		// Find the SBOM artifact (application/spdx+json)
		for _, desc := range referrers {
			if desc.ArtifactType == "application/spdx+json" {
				sbomDescriptor = &desc
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get referrers: %w", err)
	}

	if sbomDescriptor == nil {
		return nil, fmt.Errorf("no SBOM artifact found attached to image")
	}

	// Fetch the manifest to get the layers
	manifestData, err := repo.Fetch(ctx, *sbomDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SBOM manifest: %w", err)
	}
	defer func() {
		if err := manifestData.Close(); err != nil {
			log.Printf("Warning: failed to close manifest data: %v", err)
		}
	}()

	// Parse the manifest
	var manifest ocispec.Manifest
	manifestBytes, err := io.ReadAll(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// The SBOM content should be in the first layer
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("no layers found in SBOM manifest")
	}

	// Fetch the SBOM layer content
	sbomLayer := manifest.Layers[0]
	sbomContent, err := repo.Fetch(ctx, sbomLayer)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SBOM content: %w", err)
	}
	defer func() {
		if err := sbomContent.Close(); err != nil {
			log.Printf("Warning: failed to close SBOM content: %v", err)
		}
	}()

	// Read and parse the SBOM
	sbomData, err := io.ReadAll(sbomContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM content: %w", err)
	}

	var sbom map[string]interface{}
	if err := json.Unmarshal(sbomData, &sbom); err != nil {
		return nil, fmt.Errorf("failed to parse SBOM JSON: %w", err)
	}

	return sbom, nil
}

// setupRepository creates and configures a remote repository with Docker credentials
func setupRepository(registry, name string) (*remote.Repository, error) {
	// Setup remote repository
	repo, err := remote.NewRepository(registry + "/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create credential store from Docker config
	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store: %w", err)
	}

	// Configure authentication using Docker credential store
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(store),
	}

	return repo, nil
}

// FindPackageInSBOM searches for a package in the SPDX SBOM and returns its version
func FindPackageInSBOM(sbom map[string]interface{}, packageName string) (string, bool) {
	packages, ok := sbom["packages"].([]interface{})
	if !ok {
		return "", false
	}

	for _, pkg := range packages {
		pkgMap, ok := pkg.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := pkgMap["name"].(string)
		if !ok || name != packageName {
			continue
		}

		version, ok := pkgMap["versionInfo"].(string)
		if !ok {
			continue
		}

		return version, true
	}

	return "", false
}

// CheckContainerNeedsBuild determines if a container needs to be built based on registry SBOM
func CheckContainerNeedsBuild(name string, packageSpecs map[string]*spec.PackageSpec, containerSpec *spec.ContainerSpec, registry string) (bool, error) {
	// Query the registry for the current image
	registryImage, err := QueryImage(registry, name)
	if err != nil {
		log.Printf("Warning: failed to query registry for %s: %v", name, err)
		// If we can't query the registry, assume we need to build
		return true, nil
	}

	if !registryImage.Exists {
		return true, nil
	}

	// Try to pull and compare SBOM from registry
	registrySBOM, err := PullSBOM(name, registry)
	if err != nil {
		log.Printf("Warning: failed to pull SBOM for %s: %v (assuming needs rebuild)", name, err)
		return true, nil
	}

	// Compare package versions in SBOM with current package specs
	for _, pkgName := range containerSpec.Contents.Packages {
		if s, ok := packageSpecs[pkgName]; ok {
			currentVersion := fmt.Sprintf("%s-r%d", s.Package.Version, s.Package.Epoch)

			// Find this package in the SBOM
			sbomVersion, found := FindPackageInSBOM(registrySBOM, pkgName)
			if !found {
				log.Printf("Container %s needs rebuild: package %s not found in SBOM", name, pkgName)
				return true, nil
			}

			if currentVersion != sbomVersion {
				log.Printf("Container %s needs rebuild: package %s changed (%s -> %s)",
					name, pkgName, sbomVersion, currentVersion)
				return true, nil
			}
		}
	}

	return false, nil
}

// AttachSBOM attaches the SBOM file to the container image in the registry using ORAS Go library
func AttachSBOM(name, registry, outputDir string) error {
	ctx := context.Background()

	// Find SBOM file - apko generates files with architecture-specific names
	// e.g., centauri-x86_64.spdx.json, centauri-aarch64.spdx.json
	pattern := filepath.Join(outputDir, fmt.Sprintf("%s-*.spdx.json", name))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to search for SBOM files: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no SBOM files found matching pattern: %s", pattern)
	}

	// Use the first match (typically there will only be one for single-arch builds)
	sbomFile := matches[0]

	log.Printf("Attaching SBOM to %s/%s:latest", registry, name)

	repo, err := setupRepository(registry, name)
	if err != nil {
		return err
	}

	// Resolve the manifest for the latest tag to get the subject descriptor
	tag := "latest"
	subjectDescriptor, err := repo.Resolve(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to resolve image manifest: %w", err)
	}

	// Read SBOM data
	sbomData, err := os.ReadFile(sbomFile)
	if err != nil {
		return fmt.Errorf("failed to read SBOM: %w", err)
	}

	// Push the blob and get its descriptor
	blobDesc, err := oras.PushBytes(ctx, repo, "application/spdx+json", sbomData)
	if err != nil {
		return fmt.Errorf("failed to push SBOM blob: %w", err)
	}

	// Pack and push the artifact manifest with the subject reference
	artifactType := "application/spdx+json"
	opts := oras.PackManifestOptions{
		Layers:  []ocispec.Descriptor{blobDesc},
		Subject: &subjectDescriptor,
	}

	manifestDesc, err := oras.PackManifest(ctx, repo, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return fmt.Errorf("failed to pack and push manifest: %w", err)
	}

	log.Printf("Successfully attached SBOM (digest: %s)", manifestDesc.Digest)
	return nil
}

// CleanupSBOMs renames and cleans up SBOM files
func CleanupSBOMs(name, outputDir string) error {
	sbomPattern := filepath.Join(outputDir, "sbom-*.spdx.json")
	matches, err := filepath.Glob(sbomPattern)
	if err != nil {
		return err
	}

	for _, oldPath := range matches {
		sbomFile := filepath.Base(oldPath)

		if sbomFile == "sbom-index.spdx.json" {
			if err := os.Remove(oldPath); err != nil {
				return err
			}
			continue
		}

		suffix := strings.TrimPrefix(sbomFile, "sbom-")
		newPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s", name, suffix))
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
}
