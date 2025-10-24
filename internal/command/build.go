package command

import (
	"fmt"
	"log"

	"package-order/internal/graph"
)

// BuildConfig holds configuration for building items (packages or containers)
type BuildConfig struct {
	ItemName   string
	ItemType   string // "package" or "container" (for logging)
	Force      bool
	ForceAll   bool
	SourceDir  string
	BuildOrder []string
	Graph      *graph.Graph
}

// CheckFunc returns true if an item needs to be built
type CheckFunc func(name string) bool

// BuildFunc builds a single item
type BuildFunc func(name string) error

// CleanupFunc performs cleanup after build (may be nil)
type CleanupFunc func(name string) error

// ExecuteBuild executes a build process for items in dependency order
func ExecuteBuild(config BuildConfig, checkNeedsBuild CheckFunc, buildItem BuildFunc, cleanup CleanupFunc) error {
	for _, name := range config.BuildOrder {
		shouldForce := config.ForceAll || (config.Force && config.ItemName != "" && name == config.ItemName)

		if !shouldForce && !checkNeedsBuild(name) {
			fmt.Printf("Skipping %s (already built)\n", name)
			continue
		}

		fmt.Printf("Building %s: %s\n", config.ItemType, name)
		if err := buildItem(name); err != nil {
			return fmt.Errorf("failed to build %s %s: %w", config.ItemType, name, err)
		}

		if cleanup != nil {
			if err := cleanup(name); err != nil {
				log.Printf("Warning: cleanup failed for %s: %v", name, err)
			}
		}

		fmt.Printf("Successfully built: %s\n\n", name)
	}

	return nil
}

// ResolveBuildOrder determines what items to build based on the configuration
func ResolveBuildOrder(g *graph.Graph, itemName string) ([]string, error) {
	buildOrder, err := g.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to sort items: %w", err)
	}

	if itemName != "" {
		if _, exists := g.Nodes[itemName]; !exists {
			return nil, fmt.Errorf("item %s not found", itemName)
		}
		buildOrder = g.GetDependencies(itemName)
	}

	return buildOrder, nil
}
