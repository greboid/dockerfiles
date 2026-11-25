// Package gha provides helpers for talking to the workflow runner:
// resolving the workflow context from the environment, writing step
// outputs, environment exports and PATH additions through the runner's
// GITHUB_* files, masking secrets, and logging. It works with GitHub
// Actions and with act-based runners such as Forgejo Runner. Every
// runner interaction degrades to a no-op when the runner's files are
// absent, so the action binaries can be run by hand for debugging.
package gha

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Context describes the workflow an action runs in. It is resolved
// once, at startup, from the runner's environment instead of being
// re-read piecemeal: Forgejo Runner exports FORGEJO_* variables
// alongside the GitHub-compatible ones, and the FORGEJO_ prefix wins
// when the job carries it.
type Context struct {
	Workspace  string // directory the job's repository is checked out into
	ServerURL  string // instance base URL, e.g. https://git.example.com
	APIURL     string // REST API base URL (instance API when the runner did not advertise one)
	Repository string // owner/name of the job's repository
	Ref        string // fully qualified ref, e.g. refs/heads/main
	SHA        string // commit the workflow runs on
	Actor      string // user that triggered the workflow
	HeadRef    string // pull request head branch, on pull events
	EventName  string // event that triggered the workflow

	// ID token endpoint the runner advertises when the workflow set
	// enable-openid-connect; both empty when it did not.
	IDTokenURL   string
	IDTokenToken string

	// Token is the job token: FORGEJO_TOKEN, GITHUB_TOKEN or
	// GITEA_TOKEN, whichever is set first.
	Token string

	outputFile string
	pathFile   string
	envFile    string
}

// ContextFromEnv resolves the workflow context from the environment.
// There is no error for an environment without any runner variables:
// the fields are simply empty, and the runner file writes stay no-ops,
// so the binaries keep working by hand.
func ContextFromEnv() *Context {
	prefix := "GITHUB"
	if os.Getenv("FORGEJO_JOB") != "" {
		prefix = "FORGEJO"
	}
	lookup := func(key string) string { return os.Getenv(prefix + key) }

	c := &Context{
		Workspace:    lookup("_WORKSPACE"),
		ServerURL:    strings.TrimSuffix(lookup("_SERVER_URL"), "/"),
		Repository:   lookup("_REPOSITORY"),
		Ref:          lookup("_REF"),
		SHA:          lookup("_SHA"),
		Actor:        lookup("_ACTOR"),
		HeadRef:      lookup("_HEAD_REF"),
		EventName:    lookup("_EVENT_NAME"),
		IDTokenURL:   os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"),
		IDTokenToken: os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"),
		Token:        FirstNonEmpty(lookup("_TOKEN"), os.Getenv("GITHUB_TOKEN"), os.Getenv("GITEA_TOKEN")),
		outputFile:   lookup("_OUTPUT"),
		pathFile:     lookup("_PATH"),
		envFile:      lookup("_ENV"),
	}
	c.APIURL = FirstNonEmpty(lookup("_API_URL"), apiFromServer(c.ServerURL))
	if c.Workspace == "" {
		c.Workspace = Workspace()
	}
	return c
}

// apiFromServer derives the instance REST API base from the server URL
// (Gitea/Forgejo layout), or returns "" when there is no server URL.
func apiFromServer(server string) string {
	if server == "" {
		return ""
	}
	return server + "/api/v1"
}

// Host returns the instance host name without scheme or path — the
// registry host for the instance's container registry, and the domain
// of noreply e-mail addresses.
func (c *Context) Host() string {
	h := c.ServerURL
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	if i := strings.IndexByte(h, '/'); i >= 0 {
		h = h[:i]
	}
	return h
}

// Owner returns the owner (first path segment) of the repository.
func (c *Context) Owner() string {
	owner, _, _ := strings.Cut(c.Repository, "/")
	return owner
}

// RepoURL returns the instance URL of the repository, without a .git
// suffix.
func (c *Context) RepoURL() string {
	return c.ServerURL + "/" + c.Repository
}

// ResolvePath resolves path against the workspace; absolute paths pass
// through unchanged.
func (c *Context) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.Workspace, path)
}

// Output records a step output via the runner's $GITHUB_OUTPUT file.
func (c *Context) Output(key, value string) error {
	return writeKeyValue(c.outputFile, key, value)
}

// ExportEnv exports an environment variable for subsequent steps via
// $GITHUB_ENV.
func (c *Context) ExportEnv(key, value string) error {
	return writeKeyValue(c.envFile, key, value)
}

// AddPath adds dir to PATH for subsequent steps via $GITHUB_PATH.
func (c *Context) AddPath(dir string) error {
	if c.pathFile == "" {
		return nil
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(c.Workspace, dir)
	}
	return appendLine(c.pathFile, dir)
}

// Mask asks the runner to redact value from workflow logs, via the
// add-mask workflow command on stdout. It is a no-op for empty or
// multi-line values, which the command format cannot express.
func (c *Context) Mask(value string) {
	Mask(value)
}

// Output records a step output via $GITHUB_OUTPUT. Package-level
// counterpart of Context.Output, resolving the file from the
// environment; a no-op when the runner did not provide one.
func Output(key, value string) error {
	return writeKeyValue(os.Getenv("GITHUB_OUTPUT"), key, value)
}

// ExportEnv exports an environment variable for subsequent steps via
// $GITHUB_ENV.
func ExportEnv(key, value string) error {
	return writeKeyValue(os.Getenv("GITHUB_ENV"), key, value)
}

// AddPath adds dir to PATH for subsequent steps via $GITHUB_PATH.
// Relative directories are resolved against the workspace.
func AddPath(dir string) error {
	file := os.Getenv("GITHUB_PATH")
	if file == "" {
		return nil
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(Workspace(), dir)
	}
	return appendLine(file, dir)
}

// Mask asks the runner to redact value from workflow logs.
func Mask(value string) {
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return
	}
	fmt.Printf("::add-mask::%s\n", value)
}

// Workspace returns the directory the workflow was checked out into,
// falling back to the current working directory.
func Workspace() string {
	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		return ws
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// FirstNonEmpty returns the first argument that is not empty.
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Truthy parses a boolean in the spellings the runners accept
// (true/yes/on/1 and their negations, case-insensitive), returning def
// for unset values and falling back to def for anything unrecognized.
func Truthy(v string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return def
	case "true", "yes", "on", "1":
		return true
	case "false", "no", "off", "0":
		return false
	}
	return def
}

// IntValue parses v as a whole number for the named input, returning
// def when empty and an error naming the input when malformed.
func IntValue(name, v string, def int) (int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("the %s input must be a whole number", name)
	}
	return n, nil
}

// writeKeyValue appends key/value to the runner file in its format,
// using a heredoc block for multi-line values. An empty file name is a
// no-op.
func writeKeyValue(file, key, value string) error {
	if file == "" {
		return nil
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if strings.ContainsAny(value, "\n\r") {
		delim := fmt.Sprintf("gha-%s-%d", strings.ReplaceAll(key, "-", "_"), time.Now().UnixNano())
		_, err = fmt.Fprintf(f, "%s<<%s\n%s\n%s\n", key, delim, value, delim)
	} else {
		_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	}
	return err
}

// appendLine appends a bare line (a PATH entry) to a runner file.
func appendLine(file, line string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	return err
}

// ConfigureLogging returns the action's logger. It writes
// "[<name>] message" lines to stdout so they interleave with the
// runner's step log; debug records are only included when debug is set.
func ConfigureLogging(name string, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(&prefixHandler{prefix: "[" + name + "] ", level: level, w: os.Stdout, mu: &sync.Mutex{}})
}

// prefixHandler is a minimal slog handler rendering records as
// "[prefix] message key=value ..." lines without timestamps.
type prefixHandler struct {
	prefix string
	level  slog.Level
	w      io.Writer
	mu     *sync.Mutex
}

func (h *prefixHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *prefixHandler) Handle(_ context.Context, r slog.Record) error {
	if h.mu == nil {
		h.mu = &sync.Mutex{}
	}
	var b []byte
	b = append(b, h.prefix...)
	b = append(b, r.Message...)
	r.Attrs(func(a slog.Attr) bool {
		b = append(b, ' ')
		b = append(b, a.Key...)
		b = append(b, '=')
		b = append(b, a.Value.String()...)
		return true
	})
	b = append(b, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(b)
	return err
}

func (h *prefixHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	var prefix strings.Builder
	prefix.WriteString(h.prefix)
	for _, a := range attrs {
		prefix.WriteString(a.Key + "=" + a.Value.String() + " ")
	}
	return &prefixHandler{prefix: prefix.String(), level: h.level, w: h.w, mu: h.mu}
}

func (h *prefixHandler) WithGroup(string) slog.Handler { return h }
