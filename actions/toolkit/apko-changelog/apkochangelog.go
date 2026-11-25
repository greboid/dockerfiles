// Package apkochangelog compares two apko lock files and produces a
// commit-message style changelog. It lists every package change per
// architecture — version bumps, additions and removals, in a stable
// sorted order — and reports repository changes separately as
// non-package changes. The result is exposed as the message step
// output, with changes set to true when any package changed.
package apkochangelog

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

type packageInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
}

type repositoryInfo struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Architecture string `json:"architecture"`
}

type lockFile struct {
	Contents struct {
		Repositories []repositoryInfo `json:"repositories"`
		Packages     []packageInfo    `json:"packages"`
	} `json:"contents"`
}

// Inputs carries the action inputs.
type Inputs struct {
	NewLockfile string
	OldLockfile string
	Version     string
}

// Run compares the lock files and reports the changelog.
func Run(c *gha.Context, log *slog.Logger, in Inputs) error {
	newFile := in.NewLockfile
	if newFile == "" {
		newFile = "base.lock.json"
	}
	oldFile := in.OldLockfile
	if oldFile == "" {
		oldFile = "old.lock.json"
	}
	if !filepath.IsAbs(newFile) {
		newFile = filepath.Join(c.Workspace, newFile)
	}
	if !filepath.IsAbs(oldFile) {
		oldFile = filepath.Join(c.Workspace, oldFile)
	}

	version := gha.FirstNonEmpty(in.Version, os.Getenv("VERSTRING"))

	newLock, err := readLockFile(newFile)
	if err != nil {
		return err
	}
	oldLock, err := readLockFile(oldFile)
	if err != nil {
		return err
	}

	packageChanges := formatPackageChanges(oldLock.Contents.Packages, newLock.Contents.Packages)
	repositoryChanges := formatRepositoryChanges(oldLock.Contents.Repositories, newLock.Contents.Repositories)

	header := "Updating"
	if version != "" {
		header += fmt.Sprintf(" to %s", version)
	}
	var message strings.Builder
	if packageChanges != "" {
		message.WriteString(fmt.Sprintf("%s with the following package changes:\n%s", header, packageChanges))
	} else {
		message.WriteString(fmt.Sprintf("%s with no package changes.\n", header))
	}
	if repositoryChanges != "" {
		message.WriteString("\nNon-package changes:\n")
		message.WriteString(repositoryChanges)
	}

	log.Info(message.String())

	if err := c.Output("message", message.String()); err != nil {
		return err
	}
	return c.Output("changes", boolOutput(packageChanges != ""))
}

func boolOutput(b bool) string {
	if b {
		return "true"
	}
	return ""
}

func readLockFile(filename string) (*lockFile, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	var lock lockFile
	if err := json.Unmarshal(fileBytes, &lock); err != nil {
		return nil, fmt.Errorf("unable to parse file %s: %w", filename, err)
	}
	return &lock, nil
}

func formatPackageChanges(oldPkgs, newPkgs []packageInfo) string {
	oldGrouped := groupVersions(oldPkgs)
	newGrouped := groupVersions(newPkgs)

	archs := map[string]bool{}
	for arch := range oldGrouped {
		archs[arch] = true
	}
	for arch := range newGrouped {
		archs[arch] = true
	}
	sortedArchs := make([]string, 0, len(archs))
	for arch := range archs {
		sortedArchs = append(sortedArchs, arch)
	}
	sort.Strings(sortedArchs)

	var out strings.Builder
	for _, arch := range sortedArchs {
		lines := archChangeLines(oldGrouped[arch], newGrouped[arch])
		if len(lines) == 0 {
			continue
		}
		out.WriteString(fmt.Sprintf(" - %s\n", arch))
		for _, line := range lines {
			out.WriteString(fmt.Sprintf("   - %s\n", line))
		}
	}
	return out.String()
}

func groupVersions(pkgs []packageInfo) map[string]map[string]string {
	grouped := map[string]map[string]string{}
	for _, pkg := range pkgs {
		if grouped[pkg.Architecture] == nil {
			grouped[pkg.Architecture] = map[string]string{}
		}
		grouped[pkg.Architecture][pkg.Name] = pkg.Version
	}
	return grouped
}

func archChangeLines(oldVersions, newVersions map[string]string) []string {
	names := map[string]bool{}
	for name := range oldVersions {
		names[name] = true
	}
	for name := range newVersions {
		names[name] = true
	}
	var lines []string
	for name := range names {
		oldVersion, inOld := oldVersions[name]
		newVersion, inNew := newVersions[name]
		switch {
		case inOld && inNew && oldVersion != newVersion:
			lines = append(lines, fmt.Sprintf("%s: %s => %s", name, oldVersion, newVersion))
		case !inOld && inNew:
			lines = append(lines, fmt.Sprintf("%s: %s (added)", name, newVersion))
		case inOld && !inNew:
			lines = append(lines, fmt.Sprintf("%s: %s (removed)", name, oldVersion))
		}
	}
	sort.Strings(lines)
	return lines
}

func formatRepositoryChanges(oldRepos, newRepos []repositoryInfo) string {
	oldByName := map[string]string{}
	for _, repo := range oldRepos {
		oldByName[repo.Name] = repo.URL
	}
	newByName := map[string]string{}
	for _, repo := range newRepos {
		newByName[repo.Name] = repo.URL
	}

	names := map[string]bool{}
	for name := range oldByName {
		names[name] = true
	}
	for name := range newByName {
		names[name] = true
	}
	var lines []string
	for name := range names {
		oldURL, inOld := oldByName[name]
		newURL, inNew := newByName[name]
		switch {
		case inOld && inNew && oldURL != newURL:
			lines = append(lines, fmt.Sprintf("repository: %s: %s => %s", name, oldURL, newURL))
		case !inOld && inNew:
			lines = append(lines, fmt.Sprintf("repository added: %s", name))
		case inOld && !inNew:
			lines = append(lines, fmt.Sprintf("repository removed: %s", name))
		}
	}
	sort.Strings(lines)

	var out strings.Builder
	for _, line := range lines {
		out.WriteString(fmt.Sprintf(" - %s\n", line))
	}
	return out.String()
}
