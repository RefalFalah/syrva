package cmd

import (
	"fmt"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/spf13/cobra"
)

func (a *app) editCommand() *cobra.Command {
	return &cobra.Command{
		Use: "edit <host>", Short: "Edit server; Enter mempertahankan value lama", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, hosts, err := a.load()
			if err != nil {
				return err
			}
			current, err := host.Find(hosts, args[0])
			if err != nil {
				return err
			}
			updated, err := readHost(cmd, current)
			if err != nil {
				return err
			}
			if err := host.Update(hosts, args[0], updated); err != nil {
				return err
			}
			if err := store.Save(hosts); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Server %q diperbarui.\n", updated.Alias)
			return nil
		},
	}
}
