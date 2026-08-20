# Implementation guide · recompose the root

The imported composition root buries the router assembly inside `newServer` and asks a growing
service to edit four places per infrastructure service. This change recomposes it on the
Elemental Architecture layout: `cmd/server` reduces to the entrypoint plus three declarative
manifests (infrastructure services, middleware, routes), and the assembly machinery moves below
cmd into `internal/app`. `App` owns the infrastructure registry and drives its lifecycle;
extending a seeded service means adding an entry to a manifest, never touching the machinery.

The registry is type-keyed: `Register` adds one instance per type together with its lifecycle
declaration, and `Get` retrieves it by type. Both panic on a wiring mistake (a duplicate type,
an absent type), and every call site lives in the composition root and runs during the cold
start, so a mistake cannot reach a running process. The discipline that keeps this from
becoming a service locator: only the manifests and the application layer touch the registry;
everything below receives its dependencies through constructor parameters.

Godoc arrives with the agent's closeout pass; the code here is bare. Apply the mutations in
order. The carried tests reference the old `newServer` and fail to compile until the agent
restores the suite after this guide, so verify with `go build` and a manual serve, per the
closing section.

## 1. `template/internal/infrastructure/infrastructure.go` — replace the file

The package becomes pure machinery: the `Service` lifecycle declaration and the type-keyed
`Registry`. Construction moves to the manifest in `cmd/server`, so `logging` and
`internal/config` leave the imports. Registration order is startup order; `Shutdown` walks the
registrations in reverse with `slices.Backward`, and every service gets its shutdown attempt
before the errors join. `Register` and `Get` are parameterized methods, new in Go 1.27; the
module and toolchain move to 1.27 in mutation 8.

```go
package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
)

type Service struct {
	Name     string
	Start    func(ctx context.Context) error
	Shutdown func(ctx context.Context) error
	Check    lifecycle.ReadinessChecker
}

type Registry struct {
	values   map[reflect.Type]any
	services []Service
}

func NewRegistry() *Registry {
	return &Registry{values: make(map[reflect.Type]any)}
}

func (r *Registry) Register[T any](value T, svc Service) {
	key := reflect.TypeFor[T]()
	if _, exists := r.values[key]; exists {
		panic(fmt.Sprintf("infrastructure: %s already registered", key))
	}
	r.values[key] = value
	r.services = append(r.services, svc)
}

func (r *Registry) Get[T any]() T {
	value, ok := r.values[reflect.TypeFor[T]()]
	if !ok {
		panic(fmt.Sprintf("infrastructure: %s not registered", reflect.TypeFor[T]()))
	}
	return value.(T)
}

func (r *Registry) Start(ctx context.Context) error {
	for _, svc := range r.services {
		if svc.Start == nil {
			continue
		}
		if err := svc.Start(ctx); err != nil {
			return fmt.Errorf("%s: %w", svc.Name, err)
		}
	}
	return nil
}

func (r *Registry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, svc := range slices.Backward(r.services) {
		if svc.Shutdown == nil {
			continue
		}
		if err := svc.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", svc.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (r *Registry) Checks() []web.Check {
	var checks []web.Check
	for _, svc := range r.services {
		if svc.Check == nil {
			continue
		}
		checks = append(checks, web.Check{Name: svc.Name, Checker: svc.Check})
	}
	return checks
}
```

## 2. `template/internal/app/app.go` — new file, new directory

The application layer: `New` is the cold start (router assembly, server construction,
coordinator binding, zero I/O), `Run` is the hot start plus shutdown. The probes register on
the router's native mux, outside every module's middleware, from the coordinator's check plus
the registry's. The composite shutdown hook keeps the imported ordering: `errors.Join`
evaluates left to right, so the HTTP server drains before the infrastructure beneath it
closes. The logger is pulled from the registry once, at construction.

```go
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

type Wiring struct {
	Middleware []web.Middleware
	Modules    []*web.Module
}

type App struct {
	cfg    *config.Config
	infra  *infrastructure.Registry
	logger *slog.Logger
	lc     *lifecycle.Coordinator
	http   *web.Server
}

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
		http:   web.NewServer(cfg.Server, router),
	}
	a.bind()
	return a
}

func (a *App) bind() {
	a.lc.OnStartup(a.infra.Start)
	a.lc.OnStartup(a.http.Start)

	a.lc.OnShutdown(func(ctx context.Context) error {
		return errors.Join(
			a.http.Shutdown(ctx),
			a.infra.Shutdown(ctx),
		)
	})

	a.lc.Monitor(a.http.Err())

	a.lc.OnReady(func() {
		a.logger.Info("server ready", "addr", a.http.Addr())
	})
}

func (a *App) Run(ctx context.Context) int {
	if err := a.lc.Run(ctx, a.cfg.ShutdownTimeout.Duration()); err != nil {
		a.logger.Error("service failed", "error", err)
		return 1
	}
	a.logger.Info("server stopped")
	return 0
}
```

## 3. `template/cmd/server/infrastructure.go` — new file

The infrastructure-service manifest: every service the application composes on is constructed
and registered here, in dependency order, from the root configuration's blocks. Construction
performs no I/O; connectivity waits for `Registry.Start`. The logger is inert, so its entry
declares only a name.

```go
package main

import (
	"io"

	"github.com/standards-lab/go-core/logging"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

func newInfrastructure(w io.Writer, cfg *config.Config) (*infrastructure.Registry, error) {
	r := infrastructure.NewRegistry()

	r.Register(logging.New(w, cfg.Log), infrastructure.Service{
		Name: "logger",
	})

	return r, nil
}
```

A lifecycle-bearing service lands as one more entry in the same shape:

```go
r.Register(pool, infrastructure.Service{
	Name:     "postgres",
	Start:    pool.Start,
	Shutdown: pool.Close,
	Check:    pool,
})
```

## 4. `template/cmd/server/middleware.go` — new file

The middleware manifest: the router-level stack, outermost first. Route- and group-level
middleware belongs to the domain-service modules in the routes manifest. The function is
`middlewares` because the SDK's `middleware` package takes the singular name.

```go
package main

import (
	"log/slog"

	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk/middleware"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

func middlewares(infra *infrastructure.Registry) []web.Middleware {
	return []web.Middleware{
		middleware.RequestLogger(infra.Get[*slog.Logger]()),
	}
}
```

## 5. `template/cmd/server/routes.go` — replace the file

The route manifest replaces `newRouter`; router assembly and the probes now live in `app.New`.
Every domain service's module is constructed and mounted here, drawing its dependencies from
the registry. The baseline carries no domain services, so the manifest is empty; a domain
service lands as its module constructor and one entry, in the shape
`orders.NewModule(infra.Get[*pgxpool.Pool]())`.

```go
package main

import (
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

func routes(infra *infrastructure.Registry) []*web.Module {
	return nil
}
```

## 6. `template/cmd/server/main.go` — replace `run`

`main` is unchanged. `run` now reads as the phase sequence: signals, configuration,
infrastructure manifest, application assembly from the manifests, run.

```go
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

	infra, err := newInfrastructure(stdout, cfg)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "infrastructure init failed:", err)
		return 1
	}

	a := app.New(cfg, infra, app.Wiring{
		Middleware: middlewares(infra),
		Modules:    routes(infra),
	})

	return a.Run(ctx)
}
```

The import block gains the app package:

```go
import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)
```

## 7. Delete `template/cmd/server/server.go`

The `server` type, `newServer`, `bind`, and `serve` are subsumed by `internal/app`.

## 8. Toolchain to Go 1.27

Parameterized methods require it. Four files, one commit, per the dual-copy rule:

- `template/go.mod`: `go 1.26` becomes `go 1.27`.
- `mise.toml` and `template/mise.toml`: `go = "1.26"` becomes `go = "1.27"`.
- `.github/workflows/ci.yml` and `template/.github/workflows/ci.yml`:
  `go-version: "1.26"` becomes `go-version: "1.27"`.

go-core and go-web-sdk keep their `go 1.26` directives; a directive is a minimum, so the
resolved build is 1.27 throughout.

## Run and verify

From `template/` (the test files still reference `newServer` and are restored by the agent
next, so build and serve are the verification):

```
go build ./...
gofmt -l .
go run ./cmd/server
```

Expect `server ready` with `addr=127.0.0.1:8080` (the local overlay). From a second shell:

```
curl -i localhost:8080/healthz   # 200, {"status":"ok"}
curl -i localhost:8080/readyz    # 200, {"status":"ready","checks":[{"name":"lifecycle",...}]}
```

Ctrl-C the server: it drains and logs `server stopped`. Hand back for tests, docs, and
scaffolding.
