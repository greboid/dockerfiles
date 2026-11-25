// Package nodesetup is a pure-Go replacement for actions/setup-node:
// it downloads a Node.js distribution from nodejs.org, verifies its
// checksum, unpacks it inside the workspace and exports PATH for
// subsequent steps. On top of Node it sets up a package manager
// (pnpm by default, or yarn/npm via the package-manager input) and can
// write a project .npmrc for a given registry URL and token.
package nodesetup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/dl"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/tarx"
	"git.mouse-lake.ts.net/actions/toolkit/internal/ver"
)

const nodeDistURL = "https://nodejs.org/dist/"

// nodeUnofficialDistURL mirrors nodejs.org/dist with extra variants;
// the official builds are glibc-linked, so musl-based images (Alpine)
// need its linux-<arch>-musl tarballs instead.
const nodeUnofficialDistURL = "https://unofficial-builds.nodejs.org/download/release/"

// alpineCDN serves the Alpine package repository the musl builds'
// C++ runtime (libstdc++, libgcc) is bundled from.
const alpineCDN = "https://dl-cdn.alpinelinux.org/alpine/latest-stable/main/"

type nodeIndexEntry struct {
	Version string   `json:"version"` // "v18.19.0"
	Lts     any      `json:"lts"`     // false or a codename like "Jod"
	Files   []string `json:"files"`   // ["linux-x64", "osx-arm64-tar", ...]
}

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN).
type Inputs struct {
	NodeVersion     string
	NodeVersionFile string
	WorkingDir      string
	InstallDir      string
	Force           string
	Cache           string // accepted for setup-node familiarity; ignored
	PackageManager  string
	RegistryURL     string
	Token           string
}

// Run installs the requested Node.js toolchain into the workspace.
func Run(c *gha.Context, log *slog.Logger, in Inputs) error {
	if in.Cache != "" {
		log.Info("note: the cache input is ignored; the toolchain is kept in the workspace for the job")
	}
	ws := c.Workspace
	workDir, err := resolveWorkingDir(ws, in.WorkingDir)
	if err != nil {
		return err
	}

	versionInput := in.NodeVersion
	fileInput := in.NodeVersionFile
	if versionInput != "" && fileInput != "" {
		return fmt.Errorf("node-version and node-version-file are mutually exclusive")
	}
	if versionInput == "" && fileInput == "" {
		for _, candidate := range []string{".nvmrc", "package.json"} {
			if _, err := os.Stat(filepath.Join(workDir, candidate)); err == nil {
				fileInput = candidate
				log.Info(fmt.Sprintf("no node-version given; deriving from %s", candidate))
				break
			}
		}
	}
	if fileInput != "" {
		path := fileInput
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		v, err := nodeVersionFromFile(path)
		if err != nil {
			return err
		}
		versionInput = v
		log.Info(fmt.Sprintf("using %s from %s", v, fileInput))
	}
	if versionInput == "" {
		return fmt.Errorf("node-version, node-version-file, a .nvmrc or a package.json is required")
	}

	var index []nodeIndexEntry
	if err := dl.FetchJSON(nodeDistURL+"index.json", &index); err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	entry, filename, err := resolveNode(index, versionInput)
	if err != nil {
		return err
	}
	nodeVersion := strings.TrimPrefix(entry.Version, "v")

	distBase := nodeDistURL
	if muslImage() {
		distBase = nodeUnofficialDistURL
		filename = muslFilename(filename)
		log.Info("musl-based image detected; using musl builds from unofficial-builds.nodejs.org")
	}

	installRoot := in.InstallDir
	if installRoot == "" {
		installRoot = filepath.Join(ws, ".actions", "node")
	}
	dest := filepath.Join(installRoot, "node-"+nodeVersion)
	binDir := filepath.Join(dest, "bin")
	nodeBin := filepath.Join(binDir, "node")

	force := gha.Truthy(in.Force, false)
	if !force {
		if _, err := os.Stat(nodeBin); err == nil {
			log.Info(fmt.Sprintf("reusing cached toolchain %s", nodeVersion))
			return finish(c, binDir, nodeVersion, workDir, log, in)
		}
	}

	log.Info(fmt.Sprintf("installing Node.js %s (%s)", nodeVersion, filename))
	shasums, err := dl.Get(distBase+entry.Version+"/SHASUMS256.txt", nil)
	if err != nil {
		return fmt.Errorf("fetch SHASUMS256.txt: %w", err)
	}
	sum, err := findChecksum(shasums, filename)
	if err != nil {
		return err
	}
	archive := filepath.Join(installRoot, filename)
	if err := dl.DownloadFile(distBase+entry.Version+"/"+filename, archive, sum, log); err != nil {
		return err
	}
	tmp := dest + ".extract"
	os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	err = tarx.ExtractTarGz(f, tmp, 1)
	f.Close()
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("extract %s: %w", filename, err)
	}
	os.RemoveAll(dest)
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	os.Remove(archive)

	return finish(c, binDir, nodeVersion, workDir, log, in)
}

func finish(c *gha.Context, binDir, version, workDir string, log *slog.Logger, in Inputs) error {
	if err := ensureNodeRuns(binDir, log); err != nil {
		return err
	}
	if err := c.AddPath(binDir); err != nil {
		return err
	}
	if err := c.Output("node-version", version); err != nil {
		return err
	}
	if err := setupPackageManager(gha.FirstNonEmpty(in.PackageManager, "pnpm"), binDir, workDir, log); err != nil {
		return err
	}

	if in.RegistryURL != "" {
		token := in.Token
		if token == "" {
			token = c.Token
		}
		key, err := npmAuthKey(in.RegistryURL)
		if err != nil {
			return err
		}
		entries := []struct{ k, v string }{{"registry", strings.TrimSuffix(in.RegistryURL, "/")}}
		if token != "" {
			entries = append(entries, struct{ k, v string }{key, token})
		}
		npmrc := filepath.Join(workDir, ".npmrc")
		if err := upsertNpmrc(npmrc, entries); err != nil {
			return err
		}
		log.Info(fmt.Sprintf("wrote %s (registry %s)", npmrc, in.RegistryURL))
	}

	if dir := compatLibDir(binDir); isDir(dir) {
		ld := strings.Join(nonEmpty(dir, os.Getenv("LD_LIBRARY_PATH")), string(os.PathListSeparator))
		if err := c.ExportEnv("LD_LIBRARY_PATH", ld); err != nil {
			return err
		}
	}

	log.Info(fmt.Sprintf("Node.js %s ready at %s", version, binDir))
	return nil
}

// ensureNodeRuns smoke-executes the installed node; when it cannot
// load the image's C++ runtime (bare musl images ship no libstdc++),
// it bundles Alpine's libstdc++/libgcc into the toolchain and retries.
func ensureNodeRuns(binDir string, log *slog.Logger) error {
	err := checkNodeRuns(binDir)
	if err == nil || !muslImage() {
		return err
	}
	log.Info("the image is missing node's C++ runtime; bundling Alpine's libstdc++/libgcc")
	if berr := bundleCompatLibs(binDir, log); berr != nil {
		return fmt.Errorf("%w (bundling libstdc++/libgcc failed: %v)", err, berr)
	}
	return checkNodeRuns(binDir)
}

// setupPackageManager sets up the package manager selected by the
// package-manager input: pnpm and yarn are installed into the Node
// toolchain using its bundled npm, while npm ships with Node and needs
// nothing. An install already present in the toolchain is reused.
func setupPackageManager(mgr, binDir, workDir string, log *slog.Logger) error {
	mgr = strings.ToLower(strings.TrimSpace(mgr))
	switch mgr {
	case "npm":
		return nil
	case "pnpm", "yarn":
	default:
		return fmt.Errorf("package-manager must be pnpm, yarn or npm, not %q", mgr)
	}
	if _, err := os.Stat(filepath.Join(binDir, mgr)); err == nil {
		log.Info(fmt.Sprintf("reusing %s already installed in the toolchain", mgr))
		return nil
	}
	version, err := packageManagerVersion(workDir, mgr, log)
	if err != nil {
		return err
	}
	if version == "" {
		version = "latest"
	}
	log.Info(fmt.Sprintf("installing %s %s", mgr, version))
	// Run npm as `node npm-cli.js` instead of executing bin/npm: its
	// #!/usr/bin/env node shebang makes the kernel search PATH for
	// node, which fails on runner images that ship no Node (the reason
	// this action exists) and whose env resolves a duplicated PATH
	// differently than expected.
	npmCli, err := filepath.EvalSymlinks(filepath.Join(binDir, "npm"))
	if err != nil {
		return fmt.Errorf("resolve the toolchain's npm: %w", err)
	}
	var out bytes.Buffer
	// --prefix keeps the install inside the toolchain even when a
	// system or user npmrc redirects global installs elsewhere.
	cmd := exec.Command(filepath.Join(binDir, "node"), npmCli, "install", "--global", "--prefix", filepath.Dir(binDir), mgr+"@"+version)
	// Run from outside the project so a project .npmrc (e.g. a private
	// registry without these packages) cannot hijack the install.
	cmd.Dir = os.TempDir()
	cmd.Stdout, cmd.Stderr = &out, &out
	cmd.Env = childEnv(binDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install %s %s: %w\n%s", mgr, version, err, out.String())
	}
	return nil
}

// packageManagerVersion returns the version that the corepack-style
// packageManager field of workDir's package.json pins for mgr, e.g.
// "pnpm@9.15.0+sha512-..." for pnpm -> "9.15.0". It returns "" when
// there is no package.json, no such field, or the field names a
// different package manager.
func packageManagerVersion(workDir, mgr string, log *slog.Logger) (string, error) {
	b, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		return "", fmt.Errorf("parse package.json: %w", err)
	}
	if pkg.PackageManager == "" {
		return "", nil
	}
	spec, _, _ := strings.Cut(pkg.PackageManager, "+") // drop corepack's sha512 suffix
	if name, version, ok := strings.Cut(spec, "@"); ok && strings.EqualFold(name, mgr) {
		return version, nil
	}
	log.Info(fmt.Sprintf("packageManager %q does not name %s; installing latest", pkg.PackageManager, mgr))
	return "", nil
}

// checkNodeRuns smoke-executes the installed node so that an image
// whose libc cannot load the distribution (e.g. a glibc build on musl)
// fails here with a clear error instead of deep inside a package
// manager install.
func checkNodeRuns(binDir string) error {
	var out bytes.Buffer
	cmd := exec.Command(filepath.Join(binDir, "node"), "--version")
	cmd.Env = childEnv(binDir)
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("the installed node does not run on this image: %w\n%s", err, out.String())
	}
	return nil
}

// childEnv builds the environment for executing toolchain binaries:
// the toolchain first on PATH, plus its bundled musl compat libraries
// on LD_LIBRARY_PATH when present.
func childEnv(binDir string) []string {
	ld := os.Getenv("LD_LIBRARY_PATH")
	if dir := compatLibDir(binDir); isDir(dir) {
		ld = strings.Join(nonEmpty(dir, ld), string(os.PathListSeparator))
	}
	env := withoutEnv("PATH", "LD_LIBRARY_PATH")
	if ld != "" {
		env = append(env, "LD_LIBRARY_PATH="+ld)
	}
	return append(env, "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// compatLibDir is where bundleCompatLibs places the C++ runtime
// libraries inside the toolchain.
func compatLibDir(binDir string) string {
	return filepath.Join(filepath.Dir(binDir), "lib", "compat")
}

func nonEmpty(vals ...string) []string {
	var out []string
	for _, v := range vals {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// bundleCompatLibs downloads Alpine's libstdc++ and libgcc packages —
// the two shared libraries the musl node builds load dynamically —
// verifies their contents against the packages' embedded checksums and
// copies the libraries into lib/compat inside the toolchain.
func bundleCompatLibs(binDir string, log *slog.Logger) error {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64", "arm": "armv7"}[runtime.GOARCH]
	if arch == "" {
		return fmt.Errorf("no Alpine package architecture for %s", runtime.GOARCH)
	}
	repo := alpineCDN + arch
	indexGz, err := dl.Get(repo+"/APKINDEX.tar.gz", nil)
	if err != nil {
		return fmt.Errorf("fetch APKINDEX: %w", err)
	}
	index, err := apkIndexMember(indexGz)
	if err != nil {
		return err
	}
	pkgs := parseAPKINDEX(index)

	tmp, err := os.MkdirTemp("", "node-setup-compat")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	libDir := compatLibDir(binDir)
	for _, cl := range compatLibs {
		pkg, ok := pkgs[cl.pkg]
		if !ok || pkg.version == "" {
			return fmt.Errorf("%s not found in the Alpine index", cl.pkg)
		}
		url := fmt.Sprintf("%s/%s-%s.apk", repo, cl.pkg, pkg.version)
		archive := filepath.Join(tmp, cl.pkg+".apk")
		if err := dl.DownloadFile(url, archive, "", log); err != nil {
			return err
		}
		dst := filepath.Join(libDir, filepath.Base(cl.member))
		if err := extractCompatLib(archive, cl.member, dst); err != nil {
			return err
		}
	}
	log.Info(fmt.Sprintf("bundled libstdc++/libgcc into %s", libDir))
	return nil
}

// compatLibs maps the Alpine packages to the library members node
// needs from them.
var compatLibs = []struct{ pkg, member string }{
	{"libstdc++", "usr/lib/libstdc++.so.6"},
	{"libgcc", "usr/lib/libgcc_s.so.1"},
}

// extractCompatLib streams the library matching memberPrefix (the
// plain name or a versioned sibling of it) out of an Alpine package,
// verifying it against the package's embedded APK-TOOLS.checksum
// record, and writes it to dst.
func extractCompatLib(apkPath, memberPrefix, dst string) error {
	f, err := os.Open(apkPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("no library matching %s in %s", memberPrefix, apkPath)
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(hdr.Name)), "./")
		if name != memberPrefix && !strings.HasPrefix(name, memberPrefix+".") {
			continue
		}
		if hdr.Typeflag != tar.TypeReg || hdr.Size == 0 {
			continue // skip the plain-name symlink packages ship
		}
		want := hdr.PAXRecords["APK-TOOLS.checksum.SHA1"]
		if want == "" {
			return fmt.Errorf("%s in %s carries no apk checksum", name, apkPath)
		}
		h := sha1.New()
		part := dst + ".part"
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(part, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		_, err = io.Copy(io.MultiWriter(out, h), tr)
		out.Close()
		if err != nil {
			os.Remove(part)
			return fmt.Errorf("read %s: %w", name, err)
		}
		if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, want) {
			os.Remove(part)
			return fmt.Errorf("checksum mismatch for %s: got %s want %s", name, got, want)
		}
		return os.Rename(part, dst)
	}
}

// apkIndexMember pulls the APKINDEX member out of an APKINDEX.tar.gz.
func apkIndexMember(gz []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("gunzip APKINDEX: %w", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("APKINDEX member not found")
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == "APKINDEX" && hdr.Typeflag == tar.TypeReg {
			return io.ReadAll(tr)
		}
	}
}

// apkPackage is one package stanza of an APKINDEX.
type apkPackage struct {
	version string
}

// parseAPKINDEX reads the plain-text package stanzas of an APKINDEX.
func parseAPKINDEX(b []byte) map[string]apkPackage {
	pkgs := map[string]apkPackage{}
	var name string
	var pkg apkPackage
	flush := func() {
		if name != "" {
			pkgs[name] = pkg
		}
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "P:"):
			flush()
			name, pkg = strings.TrimPrefix(line, "P:"), apkPackage{}
		case name == "":
		case strings.HasPrefix(line, "V:"):
			pkg.version = strings.TrimPrefix(line, "V:")
		}
	}
	flush()
	return pkgs
}

// muslFilename maps an official linux tarball name onto its musl
// counterpart from unofficial-builds.nodejs.org, e.g.
// node-v22.13.1-linux-x64.tar.gz -> node-v22.13.1-linux-x64-musl.tar.gz.
func muslFilename(filename string) string {
	return strings.TrimSuffix(filename, ".tar.gz") + "-musl.tar.gz"
}

// muslImage reports whether the runner image is musl-based.
func muslImage() bool {
	return detectMusl(runtime.GOARCH, os.Stat)
}

// glibcLoaders lists where glibc images keep their dynamic loader;
// an image carrying one of these (including compatibility shims like
// Alpine's gcompat) can run the official builds.
var glibcLoaders = map[string][]string{
	"amd64": {"/lib64/ld-linux-x86-64.so.2", "/lib/ld-linux-x86-64.so.2", "/usr/lib/ld-linux-x86-64.so.2"},
	"arm64": {"/lib/ld-linux-aarch64.so.1", "/usr/lib/ld-linux-aarch64.so.1"},
	"arm":   {"/lib/ld-linux-armhf.so.3", "/usr/lib/ld-linux-armhf.so.3"},
}

// detectMusl decides musl-ness from the dynamic loaders present: musl
// images carry ld-musl-<libc-arch>.so.1 while glibc images do not.
func detectMusl(goarch string, stat func(string) (fs.FileInfo, error)) bool {
	libcArch := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
		"arm":   "armv7",
	}[goarch]
	if libcArch == "" {
		return false
	}
	exists := func(paths ...string) bool {
		for _, p := range paths {
			if _, err := stat(p); err == nil {
				return true
			}
		}
		return false
	}
	if !exists("/lib/ld-musl-"+libcArch+".so.1", "/usr/lib/ld-musl-"+libcArch+".so.1") {
		return false
	}
	return !exists(glibcLoaders[goarch]...)
}

// withoutEnv returns the current environment with entries for the
// given keys removed, so setting a variable again cannot leave
// duplicate entries whose resolution varies between C libraries.
func withoutEnv(keys ...string) []string {
	var out []string
outer:
	for _, e := range os.Environ() {
		for _, k := range keys {
			if strings.HasPrefix(e, k+"=") {
				continue outer
			}
		}
		out = append(out, e)
	}
	return out
}

// resolveWorkingDir resolves the working-dir input against the workspace
// and checks that it names an existing directory; an empty input means
// the workspace itself.
func resolveWorkingDir(ws, input string) (string, error) {
	if input == "" {
		return ws, nil
	}
	dir := input
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(ws, dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("working-dir %q: %w", input, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("working-dir %q is not a directory", input)
	}
	return dir, nil
}

// npmAuthKey builds the .npmrc auth token key for a registry URL, e.g.
// https://example.com/api/packages/owner/npm/ ->
// //example.com/api/packages/owner/npm/:_authToken.
func npmAuthKey(registryURL string) (string, error) {
	u, err := url.Parse(strings.TrimSuffix(registryURL, "/") + "/")
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid registry-url %q", registryURL)
	}
	return "//" + u.Host + u.Path + ":_authToken", nil
}

// upsertNpmrc replaces or appends key lines in a project .npmrc,
// preserving any other content.
func upsertNpmrc(path string, entries []struct{ k, v string }) error {
	var lines []string
	if b, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	}
	for _, e := range entries {
		prefix := e.k + "="
		replaced := false
		for i, line := range lines {
			if strings.HasPrefix(line, prefix) {
				lines[i] = prefix + e.v
				replaced = true
			}
		}
		if !replaced {
			lines = append(lines, prefix+e.v)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

// nodeArch maps Go arch names onto the names nodejs.org uses.
func nodeArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm":
		return "armv7l"
	default:
		return goarch // arm64, ppc64le, s390x match
	}
}

// resolveNode picks the index entry matching pattern and the linux
// tar.gz for the current architecture.
func resolveNode(index []nodeIndexEntry, pattern string) (*nodeIndexEntry, string, error) {
	p := strings.TrimSpace(strings.ToLower(pattern))
	p = strings.TrimPrefix(p, "v")

	best := -1
	bestV := ver.Version{}
	consider := func(i int, v ver.Version) {
		if best < 0 || ver.Compare(v, bestV) > 0 {
			best, bestV = i, v
		}
	}

	for i := range index {
		e := &index[i]
		v, err := ver.Parse(strings.TrimPrefix(e.Version, "v"))
		if err != nil {
			continue
		}
		// Wildcards and "latest" resolve to stable releases only; an
		// exact version may match a pre-release.
		exact := strings.Count(p, ".") >= 2 && !strings.HasPrefix(p, "lts")
		if !exact && v.Pre != "" {
			continue
		}
		switch {
		case p == "" || p == "latest" || p == "current":
			consider(i, v)
		case p == "lts/*":
			if ltsName(e) != "" {
				consider(i, v)
			}
		case strings.HasPrefix(p, "lts/"):
			if strings.EqualFold(ltsName(e), strings.TrimPrefix(p, "lts/")) {
				consider(i, v)
			}
		default:
			want, err := ver.Parse(p)
			if err != nil {
				return nil, "", fmt.Errorf("invalid node-version %q: %w", pattern, err)
			}
			if ver.Compare(v, want) == 0 {
				best = i
				bestV = v
			} else if strings.Count(p, ".") == 1 && v.Major == want.Major && v.Minor == want.Minor {
				consider(i, v)
			} else if strings.Count(p, ".") == 0 && v.Major == want.Major {
				consider(i, v)
			}
		}
	}
	if best < 0 {
		return nil, "", fmt.Errorf("no Node.js release found matching %q", pattern)
	}
	arch := nodeArch(runtime.GOARCH)
	if slices.Contains(index[best].Files, "linux-"+arch) {
		return &index[best], fmt.Sprintf("node-%s-linux-%s.tar.gz", index[best].Version, arch), nil
	}
	return nil, "", fmt.Errorf("Node.js %s has no linux/%s tar.gz", index[best].Version, arch)
}

func ltsName(e *nodeIndexEntry) string {
	if s, ok := e.Lts.(string); ok {
		return s
	}
	return ""
}

// nodeVersionFromFile reads a version from .nvmrc or the engines.node
// field of a package.json.
func nodeVersionFromFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	switch filepath.Base(path) {
	case ".nvmrc":
		line := strings.TrimSpace(strings.SplitN(string(b), "\n", 2)[0])
		if line == "" {
			return "", fmt.Errorf("%s is empty", path)
		}
		return line, nil
	default:
		var pkg struct {
			Engines struct {
				Node string `json:"node"`
			} `json:"engines"`
		}
		if err := json.Unmarshal(b, &pkg); err != nil {
			return "", fmt.Errorf("parse %s: %w", path, err)
		}
		if pkg.Engines.Node == "" {
			return "latest", nil
		}
		re := regexp.MustCompile(`\d+(?:\.\d+){0,2}`)
		if m := re.FindString(pkg.Engines.Node); m != "" {
			return m, nil
		}
		return "", fmt.Errorf("cannot derive a version from engines.node %q", pkg.Engines.Node)
	}
}

// findChecksum locates the sha256 of filename in a SHASUMS256.txt body.
func findChecksum(body []byte, filename string) (string, error) {
	for line := range strings.SplitSeq(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("%s not found in SHASUMS256.txt", filename)
}
