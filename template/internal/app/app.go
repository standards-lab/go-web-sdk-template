package app

import (
	"context"
	"io"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

// App is the application layer: it owns the infrastructure registry and the
// lifecycle coordinator, assembles the router, and runs the process.
type App struct {
	cfg    *config.Config
	infra  *infrastructure.Infrastructure
	lc     *lifecycle.Coordinator
	server *web.Server
}

func New(cfg *config.Config, w io.Writer) (*App, error) {
	lc := lifecycle.New()

	infra, err := infrastructure.New(w, cfg, lc)
	if err != nil {
		return nil, err
	}

	router := web.NewRouter()
	router.Use(middleware(infra)...)
	for _, m := range routes(infra) {
		router.Mount(m)
	}

	server := web.NewServer(cfg.Server, router)
	lc.Add(lifecycle.Service{
		Name:     "server",
		Stage:    lifecycle.StageRoot,
		Start:    server.Start,
		Shutdown: server.Shutdown,
	})
	lc.Monitor(server.Err())

	checks := append(
		[]lifecycle.Check{{Name: "lifecycle", Checker: lc}},
		lc.Checks()...,
	)
	web.RegisterHealth(router, checks...)

	lc.OnReady(func() {
		infra.Logger.Info("server ready", "addr", server.Addr())
	})

	return &App{
		cfg:    cfg,
		infra:  infra,
		lc:     lc,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) int {
	if err := a.lc.Run(ctx, a.cfg.ShutdownTimeout.Duration()); err != nil {
		a.infra.Logger.Error("service failed", "error", err)
		return 1
	}
	a.infra.Logger.Info("server stopped")
	return 0
}
