package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
)

func TestImportSelectionAndDuplicates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(t.TempDir(), "ssh config")
	if err := os.WriteFile(path, []byte("Host one two\nHostName example.com\nUser root\nHost *\nPort 2222\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, _ := config.Open(dir)
	if err := s.Save(map[string]host.Host{"one": {Hostname: "original.example.com", User: "deploy", Port: 22}}); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, dir, "two\ny\n", "import", "--file", path)
	if err != nil || !strings.Contains(out, "tidak ditimpa") || !strings.Contains(out, "1 server diimpor") {
		t.Fatalf("import = %q, %v", out, err)
	}
	hosts, err := s.Load()
	if err != nil || len(hosts) != 2 || hosts["one"].Hostname != "original.example.com" || hosts["two"].Port != 2222 {
		t.Fatalf("imported = %#v, %v", hosts, err)
	}
}

func TestImportCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte("Host dev\nUser root\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"\nn\n", "-\n", "\n"} {
		dir := t.TempDir()
		_, _ = execute(t, dir, input, "import", "--file", path)
		s, _ := config.Open(dir)
		hosts, err := s.Load()
		if err != nil || len(hosts) != 0 {
			t.Fatal("canceled import saved hosts")
		}
	}
}
