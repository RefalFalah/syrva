package ssh

import (
	"fmt"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
)

// QuoteRemote protects a literal path in a POSIX remote shell. It is never used
// for a local shell: presets intentionally remain user-authored remote commands.
func QuoteRemote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func RunArgs(h host.Host, preset host.Command) ([]string, error) {
	if strings.TrimSpace(preset.Command) == "" || strings.ContainsRune(preset.Command, '\x00') {
		return nil, fmt.Errorf("command preset kosong atau tidak valid")
	}
	args, err := ConnectArgs(h)
	if err != nil {
		return nil, err
	}
	remote := preset.Command
	if h.WorkingDirectory != "" {
		remote = "cd -- " + QuoteRemote(h.WorkingDirectory) + " && " + remote
	}
	return append(args, remote), nil
}
