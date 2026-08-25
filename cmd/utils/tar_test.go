package utils

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestTarDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	nested := filepath.Join(dir, "sub", "file")
	if err := os.WriteFile(nested, []byte("hello"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Chmod(nested, 0o640); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Symlink("sub/file", filepath.Join(dir, "link")); err != nil {
		t.Fatalf("setup: %v", err)
	}

	headers, contents := readTar(t, dir)

	if got := slices.Sorted(maps.Keys(headers)); !slices.Equal(got, []string{"link", "sub/", "sub/file"}) {
		t.Fatalf("entries: got %v, want [link sub/ sub/file]", got)
	}
	if got := contents["sub/file"]; got != "hello" {
		t.Errorf("sub/file content: got %q, want %q", got, "hello")
	}
	if got := headers["sub/file"].Mode; got != 0o640 {
		t.Errorf("sub/file mode: got %o, want %o", got, 0o640)
	}
	if got := headers["sub/"].Typeflag; got != tar.TypeDir {
		t.Errorf("sub/ typeflag: got %c, want %c", got, tar.TypeDir)
	}
	if got := headers["link"]; got.Typeflag != tar.TypeSymlink || got.Linkname != "sub/file" {
		t.Errorf("link: got typeflag %c linkname %q, want %c and %q", got.Typeflag, got.Linkname, tar.TypeSymlink, "sub/file")
	}
}

func TestTarDirectoryMissingPath(t *testing.T) {
	if _, err := TarDirectory(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("got nil, want an error")
	}
}

func readTar(t *testing.T, dir string) (map[string]*tar.Header, map[string]string) {
	t.Helper()
	data, err := TarDirectory(dir)
	if err != nil {
		t.Fatalf("TarDirectory: %v", err)
	}

	headers := map[string]*tar.Header{}
	contents := map[string]string{}
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return headers, contents
		}
		if err != nil {
			t.Fatalf("next entry: %v", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read %s: %v", header.Name, err)
		}
		headers[header.Name] = header
		contents[header.Name] = string(body)
	}
}
