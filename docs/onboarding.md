# Engineer Onboarding (GEC-114)

Get the Gigmann Executive Cockpit running locally in ~10 minutes.

## Prerequisites
- Go 1.25+, Node 20+, (optional) Docker for Postgres/pgvector + Redis.
- Read `CLAUDE.md` (operating rules) and `docs/architecture.md`.

## Fast path (in-memory, no Docker)
```bash
# Backend (serves the synthetic 12-facility network in-memory)
make backend-run                 # http://localhost:8080  (JWT_SECRET defaults in dev)

# Frontend
cd frontend && npm ci && npm run dev   # http://localhost:5173 (proxies /api → :8080)
```
Demo logins (password `ahenfie-demo` unless `DEMO_PASSWORD` is set):
- Executive: `ceo@gigmann.health`
- Manager (Kasoa): `kasoa.manager@gigmann.health`

## With persistence (Postgres + pgvector)
```bash
make dev-up                      # docker compose: pgvector/pgvector:pg16 + redis
export DATABASE_URL='postgres://gigmann:gigmann@localhost:5432/gigmann?sslmode=disable'
make backend-run                 # auto-migrates + first-run seeds
```

## Real AI (optional)
Put keys in `backend/.env` (git-ignored): `ANTHROPIC_API_KEY` (Claude narration/Ask),
`VOYAGE_API_KEY` (semantic facility search). Without them, deterministic fallbacks run.
Load with `set -a; . backend/.env; set +a; make backend-run`.

## Daily commands
```bash
make backend-test          # unit tests + coverage
make backend-cover-gate    # enforce >80%
make backend-integration   # testcontainers (needs Docker)
make lint                  # golangci-lint
make generate              # regenerate OpenAPI Go server + TS client + sqlc
cd frontend && npm run lint && npm run typecheck && npm run test:coverage && npm run build
```

## Repo map
- `backend/internal/core` — pure domain + signal engine (no I/O).
- `backend/internal/app` — use cases (authz lives here).
- `backend/internal/ports` — interfaces; `adapters/{inbound,outbound}` — implementations.
- `backend/cmd` + `internal/bootstrap` — composition root.
- `frontend/src/{screens,components,api,app}` — SPA.
- `agent_plan.md` — the backlog (replaces Jira); `docs/adr/` — decisions.

## Workflow
Branch `feature/GEC-123-...`; commit `GEC-123 ...`; update the story + dashboard in the
same PR; keep `core`/`app` framework-free (the arch test enforces it).
