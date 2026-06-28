# Internal Security Assessment (GEC-82)

> Scope: the Gigmann Executive Cockpit ("Ahenfie") backend (Go, hexagonal), React
> frontend, the deterministic signal engine, and the Postgres adapter — as of
> 2026-06-28. This is the **internal** assessment that precedes the external
> pre-production penetration test (which still requires an engaged firm + a
> deployed staging URL — see "Residual / external" below).

## Methodology
1. **Threat model** — STRIDE over the auth/API/AI trust boundaries
   ([threat-model.md](threat-model.md)).
2. **Automated controls in CI** (every push / scheduled):
   - SAST: **CodeQL** (`.github/workflows/codeql.yml`)
   - Dependencies: **govulncheck** (Go vuln DB) + `npm audit --omit=dev`
   - Secrets: **gitleaks** (`.gitleaks.toml` allowlists only generated code + test fixtures)
   - Container: **Trivy** image scan (HIGH/CRITICAL, fixable) on the distroless image
   - DAST: **OWASP ZAP** baseline (`.github/workflows/dast.yml`) against the running API
3. **Adversarial multi-agent code audits** — six independent review passes
   (backend app/adapters, frontend, signal-engine math, Postgres/SQL, the auth
   core, and AI grounding / prompt-injection / AI-side-effects), each finding
   cross-verified by independent skeptics that default to "refuted" when uncertain.
   False-positive rate was high by design (≈40–55% of raw findings rejected),
   so what remained was actioned.

## Findings & remediations
All confirmed findings were fixed and shipped; CI is green on every commit.

| Area | Finding | Severity | Fix (commit) |
|---|---|---|---|
| **Authorization (IDOR)** | Facility managers could read/mutate **any** facility (GetFacility, UpdateAlertStatus, CreateDraft, CreateTask, **ListTasks, UpdateTaskStatus**) and see all alerts/approvals/tasks — the domain had `CanAccessFacility`/`ErrForbidden` but handlers never enforced it. Now enforced across **every** facility-bearing read/write (an independent fix-review caught the two task endpoints missed in the first pass). | High | `4397e20` + fix-review |
| Dependencies | `golang.org/x/crypto` v0.51.0 — 9 HIGH SSH CVEs. | High | `6a72a9d` |
| Input handling | Malformed Ask/Draft bodies returned **500** with code `bad_request` (now 400). | Med | `fd14cbf` |
| Input handling | Preferences keys had no length cap (bloat vector). | Med | `4397e20` |
| Resource safety | Unbounded rate-limiter map (memory growth per distinct client). | Med | `4397e20` |
| Resource safety | Web Push response body leaked on the error path. | Med | `4397e20` |
| Frontend | 401-retry replayed mutations with an **empty body** (consumed Request). | Med | `b6cabf5` |
| Frontend | Silent error/loading states masked failures (approvals, search, tasks, push). | Med | `b6cabf5` |
| Data fidelity | KPI money truncated pesewas; revenue-leakage headline truncated. | Med | `b6cabf5`, `4397e20` |
| Correctness | Alert-feed keyset cursor reset when an alert changed status mid-paging. | High | `4397e20` |
| Correctness | Brief-ranking magnitude not normalised — large-currency signals outranked higher-severity ratio signals within a tier. | High | `918c5d0` |
| Correctness | Licence-expiry boundary off-by-one (exact-day expiry not flagged). | Med | `918c5d0` |
| Availability | Materialized-view refresh held an exclusive lock, blocking chart reads during the cron rebuild (now `CONCURRENTLY`, populate-guarded). | High | `55a1bbd` |
| **Auth — rate limiting** | `X-Forwarded-For` was trusted unconditionally, so a spoofed header bypassed the per-IP login/refresh rate limit. Now XFF is honoured only behind a trusted proxy (`TRUST_PROXY`), using the proxy-observed (rightmost) entry. | High | _this commit_ |
| **Auth — MFA** | TOTP codes had no single-use enforcement → a captured code was replayable within its ±1-step window. Now each time-step counter is consumed at most once per user. | High | _this commit_ |
| Auth — MFA | The MFA-confirm endpoint (6-digit TOTP) was not rate-limited; added to the brute-force-sensitive path set. | Med | _this commit_ |
| Auth — config | Defence in depth: the well-known dev placeholder secret is now rejected outside development (the empty-secret guard already existed). | Med | _this commit_ |
| Availability — ws | WebSocket connections were uncapped; added a concurrent-connection limit. | Med | _this commit_ |
| **AI grounding** | The narrated Brief / Ask answer could reference a facility id the model invented (no code-level guard). Now brief items and answer citations are validated against the engine's facility set — invented references are dropped (the AI never invents a facility). | Med | _this commit_ |
| **AI input** | The Draft `instruction` was unbounded (the Ask `question` was capped at 1000 runes; Draft was not) — a large prompt-injection/cost payload. Now capped to 1000 runes; both bounds declared in the OpenAPI (`maxLength`). | High | _this commit_ |
| CI integrity | Non-deterministic tool/action pins; secret-scan false positive on generated code. | — | `73dfbb0`, `71ea8f9`, `94580e5` |

### Explicitly assessed and **not** changed (with rationale)
- **Access tokens in `localStorage`** — accepted risk: a standard SPA pattern given
  short-lived access tokens + single-use rotating refresh; httpOnly cookies would be
  a backend-auth/CSRF rearchitecture. Revisit if XSS surface grows.
- **argon2id `t=1` with 64 MiB memory** — already exceeds OWASP's `m=46 MiB, t=1`
  baseline; not weakened.
- **KPI `deltaPct` at `previous = 0`** — returns `0` (undefined ratio); fabricating
  "+100%" would invent a figure, against the product's core principle. The
  `direction` field conveys the trend.
- **WebSocket token in `?token=`** — the browser WebSocket API cannot set an
  `Authorization` header, so the short-lived access token is passed as a query
  param. Verified that our `requestLogger` logs only the path (never the query), so
  the token does not leak via our logs; the residual is the inherent browser-WS
  limitation (proxy/history). Future hardening: mint short-lived WS-specific tokens.
- **Numeric-claim grounding in free text** — figures are computed by the engine and
  supplied to the model, which is instructed never to invent one; brief items and
  citations are now validated against the facility set, but individual numbers in
  the narrated prose are not re-extracted and checked against the figures map. Full
  numeric validation is an inherent model-trust boundary (a research-grade problem);
  it is mitigated by the constrained tool schema, the supplied `SupportingFigures`,
  the deterministic local-narrator fallback, and the brief-quality fidelity tests.

## Controls in place
- AuthN: HS256 JWT (short-lived) + single-use **rotating** refresh tokens (hashed at rest); argon2id password hashing.
- AuthZ: enforced in the app layer; managers scoped to their facility (verified, no IDOR).
- Transport/headers: HSTS (prod), CSP, CORP, strict CORS allow-list.
- Input: parameterised SQL (sqlc), app-boundary allow-list validation, per-principal rate limiting.
- AI safety: the model never invents a figure (deterministic engine) and never triggers a side-effect without explicit user confirmation.

## Residual / external (the formal pen test)
A formal **third-party penetration test** against a **deployed staging URL** is still
required to close GEC-82 — covering live auth flows, session/token handling under
real network conditions, business-logic abuse, and infra. The automated DAST (ZAP)
and this internal assessment are the inputs/scope for that engagement.
