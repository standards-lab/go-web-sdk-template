package app

import (
	"context"
	"io"
	"log/slog"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"

	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

// App is the application layer: the dependency graph that describes the
// service and the [Nodes] that name its parts.
type App struct {
	graph *graph.Graph
	nodes Nodes
	ran   bool
}

// Nodes is the App's graph, one handle per node: the single description of
// what the service is composed of. Each layer file's define function fills
// its own part of one Nodes value, and its constructors read the lower
// layers' nodes from that same value. Each field's node or group is named
// for the field, its words lowercased and spaced, so LifecycleConfig holds
// "lifecycle config", the name the lifecycle labels its errors with; the
// field names the node, and no comment repeats it.
type Nodes struct {
	// The App's inputs (app.go): the configuration and the logger's
	// writer New was given, each the same value in every Build.
	Config *graph.Node[*config.Config]
	Writer *graph.Node[io.Writer]

	// Infrastructure (infrastructure.go).
	LifecycleConfig *graph.Node[lifecycle.Config]
	Logger          *graph.Node[*slog.Logger]

	// Reactors (reactors.go): the group each reactor joins with graph.In.
	// A reactor is reached by no Use, so the group is one of [App.Roots],
	// and the server orders itself after it so it stays the top layer.
	Reactors *graph.Group

	// The request edge (server.go).
	Readiness *graph.Node[*lifecycle.Readiness]
	Router    *graph.Node[*web.Router]
	Server    *graph.Node[*web.Server]
}

// New describes the graph over cfg and a writer for the service's logger,
// and returns the App. It is cold: it constructs nothing and cannot fail;
// [App.Run] builds. cfg and w are fixed-value nodes, the config node and
// the writer node, which the layers reach with Use like any other; then
// each layer file describes its nodes, lowest first, so every handle a
// Define line orders after is already defined.
func New(cfg *config.Config, w io.Writer) *App {
	a := &App{graph: graph.New()}
	a.nodes.Config = a.graph.Value("config", cfg)
	a.nodes.Writer = a.graph.Value("writer", w)
	defineInfrastructure(a.graph, &a.nodes)
	defineAdmin(a.graph, &a.nodes)
	defineDomain(a.graph, &a.nodes)
	defineReactors(a.graph, &a.nodes)
	defineServer(a.graph, &a.nodes)
	return a
}

// Graph returns the graph Run builds from, as [New] described it. The
// service itself never calls it; it is published so a caller can, before
// Run, observe what the Build constructs with graph.Graph.Observe, or
// Replace a node's constructor with a substitute.
func (a *App) Graph() *graph.Graph { return a.graph }

// Nodes returns a handle on each of a's graph nodes, for a caller's Replace
// or Observe.
func (a *App) Nodes() Nodes { return a.nodes }

// Roots returns the service's root set, the nodes a Build of the whole
// service starts from: the server, which reaches the router, the logger,
// and the configuration through its Uses, and the reactors group, whose
// members no Use reaches. [App.Run] builds this set plus the lifecycle
// config node the Coordinator is configured from.
func (a *App) Roots() []graph.Ref {
	n := a.nodes
	return []graph.Ref{n.Server, n.Reactors}
}

// Run runs the service as a web.Service over [App.Roots] until ctx ends,
// and returns the process exit code. It builds the root set and the
// lifecycle config node and hands the System to a lifecycle Coordinator
// configured from that node. Once the server has bound, it logs "server
// ready" with the bound address. It logs "server stopped" and returns 0
// after a clean drain, or logs "service failed" with the error, which
// names the failing node, and returns 1 on a startup, runtime, or shutdown
// failure. A Build failure is written to stderr instead, as "service
// failed: " and the error, since the logger node may be what failed. An
// App runs once: a second Run panics before it builds anything.
func (a *App) Run(ctx context.Context, stderr io.Writer) int {
	// web.Service holds no state between Runs, so the once-only guard is
	// the App's.
	if a.ran {
		panic("app: Run called twice; an App runs once")
	}
	a.ran = true

	n := a.nodes
	return web.Service{
		Graph:           a.graph,
		LifecycleConfig: n.LifecycleConfig,
		Logger:          n.Logger,
		Server:          n.Server,
		Roots:           a.Roots(),
	}.Run(ctx, stderr)
}
