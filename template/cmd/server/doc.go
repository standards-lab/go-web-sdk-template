// The server binary, one concern per file:
//
//   - main.go    process entrypoint: signal context, config load, exit code
//   - server.go  the server type and its phases — newServer is the cold
//     start (construction and coordinator binding, zero I/O); serve is the
//     hot start plus shutdown, delegated to lifecycle.Run
//   - routes.go  router assembly, and the narrowing point where route groups
//     receive their dependencies
//
// Shutdown is ordered: a single composite hook drains the HTTP server before
// the infrastructure beneath it closes.
package main
