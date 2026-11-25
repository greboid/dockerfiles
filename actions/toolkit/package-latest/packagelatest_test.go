package packagelatest

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// registry is a fake package API: it answers /api/v1/packages/{owner}
// from its versions, honouring the page/limit/type/q parameters the
// action sends and recording the Authorization header of the last
// request.
type registry struct {
	versions []string
	auth     string
	owner    string
	queries  []urlValues
}

type urlValues map[string]string

func (r *registry) serve(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, "/api/v1/packages/") {
			http.NotFound(w, req)
			return
		}
		r.auth = req.Header.Get("Authorization")
		r.owner = strings.TrimPrefix(req.URL.Path, "/api/v1/packages/")
		q := req.URL.Query()
		r.queries = append(r.queries, urlValues{
			"type":  q.Get("type"),
			"q":     q.Get("q"),
			"page":  q.Get("page"),
			"limit": q.Get("limit"),
		})
		page, _ := strconv.Atoi(q.Get("page"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit <= 0 {
			limit = pageSize
		}
		start := (page - 1) * limit
		end := start + limit
		if start > len(r.versions) {
			start = len(r.versions)
		}
		if end > len(r.versions) {
			end = len(r.versions)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Total-Count", strconv.Itoa(len(r.versions)))
		fmt.Fprintf(w, "[")
		for i, v := range r.versions[start:end] {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, `{"id":%d,"type":"container","name":"repo/image","version":%q,"created_at":"2026-01-01T00:00:00Z"}`, i+start, v)
		}
		fmt.Fprintf(w, "]")
	}))
	t.Cleanup(srv.Close)
	return srv
}

// runContext scrubs the runner environment (a real job's FORGEJO_*/
// GITHUB_* variables would leak into the context), applies the given
// KEY=VALUE overrides, points the output file at a temp file, and
// returns the context plus a reader for the outputs the action wrote.
func runContext(t *testing.T, env ...string) (*gha.Context, func(string) string) {
	t.Helper()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "FORGEJO_") || strings.HasPrefix(key, "GITHUB_") ||
			strings.HasPrefix(key, "GITEA_") || strings.HasPrefix(key, "ACTIONS_") {
			t.Setenv(key, "")
		}
	}
	for _, kv := range env {
		key, value, _ := strings.Cut(kv, "=")
		t.Setenv(key, value)
	}
	out := filepath.Join(t.TempDir(), "output")
	if err := os.WriteFile(out, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", out)
	read := func(key string) string {
		t.Helper()
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		for line := range strings.SplitSeq(string(b), "\n") {
			if k, v, found := strings.Cut(line, "="); found && k == key {
				return v
			}
		}
		return "<missing>"
	}
	return gha.ContextFromEnv(), read
}

func TestLatestAcrossPages(t *testing.T) {
	// Sorted the way the API might return them (newest first): the
	// action must still pick the numerically highest version, 10 not 9.
	reg := &registry{versions: []string{
		"1.20261007.9", "1.20261007.3", "1.20261007.10", "1.20261007.2",
	}}
	server := reg.serve(t)
	old := pageSize
	pageSize = 2
	defer func() { pageSize = old }()

	c, read := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid", Type: "container", Package: "repo/image"})
	if err != nil {
		t.Fatal(err)
	}
	if got := read("version"); got != "1.20261007.10" {
		t.Errorf("version = %q, want 1.20261007.10", got)
	}
	if got := read("next"); got != "1.20261007.11" {
		t.Errorf("next = %q, want 1.20261007.11", got)
	}
	if len(reg.queries) < 3 {
		t.Errorf("expected paginated requests, got %d", len(reg.queries))
	}
	if q := reg.queries[0]; q["type"] != "container" || q["q"] != "repo/image" {
		t.Errorf("first request filters wrong: %v", q)
	}
	if q := reg.queries[1]; q["page"] != "2" {
		t.Errorf("second request page wrong: %v", q)
	}
}

func TestPrefixFilterAndNext(t *testing.T) {
	reg := &registry{versions: []string{"1.20261006.4", "1.20261007.3", "2.0.0"}}
	server := reg.serve(t)

	c, read := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid", Prefix: "1.20261007."})
	if err != nil {
		t.Fatal(err)
	}
	if got := read("version"); got != "1.20261007.3" {
		t.Errorf("version = %q, want 1.20261007.3", got)
	}
	if got := read("next"); got != "1.20261007.4" {
		t.Errorf("next = %q, want 1.20261007.4", got)
	}
}

func TestUnparseableVersionsIgnored(t *testing.T) {
	reg := &registry{versions: []string{"latest", "sha256-abc", "1.2.3"}}
	server := reg.serve(t)

	c, read := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid"})
	if err != nil {
		t.Fatal(err)
	}
	if got := read("version"); got != "1.2.3" {
		t.Errorf("version = %q, want 1.2.3", got)
	}
	if got := read("next"); got != "1.2.4" {
		t.Errorf("next = %q, want 1.2.4", got)
	}
}

func TestNoMatches(t *testing.T) {
	for _, tt := range []struct {
		prefix   string
		versions []string
		wantNext string
	}{
		{"1.20261007.", []string{"2.0.0"}, "1.20261007.0"},
		{"", []string{"latest"}, "0.0.0"},
	} {
		reg := &registry{versions: tt.versions}
		server := reg.serve(t)

		c, read := runContext(t)
		err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid", Prefix: tt.prefix})
		if err != nil {
			t.Fatal(err)
		}
		if got := read("version"); got != "" {
			t.Errorf("prefix %q: version = %q, want empty", tt.prefix, got)
		}
		if got := read("next"); got != tt.wantNext {
			t.Errorf("prefix %q: next = %q, want %q", tt.prefix, got, tt.wantNext)
		}
	}
}

func TestEmptyRegistry(t *testing.T) {
	server := (&registry{}).serve(t)

	c, read := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid", Prefix: "1."})
	if err != nil {
		t.Fatal(err)
	}
	if got := read("version"); got != "" {
		t.Errorf("version = %q, want empty", got)
	}
	if got := read("next"); got != "1.0" {
		t.Errorf("next = %q, want 1.0", got)
	}
}

func TestTokenSent(t *testing.T) {
	reg := &registry{}
	server := reg.serve(t)

	c, _ := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: server.URL, Owner: "greboid", Token: "s3cret"})
	if err != nil {
		t.Fatal(err)
	}
	if reg.auth != "token s3cret" {
		t.Errorf("Authorization = %q, want the token scheme", reg.auth)
	}
}

func TestOwnerFromContext(t *testing.T) {
	reg := &registry{}
	server := reg.serve(t)

	c, _ := runContext(t, "GITHUB_REPOSITORY=owner/repo")
	if err := Run(c, testLogger(), Inputs{ServerURL: server.URL}); err != nil {
		t.Fatal(err)
	}
	if reg.owner != "owner" {
		t.Errorf("listed owner = %q, want owner (from the workflow repository)", reg.owner)
	}
}

func TestMissingServerAndOwner(t *testing.T) {
	c, _ := runContext(t)
	if err := Run(c, testLogger(), Inputs{}); err == nil {
		t.Error("expected an error without a server URL")
	}

	reg := &registry{}
	server := reg.serve(t)
	if err := Run(c, testLogger(), Inputs{ServerURL: server.URL}); err == nil {
		t.Error("expected an error without an owner")
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	c, _ := runContext(t)
	err := Run(c, testLogger(), Inputs{ServerURL: srv.URL, Owner: "greboid"})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want a 403 status error", err)
	}
}
