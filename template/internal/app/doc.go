// Package app is the application layer: the composition and runtime seam
// between the declarative manifests in cmd and the SDKs beneath. [New] is
// the cold start — router assembly from the wiring, the probes on the native
// mux, server construction, and coordinator binding, with no I/O — and
// [App.Run] is the hot start plus shutdown, delegated to go-core's lifecycle
// coordinator. Shutdown is ordered: one composite hook drains the HTTP
// server before the infrastructure beneath it closes.
package app
