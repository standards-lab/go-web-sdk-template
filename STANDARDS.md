# go-web-sdk-template standards

The judgement calls the standards-reviewer applies to go-web-sdk-template, beyond what `mise -C template run check` enforces.

- A change that alters documented behavior updates the README, `template/README.md` and the affected `doc.go` in the same change.
- The root and `template/` copies of the CI workflow differ only in the lines that locate the subtree (`-C template`, `working_directory`, `working-directory`), and the two `.gitignore` files only by the root's management-layer entries.
- `architecture/standards/go-elemental/principles/dependencies.md`: the bottom-up line and no provider in a base, across `template/go.mod` and every import under `template/`.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: the doc.go inventory of `internal/app`, `internal/config`, `integration` and `cmd/server`, and the harness rules `integration` follows over `processtest` and `webtest`.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: the module rooted at `template/`, with `cmd/server` and `internal/app`'s layer files.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: the check and currency inside `template/`, the `integration` job, and `template/v<semver>` tags cut from the root `CHANGELOG.md`.
- `architecture/standards/go-elemental/principles/lifecycle-and-context.md`: `cmd/server`, `internal/app/stages.go`, and the stage each layer file registers at.
- `architecture/standards/go-elemental/principles/timeouts.md`: `template/config.json`'s `server` block and its `transfer_rate`.
- `architecture/principles/validation-first.md`: `internal/config`'s Finalize, over the `reads` block and the shutdown timeout; a new block validates in its own Finalize.
- `architecture/principles/context-architecture.md`: the README, `template/README.md` and each `doc.go` are the homes.
