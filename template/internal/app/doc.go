// Package app is the composition root. Each architecture layer has one
// file, which constructs the layer and owns its mount, so the layer files
// are the architecture's layer list:
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
// [App] is the assembled process. [New] is the cold start and performs no
// I/O. It constructs the infrastructure, each service registering on the
// coordinator where it is constructed, at a stage the table names; then the
// admin layer over the infrastructure, the domain over it, and the reactors
// over both. It then assembles the router from the mounts and the
// middleware stack. The server is the coordinator's root-stage service: it
// starts after every other stage and drains first, so in-flight requests
// complete before the infrastructure beneath them closes. The service's
// logger reaches the server, the middleware, and every mount's error
// writer. The probes register on the router's native mux, outside every
// module's middleware, query live on every request, and report the
// coordinator's status under the "lifecycle" name, ahead of the services'
// own checks. Wiring mistakes panic at construction. [App.Run] is the
// hot start plus shutdown, delegated to the coordinator, and returns the
// process exit code.
//
// Routes and reactors are the two ways a domain service enters the running
// process: a caller drives a route, and an occurrence the process receives
// or discovers drives a reactor. Both take *Domain; neither is a domain
// service itself. Extending the service means editing a layer file's body;
// the signatures, cmd/server, and [App.Run] stay untouched.
package app
