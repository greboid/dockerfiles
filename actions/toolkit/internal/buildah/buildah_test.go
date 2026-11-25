package buildah

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSplitRefsAndTagSuffix(t *testing.T) {
	got := SplitRefs(" a,b c,, d ")
	if !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Errorf("got %v", got)
	}
	if !reflect.DeepEqual(SplitRefs("linux/amd64, linux/arm64"), []string{"linux/amd64", "linux/arm64"}) {
		t.Errorf("platform split: %v", SplitRefs("linux/amd64, linux/arm64"))
	}
	cases := map[string]string{
		"reg.example.com/owner/image:1.2": "1.2",
		"reg.example.com/owner/image":     "latest",
		"image:v3":                        "v3",
		"image":                           "latest",
	}
	for ref, want := range cases {
		if got := TagSuffix(ref); got != want {
			t.Errorf("TagSuffix(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestBudArgs(t *testing.T) {
	got := budArgs(BudOptions{
		AuthFile:  "/run/auth.json",
		IIDFile:   "/run/iid",
		Platform:  "linux/arm64",
		Manifest:  "action-manifest-42",
		RelFile:   "Dockerfile",
		Target:    "build",
		NoCache:   true,
		BuildArgs: []string{"A=1", "B=x y"},
		Tags:      []string{"reg.example.com/owner/repo:1", "reg.example.com/owner/repo:2"},
		Context:   "/workspace",
	})
	want := []string{
		"bud", "--format", "oci", "--isolation", "chroot", "--network", "host", "--timestamp", "0",
		"--authfile", "/run/auth.json",
		"--iidfile", "/run/iid",
		"--platform", "linux/arm64",
		"--manifest", "action-manifest-42",
		"--file", "Dockerfile",
		"--target", "build",
		"--no-cache",
		"--build-arg", "A=1",
		"--build-arg", "B=x y",
		"--tag", "reg.example.com/owner/repo:1",
		"--tag", "reg.example.com/owner/repo:2",
		"/workspace",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}

	// The defaults stay lean: no auth, no platform, no manifest, no
	// target, cached build, no tags.
	got = budArgs(BudOptions{RelFile: "Containerfile", Context: "."})
	want = []string{
		"bud", "--format", "oci", "--isolation", "chroot", "--network", "host", "--timestamp", "0",
		"--file", "Containerfile", ".",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("defaults: got %v", got)
	}
}

func TestBudArgsCache(t *testing.T) {
	args := strings.Join(budArgs(BudOptions{RelFile: "Dockerfile", Context: "/c"}), " ")
	if strings.Contains(args, "--layers") || strings.Contains(args, "--pull") || strings.Contains(args, "--volume") {
		t.Errorf("no cache flags expected without a cache: %s", args)
	}
	args = strings.Join(budArgs(BudOptions{RelFile: "Dockerfile", Context: "/c", Layers: true, Volumes: []string{"/m:/go/pkg/mod"}}), " ")
	for _, want := range []string{"--layers", "--pull=newer", "--volume /m:/go/pkg/mod"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q: %s", want, args)
		}
	}
}

func TestStorageConf(t *testing.T) {
	if got, want := storageConf("vfs", "", ""), "[storage]\ndriver = \"vfs\"\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	got := storageConf("overlay", "/usr/bin/fuse-overlayfs", "/caches/buildah/storage")
	for _, want := range []string{
		`driver = "overlay"`,
		`graphroot = "/caches/buildah/storage"`,
		`runroot = "/run/containers/storage"`,
		"[storage.options.overlay]",
		`mount_program = "/usr/bin/fuse-overlayfs"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestCacheStore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BUILDAH_CACHE_DIR", dir)
	if got := cacheStore(nil, "vfs"); got != "" {
		t.Errorf("vfs must not use the cache, got %q", got)
	}
	want := filepath.Join(dir, "storage")
	if got := cacheStore(nil, "overlay"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	t.Setenv("BUILDAH_CACHE_DIR", "")
	if got := cacheStore(nil, "overlay"); got != "" {
		t.Errorf("unset cache dir should disable the cache, got %q", got)
	}
}

func TestGoModVolumes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOMODCACHE", dir)
	if got := goModVolumes(); !reflect.DeepEqual(got, []string{dir + ":/go/pkg/mod"}) {
		t.Errorf("got %v", got)
	}
	for _, v := range []string{"", "relative", filepath.Join(dir, "missing")} {
		t.Setenv("GOMODCACHE", v)
		if got := goModVolumes(); got != nil {
			t.Errorf("GOMODCACHE=%q: got %v", v, got)
		}
	}
}

func TestUniqueName(t *testing.T) {
	a, b := UniqueName("x"), UniqueName("x")
	if a == b || !strings.HasPrefix(a, "x-") {
		t.Errorf("got %q and %q", a, b)
	}
}

func TestManifestPushArgs(t *testing.T) {
	got := manifestPushArgs(ManifestPushOptions{
		AuthFile:   "/run/auth.json",
		DigestFile: "/run/digest",
		List:       "action-manifest-42",
		Dest:       "reg.example.com/owner/repo:1",
	})
	want := []string{
		"manifest", "push", "--all",
		"--authfile", "/run/auth.json",
		"--digestfile", "/run/digest",
		"action-manifest-42", "docker://reg.example.com/owner/repo:1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}

	// Bare destinations get the docker:// transport; explicit ones are
	// left alone.
	got = manifestPushArgs(ManifestPushOptions{List: "l", Dest: "ghcr.io/owner/repo:dev"})
	if last := got[len(got)-1]; last != "docker://ghcr.io/owner/repo:dev" {
		t.Errorf("bare dest: got %q", last)
	}
	got = manifestPushArgs(ManifestPushOptions{List: "l", Dest: "docker://ghcr.io/owner/repo:dev"})
	if last := got[len(got)-1]; last != "docker://ghcr.io/owner/repo:dev" {
		t.Errorf("explicit docker:// dest: got %q", last)
	}
	got = manifestPushArgs(ManifestPushOptions{List: "l", Dest: "docker-daemon:img:1"})
	if last := got[len(got)-1]; last != "docker-daemon:img:1" {
		t.Errorf("docker-daemon dest: got %q", last)
	}
}

func TestResolveBuildFile(t *testing.T) {
	ws := t.TempDir()
	ctxDir := filepath.Join(ws, "ctx")
	if err := os.MkdirAll(ctxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	containerfile := filepath.Join(ctxDir, "Containerfile")
	if err := os.WriteFile(containerfile, []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Context resolves against the workspace; the Containerfile default
	// is relative to the context.
	gotCtx, gotRel, err := ResolveBuildFile(ws, "ctx", "")
	if err != nil {
		t.Fatal(err)
	}
	if gotCtx != ctxDir || gotRel != "Containerfile" {
		t.Errorf("got %q, %q", gotCtx, gotRel)
	}

	// An explicit file relative to the context.
	if _, gotRel, err := ResolveBuildFile(ws, "ctx", "Containerfile"); err != nil || gotRel != "Containerfile" {
		t.Errorf("got %q, %v", gotRel, err)
	}

	// A file outside the context is rejected (it exists — just not in
	// the context).
	if err := os.WriteFile(filepath.Join(ws, "Containerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ResolveBuildFile(ws, "ctx", "../Containerfile"); err == nil ||
		!strings.Contains(err.Error(), "must live inside the build context") {
		t.Errorf("escaping file should error, got %v", err)
	}

	// No Containerfile or Dockerfile at all.
	empty := t.TempDir()
	if _, _, err := ResolveBuildFile(empty, "", ""); err == nil ||
		!strings.Contains(err.Error(), "no Containerfile or Dockerfile") {
		t.Errorf("missing file should error, got %v", err)
	}
}

func TestParseMounts(t *testing.T) {
	table := []byte(`overlay / overlay rw,lowerdir=... 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev,noexec 0 0
shm /dev/shm tmpfs rw,nosuid,nodev,noexec,relatime,size=64m 0 0
`)
	mounts := parseMounts(table)
	if len(mounts) != 4 {
		t.Fatalf("got %d mounts, want 4", len(mounts))
	}
	if !rootfsIsOverlay(table) {
		t.Error("overlay root not detected")
	}
	if rootfsIsOverlay([]byte("ext4 / ext4 rw 0 0\n")) {
		t.Error("ext4 root misdetected as overlay")
	}

	// The candidate is validated against the real file system only
	// after the mounts table names it tmpfs, so with /dev/shm in the
	// table the selection is deterministic on any Linux box.
	dir, kb := findTmpfs(table, []string{"/dev/shm", "/tmp"})
	if dir != "/dev/shm" {
		t.Errorf("tmpfs candidate wrong: %q", dir)
	}
	if kb <= 0 {
		t.Errorf("free space not reported: %d", kb)
	}
	if dir, _ := findTmpfs([]byte("overlay / overlay rw 0 0\n"), []string{"/tmp"}); dir != "" {
		t.Errorf("non-tmpfs candidate should not match, got %q", dir)
	}
	if !isTmpfs(table, "/dev/shm") {
		t.Error("/dev/shm should be tmpfs in the table")
	}
	if isTmpfs(table, "/proc") {
		t.Error("/proc should not be tmpfs in the table")
	}
}

func TestSetEnv(t *testing.T) {
	env := setEnv([]string{"A=1", "TMPDIR=/old", "B=2"}, "TMPDIR", "/new")
	if !reflect.DeepEqual(env, []string{"A=1", "TMPDIR=/new", "B=2"}) {
		t.Errorf("replace failed: %v", env)
	}
	env = setEnv([]string{"A=1"}, "B", "2")
	if !reflect.DeepEqual(env, []string{"A=1", "B=2"}) {
		t.Errorf("append failed: %v", env)
	}
	env = setEnv([]string{"A=1", "A=stale"}, "A", "0")
	if !reflect.DeepEqual(env, []string{"A=0"}) {
		t.Errorf("dedupe failed: %v", env)
	}
}
