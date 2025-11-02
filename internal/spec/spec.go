package spec

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"package-order/internal/graph"
)

// PackageSpec represents a melange package specification
type PackageSpec struct {
	Package struct {
		Name       string   `yaml:"name"`
		Version    string   `yaml:"version"`
		Epoch      int      `yaml:"epoch"`
		TargetArch []string `yaml:"target-architecture"`
	} `yaml:"package"`
	Environment struct {
		Contents struct {
			Packages []string `yaml:"packages"`
		} `yaml:"contents"`
	} `yaml:"environment"`
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

// ContainerSpec represents an apko container specification
type ContainerSpec struct {
	Contents struct {
		Packages []string `yaml:"packages"`
	} `yaml:"contents"`
}

// readSpec is a generic function to read and parse a YAML file into any type
func readSpec[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec T
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// ReadPackageSpec reads and parses a package YAML file
func ReadPackageSpec(path string) (*PackageSpec, error) {
	return readSpec[PackageSpec](path)
}

// ReadContainerSpec reads and parses a container YAML file
func ReadContainerSpec(path string) (*ContainerSpec, error) {
	return readSpec[ContainerSpec](path)
}

// LoadPackageGraph builds a dependency graph for packages
func LoadPackageGraph(packagesDir string) (*graph.Graph, map[string]*PackageSpec, error) {
	files, err := filepath.Glob(filepath.Join(packagesDir, "*.yaml"))
	if err != nil {
		return nil, nil, err
	}

	g := graph.New()
	specs := make(map[string]*PackageSpec)

	// First pass: collect all package names
	packageNames := make(map[string]bool)
	for _, file := range files {
		spec, err := ReadPackageSpec(file)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", file, err)
			continue
		}
		packageNames[spec.Package.Name] = true
		specs[spec.Package.Name] = spec
	}

	// Second pass: build dependency graph
	for name, spec := range specs {
		var deps []string
		for _, pkg := range spec.Environment.Contents.Packages {
			if packageNames[pkg] && pkg != name {
				deps = append(deps, pkg)
			}
		}
		sort.Strings(deps)
		g.AddNode(name, deps)
	}

	return g, specs, nil
}

// LoadContainerGraph builds a dependency graph for containers
func LoadContainerGraph(containersDir string) (*graph.Graph, map[string]*ContainerSpec, error) {
	files, err := filepath.Glob(filepath.Join(containersDir, "*.yaml"))
	if err != nil {
		return nil, nil, err
	}

	g := graph.New()
	specs := make(map[string]*ContainerSpec)

	// First pass: collect all container names and specs
	containerNames := make(map[string]bool)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		containerNames[name] = true

		spec, err := ReadContainerSpec(file)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", file, err)
			continue
		}
		specs[name] = spec
	}

	// Second pass: build dependency graph
	for name, spec := range specs {
		var deps []string
		for _, pkg := range spec.Contents.Packages {
			// Only add dependencies on other containers, not packages
			if containerNames[pkg] && pkg != name {
				deps = append(deps, pkg)
			}
		}
		sort.Strings(deps)
		g.AddNode(name, deps)
	}

	return g, specs, nil
}

// CollectContainerPackageDeps collects all package dependencies for a container
// This includes both direct package dependencies from the container spec and
// transitive dependencies through the package dependency graph
func CollectContainerPackageDeps(containerSpec *ContainerSpec, packageGraph *graph.Graph, packageNames map[string]bool) map[string]bool {
	allPackageDeps := make(map[string]bool)

	// Get direct package dependencies from container spec
	for _, pkg := range containerSpec.Contents.Packages {
		if packageNames[pkg] {
			allPackageDeps[pkg] = true
			// Also add all transitive dependencies of this package
			transitiveDeps := packageGraph.GetAllDependencies(pkg)
			for dep := range transitiveDeps {
				allPackageDeps[dep] = true
			}
		}
	}

	return allPackageDeps
}

// ExtractRepositoryFromPipeline extracts the git repository URL from a package spec
func ExtractRepositoryFromPipeline(spec *PackageSpec) (string, error) {
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
