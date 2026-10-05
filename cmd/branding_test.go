package cmd

import (
	"strings"
	"testing"
)

func TestSyrvaBranding(t *testing.T) {
	for _, flag := range []string{"--help", "--version"} {
		output, err := execute(t, t.TempDir(), "", flag)
		if err != nil || !strings.Contains(strings.ToLower(output), "syrva") {
			t.Fatalf("%s = %q, %v", flag, output, err)
		}
		if strings.Contains(strings.ToLower(output), "serva") {
			t.Fatalf("old branding in %s: %q", flag, output)
		}
	}
}
