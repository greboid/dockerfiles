package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

func TestParseAssets(t *testing.T) {
	ws := t.TempDir()
	got, err := parseAssets(ws, "bin/melange\n  bin/melange.sha256 other.apk")
	if err != nil {
		t.Fatal(err)
	}
	want := []asset{
		{path: filepath.Join(ws, "bin/melange"), name: "melange"},
		{path: filepath.Join(ws, "bin/melange.sha256"), name: "melange.sha256"},
		{path: filepath.Join(ws, "other.apk"), name: "other.apk"},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("asset %d = %+v want %+v", i, got[i], want[i])
		}
	}

	// name=path renames; absolute paths pass through untouched.
	got, err = parseAssets(ws, "checksums=/tmp/melange.sha256 melange=bin/melange")
	if err != nil {
		t.Fatal(err)
	}
	if got[0].name != "checksums" || got[0].path != "/tmp/melange.sha256" {
		t.Errorf("renamed asset = %+v", got[0])
	}
	if got[1].name != "melange" || got[1].path != filepath.Join(ws, "bin/melange") {
		t.Errorf("second asset = %+v", got[1])
	}
}

func TestParseAssetsErrors(t *testing.T) {
	for _, in := range []string{
		"=bin/melange",      // empty name
		"name=",             // empty path
		"a=b a=c",           // duplicate names
		"sub/dir=name=path", // path separator in a custom name
	} {
		if _, err := parseAssets(".", in); err == nil {
			t.Errorf("parseAssets(%q) should fail", in)
		}
	}
	if assets, err := parseAssets(".", "  \n\t "); err != nil || assets != nil {
		t.Errorf("empty input should give no assets, got %v, %v", assets, err)
	}
}

func TestTagFromRef(t *testing.T) {
	if got := tagFromRef("refs/tags/v1.2.3"); got != "v1.2.3" {
		t.Errorf("got %q", got)
	}
	for _, ref := range []string{"refs/heads/main", "refs/pull/7/merge", "", "refs/tags/"} {
		if got := tagFromRef(ref); got != "" {
			t.Errorf("tagFromRef(%q) = %q, want empty", ref, got)
		}
	}
}

func TestStatusError(t *testing.T) {
	e := &statusError{op: "POST release v1", status: 409, body: `{"message":"Release already has a Tag"}`}
	if !e.conflict() {
		t.Error("409 should be a conflict")
	}
	if s := e.Error(); !strings.Contains(s, "Conflict") || !strings.Contains(s, "Release already has a Tag") {
		t.Errorf("Error() = %q", s)
	}
	if (&statusError{op: "x", status: 500}).conflict() {
		t.Error("500 should not be a conflict")
	}
}

func TestAPIMessage(t *testing.T) {
	if got := apiMessage([]byte(`{"message":"boom"}`)); got != "boom" {
		t.Errorf("got %q", got)
	}
	if got := apiMessage([]byte("plain text")); got != "plain text" {
		t.Errorf("got %q", got)
	}
}

// mockForgejo answers the release API endpoints Run talks to, with the
// /api/v1 prefix and the repository path the action builds.
type mockForgejo struct {
	t        *testing.T
	lookup   func() (int, bool) // release id, exists
	releases int                // successful creates
	deleted  []int64
	uploads  []string

	// failingFirstCreate makes the first create answer 409, as a
	// concurrent run re-creating the release would.
	failingFirstCreate bool
}

func (m *mockForgejo) handler() http.Handler {
	api := "/api/v1/repos/o/r"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "token tok" {
			m.t.Errorf("Authorization = %q", got)
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == api+"/releases/tags/v1":
			id, ok := m.lookup()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.writeJSON(w, http.StatusOK, releaseOut{ID: int64(id), TagName: "v1", HTMLURL: "https://example.com/o/r/releases/tag/v1"})
		case r.Method == http.MethodPost && r.URL.Path == api+"/releases":
			var opt createReleaseOption
			if err := json.NewDecoder(r.Body).Decode(&opt); err != nil {
				m.t.Errorf("create payload: %v", err)
			}
			if opt.TagName != "v1" {
				m.t.Errorf("create payload = %+v", opt)
			}
			if m.releases == 0 && m.failingFirstCreate {
				m.releases++
				w.WriteHeader(http.StatusConflict)
				io.WriteString(w, `{"message":"Release already has a Tag"}`)
				return
			}
			m.releases++
			m.writeJSON(w, http.StatusCreated, releaseOut{ID: 7, TagName: "v1", HTMLURL: "https://example.com/o/r/releases/tag/v1"})
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, api+"/releases/"):
			var id int64
			if _, err := fmt.Sscanf(r.URL.Path, api+"/releases/%d", &id); err != nil {
				m.t.Errorf("delete path %q: %v", r.URL.Path, err)
			}
			m.deleted = append(m.deleted, id)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/assets"):
			m.uploads = append(m.uploads, r.URL.Query().Get("name"))
			m.writeJSON(w, http.StatusCreated, map[string]any{"id": 1, "name": r.URL.Query().Get("name")})
		default:
			m.t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func (m *mockForgejo) writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		m.t.Errorf("encode response: %v", err)
	}
}

// scrubRunnerEnv clears every runner-injected variable (restored by
// t.Setenv at test end), so ContextFromEnv stays deterministic even
// inside a job container, where FORGEJO_*/GITHUB_*/ACTIONS_* variables
// describe the real job and would otherwise leak into the context.
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

// runTool drives Run against the mock with one asset on disk and the
// runner's output file captured.
func runTool(t *testing.T, srv *httptest.Server, ws string, in Inputs) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "output")
	scrubRunnerEnv(t)
	t.Setenv("FORGEJO_JOB", "") // force the GITHUB prefix even inside a Forgejo job
	t.Setenv("GITHUB_OUTPUT", out)
	c := gha.ContextFromEnv()
	c.APIURL = srv.URL + "/api/v1"
	c.Repository = "o/r"
	c.Workspace = ws
	if err := Run(context.Background(), c, slog.New(slog.DiscardHandler), in); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

func TestRunCreatesAndUploads(t *testing.T) {
	m := &mockForgejo{t: t, lookup: func() (int, bool) { return 0, false }}
	srv := httptest.NewServer(m.handler())
	defer srv.Close()

	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "bin/tool"), []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "bin/tool.sha256"), []byte("hash"), 0o644); err != nil {
		t.Fatal(err)
	}

	outputs := runTool(t, srv, ws, Inputs{
		Tag:       "v1",
		Commitish: "abc",
		Body:      "notes",
		Assets:    "bin/tool bin/tool.sha256",
		Token:     "tok",
	})
	if m.releases != 1 || len(m.deleted) != 0 || len(m.uploads) != 2 {
		t.Errorf("releases=%d deleted=%v uploads=%v", m.releases, m.deleted, m.uploads)
	}
	if m.uploads[0] != "tool" || m.uploads[1] != "tool.sha256" {
		t.Errorf("uploads = %v", m.uploads)
	}
	for _, want := range []string{"id=7", "assets=2", "url=https://example.com/o/r/releases/tag/v1"} {
		if !strings.Contains(outputs, want) {
			t.Errorf("outputs missing %q: %s", want, outputs)
		}
	}
}

func TestRunReplacesExisting(t *testing.T) {
	m := &mockForgejo{t: t, lookup: func() (int, bool) { return 5, true }}
	srv := httptest.NewServer(m.handler())
	defer srv.Close()

	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "tool"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	runTool(t, srv, ws, Inputs{Tag: "v1", Assets: "tool", Token: "tok"})
	if m.releases != 1 || len(m.deleted) != 1 || m.deleted[0] != 5 || len(m.uploads) != 1 {
		t.Errorf("releases=%d deleted=%v uploads=%v, want 1, [5], 1", m.releases, m.deleted, m.uploads)
	}
}

// TestRunRetriesAfterConflict covers the race path: between the lookup
// and the create a concurrent run re-created the release, the create
// answers 409, and Run replaces that one and tries once more.
func TestRunRetriesAfterConflict(t *testing.T) {
	seen := 0
	m := &mockForgejo{t: t, lookup: func() (int, bool) {
		seen++
		if seen == 1 { // the pre-create lookup: nothing there yet
			return 0, false
		}
		return 6, true // the concurrent release
	}}
	m.failingFirstCreate = true
	srv := httptest.NewServer(m.handler())
	defer srv.Close()

	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "tool"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	runTool(t, srv, ws, Inputs{Tag: "v1", Assets: "tool", Token: "tok"})
	if m.releases != 2 || len(m.deleted) != 1 || m.deleted[0] != 6 || len(m.uploads) != 1 {
		t.Errorf("releases=%d deleted=%v uploads=%v, want 2, [6], 1", m.releases, m.deleted, m.uploads)
	}
}

func TestRunWithoutTag(t *testing.T) {
	c := gha.ContextFromEnv()
	c.APIURL = "https://example.com/api/v1"
	c.Repository = "o/r"
	c.Ref = "refs/heads/main"
	err := Run(context.Background(), c, slog.New(slog.DiscardHandler), Inputs{Token: "tok"})
	if err == nil || !strings.Contains(err.Error(), "tag input is required") {
		t.Errorf("err = %v, want a missing-tag error", err)
	}
}
