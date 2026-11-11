package app

import (
	"log/slog"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-core/logging"
)

// defineInfrastructure defines the infrastructure nodes on g into n: the
// lifecycle config the Coordinator is configured from, and the services the
// application is composed on, read from the config node New defined — the
// logger, written to the writer node, in the template baseline; a database
// pool, storage, or auth client as a service grows, each its own node,
// with any order it needs declared on its Define line. It constructs
// nothing. A service whose value implements lifecycle.Subsystem takes part
// in startup and shutdown through its own methods, and its constructor
// opens nothing: connectivity belongs to its Start, so a failed Build
// leaks no connections.
func defineInfrastructure(g *graph.Graph, n *Nodes) {
	n.LifecycleConfig = g.Define("lifecycle config", n.newLifecycleConfig)
	n.Logger = g.Define("logger", n.newLogger)
}

// newLifecycleConfig returns the config node's lifecycle block, the
// configuration [App.Run] hands the lifecycle Coordinator.
func (n *Nodes) newLifecycleConfig(s *graph.Scope) (lifecycle.Config, error) {
	return s.Use(n.Config).Config, nil
}

// newLogger constructs the service's logger over the writer node from the
// config node's log block.
func (n *Nodes) newLogger(s *graph.Scope) (*slog.Logger, error) {
	return logging.New(s.Use(n.Writer), s.Use(n.Config).Log), nil
}
