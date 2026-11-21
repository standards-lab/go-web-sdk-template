# Admin layer and database patterns

Planned: the admin layer's content and the database infrastructure patterns (the data package,
the admin route group, the conventions lint, the compose stack and its tasks) are
reference-architecture patterns go-web-service proves. The template stays engine-free and
standardizes only what the reference service has proven stable.

Open points:

- Which of the patterns the reference service has proven stable enough to standardize.
- How the template takes them while staying engine-free.
- The admin mount's isolation (its own listener, authentication, and audit), which the template
  leaves to the application until the first admin service arrives.
