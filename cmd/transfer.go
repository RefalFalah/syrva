package cmd

import (
	"fmt"
	"os"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/transfer"
	"github.com/spf13/cobra"
)

func transferFiles(cmd *cobra.Command, h host.Host, args []string, upload bool) error {
	var scpArgs []string
	var err error
	if upload {
		if len(args) != 2 {
			return fmt.Errorf("gunakan: syrva %s upload <local> <remote>", h.Alias)
		}
		local, err := transfer.LocalPath(args[0])
		if err != nil {
			return err
		}
		info, err := os.Stat(local)
		if err != nil {
			return fmt.Errorf("file lokal %q tidak dapat dibaca: %w", args[0], err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("upload hanya mendukung file biasa, bukan direktori atau device")
		}
		scpArgs, err = transfer.UploadArgs(h, args[0], args[1])
	} else {
		if len(args) != 2 {
			return fmt.Errorf("gunakan: syrva %s download <remote> <local>", h.Alias)
		}
		scpArgs, err = transfer.DownloadArgs(h, args[0], args[1])
	}
	if err != nil {
		return err
	}
	return executor(cmd).Run(cmd.Context(), "scp", scpArgs)
}
