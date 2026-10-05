package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseArchives(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "binary")
	if err := os.WriteFile(source, []byte("test binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := []entry{{"syrva", source, 0o755}, {"examples/hosts.example.yaml", source, 0o644}}
	for _, isZip := range []bool{false, true} {
		path := filepath.Join(dir, "archive")
		if err := writeArchive(path, entries, isZip); err != nil {
			t.Fatal(err)
		}
		if isZip {
			reader, err := zip.OpenReader(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(reader.File) != 2 || reader.File[0].Name != "syrva" || reader.File[0].Mode().Perm() != 0o755 {
				t.Fatal("invalid ZIP contents")
			}
			_ = reader.Close()
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		compressed, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		reader := tar.NewReader(compressed)
		for _, entry := range entries {
			header, err := reader.Next()
			if err != nil || header.Name != entry.name || header.Mode != entry.mode {
				t.Fatalf("invalid TAR entry = %#v, %v", header, err)
			}
			data, err := io.ReadAll(reader)
			if err != nil || string(data) != "test binary" {
				t.Fatal("invalid TAR data")
			}
		}
		_ = compressed.Close()
		_ = f.Close()
	}
}
