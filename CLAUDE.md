# go-web-sdk-template

The `go-minimal` standard's web service template: the minimal runnable web service new services
are generated from with `gonew`, built on `go-core` and `go-web-sdk`. It is the seam between
the standard's SDK level and its reference level. Managed with the marathon workflow; start
from `context/README.md`.

## Conventions are settled in the repository

The design and conventions for this template are recorded in `context/design/` — that is the
authority. Keep them there; do not restate them here.

## Role boundary

go-web-sdk-template is a marathon **code** project (`.claude/marathon.toml` declares
`kind = "code"`). The developer owns the production Go source — they apply it and answer for
it. The agent writes everything else: tests, godoc and `doc.go`, prose documentation, the files
in `context/`, the implementation guide, and the reset file.

## Repository specifics

- **Module layout.** The module is `github.com/standards-lab/go-web-sdk-template/template`,
  rooted at `template/`; the repository root above it is the template's management layer and
  never ships in the module. See `context/design/template-subtree.md`.
- **Dependencies.** `go-core` and `go-web-sdk` at pinned releases, on Go 1.27. The template is
  engine-free: no data engine declared, no provider imported.
- **Generation.** `gonew github.com/standards-lab/go-web-sdk-template/template@latest
  example.com/newsvc`. Everything in the subtree must survive the path rewrite; re-check
  whenever a file is added.
- **Releases.** The module is tagged `template/v<semver>` from the root `CHANGELOG.md`, cut by
  `.github/workflows/release.yml`.
- **Tasks.** `template/mise.toml` carries the tasks; development runs inside `template/`. The
  subtree's `.gitignore` and CI workflow are separate copies of the root ones (the dual-copy
  rule: a change to either copy lands in the other in the same commit).
- **Public repo.** This repository is public on GitHub.
