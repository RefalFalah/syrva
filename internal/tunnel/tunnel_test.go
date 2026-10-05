package tunnel

import (
	"reflect"
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func TestTunnelArgs(t *testing.T) {
	h := host.Host{Alias: "dev", Hostname: "localhost", User: "root", Port: 2222}
	tun := host.Tunnel{LocalPort: 3307, RemoteHost: "127.0.0.1", RemotePort: 3306}
	got, err := Args(h, tun)
	want := []string{"-p", "2222", "-l", "root", "-N", "-o", "ExitOnForwardFailure=yes", "-L", "127.0.0.1:3307:127.0.0.1:3306", "--", "localhost"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Args = %#v, %v", got, err)
	}
	tun.RemoteHost = "::1"
	got, err = Args(h, tun)
	if err != nil || got[8] != "127.0.0.1:3307:[::1]:3306" {
		t.Fatalf("IPv6 = %#v, %v", got, err)
	}
	tun.LocalPort = 0
	if _, err := Args(h, tun); err == nil {
		t.Fatal("invalid local port accepted")
	}
}
