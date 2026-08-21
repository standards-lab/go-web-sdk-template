// Package app is the composition root: [App] is the primitive that
// orchestrates the root composition behavior, and this package holds the
// manifests it composes from — middleware.go declares the router-level
// middleware stack, outermost first, and routes.go declares the
// domain-service modules to mount. Extending the service means editing a
// manifest body; the signatures, cmd/server, and [App.Run] stay untouched.
//
// [New] is the cold start, with no I/O: it creates the lifecycle
// coordinator, constructs the infrastructure (each service registering on
// the coordinator where it is constructed), assembles the router from the
// manifests, and declares the server as the coordinator's root-stage
// service — started after every infrastructure stage and drained first, so
// in-flight requests complete before the infrastructure beneath them closes.
// The probes register on the router's native mux, outside every module's
// middleware, aggregating the coordinator under the "lifecycle" name ahead
// of the services' own checks. Wiring mistakes panic at construction.
//
// [App.Run] is the hot start plus shutdown, delegated to the coordinator,
// and returns the process exit code.
package app
