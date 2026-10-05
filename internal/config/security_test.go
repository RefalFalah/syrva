package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLErrorDoesNotEchoValues(t *testing.T) {
	s, _ := Open(t.TempDir())
	content := "hosts:\n  dev: {host: localhost, user: root, port: accidental-secret-value}\n"
	if err := os.WriteFile(filepath.Join(s.Dir, "hosts.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load()
	if err == nil || strings.Contains(err.Error(), "accidental-secret-value") || !strings.Contains(err.Error(), "baris") {
		t.Fatalf("unsafe YAML error: %v", err)
	}
}

func TestYAMLAnchorsKeepDefaultPort(t *testing.T) {
	for _, content := range []string{
		"hosts:\n  first: &base {host: localhost, user: root}\n  second: *base\n",
		"hosts:\n  first: &base {host: localhost, user: root}\n  second: {<<: *base, host: example.com}\n",
	} {
		s, _ := Open(t.TempDir())
		if err := os.WriteFile(filepath.Join(s.Dir, "hosts.yaml"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		hosts, err := s.Load()
		if err != nil || hosts["second"].Port != 22 {
			t.Fatalf("anchor = %#v, %v", hosts, err)
		}
	}
}

func TestSettingsValidation(t *testing.T) {
	for _, content := range []string{"version: 9\n", "unknown: true\n", "version: [\n"} {
		s, _ := Open(t.TempDir())
		if err := os.WriteFile(filepath.Join(s.Dir, "config.yaml"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Load(); err == nil {
			t.Fatalf("accepted settings %q", content)
		}
	}
}
