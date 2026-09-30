package app

import "github.com/standards-lab/go-core/lifecycle"

// The stage table: every lifecycle stage the process uses, named once, in
// the process's dependency order. A stage is the composition root's
// decision, so each layer file registers its services at a stage named
// here and at no number of its own. The coordinator starts the stages
// ascending, the services within one concurrently, and drains them in
// reverse, so the table read upward is the drain order.
const (
	// stageInfrastructure starts the connections everything else runs
	// over: a database pool or an object store, each with its readiness
	// check (infrastructure.go). The template registers nothing here yet.
	stageInfrastructure = 0

	// stageRoot is the request edge: the server (app.go), started after
	// every other stage and drained first.
	stageRoot = lifecycle.StageRoot
)
