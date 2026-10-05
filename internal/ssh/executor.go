package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type Executor struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type ExitError struct {
	Program string
	Code    int
}

func (e *ExitError) Error() string {
	if e.Program == "ssh" && e.Code == 255 {
		return "Koneksi SSH gagal. Periksa hostname, jaringan, key, dan pesan OpenSSH di atas."
	}
	if e.Program == "scp" {
		return fmt.Sprintf("Transfer SCP gagal (exit code %d). Periksa path, izin, dan pesan di atas; gunakan OpenSSH 9+ dengan dukungan SFTP.", e.Code)
	}
	return fmt.Sprintf("%s berhenti dengan exit code %d.", e.Program, e.Code)
}

func (e Executor) Run(ctx context.Context, program string, args []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := exec.LookPath(program)
	if err != nil {
		switch program {
		case "ssh":
			return fmt.Errorf("OpenSSH tidak ditemukan.\n\nPastikan command \"ssh\" tersedia di PATH.")
		case "scp":
			return fmt.Errorf("SCP tidak ditemukan.\n\nPastikan command \"scp\" tersedia di PATH.")
		default:
			return fmt.Errorf("executable %q tidak ditemukan di PATH", program)
		}
	}
	// Execute an argv vector directly: no cmd.exe, PowerShell, or /bin/sh.
	process := exec.CommandContext(ctx, path, args...)
	process.Stdin, process.Stdout, process.Stderr = e.Stdin, e.Stdout, e.Stderr
	process.Cancel = func() error { return interrupt(process) }
	process.WaitDelay = 2 * time.Second
	err = process.Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err == nil {
		return nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return &ExitError{Program: program, Code: exit.ExitCode()}
	}
	return fmt.Errorf("tidak dapat menjalankan %s: %w", program, err)
}
