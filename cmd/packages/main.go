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
}

type Graph struct {
	nodes map[string][]string
	all   map[string]bool
	dir   string
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string][]string),
		all:   make(map[string]bool),
	}
}

func (g *Graph) AddPackage(name string, deps []string) {
	g.all[name] = true
	g.nodes[name] = deps
	for _, dep := range deps {
		g.all[dep] = true
	}
}

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
			return fmt.Errorf("circular dependency detected involving package: %s", node)
		}

		inProgress[node] = true

		if deps, exists := g.nodes[node]; exists {
			for _, dep := range deps {
				if _, isPackage := g.nodes[dep]; isPackage {
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

	nodes := make([]string, 0, len(g.nodes))
	for node := range g.nodes {
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

func (g *Graph) GetLevels() ([][]string, error) {
	if _, err := g.TopologicalSort(); err != nil {
		return nil, err
	}

	levels := make(map[string]int)

	var getLevel func(string) int
	getLevel = func(node string) int {
		if level, exists := levels[node]; exists {
			return level
		}

		maxDepLevel := -1
		if deps, exists := g.nodes[node]; exists {
			for _, dep := range deps {
				if _, isPackage := g.nodes[dep]; isPackage {
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

	for node := range g.nodes {
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

// getDependencies returns a list of packages to build for a specific package,
// including all its dependencies in topological order
func (g *Graph) getDependencies(packageName string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(string)
	visit = func(node string) {
		if visited[node] {
			return
		}

		// Visit dependencies first
		if deps, exists := g.nodes[node]; exists {
			for _, dep := range deps {
				if _, isPackage := g.nodes[dep]; isPackage {
					visit(dep)
				}
			}
		}

		visited[node] = true
		result = append(result, node)
	}

	visit(packageName)
	return result
}

func (g *Graph) LoadFromDirectory(dir string) error {
	g.dir = dir

	// Find all YAML files
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read packages directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no YAML files found in %s directory", dir)
	}

	// First pass: collect all package names
	packageNames := make(map[string]bool)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", file, err)
			continue
		}

		var spec PackageSpec
		if err = yaml.Unmarshal(data, &spec); err != nil {
			log.Printf("Warning: failed to parse %s: %v", file, err)
			continue
		}

		packageNames[spec.Package.Name] = true
	}

	// Second pass: build dependency graph
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var spec PackageSpec
		if err = yaml.Unmarshal(data, &spec); err != nil {
			continue
		}

		var deps []string
		for _, pkg := range spec.Environment.Contents.Packages {
			if packageNames[pkg] && pkg != spec.Package.Name {
				deps = append(deps, pkg)
			}
		}

		g.AddPackage(spec.Package.Name, deps)
	}

	return nil
}

// cleanupBubblewrapDirs removes stale bubblewrap-guest* directories from /tmp
// Uses chmod -R to fix permissions before removal since bubblewrap creates
// directories with restricted permissions
func cleanupBubblewrapDirs() error {
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

func (g *Graph) PrintTree() error {
	levels, err := g.GetLevels()
	if err != nil {
		return err
	}

	rootPkgs := levels[0]
	for i, pkg := range rootPkgs {
		isLastRoot := i == len(rootPkgs)-1

		var dependents []string
		for p, deps := range g.nodes {
			for _, dep := range deps {
				if dep == pkg {
					dependents = append(dependents, p)
					break
				}
			}
		}
		sort.Strings(dependents)

		if isLastRoot {
			fmt.Printf("└─ %s\n", pkg)
		} else if i == 0 {
			fmt.Printf("┌─ %s\n", pkg)
		} else {
			fmt.Printf("├─ %s\n", pkg)
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

func (g *Graph) BuildPackages(force bool, forceAll bool, packageName string) error {
	buildOrder, err := g.TopologicalSort()
	if err != nil {
		return err
	}

	// If a specific package is requested, filter build order to include only that package and its dependencies
	if packageName != "" {
		if _, exists := g.nodes[packageName]; !exists {
			return fmt.Errorf("package %s not found", packageName)
		}
		buildOrder = g.getDependencies(packageName)
	}

	// Clean up any stale bubblewrap directories before starting
	log.Println("Cleaning up stale bubblewrap directories before build...")
	if err := cleanupBubblewrapDirs(); err != nil {
		log.Printf("Warning: cleanup failed: %v", err)
	}

	for _, pkg := range buildOrder {
		yamlFile := filepath.Join(g.dir, fmt.Sprintf("%s.yaml", pkg))

		// Determine if we should force rebuild this specific package
		shouldForce := forceAll || (force && packageName != "" && pkg == packageName)

		// Check if package already exists unless force flag is set
		if !shouldForce {
			data, err := os.ReadFile(yamlFile)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", yamlFile, err)
			}

			var spec PackageSpec
			if err = yaml.Unmarshal(data, &spec); err != nil {
				return fmt.Errorf("failed to parse %s: %w", yamlFile, err)
			}

			// Check for each target architecture
			archList := spec.Package.TargetArch
			if len(archList) == 0 {
				// Default to x86_64 if not specified
				archList = []string{"x86_64"}
			}

			allExist := true
			for _, arch := range archList {
				apkFile := filepath.Join("./repo", arch,
					fmt.Sprintf("%s-%s-r%d.apk", spec.Package.Name, spec.Package.Version, spec.Package.Epoch))
				if _, err := os.Stat(apkFile); os.IsNotExist(err) {
					allExist = false
					break
				}
			}

			if allExist {
				fmt.Printf("Skipping %s (already built)\n", pkg)
				continue
			}
		}

		fmt.Printf("Building package: %s\n", pkg)
		cmd := exec.Command("melange", "build", yamlFile,
			"--signing-key", "melange.rsa",
			"--keyring-append", "melange.rsa.pub",
			"--out-dir", "./repo",
			"--repository-append", "./repo",
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Clean up bubblewrap dirs after a failed build
			cleanupBubblewrapDirs()
			return fmt.Errorf("failed to build package %s: %w", pkg, err)
		}

		fmt.Printf("Successfully built: %s\n\n", pkg)

		// Clean up after each successful build to prevent accumulation
		cleanupBubblewrapDirs()
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
		fmt.Fprintf(os.Stderr, "Unknown packages subcommand: %s\n\n", subcommand)
		printHelp()
		os.Exit(1)
	}
}

func buildCommand() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	force := fs.Bool("force", false, "force rebuild of specified package only (when package is specified)")
	forceAll := fs.Bool("forceall", false, "force rebuild of specified package and all its dependencies")
	fs.Parse(os.Args[2:])

	// Get optional package name from remaining args
	var packageName string
	if fs.NArg() > 0 {
		packageName = fs.Arg(0)
	}

	// When no package is specified, -force behaves like -forceall (rebuild everything)
	if packageName == "" && *force {
		*forceAll = true
		*force = false
	}

	graph := NewGraph()

	err := graph.LoadFromDirectory("packages")
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	err = graph.BuildPackages(*force, *forceAll, packageName)
	if err != nil {
		log.Fatalf("Failed to build packages: %v", err)
	}
}

func graphCommand() {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	graph := NewGraph()

	err := graph.LoadFromDirectory("packages")
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	err = graph.PrintTree()
	if err != nil {
		log.Fatalf("Failed to print dependency graph: %v", err)
	}
}

func printHelp() {
	fmt.Println("Usage: packages <command> [options] [package]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [package]    Build packages in dependency order")
	fmt.Println("                     If package is specified, only build that package and its dependencies")
	fmt.Println("  graph              Display package dependency tree")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Build options:")
	fmt.Println("  -force             When package specified: force rebuild of only that package")
	fmt.Println("                     When no package specified: force rebuild of all packages")
	fmt.Println("  -forceall          Force rebuild of specified package and all its dependencies")
	fmt.Println()
}
