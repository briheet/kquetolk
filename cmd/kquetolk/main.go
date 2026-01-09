package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/briheet/kquetolk/internal/cmd"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ret := cmd.Execute(ctx)
	os.Exit(ret)
}
