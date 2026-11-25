// Package cache is a minimal client for the artifact cache service
// that Forgejo runners expose to jobs through ACTIONS_CACHE_URL —
// the same REST protocol actions/cache speaks, minus the Node.js
// runtime. Entries are opaque blobs addressed by a (key, version)
// pair; the runner's cache proxy authenticates requests by the run
// id embedded in the URL, so no credentials are needed beyond what
// the environment already provides.
package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/dl"
)

const apiBase = "/_apis/artifactcache"

// Client talks to one job's cache endpoint. Construct it with
// NewFromEnv; a nil Client means caching is unavailable and callers
// should carry on without it.
type Client struct {
	base  string // ACTIONS_CACHE_URL with any trailing slash removed
	token string // optional bearer token; the Forgejo proxy ignores it but the standard clients send one
	log   *slog.Logger
}

// NewFromEnv returns a client for the cache advertised to this job,
// or nil when no cache exists (ACTIONS_CACHE_URL unset — running by
// hand, or a runner without a cache server). The token is taken from
// ACTIONS_CACHE_TOKEN, falling back to ACTIONS_RUNTIME_TOKEN like
// actions/cache does.
func NewFromEnv(log *slog.Logger) *Client {
	base := strings.TrimRight(os.Getenv("ACTIONS_CACHE_URL"), "/")
	if base == "" {
		return nil
	}
	token := os.Getenv("ACTIONS_CACHE_TOKEN")
	if token == "" {
		token = os.Getenv("ACTIONS_RUNTIME_TOKEN")
	}
	return &Client{base: base, token: token, log: log}
}

type lookupResponse struct {
	Result          string `json:"result"`
	ArchiveLocation string `json:"archiveLocation"`
	CacheKey        string `json:"cacheKey"`
}

// Lookup returns the download location for the entry stored under
// key/version, or ok=false when the cache holds no such entry. Keys
// are matched case-insensitively.
func (c *Client) Lookup(key, version string) (location string, ok bool, err error) {
	u := fmt.Sprintf("%s%s/cache?keys=%s&version=%s",
		c.base, apiBase, url.QueryEscape(strings.ToLower(key)), url.QueryEscape(version))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", false, err
	}
	c.decorate(req)
	resp, err := dl.Client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var lr lookupResponse
		if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
			return "", false, fmt.Errorf("lookup %s: decode response: %w", key, err)
		}
		return lr.ArchiveLocation, true, nil
	case http.StatusNoContent:
		return "", false, nil
	default:
		return "", false, fmt.Errorf("lookup %s: unexpected status %s", key, resp.Status)
	}
}

// Save uploads the file at path as the entry for key/version. The
// upload is a single request sized from the file, so path must not
// change while it runs.
func (c *Client) Save(key, version, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	size := info.Size()
	if size == 0 {
		return fmt.Errorf("save %s: refusing to cache empty file", key)
	}

	id, err := c.reserve(key, version, size)
	if err != nil {
		return err
	}
	if err := c.upload(id, path, size); err != nil {
		return err
	}
	return c.commit(id)
}

// reserve allocates a cache id for key/version and declares the
// entry's size, which the server later verifies against the bytes
// actually uploaded.
func (c *Client) reserve(key, version string, size int64) (uint64, error) {
	body, err := json.Marshal(map[string]any{
		"key":       strings.ToLower(key),
		"version":   version,
		"cacheSize": size,
	})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, c.base+apiBase+"/caches", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	c.decorate(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := dl.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("reserve %s: unexpected status %s", key, resp.Status)
	}
	var rr struct {
		CacheID uint64 `json:"cacheId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return 0, fmt.Errorf("reserve %s: decode response: %w", key, err)
	}
	return rr.CacheID, nil
}

// upload streams the whole file in one PATCH; the protocol allows
// chunked ranges but the server accepts a single full-range request,
// which keeps saves simple and restart-free.
func (c *Client) upload(id uint64, path string, size int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s%s/caches/%d", c.base, apiBase, id), f)
	if err != nil {
		return err
	}
	c.decorate(req)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size
	req.Header.Set("Content-Range", fmt.Sprintf("bytes 0-%d/*", size-1))
	resp, err := dl.Client.Do(req)
	if err != nil {
		return fmt.Errorf("upload cache %d: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload cache %d: unexpected status %s", id, resp.Status)
	}
	return nil
}

// commit finalises the entry; the server compares the uploaded bytes
// against the size declared at reserve time.
func (c *Client) commit(id uint64) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s/caches/%d", c.base, apiBase, id), nil)
	if err != nil {
		return err
	}
	c.decorate(req)
	resp, err := dl.Client.Do(req)
	if err != nil {
		return fmt.Errorf("commit cache %d: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("commit cache %d: unexpected status %s", id, resp.Status)
	}
	return nil
}

// Download streams the entry at location (a URL previously returned
// by Lookup) into dst, verifying expectedSHA256 when non-empty. It is
// a thin wrapper so callers do not need to import dl as well.
func (c *Client) Download(location, dst, expectedSHA256 string) error {
	return dl.DownloadFile(location, dst, expectedSHA256, c.log)
}

func (c *Client) decorate(req *http.Request) {
	req.Header.Set("User-Agent", "forgejo-pure-actions/1.0 (+cache)")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.log != nil {
		c.log.Info(fmt.Sprintf("cache request %s %s", req.Method, redact(req.URL)))
	}
}

// redact strips the run id from cache URLs before logging it: the id
// doubles as this run's cache credential. It splices strings rather
// than rebuilding the URL so the placeholder is not percent-escaped.
func redact(u *url.URL) string {
	s := u.String()
	i := strings.Index(s, apiBase)
	if i <= 0 {
		return s
	}
	if j := strings.LastIndex(s[:i], "/"); j >= 0 {
		return s[:j+1] + "<run>" + s[i:]
	}
	return s
}
