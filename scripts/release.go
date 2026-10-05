// Run from the repository root: go run ./scripts/release.go
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/RefalFalah/syrva/cmd"
)

type entry struct {
	name, path string
	mode       int64
}

func main() {
	if err := release(); err != nil {
		fmt.Fprintln(os.Stderr, "Release build gagal:", err)
		os.Exit(1)
	}
}

func release() error {
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return err
	}
	temp, err := os.MkdirTemp("dist", ".build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp) // Only the temporary directory created by this run.
	notices := filepath.Join(temp, "THIRD_PARTY_NOTICES.txt")
	if err := writeNotices(notices); err != nil {
		return err
	}
	var archives []string
	for _, target := range []string{"windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"} {
		parts := strings.Split(target, "/")
		name := "syrva"
		isWindows := parts[0] == "windows"
		if isWindows {
			name += ".exe"
		}
		binary := filepath.Join(temp, parts[0]+"-"+parts[1], name)
		build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", binary, ".")
		build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+parts[0], "GOARCH="+parts[1])
		build.Stdout, build.Stderr = os.Stdout, os.Stderr
		fmt.Println("Building", target)
		if err := build.Run(); err != nil {
			return fmt.Errorf("%s: %w", target, err)
		}
		files := []entry{
			{name, binary, 0o755}, {"README.md", "README.md", 0o644}, {"LICENSE", "LICENSE", 0o644},
			{"THIRD_PARTY_NOTICES.txt", notices, 0o644},
			{"examples/hosts.example.yaml", "examples/hosts.example.yaml", 0o644},
		}
		suffix := ".tar.gz"
		if isWindows {
			suffix = ".zip"
		}
		archive := fmt.Sprintf("syrva_%s_%s_%s%s", cmd.Version, parts[0], parts[1], suffix)
		if err := writeArchive(filepath.Join("dist", archive), files, isWindows); err != nil {
			return err
		}
		archives = append(archives, archive)
	}
	sort.Strings(archives)
	var checksums strings.Builder
	for _, name := range archives {
		f, err := os.Open(filepath.Join("dist", name))
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, f)
		if err := errors.Join(copyErr, f.Close()); err != nil {
			return err
		}
		fmt.Fprintf(&checksums, "%x  %s\n", hash.Sum(nil), name)
	}
	if err := os.WriteFile(filepath.Join("dist", "checksums.txt"), []byte(checksums.String()), 0o644); err != nil {
		return err
	}
	fmt.Println("Arsip dan SHA-256 tersedia di dist/ (belum dipublikasikan).")
	return nil
}

func writeNotices(path string) error {
	type module struct {
		Path, Version, Dir, Error string
	}
	// Explicit requirements in go.mod include the runtime dependencies selected
	// by go mod tidy across platforms. Download their original license files.
	download := exec.Command("go", "mod", "download", "-json")
	download.Stderr = os.Stderr
	data, err := download.Output()
	if err != nil {
		return fmt.Errorf("tidak dapat menyiapkan lisensi dependency: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	var modules []module
	for {
		var m module
		if err := decoder.Decode(&m); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		if m.Error != "" || m.Dir == "" {
			return fmt.Errorf("dependency %s tidak dapat disiapkan", m.Path)
		}
		modules = append(modules, m)
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
	root, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	modules = append([]module{{Path: "Go standard library and runtime", Dir: strings.TrimSpace(string(root))}}, modules...)
	var notices strings.Builder
	notices.WriteString("Syrva includes the following third-party software.\nOriginal license and notice texts follow. OpenSSH is external and is not bundled.\n")
	for _, m := range modules {
		files, err := os.ReadDir(m.Dir)
		if err != nil {
			return err
		}
		found := false
		for _, file := range files {
			name := strings.ToUpper(file.Name())
			if file.IsDir() || !(strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(name, "LICENCE") || strings.HasPrefix(name, "COPYING") || strings.HasPrefix(name, "NOTICE")) {
				continue
			}
			content, err := os.ReadFile(filepath.Join(m.Dir, file.Name()))
			if err != nil {
				return err
			}
			fmt.Fprintf(&notices, "\n============================================================\n%s %s / %s\n============================================================\n%s\n", m.Path, m.Version, file.Name(), content)
			found = true
		}
		if !found {
			// This legacy version declares MIT in README but predates the upstream
			// standalone license file. Keep that original text with the source.
			if m.Path == "github.com/mattn/go-localereader" && m.Version == "v0.0.1" {
				content, err := os.ReadFile(filepath.Join("licenses", "go-localereader.LICENSE"))
				if err != nil {
					return err
				}
				fmt.Fprintf(&notices, "\n============================================================\n%s %s / upstream LICENSE\n============================================================\n%s\n", m.Path, m.Version, content)
			} else {
				return fmt.Errorf("lisensi dependency %s tidak ditemukan; periksa sebelum distribusi", m.Path)
			}
		}
	}
	return os.WriteFile(path, []byte(notices.String()), 0o644)
}

func writeArchive(path string, entries []entry, isZip bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if isZip {
		writer := zip.NewWriter(f)
		for _, file := range entries {
			data, err := os.ReadFile(file.path)
			if err != nil {
				return err
			}
			header := &zip.FileHeader{Name: file.name, Method: zip.Deflate}
			header.SetMode(os.FileMode(file.mode))
			entry, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}
			if _, err := entry.Write(data); err != nil {
				return err
			}
		}
		return errors.Join(writer.Close(), f.Close())
	}
	compressed := gzip.NewWriter(f)
	writer := tar.NewWriter(compressed)
	for _, file := range entries {
		data, err := os.ReadFile(file.path)
		if err != nil {
			return err
		}
		header := &tar.Header{Name: file.name, Mode: file.mode, Size: int64(len(data)), ModTime: time.Unix(0, 0)}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}
	return errors.Join(writer.Close(), compressed.Close(), f.Close())
}
