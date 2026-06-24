# ADR-0001: Architecture & technology stack

- **Status:** accepted
- **Date:** 2026-06-24
- **Deciders:** Owner (Stanley) + engineering

## Context
The Gigmann Executive Cockpit must be a production-ready, secure, SEO-aware, AI-native product. The PoC spec
recommended a NestJS + Next.js + Tailwind stack on Vercel/Neon. The owner directed a Go-based backend with
hexagonal architecture and selected the remaining technologies interactively (see `agent_plan.md` §2).

## Decision
Adopt the following, with rationale captured per decision:

- **D-001 — Backend: Go + hexagonal (ports & adapters).** Strong typing, performance, and a clean domain core
  for the deterministic signal engine. Enforced by `internal/architecture/arch_test.go`.
- **D-002 — Auth: custom JWT in Go.** Full control, no vendor in the critical path. We own the security surface
  (rotation, lockout, MFA) — addressed in epic E2/E9 using vetted libraries.
- **D-003 — Realtime: WebSocket (`coder/websocket`) + Redis pub/sub.** Owner choice over SSE; matches the spec's
  Socket.io intent with an idiomatic Go library.
- **D-004 — Hosting: Render Blueprint (`infra/render.yaml`).** Declarative IaC; stateless API so the eventual
  Ghana-hosting move (spec §8.3) is a deployment decision, not a rebuild.
- **D-005 — Seed: Go `cmd/seed`.** One backend language; seed shares domain types.
- **D-006 — Frontend: React + Vite (SPA).** Owner preference over Next.js. SEO preserved via SSG/pre-render of
  public pages (epic E10); the private cockpit is `noindex`.
- **D-007 — UI: MUI v9 + MUI X Charts.** Owner preference over Tailwind/shadcn/Tremor. Typography:
  Fraunces (titles), Outfit (body), JetBrains Mono (statuses).
- **API contract:** REST + OpenAPI (codegen both sides). **DB:** Render-managed Postgres 16 + pgvector
  (TimescaleDB dropped — see consequences). **DB access:** pgx + sqlc. **AI:** Claude Sonnet (latest).

## Consequences
- Clean inward-pointing dependencies; domain is testable without infrastructure (94%+ coverage at kickoff).
- **TimescaleDB dropped (RESOLVED, OQ-4).** Render's managed Postgres supports pgvector but not TimescaleDB,
  and we are committed to Render. Decision: native Postgres time-series — indexes now, with declarative range
  partitioning + materialized views as the scale-up path. Dev/prod parity via the `pgvector/pgvector:pg16`
  image. (A future ADR can revisit a dedicated TSDB if data volume ever demands it.)
- SPA SEO requires deliberate pre-rendering and meta handling (owned in E10), unlike the spec's SSR assumption.

## Alternatives considered
- NestJS backend (spec default) — rejected per owner directive favouring Go.
- Auth.js / Clerk / Keycloak — rejected in favour of custom JWT for control; revisitable if ops cost grows.
- SSE for realtime — viable (push-dominant traffic) but owner chose WebSocket.
