package ssh

import (
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	input := `# sample
Host websku other
    HostName dev.example.com
    User root
    Port=2222
    IdentityFile "~/.ssh/key with space" # comment
Host github
    User git
Host * !github
    User default
    Port 22
    IdentityFile ~/.ssh/id_ed25519
`
	hosts, warnings, err := ParseConfig(strings.NewReader(input), "localuser")
	if err != nil || len(hosts) != 3 || len(warnings) != 0 {
		t.Fatalf("ParseConfig = %#v, %v, %v", hosts, warnings, err)
	}
	if hosts[0].Alias != "github" || hosts[0].User != "git" || hosts[0].IdentityFile != "" || hosts[0].Port != 22 {
		t.Fatalf("negated defaults: %#v", hosts[0])
	}
	if hosts[2].Alias != "websku" || hosts[2].Hostname != "dev.example.com" || hosts[2].Port != 2222 || hosts[2].IdentityFile != "~/.ssh/key with space" {
		t.Fatalf("host fields: %#v", hosts[2])
	}
}

func TestParserDefaultsAndFirstValue(t *testing.T) {
	input := "User global\nHost dev\n user root\n Port 2200\n Port 2222\nHost dev\n HostName DEV.example.com\nHost *\n Port 22\n"
	hosts, _, err := ParseConfig(strings.NewReader(input), "fallback")
	if err != nil || len(hosts) != 1 || hosts[0].User != "global" || hosts[0].Port != 2200 || hosts[0].Hostname != "DEV.example.com" {
		t.Fatalf("first values = %#v, %v", hosts, err)
	}
}

func TestParserUnsupportedAndInvalid(t *testing.T) {
	input := "Include config.d/*\nHost good\nUser root\nProxyJump gateway\nMatch exec dangerous\nHostName wrong\nHost bad\nPort 99999\nHost *\nUser default\nHost list\nUser root\n"
	hosts, warnings, err := ParseConfig(strings.NewReader(input), "fallback")
	if err != nil || len(hosts) != 1 || hosts[0].Alias != "good" || hosts[0].Hostname != "good" || len(warnings) < 4 {
		t.Fatalf("unsupported = %#v, %v, %v", hosts, warnings, err)
	}
	if strings.Contains(strings.Join(warnings, " "), "dangerous") {
		t.Fatal("raw Match command leaked")
	}
	for _, input := range []string{"Host\n", "Host dev\nUser root extra\n", "Host dev\nIdentityFile \"unfinished\n"} {
		if _, _, err := ParseConfig(strings.NewReader(input), "fallback"); err == nil {
			t.Errorf("accepted malformed %q", input)
		}
	}
}

func TestParserWindowsAndCRLF(t *testing.T) {
	input := "\ufeffhost win\r\n hostname=example.com\r\n user DOMAIN\\alice\r\n identityfile \"C:\\Users\\alice\\My Keys\\id_ed25519\"\r\n"
	hosts, _, err := ParseConfig(strings.NewReader(input), "fallback")
	if err != nil || len(hosts) != 1 || hosts[0].IdentityFile != `C:\Users\alice\My Keys\id_ed25519` || hosts[0].User != `DOMAIN\alice` {
		t.Fatalf("Windows config = %#v, %v", hosts, err)
	}
}
