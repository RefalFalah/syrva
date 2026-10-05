package ssh

import (
	"strings"
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func FuzzParseConfig(f *testing.F) {
	for _, seed := range []string{
		"", "Host dev\nHostName localhost\nUser root\n", "Host * !github\nUser deploy\n",
		"Host one two\r\nIdentityFile \"C:\\My Keys\\id_ed25519\"\r\n", "Match exec ignored\nPort 9\n",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		hosts, _, err := ParseConfig(strings.NewReader(input), "fallback")
		if err != nil {
			return
		}
		for _, h := range hosts {
			if err := host.Validate(h); err != nil {
				t.Fatalf("parser returned invalid host: %v", err)
			}
		}
	})
}
