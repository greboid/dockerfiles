package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// fakeServer mimics the subset of the artifact cache API the client
// uses: reserve, single-shot upload, commit, lookup and download. It
// mirrors the real server's size check at commit time.
type fakeServer struct {
	mu     sync.Mutex
	nextID uint64
	caches map[uint64]*entry
	server *httptest.Server

	// recorded for assertions
	contentRange string
}

type entry struct {
	key      string
	version  string
	size     int64
	data     []byte
	complete bool
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	fs := &fakeServer{caches: map[uint64]*entry{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/RUNID/_apis/artifactcache/cache", fs.find)
	mux.HandleFunc("/RUNID/_apis/artifactcache/caches", fs.reserve)
	mux.HandleFunc("/RUNID/_apis/artifactcache/caches/", fs.byID)
	mux.HandleFunc("/RUNID/_apis/artifactcache/artifacts/", fs.serve)
	fs.server = httptest.NewServer(mux)
	t.Cleanup(fs.server.Close)
	return fs
}

func (fs *fakeServer) client(t *testing.T) *Client {
	t.Setenv("ACTIONS_CACHE_URL", fs.server.URL+"/RUNID/")
	t.Setenv("ACTIONS_CACHE_TOKEN", "sekrit")
	return NewFromEnv(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func (fs *fakeServer) find(w http.ResponseWriter, r *http.Request) {
	keys := strings.Split(r.URL.Query().Get("keys"), ",")
	version := r.URL.Query().Get("version")
	fs.mu.Lock()
	defer fs.mu.Unlock()
	for _, e := range fs.caches {
		if !e.complete {
			continue
		}
		for _, k := range keys {
			if strings.EqualFold(e.key, k) && e.version == version {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"result":"hit","archiveLocation":%q,"cacheKey":%q}`,
					fs.server.URL+"/RUNID/_apis/artifactcache/artifacts/1", e.key)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (fs *fakeServer) reserve(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key       string `json:"key"`
		Version   string `json:"version"`
		CacheSize int64  `json:"cacheSize"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.nextID++
	fs.caches[fs.nextID] = &entry{key: req.Key, version: req.Version, size: req.CacheSize}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"cacheId":%d}`, fs.nextID)
}

func (fs *fakeServer) byID(w http.ResponseWriter, r *http.Request) {
	var id uint64
	if _, err := fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/RUNID/_apis/artifactcache/caches/"), "%d", &id); err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	fs.mu.Lock()
	e, ok := fs.caches[id]
	fs.mu.Unlock()
	if !ok {
		http.Error(w, "not reserved", http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		fs.mu.Lock()
		fs.contentRange = r.Header.Get("Content-Range")
		fs.mu.Unlock()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fs.mu.Lock()
		e.data = data
		fs.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	case http.MethodPost: // commit
		fs.mu.Lock()
		defer fs.mu.Unlock()
		if int64(len(e.data)) != e.size {
			http.Error(w, fmt.Sprintf("broken file: %d != %d", len(e.data), e.size), http.StatusInternalServerError)
			return
		}
		e.complete = true
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "method", http.StatusMethodNotAllowed)
	}
}

func (fs *fakeServer) serve(w http.ResponseWriter, r *http.Request) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	// the fake has at most one complete entry; good enough for roundtrips
	for _, e := range fs.caches {
		if e.complete {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(e.data)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func TestNewFromEnvUnset(t *testing.T) {
	t.Setenv("ACTIONS_CACHE_URL", "")
	if c := NewFromEnv(slog.New(slog.NewTextHandler(io.Discard, nil))); c != nil {
		t.Fatalf("NewFromEnv with empty ACTIONS_CACHE_URL = %v, want nil", c)
	}
}

func TestSaveLookupDownload(t *testing.T) {
	fs := newFakeServer(t)
	c := fs.client(t)
	if c == nil {
		t.Fatal("NewFromEnv returned nil with ACTIONS_CACHE_URL set")
	}

	payload := []byte("pretend this is a go toolchain tarball")
	dir := t.TempDir()
	archive := filepath.Join(dir, "go1.99.9.linux-amd64.tar.gz")
	if err := os.WriteFile(archive, payload, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, ok, err := c.Lookup("go-setup-go1.99.9.linux-amd64.tar.gz", "abc123"); err != nil || ok {
		t.Fatalf("Lookup before save = (%v, %v, %v), want miss without error", nil, ok, err)
	}

	if err := c.Save("go-setup-go1.99.9.linux-amd64.tar.gz", "abc123", archive); err != nil {
		t.Fatalf("Save: %v", err)
	}

	location, ok, err := c.Lookup("GO-SETUP-go1.99.9.linux-amd64.tar.gz", "abc123") // case-insensitive
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if !ok {
		t.Fatal("Lookup after save missed")
	}

	got := filepath.Join(dir, "downloaded.tar.gz")
	if err := c.Download(location, got, ""); err != nil {
		t.Fatalf("Download: %v", err)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(payload) {
		t.Fatalf("downloaded %q, want %q", data, payload)
	}

	fs.mu.Lock()
	cr := fs.contentRange
	fs.mu.Unlock()
	if want := fmt.Sprintf("bytes 0-%d/*", len(payload)-1); cr != want {
		t.Fatalf("Content-Range = %q, want %q", cr, want)
	}
}

func TestLookupVersionMismatch(t *testing.T) {
	fs := newFakeServer(t)
	c := fs.client(t)

	payload := []byte("x")
	archive := filepath.Join(t.TempDir(), "a.tar.gz")
	if err := os.WriteFile(archive, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Save("key", "v1", archive); err != nil {
		t.Fatalf("Save: %v", err)
	}
	_, ok, err := c.Lookup("key", "v2")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if ok {
		t.Fatal("Lookup with different version hit, want miss")
	}
}

func TestSaveEmptyFile(t *testing.T) {
	fs := newFakeServer(t)
	c := fs.client(t)
	empty := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Save("k", "v", empty); err == nil {
		t.Fatal("Save of empty file succeeded, want error")
	}
}

func TestRedact(t *testing.T) {
	u := mustURL(t, "http://127.0.0.1:8088/abc123def/_apis/artifactcache/cache?keys=k")
	got := redact(u)
	want := "http://127.0.0.1:8088/<run>/_apis/artifactcache/cache?keys=k"
	if got != want {
		t.Fatalf("redact = %q, want %q", got, want)
	}
	if s := redact(mustURL(t, "https://go.dev/dl/?mode=json")); s != "https://go.dev/dl/?mode=json" {
		t.Fatalf("redact of unrelated URL = %q", s)
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
