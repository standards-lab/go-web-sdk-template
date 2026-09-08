# go-web-sdk-template

The web service template of Go Elemental, the Standards Lab organization's Go implementation of
the Elemental Architecture: scaffolds an initial Go Elemental web service application with `gonew`, built on
go-core and go-web-sdk. Managed with the marathon workflow; start from `context/README.md`.

## Documentation lives in the repository

This repository documents its own implementation: the README states its place in the standard
and the principles it enhances, and each package's `doc.go` is the authority for its API. The
organization's [architecture repository](https://github.com/standards-lab/architecture) states the principles this repository follows
and holds nothing a reader can infer from this source. `context/` records only working
knowledge the code and the README do not express; do not restate documented design here. A
change that alters documented behavior updates the README and the package documentation in
the same effort, and a design note that generalizes past this repository is promoted to the
architecture repository through its `context/`.

## Repository specifics

- **Module layout.** The module is `github.com/standards-lab/go-web-sdk-template/template`,
  rooted at `template/`; the repository root above it is the template's management layer and
  never ships in the module.
- **Dependencies.** go-core and go-web-sdk at pinned releases, on Go 1.27. The template is
  engine-free: no data engine declared, no provider imported.
- **Generation.** `gonew github.com/standards-lab/go-web-sdk-template/template@latest
  example.com/newsvc`. Everything in the subtree must survive the path rewrite; re-check
  whenever a file is added.
- **Releases.** The module is tagged `template/v<semver>` from the root `CHANGELOG.md`, cut by
  `.github/workflows/release.yml`.
- **Tasks.** `template/mise.toml` defines the tasks; development runs inside `template/`. The
  subtree's CI workflow is a copy of the root one minus the lines that locate the subtree,
  `working-directory` and `cache-dependency-path` (the dual-copy rule: a change to either copy
  lands in the other in the same commit). The two `.gitignore` files differ deliberately: the
  root adds the management layer's entries (`.claude/plans/`, `mise.local.toml`); the subtree
  carries only what a generated service needs.
- **Public repo.** This repository is public on GitHub.
