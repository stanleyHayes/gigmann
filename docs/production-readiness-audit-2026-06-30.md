# Production-Readiness & Spec-Completeness Register — 2026-06-30

> A 5-agent adversarial audit (SEO/SSG, design/UX + a11y, testing, deploy/ops,
> feature completeness) verified the implementation against the spec (agent_plan.md
> §3 DoD, §4 baselines, the PoC checklist). **The product is genuinely built** — no
> stubs, all 11 screens API-wired, the signal engine (98.8%) + brief pipeline solid,
> observability/containers/migrations/DR real. The gaps below are bounded and mostly
> in polish, deploy topology, and owner-supplied infra.

## Fixed in this pass (shipped, CI-green)
- **Config fail-fast:** reject `JWT_SECRET < 32` chars and wildcard `CORS_ALLOWED_ORIGINS` outside dev; WARN on ephemeral in-memory repos outside dev. (`4e80bce`)
- **Deploy blueprint:** `TRUST_PROXY=true`, documented VAPID/Sentry/OTel keys, one-uncomment Postgres+Redis+cron persistence path. (`4e80bce`)
- **`.env.example` + README:** rewritten to exactly the 21 vars the code reads. (`4e80bce`)
- **k6 load test now runs in CI** (weekly + dispatch). (`ed86165`)

## Open — code gaps (prioritised)

### Backend / feature
- **GEC-71 scheduled sweeps — largely unimplemented.** `cmd/worker` has only `migrate` + `refresh-views`. Missing: a **licence-expiry sweep** and a **stalled-action sweep** (proactive Web Push), and a morning brief pre-warm cron. Plan: add `worker licence-sweep` / `worker stalled-sweep` subcommands that reuse the signal detectors + `PushService`, wired as Render crons. *Value gated on VAPID keys (owner); licence-expiry already surfaces in the brief.*
- **Brief "alive"/refresh is TTL-only** (GEC-50/68): no `POST /brief/refresh` and no `?date=` param; the signal `Input` is captured once at bootstrap, so the brief changes across restarts, not within a running day. Add a refresh endpoint + a date param if live intra-day change is required.
- **Brief-quality harness (GEC-52) asserts less than claimed** — only worst-first/grounding/Tafo/fidelity; the causal-link, bright-spot, approvals-listed, personal-greeting, and latency assertions are not in code. Add them to `brief_quality_test.go`.

### Frontend (frontend agent's active area — coordinate before editing)
- **No SSG/prerender + no `react-helmet`** — the marketing surface is hand-written static HTML (`public/welcome.html` etc.), not pre-rendered from React. Biggest SEO spec gap (§4.2). Add `vite-plugin-ssg`/a prerender step + per-route metadata + a generated sitemap/robots.
- **Pagination is client-side slicing, not cursor-fetching** — `useAlerts` hard-caps at 20; hooks fetch whole lists. Wire the API cursors (`next_cursor`) into the hooks.
- **Loading-state slips:** notifications dropdown text, Reports "Preparing…"/PDF-button text → skeletons / `ButtonLoadingDots`.
- **Fonts not preloaded** (FOUT risk); **Framer-Motion `layout` transitions unused** (only fade/slide); **no skip-link + `aria-current`** (WCAG 2.4.1 / current-page).
- **CI gate tightening (blocked on the above):** LHCI has no **SEO** gate + a11y at 0.9 (spec 0.95, and the `noindex` cockpit tanks a naive SEO gate — target the public pages); frontend **branch coverage threshold 70** (actual ~76%, so 80 needs more tests first).

### Testing
- **Contract tests** validate spec-validity + codegen-drift, not response conformance (add `openapi3filter` response validation).
- **No Redis testcontainer** integration (only Postgres).

## Owner / external (cannot be done in-repo)
- Provision paid Render **Postgres (pgvector) + Redis**; set real `DATABASE_URL`/`REDIS_URL` (then uncomment the blueprint blocks).
- Real secrets: strong `JWT_SECRET`, `ANTHROPIC_API_KEY`, `VOYAGE_API_KEY`, `VAPID_*`, `SENTRY_DSN`, OTel endpoint, `DEMO_PASSWORD`.
- CI secrets: `SONAR_TOKEN` (the SonarQube gate silently no-ops without it), `RENDER_DEPLOY_HOOK_URL`, `SMOKE_*`, `DR_*`.
- A deployed **staging URL** + DNS/custom domain, then the external **pen-test (GEC-82)** and **UAT/beta sign-off (GEC-111)**.

## Deliberate designs (assessed, not bugs)
- **Network-wide views (Brief/Ask/Metrics/Facilities) are executive-only at the `requireAuth` middleware**, not re-checked in-service. This is correct: they are *cached, network-scoped* artifacts generated outside any request (the brief warms at startup with no principal). Per-*facility* resources (detail/tasks/alerts/approvals/drafts) *are* scoped in-service via `CanAccessFacility`. Fail-closed + tested (`TestNetworkAggregateViewsAreExecutiveOnly`).
- **In-memory demo posture** in production is intentional (free-tier); now WARN-logged and one-uncomment from persistence.
- **JWT HS256** (vs the AC's EdDSA) — fine for a single-service PoC.
- **Numeric grounding on live-AI prose** is a documented, accepted model-trust limit (strict `emit_brief` tool schema + supplied figures + local-narrator fallback); facility/citation grounding *is* code-enforced.
- **Managers get 403 on the network brief** (not a facility-scoped brief) — the applied fail-closed product decision.
