# Domain composition

Whether the template should pre-stage packages for domain services — a `domain` package where
they are constructed, an `entity` package beside it — or keep the `routes` manifest
(`internal/app/routes.go`) as the only place a domain service enters the application.

Deliberately deferred: no domain service exists yet, and the template grows only patterns
proven in application code. go-web-service settles the question — how its domain services are
constructed, what they take from `Infrastructure`, and how many files a new one touches will
show what structure, if any, the template should standardize. Revisit when the reference
service has its first domain services built.
