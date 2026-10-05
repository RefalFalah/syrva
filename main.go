package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RefalFalah/syrva/cmd"
	"github.com/RefalFalah/syrva/internal/ssh"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.NewRoot().ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return 130
		}
		fmt.Fprintln(os.Stderr, err)
		var exit *ssh.ExitError
		if errors.As(err, &exit) && exit.Code > 0 {
			return exit.Code
		}
		return 1
	}
	return 0
}
