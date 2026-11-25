// Package dl provides small download helpers with checksum
// verification for the setup-* actions.
package dl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client is used for all downloads; the timeout is generous because
// toolchain archives are large.
var Client = &http.Client{Timeout: 30 * time.Minute}

const userAgent = "forgejo-pure-actions/1.0 (+go)"

// FetchJSON GETs url and decodes the JSON body into out.
func FetchJSON(url string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Get issues a simple GET, returning the body, so callers can add
// headers such as Authorization.
func Get(url string, header map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// DownloadFile streams url into dst, verifying expectedSHA256 (hex,
// optional) along the way. It writes a temporary sibling first and
// renames it into place so failures never leave partial files.
func DownloadFile(url, dst, expectedSHA256 string, log *slog.Logger) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	part := dst + ".part"
	f, err := os.Create(part)
	if err != nil {
		return err
	}
	h := sha256.New()
	pr := &progressReader{r: resp.Body, every: 16 << 20, log: log, url: url}
	_, err = io.Copy(io.MultiWriter(f, h), pr)
	f.Close()
	if err != nil {
		os.Remove(part)
		return fmt.Errorf("download %s: %w", url, err)
	}
	if expectedSHA256 != "" {
		got := hex.EncodeToString(h.Sum(nil))
		if !strings.EqualFold(got, expectedSHA256) {
			os.Remove(part)
			return fmt.Errorf("checksum mismatch for %s: got %s want %s", url, got, expectedSHA256)
		}
	}
	return os.Rename(part, dst)
}

type progressReader struct {
	r     io.Reader
	every int64
	seen  int64
	total int64
	log   *slog.Logger
	url   string
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.total += int64(n)
	if p.log != nil && p.total-p.seen >= p.every {
		p.seen = p.total
		p.log.Info(fmt.Sprintf("downloaded %d MiB ...", p.total>>20))
	}
	return n, err
}
