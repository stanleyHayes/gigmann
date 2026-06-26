# ADR-0004: Deterministic signal engine with grounded, cached Claude narration

- **Status:** accepted
- **Date:** 2026-06-26
- **Deciders:** Owner (Stanley) + engineering

## Context
The product's defining promise (CLAUDE.md §1) is that **a deterministic signal engine computes all
numbers and the AI never invents a figure** — Claude only narrates and prioritises. We needed an
architecture that makes that promise structurally enforceable rather than a matter of prompt
discipline, and we needed the Daily Brief to feel instant despite Claude taking ~15s to narrate.

## Decision
- **Numbers come only from the domain.** The signal engine in `internal/core/**` (pure Go) computes
  every metric, threshold breach, and ranking. The narration layer receives those computed figures as
  structured input and is asked to phrase and order them — never to calculate.
- **Constrained tool output.** The Anthropic adapter forces Claude through strict tools
  (`emit_brief` / `emit_answer`) whose schemas carry the already-computed values. A **grounding
  guardrail** rejects narration that introduces figures or entities not present in the supplied data,
  so a hallucinated number fails closed rather than reaching the executive.
- **Read-through cache with background refresh** (`internal/app/cached_brief.go`). A cold cache
  generates synchronously; a warm cache serves immediately and refreshes in the background. The server
  warms the brief at startup so the first real request is hot.
- **Deterministic fallback.** With no `ANTHROPIC_API_KEY`, a deterministic narrator renders the same
  computed figures in templated prose, so the cockpit is fully functional and testable offline.

## Consequences
- The "never invent a figure" rule is enforced in code (constrained tools + guardrail + domain-only
  computation), not just requested in a prompt. The domain and signal engine are tested toward ~100%
  coverage independent of any model.
- The brief is effectively instant after warm-up; the ~15s model latency is hidden by the cache and
  the startup warm. The trade-off is brief staleness up to the cache TTL, which is acceptable for a
  daily executive brief and bounded by the background refresh.
- The model is swappable and non-critical: an outage or missing key degrades to the deterministic
  narrator rather than failing the brief.
- AI output is **read-only narration** — consistent with CLAUDE.md §7, it never triggers a
  side-effect; actions (approvals, tasks) require explicit user confirmation through their own
  endpoints.

## Alternatives considered
- **Let the model compute or restate numbers freely** — rejected: directly violates the core promise
  and invites hallucinated figures in front of an executive.
- **Synchronous narration on every brief request** — rejected: ~15s per request is not acceptable for
  the hero surface; caching + startup warm removes it from the critical path.
- **No deterministic fallback (hard dependency on Claude)** — rejected: would make local dev, tests,
  and any model outage block the entire brief.
