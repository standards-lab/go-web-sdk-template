# Changelog

All notable changes to the `github.com/standards-lab/go-web-sdk-template/template` module are
documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). A
generated service starts its own changelog; this one records the template's.

## [v0.5.0] - 2026-08-28

The template composes on go-web-sdk v0.5.0 and models the API-module convention the reference
service proved: one `/api` group, shipped initialized and empty, with the config root at the
route build point so each handler is handed its policy at the construction site.

### Changed

- `internal/app`: `routes` widens to `routes(dom *domain.Domain, cfg *config.Config)` and its
  body ships the initialized empty `/api` module — an application mounts its domain-service
  route groups into the one group, each handler handed its policy from `cfg` where it is
  constructed. The empty group serves no route; the baseline's behavior is unchanged.
- `internal/domain`: the package doc states the layering — domain services are defined in the
  module's base-layer domain packages and constructed here; the composition root registers and
  composes only.
- The go-web-sdk pin moves to v0.5.0. Nothing the template imports changed in that release;
  the pin is the committed steady state for generated services, which start with `ParseQuery`
  and `ErrorWriter` available.

## [v0.4.0] - 2026-08-24

The composition root now builds on go-core's `process` package, and the suite gains the
configtest convention. A generated service starts with the pre-infrastructure main sequence
imported rather than inlined, and with one package that knows the configuration's required
fields.

### Added

- `internal/config/configtest` — hermetically valid configuration for the suites: `Config`
  returns a finalized root config whose composition performs no I/O. When a subsystem's block
  gains a required field, it is set here once and every consuming test adapts. The app test
  builds its config through it.

### Changed

- `cmd/server` composes its run function on go-core's `process` package: `SignalContext`
  supplies the signal-derived root context and `Fail` the pre-logger failure reporting,
  replacing the inline signal wiring and bare exit literals.
- The end-to-end app test is marked as bound to the baseline's inert infrastructure, naming
  the startup-contract tests that replace it when the first subsystem with a lifecycle
  arrives.
- Pins move to go-core v0.3.0 and go-web-sdk v0.3.1.

## [v0.3.0] - 2026-08-21

Two new composition-root layers, `internal/domain` and `internal/reactors`, sit between
`internal/infrastructure` and `internal/app`; both ship empty, matching the empty `routes()`
build point already in the baseline. The module depends on
`github.com/standards-lab/go-web-sdk v0.3.0`.

### Added

- **`internal/domain`** composes the application's domain services over infrastructure. `New`
  takes no lifecycle coordinator: domain services own no resource and never run, so there's
  nothing to register.
- **`internal/reactors`** composes the application's event-driven entry points — components
  that watch a source of occurrences and dispatch each one to a domain service call, the
  inbound counterpart to a route. `New` takes the coordinator and registers each one on it.

### Changed

- **`internal/app`**: construction order is infrastructure, then domain, then reactors, then
  the router. `routes()` takes `*domain.Domain` in place of `*infrastructure.Infrastructure`;
  `middleware()` keeps taking `*infrastructure.Infrastructure`, since the template's one
  middleware, request logging, is a cross-cutting infrastructure concern, not domain logic.
  `App` drops its `infra` field — nothing on `App` can reach a pool once it's gone — and gains
  `logger`, populated from `infra.Logger` at construction. The `RegisterHealth` call passes the
  coordinator directly, matching go-web-sdk v0.3.0's live-query signature.

## [v0.2.0] - 2026-08-21

The composition root moves into the application layer and the type-keyed registry is deleted:
infrastructure is a struct of concrete fields, lifecycle declarations go directly to go-core's
staged coordinator, and `cmd/server` shrinks to the entrypoint. The module depends on
`github.com/standards-lab/go-core v0.2.0` and `github.com/standards-lab/go-web-sdk v0.2.0`.

### Changed

- **`internal/app`** is the composition root: the `App` structure orchestrates the root
  composition from the package's build points, route registration (`routes.go`) and the
  middleware stack (`middleware.go`). `New(cfg, w)` creates the coordinator, constructs the
  infrastructure, assembles the router from the registered routes and the middleware stack,
  and declares the server as the coordinator's root-stage service — started after every
  infrastructure stage and drained first, replacing the composite shutdown hook — before
  snapshotting the readiness checks.
- **`internal/infrastructure`** constructs the application's services into the concrete
  fields of `Infrastructure`. `New(w, cfg, lc)` registers each service that has a lifecycle
  on the coordinator where it is constructed. A wiring mistake is a compile error at the
  field access instead of a cold-start panic.
- **`cmd/server`** is the entrypoint alone: the signal context, the config load, `app.New`,
  and the exit code. The infrastructure constructor, route registration, and the middleware
  stack moved into `internal`, so a growing service never edits the binary package — and the
  package-main test exception leaves the template with them.
- The server starts only after the infrastructure stages complete; previously the two startup
  hooks ran concurrently.

### Removed

- The type-keyed `infrastructure.Registry`: the `reflect` map, the parameterized `Register`
  and `Get`, its `Service` type, and its ordered start and reverse shutdown, which now belong
  to `lifecycle.Coordinator.Add`. Roles sharing a type are distinct fields; the wrapper-type
  rule existed for the map key and goes with it.

## [v0.1.0] - 2026-08-20

The first release of the web service template: the `template/` subtree module, carried from
go-service-template v0.2.0 and recomposed on the Elemental Architecture layout. The module
depends on `github.com/standards-lab/go-core v0.1.0` and
`github.com/standards-lab/go-web-sdk v0.1.0`, on Go 1.27.

### Added

- **`cmd/server`** — the entrypoint and the three manifests that declare the service's
  composition, one concern per file. `main.go` runs the phase sequence through a testable
  `run(stdout, stderr) int`: signal context, configuration load, infrastructure manifest,
  application assembly, run. `setInfrastructure` constructs and registers every infrastructure
  service in one place, `setMiddleware` declares the router-level stack outermost first
  (`middleware.RequestLogger` in the baseline), and `setRoutes` mounts the domain-service
  modules (none in the baseline; the manifest is the build point).
- **`internal/app`** — the application layer. `New` is the cold start: router assembly from
  the wiring, the `/healthz` and `/readyz` probes on the native mux from the coordinator's
  check plus the registry's, server construction, and lifecycle binding, with no I/O. `Run` is
  the hot start plus shutdown, delegated to go-core's coordinator, with one composite hook
  draining the HTTP server before the infrastructure beneath it closes.
- **`internal/infrastructure`** — the type-keyed registry of infrastructure services.
  `Register` carries a service's handle and lifecycle declaration in one call and `Get`
  retrieves by type (parameterized methods, Go 1.27); one instance per type, with roles
  sharing a type distinguished by defined wrapper types. Startup runs in registration order,
  shutdown in reverse joining every error, and the registered checks feed the readiness probe.
- **`internal/config`** — the layered configuration root: `Config` composes the log and server
  blocks with the service's shutdown timeout under the one `envPrefix` constant, and `Load`
  reads base, environment overlay, and secret files before finalizing.
- **Configuration files** — `config.json` with every knob at its default and the committed
  `config.local.json` overlay (loopback host, debug logging).
- **Developer tooling** — mise tasks wrapping plain Go commands, CI running vet, race tests,
  and lint inside `template/`, and releases cut from this changelog on `template/v*` tags.
- **Generation** — `gonew github.com/standards-lab/go-web-sdk-template/template@latest`
  copies the subtree as a running service; the starter README carries the after-generation
  identity steps.
