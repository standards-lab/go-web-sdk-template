package app

import (
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-web-sdk"
)

func defineAdmin(g *graph.Graph, n *Nodes) {}

func mountAdmin(s *graph.Scope, n *Nodes) *web.Group {
	return web.NewGroup("/admin")
}
