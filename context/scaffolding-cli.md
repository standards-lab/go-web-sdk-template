# Scaffolding CLI

Planned: a CLI that adds a proven infrastructure integration to a service already generated with
`gonew`. It would add the integration's configuration block, its `Infrastructure` field, and its
construction and lifecycle declaration in `newInfrastructure` (`internal/app/infrastructure.go`).

It waits until the reference service has proven several documented integration patterns, so the
scaffolds standardize something stable.
