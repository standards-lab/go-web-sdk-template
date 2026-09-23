# go-web-sdk-template

go-web-sdk-template is the web service template of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. It scaffolds an initial Go
Elemental web service application with `gonew`, built on go-core and go-web-sdk at pinned
releases, and is the first code expression of the Elemental Architecture's application layout.

The README, the starter README, and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the READMEs do not express.

## Capability map

Every capability below is built except the candidate directions, and the code and each
package's `doc.go` are authoritative for it.

- **Runnable baseline**: the template ships the composition root `internal/app` (one file per
  layer on go-core's staged coordinator), the `internal/config` root with its `reads` policy
  block, the `cmd/server` entrypoint, the probes, the staged drain, and the suite.
  `internal/app/doc.go` and the template's README state the layout. The admin, domain, and
  reactor layers ship empty, because what they compose is the application author's decision,
  and the `/admin` group serves on the API listener until the application settles its isolation.
- **Generation**: three constraints keep the module cleanly copyable with `gonew`: the subtree
  boundary, every file surviving the path rewrite, and the starter README's identity steps. Each
  added file is re-checked against them.
- **Integration tier**: the `integration` package runs over the SDKs' toolkit, through the
  `integration` task and the CI job on merge to main. The tier is engine-free. Planned: with its
  first backing service, the template adopts the reference service's isolated compose project.
- **CI and release**: CI runs vet, race tests, and lint inside `template/` on every pull
  request and the integration tier on merge to main; releases are `template/v*` tags cut from
  the root `CHANGELOG.md`.
- **Candidate direction**: the admin layer's content and the database infrastructure patterns
  (the data package, the admin route group, the conventions lint, the compose stack and its
  tasks) are reference-architecture patterns go-web-service proves. The template stays
  engine-free and standardizes only what the reference service has proven stable.
- **Candidate direction**: the scaffolding CLI (`scaffolding-cli.md`), which waits on the
  reference service.
