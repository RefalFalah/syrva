package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func TestLoadCreatesEmptyConfig(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "nested", "syrva"))
	hosts, err := s.Load()
	if err != nil || len(hosts) != 0 || hosts == nil {
		t.Fatalf("Load = %v, %v", hosts, err)
	}
	for _, name := range []string{"config.yaml", "hosts.yaml"} {
		info, err := os.Stat(filepath.Join(s.Dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Fatalf("insecure permissions: %v", info.Mode())
		}
	}
}

func TestSaveRoundTrip(t *testing.T) {
	s, _ := Open(t.TempDir())
	want := map[string]host.Host{"dev": {
		Alias: "dev", Name: "Development", Hostname: "dev.example.com", User: "root", Port: 2222,
		IdentityFile: "~/.ssh/id_ed25519", Tags: []string{"dev", "web"}, WorkingDirectory: "/var/www/app",
		Commands: map[string]host.Command{"logs": {Description: "Logs", Command: "tail -f app.log"}},
		Tunnels:  map[string]host.Tunnel{"db": {LocalPort: 3307, RemoteHost: "127.0.0.1", RemotePort: 3306}},
	}}
	for i := 0; i < 2; i++ {
		if err := s.Save(want); err != nil {
			t.Fatal(err)
		}
		got, err := s.Load()
		if err != nil || !reflect.DeepEqual(want, got) {
			t.Fatalf("round trip = %#v, %v", got, err)
		}
	}
}

func TestLoadValidation(t *testing.T) {
	for _, tt := range []struct {
		name, content string
		valid         bool
	}{
		{"empty", "", true},
		{"null", "hosts: null\n", true},
		{"default port", "hosts:\n  dev:\n    host: localhost\n    user: root\n", true},
		{"zero port", "hosts:\n  dev: {host: localhost, user: root, port: 0}\n", false},
		{"invalid port", "hosts:\n  dev: {host: localhost, user: root, port: 65536}\n", false},
		{"missing user", "hosts:\n  dev: {host: localhost}\n", false},
		{"reserved alias", "hosts:\n  list: {host: localhost, user: root}\n", false},
		{"password", "hosts:\n  dev: {host: localhost, user: root, password: secret}\n", false},
		{"bad yaml", "hosts: [\n", false},
		{"multiple docs", "hosts: {}\n---\nhosts: {}\n", false},
		{"duplicate", "hosts:\n  dev: {}\n  dev: {}\n", false},
		{"null host", "hosts:\n  dev: null\n", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := Open(t.TempDir())
			if err := os.WriteFile(filepath.Join(s.Dir, "hosts.yaml"), []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			hosts, err := s.Load()
			if (err == nil) != tt.valid {
				t.Fatalf("Load: %v", err)
			}
			if tt.name == "default port" && hosts["dev"].Port != 22 {
				t.Fatal("missing default port")
			}
		})
	}
}

func TestInvalidSavePreservesConfig(t *testing.T) {
	s, _ := Open(t.TempDir())
	if err := s.Save(nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.Dir, "hosts.yaml")
	before, _ := os.ReadFile(path)
	if err := s.Save(map[string]host.Host{"bad alias": {}}); err == nil {
		t.Fatal("expected validation error")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("invalid save changed existing file")
	}
}

func TestDefaultDir(t *testing.T) {
	// A literal tilde is valid in a path, including Windows short names used
	// by hosted runners. Only a relative path suggests an unexpanded home.
	for _, name := range []string{"config", "RUNNER~1"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join(t.TempDir(), name)
			switch runtime.GOOS {
			case "windows":
				t.Setenv("APPDATA", base)
			case "darwin":
				t.Setenv("HOME", base)
				base = filepath.Join(base, "Library", "Application Support")
			default:
				t.Setenv("XDG_CONFIG_HOME", base)
			}
			got, err := DefaultDir()
			if err != nil || got != filepath.Join(base, "syrva") {
				t.Fatalf("DefaultDir = %q, %v", got, err)
			}
			if !filepath.IsAbs(got) {
				t.Fatalf("config directory must be absolute: %q", got)
			}
		})
	}
}

func TestDocumentedExample(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "hosts.example.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	s, _ := Open(t.TempDir())
	if err := os.WriteFile(filepath.Join(s.Dir, "hosts.yaml"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	hosts, err := s.Load()
	if err != nil || len(hosts) != 1 || hosts["websku-dev"].Commands["logs"].Command == "" || hosts["websku-dev"].Tunnels["mysql"].LocalPort != 3307 {
		t.Fatalf("example config = %#v, %v", hosts, err)
	}
}
