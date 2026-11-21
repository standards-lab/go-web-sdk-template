// Package app is the composition root. It describes the service as one
// go-core dependency graph, and each architecture layer has one file, which
// defines the layer's nodes in its define function and owns its mount, so
// the layer files are the architecture's layer list:
//
//   - infrastructure.go: defineInfrastructure, the lifecycle configuration
//     and the services the application is composed on (the logger, in the
//     template baseline);
//   - admin.go: defineAdmin, the administrative services, and their /admin
//     mount;
//   - domain.go: defineDomain, the domain services, and their /api mount;
//   - reactors.go: defineReactors, the process-lifetime entry points an
//     occurrence drives, each a node whose value is a go-core
//     reactor.Reactor;
//   - server.go: defineServer, the request edge: the readiness the probes
//     report, the router, and the server;
//   - routes.go: the list of mounts;
//   - middleware.go: the router-level middleware stack, outermost first.
//
// [Nodes] holds one handle per node, the single description of what the
// service is composed of: each define function fills its own part of one
// Nodes value, declaring each node's order and group membership on its
// Define line, and its constructors, methods on *Nodes, read the lower
// layers' nodes from it with Use; [Nodes] states how each field names its
// node. [App] is the described process. [New] is the cold start: it
// describes the graph, layer by layer, lowest first, constructs nothing,
// and cannot fail. [App.Graph] and [App.Nodes] publish the graph and its
// handles, so a caller can Observe or Replace a node before Run. [App.Run]
// is the hot start plus shutdown: it runs a go-web-sdk web.Service, which
// builds [App.Roots], the server and the reactors group, plus the lifecycle
// config node, hands the System to go-core's lifecycle Coordinator, and
// returns the process exit code. A constructor's error fails the Build,
// and Run writes "service failed: " and the error, which names the failing
// node, to the stderr it is given; a wiring mistake panics during Build.
//
// What a node takes part in is inferred from its value's methods: a
// lifecycle.Starter or Stopper is started or stopped, a Subsystem is both, a
// ReadyStarter (a Starter that also implements Ready) joins the readiness
// probe, and a Monitored has its runtime error watched. A value with Ready
// but no Start joins no readiness check. Nodes start in layer order and
// drain in reverse. The server orders itself after the reactors group, so
// it is alone in the top layer: it starts after every other node and drains
// first, so in-flight requests complete before what they run on closes. The
// service's logger reaches the server, the middleware, and every mount's
// error writer. The probes register on the router's native mux, outside
// every module's middleware, query live on every request through the
// readiness node's value, and report the Coordinator's status under the
// "lifecycle" name, ahead of the nodes' own checks in layer, then
// definition, order.
//
// Routes and reactors are the two ways the running process is entered: a
// caller drives a route, and an occurrence the process receives or
// discovers drives a reactor. A route reads the domain nodes, and a
// reactor may; neither is a domain service itself. Extending the service
// means defining a node in its layer's define function, and joining a
// reactor to the Nodes.Reactors group with graph.In on its Define line;
// the signatures, cmd/server, and [App.Run] stay untouched.
//
// # The layers
//
// The infrastructure layer holds the lifecycle config the Coordinator is
// configured from and the services the application is composed on, read
// from the config node: the logger, written to the writer node, in the
// template baseline; a database pool, storage, or auth client as a service
// grows, each its own node, with any order it needs declared on its Define
// line. A service whose value implements lifecycle.Subsystem takes part in
// startup and shutdown through its own methods, and its constructor opens
// nothing: connectivity belongs to its Start, so a failed Build leaks no
// connections.
//
// The admin layer is the administrative counterpart of the domain layer:
// one node per admin domain, each administering one infrastructure service
// over the library mechanisms it triggers and reading that service's node
// with Use. One that verifies and corrects the state of the infrastructure
// it administers is a lifecycle participant, and the domains that depend
// on that state Use it. Each admin domain's route group mounts into the
// /admin group, with its service and the logger for its error writer read
// with Use. The template defines none and ships the /admin group empty;
// which admin services a service composes follows the infrastructure it
// adopts. In production the admin mount belongs on its own listener,
// authenticated and unreachable from the public API's network path, an
// isolation the application settles when the first admin service arrives.
//
// The domain layer has one node per domain layer. Domain services are
// defined in the module's base-layer domain packages, one package per
// layer under domain/, and each node's constructor builds its service from
// the infrastructure nodes it Uses, never from Nodes itself. Domain
// services own no resource and never run, so none is a lifecycle
// participant. A package that does own a resource and run belongs in the
// infrastructure layer, if it is built once at startup, or the reactors
// layer, if it reacts to an external occurrence for the life of the
// process. Each domain layer's route group mounts into the /api group, its
// handler handed, at the construction site, its service, its policy from
// the config node (cfg.Reads.Limits() for a collection read), and the
// logger for its error writer (web.NewErrorWriter(logger, ...)). The
// template defines none and ships the /api group empty.
//
// A reactor runs for the process lifetime, driven by an occurrence (a
// subscription, an interval, a wake on demand) rather than a caller, the
// inbound counterpart to a route. It often dispatches each occurrence to a
// domain service call, but need not. Each reactor is its own node, a
// lifecycle participant whose constructor Uses the infrastructure nodes for
// the transport connection it owns and the domain nodes it dispatches to.
// The node's value is a go-core reactor.Reactor, which reactor.New returns
// over a Source, a handler, and a reactor.Options. The Source is
// reactor.Every for a plain interval, or reactor.Wake for an interval that
// a nudge can also trigger. reactor.Options.CancelAfter sets how long a
// drain lets the handling in flight run before it cancels it, and
// reactor.CancelWithin returns half the shutdown timeout for that field;
// under the zero Options, the drain cancels the handling only when its
// context ends. The Coordinator infers from the value's methods that the
// reactor starts with its layer, drains before the nodes it uses, joins
// the readiness probe under its node's name, and ends the run on its
// failure. The template defines none.
//
// The request edge's readiness is a fresh node, so each Build's starts
// unbound, ready for the Coordinator that Build's System is handed to. The
// probes register with the zero web.Problem, which keeps the SDK's
// readiness defaults: type about:blank, status 503 with its status text as
// the title, the generic detail, and the checks extension member. A service
// passes its own not-ready problem type once it has one.
//
// The middleware stack reads infrastructure nodes only: request logging,
// and cross-cutting concerns like it, need infrastructure primitives, not
// domain services. A middleware that has to reach a domain service is
// domain logic, and belongs in a route or a reactor instead. RequestID
// runs ahead of RequestLogger, so the log record carries the request id.
// Recoverer turns a handler's panic into a logged 500, which RequestLogger
// records with its status; it works on either side of RequestLogger, and
// the template places it inside. A BodyLimit, when a service adds one,
// goes ahead of RequestLogger and Recoverer; the chain order in
// go-web-sdk's middleware package documentation states why.
package app
