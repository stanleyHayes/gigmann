# Multi-Agent Adversarial Audit — 2026-06-29

> Follow-up sweep to the internal assessment ([assessment.md](assessment.md)), run while
> several agents were landing new features (MFA recovery codes/disable/QR, Reports PDF
> export). Six review dimensions ran in parallel — auth/MFA, Postgres/SQL, frontend,
> API/HTTP security, the deterministic signal engine, and test/CI determinism — each
> finding cross-verified by an independent skeptic that defaults to "refuted". 40 raw
> findings → 31 confirmed → triaged below. The skeptics over-flagged (≈ a third of the
> "confirmed" set were accepted-risk re-raises, already-fixed, or false), so each item
> was re-judged by hand.

## Fixed (shipped, CI green)

| Finding | Sev | Fix | Commit |
|---|---|---|---|
| Facility NL search (`SearchFacilities`) was authenticated but **not executive-only** — a facility manager could search the whole-network roster (same disclosure surface as `ListFacilities`). | High | Added to `executiveOperations` (fail-closed; matches the assessment's Option A). Test added. | `4772a12` |
| Facility search **query was unbounded** before being sent to the embedder (a paid/remote call) — cost/abuse vector. | High | Bounded to 256 runes at the app boundary (`Resolve`). Test added. | `4772a12` |
| Signal engine: `licence_expiry` used a flat `Magnitude: 1`, so expiries never ranked by urgency and dominated their tier. | High | Normalised within the alert window (sooner/expired ranks higher), matching the other detectors. Test added. | `4772a12` |
| API responses had **no `Cache-Control`** (ZAP [10049] storable/cacheable). | Med | `Cache-Control: no-store` in `securityHeaders`. | `6b95520` |
| MFA migration `recovery_code_hashes NOT NULL`, but the credential upsert sent a Go `nil` slice → SQL `NULL` (broke credential upsert/login on Postgres). | High | Nil-guard `[]string{}` in `credentialParams`. | `858ca28` |
| `gosec` G118 broke the lint gate on intentional detached goroutines. | — | Excluded G118 in `.golangci.yml` (documented). | `858ca28` |

## Open — flagged to lane owners (not edited here, to avoid colliding with active work)

| Finding | Sev | Owner | Location | Recommended fix |
|---|---|---|---|---|
| **Recovery-code double-spend (TOCTOU).** `consumeRecoveryCode` is a non-atomic read-modify-write (load account → remove hash → save); concurrent logins can spend one recovery code twice → MFA bypass. | **High** | MFA (Codex) | `backend/internal/app/auth_service.go` (`consumeRecoveryCode`) | Consume atomically in one SQL statement: `UPDATE credentials SET recovery_code_hashes = array_remove(recovery_code_hashes, $2) WHERE user_id = $1 AND $2 = ANY(recovery_code_hashes) RETURNING ...` — success only if a row changed. (A `sync.Mutex` keyed per user closes the single-instance race but not multi-instance.) |
| **PDF pagination math wrong.** `position = heightLeft - imgHeight + margin` mis-places content on pages 2+ (computes a large negative Y); any multi-page export is malformed. | **High** | Reports (Kimi) | `frontend/src/screens/exportBrief.ts:~161` | Track a running offset; place each new page at `position = margin` and advance the source slice. Add a test with a canvas taller than one page (current mock canvas has height 0, so the loop never runs). |
| **Silent clipboard failure for recovery codes.** `void navigator.clipboard?.writeText(...)` swallows rejection — if copy fails (permissions/insecure context) the user gets no feedback for one-time codes. | High | Settings (Codex/Kimi) | `frontend/src/screens/SettingsScreen.tsx:~46` | `try/await/catch` with success + error feedback (Alert/Snackbar). |
| **No PDF error/loading state.** `onDownloadPdf` has no try/catch and `void`s the promise; html2canvas/jsPDF failure is silent. | Med | Reports (Kimi) | `frontend/src/screens/ReportsScreen.tsx:~36` | try/catch/finally with loading + error UI (the codebase's established pattern). |
| **QR error/loading semantics.** Failure renders plain `Typography` (not `Alert`); loading `Box` has `aria-label` but no `role`. | Med | Settings (Kimi) | `frontend/src/components/MfaQrCode.tsx:~44` | Use `Alert severity="error"`; add `role="status"` to the loading box. |
| **Missing Postgres integration test** for recovery-code consume + `DisableMFA` nil round-trip (the nil→NULL bug above was only caught by an unrelated integration test; the consume path is memory-repo-only). | Med | MFA (Codex) | `backend/internal/adapters/outbound/postgres/persistence_integration_test.go` | Add a round-trip test: enroll → persist hashes → consume one → disable (nil) → reload. |
| Possibly time-boundary-flaky MFA enrollment test (`time.Now()` ±1 step). | Med | MFA (Codex) | `backend/internal/app/auth_service_test.go:~247` | Pin the step like `TestMFAEnrollAndStepUp` (`time.Unix((now/30+1)*30, 0)`). |

## Assessed and **not** changed (rationale)

- **In-memory TOTP single-use + rate-limiter "not distributed".** Accepted risk: the
  deployment is single-instance (Render); a distributed (Redis) store is a deliberate
  non-goal until horizontal scaling is real. Already documented.
- **`no-store` on auth endpoints / `recovery_code_hashes` nil→NULL.** Already fixed
  (`6b95520`, `858ca28`); the audit read a pre-fix tree.
- **CORS "capitalization bypass".** False: the browser sets `Origin`; an exact-match
  allow-list of known origins is correct and cannot be bypassed by casing the header.
- **Recovery-code loop "timing leak".** Marginal: codes are 80-bit, ≤10 per user, and the
  argon2id verify per iteration dominates timing; position-in-list leakage is not useful
  without a valid code. Not worth a constant-time-loop rewrite.
- **CI tool pins (govulncheck `@latest`, Trivy pinned+retry, `go-version: stable`).** The
  Trivy pinned-image + retry loop is deliberate (Docker Hub flakiness). Go-version
  consistency is a minor follow-up, not a correctness issue.
