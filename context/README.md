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

- **Runnable baseline** — built: the composition root on the Elemental Architecture layout, in
  four layers — `internal/infrastructure`, `internal/domain`, `internal/reactors`, and
  `internal/app` — each constructed in that order onto go-core's staged coordinator, plus the
  `internal/config` root and the `cmd/server` entrypoint, with the probes, the staged drain, and
  the test suite. The route build point ships the one initialized empty `/api` module, with the
  config root in scope so handlers take their policy at the construction site. `domain` and
  `reactors` ship empty; what they compose is the application author's decision.
- **Generation** — the constraints that keep the module cleanly copyable with `gonew`: the
  subtree boundary, everything surviving the path rewrite, the starter README's identity steps.
  Re-checked whenever a file is added.
- **CI and release** — built: CI runs vet, race tests, and lint inside `template/`; releases are
  `template/v*` tags cut from the root `CHANGELOG.md`.
- **Candidate direction** — the scaffolding CLI (`concepts/scaffolding-cli.md`); it waits on the
  reference service, and the roadmap re-plan decides what is next.
