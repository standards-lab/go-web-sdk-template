# go-web-sdk-template

go-web-sdk-template is the web service template of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. It scaffolds an initial Go
Elemental web service application with `gonew`, built on go-core and go-web-sdk at pinned
releases, and is the first code expression of the Elemental Architecture's application layout.

The README, the starter README, and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the READMEs do not express.

## Capability map

- **Runnable baseline** (built): [internal/app/doc.go](../template/internal/app/doc.go).
- **Generation** (built): [README.md](../README.md).
- **Integration tier** (built): [integration/doc.go](../template/integration/doc.go).
- **CI and release** (built): [ci.yml](../.github/workflows/ci.yml).
- **Admin layer and database patterns** (planned): [admin-and-database-patterns.md](admin-and-database-patterns.md).
- **Scaffolding CLI** (planned): [scaffolding-cli.md](scaffolding-cli.md).
