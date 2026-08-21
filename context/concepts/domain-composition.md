# Domain composition

Whether the template should pre-stage packages for domain services: a `domain` package where
domain services are constructed and an `entity` package for the entities they operate on, or
whether route registration in `internal/app/routes.go` stays the only build point through
which a domain service enters the application.

Deliberately deferred. No domain service exists yet anywhere in the ecosystem, and the
template standardizes only patterns proven in application code first. go-web-service supplies
the evidence: how its domain services are constructed, which `Infrastructure` fields they
consume, and how much structure their assembly actually needs. Revisit once go-web-service has
built its first domain services.
