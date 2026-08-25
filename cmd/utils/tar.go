package utils

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// TarDirectory packs dir into an uncompressed tar whose entry names are relative to dir.
func TarDirectory(dir string) ([]byte, error) {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		return writeTarEntry(tw, path, filepath.ToSlash(name), entry)
	})
	if err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeTarEntry(tw *tar.Writer, path, name string, entry fs.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSocket != 0 {
		return nil
	}

	link := ""
	if info.Mode()&fs.ModeSymlink != 0 {
		if link, err = os.Readlink(path); err != nil {
			return err
		}
	}
	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	header.Name = name
	if entry.IsDir() {
		header.Name += "/"
	}
	if err = tw.WriteHeader(header); err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}

	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(tw, f)
	return err
}
