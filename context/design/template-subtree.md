# The template subtree

Why the module is rooted at `template/` rather than the repository root, and the duties the
layout creates.

## The generation boundary is structural

A module's zip is its `go.mod`-rooted subtree: the proxy packs every committed file below
`go.mod` and nothing above it, and `gonew` copies that zip with the module path rewritten. With
the module at the repository root, every generated service inherited the template's management
layer — `context/`, `.claude/`, `CLAUDE.md`, the changelog — and excluding any of it required
post-generation cleanup instructions. With the module at `template/`, the boundary is the
directory line: a generated service receives the subtree, the management layer cannot ship, and
no cleanup instructions exist to be skipped.

## Costs accepted

- The module path is `github.com/standards-lab/go-web-sdk-template/template`. Generated
  services never see it: `gonew` rewrites the whole path.
- Version tags are `template/v*`. The prefix is not a naming choice — the proxy resolves a
  subdirectory module's versions from tags prefixed with exactly that subdirectory — and it is
  the one place the directory name is externally visible, which is why the directory is named
  `template` rather than `src`.
- The root and subtree each keep their own `.gitignore` and CI workflow. The root workflow runs
  inside `template/` (`working-directory` on its steps); the subtree copies serve the generated
  service and are inert in this repository, since GitHub executes only the root `.github`. A
  change to either copy is applied to the other in the same commit. `mise.toml` lives only in
  the subtree — development runs inside `template/`, so the root carries no copy.

## What the layers carry

The root: `context/`, `CLAUDE.md` and `.claude/`, the template's README and `CHANGELOG.md` (the
source of the release notes), the committed `go.work` that lets `go` commands and gopls resolve
the module from the root, the CI and release workflows, and the license. The subtree: the
service source, its config files, `mise.toml` and the tooling copies, the starter README
addressed to the generated service, and the license — pkg.go.dev requires a license inside the
module subtree to render documentation.

Deliberately absent from the subtree: release automation and a changelog. Both encode
conventions the generated service's author owns.
