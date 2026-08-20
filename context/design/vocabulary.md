# Vocabulary

The Elemental Architecture terms the template embodies. The architecture is defined once, at
the organization level, in standards-lab's `context/design/elemental-architecture.md`; this
note maps its elements onto this repository and adds nothing to the definition.

- **Application** — what a generated service is: the deployable unit, here a web service. The
  template's `cmd/server` is its composition root, and the composition root only declares: the
  three manifests name the service's infrastructure services, middleware, and domain-service
  modules.
- **Application layer** — `internal/app`. It owns the infrastructure registry and the process
  lifecycle, assembles the transport, and runs the process. The machinery lives here so the
  manifests stay declarative.
- **Infrastructure Services** — the entries in `setInfrastructure`, held by the type-keyed
  registry in `internal/infrastructure`. The baseline registers the logger; a database pool,
  storage, or auth client lands as one more entry.
- **Domain Service** — arrives with a generated service, not the template. Its module is the
  unit of registration in `setRoutes`, and the baseline's empty manifest is the seam it lands
  in.
- **Entity** — arrives with the domain services; the template reserves no vocabulary below it.

"Feature" is not an element. The routes manifest registers domain-service modules, and prose
that reaches for "feature" means one Entity and its Domain Service.
