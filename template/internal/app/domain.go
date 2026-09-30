package app

import (
	"log/slog"

	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

// Domain composes the application's domain services, one field per domain
// layer. Domain services are defined in the module's base-layer domain
// packages, one package per layer under domain/, and constructed here from
// the infrastructure fields they depend on. The template ships it empty;
// which services it composes is the application author's decision.
type Domain struct{}

// newDomain wires the domain layer over infra: each domain package's service
// is constructed here from the infrastructure fields it uses, never the
// Infrastructure struct itself. It takes no lifecycle coordinator: domain
// services own no resource and never run, so there is nothing in scope to
// register. A package that does own a resource and run belongs in
// infrastructure.go, if it is built once at startup, or reactors.go, if it
// reacts to an external occurrence for the life of the process.
func newDomain(infra *Infrastructure) *Domain {
	return &Domain{}
}

// mountAPI builds the API mount, /api, with each domain layer's route group
// mounted into it. Each handler is handed, at the construction site, its
// policy from cfg (cfg.Reads.Limits() for a collection read) and logger for
// its error writer (web.NewErrorWriter(logger, ...)). The template ships
// the group initialized and empty.
func mountAPI(dom *Domain, cfg *config.Config, logger *slog.Logger) *web.Group {
	return web.NewGroup("/api")
}
