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

- **Runnable baseline** — built: the composition root `internal/app` as one file per layer on
  go-core's staged coordinator, the `internal/config` root with the `reads` policy block, the
  `cmd/server` entrypoint, the probes, the staged drain, and the suite; `internal/app/doc.go`
  and the template's README state the layout. The admin, domain, and reactor layers ship
  empty, since what they compose is the application author's decision, and the `/admin` group
  serves on the API listener until the application settles its isolation.
- **Generation** — the constraints that keep the module cleanly copyable with `gonew`: the
  subtree boundary, everything surviving the path rewrite, the starter README's identity steps.
  Re-checked whenever a file is added.
- **Integration tier** — built (`v1.data.sql.tasks.toolkit`, 2026-09-07): the `integration`
  package over the SDKs' toolkit, the `integration` task, and the CI job on merge to main.
  Engine-free: the isolated compose project is the reference service's pattern, adopted with
  a first backing service. The landing zone page is due in `v1.alignment.docs`.
- **CI and release** — built: CI runs vet, race tests, and lint inside `template/` on every pull
  request and the integration tier on merge to main; releases are `template/v*` tags cut from
  the root `CHANGELOG.md`.
- **Candidate direction** — the admin layer's content and the database infrastructure patterns
  (the data package, the admin route group, the conventions lint, the compose stack and its
  tasks) are reference-architecture patterns: go-web-service proved them at
  `v1.data.sql.integration.service` (2026-09-06) and the docs pass documents them. The template
  stays engine-free and standardizes only what the reference service has proven stable.
- **Candidate direction** — the scaffolding CLI (`concepts/scaffolding-cli.md`); it waits on
  the reference service.
