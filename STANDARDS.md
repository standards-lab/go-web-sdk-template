# go-web-sdk-template standards

The judgement calls the standards-reviewer applies to go-web-sdk-template.

- A change that alters documented behavior updates the README, `template/README.md` and the affected `doc.go` in the same change.
- The root and `template/` copies of the CI workflow are kept in step and differ only in the lines that locate the subtree (`-C template`, `working_directory`, `working-directory`); the two `.gitignore` files differ only by the root's management-layer entries.
- `architecture/standards/go-elemental/principles/dependencies.md`: the bottom-up line and no provider in a base, held by review; `template/go.mod` requires go-core and go-web-sdk alone, and nothing under `template/` imports a provider.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: the doc.go inventory of `internal/app`, `internal/config`, `integration` and `cmd/server`, held by review, and the harness rules `integration` follows over `processtest` and `webtest`.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: the module rooted at `template/`, `internal/app`'s file per layer, and the `cmd/server` to `internal/app` import direction.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: the check and currency inside `template/`, the `integration` job, and `template/v<semver>` tags cut from the root `CHANGELOG.md`.
- `architecture/standards/go-elemental/principles/lifecycle-and-context.md`: `cmd/server` owns the signal context and the exit code through go-core's `process`, `stages.go` names every stage a layer file registers at, and the domain layer registers nothing.
- `architecture/standards/go-elemental/principles/timeouts.md`: `template/config.json`'s `server` block, which keeps the read and write timeouts tight and carries `transfer_rate`; no route under `template/` widens them except through go-web-sdk's `Transfer`.
- `architecture/principles/composition-root.md`: `internal/app` is the composition root, handed the configuration `cmd/server` loads; `cmd/server` never changes as a generated service grows.
- `architecture/principles/validation-first.md`: `internal/config`'s Finalize validates the `reads` block and the shutdown timeout, so `app.New` never receives an unvalidated configuration; a new block validates in its own Finalize.
- `architecture/principles/context-architecture.md`: the README, `template/README.md` and each `doc.go` are the homes; `context/` records only what they do not express.
