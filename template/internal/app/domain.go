package app

import (
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-web-sdk"
)

func defineDomain(g *graph.Graph, n *Nodes) {}

func mountAPI(s *graph.Scope, n *Nodes) *web.Group {
	return web.NewGroup("/api")
}
