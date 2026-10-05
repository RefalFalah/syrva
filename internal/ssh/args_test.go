package ssh

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func TestConnectArgs(t *testing.T) {
	h := host.Host{Alias: "dev", Hostname: "dev.example.com", User: "root", Port: 2222, IdentityFile: "~/.ssh/key with space"}
	args, err := ConnectArgs(h)
	home, _ := os.UserHomeDir()
	want := []string{"-p", "2222", "-l", "root", "-i", filepath.Join(home, ".ssh", "key with space"), "--", "dev.example.com"}
	if err != nil || !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, %v", args, err)
	}
	h.IdentityFile = ""
	h.Hostname = "[::1]"
	args, err = ConnectArgs(h)
	if err != nil || args[len(args)-1] != "::1" || len(args) != 6 {
		t.Fatalf("IPv6 args = %v, %v", args, err)
	}
	h.Hostname = "-oProxyCommand=evil"
	if _, err := ConnectArgs(h); err == nil {
		t.Fatal("option injection accepted")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	for _, path := range []string{"~/.ssh/key", `~\.ssh/key`} {
		got, err := ExpandPath(path)
		if err != nil || got != filepath.Join(home, ".ssh", "key") {
			t.Fatalf("ExpandPath(%q) = %q, %v", path, got, err)
		}
	}
	if _, err := ExpandPath("~other/.ssh/key"); err == nil {
		t.Fatal("~user accepted")
	}
}
