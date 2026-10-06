package app

import (
	"github.com/standards-lab/go-core/lifecycle"
)

// Reactors composes the application's reactors: the entry points that run
// for the process lifetime, driven by an occurrence (a subscription, an
// interval, a wake on demand) rather than a caller, the inbound counterpart
// to a route. A reactor often dispatches each occurrence to a domain service
// call, but need not. The template ships it empty; which reactors it runs,
// and what drives them, is the application author's decision.
type Reactors struct{}

// newReactors constructs the reactors and registers each on lc. It takes
// infra for the transport connections a reactor owns and dom for the domain
// calls it dispatches to — the two halves a reactor joins. Each reactor owns
// a connection and runs for the process lifetime, so it registers on the
// coordinator the same as an infrastructure service.
func newReactors(
	infra *Infrastructure,
	dom *Domain,
	lc *lifecycle.Coordinator,
) (*Reactors, error) {
	return &Reactors{}, nil
}
