package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
)

func TestHostCommandUsage(t *testing.T) {
	dir := t.TempDir()
	s, _ := config.Open(dir)
	if err := s.Save(map[string]host.Host{"dev": {Hostname: "localhost", User: "root", Port: 22}}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"dev", "unknown"}, {"dev", "info", "extra"}, {"dev", "run", "one", "two"},
		{"dev", "upload"}, {"dev", "download"}, {"dev", "tunnel"}, {"dev", "tunnel", "missing"},
	} {
		if _, err := execute(t, dir, "", args...); err == nil {
			t.Fatalf("accepted invalid args %v", args)
		}
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := execute(t, dir, "", "dev", "info"); err != nil {
		t.Fatal("info unexpectedly required an executable:", err)
	}
	if _, err := execute(t, dir, "", "dev"); err == nil || !strings.Contains(err.Error(), "OpenSSH tidak ditemukan") {
		t.Fatalf("missing SSH = %v", err)
	}
}

func TestCanceledPrompt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// A pre-canceled CLI must return without starting a blocking input read.
	root := NewRoot()
	root.SetIn(strings.NewReader(""))
	root.SetOut(&lockedBuffer{})
	root.SetArgs([]string{"--config-dir", t.TempDir(), "add"})
	if err := root.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled prompt = %v", err)
	}
}
