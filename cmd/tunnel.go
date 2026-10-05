package cmd

import (
	"fmt"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/tunnel"
	"github.com/spf13/cobra"
)

func openTunnel(cmd *cobra.Command, h host.Host, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("gunakan: syrva %s tunnel <name>", h.Alias)
	}
	t, exists := h.Tunnels[args[0]]
	if !exists {
		return fmt.Errorf("tunnel %q tidak ditemukan pada server %q.\nTunnel tersedia: %s", args[0], h.Alias, strings.Join(sortedKeys(h.Tunnels), ", "))
	}
	sshArgs, err := tunnel.Args(h, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Membuka tunnel %s: 127.0.0.1:%d → %s:%d via %s\nCtrl+C untuk berhenti.\n", args[0], t.LocalPort, t.RemoteHost, t.RemotePort, h.Alias)
	return executor(cmd).Run(cmd.Context(), "ssh", sshArgs)
}
