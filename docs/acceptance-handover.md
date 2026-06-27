# Acceptance & Handover Package (GEC-117)

Summary of what is delivered, how to verify it, and what remains. Pair with
`docs/architecture.md`, `docs/onboarding.md`, and `agent_plan.md` (per-story status).

## Acceptance test matrix (manual smoke)
| Area | Steps | Expected |
|---|---|---|
| Auth | Login (exec + manager); bad password; MFA when enrolled | 200 + token; generic failure; step-up |
| Daily Brief | Open Today | Narrated prose + worst-first items with figures; copy/download work |
| Network | Open Network; open a facility | 12 facilities; drill-down (inventory/staff/alerts) |
| Quick search | "Kasoa polyclinic" in 🔍 | Ranked matches → facility detail |
| KPIs | Open KPIs | Revenue/patients/occupancy/denial trends |
| Ask | Ask a question | Grounded answer + citations; no invented figures |
| My Day | Move a task todo→done | Persists |
| Approvals | Approve one; re-decide | 200; re-decide → 409; manager → 403 |
| Settings | Toggle theme; reload | Choice remembered |

## Automated verification
- Backend: `make backend-cover-gate` (>80%, currently ~88%), `make lint`,
  `make backend-integration` (testcontainers; runs in CI). Persistence verticals
  were additionally runtime-verified against native Postgres 18 + pgvector 0.8.3.
- Frontend: `npm run lint && npm run typecheck && npm run test:coverage && npm run build`.
- CI (`.github/workflows/ci.yml`): backend gate, frontend, SonarQube, codegen-drift,
  integration, secret-scan, govulncheck.

## Delivered (epics)
Domain + signal engine (~100% covered), deterministic grounded narration, full REST
API (brief/metrics/facilities + detail/auth+MFA/approvals/tasks/ask/facility-search/
preferences/alerts), Postgres + pgvector persistence, React SPA (all core screens),
PWA, theming, a11y baseline, CI, Render Blueprint, ADRs + docs.

## Environment reference
Required: `JWT_SECRET` (outside dev). Optional: `DATABASE_URL`, `REDIS_URL`,
`ANTHROPIC_API_KEY`/`ANTHROPIC_MODEL`, `VOYAGE_API_KEY`/`VOYAGE_MODEL`,
`CORS_ALLOWED_ORIGINS`, `DEMO_PASSWORD`. See `backend/.env.example` + `infra/render.yaml`.

## Known gaps / not in PoC scope
Realtime (WebSocket GEC-67) + push (GEC-69); public marketing site + SEO
(GEC-83/84/85/118); Sentry/SLOs/backups (GEC-92/94/95) and a penetration test
(GEC-82) — these need live infra, third-party accounts, or a human process and are
tracked in `agent_plan.md`. The pasted Anthropic key from development **must be
rotated** before handover.
