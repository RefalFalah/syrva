package cmd

import (
	"fmt"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
	"github.com/spf13/cobra"
)

func executor(cmd *cobra.Command) ssh.Executor {
	return ssh.Executor{Stdin: cmd.InOrStdin(), Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr()}
}

func (a *app) hostCommand(cmd *cobra.Command, h host.Host, args []string) error {
	if len(args) == 0 {
		return connect(cmd, h)
	}
	switch args[0] {
	case "info":
		if len(args) != 1 {
			return fmt.Errorf("gunakan: syrva %s info", h.Alias)
		}
		return printInfo(cmd.OutOrStdout(), h)
	case "run":
		return runPreset(cmd, h, args[1:])
	case "upload", "download":
		return transferFiles(cmd, h, args[1:], args[0] == "upload")
	case "tunnel":
		return openTunnel(cmd, h, args[1:])
	}
	return fmt.Errorf("command host %q tidak dikenal. Gunakan syrva %s [info|run|upload|download|tunnel]", args[0], h.Alias)
}

func connect(cmd *cobra.Command, h host.Host) error {
	args, err := ssh.ConnectArgs(h)
	if err != nil {
		return err
	}
	return executor(cmd).Run(cmd.Context(), "ssh", args)
}
