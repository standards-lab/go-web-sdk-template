package main

import (
	"context"
	"errors"
	"io"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

type server struct {
	cfg   *config.Config
	infra *infrastructure.Infrastructure
	lc    *lifecycle.Coordinator
	http  *web.Server
}

func newServer(w io.Writer, cfg *config.Config) (*server, error) {
	infra, err := infrastructure.New(w, cfg)
	if err != nil {
		return nil, err
	}

	lc := lifecycle.New()
	srv := &server{
		cfg:   cfg,
		infra: infra,
		lc:    lc,
		http:  web.NewServer(cfg.Server, newRouter(infra, lc)),
	}
	srv.bind()
	return srv, nil
}

func (s *server) bind() {
	s.lc.OnStartup(s.infra.Start)
	s.lc.OnStartup(s.http.Start)

	// One ordered hook: errors.Join evaluates left to right, so the server
	// drains fully before the infrastructure beneath it shuts down.
	s.lc.OnShutdown(func(ctx context.Context) error {
		return errors.Join(
			s.http.Shutdown(ctx),
			s.infra.Shutdown(ctx),
		)
	})

	s.lc.Monitor(s.http.Err())

	s.lc.OnReady(func() {
		s.infra.Logger.Info("server ready", "addr", s.http.Addr())
	})
}

func (s *server) serve(ctx context.Context) int {
	if err := s.lc.Run(ctx, s.cfg.ShutdownTimeout.Duration()); err != nil {
		s.infra.Logger.Error("service failed", "error", err)
		return 1
	}
	s.infra.Logger.Info("server stopped")
	return 0
}
