package gha

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scrubRunnerEnv clears every runner-injected variable (restored by
// t.Setenv at test end), so the scenarios below are deterministic even
// inside a job container, where FORGEJO_*/GITHUB_*/ACTIONS_* variables
// describe the real job and would otherwise leak into ContextFromEnv.
func scrubRunnerEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "FORGEJO_") || strings.HasPrefix(key, "GITHUB_") ||
			strings.HasPrefix(key, "GITEA_") || strings.HasPrefix(key, "ACTIONS_") {
			t.Setenv(key, "")
		}
	}
}

func TestContextFromEnvGitHub(t *testing.T) {
	scrubRunnerEnv(t)
	t.Setenv("FORGEJO_JOB", "") // force the GITHUB prefix even inside a Forgejo job
	t.Setenv("GITHUB_JOB", "build")
	t.Setenv("GITHUB_WORKSPACE", "/job/ws")
	t.Setenv("GITHUB_SERVER_URL", "https://git.example.com/")
	t.Setenv("GITHUB_API_URL", "https://git.example.com/api/v1")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("GITHUB_REF", "refs/heads/main")
	t.Setenv("GITHUB_TOKEN", "tok")

	c := ContextFromEnv()
	if c.Workspace != "/job/ws" || c.Repository != "owner/repo" || c.Ref != "refs/heads/main" {
		t.Fatalf("context fields not resolved: %+v", c)
	}
	if c.ServerURL != "https://git.example.com" {
		t.Errorf("trailing slash should be trimmed, got %q", c.ServerURL)
	}
	if c.APIURL != "https://git.example.com/api/v1" {
		t.Errorf("API URL wrong: %q", c.APIURL)
	}
	if c.Token != "tok" {
		t.Errorf("token wrong: %q", c.Token)
	}
	if c.Owner() != "owner" {
		t.Errorf("owner wrong: %q", c.Owner())
	}
	if got := c.RepoURL(); got != "https://git.example.com/owner/repo" {
		t.Errorf("repo URL wrong: %q", got)
	}
	if got := c.Host(); got != "git.example.com" {
		t.Errorf("host wrong: %q", got)
	}
	if got := c.ResolvePath("sub/dir"); got != "/job/ws/sub/dir" {
		t.Errorf("relative path should resolve against workspace, got %q", got)
	}
	if got := c.ResolvePath("/abs"); got != "/abs" {
		t.Errorf("absolute path should pass through, got %q", got)
	}
}

func TestContextFromEnvForgejoPrefixWins(t *testing.T) {
	scrubRunnerEnv(t)
	t.Setenv("FORGEJO_JOB", "build")
	t.Setenv("FORGEJO_REPOSITORY", "owner/forge")
	t.Setenv("GITHUB_REPOSITORY", "owner/hub")
	t.Setenv("FORGEJO_SERVER_URL", "https://git.example.com")

	c := ContextFromEnv()
	if c.Repository != "owner/forge" {
		t.Errorf("FORGEJO_ prefix should win, got %q", c.Repository)
	}
	// Without an explicit API URL it is derived from the server URL.
	if c.APIURL != "https://git.example.com/api/v1" {
		t.Errorf("API URL should be derived: %q", c.APIURL)
	}
}

func TestContextFromEnvEmpty(t *testing.T) {
	// Without any runner job marker the context still resolves (to
	// whatever GITHUB_* variables exist, usually none when run by
	// hand) and stays usable.
	if c := ContextFromEnv(); c == nil {
		t.Fatal("nil context")
	}
}

func TestContextOutput(t *testing.T) {
	t.Setenv("FORGEJO_JOB", "")
	dir := t.TempDir()
	file := filepath.Join(dir, "output")
	t.Setenv("GITHUB_JOB", "build")
	t.Setenv("GITHUB_OUTPUT", file)

	c := ContextFromEnv()
	if err := c.Output("tags", "a,b"); err != nil {
		t.Fatal(err)
	}
	if err := c.Output("multi", "one\ntwo"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.HasPrefix(got, "tags=a,b\n") {
		t.Errorf("single-line output wrong: %q", got)
	}
	if !strings.Contains(got, "multi<<gha-multi-") {
		t.Errorf("heredoc opener wrong: %q", got)
	}
}

func TestOutputHeredoc(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "output")
	t.Setenv("GITHUB_OUTPUT", file)

	if err := Output("tags", "a,b"); err != nil {
		t.Fatal(err)
	}
	if err := Output("multi", "one\ntwo"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.HasPrefix(got, "tags=a,b\n") {
		t.Errorf("single-line output wrong: %q", got)
	}
	if !strings.Contains(got, "multi<<gha-multi-") {
		t.Errorf("heredoc opener wrong: %q", got)
	}
	if !strings.Contains(got, "one\ntwo\n") {
		t.Errorf("heredoc body wrong: %q", got)
	}
}

func TestAddPathResolvesRelativeToWorkspace(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "path")
	t.Setenv("GITHUB_PATH", file)
	t.Setenv("GITHUB_WORKSPACE", filepath.Join(dir, "ws"))

	if err := AddPath(".actions/go/go1.22.5/bin"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "ws", ".actions", "go", "go1.22.5", "bin") + "\n"
	if string(b) != want {
		t.Errorf("got %q want %q", string(b), want)
	}
}

func TestMask(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Mask("jwt-secret")
	Mask("with\nnewline")
	w.Close()
	os.Stdout = old
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "::add-mask::jwt-secret\n" {
		t.Errorf("got %q, want a single add-mask command", string(b))
	}
}

func TestTruthy(t *testing.T) {
	cases := []struct {
		v    string
		def  bool
		want bool
	}{
		{"", true, true},
		{"", false, false},
		{"TRUE", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"1", false, true},
		{"no", true, false},
		{"Off", true, false},
		{"0", true, false},
		{"garbage", true, true},
	}
	for _, tc := range cases {
		if got := Truthy(tc.v, tc.def); got != tc.want {
			t.Errorf("Truthy(%q, %v) = %v, want %v", tc.v, tc.def, got, tc.want)
		}
	}
}

func TestIntValue(t *testing.T) {
	if n, err := IntValue("retries", " 7 ", 1); err != nil || n != 7 {
		t.Errorf("got %d, %v", n, err)
	}
	if n, err := IntValue("retries", "", 3); err != nil || n != 3 {
		t.Errorf("empty should give default, got %d, %v", n, err)
	}
	if _, err := IntValue("retries", "soon", 3); err == nil ||
		err.Error() != "the retries input must be a whole number" {
		t.Errorf("malformed value should error naming the input, got %v", err)
	}
}

func TestConfigureLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(&prefixHandler{prefix: "[go-test] ", level: slog.LevelDebug, w: &buf})
	logger.Info("running tests", "packages", "./...")
	logger.Debug("debug detail")
	logger.Warn("careful")
	out := buf.String()
	if !strings.Contains(out, "[go-test] running tests packages=./...\n") {
		t.Errorf("info line wrong: %q", out)
	}
	if !strings.Contains(out, "[go-test] debug detail\n") {
		t.Errorf("debug line wrong: %q", out)
	}
	if !strings.Contains(out, "[go-test] careful\n") {
		t.Errorf("warn line wrong: %q", out)
	}
	if strings.Contains(out, "time=") || strings.Contains(out, "level=") {
		t.Errorf("records should be plain lines, got %q", out)
	}

	// At the default level, debug records are filtered out.
	buf.Reset()
	logger = slog.New(&prefixHandler{prefix: "[x] ", level: slog.LevelInfo, w: &buf})
	logger.Debug("hidden")
	logger.Info("shown")
	if strings.Contains(buf.String(), "hidden") || !strings.Contains(buf.String(), "shown") {
		t.Errorf("level filtering wrong: %q", buf.String())
	}
}
