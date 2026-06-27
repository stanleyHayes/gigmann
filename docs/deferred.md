# Deferred / externally-blocked work

Items that need live infrastructure, a third-party account, a browser runtime, or a
human/business process — with the design + what's prepared, so they can be picked up
directly.

## GEC-67 — Realtime (WebSocket live updates)
**Design:** `coder/websocket` hub on the API (per-user channels) fed by Redis
pub/sub; the SPA opens a socket after auth and invalidates the relevant TanStack
Query caches on a server "changed" event. **Needs:** Redis enabled on Render + a
connection-scaling decision (single instance vs. fan-out). Not built — the current
brief is cached + pre-warmed, which covers the demo without realtime.

## GEC-68/70 — Material-change brief invalidation & alert dedup
Hooks for the realtime path above: emit a "material change" event when a signal
crosses a threshold (invalidate the brief cache) and dedup alerts by an idempotency
key (`facility:type` + window). Small once GEC-67 lands.

## GEC-69 — Push notifications (critical only)
**Design:** Web Push (VAPID) — SW `push` handler + a `subscriptions` table; the API
sends only `critical` alerts. **Needs:** generated VAPID keys, browser testing, and
user opt-in UX. Browser-runtime + keys required.

## GEC-82 — Pre-production penetration test
**Prepared:** threat model (`docs/security/threat-model.md`) + CI SAST/deps/secret/
container scanning. **Needs:** a deployed staging URL (GEC-107/111) and an engaged
pen-test firm/scope — a human/procurement process.

## GEC-83/84/85/118 — Public marketing site + SEO
The cockpit is correctly `noindex` (private app). SEO (pre-render, JSON-LD, sitemap)
attaches to a **public marketing site (GEC-118)**, which is a separate deliverable
needing brand/content direction. The SPA SEO approach is recorded in ADR-0001 (D-006).

## GEC-94/95 — SLOs/alerting & backups/DR
Artifacts shipped (`infra/observability/slo.md` + `alert-rules.yml`,
`docs/backups-and-dr.md`); the **alert receiver** and a **tested restore drill** need
the live Render project + a notification channel account.
