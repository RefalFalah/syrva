package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
)

// These local executables stand in for ssh/scp. They record argv and inherit I/O,
// so the complete CLI route is tested without connecting to any server.
func TestMain(m *testing.M) {
	if os.Getenv("SYRVA_NATIVE_HELPER") == "1" {
		if os.Getenv("SYRVA_NATIVE_WAIT") == "1" {
			fmt.Fprintln(os.Stdout, "ready")
			for {
				time.Sleep(time.Hour)
			}
		}
		input, _ := io.ReadAll(os.Stdin)
		_ = json.NewEncoder(os.Stdout).Encode(struct {
			Args  []string
			Input string
		}{os.Args[1:], string(input)})
		fmt.Fprint(os.Stderr, "native helper stderr")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func nativeHelpers(t *testing.T) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	for _, name := range []string{"ssh", "scp"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if err := os.WriteFile(filepath.Join(bin, name), data, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("SYRVA_NATIVE_HELPER", "1")
}

func TestNativeHostCommands(t *testing.T) {
	nativeHelpers(t)
	dir := t.TempDir()
	s, _ := config.Open(dir)
	if err := s.Save(map[string]host.Host{"dev": {
		Hostname: "localhost", User: "root", Port: 2222, WorkingDirectory: "/var/www/app",
		Commands: map[string]host.Command{"logs": {Command: "tail -f app.log"}},
		Tunnels:  map[string]host.Tunnel{"mysql": {LocalPort: 3307, RemoteHost: "127.0.0.1", RemotePort: 3306}},
	}}); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(t.TempDir(), "backup file.zip")
	if err := os.WriteFile(local, []byte("local file"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"dev"}, `"-l","root"`},
		{[]string{"dev", "run", "logs"}, "cd -- '/var/www/app'"},
		{[]string{"dev", "upload", local, "/tmp/"}, "localhost:/tmp/"},
		{[]string{"dev", "download", "/tmp/file.zip", "."}, "localhost:/tmp/file.zip"},
		{[]string{"dev", "tunnel", "mysql"}, "127.0.0.1:3307:127.0.0.1:3306"},
	} {
		out, err := execute(t, dir, "inherited stdin", tt.args...)
		if err != nil || !strings.Contains(out, tt.want) || !strings.Contains(out, "inherited stdin") || !strings.Contains(out, "native helper stderr") {
			t.Fatalf("%v = %q, %v", tt.args, out, err)
		}
	}
}

type readyWriter struct{ ready chan struct{} }

func (w readyWriter) Write(p []byte) (int, error) {
	select {
	case <-w.ready:
	default:
		close(w.ready)
	}
	return len(p), nil
}

func TestRunningProcessCancellation(t *testing.T) {
	nativeHelpers(t)
	t.Setenv("SYRVA_NATIVE_WAIT", "1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- (ssh.Executor{Stdin: strings.NewReader(""), Stdout: readyWriter{ready}, Stderr: io.Discard}).Run(ctx, "ssh", nil)
	}()
	select {
	case <-ready:
		cancel()
	case <-time.After(10 * time.Second):
		t.Fatal("helper did not start")
	}
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("canceled child did not stop")
	}
}
