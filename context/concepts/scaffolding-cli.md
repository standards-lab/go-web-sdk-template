# Scaffolding CLI

A CLI tool that scaffolds infrastructure service integrations into a generated web service:
where `gonew` generates the baseline once, the tool would add a proven integration — a service
from the infrastructure libraries, wired the way the reference architecture documents — into an
existing generated service, landing its configuration block, its `Infrastructure` field, and
its lifecycle declaration in `infrastructure.New`.

Deferred until the reference architecture and its supported patterns are robust enough to make
the scaffolds worth standardizing. Revisit when the reference service has several documented
layers proven.
