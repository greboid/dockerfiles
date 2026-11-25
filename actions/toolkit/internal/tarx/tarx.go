// Package tarx extracts (optionally gzipped) tar archives safely,
// with support for stripping leading path components.
package tarx

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTarGz extracts a gzipped tarball from r into dest. If
// stripComponents is greater than zero that many leading path
// components are removed from each entry (used to drop the versioned
// top-level directory of Go and Node tarballs).
func ExtractTarGz(r io.Reader, dest string, stripComponents int) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()
	return Extract(tar.NewReader(gz), dest, stripComponents)
}

// Extract extracts a tar stream into dest.
func Extract(tr *tar.Reader, dest string, strip int) error {
	dest = filepath.Clean(dest)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.ToSlash(filepath.Clean(hdr.Name))
		name = strings.TrimPrefix(name, "./")
		parts := strings.Split(name, "/")
		if len(parts) <= strip {
			continue // the stripped prefix itself (e.g. "go/")
		}
		rel := filepath.Join(parts[strip:]...)
		if rel == "" || rel == "." {
			continue
		}
		target := filepath.Join(dest, rel)
		if target != dest && !strings.HasPrefix(target, dest+string(os.PathSeparator)) {
			return fmt.Errorf("archive entry %q escapes destination", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(hdr.Mode) & 0o777
			if mode == 0 {
				mode = 0o644
			}
			if err := writeFile(target, tr, mode); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return fmt.Errorf("symlink %q: %w", hdr.Name, err)
			}
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			source := linkTarget(hdr.Linkname, dest, strip)
			os.Remove(target)
			if err := os.Link(source, target); err != nil {
				return fmt.Errorf("hardlink %q -> %q: %w", hdr.Name, hdr.Linkname, err)
			}
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			// metadata only
		default:
			return fmt.Errorf("unsupported tar entry type %q for %s", hdr.Typeflag, hdr.Name)
		}
	}
}

func linkTarget(linkname, dest string, strip int) string {
	name := filepath.ToSlash(filepath.Clean(linkname))
	parts := strings.Split(name, "/")
	if len(parts) > strip {
		return filepath.Join(append([]string{dest}, parts[strip:]...)...)
	}
	return filepath.Join(dest, name)
}

func writeFile(target string, r io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(target)
		return fmt.Errorf("write %s: %w", target, err)
	}
	return f.Close()
}
