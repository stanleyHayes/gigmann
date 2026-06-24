# AGENTS.md

Operating guide for any automated agent or contributor working in this repository. Companion to
[`CLAUDE.md`](./CLAUDE.md) (the full rules) and [`agent_plan.md`](./agent_plan.md) (the plan/tracker).

## Quick orientation
- **Product:** Gigmann Executive Cockpit ("Ahenfie") — AI chief-of-staff for a Ghana hospital network.
- **Plan/tracker:** `agent_plan.md` (replaces Jira). Work maps to story IDs `GEC-###`.
- **Stack:** Go (hexagonal) · Chi · REST+OpenAPI · pgx+sqlc · Postgres+pgvector (Render) · Redis ·
  WebSocket · Claude Sonnet · React+Vite · MUI v9 · TanStack Query.

## Golden rules
1. **Deterministic compute, AI narration.** Numbers are computed in code; the model never invents figures.
2. **Respect hexagonal boundaries.** `internal/core` and `internal/app` import no adapters/frameworks
   (enforced by an architecture test).
3. **Tests + coverage > 80% on every change.** SonarQube quality gate must pass.
4. **Reference a story.** Branch/commit/PR carry `GEC-###`; update `agent_plan.md` in the same PR.
5. **Don't touch unrelated code.** Minimal, reviewable diffs.
6. **Security first.** Secrets in env only; validate inputs; enforce authz in the app layer; no PII in logs.

## Roles (from the engineering manuals)
- **Claude** — planning, documentation, structured implementation.
- **Kimi** — research and analysis.
- **Codex** — code generation.

## Where things live
```
backend/internal/core/      domain entities + signal engine (pure)
backend/internal/ports/     interfaces
backend/internal/app/       use cases (authz here)
backend/internal/adapters/  inbound (http/chi) + outbound (postgres/redis/anthropic)
backend/cmd/                composition root
backend/api/                OpenAPI spec (contract)
frontend/src/               React + Vite SPA (MUI v9)
infra/render.yaml           Render Blueprint
docs/adr/                   architecture decisions
```

## Before you open a PR
- [ ] `make backend-cover-gate` passes (>80%).
- [ ] `make backend-lint` / frontend lint+typecheck clean.
- [ ] OpenAPI + generated code regenerated (for API changes).
- [ ] `agent_plan.md` story status + dashboard updated.
- [ ] No secrets, no PII in logs, no unrelated changes.
