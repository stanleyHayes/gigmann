# ADR-0006: Natural-language retrieval via embeddings (pgvector + Voyage)

- **Status:** accepted
- **Date:** 2026-06-27
- **Deciders:** Owner (Stanley) + engineering

## Context
GEC-13 (spec §6.4) wants grounded natural-language queries that fuzzy-resolve to
facilities — "how is the Kasoa polyclinic doing" should find the `kasoa` facility.
Anthropic exposes no embeddings API, so an embedding provider had to be chosen.
The product also follows a **deterministic-fallback** rule (ADR-0004): it must run
offline with no external key, the way the Daily Brief degrades to the local
narrator.

## Decision
- **Embedding provider: Voyage AI** (Anthropic's recommended embedding partner),
  via its REST API (`POST /v1/embeddings`, `Authorization: Bearer`). Voyage has no
  Go SDK, so the adapter calls REST directly (no new dependency). Selected by
  `VOYAGE_API_KEY`; default model `voyage-3.5-lite` at `output_dimension` 1024.
- **Deterministic local fallback.** When `VOYAGE_API_KEY` is unset, a local
  embedder (feature-hashed bag-of-words → unit vector) is used. It is **lexical,
  not semantic**, but deterministic, offline, and good enough that shared words
  rank the right facility first — verified on the seeded network.
- **Storage: pgvector.** `facility_embeddings(facility_id, content, vector(1024))`
  with an **HNSW** cosine index; an in-memory brute-force repo serves the no-DB
  path. Vectors are passed as a `::vector` text literal, so no `pgvector-go`
  dependency is needed.
- **Write path.** Facilities are embedded (name/region/town/type/manager) at
  first-run, idempotently and best-effort — a provider failure logs a warning and
  leaves search empty; it never blocks startup.
- **Query path.** `FacilitySearchService` embeds the query (`input_type=query`),
  ANN-searches, and returns ranked matches, exposed at the authed, read-only
  `GET /api/v1/facilities/search?q=…`.
- **Fixed dimension (1024)** aligns the column, the Voyage request, and the local
  embedder so write and query vectors share one space.

## Consequences
- Runtime-verified against native **Postgres 18 + pgvector 0.8.3**: the full
  migration chain (incl. HNSW), the write path, and NL resolution all pass — even
  the lexical local embedder resolves "Assin Fosu specialist hospital" →
  `assin-fosu` (score 0.89), "Tamale North clinic" → `tamale-north`. Semantic
  quality improves with Voyage configured.
- The vector path requires pgvector — available on Render's managed Postgres and
  in the CI testcontainers image (`pgvector/pgvector:pg16`); the in-memory path
  needs nothing.
- Switching embedder (local ↔ Voyage, or model/dimension) changes the vector
  space, so embeddings must be regenerated. Acceptable: they are derived data and
  the write path is idempotent (re-seed after a truncate).
- Voyage is an added external dependency and cost when enabled; it is optional and
  off by default.

## Alternatives considered
- **OpenAI / other embeddings** — rejected: not Anthropic-aligned; Voyage is the
  recommended partner.
- **A `pgvector-go` dependency** — rejected as unnecessary; the `::vector` text
  cast keeps the dependency surface flat.
- **Semantic ranking in SQL only** — n/a; embeddings are the mechanism, and KPI
  numbers remain a separate, Go-computed concern (ADR-0004).
- **Skip embeddings, keep keyword matching** — rejected: misses the fuzzy/NL
  resolution the spec calls for.
