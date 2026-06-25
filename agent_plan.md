# Agent Plan — Gigmann Executive Cockpit ("Ahenfie")

> **This file replaces Jira.** It is the single source of truth for epics, stories, status, and traceability
> until a Jira instance is connected. Every code change must reference a story ID (`GEC-###`). Keep statuses
> current — update this file in the same PR that does the work.

| | |
|---|---|
| **Project** | Gigmann Executive Cockpit |
| **Codename** | Ahenfie ("the seat from which the chief governs everything") |
| **Project key** | `GEC` |
| **Client / persona** | Gigmann Medicals — Sammy Adjei (CEO/owner), 12-facility hospital & diagnostics network, Ghana |
| **Prepared for** | XCreativs Technologies — engineering & product |
| **Source docs** | `Gigmann_Cockpit_PoC_Spec.pdf` · `AI_Development_Workflow_Training_Manual.docx` · `AI_Native_Software_Engineering_Operations_Manual.docx` |
| **Plan version** | 1.0 |
| **Last updated** | 2026-06-24 |
| **Goal of this plan** | Take the cockpit from PoC to a **production-ready** product with **security** and **SEO** designed in from day one. |

---

## 0. How to use this file

### 0.1 Status legend
| Symbol | Meaning |
|---|---|
| `☐` | To Do |
| `◐` | In Progress |
| `☑` | Done (meets global Definition of Done) |
| `⊘` | Blocked (note the blocker inline) |
| `▷` | Deferred / out of current scope |

### 0.2 Workflow states (from the Eng-Ops manual)
`Backlog → Sprint Planning → Development → Code Review → QA → Staging → UAT → Beta → Production → Done`
A story's checkbox flips to `☑` only after it has passed the **global Definition of Done** (§3).

### 0.3 Story template (every story carries these — per the Dev Workflow manual)
```
#### <status> GEC-### — <title>   · <points> SP · Phase: <SDLC phase>
- User story:        As <role>, I want <capability>, so that <value>.
- Business value:    <why it matters to Sammy / the business>
- Acceptance criteria:
  - [ ] <testable condition>
- Technical notes:   <implementation guidance / constraints>
- Definition of done: Global DoD (§3) + <story-specific extras>
- Dependencies:      GEC-### (or "none")
```

### 0.4 GitHub conventions (mandated by both manuals)
- **Branch:** `feature/GEC-123-short-description` (also `fix/`, `chore/`, `docs/`)
- **Commit:** `GEC-123 implement signal trend detection`
- **PR title:** `GEC-123 Add signal trend detection`
- A merged PR is the only thing that flips a story to `☑`. PRs must link the story ID and update this file.

### 0.5 Estimation
Story points (Fibonacci: 1, 2, 3, 5, 8, 13). 1 SP ≈ a few hours; 8+ SP should usually be split.

---

## 1. Progress dashboard

| Epic | Title | Stories | Points | Status |
|---|---|---|---|---|
| **E0** | Foundations & Engineering Operations | 9 | 41 | ◐ In progress — GEC-1/2/5/9 done; 3/4/6/7/8 in progress |
| **E1** | Domain Model, Data Layer & Synthetic Network | 8 | 47 | ◐ In progress — GEC-10 done |
| **E2** | Authentication & Authorization | 7 | 39 | ☐ Not started |
| **E3** | Core Domain APIs (REST + OpenAPI) | 9 | 52 | ☐ Not started |
| **E4** | Signal Engine (deterministic) | 7 | 42 | ☐ Not started |
| **E5** | Intelligence Service (Claude) | 8 | 55 | ☐ Not started |
| **E6** | The Daily Brief (hero, end-to-end) | 5 | 34 | ☐ Not started |
| **E7** | Cockpit Frontend (React + Vite) | 14 | 100 | ☐ Not started |
| **E8** | Realtime, Notifications & Alerts | 5 | 26 | ☐ Not started |
| **E9** | Security Hardening & Compliance | 11 | 63 | ☐ Not started |
| **E10** | SEO & Web Performance | 7 | 31 | ☐ Not started |
| **E11** | Observability & Reliability | 7 | 37 | ☐ Not started |
| **E12** | Quality, Testing & CI Gates | 8 | 44 | ☐ Not started |
| **E13** | Deployment, Infra & Release | 7 | 38 | ☐ Not started |
| **E14** | Documentation, Governance & Handover | 6 | 24 | ☐ Not started |
| | **Total** | **118** | **673** | |

> Keep this table in sync as stories close. "Status" rolls up from the stories below.

---

## 2. Confirmed technology stack

Locked via the technology-selection prompt on 2026-06-24. Items marked **[deviation]** differ from the spec's
recommended stack (§8 of the PoC spec) — see §2.2 for the rationale (these double as ADR seeds).

### 2.1 Stack table
| Layer | Choice | Notes |
|---|---|---|
| **Backend language** | **Go (Golang) 1.23+** | **[deviation]** spec recommended NestJS. Hexagonal architecture. |
| **Architecture** | Hexagonal (Ports & Adapters) + DDD-lite | Domain core has zero framework/infra imports. |
| **HTTP framework** | **Chi router** | Idiomatic `net/http`; clean inbound adapter, easy to test. |
| **API contract** | **REST + OpenAPI 3.1** | `oapi-codegen` for Go server stubs; typed TS client for Next.js. |
| **Auth** | **Custom JWT in Go** | **[deviation]** spec recommended Auth.js. Access+refresh, rotation, RBAC, optional TOTP MFA. |
| **Frontend** | **React + Vite + TypeScript (SPA)** | **[deviation]** spec said Next.js. Public pages pre-rendered (SSG) for SEO; cockpit is a client-side SPA + PWA. |
| **UI / styling** | **MUI v9** (Material UI; latest stable at build time) | **[deviation]** spec said Tailwind+shadcn+Tremor. MUI theming carries the "command instrument" design language. |
| **Charts / data-viz** | **MUI X Charts** | KPI tiles, trends, reports; shares the MUI theme. (ECharts optional for the Ghana map view.) |
| **Forms / validation** | **React Hook Form + Zod** | Zod schemas shared with the typed API client. |
| **Motion** | Framer Motion | "Alive/thinking" feel: layout transitions, 3D reveals, parallax, circular theme-toggle reveal. |
| **Typography** | Fraunces (titles) · Outfit (body) · JetBrains Mono (statuses) | Owner direction; mono is configurable. |
| **Client data** | TanStack Query | Caching + background refresh + live feel. |
| **Database** | PostgreSQL 16 + pgvector (Render-managed) | Relational core; pgvector for NL retrieval. Time-series on native Postgres (indexes; partitioning + materialized views if volume grows). **TimescaleDB dropped** — unsupported on Render. |
| **DB access (Go)** | `pgx` + `sqlc` | Compile-time-checked SQL; repositories implement domain ports. |
| **Migrations** | `golang-migrate` (or `goose`) | Versioned, run in CD before deploy. |
| **Cache / realtime** | Redis + **WebSocket** (`coder/websocket`) | Live "always awake" updates; aligns with the spec's Socket.io intent via an idiomatic Go WS library. |
| **AI** | Anthropic **Claude Sonnet** (latest, e.g. `claude-sonnet-4-6`) | Brief, NL query, generated actions. Structured JSON outputs; strict grounding. |
| **Backend hosting** | **Render** via `render.yaml` Blueprint (IaC) | **[deviation]** spec said Vercel+Neon+Railway/Fly. Single Blueprint: Go API + Postgres + Redis + worker. |
| **Frontend hosting** | Render static site (Vite build) | Static SPA + pre-rendered public pages on Render; unified with the backend Blueprint. |
| **Data seeding** | Go `cmd/seed` service | **[deviation]** spec said Node generator. Reseedable from a single seed → reproducible demo state. |
| **CI/CD** | GitHub Actions | lint → test+coverage → SonarQube gate → build → deploy. |
| **Code quality** | **SonarQube** + `golangci-lint` | Quality gate must pass on every push. |
| **Test coverage** | **> 80% on every push** (enforced gate) | Go: `go test -coverprofile`; Frontend: Vitest/Jest + Playwright. |
| **Observability** | OpenTelemetry + Prometheus + Grafana + Sentry | Traces, metrics, logs, error tracking. |
| **Secrets** | Render env groups / GitHub Actions secrets; never in repo | Mandated by both manuals. |

### 2.2 Stack decisions & deviations (ADR seeds)
- **D-001 Go over NestJS.** Owner directive. Hexagonal Go gives strong typing, performance, and a clean
  domain core for the deterministic signal engine. *Risk:* spec's frontend code samples assume TS backend — N/A since boundary is OpenAPI.
- **D-002 Custom JWT over Auth.js.** Full control and no vendor in the critical path. *Risk:* we own the full
  auth security surface (rotation, lockout, reset, MFA) — covered in **E2** + **E9**. Use vetted libs
  (`golang-jwt`, `argon2id`), never hand-roll crypto primitives.
- **D-003 WebSocket via `coder/websocket`.** Owner chose WebSocket (the spec's Socket.io intent) over the SSE
  alternative; idiomatic Go WS library with Redis pub/sub fan-out behind it. Traffic is push-dominant.
- **D-006 React + Vite SPA over Next.js.** Owner preference. SEO preserved by pre-rendering/SSG the public
  pages; the private (`noindex`) cockpit stays a client-side SPA. *Risk:* SPA SEO needs deliberate prerender +
  meta handling — owned in E10.
- **D-007 MUI v9 over Tailwind+shadcn+Tremor.** Owner preference. MUI X Charts replaces Tremor for data-viz;
  Fraunces/Outfit/JetBrains Mono typography per owner direction.
- **D-004 Render Blueprint.** Single declarative `render.yaml` for all services; clean PoC→prod path.
  *Constraint:* no Africa region today → record latency budget; **§8.3 spec** Ghana-hosting remains a
  later deployment decision, architected-for now (12-factor, stateless API, externalised state).
- **D-005 Go seed service over Node.** Keep one backend language; seed logic shares domain types.

### 2.3 Open technology questions (resolve before the epic that needs them)
- **OQ-1** ~~Frontend host~~ — RESOLVED: Render static site (Vite build). No Next.js → no edge/ISR consideration.
- **OQ-2** Claude access: direct Anthropic API vs a gateway (fallback/observability). *Default: direct API + cache; decide before E5.*
- **OQ-3** Production data residency: confirm whether Ghana Data Protection Act (Act 843) hosting is required at GA or post-GA. *Affects E9/E13.*
- **OQ-4** ~~Time-series engine on Render~~ — RESOLVED (2026-06-24): **Render-managed Postgres 16 + pgvector**; TimescaleDB dropped (unsupported on Render). Time-series via native Postgres (indexes; range partitioning + materialized views if volume grows).

---

## 3. Global Definition of Done (applies to every story)

A story is `☑` only when **all** of these hold:
- [ ] Code implemented to the company coding standards; no unrelated changes.
- [ ] **Unit + integration tests written; total coverage stays > 80%** (CI-enforced).
- [ ] **SonarQube quality gate passes** (no new bugs/vulnerabilities/code smells above threshold; duplication < 3%).
- [ ] `golangci-lint` / frontend lint + typecheck clean.
- [ ] Security checklist for the change satisfied (input validation, authz, secrets, logging — see §4.1).
- [ ] OpenAPI spec updated and the generated client/server stubs regenerated (for API stories).
- [ ] Observability added where relevant (logs/metrics/traces; no PII in logs).
- [ ] Docs updated (README/ADR/CLAUDE.md/AGENTS.md) when behaviour or workflow changes.
- [ ] PR reviewed and approved; PR + commits reference `GEC-###`.
- [ ] PR merged; **this `agent_plan.md` updated** (status + dashboard).
- [ ] Acceptance criteria all checked and demonstrated.

---

## 4. Cross-cutting baselines (the "things we might not have thought of")

### 4.1 Security baseline (full work in E9)
Target **OWASP ASVS L2**. Every feature inherits: input validation (allow-list), output encoding, parameterised
SQL only, authz checks at the use-case boundary, rate limiting, secure headers + CSP, secrets in env only,
structured audit logging (who/what/when, **no PII or secrets in logs**), encryption in transit (TLS 1.2+) and at
rest, dependency + container scanning in CI, and a documented threat model. Align to **Ghana Data Protection Act,
2012 (Act 843)** even though demo data is synthetic — architect for real-data later.

### 4.2 SEO & web-performance baseline (full work in E10)
Because the app is a **React + Vite SPA**, SEO is delivered by **pre-rendering/SSG the public & marketing pages**
at build time (e.g. `vite-plugin-ssg` / a prerender step) with semantic HTML, accurate per-route metadata
(react-helmet), Open Graph/Twitter cards, JSON-LD structured data, `sitemap.xml` + `robots.txt`, and canonical
URLs. Performance budgets meet **Core Web Vitals** (LCP < 2.5s, INP < 200ms, CLS < 0.1); Lighthouse SEO + a11y
≥ 95; i18n-ready (en-GH). The authenticated cockpit is a client-side SPA and `noindex` (private); the
public/marketing surface is statically rendered and fully optimised.

### 4.3 Accessibility baseline
**WCAG 2.2 AA**: keyboard nav, focus states, ARIA where needed, colour-contrast (status colours must not be the
only signal), reduced-motion support for Framer Motion.

### 4.4 Testing strategy (full work in E12)
Test pyramid: many unit tests (domain/signal engine ~100%), integration tests for adapters (testcontainers
Postgres/Redis), contract tests against the OpenAPI spec, e2e (Playwright) for the demo-critical path, and a
load/latency check on the brief endpoint. Coverage gate **> 80%** blocks merge.

### 4.5 The hero is sacred (from the spec)
The Daily Brief's four qualities — **alive, personal, smart, fast** — are the top acceptance criterion of the
whole project (spec §2). Brief quality and the demo narrative (spec §3.3) gate everything in E6/E7.

### 4.6 Design & UX standards (owner directives — apply everywhere)
- **Loading states:** use **skeleton loaders** for content/data regions (never spinners/circular or text
  "Loading…"). Use **animated dots** for in-button loading (submit/save/approve/send).
- **Pagination:** **always paginate** lists that can grow — network feed, attention feed, tasks, approvals,
  alerts, staff, reports, facility lists. Prefer **cursor-based** pagination on the API; pager or infinite-scroll
  in the UI.
- **Motion:** **layout transitions** (Framer Motion `layout`) across the cockpit; restrained "alive" motion;
  always honour `prefers-reduced-motion`.
- **Marketing/public-site signature animations:** **parallax** scrolling, **circular reveal** for the light/dark
  theme toggle, and **3D reveal** animations on key sections. Must not breach Core Web Vitals budgets (E10).
- **Typography:** **Fraunces** (titles/display), **Outfit** (body), **JetBrains Mono** (statuses/figures).
  Self-hosted, `font-display: swap`, preloaded — no layout shift.
- **Theme:** light + dark, dark-mode-capable; status conveyed by more than colour alone (a11y, §4.3).

---

# Epics & Stories

## E0 — Foundations & Engineering Operations
*Goal: a repo, toolchain, and CI/CD that make every later story testable, secure, and traceable from line one.*

#### ☑ GEC-1 — Monorepo scaffold & hexagonal skeleton · 5 SP · Phase: Development
> **Done 2026-06-24:** repo laid out; Go backend builds; `internal/architecture/arch_test.go` enforces the core/app boundary; backend tests pass at 94.3% coverage.
- User story: As an engineer, I want a clean monorepo with a hexagonal Go layout, so that domain logic stays free of infrastructure.
- Business value: Enforces the architecture the owner mandated; keeps the codebase testable and maintainable.
- Acceptance criteria:
  - [ ] Repo layout created: `backend/` (`cmd/`, `internal/core`, `internal/app`, `internal/ports`, `internal/adapters/{inbound,outbound}`), `frontend/`, `infra/`, `.github/`.
  - [ ] Domain core compiles with **no** imports of chi/pgx/redis/anthropic.
  - [ ] `go.mod`, `Makefile` (`make build/test/lint/run/seed/migrate`) present.
  - [ ] Frontend bootstrapped (React + Vite + TS + MUI v9; React Router; TanStack Query).
- Technical notes: Enforce import boundaries with `go-arch-lint` or a custom `depguard` rule in `golangci-lint`.
- Definition of done: Global DoD.
- Dependencies: none.

#### ☑ GEC-2 — `CLAUDE.md` & `AGENTS.md` · 3 SP · Phase: Development
> **Done 2026-06-24:** both files written at repo root with coding/hexagonal/security/GitHub/AI rules.
- User story: As an AI/dev contributor, I want project rules documented, so that all work follows company standards.
- Business value: Mandated by both manuals; keeps AI + humans aligned and auditable.
- Acceptance criteria:
  - [ ] `CLAUDE.md` and `AGENTS.md` at repo root.
  - [ ] Each documents: coding standards, hexagonal rules, story workflow (this file), GitHub conventions, AI operating procedures, "do not modify unrelated code".
- Technical notes: Reference this `agent_plan.md` as the Jira substitute.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ◐ GEC-3 — CI pipeline: lint + test + coverage gate · 5 SP · Phase: Development
> **In progress:** `.github/workflows/ci.yml` written; coverage-gate logic verified locally (`make backend-cover-gate`). Remaining: run on GitHub + PR coverage reporting.
- User story: As a team, I want every push linted and tested with a coverage gate, so that quality is automatic.
- Business value: Owner requires >80% coverage on every push; prevents regressions.
- Acceptance criteria:
  - [ ] GitHub Actions runs on every push/PR: `golangci-lint`, `go test -race -coverprofile`, frontend lint+typecheck+test.
  - [ ] Build **fails** if combined coverage < 80%.
  - [ ] Coverage reported on the PR.
- Technical notes: Cache Go/npm deps; matrix backend/frontend jobs.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ◐ GEC-4 — SonarQube quality gate · 5 SP · Phase: Development
> **In progress:** `sonar-project.properties` + CI job written. Remaining: connect a SonarQube/SonarCloud project, set `SONAR_TOKEN`/`SONAR_HOST_URL`, enforce gate.
- User story: As a team, I want SonarQube analysis gating merges, so that code health is enforced.
- Business value: Owner-mandated quality bar; catches bugs/vulns/smells early.
- Acceptance criteria:
  - [ ] SonarQube (or SonarCloud) project configured for Go + TS.
  - [ ] Quality gate wired into CI; **failing gate blocks merge**.
  - [ ] Coverage data fed to Sonar from CI.
- Technical notes: `sonar-project.properties`; exclude generated code (OpenAPI/sqlc) from duplication.
- Definition of done: Global DoD.
- Dependencies: GEC-3.

#### ☑ GEC-5 — OpenAPI tooling & codegen pipeline · 5 SP · Phase: Development
> **Done 2026-06-24:** `backend/api/openapi.yaml` (3.0.3) → oapi-codegen strict Chi server (`openapi_gen.go`; router implements the generated interface) + openapi-typescript TS client (`frontend/src/api/`). `make generate` regenerates both; CI `codegen-drift` job fails on staleness; generated `*_gen.go` excluded from the coverage gate.
- User story: As an engineer, I want one OpenAPI spec generating Go stubs and a TS client, so that the contract can't drift.
- Business value: Single source of truth for the API; type-safe frontend.
- Acceptance criteria:
  - [ ] `backend/api/openapi.yaml` (OpenAPI 3.1) skeleton.
  - [ ] `oapi-codegen` generates chi server interfaces + models.
  - [ ] TS client generated for the frontend; `make generate` regenerates both.
  - [ ] CI fails if generated code is stale (drift check).
- Technical notes: Treat the spec as the contract — handlers implement generated interfaces.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ◐ GEC-6 — Config, secrets & 12-factor setup · 3 SP · Phase: Development
> **In progress:** typed env config loader with fail-fast validation + `.env.example` done & tested. Remaining: gitleaks secret-scanning in CI.
- User story: As an operator, I want config via env with validated startup, so that no secrets live in code.
- Business value: Security baseline; mandated by both manuals.
- Acceptance criteria:
  - [ ] Typed config loader (env → struct) with fail-fast validation.
  - [ ] `.env.example` documents every var; real `.env` git-ignored.
  - [ ] No secret literals anywhere; secret-scanning (gitleaks) in CI.
- Technical notes: Distinct config per env (dev/staging/prod); Render env groups later.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ◐ GEC-7 — Structured logging & error model · 3 SP · Phase: Development
> **In progress:** `slog` JSON logging wired in the composition root. Remaining: central typed error → RFC 9457 problem responses + no-PII-in-logs test.
- User story: As an operator, I want consistent structured logs and a typed error model, so that issues are traceable.
- Business value: Foundation for observability and debugging; avoids PII leaks.
- Acceptance criteria:
  - [ ] `slog` (or zerolog) JSON logging with request/trace IDs.
  - [ ] Central error type → HTTP problem responses (RFC 9457).
  - [ ] **No PII/secrets in logs** (verified by a test/lint).
- Technical notes: Inject logger via context; redact known sensitive fields.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ◐ GEC-8 — Local dev environment (docker-compose) · 3 SP · Phase: Development
> **In progress:** `docker-compose.yml` (pgvector/pgvector:pg16, Redis) + `make dev-up`. Remaining: wire API into compose and run seed (needs E1).
- User story: As a dev, I want one command to run Postgres + pgvector + Redis + API locally, so that onboarding is fast.
- Business value: Cuts onboarding to minutes (Eng-Ops onboarding goal).
- Acceptance criteria:
  - [ ] `docker-compose.yml` with Postgres + pgvector image, Redis, API, frontend.
  - [ ] `make dev` boots the full stack; seed runs against it.
- Technical notes: Pin image versions; healthchecks before API starts.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ☑ GEC-9 — Architecture Decision Records (ADRs) · 2 SP · Phase: Development
> **Done 2026-06-24:** `docs/adr/` with MADR template + ADR-0001 capturing decisions D-001..D-007.
- User story: As the team, I want ADRs capturing key decisions, so that the "why" survives.
- Business value: Governance + onboarding; records the §2.2 deviations.
- Acceptance criteria:
  - [ ] `docs/adr/` with MADR template; ADRs 001–005 written from §2.2.
- Technical notes: One ADR per significant decision going forward.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

---

## E1 — Domain Model, Data Layer & Synthetic Network
*Goal: the relational + time-series data model from spec §7, and a reseedable, Ghana-grounded 12-facility network (spec §4, §10, Appendices A & C).*

#### ☑ GEC-10 — Domain entities & value objects · 5 SP · Phase: Solution Design / Development
> **Done 2026-06-25:** value objects (money/severity/payer) + entities (facility expanded with lifecycle/payer-mix/geo, metric, inventory, staff, alert, task, approval, brief, insight, user). Pure domain, ~100% covered.
- User story: As an engineer, I want pure domain types for the cockpit, so that business rules live in the core.
- Business value: Clean hexagonal core; rules testable without infra.
- Acceptance criteria:
  - [ ] Entities: Facility, FacilityMetric, InventoryItem, Staff, Alert, Task, Approval, Brief, Insight, User.
  - [ ] Value objects: Cedis (money), PayerMix, Severity, FacilityStatus, Region.
  - [ ] Invariants enforced in constructors; 100% unit-tested.
- Technical notes: Money in minor units; never float for currency. Mirror spec §7 fields.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ☐ GEC-11 — Postgres schema & migrations · 5 SP · Phase: Development
- User story: As an engineer, I want versioned migrations for the full schema, so that environments are reproducible.
- Business value: Reliable, auditable data layer.
- Acceptance criteria:
  - [ ] Tables per spec §7: facilities, facility_metrics, inventory_items, staff, alerts, tasks, approvals, briefs, insights, users.
  - [ ] Constraints, indexes, FKs; enums for status/severity/role/type.
  - [ ] `migrate up/down` both work cleanly.
- Technical notes: `golang-migrate`; keep migrations forward-only in prod.
- Definition of done: Global DoD.
- Dependencies: GEC-10.

#### ☐ GEC-12 — Time-series metrics on native Postgres · 5 SP · Phase: Development
- User story: As an engineer, I want fast week-over-week metric queries on plain Postgres, so that trends work on Render (no TimescaleDB).
- Business value: Powers KPI trends and the signal engine while staying within Render's managed Postgres.
- Acceptance criteria:
  - [ ] `facility_metrics` indexed on `(facility_id, date)` for efficient WoW / trailing-window queries.
  - [ ] Materialized view(s) for common aggregates, refreshed by the cron worker (GEC-71).
  - [ ] Declarative range partitioning by time documented as the scale-up path (enabled only if volume warrants).
  - [ ] Query timings documented on the seeded network.
- Technical notes: daily/weekly granularity per spec §7; volume is small (12 facilities) so indexes suffice initially.
- Definition of done: Global DoD.
- Dependencies: GEC-11.

#### ☐ GEC-13 — pgvector for NL retrieval · 3 SP · Phase: Development
- User story: As an engineer, I want pgvector enabled with embeddings on facility notes/names, so that NL Ask can fuzzy-match.
- Business value: Enables grounded natural-language query (spec §6.4).
- Acceptance criteria:
  - [ ] `vector` extension + embedding columns + ANN index.
  - [ ] Embedding write path on relevant text fields.
- Technical notes: Choose embedding model; store dimension in config.
- Definition of done: Global DoD.
- Dependencies: GEC-11.

#### ☐ GEC-14 — Repository adapters (ports → Postgres) · 8 SP · Phase: Development
- User story: As an engineer, I want repository adapters implementing domain ports via pgx/sqlc, so that the core stays infra-free.
- Business value: Testable persistence; swappable storage.
- Acceptance criteria:
  - [ ] One repository per aggregate implementing its port interface.
  - [ ] sqlc-generated queries; parameterised SQL only (no string concat).
  - [ ] Integration tests via testcontainers Postgres.
- Technical notes: Transactions via a UnitOfWork port; map DB errors to domain errors.
- Definition of done: Global DoD.
- Dependencies: GEC-11, GEC-10.

#### ☐ GEC-15 — Synthetic network generator (`cmd/seed`) · 8 SP · Phase: Development
- User story: As the team, I want a reseedable generator for the 12-facility Ghana network, so that the demo is reproducible.
- Business value: The demo lands every time; data feels "unmistakably Ghanaian" (spec §4/§10).
- Acceptance criteria:
  - [ ] Generates all 12 facilities from Appendix A (beds, region/town, payer mix, patients/mo, revenue).
  - [ ] Time-series with texture: weekday/weekend patterns, rainy-season malaria peaks driving volume + stock burn.
  - [ ] Grounded specifics only: cedis, NHIS, MoMo, RDT kits, Ghanaian names/roles. **No placeholder data.**
  - [ ] **Single seed reproduces the identical network** on demand.
- Technical notes: Deterministic PRNG seeded from config; idempotent re-seed.
- Definition of done: Global DoD + data review for "feels Ghanaian".
- Dependencies: GEC-11, GEC-12.

#### ☐ GEC-16 — Planted demo stories (Appendix C) · 5 SP · Phase: Development
- User story: As the team, I want the Appendix C narratives baked into the seed, so that the brief always surfaces the same compelling story.
- Business value: Guarantees the hero moment in every demo.
- Acceptance criteria:
  - [ ] Tafo claims breakdown (rev −22%, ~GH₵78k unbilled, claims recorded-not-submitted).
  - [ ] Asokwa RDT stock-out (~5 days left, 7-day lead time).
  - [ ] Adansi star week (+14% OPD, clean claims). Kasoa NHIS denial spike. Tamale attrition/licence expiry. Cape Coast idle theatre. Sunyani ramping. Nima footfall/wait. 3 approvals (GH₵85k ultrasound, new MO Kasoa, generator Nima).
- Technical notes: Encode as deterministic deltas so the signal engine flags them naturally — not hard-coded brief text.
- Definition of done: Global DoD.
- Dependencies: GEC-15.

#### ☐ GEC-17 — Reference data & licences/staff roles · 3 SP · Phase: Development
- User story: As the team, I want Ghanaian staff roles and licence-expiry data modelled, so that staff signals work.
- Business value: Realism + drives staff-risk signals (spec §6.3).
- Acceptance criteria:
  - [ ] Roles: Medical Officer, Physician Assistant, Nurse, Midwife, Lab Technician, Pharmacist, Records Officer, NHIS Claims Officer.
  - [ ] Licence numbers + expiry; attrition-risk flags seeded (esp. Tamale).
- Technical notes: Expiry windows feed E4 staff signals.
- Definition of done: Global DoD.
- Dependencies: GEC-15.

---

## E2 — Authentication & Authorization
*Goal: real, production-grade custom JWT auth in Go with RBAC for executive vs facility manager (spec §7 users, §8.4 "real auth, fake data").*

#### ☐ GEC-18 — Password & credential security · 5 SP · Phase: Development
- User story: As a user, I want my credentials stored securely, so that my account is safe.
- Business value: Core security; protects the platform.
- Acceptance criteria:
  - [ ] Passwords hashed with **argon2id** (tuned params); never logged.
  - [ ] Strength policy + breached-password check (optional k-anon HIBP).
  - [ ] Constant-time comparisons; generic auth-failure messages.
- Technical notes: Use vetted libs; no custom crypto.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-14.

#### ☐ GEC-19 — JWT issuance & verification · 5 SP · Phase: Development
- User story: As the system, I want signed access tokens with claims, so that requests can be authenticated statelessly.
- Business value: Stateless, scalable auth.
- Acceptance criteria:
  - [ ] Short-lived access JWT (e.g. 15 min), asymmetric signing (EdDSA/RS256).
  - [ ] Claims: sub, role, facility_id (managers), exp/iat/jti.
  - [ ] Auth middleware (chi) validates signature/exp/claims; injects principal into context.
  - [ ] Key rotation supported (kid).
- Technical notes: `golang-jwt`; keys from secret store; reject `alg:none`.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-18.

#### ☐ GEC-20 — Refresh tokens with rotation · 5 SP · Phase: Development
- User story: As a user, I want to stay signed in safely, so that I'm not logged out constantly but a stolen token is contained.
- Business value: Security + UX balance.
- Acceptance criteria:
  - [ ] Refresh tokens stored hashed (Redis/DB) with rotation + reuse-detection (revoke family on reuse).
  - [ ] Logout revokes; device/session list optional.
- Technical notes: httpOnly+Secure+SameSite cookies for refresh; never in localStorage.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-19.

#### ☐ GEC-21 — RBAC & authorization at use-case boundary · 5 SP · Phase: Development
- User story: As the system, I want role/facility scoping enforced in application services, so that managers see only their facility.
- Business value: Prevents data exposure; correct multi-role behaviour.
- Acceptance criteria:
  - [ ] Roles: `executive` (network-wide), `facility_manager` (own facility only).
  - [ ] Authz enforced in the application layer (not just handlers); facility-scoping on every query.
  - [ ] Tests prove a manager cannot access another facility (IDOR-proof).
- Technical notes: Policy as a domain/app concern; deny-by-default.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-19.

#### ☐ GEC-22 — Auth endpoints (login/refresh/logout/me) · 3 SP · Phase: Development
- User story: As a user, I want login/refresh/logout/me endpoints, so that I can use the cockpit.
- Business value: Usable auth surface.
- Acceptance criteria:
  - [ ] `POST /auth/login`, `/auth/refresh`, `/auth/logout`, `GET /auth/me` in OpenAPI.
  - [ ] Rate-limited; brute-force lockout (E9 ties in).
- Technical notes: Implement generated interfaces (GEC-5).
- Definition of done: Global DoD.
- Dependencies: GEC-20, GEC-5.

#### ☐ GEC-23 — Optional TOTP MFA · 5 SP · Phase: Development
- User story: As an executive, I want optional 2FA, so that my high-value account is harder to compromise.
- Business value: Executive accounts are high-value targets.
- Acceptance criteria:
  - [ ] TOTP enrol/verify/disable; recovery codes (hashed).
  - [ ] Enforced when enabled; clear recovery flow.
- Technical notes: `pquerna/otp`; rate-limit verification.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-22.

#### ☐ GEC-24 — Frontend auth integration · 3 SP · Phase: Development
- User story: As a user, I want sign-in to "just work" in the Next.js app, so that I reach the cockpit smoothly.
- Business value: Smooth entry to the demo/product.
- Acceptance criteria:
  - [ ] Login UI; tokens via httpOnly cookies; protected routes redirect.
  - [ ] Silent refresh; role-aware navigation (executive vs manager).
- Technical notes: React Router protected-route guards; tokens in httpOnly cookies, never exposed to JS.
- Definition of done: Global DoD.
- Dependencies: GEC-22, GEC-39 (frontend shell).

---

## E3 — Core Domain APIs (REST + OpenAPI)
*Goal: the REST surface the cockpit needs — facilities, metrics, inventory, staff, alerts, tasks, approvals, briefs/insights, users (spec §5).*

#### ☐ GEC-25 — Facilities API · 5 SP · Phase: Development
- User story: As the cockpit, I want facility list/detail endpoints, so that I can render the network and drill-downs.
- Business value: Powers Network view + Facility detail (spec §5.2/§5.3).
- Acceptance criteria:
  - [ ] `GET /facilities` (list + status colour + headline numbers), `GET /facilities/{id}` (detail).
  - [ ] Sort/filter by status, region, revenue, attention.
  - [ ] **Cursor-based pagination** on the list endpoint (canonical pattern for all list APIs, §4.6).
  - [ ] Facility-scoped for managers (E2).
- Technical notes: Use cases in `internal/app`; handlers thin.
- Definition of done: Global DoD.
- Dependencies: GEC-14, GEC-21, GEC-5.

#### ☐ GEC-26 — Metrics & KPI API · 5 SP · Phase: Development
- User story: As the cockpit, I want network + per-facility KPIs with trends, so that executive KPIs and drill-through work.
- Business value: Spec §5.4 executive KPIs; kills "dashboards side by side".
- Acceptance criteria:
  - [ ] Headline metrics: network revenue, patients seen, occupancy, NHIS outstanding, unbilled, payer mix, per-facility margin (Appendix B defs).
  - [ ] WoW movement; drill-through to per-facility contributors; facility ranking/comparison.
- Technical notes: Backed by native Postgres indexes/materialized views (GEC-12).
- Definition of done: Global DoD.
- Dependencies: GEC-12, GEC-25.

#### ☐ GEC-27 — Inventory API · 3 SP · Phase: Development
- User story: As the cockpit, I want inventory data, so that stock-out projections and facility detail render.
- Business value: Feeds stock-out signal (Asokwa story).
- Acceptance criteria:
  - [ ] `GET /facilities/{id}/inventory`; fields: stock_level, daily_burn, reorder_point, lead_time_days, unit_cost.
- Technical notes: Read model for signal engine.
- Definition of done: Global DoD.
- Dependencies: GEC-14.

#### ☐ GEC-28 — Staff API · 3 SP · Phase: Development
- User story: As the cockpit, I want staff data, so that staff snapshots and licence-expiry warnings show.
- Business value: Spec §5.3; feeds staff signals.
- Acceptance criteria:
  - [ ] `GET /facilities/{id}/staff`; headcount by role, licence expiry, attrition risk.
- Technical notes: Drives Tamale attrition story.
- Definition of done: Global DoD.
- Dependencies: GEC-14, GEC-17.

#### ☐ GEC-29 — Alerts & Attention Feed API · 5 SP · Phase: Development
- User story: As the cockpit, I want a prioritised, dismissible attention feed, so that exceptions surface and resolve.
- Business value: Spec §5.5 attention feed.
- Acceptance criteria:
  - [ ] `GET /alerts` (ranked, **cursor-paginated** per §4.6), `PATCH /alerts/{id}` (dismiss/resolve/act).
  - [ ] Resolved items drop off; new ones surface.
- Technical notes: Alerts produced by signal engine (E4).
- Definition of done: Global DoD.
- Dependencies: GEC-14.

#### ☐ GEC-30 — Tasks / "My Day" API · 5 SP · Phase: Development
- User story: As Sammy, I want a personal task system tied to facilities and brief items, so that I can run my day.
- Business value: Spec §5.7 "My Day".
- Acceptance criteria:
  - [ ] CRUD tasks (title, detail, facility_id nullable, priority, status, due_date, assigned_to, source).
  - [ ] Task lists **paginated** (§4.6).
  - [ ] "Turn this into a task" from a brief item/alert (source = brief/alert).
- Technical notes: Source linkage for traceability.
- Definition of done: Global DoD.
- Dependencies: GEC-14.

#### ☐ GEC-31 — Approvals & decision routing API · 5 SP · Phase: Development
- User story: As Sammy, I want an approval queue I can act on from my phone, so that governance flows to one place.
- Business value: Spec §5.8.
- Acceptance criteria:
  - [ ] `GET /approvals`, `POST /approvals/{id}/decision` (approve/decline/ask) with decision logged.
  - [ ] Types: capital/hire/reorder; carries context to decide.
  - [ ] Seeds the 3 Appendix-C approvals.
- Technical notes: Immutable decision log (audit).
- Definition of done: Global DoD + audit logging.
- Dependencies: GEC-14.

#### ☐ GEC-32 — Delegation & follow-through API · 3 SP · Phase: Development
- User story: As Sammy, I want to assign actions and see completion, so that nothing falls through.
- Business value: Spec §5.9.
- Acceptance criteria:
  - [ ] Assign action to a facility manager; completion status; stalled-action follow-ups surface.
- Technical notes: Reuses tasks + alerts.
- Definition of done: Global DoD.
- Dependencies: GEC-30, GEC-29.

#### ☐ GEC-33 — Users & personalisation API · 3 SP · Phase: Development
- User story: As Sammy, I want the cockpit to learn what I watch, so that it prioritises what I care about.
- Business value: Spec §5.12 personalisation (simulated learning in PoC).
- Acceptance criteria:
  - [ ] `GET/PATCH /me/preferences` (watched metrics, thresholds).
  - [ ] Preferences influence brief/feed prioritisation.
- Technical notes: JSON preferences on users (spec §7).
- Definition of done: Global DoD.
- Dependencies: GEC-22.

---

## E4 — Signal Engine (deterministic)
*Goal: spec §6.3 — numbers, thresholds, deltas, projections computed **in code, never by the model**. Pure domain logic, ~100% test coverage.*

#### ☐ GEC-34 — Signal engine framework · 5 SP · Phase: Development
- User story: As the system, I want a pluggable signal framework, so that detectors emit comparable, ranked signals.
- Business value: Foundation of trustworthy intelligence (spec §6.1).
- Acceptance criteria:
  - [ ] `Signal{type, facility, severity, magnitude, supporting_figures, headline}` produced by detectors implementing a common interface.
  - [ ] Ranking by impact; pure functions, no I/O in core.
- Technical notes: Detectors live in `internal/core/signal`; fed by read models.
- Definition of done: Global DoD + ~100% unit coverage.
- Dependencies: GEC-10.

#### ☐ GEC-35 — Trend & delta detection · 5 SP · Phase: Development
- User story: As the system, I want WoW/trailing-window movement detection, so that revenue/volume/occupancy swings are flagged.
- Business value: Surfaces Tafo's −22% revenue (hero story).
- Acceptance criteria:
  - [ ] Flags swings beyond configurable thresholds on revenue, volume, occupancy, claims.
  - [ ] Each signal carries its own numbers.
- Technical notes: Threshold config externalised.
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34, GEC-12.

#### ☐ GEC-36 — Stock-out projection · 5 SP · Phase: Development
- User story: As the system, I want stock-out projection, so that imminent run-outs inside the reorder window are flagged.
- Business value: Asokwa "approve reorder" story.
- Acceptance criteria:
  - [ ] `days_left = stock_level / daily_burn`; flag when `days_left < lead_time_days`.
  - [ ] Severity scaled by margin to lead time.
- Technical notes: Pure calc from inventory read model.
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34, GEC-27.

#### ☐ GEC-37 — Claims health detection · 5 SP · Phase: Development
- User story: As the system, I want claims-health detection, so that submission gaps and denial spikes surface.
- Business value: The causal insight that makes Sammy believe ("revenue down *because* claims not submitted").
- Acceptance criteria:
  - [ ] Detect submission gaps (revenue recorded, claims not submitted), denial-rate spikes, growing NHIS outstanding.
  - [ ] Connects Tafo revenue drop ↔ unsubmitted claims; Kasoa denial spike.
- Technical notes: This is the diagnostic leap (spec §2.3).
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34, GEC-26.

#### ☐ GEC-38 — Revenue leakage detection · 3 SP · Phase: Development
- User story: As the system, I want unbilled-service detection, so that silent revenue loss is surfaced.
- Business value: Appendix B "unbilled (leakage)".
- Acceptance criteria:
  - [ ] Flags services delivered but unbilled (e.g. Tafo ~GH₵78k).
- Technical notes: From metrics deltas.
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34.

#### ☐ GEC-39 — Staff signals · 3 SP · Phase: Development
- User story: As the system, I want staff-risk detection, so that licence expiries and attrition risk surface.
- Business value: Tamale attrition/licence story.
- Acceptance criteria:
  - [ ] Flags approaching licence expiries, attrition-risk indicators, deployment imbalances.
- Technical notes: From staff read model (GEC-17/28).
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34, GEC-28.

#### ☐ GEC-40 — Network pulse composite · 3 SP · Phase: Development
- User story: As the cockpit, I want a single composite network-health indicator, so that "how is my whole network right now?" is one glance.
- Business value: Spec §5.2 network pulse.
- Acceptance criteria:
  - [ ] Composite score from active signals + KPIs; deterministic and explainable.
- Technical notes: Weighting documented; feeds the Network view.
- Definition of done: Global DoD + ~100% coverage.
- Dependencies: GEC-34, GEC-35..39.

---

## E5 — Intelligence Service (Claude)
*Goal: spec §6 — Claude **narrates** computed signals, **never invents figures**. Daily Brief pipeline, grounded NL query, generated actions, caching, graceful fallback.*

#### ☐ GEC-41 — Anthropic adapter & prompt architecture · 5 SP · Phase: Development
- User story: As the system, I want a Claude client adapter with a stable system prompt, so that intelligence is consistent and swappable.
- Business value: Real AI on the magic touchpoints (spec decisions-locked).
- Acceptance criteria:
  - [ ] Outbound port + Anthropic adapter (Sonnet), retries/timeouts.
  - [ ] Stable system role: "you are Sammy's chief of staff…"; strict "use only supplied figures" instruction.
  - [ ] Model/version in config; per-call structured context.
- Technical notes: Read `claude-api` reference before coding; never log full prompts with data unredacted.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-6.

#### ☐ GEC-42 — Context assembly · 5 SP · Phase: Development
- User story: As the system, I want flagged signals + facts packaged into a structured context object, so that Claude has exactly what it needs.
- Business value: Grounding = trustworthy intelligence.
- Acceptance criteria:
  - [ ] Assemble snapshot: all 12 facilities' latest KPIs, WoW deltas, open approvals, active alerts.
  - [ ] Package signals (E4) + relevant facts into a typed context.
- Technical notes: Deterministic, size-bounded; the pipeline's step 3 (spec §6.2).
- Definition of done: Global DoD.
- Dependencies: GEC-40, GEC-26, GEC-29, GEC-31.

#### ☐ GEC-43 — Structured brief generation · 8 SP · Phase: Development
- User story: As the system, I want Claude to return structured brief JSON, so that the UI renders items with inline actions.
- Business value: The hero output.
- Acceptance criteria:
  - [ ] Output: prose brief + items[] `{severity, facility, headline, explanation, suggested_actions}`.
  - [ ] Top items selected by impact (worst first); causes connected where data supports.
  - [ ] **Schema-validated**; on invalid output, retry/repair; never fabricated figures.
- Technical notes: Anthropic structured/JSON outputs; validate against a strict schema.
- Definition of done: Global DoD + brief-quality review.
- Dependencies: GEC-41, GEC-42.

#### ☐ GEC-44 — Grounded NL query + retrieval · 8 SP · Phase: Development
- User story: As Sammy, I want to ask my business anything in plain English, so that I can interrogate the network.
- Business value: Spec §5.6/§6.4 "Ask" — the close.
- Acceptance criteria:
  - [ ] Interpret question → identify facilities/metrics/timeframe → retrieve via structured queries + pgvector fuzzy match.
  - [ ] Answer **only** from retrieved data; if unsupported, say so (no fabrication).
  - [ ] Examples answered: "which facility needs me this week?", "how is Kasoa's NHIS doing?".
- Technical notes: Guardrail enforced both in prompt and by post-checks.
- Definition of done: Global DoD + guardrail tests.
- Dependencies: GEC-41, GEC-13, GEC-26.

#### ☐ GEC-45 — Generated actions & documents · 5 SP · Phase: Development
- User story: As Sammy, I want the system to produce work, so that the cockpit *does* work, not just shows it.
- Business value: Spec §6.5 — second wow ("Message the manager").
- Acceptance criteria:
  - [ ] Draft a firm, professional WhatsApp-style manager message; board-ready facility summary; draft network report.
  - [ ] Each editable before it leaves his hands.
- Technical notes: Same grounding rules; outputs returned as editable drafts.
- Definition of done: Global DoD.
- Dependencies: GEC-44.

#### ☐ GEC-46 — Caching & morning pre-warm · 5 SP · Phase: Development
- User story: As Sammy, I want the brief instant on open, so that it feels fast (a hero quality).
- Business value: "Fast" is a top-4 brief quality (spec §2 mandate).
- Acceptance criteria:
  - [ ] Daily brief cached (Redis) per day; regenerates on demand or on material change.
  - [ ] Repeated NL queries cached; morning brief pre-warmed via scheduled job.
- Technical notes: Cache keys include data-version; invalidate on material change.
- Definition of done: Global DoD + latency budget met.
- Dependencies: GEC-43, GEC-44.

#### ☐ GEC-47 — Graceful AI fallback · 3 SP · Phase: Development
- User story: As Sammy, I want the cockpit to never show a broken screen, so that a mid-demo API outage doesn't kill it.
- Business value: Spec §6.6 fallback; protects the demo.
- Acceptance criteria:
  - [ ] On API failure, serve last cached brief and degrade gracefully (no error screens).
  - [ ] User-visible "showing last brief" state.
- Technical notes: Circuit-breaker around the adapter.
- Definition of done: Global DoD + chaos test (kill AI mid-flow).
- Dependencies: GEC-46.

#### ☐ GEC-48 — AI cost, latency & abuse controls · 3 SP · Phase: Development
- User story: As an operator, I want AI usage bounded and observed, so that cost/latency stay predictable and abuse is contained.
- Business value: Production cost control + security.
- Acceptance criteria:
  - [ ] Per-user rate limits on Ask; token/cost metrics emitted.
  - [ ] Prompt-injection mitigations on NL input; output never executes actions automatically.
- Technical notes: Treat user input as untrusted; never let model output trigger side-effects without explicit user confirm.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-44, GEC-11 (observability hooks).

---

## E6 — The Daily Brief (hero, end-to-end)
*Goal: wire E1+E4+E5 into the one feature that closes the deal (spec §2, §5.1, §6.2). Quality here outranks everything.*

#### ☐ GEC-49 — Brief pipeline orchestration · 5 SP · Phase: Development
- User story: As the system, I want the full assemble→compute→context→generate→render→cache pipeline, so that the brief generates live each morning.
- Business value: The hero pipeline (spec §6.2).
- Acceptance criteria:
  - [ ] Use case runs all six steps; produces persisted `briefs` row with `source_signal_ids`.
  - [ ] Refreshable on demand.
- Technical notes: Application service coordinating ports only.
- Definition of done: Global DoD.
- Dependencies: GEC-42, GEC-43, GEC-46.

#### ☐ GEC-50 — Brief API endpoint · 3 SP · Phase: Development
- User story: As the cockpit, I want `GET /brief` (+ refresh), so that the Home screen can render it.
- Business value: Frontend contract for the hero.
- Acceptance criteria:
  - [ ] `GET /brief?date=` returns prose + items + inline actions; `POST /brief/refresh`.
  - [ ] Facility-scoped for managers.
- Technical notes: OpenAPI-first.
- Definition of done: Global DoD.
- Dependencies: GEC-49, GEC-5.

#### ☐ GEC-51 — Inline brief actions · 5 SP · Phase: Development
- User story: As Sammy, I want each brief item to act (explain, message manager, approve, open facility), so that I can act without leaving the brief.
- Business value: Spec §2.4/§5.1 "actionable inline".
- Acceptance criteria:
  - [ ] "Why?" digs deeper live; "Message the manager" drafts a sendable message; "Approve" signs; "Open facility" drills in.
- Technical notes: Wire to E3/E5 endpoints.
- Definition of done: Global DoD.
- Dependencies: GEC-50, GEC-45, GEC-31.

#### ☐ GEC-52 — Brief-quality acceptance harness · 8 SP · Phase: QA
- User story: As the team, I want an automated check that the brief meets its four qualities, so that we protect the hero.
- Business value: Brief quality is the project's top acceptance criterion (spec mandate).
- Acceptance criteria:
  - [ ] Golden tests on the seeded network: brief surfaces Tafo first (worst), connects revenue↔claims, names Adansi bright spot, lists the 2+ approvals, reassures on the rest.
  - [ ] **Alive** (changes with data/day), **personal** (greets by name, his idiom), **smart** (causal link present), **fast** (within latency budget) — each asserted.
  - [ ] No fabricated figures (all numbers traceable to DB).
- Technical notes: Combination of deterministic assertions + structural checks on AI output.
- Definition of done: Global DoD + sign-off that "the magic lands".
- Dependencies: GEC-51, GEC-16.

#### ☐ GEC-53 — Demo-narrative e2e (§3.3) · 5 SP · Phase: QA
- User story: As the team, I want the full demo narrative automated end-to-end on a phone viewport, so that it runs flawlessly twice in a row.
- Business value: The demo-readiness gate (spec §11.2).
- Acceptance criteria:
  - [ ] Playwright runs: open cold → read brief → tap "Why?" on Tafo → approve Asokwa reorder + draft message → open Network (12 alive tiles) → drill in → Ask a question → roadmap.
  - [ ] Passes twice consecutively with real Claude, no broken screen.
- Technical notes: Mobile + desktop viewports.
- Definition of done: Global DoD + green twice in a row.
- Dependencies: GEC-52, E7 core screens.

---

## E7 — Cockpit Frontend (React + Vite)
*Goal: spec §5 + §9 — mobile-first, desktop-strong, "command instrument" design, the "alive" feel. PWA.*

#### ☐ GEC-54 — Design system & "command instrument" language · 8 SP · Phase: Development
- User story: As a user, I want a premium, calm, fast UI, so that it feels like the seat of someone who runs an empire.
- Business value: Spec §9.1 design mandate; carries the owner's design directives (§4.6).
- Acceptance criteria:
  - [ ] **MUI v9 theme**: tokens (spacing, radius, elevation), status palette (critical/watch/good), light + dark modes.
  - [ ] **Typography** wired into the theme: Fraunces (titles), Outfit (body), JetBrains Mono (statuses); self-hosted + preloaded, no layout shift.
  - [ ] **MUI X Charts** themed to match (KPI tiles/trends).
  - [ ] **Skeleton** loaders for content + **animated-dots** button-loading component; status never colour-only (a11y).
  - [ ] Framer Motion presets for **layout transitions** + restrained "alive" motion (reduced-motion aware).
- Technical notes: Single ThemeProvider; wrap MUI primitives for consistency; loading components are shared.
- Definition of done: Global DoD + design review.
- Dependencies: GEC-1.

#### ☐ GEC-55 — App shell, routing & PWA · 5 SP · Phase: Development
- User story: As a user, I want an installable app shell with bottom nav (mobile) and multi-pane (desktop), so that it feels like a real app.
- Business value: Spec §9.3 mobile-first / desktop-strong; PWA.
- Acceptance criteria:
  - [ ] React Router routes; bottom nav (mobile), multi-pane layout (desktop), thumb-reachable actions.
  - [ ] PWA via `vite-plugin-pwa` (manifest + service worker); installable; offline shell.
  - [ ] **Layout transitions** between routes (Framer Motion).
- Technical notes: TanStack Query provider + MUI ThemeProvider at the root.
- Definition of done: Global DoD.
- Dependencies: GEC-54.

#### ☐ GEC-56 — Home / The Brief screen (hero) · 8 SP · Phase: Development
- User story: As Sammy, I want the brief at the top the moment I open the app, so that it briefs me before I ask.
- Business value: The hero screen (spec §5.1/§9.2).
- Acceptance criteria:
  - [ ] Renders prose + items (worst first, severity dots), inline actions, attention feed, approvals waiting.
  - [ ] "Subtle motion as the brief composes and numbers settle"; fast first paint.
- Technical notes: Pre-warmed cache → instant; skeleton → composed transition.
- Definition of done: Global DoD + matches GEC-52.
- Dependencies: GEC-50, GEC-55.

#### ☐ GEC-57 — Network single-pane view · 5 SP · Phase: Development
- User story: As Sammy, I want all 12 facilities as living tiles with a network pulse, so that I command the whole empire on one screen.
- Business value: Spec §5.2.
- Acceptance criteria:
  - [ ] 12 tiles (name, region, status colour, 1–2 headline numbers); network pulse at top.
  - [ ] Sort/filter by status/region/revenue/attention; problems float to top.
- Technical notes: Live updates via SSE (E8).
- Definition of done: Global DoD.
- Dependencies: GEC-25, GEC-40, GEC-55.

#### ☐ GEC-58 — Facility detail (drill-down) · 5 SP · Phase: Development
- User story: As Sammy, I want one facility in depth one tap away, so that I can investigate.
- Business value: Spec §5.3.
- Acceptance criteria:
  - [ ] KPI trends (WoW), facility AI notes/alerts, staff snapshot (licence warnings), quick actions (message manager, create task, open approval, generate summary).
- Technical notes: Reuse KPI/charts components.
- Definition of done: Global DoD.
- Dependencies: GEC-26, GEC-28, GEC-29.

#### ☐ GEC-59 — Executive KPIs screen · 5 SP · Phase: Development
- User story: As Sammy, I want portfolio-wide KPIs with ranking and drill-through, so that I think like an owner.
- Business value: Spec §5.4.
- Acceptance criteria:
  - [ ] Headline metrics + facility ranking/comparison + drill-through to contributors.
- Technical notes: Tremor charts; Appendix B definitions surfaced as tooltips.
- Definition of done: Global DoD.
- Dependencies: GEC-26.

#### ☐ GEC-60 — Ask screen (NL query + generated docs) · 8 SP · Phase: Development
- User story: As Sammy, I want a plain-English query box with generated-document output, so that I interrogate and command in words.
- Business value: Spec §5.6 — the close.
- Acceptance criteria:
  - [ ] Single input; grounded answers; generated drafts shown editable; "data can't support that" handled gracefully.
- Technical notes: Streaming response with "thinking" motion.
- Definition of done: Global DoD.
- Dependencies: GEC-44, GEC-45, GEC-55.

#### ☐ GEC-61 — My Day screen · 5 SP · Phase: Development
- User story: As Sammy, I want a clean personal task board tied to facilities, so that I run my day.
- Business value: Spec §5.7.
- Acceptance criteria:
  - [ ] Tasks with priority/due/status; "turn this into a task" from brief/alert; fast board.
- Technical notes: Optimistic updates via TanStack Query.
- Definition of done: Global DoD.
- Dependencies: GEC-30.

#### ☐ GEC-62 — Approvals screen · 3 SP · Phase: Development
- User story: As Sammy, I want a decision queue I act on from my phone, so that governance is one place.
- Business value: Spec §5.8.
- Acceptance criteria:
  - [ ] Queue with context; approve/decline/ask; decision logged + reflected immediately.
- Technical notes: Surfaces the 3 Appendix-C approvals.
- Definition of done: Global DoD.
- Dependencies: GEC-31.

#### ☐ GEC-63 — Reports screen (generate & export) · 5 SP · Phase: Development
- User story: As Sammy, I want one-tap network/investor/board reports from live data, so that reporting isn't hand-assembled.
- Business value: Spec §5.10.
- Acceptance criteria:
  - [ ] Generate + export (PDF) network report; per-investor/per-facility cuts.
- Technical notes: Server-side render → PDF; grounded in DB.
- Definition of done: Global DoD.
- Dependencies: GEC-45, GEC-26.

#### ☐ GEC-64 — Delegation & follow-through UI · 3 SP · Phase: Development
- User story: As Sammy, I want to assign actions and see completion/stalls, so that nothing falls through.
- Business value: Spec §5.9.
- Acceptance criteria:
  - [ ] Assign to manager; status; stalled follow-ups surface in the feed.
- Technical notes: Builds on My Day + Alerts.
- Definition of done: Global DoD.
- Dependencies: GEC-32, GEC-61.

#### ☐ GEC-65 — Personalisation & settings UI · 3 SP · Phase: Development
- User story: As Sammy, I want to tune which metrics/facilities are watched, so that the cockpit learns what I care about.
- Business value: Spec §5.12.
- Acceptance criteria:
  - [ ] Tunable priorities/thresholds; affects brief/feed ordering.
- Technical notes: Writes to preferences (GEC-33).
- Definition of done: Global DoD.
- Dependencies: GEC-33.

#### ☐ GEC-66 — The "alive" details & motion polish · 5 SP · Phase: Polish
- User story: As Sammy, I want subtle live motion, so that the cockpit feels like it's always awake and thinking.
- Business value: Spec §9.4 — protects the magic.
- Acceptance criteria:
  - [ ] Brief composes, numbers settle, tiles/pulse shift on live updates; honours `prefers-reduced-motion`.
- Technical notes: Framer Motion; performance-budget aware.
- Definition of done: Global DoD + design sign-off.
- Dependencies: GEC-56, GEC-57, GEC-67.

#### ☐ GEC-118 — Public / marketing site & signature animations · 8 SP · Phase: Development
- User story: As a prospect, I want a striking public site, so that the product feels premium before I even sign in.
- Business value: Brand + conversion; it is the SEO surface (E10) and the first impression.
- Acceptance criteria:
  - [ ] Public landing/marketing pages, separate from the authed cockpit, pre-rendered for SEO (GEC-83).
  - [ ] **Parallax** scrolling, **3D reveal** animations on key sections, **circular reveal** on the light/dark theme toggle.
  - [ ] Fraunces/Outfit/JetBrains Mono typography; fully responsive; honours `prefers-reduced-motion`.
  - [ ] Animations stay within Core Web Vitals budgets (GEC-86).
- Technical notes: Framer Motion; lazy-load heavy 3D/animation assets; theme toggle uses the View Transitions API where supported, with a clip-path circular-reveal fallback.
- Definition of done: Global DoD + design review.
- Dependencies: GEC-54.

---

## E8 — Realtime, Notifications & Alerts
*Goal: the "always awake" channel (spec §8.2) — push live updates; quiet-by-default notifications.*

#### ☐ GEC-67 — WebSocket live update channel · 5 SP · Phase: Development
- User story: As the cockpit, I want a live channel, so that new alerts and brief updates appear without refresh.
- Business value: Reinforces "always awake" (spec §8.2/§9.4).
- Acceptance criteria:
  - [ ] Authenticated WebSocket endpoint (`coder/websocket`), JWT-authed; auto-reconnect with backoff; facility-scoped events.
  - [ ] Tiles/pulse/feed update live; heartbeat/ping-pong keepalive.
- Technical notes: Chi handler upgrades to WS; Redis pub/sub fan-out across instances.
- Definition of done: Global DoD.
- Dependencies: GEC-29, GEC-21.

#### ☐ GEC-68 — Material-change brief invalidation · 3 SP · Phase: Development
- User story: As the system, I want the brief to regenerate on material change, so that it stays current within the day.
- Business value: Keeps the hero "alive".
- Acceptance criteria:
  - [ ] Material changes invalidate cached brief and push an update.
- Technical notes: Define "material change" thresholds.
- Definition of done: Global DoD.
- Dependencies: GEC-46, GEC-67.

#### ☐ GEC-69 — Push notifications (critical only) · 5 SP · Phase: Development
- User story: As Sammy, I want push notifications only for things that genuinely need me, so that notifications stay trusted.
- Business value: Spec §5.11 "quiet by default".
- Acceptance criteria:
  - [ ] Web Push for stock-out imminent, sharp revenue drop, approval waiting.
  - [ ] Quiet-by-default; per-user thresholds.
- Technical notes: Web Push API + service worker (PWA).
- Definition of done: Global DoD.
- Dependencies: GEC-55, GEC-29.

#### ☐ GEC-70 — Alert lifecycle & dedup · 3 SP · Phase: Development
- User story: As the system, I want alerts deduped and lifecycle-managed, so that the feed stays trustworthy.
- Business value: Avoids alert fatigue.
- Acceptance criteria:
  - [ ] Dedup repeated signals; resolve/dismiss/escalate; no duplicate pushes.
- Technical notes: Idempotency keys on alerts.
- Definition of done: Global DoD.
- Dependencies: GEC-29.

#### ☐ GEC-71 — Scheduled jobs (pre-warm, follow-ups) · 5 SP · Phase: Development
- User story: As an operator, I want scheduled jobs, so that the morning brief is pre-warmed and stalled follow-ups surface.
- Business value: Fast brief + delegation follow-through.
- Acceptance criteria:
  - [ ] Cron-style worker: morning pre-warm, stalled-action sweep, licence-expiry sweep.
- Technical notes: Render cron/worker service; idempotent.
- Definition of done: Global DoD.
- Dependencies: GEC-46, GEC-32, GEC-39.

---

## E9 — Security Hardening & Compliance
*Goal: take "real auth, fake data" (spec §8.4) to OWASP ASVS L2 and architect for Ghana DPA (Act 843). The "things we might not have thought of".*

#### ☐ GEC-72 — Threat model & security requirements · 3 SP · Phase: Solution Design
- User story: As the team, I want a documented threat model, so that we design controls deliberately.
- Business value: Proactive security; informs all later stories.
- Acceptance criteria:
  - [ ] STRIDE threat model; trust boundaries; abuse cases; mapped mitigations.
- Technical notes: Living doc in `docs/security/`.
- Definition of done: Global DoD.
- Dependencies: GEC-9.

#### ☐ GEC-73 — Input validation & output encoding · 5 SP · Phase: Development
- User story: As the system, I want strict validation everywhere, so that injection/XSS are prevented.
- Business value: Closes the biggest vuln classes.
- Acceptance criteria:
  - [ ] Allow-list validation on all inputs (server-side); parameterised SQL only; safe templating/encoding.
  - [ ] Negative tests for SQLi/XSS payloads.
- Technical notes: Validate at the edge of the app layer.
- Definition of done: Global DoD + security review.
- Dependencies: GEC-25..33.

#### ☐ GEC-74 — Rate limiting & brute-force protection · 3 SP · Phase: Development
- User story: As the system, I want rate limiting and lockout, so that abuse and credential-stuffing are contained.
- Business value: Protects auth + AI cost.
- Acceptance criteria:
  - [ ] Per-IP/user limits on auth + Ask; exponential lockout; 429 with retry-after.
- Technical notes: Redis token bucket; tie to GEC-22/48.
- Definition of done: Global DoD.
- Dependencies: GEC-22, GEC-48.

#### ☐ GEC-75 — Security headers & CSP · 3 SP · Phase: Development
- User story: As the system, I want strict security headers, so that the browser enforces our security posture.
- Business value: Defence-in-depth.
- Acceptance criteria:
  - [ ] HSTS, strict CSP (nonce-based), X-Content-Type-Options, Referrer-Policy, Permissions-Policy, frame-ancestors.
  - [ ] CSP verified not to break the app.
- Technical notes: Static-host (Render) headers + API middleware; CSP nonces for the SPA.
- Definition of done: Global DoD.
- Dependencies: GEC-55.

#### ☐ GEC-76 — CORS & CSRF protection · 2 SP · Phase: Development
- User story: As the system, I want correct CORS and CSRF defences, so that cross-origin abuse is blocked.
- Business value: Prevents session-riding attacks.
- Acceptance criteria:
  - [ ] Strict origin allow-list; SameSite cookies; CSRF tokens for cookie-auth mutations.
- Technical notes: Aligns with GEC-20 cookie model.
- Definition of done: Global DoD.
- Dependencies: GEC-20.

#### ☐ GEC-77 — Audit logging · 3 SP · Phase: Development
- User story: As the business, I want immutable audit logs of sensitive actions, so that decisions are accountable.
- Business value: Governance; approvals are decisions of record.
- Acceptance criteria:
  - [ ] Append-only audit log for auth events, approvals, role changes, exports (who/what/when).
  - [ ] **No PII/secrets** in logs; tamper-evident.
- Technical notes: Separate audit store/table; retention policy.
- Definition of done: Global DoD.
- Dependencies: GEC-7, GEC-31.

#### ☐ GEC-78 — Encryption in transit & at rest · 3 SP · Phase: Development
- User story: As the business, I want data encrypted in transit and at rest, so that data is protected.
- Business value: Baseline + DPA readiness.
- Acceptance criteria:
  - [ ] TLS 1.2+ enforced end-to-end; DB/Redis/backups encrypted at rest; secrets encrypted.
- Technical notes: Render-managed TLS + at-rest; document key management.
- Definition of done: Global DoD.
- Dependencies: GEC-6.

#### ☐ GEC-79 — Dependency, SAST & secret scanning in CI · 3 SP · Phase: Development
- User story: As the team, I want automated security scanning, so that vulns and leaked secrets are caught pre-merge.
- Business value: Shift-left security.
- Acceptance criteria:
  - [ ] `govulncheck`, `npm audit`/`osv-scanner`, gitleaks, SAST (Sonar/CodeQL) in CI; high severities block.
- Technical notes: Triage workflow for findings.
- Definition of done: Global DoD.
- Dependencies: GEC-3, GEC-4.

#### ☐ GEC-80 — Container & image hardening + DAST · 3 SP · Phase: Staging
- User story: As an operator, I want hardened images and a DAST pass, so that the deployed surface is minimal and tested.
- Business value: Runtime security.
- Acceptance criteria:
  - [ ] Distroless/minimal base, non-root user, no shell where avoidable; Trivy scan in CI.
  - [ ] OWASP ZAP baseline DAST against staging.
- Technical notes: Multi-stage Go build.
- Definition of done: Global DoD.
- Dependencies: GEC-99 (staging).

#### ☐ GEC-81 — Ghana Data Protection Act (Act 843) alignment · 3 SP · Phase: Solution Design
- User story: As the business, I want DPA-aligned data handling, so that the move to real data is a deployment decision, not a rebuild.
- Business value: Spec §8.3 production note.
- Acceptance criteria:
  - [ ] Data inventory/classification; lawful-basis & retention notes; data-subject-rights design (access/erasure); residency plan (Ghana hosting path).
  - [ ] Privacy policy + consent surfaces stubbed.
- Technical notes: Synthetic data now, but architecture must support PII controls.
- Definition of done: Global DoD.
- Dependencies: GEC-72.

#### ☐ GEC-82 — Pre-production penetration test · 5 SP · Phase: Staging
- User story: As the business, I want a pen-test before GA, so that real-world weaknesses are found and fixed.
- Business value: Production confidence.
- Acceptance criteria:
  - [ ] Pen-test (internal or external) against staging; findings triaged; criticals/highs fixed; report archived.
- Technical notes: Scope = auth, authz/IDOR, AI input, exports.
- Definition of done: Global DoD + no open criticals.
- Dependencies: GEC-80.

---

## E10 — SEO & Web Performance
*Goal: the public surface is fully optimised; the private cockpit is fast and `noindex`. Meets Core Web Vitals.*

#### ☐ GEC-83 — Pre-render public pages & metadata (SPA SEO) · 5 SP · Phase: Development
- User story: As a visitor, I want fast, crawlable public pages, so that the product is discoverable.
- Business value: SEO requirement, delivered without Next.js.
- Acceptance criteria:
  - [ ] Public/marketing routes **pre-rendered/SSG** at build (`vite-plugin-ssg` or a prerender step) with accurate `<title>`/meta/canonical.
  - [ ] Per-route meta via react-helmet (or equivalent); cockpit routes `noindex`.
- Technical notes: Separate the public bundle/layout from the authed SPA; serve static HTML to crawlers.
- Definition of done: Global DoD.
- Dependencies: GEC-55, GEC-118.

#### ☐ GEC-84 — Structured data (JSON-LD) & Open Graph · 3 SP · Phase: Development
- User story: As a visitor/sharer, I want rich previews and structured data, so that the product looks credible in search/social.
- Business value: Click-through + SEO.
- Acceptance criteria:
  - [ ] Organization/SoftwareApplication JSON-LD; OG + Twitter cards with images.
- Technical notes: Validate with Rich Results test.
- Definition of done: Global DoD.
- Dependencies: GEC-83.

#### ☐ GEC-85 — Sitemap, robots & canonicalization · 2 SP · Phase: Development
- User story: As a crawler, I want a sitemap and robots rules, so that indexing is correct.
- Business value: SEO hygiene.
- Acceptance criteria:
  - [ ] `sitemap.xml` (public only), `robots.txt` (disallow cockpit), canonical tags.
- Technical notes: Generated at build.
- Definition of done: Global DoD.
- Dependencies: GEC-83.

#### ☐ GEC-86 — Core Web Vitals & performance budgets · 5 SP · Phase: Polish
- User story: As a user, I want fast loads and interactions, so that the product feels premium.
- Business value: CWV affects SEO + the "fast" hero quality.
- Acceptance criteria:
  - [ ] LCP < 2.5s, INP < 200ms, CLS < 0.1 on target devices; budgets enforced in CI (Lighthouse CI).
- Technical notes: Code-split, image optimisation, font strategy, route prefetch.
- Definition of done: Global DoD + budgets green.
- Dependencies: GEC-55.

#### ☐ GEC-87 — Image, font & asset optimization · 3 SP · Phase: Polish
- User story: As a user, I want optimised assets, so that pages are light and fast.
- Business value: CWV + cost.
- Acceptance criteria:
  - [ ] Optimized responsive images (AVIF/WebP), lazy-loading; self-hosted optimized fonts (no layout shift).
- Technical notes: `vite-imagetools` for images; preload Fraunces/Outfit/JetBrains Mono; `font-display: swap`.
- Definition of done: Global DoD.
- Dependencies: GEC-86.

#### ☐ GEC-88 — Accessibility (WCAG 2.2 AA) · 5 SP · Phase: QA
- User story: As any user, I want an accessible cockpit, so that it's usable and compliant (and SEO-friendly).
- Business value: Inclusion + SEO + risk.
- Acceptance criteria:
  - [ ] Keyboard nav, focus management, ARIA, contrast; status not colour-only; reduced-motion.
  - [ ] axe + Lighthouse a11y ≥ 95; manual screen-reader pass on hero path.
- Technical notes: Bake a11y checks into Playwright.
- Definition of done: Global DoD.
- Dependencies: GEC-54.

#### ☐ GEC-89 — i18n-readiness (en-GH) · 3 SP · Phase: Development
- User story: As the business, I want the UI i18n-ready with Ghanaian English/locale, so that cedis/dates/number formats are correct.
- Business value: Realism + future locales.
- Acceptance criteria:
  - [ ] Locale framework wired; GH₵ currency, date/number formats; copy externalised.
- Technical notes: react-i18next (or LinguiJS); default en-GH; structure for future locales.
- Definition of done: Global DoD.
- Dependencies: GEC-55.

---

## E11 — Observability & Reliability
*Goal: see, measure, and recover. The "things we might not have thought of" for running it in production.*

#### ☐ GEC-90 — OpenTelemetry tracing · 5 SP · Phase: Development
- User story: As an operator, I want distributed traces, so that I can debug the brief pipeline and slow requests.
- Business value: Fast diagnosis in prod.
- Acceptance criteria:
  - [ ] OTel traces across HTTP → app → DB/Redis → Anthropic; trace IDs in logs.
- Technical notes: Export to a collector/backend (Grafana Tempo/Honeycomb).
- Definition of done: Global DoD.
- Dependencies: GEC-7.

#### ☐ GEC-91 — Metrics & dashboards · 5 SP · Phase: Development
- User story: As an operator, I want Prometheus metrics + Grafana dashboards, so that I see system health at a glance.
- Business value: Operability.
- Acceptance criteria:
  - [ ] RED metrics per endpoint; brief latency, AI cost/tokens, cache hit-rate; Grafana dashboards.
- Technical notes: `/metrics` endpoint (protected).
- Definition of done: Global DoD.
- Dependencies: GEC-90.

#### ☐ GEC-92 — Error tracking (Sentry) · 2 SP · Phase: Development
- User story: As the team, I want client + server error tracking, so that we catch issues fast.
- Business value: Reliability.
- Acceptance criteria:
  - [ ] Sentry on frontend + backend; releases tagged; **PII scrubbed**.
- Technical notes: Source maps uploaded in CI.
- Definition of done: Global DoD.
- Dependencies: GEC-7.

#### ☐ GEC-93 — Health checks & probes · 2 SP · Phase: Development
- User story: As the platform, I want liveness/readiness endpoints, so that deploys and restarts are safe.
- Business value: Zero-downtime deploys.
- Acceptance criteria:
  - [ ] `/healthz` (liveness), `/readyz` (deps); used by Render health checks.
- Technical notes: Readiness checks DB/Redis.
- Definition of done: Global DoD.
- Dependencies: GEC-6.

#### ☐ GEC-94 — SLOs & alerting · 3 SP · Phase: Staging
- User story: As the team, I want SLOs and alerts, so that we know before users do.
- Business value: Proactive reliability.
- Acceptance criteria:
  - [ ] SLOs (availability, brief latency, error rate) defined; alerts wired to a channel; runbook links.
- Technical notes: Alert on burn-rate, not single spikes.
- Definition of done: Global DoD.
- Dependencies: GEC-91.

#### ☐ GEC-95 — Backups & disaster recovery · 5 SP · Phase: Staging
- User story: As the business, I want backups and a tested restore, so that data loss is recoverable.
- Business value: Production must-have.
- Acceptance criteria:
  - [ ] Automated Postgres backups; documented + **tested** restore; RPO/RTO stated.
- Technical notes: Render managed backups + periodic restore drill.
- Definition of done: Global DoD + successful restore test.
- Dependencies: GEC-99.

#### ☐ GEC-96 — Runbooks & incident process · 2 SP · Phase: Hypercare
- User story: As on-call, I want runbooks and an incident process, so that we respond consistently.
- Business value: Hypercare phase (Eng-Ops §12).
- Acceptance criteria:
  - [ ] Runbooks for top failure modes (AI down, DB down, deploy rollback); incident severity + comms template.
- Technical notes: Link from alerts (GEC-94).
- Definition of done: Global DoD.
- Dependencies: GEC-94.

---

## E12 — Quality, Testing & CI Gates
*Goal: the test pyramid + gates that make >80% coverage and SonarQube real, not aspirational.*

#### ☐ GEC-97 — Domain & signal-engine unit tests · 5 SP · Phase: Development
- User story: As the team, I want the domain + signal engine near-fully unit-tested, so that the math is trustworthy.
- Business value: Trust = the product's core value (spec §6.1).
- Acceptance criteria:
  - [ ] ~100% coverage on `internal/core`; table-driven tests; edge cases (zero burn, missing data).
- Technical notes: No I/O in these tests.
- Definition of done: Global DoD.
- Dependencies: GEC-34..40.

#### ☐ GEC-98 — Integration tests (testcontainers) · 5 SP · Phase: Development
- User story: As the team, I want adapter integration tests against real Postgres/Redis, so that persistence is correct.
- Business value: Catches real DB issues.
- Acceptance criteria:
  - [ ] testcontainers Postgres + pgvector + Redis; migrations applied; repo round-trips verified.
- Technical notes: Runs in CI (Docker).
- Definition of done: Global DoD.
- Dependencies: GEC-14.

#### ☐ GEC-99 — API contract tests · 3 SP · Phase: Development
- User story: As the team, I want contract tests against the OpenAPI spec, so that the API never drifts from its contract.
- Business value: Frontend/back-end stay in sync.
- Acceptance criteria:
  - [ ] Requests/responses validated against `openapi.yaml`; CI fails on mismatch.
- Technical notes: schemathesis or generated client assertions.
- Definition of done: Global DoD.
- Dependencies: GEC-5, GEC-25.

#### ☐ GEC-100 — Frontend unit/component tests · 5 SP · Phase: Development
- User story: As the team, I want component tests, so that UI logic is covered.
- Business value: Coverage gate + regression safety.
- Acceptance criteria:
  - [ ] Vitest + Testing Library; key components/hooks covered; coverage counts toward the 80% gate.
- Technical notes: Mock the typed API client.
- Definition of done: Global DoD.
- Dependencies: GEC-54.

#### ☐ GEC-101 — E2E tests (Playwright) · 5 SP · Phase: QA
- User story: As the team, I want e2e coverage of critical journeys, so that the demo path can't silently break.
- Business value: Protects the close.
- Acceptance criteria:
  - [ ] Auth, brief render+actions, network drill, Ask, approval — mobile + desktop.
- Technical notes: Feeds GEC-53.
- Definition of done: Global DoD.
- Dependencies: GEC-56, GEC-57, GEC-60.

#### ☐ GEC-102 — Load & latency test (brief endpoint) · 3 SP · Phase: Staging
- User story: As the team, I want a load test on the brief/Ask paths, so that latency budgets hold under load.
- Business value: "Fast" hero quality under real conditions.
- Acceptance criteria:
  - [ ] k6 test; p95 within budget at expected concurrency; cache effectiveness verified.
- Technical notes: Mock Anthropic for deterministic load runs.
- Definition of done: Global DoD.
- Dependencies: GEC-46.

#### ☐ GEC-103 — Coverage gate enforcement (>80%) · 2 SP · Phase: Development
- User story: As the team, I want the 80% gate enforced and visible, so that quality can't regress.
- Business value: Owner-mandated.
- Acceptance criteria:
  - [ ] Combined backend+frontend coverage gate blocks merge below 80%; trend visible.
- Technical notes: Aggregate coverage reporting.
- Definition of done: Global DoD.
- Dependencies: GEC-3, GEC-100.

#### ☐ GEC-104 — Mutation testing (core) · 3 SP · Phase: QA
- User story: As the team, I want mutation testing on the signal engine, so that tests are meaningful, not just coverage theatre.
- Business value: Real confidence in the math.
- Acceptance criteria:
  - [ ] `go-mutesting` (or equivalent) on `internal/core`; mutation score threshold set; gaps addressed.
- Technical notes: Run nightly, not per-PR (cost).
- Definition of done: Global DoD.
- Dependencies: GEC-97.

---

## E13 — Deployment, Infra & Release
*Goal: render.yaml Blueprint + the multi-stage release path (Eng-Ops §7–§11): staging → UAT → beta → production.*

#### ◐ GEC-105 — render.yaml Blueprint · 5 SP · Phase: Development
> **In progress:** `infra/render.yaml` (API + frontend + Redis + Postgres + secrets group) + backend Dockerfile written. Remaining: deploy from Blueprint and run first migration (`CREATE EXTENSION vector`).
- User story: As an operator, I want all services declared as IaC, so that environments are reproducible.
- Business value: Owner's hosting choice; reproducible infra.
- Acceptance criteria:
  - [ ] `infra/render.yaml` declares: Go API (web), worker/cron, Postgres, Redis, frontend; env groups; health checks.
  - [ ] Spins up a working environment from the Blueprint.
- Technical notes: Stateless API; externalised state for the future Ghana-hosting move (D-004).
- Definition of done: Global DoD.
- Dependencies: GEC-93.

#### ☐ GEC-106 — Dockerfiles (multi-stage) · 3 SP · Phase: Development
- User story: As an operator, I want small, secure images, so that deploys are fast and hardened.
- Business value: Performance + security.
- Acceptance criteria:
  - [ ] Multi-stage Go build → distroless/minimal, non-root; frontend image; both scanned (GEC-80).
- Technical notes: Reproducible builds; pinned bases.
- Definition of done: Global DoD.
- Dependencies: GEC-1.

#### ☐ GEC-107 — Environments: dev/staging/prod · 3 SP · Phase: Development
- User story: As the team, I want isolated environments, so that we can promote changes safely.
- Business value: Safe release path.
- Acceptance criteria:
  - [ ] Three environments via Blueprint; separate secrets/DBs; seed in non-prod only.
- Technical notes: Prod uses no synthetic seed unless explicitly a demo env.
- Definition of done: Global DoD.
- Dependencies: GEC-105.

#### ☐ GEC-108 — CD: build → migrate → deploy · 5 SP · Phase: Development
- User story: As the team, I want automated deploys with migrations, so that releases are one-click and safe.
- Business value: Eng-Ops §10 automation (GitHub → CI/CD → Production).
- Acceptance criteria:
  - [ ] On main merge (after gates): build, run migrations, deploy; rollback on failed health check.
  - [ ] Deploy status reflected back (Jira-substitute note in PR/this file).
- Technical notes: Migrations run before traffic shift.
- Definition of done: Global DoD.
- Dependencies: GEC-105, GEC-3, GEC-4.

#### ☐ GEC-109 — Zero-downtime releases & rollback · 3 SP · Phase: Staging
- User story: As an operator, I want rolling deploys and fast rollback, so that releases don't break the demo/prod.
- Business value: Reliability.
- Acceptance criteria:
  - [ ] Rolling/health-gated deploy; documented one-command rollback; tested.
- Technical notes: Backward-compatible migrations (expand/contract).
- Definition of done: Global DoD.
- Dependencies: GEC-108.

#### ☐ GEC-110 — Feature flags · 3 SP · Phase: Development
- User story: As the team, I want feature flags, so that we can ship dark and control rollout (beta phase).
- Business value: Eng-Ops §9 beta; safe experimentation.
- Acceptance criteria:
  - [ ] Flag system (config or service); flags for in-progress features; documented.
- Technical notes: Keep flag count low; clean up stale flags.
- Definition of done: Global DoD.
- Dependencies: GEC-6.

#### ☐ GEC-111 — Staging smoke + UAT + beta gates · 5 SP · Phase: UAT/Beta
- User story: As the team, I want staging smoke tests and UAT/beta gates, so that releases follow the Eng-Ops SDLC.
- Business value: Eng-Ops §7–§9 approval gates.
- Acceptance criteria:
  - [ ] Post-deploy smoke suite on staging; UAT checklist + sign-off; limited beta rollout with feedback capture.
  - [ ] **Demo-readiness gate (GEC-53) green twice** before any stakeholder showing.
- Technical notes: Smoke = subset of e2e.
- Definition of done: Global DoD + UAT sign-off recorded.
- Dependencies: GEC-101, GEC-53, GEC-110.

---

## E14 — Documentation, Governance & Handover
*Goal: Eng-Ops §11 deliverables + the AI-governance docs both manuals mandate.*

#### ☐ GEC-112 — API documentation · 2 SP · Phase: Development
- User story: As a consumer, I want browsable API docs, so that integration is easy.
- Business value: DX + handover.
- Acceptance criteria:
  - [ ] Swagger UI/Redoc served from the OpenAPI spec; kept current via codegen.
- Technical notes: Auto from `openapi.yaml`.
- Definition of done: Global DoD.
- Dependencies: GEC-5.

#### ☐ GEC-113 — Architecture & data-model docs · 3 SP · Phase: Development
- User story: As the team, I want architecture + ERD docs, so that the system is understandable.
- Business value: Onboarding + governance.
- Acceptance criteria:
  - [ ] Hexagonal diagram, context map, ERD, signal-engine + brief-pipeline diagrams in `docs/`.
- Technical notes: Diagrams as code (Mermaid) where possible.
- Definition of done: Global DoD.
- Dependencies: GEC-11, GEC-49.

#### ☐ GEC-114 — Onboarding guide · 2 SP · Phase: Development
- User story: As a new hire, I want a guide, so that I can complete the onboarding project (manuals' onboarding flow).
- Business value: Manuals' New Employee Onboarding.
- Acceptance criteria:
  - [ ] README quickstart; links to manuals, CLAUDE.md, AGENTS.md, this plan, workflow.
- Technical notes: One command to a running local stack (GEC-8).
- Definition of done: Global DoD.
- Dependencies: GEC-8, GEC-2.

#### ☐ GEC-115 — User guide & training material · 3 SP · Phase: Sign-off
- User story: As the client, I want a user guide and training material, so that the team can adopt the cockpit.
- Business value: Eng-Ops §11 deliverables.
- Acceptance criteria:
  - [ ] User guide (mobile + desktop), short training material, FAQ.
- Technical notes: Screenshots from the real app.
- Definition of done: Global DoD.
- Dependencies: E7 complete.

#### ☐ GEC-116 — Release notes automation · 2 SP · Phase: Production
- User story: As the team, I want generated release notes, so that stakeholders are informed automatically.
- Business value: Eng-Ops §10 automation; spec §5.10 spirit.
- Acceptance criteria:
  - [ ] Release notes generated from merged PRs/changelog on deploy; stakeholder notification.
- Technical notes: Conventional-commit driven.
- Definition of done: Global DoD.
- Dependencies: GEC-108.

#### ☐ GEC-117 — Acceptance & handover package · 2 SP · Phase: Sign-off
- User story: As the client, I want a handover package, so that we can sign off the project.
- Business value: Eng-Ops §11 official sign-off.
- Acceptance criteria:
  - [ ] Final demo; acceptance checklist mapped to spec §12 DoD; handover docs + acceptance certificate.
- Technical notes: Validate against spec §12 "Definition of done".
- Definition of done: Global DoD + client sign-off.
- Dependencies: GEC-111, GEC-115.

---

## 5. Spec "Definition of Done" alignment (spec §12)
The PoC's own DoD maps to these stories — all must be `☑` for the PoC to be "done":
- Daily Brief live from real computed signals, prioritised, Tafo revenue↔claims connected, working inline actions → **GEC-43, GEC-49, GEC-51, GEC-52**.
- Full demo narrative end-to-end on a phone with real Claude, no broken screen, reproducible → **GEC-53, GEC-47**.
- All 12 facilities present, alive, drillable; pulse reflects real state → **GEC-57, GEC-40, GEC-67**.
- Ask answers grounded questions + generates manager message + facility/board summary → **GEC-44, GEC-45, GEC-60**.
- My Day, Approvals, Reports functional → **GEC-61, GEC-62, GEC-63, GEC-30, GEC-31**.
- Genuinely responsive (mobile-first + first-class desktop) → **GEC-55, GEC-66, GEC-86**.
- Fully Ghana-grounded data, no placeholders → **GEC-15, GEC-16, GEC-17, GEC-89**.

---

## 6. Suggested build sequence (maps to spec §11 phases)
1. **Phase 0 — Foundations:** E0, then E1 (schema + seed + planted stories). *Milestone: seeded 12-facility network.*
2. **Phase 1 — The hero:** E4 (signal engine) → E5 (intelligence) → E6 (Daily Brief end-to-end). *Milestone: the brief generates live and the magic lands (GEC-52).*
3. **Phase 2 — Command surfaces:** E2 (auth) + E3 (APIs) + E7 (Network, Facility detail, KPIs, Attention feed) + E8 (realtime). *Milestone: drill the whole network.*
4. **Phase 3 — Act & ask:** Ask, My Day, Approvals, Delegation, Reports (E5/E7). *Milestone: the cockpit does work.*
5. **Phase 4 — Polish & demo + production-readiness:** E9 (security), E10 (SEO/perf/a11y), E11 (observability), E12 (test gates), E13 (deploy/release), E14 (docs/handover). *Milestone: demo-ready twice in a row (GEC-53) and production-ready.*

> Security (E9), SEO (E10), observability (E11), and test gates (E12) are **continuous**, not a final phase — they're listed as epics for tracking but their stories are pulled forward into each feature's DoD.

---

## 7. Risk register
| ID | Risk | Impact | Mitigation | Owner story |
|---|---|---|---|---|
| R1 | Brief quality slips (not alive/personal/smart/fast) | Kills the deal | Treat as top acceptance criterion; automated quality harness | GEC-52 |
| R2 | AI fabricates figures | Destroys trust | Deterministic compute + strict grounding + guardrail tests | GEC-43, GEC-44, GEC-48 |
| R3 | AI API outage mid-demo | Broken demo | Cache + graceful fallback + chaos test | GEC-47 |
| R4 | Custom-auth security gaps | Breach / data exposure | Vetted libs, threat model, pen-test, ASVS L2 | E2, GEC-72, GEC-82 |
| R5 | No Africa hosting region (latency) | Sluggish feel in Ghana | Region selection (jnb/af-south path), perf budgets, architect for Ghana hosting | D-004, GEC-86 |
| R6 | Coverage gate stalls delivery | Slower velocity | Write tests with features (TDD-leaning), mutation testing for real confidence | GEC-103, GEC-104 |
| R7 | Synthetic data feels generic | Demo doesn't land | Ghana-grounded generator + planted stories + data review | GEC-15, GEC-16 |
| R8 | Scope creep into clinical AI / interoperability | Regulatory + timeline risk | Explicitly out of scope (spec §3.2); roadmap only | — |

---

## 8. Changelog
| Date | Change | By |
|---|---|---|
| 2026-06-24 | Initial plan created from PoC spec + both manuals; stack locked (Go/Chi/REST+OpenAPI/Custom JWT/Render Blueprint); 15 epics, 117 stories. | Claude |
| 2026-06-24 | Frontend/data stack selected per owner: React+Vite SPA (not Next.js), MUI v9 (not Tailwind/Tremor), MUI X Charts, TanStack Query, React Hook Form+Zod; pgx+sqlc; Postgres+Timescale+pgvector; **WebSocket** realtime; Claude Sonnet. Added §4.6 design/UX standards (skeleton + animated-dot loaders, pagination, layout transitions, marketing parallax/3D/circular-reveal, Fraunces/Outfit/JetBrains Mono) and story GEC-118. 118 stories. | Claude |
| 2026-06-24 | **Phase 0 kickoff scaffolded** (latest versions of everything): git repo; Go hexagonal backend (Chi v5.3, build + 94.3% test coverage, arch boundary test); CLAUDE.md/AGENTS.md; CI workflow + SonarQube + golangci v2 config; render.yaml + Dockerfile + docker-compose; React 19 + Vite + MUI v9 + MUI X Charts 9 frontend (100% test coverage). GEC-1/2/9 done; GEC-3/4/6/7/8/105 in progress. | Claude |
| 2026-06-24 | **DB finalised for Render: dropped TimescaleDB** → Render-managed Postgres 16 + pgvector with native time-series (indexes; partitioning/materialized views if needed). Updated stack table, GEC-12, GEC-98, docker-compose (`pgvector/pgvector:pg16`), render.yaml, ADR-0001, and docs. OQ-4 resolved. | Claude |
| 2026-06-24 | **GEC-5 done** — OpenAPI 3.0.3 contract + codegen: oapi-codegen strict Chi server (router implements the generated interface) + openapi-typescript client (openapi-fetch). `make generate` + CI drift job; generated code excluded from the coverage gate (backend 96.8%, frontend 100%). | Claude |
| 2026-06-25 | **GEC-10 done** — domain model per spec §7: value objects money(Cedis)/severity/payer + entities facility(expanded), metric, inventory, staff, alert, task, approval, brief, insight, user. Pure, ~100% covered (backend gate 99.2%). | Claude |
