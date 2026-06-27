# Threat Model — Gigmann Executive Cockpit (GEC-72)

Scope: the cockpit API (Go) + SPA (React) deployed on Render. Methodology: STRIDE
over the main trust boundaries. The PoC runs on **synthetic data** (no real patient
PII), which materially lowers data-sensitivity risk; controls below are designed to
hold when real facility data is introduced.

## Trust boundaries
1. Browser ↔ API (public internet, TLS-terminated at Render).
2. API ↔ Claude / Voyage (outbound HTTPS to third parties).
3. API ↔ Postgres/Redis (private network, TLS).
4. CI/CD ↔ repo + Render (secrets in env groups / GitHub secrets).

## Assets
Executive credentials & sessions; facility operational/financial figures;
AI prompts/outputs; signing & API secrets.

## STRIDE

| Category | Threat | Control (where) |
|---|---|---|
| **Spoofing** | Stolen/forged token | HS256 JWT verified server-side; short access TTL (15m); single-use **rotating** refresh tokens (hashes only); TOTP MFA step-up (`core/mfa`). Refresh re-reads the live account so revoked privileges don't persist. |
| **Tampering** | Mutated request / SQL injection | Parameterised SQL only (sqlc / `::vector` casts); input validated/sanitised at the app boundary; OpenAPI strict server rejects malformed bodies. |
| **Repudiation** | "I didn't approve that" | Audit log (`AuditLogger`) records auth + decision events (actor/action/target/outcome) via structured slog. Approvals are explicit, user-initiated, never AI-triggered. |
| **Information disclosure** | PII/secret leakage; IDOR | Never log PII or secrets (slog discipline); secrets only via env groups; managers scoped to their facility at the use-case boundary; generic auth errors prevent account enumeration; only token *hashes* stored. |
| **Denial of service** | Auth brute force / AI cost abuse | Per-IP fixed-window rate limiting on auth paths; per-principal rate limiting + bounded answers on Ask; request timeouts; the brief is cached (model not on the hot path). |
| **Elevation of privilege** | Manager acting as executive | RBAC enforced in `app` (e.g. `ApprovalService.Decide` is executive-only → 403); the `arch_test` keeps authz out of adapters. |

## AI-specific risks
- **Hallucinated figures** → structurally prevented: numbers come only from the Go
  signal engine; Claude uses constrained `emit_*` tools; a grounding guardrail
  rejects unsupplied figures/entities; AI output is read-only and never triggers a
  side-effect without explicit user confirmation (CLAUDE.md §7).
- **Prompt injection via NL input** → the model cannot call tools that mutate state;
  Ask answers are grounded in the computed context only.

## Residual risks / follow-ups
- HSTS + a strict CSP (GEC-75), CodeQL/SBOM/Trivy in CI (GEC-79/80), a real
  penetration test against staging (GEC-82), and DPA alignment (GEC-81) are tracked
  separately. WebSocket and push surfaces (GEC-67/69) are not yet built.
