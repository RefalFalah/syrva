package host

import "testing"

func TestValidation(t *testing.T) {
	base := Host{Alias: "dev", Hostname: "localhost", User: "root", Port: 22}
	if err := Validate(base); err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{"", "-oProxyCommand=x", "a b", "add", "help", "a/b"} {
		if ValidateAlias(alias) == nil {
			t.Errorf("accepted alias %q", alias)
		}
	}
	for _, hostname := range []string{"", "-oSomething", "root@host", "host;cmd", "a\nb", "[::1", "::1]", "127.0.0.1%eth0"} {
		if ValidateHostname(hostname) == nil {
			t.Errorf("accepted hostname %q", hostname)
		}
	}
	for _, hostname := range []string{"dev.example.com", "127.0.0.1", "::1", "[::1]", "fe80::1%eth0", "[fe80::1%eth0]"} {
		if err := ValidateHostname(hostname); err != nil {
			t.Error(err)
		}
	}
	for _, port := range []int{-1, 0, 65536} {
		if ValidatePort(port) == nil {
			t.Errorf("accepted port %d", port)
		}
	}
}
