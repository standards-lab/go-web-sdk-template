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
| `mise run check` | build, vet, `gofmt -l`, `go fix -diff`, `go mod tidy -diff`, `go test -race`, `golangci-lint run`, `lint-context`, `lint-references` | The one read-only check, as CI runs it |
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

`internal/app`'s graph tests check the composition itself. Each `Nodes` field names its node
(go-core's `graphtest.CheckNames`). A Build of the root set, `App.Roots` plus the lifecycle
configuration node, builds exactly what `Run` builds; in it the server is alone in the top layer,
a reactor sits below the server by its membership of the reactors group alone, and no order
between nodes is dropped but one: the server's order after the reactors group, while the group
is empty. Each check is shown failing on a fixture that breaks it, with the layer table.

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

[Building out the service](docs/building-out.md) covers how the layer files define and mount
their nodes, and walks through adding a domain service, an infrastructure service, an admin
service, a reactor, middleware, and configuration.
