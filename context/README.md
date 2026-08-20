# go-web-sdk-template context

The `go-minimal` standard's web service template: the minimal runnable web service new services
are generated from with `gonew`, built on `go-core` and `go-web-sdk` at pinned releases. The
template is the seam between the standard's SDK level and its reference level, and the first
code expression of the Elemental Architecture's application layout
(`design/vocabulary.md`).

The template stays minimal: the SDKs carry the bulk of development change, and a generated
service keeps pace by updating its `go-core` and `go-web-sdk` pins. The baseline is engine-free
— no data engine declared, no provider imported; a generated service chooses its providers in
its own composition root. What the template owns is the composition pattern:
the three manifests in `cmd/server` and the machinery in `internal/` they feed
(`design/baseline.md`).

## Capability map

The code and each package's `doc.go` are authoritative for what is built; an unbuilt capability
gains written detail when a session is about to build it.

- **Runnable baseline** — built: the phase-structured composition root recomposed on the
  Elemental Architecture layout — the `cmd/server` manifests, the `internal/app` application
  layer, the type-keyed `internal/infrastructure` registry, and the `internal/config` root —
  with the probes, the ordered drain, and the test suite.
- **Generation** — the constraints that keep the module cleanly copyable with `gonew`: the
  subtree boundary (`design/template-subtree.md`), everything surviving the path rewrite, the
  starter README's identity steps. Re-checked whenever a file is added.
- **CI and release** — built: CI runs vet, race tests, and lint inside `template/`; releases
  are `template/v*` tags cut from the root `CHANGELOG.md`.
- **Candidate direction** — the scaffolding CLI (`concepts/scaffolding-cli.md`) and the
  graduation of the application machinery into the SDKs (`concepts/sdk-promotion.md`); each
  waits on the reference architecture, and the roadmap re-plan decides what is next.

## How this repository works

- The module is rooted at `template/`; the repository root is the management layer and never
  ships in a generated service. See `design/template-subtree.md`.
- The baseline's architecture and principles are settled in `design/baseline.md`.
- The vocabulary the template embodies is defined in `design/vocabulary.md`, citing the
  organization's Elemental Architecture definition.
