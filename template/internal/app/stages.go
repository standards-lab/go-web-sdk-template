package app

import "github.com/standards-lab/go-core/lifecycle"

// The stage table: every lifecycle stage the process uses, named once, in
// the process's dependency order. A stage is the composition root's
// decision, so each layer file registers its services at a stage named
// here, never at a number of its own. The coordinator starts the stages
// ascending, the services within one concurrently, and drains them in
// reverse, so the table read upward is the drain order.
const (
	// stageInfrastructure holds the connections everything else runs over,
	// such as a database pool or an object store, each with its readiness
	// check (infrastructure.go). The template registers nothing at this
	// stage; the constant stays unused until the first service registers here.
	stageInfrastructure = 0 //nolint:unused // the first infrastructure service registers here

	// stageRoot is the request edge: the server (app.go), which starts
	// after every other stage and drains first.
	stageRoot = lifecycle.StageRoot
)
