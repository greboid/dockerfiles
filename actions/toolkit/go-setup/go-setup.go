// Package gosetup is a pure-Go replacement for actions/setup-go: it
// downloads a Go toolchain from go.dev, verifies its checksum, unpacks
// it inside the workspace and exports GOROOT/PATH for subsequent
// steps. When the runner advertises a cache (ACTIONS_CACHE_URL, the
// standard protocol actions/cache uses) toolchain archives are stored
// and restored through it, so repeat jobs skip the go.dev download.
package gosetup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/cache"
	"git.mouse-lake.ts.net/actions/toolkit/internal/dl"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/tarx"
	"git.mouse-lake.ts.net/actions/toolkit/internal/ver"
)

const downloadListURL = "https://go.dev/dl/?mode=json&include=all"

type goFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHA256   string `json:"sha256"`
}

type goRelease struct {
	Version string   `json:"version"` // "go1.22.5"
	Stable  bool     `json:"stable"`
	Files   []goFile `json:"files"`
}

// Inputs carries the action inputs.
type Inputs struct {
	GoVersion     string
	GoVersionFile string
	InstallDir    string
	Cache         string
	Force         string
}

// Run installs the requested Go toolchain into the workspace.
func Run(c *gha.Context, log *slog.Logger, in Inputs) error {
	ws := c.Workspace

	versionInput := in.GoVersion
	fileInput := in.GoVersionFile
	if versionInput != "" && fileInput != "" {
		return fmt.Errorf("go-version and go-version-file are mutually exclusive")
	}
	if versionInput == "" && fileInput == "" {
		if _, err := os.Stat(filepath.Join(ws, "go.mod")); err == nil {
			fileInput = "go.mod"
			log.Info("no go-version given; deriving from go.mod")
		}
	}
	if fileInput != "" {
		path := fileInput
		if !filepath.IsAbs(path) {
			path = filepath.Join(ws, path)
		}
		v, err := goVersionFromMod(path)
		if err != nil {
			return err
		}
		versionInput = v
		log.Info(fmt.Sprintf("go.mod requires %s", v))
	}
	if versionInput == "" {
		return fmt.Errorf("go-version, go-version-file or a go.mod in the workspace is required")
	}

	var releases []goRelease
	if err := dl.FetchJSON(downloadListURL, &releases); err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	release, file, err := resolveGo(releases, versionInput)
	if err != nil {
		return err
	}
	goVersion := strings.TrimPrefix(release.Version, "go")

	installRoot := in.InstallDir
	if installRoot == "" {
		installRoot = filepath.Join(ws, ".actions", "go")
	}
	dest := filepath.Join(installRoot, "go"+goVersion)
	binDir := filepath.Join(dest, "bin")
	goBin := filepath.Join(binDir, "go")

	force := gha.Truthy(in.Force, false)
	if !force {
		if _, err := os.Stat(goBin); err == nil {
			log.Info(fmt.Sprintf("reusing cached toolchain %s", goVersion))
			return export(c, binDir, dest, goVersion, log)
		}
	}

	// The cache key carries the archive filename (version, OS, arch)
	// and the entry version is its sha256, so a hit can only ever hand
	// back the exact bytes go.dev serves for that release.
	cacheKey := "go-setup-" + file.Filename
	var cacheClient *cache.Client
	if gha.Truthy(in.Cache, true) {
		cacheClient = cache.NewFromEnv(log)
		if cacheClient == nil {
			log.Info("no runner cache advertised; downloading from go.dev")
		}
	}

	log.Info(fmt.Sprintf("installing Go %s (%s)", goVersion, file.Filename))
	archive := filepath.Join(installRoot, file.Filename)
	fromCache := false
	if cacheClient != nil {
		location, ok, err := cacheClient.Lookup(cacheKey, file.SHA256)
		switch {
		case err != nil:
			log.Info(fmt.Sprintf("cache lookup failed, falling back to go.dev: %v", err))
		case !ok:
			log.Info(fmt.Sprintf("cache miss for %s", file.Filename))
		default:
			if err := cacheClient.Download(location, archive, file.SHA256); err != nil {
				log.Info(fmt.Sprintf("cache download failed, falling back to go.dev: %v", err))
			} else {
				fromCache = true
			}
		}
	}
	if !fromCache {
		if err := dl.DownloadFile("https://go.dev/dl/"+file.Filename, archive, file.SHA256, log); err != nil {
			return err
		}
	}
	tmp, err := os.MkdirTemp(installRoot, ".extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	err = tarx.ExtractTarGz(f, tmp, 1)
	f.Close()
	if err != nil {
		return fmt.Errorf("extract %s: %w", file.Filename, err)
	}
	os.RemoveAll(dest)
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	if cacheClient != nil && !fromCache {
		// Best effort: a failed save must not fail the job.
		if err := cacheClient.Save(cacheKey, file.SHA256, archive); err != nil {
			log.Info(fmt.Sprintf("cache save failed (continuing): %v", err))
		} else {
			log.Info(fmt.Sprintf("cached %s for future jobs", file.Filename))
		}
	}
	os.Remove(archive)

	return export(c, binDir, dest, goVersion, log)
}

func export(c *gha.Context, binDir, goroot, version string, log *slog.Logger) error {
	if err := c.AddPath(binDir); err != nil {
		return err
	}
	if err := c.ExportEnv("GOROOT", goroot); err != nil {
		return err
	}
	if err := c.Output("go-version", version); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Go %s ready at %s", version, goroot))
	return nil
}

// resolveGo picks the release matching pattern and the linux archive
// for the current architecture.
func resolveGo(releases []goRelease, pattern string) (*goRelease, *goFile, error) {
	var best *goRelease
	bestV := ver.Version{}
	p := strings.TrimSpace(pattern)
	p = strings.TrimPrefix(strings.TrimPrefix(p, "go"), "v")
	p = strings.TrimSuffix(p, ".x") // "1.x" and "1.22.x" are wildcards

	// "", "stable", "latest" and "1"/"1.x" all mean: newest stable release.
	majorOnly := p == "" || p == "stable" || p == "latest" || p == "1" || p == "1.x"
	var want ver.Version
	var wantErr error
	if !majorOnly {
		want, wantErr = ver.Parse(p)
		if wantErr != nil {
			return nil, nil, fmt.Errorf("invalid go-version %q: %w", pattern, wantErr)
		}
	}

	for i := range releases {
		r := &releases[i]
		v, err := ver.Parse(strings.TrimPrefix(r.Version, "go"))
		if err != nil {
			continue
		}
		if majorOnly {
			if r.Stable && (best == nil || ver.Compare(v, bestV) > 0) {
				best, bestV = r, v
			}
			continue
		}
		exact := strings.Count(p, ".") >= 2
		if exact {
			if ver.Compare(v, want) == 0 {
				best = r
				break
			}
			continue
		}
		matches := v.Major == want.Major &&
			(strings.Count(p, ".") == 0 || v.Minor == want.Minor)
		if matches && r.Stable && (best == nil || ver.Compare(v, bestV) > 0) {
			best, bestV = r, v
		}
	}
	if best == nil {
		return nil, nil, fmt.Errorf("no Go release found matching %q", pattern)
	}
	for i := range best.Files {
		fl := &best.Files[i]
		if fl.OS == "linux" && fl.Arch == runtime.GOARCH && strings.HasSuffix(fl.Filename, ".tar.gz") {
			return best, fl, nil
		}
	}
	return nil, nil, fmt.Errorf("Go %s has no linux/%s archive", best.Version, runtime.GOARCH)
}

// goVersionFromMod extracts the toolchain (preferred) or go directive
// from a go.mod file.
func goVersionFromMod(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	toolchain, goLine := "", ""
	for line := range strings.SplitSeq(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "toolchain":
			if toolchain == "" {
				toolchain = strings.TrimPrefix(fields[1], "go")
			}
		case "go":
			if goLine == "" {
				goLine = fields[1]
			}
		}
	}
	if toolchain != "" {
		return toolchain, nil
	}
	if goLine != "" {
		return goLine, nil
	}
	return "", fmt.Errorf("no go or toolchain directive found in %s", path)
}
