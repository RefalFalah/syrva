//go:build !windows

package ssh

import (
	"os"
	"os/exec"
)

func interrupt(cmd *exec.Cmd) error {
	return cmd.Process.Signal(os.Interrupt)
}
