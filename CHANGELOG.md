# Changelog

All notable changes to the `github.com/standards-lab/go-web-sdk-template/template` module are
documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). A
generated service starts its own changelog; this one records the template's.

## [Unreleased]

## [v0.7.0] - 2026-09-07

The template gains its integration tier, engine-free, over the toolkit the SDKs now ship beside
what it exercises. The module depends on `github.com/standards-lab/go-core v0.4.0` and
`github.com/standards-lab/go-web-sdk v0.7.0`.

### Added

- `integration` — the integration tier: an untagged harness that builds `cmd/server` once per
  run through go-core's `process/processtest`, runs it as a subprocess configured by `APP_*`
  variables on a reserved port with no overlay applied, and observes its liveness probe through
  go-web-sdk's `webtest` (`Main`, `Start`, `Launch`, `Service.Ready`, `Service.Client`); and,
  under the `integration` build tag, the suite asserting the baseline: boot, both probes with
  the `lifecycle` check, the drain to exit 0, and two instances side by side.
- `mise run integration` runs the tagged suite against the built service; `vet` and `lint`
  carry the tag. CI gains `workflow_dispatch` and an `integration` job on merge to main and on
  demand, in both copies of the workflow. The README states the two tiers and the isolated
  compose project a service adopts with its first backing service.

### Changed

- The module builds on go-core v0.4.0 and go-web-sdk v0.7.0 (from v0.3.0 and v0.6.0).

## [v0.6.0] - 2026-09-06

The composition root is laid out as one file per layer of the architecture, and the template
gains the administrative layer as shape and the service-owned read policy block. The module
depends on `github.com/standards-lab/go-web-sdk v0.6.0`.

### Added

- `internal/app/admin.go` — the administrative layer: the empty `Admin` composition, `newAdmin`
  taking the config root and the coordinator so the first admin service registers its
  lifecycle stage and takes its switches without a signature change, and `mountAdmin`
  building the empty `/admin` group. The mount serves on the API listener; its isolation is
  the application's decision before the first admin service is mounted, and the starter
  README says so.
- `internal/config`: the `reads` block, `ReadsConfig`, the single source of the page-size
  policy each handler constructor receives as `web.Limits` at the route build point —
  `default_size` 20 and `max_size` 100, overridable as `APP_READS_DEFAULT_SIZE` and
  `APP_READS_MAX_SIZE`, validated to the invariant `web.ParseQuery` panics on. `config.json`
  carries the defaults.

### Changed

- `internal/app` is the whole composition root: `infrastructure.go`, `admin.go`, `domain.go`,
  and `reactors.go` each construct their layer and own their mount, `routes.go` is the list of
  mounts, and `middleware.go` the router-level stack. `New` constructs infrastructure, the
  admin layer, the domain, and the reactors in that order. The package's file list is the
  architecture's layer list.
- The go-web-sdk pin moves to v0.6.0. Nothing the template calls changed in that release; a
  generated service starts with the error-returning handler adapter, `DecodeJSON`, `IfMatch`,
  and the bracket operator grammar in `ParseQuery` available.

### Removed

- The `internal/infrastructure`, `internal/domain`, and `internal/reactors` packages. Each had
  one consumer, and the reactors constructor taking the domain type forced them into one
  package; their doc comments moved onto the layer files' types and constructors.

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

[Unreleased]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.7.0...HEAD
[v0.7.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.6.0...template/v0.7.0
[v0.6.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.5.0...template/v0.6.0
[v0.5.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.4.0...template/v0.5.0
[v0.4.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.3.0...template/v0.4.0
[v0.3.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.2.0...template/v0.3.0
[v0.2.0]: https://github.com/standards-lab/go-web-sdk-template/compare/template/v0.1.0...template/v0.2.0
[v0.1.0]: https://github.com/standards-lab/go-web-sdk-template/releases/tag/template/v0.1.0
