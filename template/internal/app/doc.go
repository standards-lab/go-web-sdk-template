// Package app is the composition root, laid out as one file per layer of
// the architecture, so the package's file list is the architecture's layer
// list:
//
//   - infrastructure.go: [Infrastructure], the services the application is
//     composed on;
//   - admin.go: [Admin], the administrative services, and their /admin mount;
//   - domain.go: [Domain], the domain services, and their /api mount;
//   - reactors.go: [Reactors], the event-driven entry points;
//   - stages.go: the stage table every layer file registers from;
//   - routes.go: the list of mounts;
//   - middleware.go: the router-level middleware stack, outermost first.
//
// [App] is the assembled process. [New] is the cold start, with no I/O: it
// constructs infrastructure (each service registering on the coordinator
// where it is constructed, at a stage the table names), the admin layer
// over it, the domain over it, and the reactors over both, then assembles
// the router from the mounts and the middleware stack. The server is the
// coordinator's root-stage service, started after every other stage and
// drained first, so in-flight requests complete before the infrastructure
// beneath them closes. The service's logger reaches the server, the
// middleware, and every mount's error writer. The probes register on the
// router's native mux, outside every module's middleware, and report the
// coordinator under the "lifecycle" name. Wiring mistakes panic at
// construction. [App.Run] is the hot start plus shutdown, delegated to the
// coordinator, and returns the process exit code.
//
// Routes and reactors are the two ways a domain service enters the running
// process: a route is driven by a caller, a reactor by an occurrence the
// process receives or discovers. Both take *Domain; neither is a domain
// service itself. Extending the service means editing a layer file's body;
// the signatures, cmd/server, and [App.Run] stay untouched.
package app
