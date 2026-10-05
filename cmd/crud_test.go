package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func execute(t *testing.T, dir, input string, args ...string) (string, error) {
	t.Helper()
	root := NewRoot()
	var output lockedBuffer
	root.SetIn(strings.NewReader(input))
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs(append([]string{"--config-dir", dir}, args...))
	err := root.Execute()
	return output.String(), err
}

func TestCRUDCommands(t *testing.T) {
	dir := t.TempDir()
	if _, err := execute(t, dir, "dev\nDevelopment\nlocalhost\nroot\n\n\ndev, web\n/var/www\n", "add"); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, dir, "", "dev", "info")
	if err != nil || !strings.Contains(out, "Development") || !strings.Contains(out, "22") {
		t.Fatalf("info = %q, %v", out, err)
	}
	if _, err := execute(t, dir, strings.Repeat("\n", 8), "edit", "dev"); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"\n", "n\n", ""} {
		_, _ = execute(t, dir, input, "remove", "dev")
		s, _ := config.Open(dir)
		hosts, err := s.Load()
		if err != nil || len(hosts) != 1 {
			t.Fatal("remove without explicit yes changed config")
		}
	}
	if _, err := execute(t, dir, "y\n", "remove", "dev"); err != nil {
		t.Fatal(err)
	}
	out, err = execute(t, dir, "", "list")
	if err != nil || !strings.Contains(out, "Belum ada server") {
		t.Fatalf("empty list = %q, %v", out, err)
	}
}

func TestAddEOFDoesNotSave(t *testing.T) {
	dir := t.TempDir()
	if _, err := execute(t, dir, "dev\n", "add"); err == nil {
		t.Fatal("incomplete form accepted")
	}
	s, _ := config.Open(dir)
	hosts, err := s.Load()
	if err != nil || len(hosts) != 0 {
		t.Fatal("incomplete form saved")
	}
}

func TestUnknownHostCommand(t *testing.T) {
	_, err := execute(t, t.TempDir(), "", "unknown", "info")
	if err == nil || !strings.Contains(err.Error(), `Server "unknown" tidak ditemukan`) {
		t.Fatalf("unknown host = %v", err)
	}
}

func TestEditRenameAndClearPreservesPresets(t *testing.T) {
	dir := t.TempDir()
	s, _ := config.Open(dir)
	original := host.Host{
		Hostname: "localhost", User: "root", Port: 22, Name: "Dev", IdentityFile: "~/.ssh/key",
		Tags: []string{"dev"}, WorkingDirectory: "/var/www",
		Commands: map[string]host.Command{"logs": {Command: "tail -f app.log"}},
		Tunnels:  map[string]host.Tunnel{"db": {LocalPort: 3307, RemoteHost: "localhost", RemotePort: 3306}},
	}
	if err := s.Save(map[string]host.Host{"dev": original}); err != nil {
		t.Fatal(err)
	}
	if _, err := execute(t, dir, "renamed\n-\n\n\n\n-\n-\n-\n", "edit", "dev"); err != nil {
		t.Fatal(err)
	}
	hosts, err := s.Load()
	h := hosts["renamed"]
	if err != nil || len(hosts) != 1 || h.Name != "" || h.IdentityFile != "" || h.WorkingDirectory != "" || len(h.Tags) != 0 || !reflect.DeepEqual(h.Commands, original.Commands) || !reflect.DeepEqual(h.Tunnels, original.Tunnels) {
		t.Fatalf("edited host = %#v, %v", h, err)
	}
}

func TestHelpDoesNotCreateConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not-created")
	out, err := execute(t, dir, "", "--help")
	if err != nil || !strings.Contains(out, "upload") || !strings.Contains(out, "tunnel") {
		t.Fatalf("help = %q, %v", out, err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("help created config")
	}
}
