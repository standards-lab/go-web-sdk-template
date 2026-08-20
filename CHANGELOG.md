# Changelog

All notable changes to the `github.com/standards-lab/go-web-sdk-template/template` module are
documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). A
generated service starts its own changelog; this one records the template's.

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
