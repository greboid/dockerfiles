package gitcmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

func TestSetupStaticToken(t *testing.T) {
	c := &gha.Context{Token: "job-token", Actor: "someone"}

	a, err := Setup(t.Context(), c, nil, "static-token", "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Cleanup()

	if a.askpass == "" {
		t.Fatal("static token should use an askpass helper")
	}
	if _, err := os.Stat(a.askpass); err != nil {
		t.Fatalf("askpass helper missing: %v", err)
	}
	if info, err := os.Stat(a.askpass); err == nil && info.Mode().Perm() != 0o700 {
		t.Errorf("askpass helper should be 0700, got %v", info.Mode().Perm())
	}
	env := a.Environ()
	if !containsPrefix(env, "GIT_ASKPASS=") || !containsPrefix(env, "GIT_TERMINAL_PROMPT=0") {
		t.Errorf("environment wrong: %v", env)
	}

	a.Cleanup()
	if _, err := os.Stat(a.askpass); !os.IsNotExist(err) {
		t.Errorf("askpass helper should be removed, got %v", err)
	}
}

func TestSetupBearerToken(t *testing.T) {
	// An audience with an endpoint always wins over the job token: it
	// is a JWT presented as a header, not basic auth. The OIDC fetch
	// itself is exercised in the oidc package; here only the resulting
	// environment matters, so point the endpoint at an unreachable
	// stub... actually, serve one.
	srv := newIDTokenServer(t, "jwt-value")

	c := &gha.Context{Actor: "someone", IDTokenURL: srv.URL, IDTokenToken: "req"}
	a, err := Setup(t.Context(), c, nil, "", "my-audience", "")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Cleanup()

	if a.askpass != "" {
		t.Error("bearer tokens should not use an askpass helper")
	}
	env := a.Environ()
	if !containsPrefix(env, "GIT_CONFIG_COUNT=1") ||
		!containsPrefix(env, "GIT_CONFIG_KEY_0=http.extraheader") ||
		!containsPrefix(env, "GIT_CONFIG_VALUE_0=Authorization: Bearer jwt-value") {
		t.Errorf("environment wrong: %v", env)
	}
}

func TestSetupHonorsExistingConfigCount(t *testing.T) {
	t.Setenv("GIT_CONFIG_COUNT", "2")
	t.Setenv("GIT_CONFIG_KEY_1", "preexisting")

	srv := newIDTokenServer(t, "jwt-value")
	c := &gha.Context{IDTokenURL: srv.URL, IDTokenToken: "req"}
	a, err := Setup(t.Context(), c, nil, "", "aud", "")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Cleanup()

	if !containsPrefix(a.Environ(), "GIT_CONFIG_COUNT=3") ||
		!containsPrefix(a.Environ(), "GIT_CONFIG_KEY_2=http.extraheader") {
		t.Errorf("existing entries not honored: %v", a.Environ())
	}
}

func TestSetupNoToken(t *testing.T) {
	c := &gha.Context{}
	if _, err := Setup(t.Context(), c, nil, "", "", ""); err == nil ||
		!strings.Contains(err.Error(), "no token for git authentication") {
		t.Errorf("missing token should error clearly, got %v", err)
	}
}

func TestAskpassScript(t *testing.T) {
	script := askpassScript("user'name", "to'ken")
	want := `#!/bin/sh
case "$1" in
	*Username*) printf %s 'user'\''name' ;;
	*Password*) printf %s 'to'\''ken' ;;
	*) printf "\n" ;;
esac
`
	if script != want {
		t.Errorf("script wrong:\n%s", script)
	}
}

func TestDefaultIdentity(t *testing.T) {
	// Run git with no global or system config so no identity is set.
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")

	c := &gha.Context{Actor: "octo", ServerURL: "https://git.example.com"}
	env, err := DefaultIdentity(c, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GIT_AUTHOR_NAME=octo",
		"GIT_AUTHOR_EMAIL=octo@users.noreply.git.example.com",
		"GIT_COMMITTER_NAME=octo",
		"GIT_COMMITTER_EMAIL=octo@users.noreply.git.example.com",
	}
	if len(env) != len(want) {
		t.Fatalf("got %v, want %v", env, want)
	}
	for i := range want {
		if env[i] != want[i] {
			t.Errorf("env[%d] = %q, want %q", i, env[i], want[i])
		}
	}

	// An explicit name wins over the actor.
	env, _ = DefaultIdentity(c, "named")
	if !containsPrefix(env, "GIT_AUTHOR_NAME=named") {
		t.Errorf("explicit name not used: %v", env)
	}
}

func TestCurrentBranch(t *testing.T) {
	c := &gha.Context{Ref: "refs/heads/from-ref"}

	// A directory that is no repository: rev-parse fails, the ref
	// fallback answers.
	if got := CurrentBranch(c, t.TempDir()); got != "from-ref" {
		t.Errorf("got %q", got)
	}

	// Neither a repository nor a branch ref: nothing.
	if got := CurrentBranch(&gha.Context{Ref: "refs/tags/v1"}, t.TempDir()); got != "" {
		t.Errorf("got %q", got)
	}
}

func containsPrefix(env []string, prefix string) bool {
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			return true
		}
	}
	return false
}

// newIDTokenServer serves the ID token endpoint response the runner
// would.
func newIDTokenServer(t *testing.T, jwt string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":"` + jwt + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}
