package tarx

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildTarGz builds an in-memory tar.gz from name -> content entries;
// names ending in "/" are directories, entries starting with "link:"
// are symlinks.
func buildTarGz(t *testing.T, entries ...[2]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		name, content := e[0], e[1]
		if strings.HasSuffix(name, "/") {
			if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if target, ok := strings.CutPrefix(content, "link:"); ok {
			if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeSymlink, Linkname: target, Mode: 0o777}); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func TestExtractTarGzStripsTopDir(t *testing.T) {
	dest := t.TempDir()
	body := buildTarGz(t,
		[2]string{"go-1.2.3/"},
		[2]string{"go-1.2.3/bin", "link:../lib/go"},
		[2]string{"go-1.2.3/VERSION", "go1.2.3"},
		[2]string{"go-1.2.3/pkg/tool", "binary"},
	)
	if err := ExtractTarGz(bytes.NewReader(body), dest, 1); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "VERSION")); err != nil || string(b) != "go1.2.3" {
		t.Errorf("VERSION wrong: %q %v", b, err)
	}
	if fi, err := os.Lstat(filepath.Join(dest, "bin")); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("bin should be a symlink: %v %v", fi, err)
	}
}

func TestExtractRejectsTraversal(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "../escape", Typeflag: tar.TypeReg, Mode: 0o644, Size: 1}); err != nil {
		t.Fatal(err)
	}
	tw.Write([]byte("x"))
	tw.Close()
	dest := t.TempDir()
	if err := Extract(tar.NewReader(&buf), dest, 0); err == nil {
		t.Error("traversal attempt should be rejected")
	}
}
