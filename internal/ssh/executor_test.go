package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// The helper is a local test process, never an SSH server or a shell.
func TestExecutorHelper(t *testing.T) {
	if os.Getenv("SYRVA_EXEC_HELPER") != "1" {
		return
	}
	if os.Args[len(os.Args)-1] == "fail" {
		os.Exit(7)
	}
	data, _ := io.ReadAll(os.Stdin)
	fmt.Fprint(os.Stdout, string(data))
	fmt.Fprint(os.Stderr, "helper stderr")
	os.Exit(0)
}

func TestExecutorInteractiveIO(t *testing.T) {
	t.Setenv("SYRVA_EXEC_HELPER", "1")
	path, _ := os.Executable()
	var out, stderr bytes.Buffer
	runner := Executor{Stdin: strings.NewReader("interactive input"), Stdout: &out, Stderr: &stderr}
	err := runner.Run(context.Background(), path, []string{"-test.run=^TestExecutorHelper$", "--", "echo"})
	if err != nil || out.String() != "interactive input" || stderr.String() != "helper stderr" {
		t.Fatalf("Run = %v, %q, %q", err, out.String(), stderr.String())
	}
	err = runner.Run(context.Background(), path, []string{"-test.run=^TestExecutorHelper$", "--", "fail"})
	var exit *ExitError
	if !errors.As(err, &exit) || exit.Code != 7 {
		t.Fatalf("exit = %v", err)
	}
}

func TestMissingExecutables(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, program := range []string{"ssh", "scp"} {
		err := (Executor{}).Run(context.Background(), program, nil)
		if err == nil || !strings.Contains(err.Error(), "tidak ditemukan") || !strings.Contains(err.Error(), "PATH") {
			t.Fatalf("missing %s = %v", program, err)
		}
	}
}

func TestCanceledExecutor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Executor{}).Run(ctx, "ssh", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
}
