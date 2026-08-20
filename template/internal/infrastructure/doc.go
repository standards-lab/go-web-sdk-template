// Package infrastructure is the type-keyed registry of the services an
// application is composed on — the logger in the template baseline; a
// database pool, storage, or auth client as a service grows. A service is
// constructed and registered once, in the composition root's manifest, and
// [Registry.Register] carries its handle and its lifecycle declaration in
// the same call. The registry then drives startup in registration order and
// shutdown in reverse ([Registry.Start], [Registry.Shutdown]), and feeds the
// readiness probe ([Registry.Checks]).
//
// Retrieval is by type ([Registry.Get]) and stops at the composition layer:
// the manifests and the application layer read the registry, and domain
// packages receive their dependencies as constructor parameters, never the
// registry itself. One instance per type is the contract; roles sharing a
// type (a write pool and a read pool) are distinguished by defined wrapper
// types, and a dynamic set of like services registers as one service that
// owns its members.
package infrastructure
