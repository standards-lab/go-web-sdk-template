package app_test

import (
	"io"
	"testing"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-web-sdk"

	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config/configtest"
)

// This file pins the graph's one structural property the composition root
// owns: the server is alone in the top layer. What that position buys at
// run time — startup in layer order, readiness held until every check
// passes, the server's runtime error watched as Monitored, and the drain in
// reverse layer order — is pinned in go-core's lifecycle tests and
// go-web-sdk's health and server tests, not here.

// buildAsRun builds a's graph from the roots Run builds from, the config,
// the logger, the server, and the reactors, plus extra.
func buildAsRun(t *testing.T, a *app.App, extra ...graph.Ref) *graph.System {
	t.Helper()
	n := a.Nodes()
	roots := append([]graph.Ref{n.Config, n.Logger, n.Server}, n.Reactors...)
	sys, err := a.Graph().Build(append(roots, extra...)...)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return sys
}

// serverAloneOnTop reports whether the System's last layer holds the server
// node alone, and returns that layer.
func serverAloneOnTop(sys *graph.System, n app.Nodes) ([]graph.Dependency, bool) {
	layers := sys.Layers()
	if len(layers) == 0 {
		return nil, false
	}
	top := layers[len(layers)-1]
	alone := len(top) == 1 &&
		top[0].Name == n.Server.Name() &&
		top[0].Value == any(sys.Get(n.Server))
	return top, alone
}

// The server is the only node in the top layer of the graph Run builds, so
// it starts after every other node and drains first. A node defined above
// it, or beside it at the top, fails here; the second case shows the check
// catches one.
func TestGraph_ServerAloneInTopLayer(t *testing.T) {
	t.Run("as defined", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		sys := buildAsRun(t, a)
		if top, ok := serverAloneOnTop(sys, a.Nodes()); !ok {
			t.Errorf("top layer = %+v, want the server node alone", names(top))
		}
	})

	t.Run("a node above the server", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		n := a.Nodes()
		above := a.Graph().Define("above", func(s *graph.Scope) (*web.Server, error) {
			return s.Use(n.Server), nil
		})
		sys := buildAsRun(t, a, above)
		if top, ok := serverAloneOnTop(sys, n); ok {
			t.Errorf("top layer = %+v, reported the server alone with a node above it", names(top))
		}
	})
}

// names lists the layer's node names.
func names(layer []graph.Dependency) []string {
	out := make([]string, len(layer))
	for i, d := range layer {
		out[i] = d.Name
	}
	return out
}
