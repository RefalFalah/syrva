package ssh

import (
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func TestRunArgs(t *testing.T) {
	h := host.Host{Alias: "dev", Hostname: "localhost", User: "root", Port: 22, WorkingDirectory: "/var/www/user's app; $(touch /tmp/no)"}
	args, err := RunArgs(h, host.Command{Command: "tail -f app.log"})
	want := "cd -- '/var/www/user'\"'\"'s app; $(touch /tmp/no)' && tail -f app.log"
	if err != nil || args[len(args)-1] != want {
		t.Fatalf("RunArgs = %#v, %v", args, err)
	}
	h.WorkingDirectory = ""
	args, err = RunArgs(h, host.Command{Command: "echo hello"})
	if err != nil || args[len(args)-1] != "echo hello" {
		t.Fatal("command changed without working directory")
	}
	if _, err := RunArgs(h, host.Command{}); err == nil {
		t.Fatal("empty command accepted")
	}
}
