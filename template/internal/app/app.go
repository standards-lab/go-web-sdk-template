package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

// Wiring carries the manifests the composition root declares: the
// router-level middleware stack, outermost first, and the domain-service
// modules to mount.
type Wiring struct {
	Middleware []web.Middleware
	Modules    []*web.Module
}

// App is the application layer: it owns the infrastructure registry and the
// lifecycle coordinator, assembles the router, and runs the process.
type App struct {
	cfg    *config.Config
	infra  *infrastructure.Registry
	logger *slog.Logger
	lc     *lifecycle.Coordinator
	server *web.Server
}

// New is the cold start: router assembly from the wiring, server
// construction, and coordinator binding, with no I/O. The probes register on
// the router's native mux, outside every module's middleware, from the
// coordinator's check plus the registry's, and the logger is pulled from the
// registry once, here. Wiring mistakes panic at construction.
func New(cfg *config.Config, infra *infrastructure.Registry, wiring Wiring) *App {
	logger := infra.Get[*slog.Logger]()
	lc := lifecycle.New()

	router := web.NewRouter()
	router.Use(wiring.Middleware...)
	for _, m := range wiring.Modules {
		router.Mount(m)
	}

	checks := append(
		[]web.Check{{Name: "lifecycle", Checker: lc}},
		infra.Checks()...,
	)
	web.RegisterHealth(router, checks...)

	a := &App{
		cfg:    cfg,
		infra:  infra,
		logger: logger,
		lc:     lc,
		server: web.NewServer(cfg.Server, router),
	}
	a.bind()
	return a
}

func (a *App) bind() {
	a.lc.OnStartup(a.infra.Start)
	a.lc.OnStartup(a.server.Start)

	// One ordered hook: errors.Join evaluates left to right, so the server
	// drains fully before the infrastructure beneath it shuts down.
	a.lc.OnShutdown(func(ctx context.Context) error {
		return errors.Join(
			a.server.Shutdown(ctx),
			a.infra.Shutdown(ctx),
		)
	})

	a.lc.Monitor(a.server.Err())

	a.lc.OnReady(func() {
		a.logger.Info("server ready", "addr", a.server.Addr())
	})
}

// Run is the hot start plus shutdown, delegated to the coordinator, and
// returns the process exit code.
func (a *App) Run(ctx context.Context) int {
	if err := a.lc.Run(ctx, a.cfg.ShutdownTimeout.Duration()); err != nil {
		a.logger.Error("service failed", "error", err)
		return 1
	}
	a.logger.Info("server stopped")
	return 0
}
