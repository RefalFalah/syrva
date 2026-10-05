package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/spf13/cobra"
)

func (a *app) listCommand() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "Tampilkan daftar server lokal", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, hosts, err := a.load()
			if err != nil {
				return err
			}
			if len(hosts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Belum ada server. Gunakan syrva add atau syrva import.")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tHOST\tUSER\tTAGS")
			for _, h := range host.Sorted(hosts) {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", h.Alias, h.Hostname, h.User, strings.Join(h.Tags, ", "))
			}
			return w.Flush()
		},
	}
}
