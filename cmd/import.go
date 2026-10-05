package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
	"github.com/spf13/cobra"
)

func (a *app) importCommand() *cobra.Command {
	var source string
	command := &cobra.Command{
		Use: "import", Short: "Pilih dan impor host dari ~/.ssh/config", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if source == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("home directory tidak ditemukan: %w", err)
				}
				source = filepath.Join(home, ".ssh", "config")
			}
			path, err := ssh.ExpandPath(source)
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("SSH config %q tidak dapat dibaca: %w", path, err)
			}
			defaultUser := os.Getenv("USER")
			if current, err := user.Current(); err == nil {
				defaultUser = current.Username
			} else if value := os.Getenv("USERNAME"); value != "" {
				defaultUser = value
			}
			found, warnings, parseErr := ssh.ParseConfig(file, defaultUser)
			closeErr := file.Close()
			if parseErr != nil {
				return parseErr
			}
			if closeErr != nil {
				return closeErr
			}
			for _, warning := range warnings {
				fmt.Fprintln(cmd.OutOrStdout(), "Catatan:", warning)
			}
			if len(found) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Tidak ada host literal yang dapat diimpor. Host wildcard tidak dijadikan server.")
				return nil
			}
			store, hosts, err := a.load()
			if err != nil {
				return err
			}
			eligible := map[string]host.Host{}
			fmt.Fprint(cmd.OutOrStdout(), "SSH hosts ditemukan:\n\n")
			for _, h := range found {
				if _, exists := hosts[h.Alias]; exists {
					fmt.Fprintf(cmd.OutOrStdout(), "[ ] %s (sudah ada; tidak ditimpa)\n", h.Alias)
				} else {
					eligible[h.Alias] = h
					fmt.Fprintf(cmd.OutOrStdout(), "[✓] %s (%s@%s)\n", h.Alias, h.User, h.Hostname)
				}
			}
			if len(eligible) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Semua host sudah ada. Tidak ada perubahan.")
				return nil
			}
			p := newPrompt(cmd)
			selection, err := p.ask("Pilih alias (koma; Enter=semua, '-'=batal)", strings.Join(sortedKeys(eligible), ", "), true)
			if err != nil {
				return err
			}
			if selection == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Dibatalkan.")
				return nil
			}
			selected := map[string]bool{}
			for _, alias := range strings.Split(selection, ",") {
				alias = strings.TrimSpace(alias)
				if _, exists := eligible[alias]; !exists {
					return fmt.Errorf("alias %q tidak tersedia untuk import; host lama tidak dapat ditimpa", alias)
				}
				selected[alias] = true
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nPilihan import:")
			for _, alias := range sortedKeys(eligible) {
				mark := " "
				if selected[alias] {
					mark = "✓"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", mark, alias)
			}
			confirmed, err := p.confirm("Import selected hosts?", true)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Dibatalkan.")
				return nil
			}
			for _, alias := range sortedKeys(selected) {
				if err := host.Add(hosts, eligible[alias]); err != nil {
					return err
				}
			}
			if err := store.Save(hosts); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d server diimpor.\n", len(selected))
			return nil
		},
	}
	command.Flags().StringVar(&source, "file", "", "Path SSH config (default: ~/.ssh/config)")
	return command
}
