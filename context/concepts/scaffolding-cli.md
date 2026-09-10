# Scaffolding CLI

A CLI tool that scaffolds infrastructure service integrations into a generated web service.
Where `gonew` generates the baseline once, this tool would add a proven integration (a service
from the infrastructure libraries, wired the way the reference architecture documents) into an
existing generated service. It would land the integration's configuration block, its
`Infrastructure` field, and its lifecycle declaration in `infrastructure.New`.

The tool is deferred until the reference architecture and its supported patterns are robust
enough to make these integration scaffolds worth standardizing. Revisit when the reference
service has several documented layers proven.
