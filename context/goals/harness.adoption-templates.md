# goal · harness.adoption-templates

- **State:** building
- **Task:** adopt
- **Branch:** adopt

## Tasks

1. [ ] adopt

## Task brief · adopt

**Problem**

The two Go templates predate marathon 0.22.0: no pin, no context lints, comments outside the five
kinds, and no rule that suppressions give a reason. They scaffold that state into every generated
service and CLI. The task adopts both and releases them, so new projects start adopted.

**Behaviors**
1. The web template module requires go-web-sdk v0.18.0. Both repositories' CI copies, root and
   template, run mise-action v5.2.10.
2. The template's mise configuration pins marathon 0.22.0 with auto-install off and a `See` line.
   Currency reports a newer marathon tag when one exists.
3. The check runs the Go gates, then lint-context and lint-references in the project root and
   again at the repository top level when the project is nested, running all of them before
   failing. A lint missing from PATH fails with `FAIL: <lint> is not on PATH; run mise install`.
   The task description names both lints. A generated project runs them once, at its root.
4. lint-context passes at both roots: map entries are a name, `(built)` or `(planned)`, and a
   relative link to their home; the candidate directions are planned entries with their open
   points in notes; the web starter README is within budget and links the `docs/` page holding its
   moved detail, content unchanged.
5. Every comment in tracked code, tests, scripts, and configuration, generated code included, is
   one of the five kinds. Unexported and test doc comments go; their reasoning moves into package
   docs, names, the starter README, or its `docs/` page. The currency script's sections become
   named functions. Configuration headers are a short description plus `See` lines.
6. Each template's golangci-lint configuration requires every suppression to name its linter and
   give a reason (nolintlint with require-explanation and require-specific).
7. Each STANDARDS.md has the comment rule covering every configuration file, and its
   release-and-ci item names the lints.
8. The root and template copies of CI and `.gitignore` still differ only in subtree lines and
   management-layer entries. CI is green.
9. Each CHANGELOG cuts its release section listing what a generated project inherits (the
   suppression rule, the pin and lints in the check, the comment and README moves, and for web the
   go-web-sdk v0.18.0 requirement); link references are updated and Unreleased is left empty.
10. Both tags are pushed and their release workflows succeed.

**Test seams**
- `mise -C template run check` and `mise -C template run currency` in each repository; the check
  in a gonew-generated copy of each template.

**Slices**
1. go-web-sdk-template upgrade: go-web-sdk v0.18.0 and mise-action v5.2.10.
2. go-cli-sdk-template upgrade: mise-action v5.2.10.
3. go-web-sdk-template comments (behaviors 5, 6).
4. go-web-sdk-template pin, check, context, STANDARDS (behaviors 2-4, 7).
5. go-cli-sdk-template comments (behaviors 5, 6).
6. go-cli-sdk-template pin, check, context, STANDARDS (behaviors 2-4, 7).
7. Release sections in both CHANGELOGs (behavior 9).

**Out of scope**

docinfer; adopting already-generated services; using `web.NewClient` in the template; changing the
template's behavior.

**Door**

One-way for the two tags, which reach every future `gonew`. Everything before the tags is two-way.

**Release**

go-web-sdk-template `template/v0.16.0`, then go-cli-sdk-template `template/v0.3.1`.

## Progress

slices 7/7 committed · standards ✓ · spec ✓ · editor ✓

## Decisions

## Pending edits

- coordinator · context/harness-adoption.md, "Changelogs record requirements": a template's
  release section lists what a generated project inherits. (T2)
- architecture · standards/go-elemental/principles/release-and-ci.md, "One check per repository":
  a template's check runs the lints in `template/` and at the repository root, and generated
  projects inherit them.
