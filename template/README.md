# Service

Generated from [go-web-sdk-template](https://github.com/standards-lab/go-web-sdk-template).

## After generation

Three steps localize the service's identity:

1. Rename `envPrefix` in `internal/config/config.go` — the single constant every `APP_*`
   environment-variable name derives from.
2. Rename the `APP_ENV` key in `mise.toml`'s `[env]` block to follow the prefix.
3. Rewrite this README for the service.

Licensing and release automation are yours to define; CI arrives working (`.github/workflows/ci.yml`).

## Getting started

[mise](https://mise.jdx.dev/) provisions the toolchain and runs the tasks:

```sh
mise trust && mise install
mise run serve
```

The service logs `server ready` on `localhost:8080` (the `local` overlay binds loopback and runs
debug logging). From a second shell:

```sh
curl localhost:8080/healthz   # 200 {"status":"ok"}
curl localhost:8080/readyz    # 200 {"status":"ready","checks":[...]}
```

Ctrl-C drains in-flight requests and exits with `server stopped`.

## Tasks

Each task wraps a plain command, so the repository works without mise:

| Task | Command | What it does |
|------|---------|--------------|
| `mise run check` | build, vet, `gofmt -l`, `go fix -diff`, `go mod tidy -diff`, `go test -race`, `golangci-lint run` | The one read-only check, as CI runs it |
| `mise run currency` | `bash scripts/currency.sh` | Report requirements, Go version, tools, and action pins behind their latest |
| `mise run upgrade` | `go mod edit -go=<current minor> -toolchain=none`, `go get <direct>@latest`, `go mod tidy`, `mise upgrade --bump --local` | Upgrade the go directive, requirements, and tools, rewriting `go.mod` and `mise.toml` |
| `mise run build` | `go build ./...` | Build the module |
| `mise run vet` | `go vet -tags integration ./...` | Compile-check and vet, the integration suite included |
| `mise run serve` | `go run ./cmd/server` | Run the service locally |
| `mise run test` | `go test -race ./...` | Run the unit tier |
| `mise run integration` | `go test -race -count=1 -tags integration ./integration/` | Run the integration tier against the built service |
| `mise run fmt` | `gofmt -w .` | Format the source |
| `mise run tidy` | `go mod tidy` | Reconcile module requirements |
| `mise run lint` | `golangci-lint run --build-tags integration ./...` | Lint, the integration suite included |

## Testing

Two tiers. The unit tier is hermetic and runs on every pull request: `go test -race ./...`,
touching no service, network, or disk. The integration tier is the `integration` package: an
untagged harness that builds `cmd/server` once, runs it as a subprocess configured by `APP_*`
variables on a port it reserved, and drives it through its HTTP surface, over the toolkit the
SDKs ship for the purpose (go-core's `process/processtest`, go-web-sdk's `webtest`). Under the
`integration` build tag, the suite asserts the service's behavior through its API. It runs on
merge to main and on demand, below the per-PR rate by design.

The template's suite asserts the baseline: boot, both probes, and the drain. As the build points
fill in, each capability adds its cases beside them, and the first backing service brings its
compose stack: the task then boots that stack as an isolated compose project on its own port,
runs the suite, and tears it down with its volume on exit, so a developer's stack and data are
never touched. The reference service shows the shape.

## Configuration

Configuration layers in a fixed precedence, later sources winning:

1. `config.json` — the base file, every knob at its default.
2. `config.<APP_ENV>.json` — the environment overlay; `mise.toml` sets `APP_ENV=local`, which
   activates the committed `config.local.json` (loopback host, debug logging).
3. `secrets.json`, `secrets.<APP_ENV>.json` — gitignored secret layers.
4. `APP_*` environment variables — the final override:
   - `APP_LOG_LEVEL`, `APP_LOG_FORMAT`
   - `APP_SERVER_HOST`, `APP_SERVER_PORT`, the four server timeout variables, and
     `APP_SERVER_TRANSFER_RATE`
   - `APP_READS_DEFAULT_SIZE`, `APP_READS_MAX_SIZE`
   - `APP_SHUTDOWN_TIMEOUT`

Every file is optional — a deployment can run on the base file and environment variables alone,
or on environment variables only. Each file present decodes strictly: a key the `Config` types do
not declare fails the load.

## Building out the service

The composition root, `internal/app`, is laid out as one file per layer of the architecture,
and those files are the build points:

- `infrastructure.go` for the services the application composes on
- `admin.go` for the administrative services
- `domain.go` for the domain services
- `reactors.go` for the process-lifetime entry points an occurrence drives
- `middleware.go` for the router-level middleware

Each layer file defines its layer's nodes on go-core's dependency graph, in its define function
(`defineInfrastructure`, `defineAdmin`, `defineDomain`, `defineReactors`), and owns its mount;
`server.go` defines the request edge (the readiness the probes report, the router, the server),
and `routes.go` lists the mounts and does nothing else. The `Nodes` value holds one handle per
node: each define function fills its part, and a constructor reads the lower layers' nodes from
it with `Use`. `New` describes the graph and cannot fail; `Run` builds it and runs it under
go-core's lifecycle Coordinator. `cmd/server` is the entrypoint alone and never changes.

A domain service starts from its Entity:

1. Give the Entity its own package under `domain/`.
2. Expose its Queries and Commands as the domain service's methods.
3. Build the layer's route group in its handler.
4. Define the service's node in `defineDomain` and mount the group in `mountAPI` (`domain.go`).

The constructor draws what it uses from the infrastructure nodes it `Use`s, and the handler is
handed its policy from the config node at the construction site (`cfg.Reads.Limits()` for a
collection read) and the service's logger for its error writer (`web.NewErrorWriter(logger, ...)`).

An infrastructure service (a database pool, a storage client, an auth client) is a field on
`Nodes` plus its node, defined in `defineInfrastructure` (`internal/app/infrastructure.go`) —

```go
n.DB = g.Define("database", func(s *graph.Scope) (*db.Pool, error) {
	return db.New(s.Use(n.Config).DB, s.Use(n.Logger)), nil
})
```

Its part in the lifecycle is inferred from its value's methods, with no registration: a
`lifecycle.Starter` is started and a `lifecycle.Stopper` shut down (a `Subsystem` is both), a
`ReadinessChecker` joins `/readyz` under the node's name, and a `Monitored` has its runtime
error watched. Its constructor opens nothing; connectivity belongs to its `Start`, so a failed
build leaks no connections. Nodes start in layer order, each in a layer above the nodes it uses,
and drain in reverse; the server is alone in the top layer, so it starts last and drains first, and
in-flight requests complete before their infrastructure closes. A service defined this way
cannot be missing from the probe or the drain, and a `Nodes` field that does not exist fails
compilation at its access.

An admin service administers one infrastructure service over the mechanisms its library
provides. It is a node defined in `defineAdmin` (`internal/app/admin.go`), with its route group
mounted under `/admin` in `mountAdmin`. One that verifies and corrects the state of the
infrastructure it administers is a lifecycle participant by its methods, and the domains that
depend on that state `Use` it, so it starts ahead of them. The template serves the empty
`/admin` group on the API listener; before the first admin service is mounted, settle the
mount's isolation — its own listener, authentication, and audit — because an administrative
surface on the public port is exposed the moment it serves a route.

A reactor is an entry point that runs for the process lifetime, driven by an occurrence (a
subscription, an interval, a wake on demand) rather than a caller. Each is a node defined in
`defineReactors` (`internal/app/reactors.go`), its lifecycle inferred the same as an
infrastructure service's. No node uses a reactor, so it is a Build root: append it to
`Nodes.Reactors`, which `Run` builds and the server orders itself after. A reactor often
dispatches each occurrence to a domain service call, but need not; a background worker the
service runs is a reactor too.

Middleware that applies to every route stacks in `middleware` (`internal/app/middleware.go`),
outermost first. The template ships `RequestID`, `RequestLogger`, and `Recoverer`, in the chain
order that go-web-sdk's `middleware` package documents. Middleware scoped to one domain service
belongs on its route group.

Configuration grows by adding fields to `Config` in `internal/config/config.go` and delegating
to their `Merge` and `Finalize` in the existing shape; the `reads` block is the model for a
service-owned block. The service keeps pace with its SDKs by updating its `go-core` and
`go-web-sdk` pins and applying whatever adjustments the release notes call for.
