package app

import (
	"github.com/standards-lab/go-core/graph"
)

func defineReactors(g *graph.Graph, n *Nodes) {
	n.Reactors = g.Group("reactors")
}
