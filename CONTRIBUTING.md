# Contributing

1. Create a feature branch.
2. Keep business rules inside the owning `internal` module.
3. Depend on interfaces at module boundaries; wire concrete implementations in `cmd/api`.
4. Add or update unit tests for behavior changes.
5. Run `make check` and `docker compose config`.
6. Use migrations for schema changes; include both `.up.sql` and `.down.sql`.
7. Do not commit `.env`, credentials, dumps, or generated binaries.

Pull requests should describe the behavior, design trade-offs, and verification performed.
