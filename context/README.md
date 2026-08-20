# go-web-sdk-template

The web service template of Go Minimal, the Standards Lab organization's minimal-dependency Go
standard: scaffolds an initial Go Minimal web service application with `gonew`, built on
go-core and go-web-sdk at pinned releases, and the first code expression of the Elemental
Architecture's application layout.

The design and conventions of this repository are documented in the organization's
[documentation landing zone](https://github.com/standards-lab/docs); this context records only
working knowledge the landing zone and the code do not express. The repository page is
[go-web-sdk-template](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/go-web-sdk-template/index.md),
under the [Go Minimal](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/index.md)
standard, with the design detailed in
[The baseline](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/go-web-sdk-template/baseline.md),
[The elements in the template](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/go-web-sdk-template/elements.md),
and
[The template subtree](https://github.com/standards-lab/docs/blob/main/standards/go-minimal/go-web-sdk-template/template-subtree.md).

## Capability map

The code and each package's `doc.go` are authoritative for what is built; the landing zone
documents the design.

- **Runnable baseline** — built: the phase-structured composition root on the Elemental
  Architecture layout — the `cmd/server` manifests, the `internal/app` application layer, the
  type-keyed `internal/infrastructure` registry, and the `internal/config` root — with the
  probes, the ordered drain, and the test suite.
- **Generation** — the constraints that keep the module cleanly copyable with `gonew`: the
  subtree boundary, everything surviving the path rewrite, the starter README's identity steps.
  Re-checked whenever a file is added.
- **CI and release** — built: CI runs vet, race tests, and lint inside `template/`; releases are
  `template/v*` tags cut from the root `CHANGELOG.md`.
- **Candidate direction** — the scaffolding CLI (`concepts/scaffolding-cli.md`) and the
  graduation of the application machinery into the SDKs (`concepts/sdk-promotion.md`); each
  waits on the reference architecture, and the roadmap re-plan decides what is next.
