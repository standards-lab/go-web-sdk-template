// Package config declares the service's layered configuration.
//
//   - [Config] is the root: the library capability blocks (log, server),
//     the service-owned read policy, and the shutdown timeout, with Merge
//     and Finalize delegating to each block.
//   - [ReadsConfig] is the service-owned read policy, handed to a handler
//     as web.Limits by its Limits method; [ReadsEnv] records the
//     environment-variable names its Finalize read.
//   - [Load] reads the layered files, strictly, and finalizes the result
//     under the service's env prefix.
//
// The unexported envPrefix const is the single place a seeded service
// renames its environment namespace.
package config
