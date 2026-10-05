package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
	"github.com/spf13/cobra"
)

func runPreset(cmd *cobra.Command, h host.Host, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("gunakan: syrva %s run [preset]", h.Alias)
	}
	if len(args) == 0 {
		if len(h.Commands) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Belum ada preset untuk %q. Tambahkan commands di hosts.yaml.\n", h.Alias)
			return nil
		}
		fmt.Fprint(cmd.OutOrStdout(), "Command tersedia:\n\n")
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 3, ' ', 0)
		for _, name := range sortedKeys(h.Commands) {
			fmt.Fprintf(w, "%s\t%s\n", name, h.Commands[name].Description)
		}
		return w.Flush()
	}
	preset, exists := h.Commands[args[0]]
	if !exists {
		return fmt.Errorf("preset %q tidak ditemukan pada server %q.\nCommand tersedia: %s", args[0], h.Alias, strings.Join(sortedKeys(h.Commands), ", "))
	}
	sshArgs, err := ssh.RunArgs(h, preset)
	if err != nil {
		return err
	}
	return executor(cmd).Run(cmd.Context(), "ssh", sshArgs)
}
