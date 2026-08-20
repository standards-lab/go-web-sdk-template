// Package infrastructure assembles the service's shared subsystems — only
// the logger in the template baseline — from the root configuration, and owns
// their participation in the process lifecycle: construction ([New], the cold
// start, no I/O), startup ([Infrastructure.Start], the hot half), ordered
// teardown ([Infrastructure.Shutdown]), and the readiness checks the router
// registers ([Infrastructure.Checks]). A new subsystem lands in all four
// places in this one package, so it cannot be constructed yet missing from
// startup, teardown, or the probe.
//
// The instance is passed, never global, and it stops at the composition
// layer: domain packages receive narrow primitives (*slog.Logger,
// configuration values) extracted at the root, never this struct and never
// the root config.
package infrastructure
