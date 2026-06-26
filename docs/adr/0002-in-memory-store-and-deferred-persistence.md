# ADR-0002: In-memory store with deferred Postgres persistence

- **Status:** accepted
- **Date:** 2026-06-26
- **Deciders:** Owner (Stanley) + engineering

## Context
ADR-0001 committed to Render-managed Postgres 16 + pgvector via pgx/sqlc, and `agent_plan.md`
carries the persistence stories (GEC-11/12/13). The integration-test strategy uses Testcontainers,
which needs a working Docker engine. During this build phase the local Docker engine was
unresponsive (daemon process up, API probes hang), so the Postgres adapter and its integration
tests could not be exercised end-to-end. We still needed a runnable, demoable cockpit — and the
brief, KPIs, approvals, tasks, and auth all have to work for the hero path to be evaluable.

## Decision
Ship the application against **in-memory adapters** that implement the same outbound ports
(`UserRepository`, `RefreshTokenStore`, `ApprovalRepository`, `TaskRepository`, …) the Postgres
adapters will implement. Seed deterministic demo data at startup. Relax `internal/config` so that
**only `JWT_SECRET` is required** outside dev; `DATABASE_URL` and `ANTHROPIC_API_KEY` are optional
with safe fallbacks (in-memory store; deterministic narrator). Comment out the Postgres/Redis blocks
in `infra/render.yaml` so the Blueprint deploys the stateless demo without provisioning stateful
services.

Because everything is already behind ports, switching to Postgres is **wiring in the composition
root** (`cmd/**`) plus the adapter + its integration tests — no domain, app, or HTTP change.

## Consequences
- The product is fully runnable and deployable today; the hero path (Daily Brief, KPIs, approvals,
  tasks, auth, MFA) is demonstrable without infrastructure.
- **Data is not durable** — a restart reseeds. This is acceptable for the PoC/demo and is called out
  in the README and the render.yaml comments. Not acceptable for production: GEC-11/12/13 remain open.
- Refresh-token rotation, rate-limit windows, and the brief cache live in process memory, so they do
  not survive a restart or span multiple instances. The Render service therefore runs a single
  instance until Postgres + Redis are wired.
- No CI integration coverage for the Postgres adapter yet — the `integration` job is a no-op stub
  until a Docker-backed runner is available.

## Alternatives considered
- **Block on Postgres** — rejected: leaves nothing demoable while Docker is down, and the port
  boundary already makes the swap cheap and low-risk later.
- **SQLite as a stand-in** — rejected: a second SQL dialect to maintain, no pgvector, and it would
  not exercise the real (pgx/sqlc) path, so it buys little over an honest in-memory store.
- **A hosted dev Postgres for local work** — deferred: viable, but a network dependency for every
  `make backend-run`; revisit if the local Docker outage persists.
