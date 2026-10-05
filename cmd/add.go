package cmd

import (
	"fmt"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/spf13/cobra"
)

func (a *app) addCommand() *cobra.Command {
	return &cobra.Command{
		Use: "add", Short: "Tambahkan server secara interaktif", Args: cobra.NoArgs,
		RunE: a.addHost,
	}
}

func (a *app) addHost(cmd *cobra.Command, _ []string) error {
	store, hosts, err := a.load()
	if err != nil {
		return err
	}
	h, err := readHost(cmd, host.Host{Port: 22})
	if err != nil {
		return err
	}
	if err := host.Add(hosts, h); err != nil {
		return err
	}
	if err := store.Save(hosts); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Server %q ditambahkan.\n", h.Alias)
	return nil
}
