// The server binary: the process entrypoint and the three manifests that
// declare its composition, one concern per file.
//
//   - main.go            entrypoint: signal context, config load, manifest
//     evaluation, exit code
//   - infrastructure.go  setInfrastructure, the infrastructure-service
//     manifest: every service constructed and registered once, in
//     dependency order
//   - middleware.go      setMiddleware, the router-level middleware stack,
//     outermost first
//   - routes.go          setRoutes, the domain-service modules to mount
//
// The machinery the manifests feed lives in internal/app; extending the
// service means adding an entry to a manifest, never touching the machinery.
package main
