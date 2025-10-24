package graph

import (
	"fmt"
	"sort"
)

// Graph represents a dependency graph structure
type Graph struct {
	Nodes map[string][]string
}

// New creates a new Graph instance
func New() *Graph {
	return &Graph{
		Nodes: make(map[string][]string),
	}
}

// AddNode adds a node with its dependencies to the graph
func (g *Graph) AddNode(name string, deps []string) {
	g.Nodes[name] = deps
}

// TopologicalSort returns a topologically sorted list of nodes
func (g *Graph) TopologicalSort() ([]string, error) {
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(node string) error {
		if visited[node] {
			return nil
		}
		if inProgress[node] {
			return fmt.Errorf("circular dependency detected involving: %s", node)
		}

		inProgress[node] = true

		if deps, exists := g.Nodes[node]; exists {
			for _, dep := range deps {
				if _, isNode := g.Nodes[dep]; isNode {
					if err := visit(dep); err != nil {
						return err
					}
				}
			}
		}

		inProgress[node] = false
		visited[node] = true
		result = append(result, node)
		return nil
	}

	nodes := make([]string, 0, len(g.Nodes))
	for node := range g.Nodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetDependencies returns a list of nodes to process for a specific node,
// including all its dependencies in topological order
func (g *Graph) GetDependencies(nodeName string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(string)
	visit = func(node string) {
		if visited[node] {
			return
		}

		if deps, exists := g.Nodes[node]; exists {
			for _, dep := range deps {
				if _, isNode := g.Nodes[dep]; isNode {
					visit(dep)
				}
			}
		}

		visited[node] = true
		result = append(result, node)
	}

	visit(nodeName)
	return result
}

// GetAllDependencies returns all dependencies (direct and transitive) for a given node
func (g *Graph) GetAllDependencies(node string) map[string]bool {
	deps := make(map[string]bool)

	var visit func(string)
	visit = func(n string) {
		if deps[n] {
			return
		}
		deps[n] = true

		if nodeDeps, exists := g.Nodes[n]; exists {
			for _, dep := range nodeDeps {
				if _, isDep := g.Nodes[dep]; isDep {
					visit(dep)
				}
			}
		}
	}

	if nodeDeps, exists := g.Nodes[node]; exists {
		for _, dep := range nodeDeps {
			if _, isDep := g.Nodes[dep]; isDep {
				visit(dep)
			}
		}
	}

	return deps
}
