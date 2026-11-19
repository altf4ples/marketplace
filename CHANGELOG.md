# Changelog

## [1.0.0] - 2026-11-19

- Completed user, catalog, cart, order, and notification modules.
- Added Redis Hash cart with Lua-based atomic mutations and checkout compensation.
- Added transactional checkout with item price/name snapshots.
- Added order state machine, row locking, state history, and transactional DB notifications.
- Added unified JSON errors, request validation, request IDs, health checks, and graceful shutdown.
- Added Docker Compose, migrations, OpenAPI documentation, CI, and GitHub Release workflow.
