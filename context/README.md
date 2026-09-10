# go-web-sdk-template

go-web-sdk-template is the web service template of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. It scaffolds an initial Go
Elemental web service application with `gonew`, built on go-core and go-web-sdk at pinned
releases, and is the first code expression of the Elemental Architecture's application layout.

The README, the starter README, and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the READMEs do not express.

## Capability map

The code and each package's `doc.go` are authoritative for what is built.

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
  a first backing service.
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
