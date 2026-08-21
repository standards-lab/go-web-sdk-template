# go-web-sdk-template

The web service template of the `go-minimal` standard. A generated service starts as a minimal
runnable web service on [go-core](https://github.com/standards-lab/go-core) and
[go-web-sdk](https://github.com/standards-lab/go-web-sdk), carrying:

- layered configuration: a base file, environment overlays, secret layers, and environment
  variables, in fixed precedence;
- the cold/hot start lifecycle: constructed with no I/O, run under go-core's coordinator, and
  drained in order on shutdown;
- liveness and readiness probes on `/healthz` and `/readyz`, fed by the lifecycle coordinator
  and the services' named checks;
- three build points in `internal`: the infrastructure constructor, route registration, and
  the middleware stack.

New services are generated from it:

```sh
# install the gonew command if needed
go install golang.org/x/tools/cmd/gonew@latest

# generate a new web service
gonew github.com/standards-lab/go-web-sdk-template/template@latest example.com/newsvc
```

[`gonew`](https://pkg.go.dev/golang.org/x/tools/cmd/gonew) copies the module and rewrites its
path; the copy is a running service from the first build. The module lives under `template/` —
that subtree is exactly what a generated service receives, and everything above it (the
template's own planning context, changelog, and tooling) stays behind by construction. Versions
are the `template/v*` tags; `@latest` follows the newest.

[`template/README.md`](template/README.md) is the starter README every generated service
receives. It documents running, configuring, and building out the service.

## Standard

`go-web-sdk-template` is the web service template of
[Go Minimal](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/index.md), the
minimal-dependency Go standard, and its design is documented on the standard's
[go-web-sdk-template page](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/go-web-sdk-template/index.md).
Its repository-level principles:

- The template module depends on `go-core` and `go-web-sdk` at pinned releases, and nothing
  else. The template is engine-free: no data engine is declared and no provider is imported; a
  generated service chooses its providers in its own composition root.
- Everything in the module survives the `gonew` path rewrite. The module path is the only
  identity, and the generated copy runs from the first build.
- The template stays minimal and stable: it scaffolds the baseline architecture, and the SDKs
  carry the bulk of development change. A generated service keeps pace by updating its
  `go-core` and `go-web-sdk` versions. Service integrations stay out; the infrastructure
  libraries define them, and the organization's reference architecture documents how they are
  integrated.

## Repository layout

`template/` is the module a generated service copies, and development happens inside it. The
root above it maintains the template: `context/` records its design and session state, this
README documents it, `CHANGELOG.md` sources the `template/v*` release notes, and the CI
workflow runs the checks inside `template/`. The template's `.gitignore` and CI workflow are
separate copies of the root ones; a change to either copy is applied to the other in the same
commit.

## Development

```sh
cd template
mise trust && mise install
mise run test
```

`template/mise.toml` carries the tasks (`vet`, `serve`, `test`, `fmt`, `tidy`, `lint`), each
wrapping a plain command.

## License

[Apache License 2.0](LICENSE).
