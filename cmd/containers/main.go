package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type ContainerSpec struct {
	Contents struct {
		Packages []string `yaml:"packages"`
	} `yaml:"contents"`
}

type ContainerGraph struct {
	nodes map[string][]string
	all   map[string]bool
	dir   string
}

func NewContainerGraph() *ContainerGraph {
	return &ContainerGraph{
		nodes: make(map[string][]string),
		all:   make(map[string]bool),
	}
}

func (cg *ContainerGraph) AddContainer(name string, deps []string) {
	cg.all[name] = true
	cg.nodes[name] = deps
	for _, dep := range deps {
		cg.all[dep] = true
	}
}

func (cg *ContainerGraph) TopologicalSort() ([]string, error) {
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(node string) error {
		if visited[node] {
			return nil
		}
		if inProgress[node] {
			return fmt.Errorf("circular dependency detected involving container: %s", node)
		}

		inProgress[node] = true

		if deps, exists := cg.nodes[node]; exists {
			for _, dep := range deps {
				if _, isContainer := cg.nodes[dep]; isContainer {
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

	nodes := make([]string, 0, len(cg.nodes))
	for node := range cg.nodes {
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

func (cg *ContainerGraph) GetLevels() ([][]string, error) {
	if _, err := cg.TopologicalSort(); err != nil {
		return nil, err
	}

	levels := make(map[string]int)

	var getLevel func(string) int
	getLevel = func(node string) int {
		if level, exists := levels[node]; exists {
			return level
		}

		maxDepLevel := -1
		if deps, exists := cg.nodes[node]; exists {
			for _, dep := range deps {
				if _, isContainer := cg.nodes[dep]; isContainer {
					depLevel := getLevel(dep)
					if depLevel > maxDepLevel {
						maxDepLevel = depLevel
					}
				}
			}
		}

		level := maxDepLevel + 1
		levels[node] = level
		return level
	}

	for node := range cg.nodes {
		getLevel(node)
	}

	maxLevel := 0
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	result := make([][]string, maxLevel+1)
	for node, level := range levels {
		result[level] = append(result[level], node)
	}

	for i := range result {
		sort.Strings(result[i])
	}

	return result, nil
}

func (cg *ContainerGraph) getDependencies(containerName string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(string)
	visit = func(node string) {
		if visited[node] {
			return
		}

		if deps, exists := cg.nodes[node]; exists {
			for _, dep := range deps {
				if _, isContainer := cg.nodes[dep]; isContainer {
					visit(dep)
				}
			}
		}

		visited[node] = true
		result = append(result, node)
	}

	visit(containerName)
	return result
}

func (cg *ContainerGraph) LoadFromDirectory(dir string) error {
	cg.dir = dir

	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read containers directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no YAML files found in %s directory", dir)
	}

	containerNames := make(map[string]bool)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		containerNames[name] = true
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", file, err)
			continue
		}

		var spec ContainerSpec
		if err = yaml.Unmarshal(data, &spec); err != nil {
			log.Printf("Warning: failed to parse %s: %v", file, err)
			continue
		}

		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		var deps []string
		for _, pkg := range spec.Contents.Packages {
			if containerNames[pkg] && pkg != name {
				deps = append(deps, pkg)
			}
		}

		cg.AddContainer(name, deps)
	}

	return nil
}

func (cg *ContainerGraph) PrintTree() error {
	levels, err := cg.GetLevels()
	if err != nil {
		return err
	}

	rootContainers := levels[0]
	for i, container := range rootContainers {
		isLastRoot := i == len(rootContainers)-1

		var dependents []string
		for c, deps := range cg.nodes {
			for _, dep := range deps {
				if dep == container {
					dependents = append(dependents, c)
					break
				}
			}
		}
		sort.Strings(dependents)

		if isLastRoot {
			fmt.Printf("└─ %s\n", container)
		} else if i == 0 {
			fmt.Printf("┌─ %s\n", container)
		} else {
			fmt.Printf("├─ %s\n", container)
		}

		prefix := "    "
		if !isLastRoot {
			prefix = "│   "
		}

		for j, dep := range dependents {
			if j == len(dependents)-1 {
				fmt.Printf("%s└─ %s\n", prefix, dep)
			} else {
				fmt.Printf("%s├─ %s\n", prefix, dep)
			}
		}
	}

	return nil
}

func (cg *ContainerGraph) BuildContainers(force bool, forceAll bool, containerName string) error {
	buildOrder, err := cg.TopologicalSort()
	if err != nil {
		return err
	}

	if containerName != "" {
		if _, exists := cg.nodes[containerName]; !exists {
			return fmt.Errorf("container %s not found", containerName)
		}
		buildOrder = cg.getDependencies(containerName)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, name := range buildOrder {
		yamlFile := filepath.Join(cg.dir, fmt.Sprintf("%s.yaml", name))

		shouldForce := forceAll || (force && containerName != "" && name == containerName)

		if !shouldForce {
			tarFile := fmt.Sprintf("%s.tar.gz", name)
			if _, err := os.Stat(tarFile); err == nil {
				fmt.Printf("Skipping %s (already built)\n", name)
				continue
			}
		}

		fmt.Printf("Building container: %s\n", name)
		cmd := exec.Command("apko", "build",
			"--repository-append", "repo",
			"--keyring-append", "melange.rsa.pub",
			"--arch", "amd64",
			"--sbom-path", "output",
			yamlFile,
			name,
			filepath.Join("output", fmt.Sprintf("%s.tar.gz", name)),
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build container %s: %w", name, err)
		}

		fmt.Printf("Successfully built: %s\n\n", name)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "build":
		buildCommand()
	case "graph":
		graphCommand()
	case "help", "-h", "--help":
		printHelp()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown containers subcommand: %s\n\n", subcommand)
		printHelp()
		os.Exit(1)
	}
}

func buildCommand() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	force := fs.Bool("force", false, "force rebuild of specified container only (when container is specified)")
	forceAll := fs.Bool("forceall", false, "force rebuild of specified container and all its dependencies")
	fs.Parse(os.Args[2:])

	var containerName string
	if fs.NArg() > 0 {
		containerName = fs.Arg(0)
	}

	if containerName == "" && *force {
		*forceAll = true
		*force = false
	}

	graph := NewContainerGraph()

	err := graph.LoadFromDirectory("containers")
	if err != nil {
		log.Fatalf("Failed to load containers: %v", err)
	}

	err = graph.BuildContainers(*force, *forceAll, containerName)
	if err != nil {
		log.Fatalf("Failed to build containers: %v", err)
	}
}

func graphCommand() {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	graph := NewContainerGraph()

	err := graph.LoadFromDirectory("containers")
	if err != nil {
		log.Fatalf("Failed to load containers: %v", err)
	}

	err = graph.PrintTree()
	if err != nil {
		log.Fatalf("Failed to print dependency graph: %v", err)
	}
}

func printHelp() {
	fmt.Println("Usage: containers <command> [options] [container]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [container]  Build containers in dependency order")
	fmt.Println("                     If container is specified, only build that container and its dependencies")
	fmt.Println("  graph              Display container dependency tree")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Build options:")
	fmt.Println("  -force             When container specified: force rebuild of only that container")
	fmt.Println("                     When no container specified: force rebuild of all containers")
	fmt.Println("  -forceall          Force rebuild of specified container and all its dependencies")
	fmt.Println()
}
