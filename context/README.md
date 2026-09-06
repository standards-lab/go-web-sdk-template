# go-web-sdk-template

The web service template of Go Elemental, the Standards Lab organization's Go implementation of
the Elemental Architecture: scaffolds an initial Go Elemental web service application with `gonew`, built on
go-core and go-web-sdk at pinned releases, and the first code expression of the Elemental
Architecture's application layout.

The design and conventions of this repository are documented in the organization's
[documentation landing zone](https://github.com/standards-lab/docs); this context records only
working knowledge the landing zone and the code do not express. The repository page is
[go-web-sdk-template](https://github.com/standards-lab/docs/blob/main/standards/go-elemental/go-web-sdk-template/index.md),
under the [Go Elemental](https://github.com/standards-lab/docs/blob/main/standards/go-elemental/index.md)
standard, with the design detailed in
[The baseline](https://github.com/standards-lab/docs/blob/main/standards/go-elemental/go-web-sdk-template/baseline.md),
[The elements in the template](https://github.com/standards-lab/docs/blob/main/standards/go-elemental/go-web-sdk-template/elements.md),
and
[The template subtree](https://github.com/standards-lab/docs/blob/main/standards/go-elemental/go-web-sdk-template/template-subtree.md).

## Capability map

The code and each package's `doc.go` are authoritative for what is built; the landing zone
documents the design.

- **Runnable baseline** — built (template/v0.6.0): the composition root `internal/app` on the
  Elemental Architecture layout, one file per layer — `infrastructure.go`, `admin.go`,
  `domain.go`, and `reactors.go`, each constructing its layer in that order onto go-core's
  staged coordinator and owning its mount, with `routes.go` the list of mounts and
  `middleware.go` the router-level stack — plus the `internal/config` root and the `cmd/server`
  entrypoint, with the probes, the staged drain, and the test suite. The `/api` and `/admin`
  groups ship initialized and empty, with the config root in scope so handlers take their
  policy at the construction site; the `reads` block is that policy's single source. The admin,
  domain, and reactor layers ship empty; what they compose is the application author's
  decision. The `/admin` group serves on the API listener until the application settles its
  isolation, before the first admin service is mounted.
- **Generation** — the constraints that keep the module cleanly copyable with `gonew`: the
  subtree boundary, everything surviving the path rewrite, the starter README's identity steps.
  Re-checked whenever a file is added.
- **CI and release** — built: CI runs vet, race tests, and lint inside `template/`; releases are
  `template/v*` tags cut from the root `CHANGELOG.md`.
- **Candidate direction** — the admin layer's content and the database infrastructure patterns
  (the data package, the admin route group, the conventions lint, the compose stack and its
  tasks) are reference-architecture patterns: go-web-service proves them at
  `v1.data.sql.integration.service` and the docs pass documents them. The template stays
  engine-free and standardizes only what the reference service has proven stable.
- **Candidate direction** — the scaffolding CLI (`concepts/scaffolding-cli.md`); it waits on
  the reference service.
