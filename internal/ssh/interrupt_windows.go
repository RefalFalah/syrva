package ssh

import "os/exec"

func interrupt(_ *exec.Cmd) error {
	// The child inherits the console and receives Ctrl+C from Windows itself.
	// Do not call Process.Signal(os.Interrupt): Windows does not support it.
	// WaitDelay provides a bounded fallback if the child fails to stop.
	return nil
}
