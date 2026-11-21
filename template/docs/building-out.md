# Building out the service

The layer files it builds on are listed in the [starter README](../README.md#building-out-the-service).

Each layer file defines its layer's nodes on go-core's dependency graph, in its define function
(`defineInfrastructure`, `defineAdmin`, `defineDomain`, `defineReactors`), and owns its mount;
`server.go` defines the request edge (the readiness the probes report, the router, the server),
and `routes.go` lists the mounts and does nothing else. The `Nodes` value holds one handle per
node, each named for its field. Each define function fills its part, declaring each node's order
and group membership on its `Define` line, and a constructor, a method on `*Nodes`, reads the
lower layers' nodes from it with `Use`. `New` describes the graph and cannot fail.
`App.Roots` is the root set, the server and the reactors group; `Run` hands it to go-web-sdk's
`web.Service`, which builds it and runs the service under go-core's lifecycle Coordinator. A
Build failure is written to stderr as `service failed: <error>`, naming the failing node.
`cmd/server` is the entrypoint alone and never changes.

A domain service starts from its Entity:

1. Give the Entity its own package under `domain/`.
2. Expose its Queries and Commands as the domain service's methods.
3. Build the layer's route group in its handler.
4. Define the service's node in `defineDomain` and mount the group in `mountAPI` (`domain.go`).

The constructor draws what it uses from the infrastructure nodes it `Use`s, and the handler is
handed its policy from the config node at the construction site (`cfg.Reads.Limits()` for a
collection read) and the service's logger for its error writer (`web.NewErrorWriter(logger, ...)`).

An infrastructure service (a database pool, a storage client, an auth client) is a field on
`Nodes` plus its node, defined in `defineInfrastructure` (`internal/app/infrastructure.go`) and
named for the field, with its constructor a method on `*Nodes` —

```go
n.DB = g.Define("db", n.newDB)

func (n *Nodes) newDB(s *graph.Scope) (*db.Pool, error) {
	return db.New(s.Use(n.Config).DB, s.Use(n.Logger)), nil
}
```

Its part in the lifecycle is inferred from its value's methods, with no registration: a
`lifecycle.Starter` is started and a `lifecycle.Stopper` shut down (a `Subsystem` is both), a
`lifecycle.ReadyStarter` (a `Starter` that also implements `Ready`) joins `/readyz` under the node's
name, and a `Monitored` has its runtime error watched. A value with `Ready` but no `Start` joins no
`/readyz` check. Its constructor opens nothing; connectivity belongs to its `Start`, so a failed
build leaks no connections. Nodes start in layer order, each in a layer above the nodes it uses, and
drain in reverse; the server is alone in the top layer, so it starts last and drains first, and
in-flight requests complete before their infrastructure closes. A service defined this way cannot be
missing from the probe or the drain, and a `Nodes` field that does not exist fails compilation at
its access.

An admin service administers one infrastructure service over the mechanisms its library
provides. It is a node defined in `defineAdmin` (`internal/app/admin.go`), with its route group
mounted under `/admin` in `mountAdmin`. One that verifies and corrects the state of the
infrastructure it administers is a lifecycle participant by its methods, and the domains that
depend on that state `Use` it, so it starts ahead of them. The template serves the empty
`/admin` group on the API listener; before the first admin service is mounted, settle the
mount's isolation — its own listener, authentication, and audit — because an administrative
surface on the public port is exposed the moment it serves a route.

A reactor is an entry point that runs for the process lifetime, driven by an occurrence (a
subscription, an interval, a wake on demand) rather than a caller. Each is a node defined in
`defineReactors` (`internal/app/reactors.go`), whose value is a go-core `reactor.Reactor`.
`reactor.New` builds one over a Source, a handler, and a `reactor.Options`; the Source is
`reactor.Every` for a plain interval, or `reactor.Wake` for an interval a nudge can also trigger.
`reactor.Options.CancelAfter` sets how long the reactor's drain lets the handling in flight run
before it cancels it, and `reactor.CancelWithin` returns half the shutdown timeout for that
field. Under the zero `reactor.Options{}`, the drain cancels the handling only when its context
ends. No node uses a reactor, so it joins the `Nodes.Reactors` group with `graph.In` on its
`Define` line: the group is one of `App.Roots`, and the server orders itself after it. A reactor
that expires sessions every minute is defined this way:

```go
n.Expiry = g.Define("expiry", n.newExpiry, graph.In(n.Reactors))

func (n *Nodes) newExpiry(s *graph.Scope) (*reactor.Reactor[time.Time], error) {
	return reactor.New(reactor.Every(time.Minute), s.Use(n.Sessions).Expire, reactor.Options{
		CancelAfter: reactor.CancelWithin(s.Use(n.Config).ShutdownTimeout.Duration()),
	}), nil
}
```

Its lifecycle is inferred the same as an infrastructure service's: the Coordinator starts it,
drains it, reports its readiness under the node's name, and ends the run on its failure, with no
registration. A reactor often dispatches each occurrence to a domain service call, but need not;
a background worker the service runs is a reactor too.

Middleware that applies to every route stacks in `middleware` (`internal/app/middleware.go`),
outermost first. The template ships `RequestID`, `RequestLogger`, and `Recoverer`, in the chain
order that go-web-sdk's `middleware` package documents. Middleware scoped to one domain service
belongs on its route group.

Configuration grows by adding fields to `Config` in `internal/config/config.go` and delegating
to their `Merge` and `Finalize` in the existing shape; the `reads` block is the model for a
service-owned block. The service keeps pace with its SDKs by updating its `go-core` and
`go-web-sdk` pins and applying whatever adjustments the release notes call for.
