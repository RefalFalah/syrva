package cmd

import (
	"fmt"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/spf13/cobra"
)

func (a *app) removeCommand() *cobra.Command {
	return &cobra.Command{
		Use: "remove <host>", Short: "Hapus server setelah konfirmasi", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, hosts, err := a.load()
			if err != nil {
				return err
			}
			if _, err := host.Find(hosts, args[0]); err != nil {
				return err
			}
			confirmed, err := newPrompt(cmd).confirm(fmt.Sprintf("Hapus server %q?", args[0]), false)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Dibatalkan.")
				return nil
			}
			if err := host.Remove(hosts, args[0]); err != nil {
				return err
			}
			if err := store.Save(hosts); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Server %q dihapus.\n", args[0])
			return nil
		},
	}
}
