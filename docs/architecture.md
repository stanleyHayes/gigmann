# Architecture — Gigmann Executive Cockpit ("Ahenfie")

A production-shaped, AI-native executive cockpit for a 12-facility healthcare
network in Ghana. The hero feature is the **Daily Brief**. A **deterministic
signal engine computes every number**; Claude only narrates and prioritises — the
AI never invents a figure (see [ADR-0004](adr/0004-deterministic-engine-grounded-narration.md)).

## 1. Style: hexagonal (ports & adapters)

```
            inbound adapters                 outbound adapters
          ┌───────────────────┐            ┌────────────────────────────┐
HTTP ───► │ adapters/inbound/  │            │ adapters/outbound/postgres │ ──► Postgres + pgvector
 (Chi)    │ httpapi (strict    │            │ adapters/outbound/memory   │ (in-memory fallback)
          │ OpenAPI server)    │            │ adapters/outbound/anthropic│ ──► Claude (Messages API)
          └─────────┬─────────┘            │ adapters/outbound/voyage   │ ──► Voyage embeddings
                    │                       │ adapters/outbound/local*   │ (deterministic fallbacks)
                    ▼                       │ passwordhash / token / audit│
          ┌───────────────────┐            └──────────────┬─────────────┘
          │  internal/app      │  use cases (orchestrate ports, hold authz)
          └─────────┬─────────┘
                    ▼
          ┌───────────────────┐   ┌───────────────────┐
          │ internal/ports    │◄──│ internal/core     │  pure domain + signal engine
          │ (interfaces)      │   │ (no I/O, ~100% cov)│
          └───────────────────┘   └───────────────────┘
```

Dependencies point **inward**: adapters → app → ports → core. The rule is enforced
by `internal/architecture/arch_test.go` (core/app/ports may not import adapters or
frameworks). `cmd/api` + `internal/bootstrap` are the composition root (wiring only).

## 2. Stack
- **Backend:** Go 1.25, Chi router, oapi-codegen strict server, pgx + sqlc,
  pgvector, golang-jwt, argon2id, Prometheus client, log/slog.
- **Frontend:** React 19 + Vite (Rolldown), MUI v9 + MUI X Charts, TanStack Query,
  React Router v7 (lazy routes), openapi-fetch typed client, vite-plugin-pwa.
- **AI:** Claude (narration/Ask, grounded + constrained tools); Voyage embeddings
  for NL facility search; deterministic local fallbacks for both.
- **Infra:** Render Blueprint (`infra/render.yaml`); CI in `.github/workflows/ci.yml`.

## 3. Request flow (Daily Brief)
1. `GET /api/v1/brief` → `httpapi` handler → `BriefGenerator` (cached) use case.
2. The **signal engine** (`internal/core/signal`) computes signals from the read
   models (metrics/inventory/staff) and ranks them worst-first.
3. The **Narrator** port phrases the *already-computed* figures via Claude's strict
   `emit_brief` tool; a grounding guardrail rejects any figure/entity not supplied.
4. A read-through cache serves the brief instantly; a background refresh + startup
   warm hide the model latency. With no API key, a deterministic local narrator
   renders the same figures.

## 4. Data model (Postgres)
`migrations/000001..000004`: facilities, facility_metrics (+ index, materialized
view `network_daily_metrics`), inventory_items, staff, alerts, tasks, approvals,
briefs, insights, users + credentials + refresh_tokens, facility_embeddings
(`vector(1024)` + HNSW). Money is stored in **minor units (pesewas, bigint)** —
never float. See [ADR-0005](adr/0005-metrics-storage-and-aggregates.md) and
[ADR-0006](adr/0006-nl-retrieval-embeddings.md).

## 5. Persistence & fallbacks
`DATABASE_URL` switches all repositories to Postgres (migrations auto-applied with
an advisory lock; first-run seed is atomic). Without it, the app runs entirely
in-memory from the synthetic network. See
[ADR-0002](adr/0002-in-memory-store-and-deferred-persistence.md).

## 6. Security posture
HS256 JWT access tokens + single-use rotating refresh tokens (hashes only); RBAC at
the use-case boundary; managers scoped to their facility (no IDOR); per-IP rate
limiting on auth; CORS allow-list; security headers; structured request + audit
logging; TOTP MFA. See [ADR-0003](adr/0003-jwt-and-refresh-token-rotation.md),
[docs/security/threat-model.md](security/threat-model.md), and §7 of `CLAUDE.md`.

## 7. ADR index
- [0001 Architecture & stack](adr/0001-architecture-and-stack.md)
- [0002 In-memory store & deferred persistence](adr/0002-in-memory-store-and-deferred-persistence.md)
- [0003 JWT & refresh-token rotation](adr/0003-jwt-and-refresh-token-rotation.md)
- [0004 Deterministic engine & grounded narration](adr/0004-deterministic-engine-grounded-narration.md)
- [0005 Metrics storage & aggregates](adr/0005-metrics-storage-and-aggregates.md)
- [0006 NL retrieval embeddings](adr/0006-nl-retrieval-embeddings.md)
