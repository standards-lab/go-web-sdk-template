# The baseline

The design of the template baseline: the architecture the template scaffolds, and the
principles that keep it minimal. This is the durable design reference for the project.

## Architecture

A single Go module, `github.com/standards-lab/go-web-sdk-template/template`, built on `go-core`
and `go-web-sdk` at pinned releases. The composition root is `cmd/server`: the entrypoint plus
the three manifests that declare the service's composition, one concern per file. The machinery
the manifests feed sits below cmd, in `internal/`.

- `main.go` is the process entrypoint. It owns the signal-derived root context and the exit
  code, and runs the phase sequence through a testable `run(stdout, stderr io.Writer) int`:
  configuration load, the infrastructure manifest, application assembly, run.
- `cmd/server/infrastructure.go` is the infrastructure-service manifest. `setInfrastructure`
  constructs every service the application composes on and registers each once, in dependency
  order; a registration carries the handle and the lifecycle declaration in the same call.
- `cmd/server/routes.go` is the route manifest. `setRoutes` mounts the domain-service modules,
  each constructor drawing its dependencies from the registry, so a module's signature declares
  what its domain service uses. The baseline mounts none; the manifest is the build point.
- `cmd/server/middleware.go` is the middleware manifest. `setMiddleware` stacks the
  router-level middleware outermost first; middleware scoped to one domain service belongs on
  its module.
- `internal/app` is the application layer. `New` is the cold start — router assembly from the
  wiring, the probes on the native mux from the coordinator's check plus the registry's, server
  construction, and lifecycle binding, with no I/O — and `Run` is the hot start plus shutdown,
  delegated to go-core's coordinator. One composite hook drains the HTTP server before the
  infrastructure beneath it closes.
- `internal/infrastructure` is the type-keyed registry. `Register` stores one instance per type
  with its lifecycle declaration; `Get` retrieves by type, and both panic on a wiring mistake
  at cold start. Startup runs in registration order, shutdown in reverse joining every error,
  and the registered checks feed the readiness probe. Retrieval stops at the composition layer:
  the manifests and the application layer read the registry, and domain packages receive their
  dependencies as constructor parameters. Roles sharing a type (a write pool and a read pool)
  register as defined wrapper types; a dynamic set of like services registers as one service
  that owns its members.
- `internal/config` declares the root configuration. `Config` composes the library blocks with
  the service's shutdown timeout, and `Load` reads the layered files and finalizes them under
  the one `envPrefix` constant.

`/readyz` reads the lifecycle coordinator and every registered check: not ready until startup
completes, and not ready again once draining begins.

## Principles

- **Minimal and stable.** The template scaffolds the initial baseline architecture, not a
  framework to track. The SDKs carry the bulk of development change; a generated service keeps
  pace by updating its `go-core` and `go-web-sdk` versions and applying whatever adjustments
  the release notes call for. Capability integrations stay out: they live in the capability
  repositories, and how one is integrated is documented where it is integrated. The template
  owns the composition pattern — the manifests, the application layer, and the registry a new
  capability lands in.
- **Engine-free.** The baseline declares no data engine and depends on no provider. A generated
  service chooses its providers in its own composition root, and that root, never the template,
  is where a provider is imported.
- **Generatable.** `gonew github.com/standards-lab/go-web-sdk-template/template@latest` copies
  the module and rewrites its path. Everything in the subtree must survive that rewrite: the
  module path is the only identity, and the copy is a running service from the first build.
- **Distributed as a module.** The template is a real, versioned module resolved through the
  public Go proxy, so generation works anywhere Go does — `gonew` fills the role project
  scaffolding tools such as `dotnet new` fill in other ecosystems, with no dependency on a
  platform's template feature.
- **Toolchain.** Go 1.27: the registry's `Register` and `Get` are parameterized methods, which
  1.27 introduced.
