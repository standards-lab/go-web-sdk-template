package app

import (
	"github.com/standards-lab/go-core/graph"
)

// defineReactors defines the reactors on g into n: the entry points that
// run for the process lifetime, driven by an occurrence (a subscription, an
// interval, a wake on demand) rather than a caller, the inbound counterpart
// to a route. A reactor often dispatches each occurrence to a domain service
// call, but need not. Each reactor is its own node, a lifecycle participant
// whose constructor Uses the infrastructure nodes for the transport
// connection it owns and the domain nodes it dispatches to, the two halves
// a reactor joins. No node Uses a reactor, so each is appended to
// n.Reactors, which Run builds as roots and the server orders itself after.
// The template defines none; which reactors it runs, and what drives them,
// is the application author's decision.
func defineReactors(g *graph.Graph, n *Nodes) {}
