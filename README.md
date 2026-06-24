# Gigmann Executive Cockpit ("Ahenfie")

An AI-native executive "chief of staff" for a multi-facility healthcare network in Ghana.
The hero is the **Daily Brief** — a morning brief that surfaces the three things that need the CEO today,
worst first, in plain English. A deterministic signal engine computes the numbers; Claude narrates them.

> **Planning / tracking:** [`agent_plan.md`](./agent_plan.md) is the source of truth for epics, stories,
> and status (it stands in for Jira). Every change references a story ID `GEC-###`.
> **Rules:** see [`CLAUDE.md`](./CLAUDE.md) and [`AGENTS.md`](./AGENTS.md).

## Stack
- **Backend:** Go (hexagonal / ports & adapters) · Chi · REST + OpenAPI · pgx + sqlc · custom JWT
- **Data:** PostgreSQL 16 + pgvector (Render-managed) · Redis · WebSocket (`coder/websocket`)
- **AI:** Anthropic Claude (Sonnet) — deterministic signal engine narrated, never invented
- **Frontend:** React + Vite (SPA) · MUI v9 · MUI X Charts · TanStack Query · React Hook Form + Zod
- **Infra:** Render Blueprint (`infra/render.yaml`) · GitHub Actions · SonarQube · >80% coverage gate

## Layout
```
backend/    Go API (hexagonal): cmd/ internal/{core,app,ports,adapters} api/ migrations/
frontend/   React + Vite SPA (MUI)
infra/       render.yaml Blueprint
docs/        ADRs and architecture docs
.github/     CI workflows
```

## Quickstart
```bash
# Local dependencies (Postgres 16 + pgvector, Redis)
make dev-up            # docker compose up

# Backend
make backend-test      # go test with coverage
make backend-run       # run the API locally

# Frontend
cd frontend && npm install && npm run dev
```

See the `Makefile` for all targets. New here? Read `agent_plan.md` §0, then `CLAUDE.md` and `AGENTS.md`.
