package cmd

import (
	"strings"
	"testing"

	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
)

func TestPresetListingIsLocal(t *testing.T) {
	dir := t.TempDir()
	s, _ := config.Open(dir)
	if err := s.Save(map[string]host.Host{"dev": {
		Hostname: "localhost", User: "root", Port: 22,
		Commands: map[string]host.Command{"logs": {Description: "Application logs", Command: "tail -f app.log"}},
	}}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	out, err := execute(t, dir, "", "dev", "run")
	if err != nil || !strings.Contains(out, "Application logs") || !strings.Contains(out, "logs") {
		t.Fatalf("list preset = %q, %v", out, err)
	}
	_, err = execute(t, dir, "", "dev", "run", "unknown")
	if err == nil || !strings.Contains(err.Error(), "logs") {
		t.Fatalf("unknown preset = %v", err)
	}
}
