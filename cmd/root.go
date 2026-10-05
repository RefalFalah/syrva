package cmd

import (
	"github.com/RefalFalah/syrva/internal/config"
	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/tui"
	"github.com/spf13/cobra"
)

const Version = "0.1.0"

type app struct {
	configDir string
}

func (a *app) load() (*config.Store, map[string]host.Host, error) {
	store, err := config.Open(a.configDir)
	if err != nil {
		return nil, nil, err
	}
	hosts, err := store.Load()
	return store, hosts, err
}

// NewRoot returns an independent command tree, including its config location.
func NewRoot() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:   "syrva [host]",
		Short: "Syrva — terminal-first server management.",
		Long: `Syrva — terminal-first server management.

Tanpa argumen, buka TUI server selector. Alias host langsung membuka native SSH.

Command host:
  syrva <host>                         Connect SSH interaktif
  syrva <host> info                    Informasi lokal
  syrva <host> run [preset]            Daftar/jalankan preset
  syrva <host> upload <local> <remote> Upload file via scp
  syrva <host> download <remote> <local>
  syrva <host> tunnel <name>           Tunnel foreground (Ctrl+C berhenti)`,
		Example:       "  syrva add\n  syrva websku-dev\n  syrva websku-dev run logs\n  syrva websku-dev tunnel mysql",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, hosts, err := a.load()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				selection, err := tui.Select(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), host.Sorted(hosts))
				if err != nil {
					return err
				}
				// Bubble Tea restores the terminal before prompts or native SSH take over.
				if selection.Add {
					return a.addHost(cmd, nil)
				}
				if selection.Alias == "" {
					return nil
				}
				args = []string{selection.Alias}
			}
			h, err := host.Find(hosts, args[0])
			if err != nil {
				return err
			}
			return a.hostCommand(cmd, h, args[1:])
		},
	}
	root.PersistentFlags().StringVar(&a.configDir, "config-dir", "", "Direktori konfigurasi (default: user config directory/syrva)")
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(a.listCommand(), a.addCommand(), a.editCommand(), a.removeCommand(), a.importCommand())
	return root
}
