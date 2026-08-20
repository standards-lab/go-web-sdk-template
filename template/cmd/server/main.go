package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "config load failed:", err)
		return 1
	}

	srv, err := newServer(stdout, cfg)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "server init failed:", err)
		return 1
	}

	return srv.serve(ctx)
}
