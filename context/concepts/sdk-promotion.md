# SDK promotion

The template's application machinery is the proving ground for API the SDKs may own: the
graduation rule says a proven pattern sinks to the lowest level at which it is generic. Two
candidates settled into shape this session and wait on use.

- **The infrastructure registry** (`internal/infrastructure`) is application-generic — nothing
  in it is specific to a web service — so it graduates toward `go-core`, beside `lifecycle`.
  Its `Register` and `Get` are parameterized methods, so `go-core` moves to Go 1.27 before the
  package can land there.
- **The application layer** (`internal/app`) splits. The router assembly, probe registration,
  and server binding are web-specific and graduate toward `go-web-sdk`; the lifecycle
  composition around the registry is application-generic and follows the registry toward
  `go-core`. Housing an `app` package in the SDK also resolves the naming collision a
  reference service otherwise carries between its own server type and `web.Server`.

Neither promotion happens from this repository. The graduation plan session decides the split
and the receiving APIs once the reference service has exercised them, and the roadmap re-plan
sequences it. Until then the template's copies are the authority, and a generated service uses
them as ordinary source.
