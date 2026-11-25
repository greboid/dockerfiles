// Package buildah prepares an unprivileged action container for
// buildah operation and wraps buildah invocations in that environment.
//
// The job/action container buildah runs in is unprivileged: its
// auto-detected overlay storage driver cannot mount, so the driver is
// forced to overlay through fuse-overlayfs when that actually
// initializes (a present /dev/fuse and binary do not guarantee the
// mount is permitted), else vfs; isolation is chroot, which needs no
// OCI runtime. On
// overlayfs-rooted containers TMPDIR moves to a tmpfs: buildah mounts
// an overlay over the build context with its scaffolding under TMPDIR,
// and inside the user namespace buildah always re-executes into, that
// overlay cannot be stacked on the container's own overlayfs root.
// Pulled image layers also stage under TMPDIR, so the tmpfs must
// additionally be large enough — with docker's default 64MB /dev/shm
// larger pulls fail with "no space left on device".
//
// When BUILDAH_CACHE_DIR names a directory shared between jobs, images
// and layers persist there and builds reuse them. When GOMODCACHE is
// set, that directory is mounted over /go/pkg/mod in RUN instructions.
package buildah

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Toolchain carries the environment buildah invocations run with;
// construct it with Prepare.
type Toolchain struct {
	driver  string
	env     []string
	layers  bool
	volumes []string
}

// Driver reports the storage driver selected by Prepare.
func (t *Toolchain) Driver() string { return t.driver }

// Prepare probes the container and pins the storage driver, isolation
// and (when the root is overlayfs) a tmpfs TMPDIR. It fails with a
// clear error when buildah or its network helper netavark are missing
// — our action images bake both in; elsewhere install them first.
func Prepare(ctx context.Context, log *slog.Logger) (*Toolchain, error) {
	if _, err := exec.LookPath("buildah"); err != nil {
		return nil, fmt.Errorf("buildah not found on PATH; use an action image that ships it")
	}
	if !netavarkInstalled() {
		return nil, fmt.Errorf("netavark not found in podman's libexec directories; buildah bud needs it — use an action image that ships it")
	}

	driver, mountProgram := probeDriver(ctx)
	if driver == "vfs" && log != nil {
		log.Info("fuse-overlayfs unavailable; using the vfs storage driver (slower)")
	}
	graphroot := cacheStore(log, driver)

	mounts, _ := os.ReadFile("/proc/mounts")
	env := os.Environ()
	if rootfsIsOverlay(mounts) {
		dir, kb := findTmpfs(mounts, tmpfsCandidates())
		if dir != "" {
			os.Setenv("TMPDIR", dir)
			env = setEnv(env, "TMPDIR", dir)
			if kb < 409600 && log != nil {
				log.Warn(fmt.Sprintf("TMPDIR tmpfs %s has only %dMB free; pulled layers also stage there, so larger images will fail with \"no space left on device\" — give the job container a bigger tmpfs (runner container.options: --shm-size=1g)", dir, kb/1024))
			}
		} else if log != nil {
			log.Warn("overlayfs container root but no tmpfs-backed TMPDIR; the overlay buildah mounts over the build context may fail")
		}
	}

	// Pin the storage driver, overriding whatever the distro shipped.
	// The default path is the standard one; exporting
	// CONTAINERS_STORAGE_CONF keeps child buildah invocations on the
	// same file.
	conf := os.Getenv("CONTAINERS_STORAGE_CONF")
	if conf == "" {
		conf = "/etc/containers/storage.conf"
	}
	if err := os.MkdirAll(filepath.Dir(conf), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(conf, []byte(storageConf(driver, mountProgram, graphroot)), 0o644); err != nil {
		return nil, err
	}
	env = setEnv(env, "CONTAINERS_STORAGE_CONF", conf)
	env = setEnv(env, "BUILDAH_ISOLATION", "chroot")

	if log != nil {
		log.Info(fmt.Sprintf("buildah storage driver: %s (isolation: chroot)", driver))
	}
	return &Toolchain{driver: driver, env: env, layers: graphroot != "", volumes: goModVolumes()}, nil
}

// probeDriver returns the storage driver to use and, for overlay, the
// fuse-overlayfs path to mount it with.
func probeDriver(ctx context.Context) (driver, mountProgram string) {
	path, err := exec.LookPath("fuse-overlayfs")
	if err != nil {
		return "vfs", ""
	}
	probe, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := exec.CommandContext(probe, "buildah", "--storage-driver", "overlay", "--storage-opt", "overlay.mount_program="+path, "info").Run(); err != nil {
		return "vfs", ""
	}
	return "overlay", path
}

// cacheStore returns the persistent image store under
// BUILDAH_CACHE_DIR, or "" for the container's own storage. vfs is
// excluded because it stores a full copy of the image per layer.
func cacheStore(log *slog.Logger, driver string) string {
	dir := os.Getenv("BUILDAH_CACHE_DIR")
	if dir == "" {
		return ""
	}
	if driver != "overlay" {
		if log != nil {
			log.Warn("BUILDAH_CACHE_DIR ignored: the cache needs the overlay storage driver")
		}
		return ""
	}
	store := filepath.Join(dir, "storage")
	if err := os.MkdirAll(store, 0o700); err != nil {
		if log != nil {
			log.Warn("BUILDAH_CACHE_DIR ignored: " + err.Error())
		}
		return ""
	}
	if log != nil {
		log.Info("buildah cache: " + store)
	}
	return store
}

// storageConf renders storage.conf. With a shared graphroot the runroot
// stays inside the container: it holds per-container mount state.
func storageConf(driver, mountProgram, graphroot string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[storage]\ndriver = %q\n", driver)
	if graphroot != "" {
		fmt.Fprintf(&b, "graphroot = %q\nrunroot = %q\n", graphroot, "/run/containers/storage")
	}
	if mountProgram != "" {
		fmt.Fprintf(&b, "\n[storage.options.overlay]\nmount_program = %q\n", mountProgram)
	}
	return b.String()
}

// goModVolumes returns the bud --volume mounting GOMODCACHE at the
// module cache of the golang image (GOPATH=/go).
func goModVolumes() []string {
	dir := os.Getenv("GOMODCACHE")
	if !filepath.IsAbs(dir) {
		return nil
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil
	}
	return []string{dir + ":/go/pkg/mod"}
}

// UniqueName returns prefix with a random suffix. Local names must be
// unique because the image store can be shared between containers.
func UniqueName(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}

// RemoveManifest deletes a local manifest list so it does not stay in a
// persistent store. Failure only leaves the list behind, so it is ignored.
func (t *Toolchain) RemoveManifest(ctx context.Context, name string) {
	_ = t.Command(ctx, "manifest", "rm", name).Run()
}

// Command returns an exec.Cmd for buildah with the prepared
// environment. The caller wires up streams and runs it.
func (t *Toolchain) Command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "buildah", args...)
	cmd.Env = t.env
	return cmd
}

// Run executes buildah with the prepared environment, streaming its
// output into the workflow log.
func (t *Toolchain) Run(ctx context.Context, args ...string) error {
	cmd := t.Command(ctx, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Output executes buildah with the prepared environment and returns
// its stdout.
func (t *Toolchain) Output(ctx context.Context, args ...string) (string, error) {
	var out bytes.Buffer
	cmd := t.Command(ctx, args...)
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// setEnv replaces or appends KEY=VALUE in an environment slice.
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	found := false
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			if !found {
				out = append(out, prefix+value)
				found = true
			}
			continue
		}
		out = append(out, kv)
	}
	if !found {
		out = append(out, prefix+value)
	}
	return out
}

// Login runs buildah login against registry, feeding the password over
// stdin so it never shows in argv.
func Login(ctx context.Context, authfile, registry, username, password string) error {
	cmd := exec.CommandContext(ctx, "buildah", "login",
		"--authfile", authfile, "--username", username, "--password-stdin", registry)
	cmd.Stdin = strings.NewReader(password + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BudOptions describes a buildah bud invocation, in the shape the
// docker-build and docker-push actions share.
type BudOptions struct {
	AuthFile  string   // registry credentials for FROM-instructions pulls
	IIDFile   string   // file the built image ID is written to
	Platform  string   // target platform (os/arch[/variant]); unset builds the host platform
	Manifest  string   // manifest list to add the image to, created when missing
	Tags      []string // full image references to tag the image with
	RelFile   string   // Containerfile/Dockerfile, relative to the context
	Target    string   // build target stage
	NoCache   bool
	BuildArgs []string // KEY=VALUE, one per entry; values may contain spaces
	Context   string   // build context directory
	Layers    bool     // keep layers for reuse and refresh base images; set by Bud
	Volumes   []string // SRC:DEST mounts for RUN instructions; set by Bud
}

// Bud builds an OCI image with the prepared environment.
func (t *Toolchain) Bud(ctx context.Context, o BudOptions) error {
	o.Layers, o.Volumes = t.layers, t.volumes
	return t.Run(ctx, budArgs(o)...)
}

// budArgs renders the buildah bud invocation Bud runs, so tests can
// inspect it without a buildah binary.
func budArgs(o BudOptions) []string {
	args := []string{"bud", "--format", "oci", "--isolation", "chroot", "--network", "host", "--timestamp", "0"}
	if o.AuthFile != "" {
		args = append(args, "--authfile", o.AuthFile)
	}
	if o.IIDFile != "" {
		args = append(args, "--iidfile", o.IIDFile)
	}
	if o.Platform != "" {
		args = append(args, "--platform", o.Platform)
	}
	if o.Manifest != "" {
		args = append(args, "--manifest", o.Manifest)
	}
	args = append(args, "--file", o.RelFile)
	if o.Target != "" {
		args = append(args, "--target", o.Target)
	}
	if o.NoCache {
		args = append(args, "--no-cache")
	}
	if o.Layers {
		// A persistent store would otherwise keep floating base tags stale.
		args = append(args, "--layers", "--pull=newer")
	}
	for _, v := range o.Volumes {
		args = append(args, "--volume", v)
	}
	for _, ba := range o.BuildArgs {
		args = append(args, "--build-arg", ba)
	}
	for _, tag := range o.Tags {
		args = append(args, "--tag", tag)
	}
	args = append(args, o.Context)
	return args
}

// ManifestPushOptions describes a buildah manifest push invocation,
// publishing a manifest list built through BudOptions.Manifest.
type ManifestPushOptions struct {
	AuthFile   string // registry credentials
	DigestFile string // file the digest of the pushed list is written to
	List       string // local name of the manifest list
	Dest       string // destination reference; docker:// implied when no transport is given
}

// ManifestPush publishes a manifest list to its destination reference,
// pushing the platform images the list references along with it.
func (t *Toolchain) ManifestPush(ctx context.Context, o ManifestPushOptions) error {
	return t.Run(ctx, manifestPushArgs(o)...)
}

// manifestPushArgs renders the buildah manifest push invocation
// ManifestPush runs, so tests can inspect it without a buildah binary.
func manifestPushArgs(o ManifestPushOptions) []string {
	args := []string{"manifest", "push", "--all"}
	if o.AuthFile != "" {
		args = append(args, "--authfile", o.AuthFile)
	}
	if o.DigestFile != "" {
		args = append(args, "--digestfile", o.DigestFile)
	}
	args = append(args, o.List, dockerRef(o.Dest))
	return args
}

// dockerRef prefixes an image reference with the docker:// transport
// when it has none — buildah push implies it for bare references,
// manifest push does not. References naming another transport (a
// scheme:// or the docker-daemon:/containers-storage: forms) are
// returned unchanged.
func dockerRef(ref string) string {
	if strings.Contains(ref, "://") ||
		strings.HasPrefix(ref, "docker-daemon:") ||
		strings.HasPrefix(ref, "containers-storage:") {
		return ref
	}
	return "docker://" + ref
}

// ResolveBuildFile resolves the build context directory and the
// Containerfile/Dockerfile to build with: context resolves against the
// workspace; file, relative to the context, defaults to Containerfile
// when one exists, else Dockerfile, and must live inside the context.
func ResolveBuildFile(workspace, context, file string) (ctxDir, relFile string, err error) {
	ctxDir = context
	if ctxDir == "" {
		ctxDir = "."
	}
	if !filepath.IsAbs(ctxDir) {
		ctxDir = filepath.Join(workspace, ctxDir)
	}
	if info, err := os.Stat(ctxDir); err != nil || !info.IsDir() {
		return "", "", fmt.Errorf("context directory %s does not exist", ctxDir)
	}

	if file == "" {
		switch {
		case fileExists(filepath.Join(ctxDir, "Containerfile")):
			file = "Containerfile"
		case fileExists(filepath.Join(ctxDir, "Dockerfile")):
			file = "Dockerfile"
		default:
			return "", "", fmt.Errorf("no Containerfile or Dockerfile in %s (set the file input to point at one)", ctxDir)
		}
	}
	absFile := file
	if !filepath.IsAbs(absFile) {
		absFile = filepath.Join(ctxDir, file)
	}
	if !fileExists(absFile) {
		return "", "", fmt.Errorf("%s not found (the file input is relative to the build context)", file)
	}
	rel, err := filepath.Rel(ctxDir, absFile)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("%s must live inside the build context", file)
	}
	return ctxDir, rel, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// SplitRefs splits a comma- or whitespace-separated list of image
// references — safe because image references cannot contain
// whitespace, which is the invariant the tag lists rely on.
func SplitRefs(s string) []string {
	return strings.Fields(strings.ReplaceAll(s, ",", " "))
}

// TagSuffix returns the tag part of an image reference: the text after
// the last colon of its last path component, defaulting to latest.
func TagSuffix(ref string) string {
	last := ref[strings.LastIndexByte(ref, '/')+1:]
	if i := strings.LastIndexByte(last, ':'); i >= 0 {
		return last[i+1:]
	}
	return "latest"
}

// tmpfsCandidates lists the directories to consider for the relocated
// TMPDIR, in preference order: the current TMPDIR, then the usual
// tmpfs mounts of a job container.
func tmpfsCandidates() []string {
	candidates := []string{}
	if tmp := os.Getenv("TMPDIR"); tmp != "" {
		candidates = append(candidates, tmp)
	}
	return append(candidates, "/dev/shm", "/tmp")
}

// netavarkInstalled reports whether the netavark network helper is
// present in one of the directories podman/buildah search by default.
// bud needs it even with --network host, and it is not on PATH, so
// exec.LookPath cannot find it.
func netavarkInstalled() bool {
	for _, d := range []string{
		"/usr/local/libexec/podman", "/usr/local/lib/podman",
		"/usr/libexec/podman", "/usr/lib/podman",
	} {
		if info, err := os.Stat(filepath.Join(d, "netavark")); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return true
		}
	}
	return false
}

// mount is one entry of a /proc/mounts table.
type mount struct {
	point  string
	fstype string
}

// parseMounts extracts mount point and file system type from a
// /proc/mounts table.
func parseMounts(table []byte) []mount {
	var out []mount
	for line := range strings.SplitSeq(string(table), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// Octal escapes in the device/mountpoint fields do not matter
		// for the two mount points looked at here ("/" and the tmpfs
		// candidates, none of which need escaping).
		out = append(out, mount{point: fields[1], fstype: fields[2]})
	}
	return out
}

// rootfsIsOverlay reports whether the container's root filesystem is
// overlayfs (docker's overlay2 driver) — the usual case for runner
// containers.
func rootfsIsOverlay(table []byte) bool {
	for _, m := range parseMounts(table) {
		if m.point == "/" && m.fstype == "overlay" {
			return true
		}
	}
	return false
}

// findTmpfs returns the first candidate that is an existing, writable
// directory mounted as tmpfs, together with its free space in KB.
func findTmpfs(table []byte, candidates []string) (string, int) {
	for _, d := range candidates {
		if !isTmpfs(table, d) {
			continue
		}
		if info, err := os.Stat(d); err != nil || !info.IsDir() {
			continue
		}
		if err := syscall.Access(d, 0o2); err != nil {
			continue
		}
		var st syscall.Statfs_t
		if err := syscall.Statfs(d, &st); err != nil {
			return d, 0
		}
		return d, int(st.Bavail * uint64(st.Bsize))
	}
	return "", 0
}

// isTmpfs reports whether the mounts table lists dir as tmpfs.
func isTmpfs(table []byte, dir string) bool {
	for _, m := range parseMounts(table) {
		if m.point == dir && m.fstype == "tmpfs" {
			return true
		}
	}
	return false
}
