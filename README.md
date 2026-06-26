# Gigmann Executive Cockpit ("Ahenfie")

An AI-native executive "chief of staff" for a 12-facility healthcare network in Ghana.
The hero is the **Daily Brief** — a morning brief that surfaces the few things that need the CEO today,
worst first, in plain English. **A deterministic signal engine computes every number; Claude only narrates and
prioritises them — the AI never invents a figure.**

> **Planning / tracking:** [`agent_plan.md`](./agent_plan.md) is the source of truth for epics, stories, and
> status (it stands in for Jira). Every change references a story ID `GEC-###`.
> **Rules:** [`CLAUDE.md`](./CLAUDE.md) and [`AGENTS.md`](./AGENTS.md).

## Stack
- **Backend:** Go 1.25, hexagonal (ports & adapters) · Chi · REST + OpenAPI 3 (oapi-codegen) · custom HS256 JWT
- **AI:** Anthropic Claude (Sonnet 4.6) — narrates the deterministic signal engine via a strict tool, grounded
- **Frontend:** React 19 + Vite 8 (SPA) · MUI v9 · MUI X Charts · TanStack Query · React Router v7 (lazy routes)
- **Infra:** Render Blueprint (`infra/render.yaml`) · GitHub Actions CI · golangci-lint v2 (all linters) · >80% coverage gate
- **Data (demo):** in-memory, seeded from a deterministic synthetic network. PostgreSQL 16 + pgvector + Redis are
  wired in the Blueprint but commented out until the persistence layer (GEC-11/12/13) lands.

## Run it locally
```bash
# 1. Backend API (port 8080) — in-memory, no database required
make backend-run

# 2. Frontend SPA (port 5173) — Vite dev-proxies /api -> :8080
cd frontend && npm install && npm run dev
```
Then open http://localhost:5173 and **sign in**:

| Field    | Value                                         |
|----------|-----------------------------------------------|
| Email    | `ceo@gigmann.health`                          |
| Password | `ahenfie-demo` (override via `DEMO_PASSWORD`) |

### Live Claude brief & Ask (optional)
Without a key, a deterministic **local narrator** renders the brief/answers so everything works offline.
To use the real model, drop a key into `backend/.env` (git-ignored — never commit it):
```
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-sonnet-4-6   # optional
```
Run the API with it loaded: `set -a; . backend/.env; set +a; make backend-run`.
The brief is generated once and cached (served in ~30 ms); `/ask` calls the model per question.

## Screens
`Today` (the Daily Brief) · `Network` (all 12 facilities, worst-first) · `Executive KPIs` (revenue / patients /
NHIS denial rate / occupancy with 14-day trends) · `Ask` (grounded NL Q&A) · `My Day` (tasks) ·
`Approvals` (decide with a confirmation step — no AI-triggered side-effects).

## API
```
GET  /healthz · /readyz                                  liveness / readiness
POST /api/v1/auth/login | /auth/refresh | /auth/logout   (login/refresh rate-limited 10/min/IP)
GET  /api/v1/auth/me                                     (Bearer-token protected)
GET  /api/v1/facilities | /brief | /metrics             (deterministic; brief cached)
GET  /api/v1/approvals    POST /api/v1/approvals/{id}/decision   (executive-only)
GET  /api/v1/tasks        POST /api/v1/tasks/{id}/status
POST /api/v1/ask                                         (grounded natural-language query)
```
All `/api/v1/**` business endpoints require a valid access token. The OpenAPI contract is
[`backend/api/openapi.yaml`](./backend/api/openapi.yaml); the Go server stubs and the typed TS client are
generated from it (`make generate`).

### Security
Argon2id password hashing · HS256 JWT access tokens (15 min) + single-use rotating refresh tokens (7 days) ·
RBAC enforced at the use-case boundary (executive vs facility-manager, facility scoping) · per-IP login rate
limiting · CORS allow-list · security headers · structured request + audit logging (no PII).

## Quality gates
```bash
make test            # backend: go test -race + coverage gate (>80%, ~93% today)
make lint            # backend: golangci-lint v2 (all linters)
make backend-integration   # testcontainers (needs Docker); live Claude test: go test -tags=integration with a key
cd frontend && npm run lint && npm run typecheck && npm run test:coverage
```

## Deploy (Render)
[`infra/render.yaml`](./infra/render.yaml) is a one-click Blueprint: a Dockerised Go **web service** (`gigmann-api`,
health check `/healthz`) plus a **static SPA** (`gigmann-frontend`). In the `gigmann-secrets` env group set
**`JWT_SECRET`** (required) and optionally `ANTHROPIC_API_KEY` / `DEMO_PASSWORD`. The API reads
`CORS_ALLOWED_ORIGINS` (the SPA URL) and the SPA builds with `VITE_API_BASE_URL` (the API URL). The demo runs
fully in-memory, so no database is needed to boot.

## Layout
```
backend/    Go API: cmd/api · internal/{core,app,ports,adapters,intel,seed,config,bootstrap} · api/openapi.yaml
frontend/   React + Vite SPA: src/{app,screens,components,api,auth}
infra/      render.yaml Blueprint
docs/       ADRs / architecture
```
New here? Read `agent_plan.md` §0, then `CLAUDE.md` and `AGENTS.md`. Run `make help` for all targets.
