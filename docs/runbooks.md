# Runbooks & Incident Process (GEC-96)

On-call reference for the Gigmann Executive Cockpit. The system is designed to
**degrade, not fail**: most dependency outages drop to a working fallback.

## Severity
- **SEV1** — cockpit unreachable / auth broken for all users.
- **SEV2** — a core surface degraded (brief stale, search down).
- **SEV3** — cosmetic / single-feature.

## Health signals
- Liveness: `GET /healthz` (and `/readyz`). Metrics: `GET /metrics` (Prometheus).
- Logs: structured JSON (slog) — filter by `request_id`. Audit events: `action=auth.*` / `approval.*`.

## Scenarios

### AI (Claude) down or slow
- **Symptom:** brief/Ask slow or erroring; logs show Anthropic errors.
- **Behaviour:** the brief is **cached** and served from cache; if `ANTHROPIC_API_KEY`
  is unset/invalid the app uses the **deterministic local narrator** — figures are
  identical, prose is templated. No action needed for correctness.
- **Action:** verify key/quota; if persistent, leave the local narrator (set no key)
  and note degraded prose. No data loss.

### Embeddings (Voyage) down
- **Symptom:** facility search returns poor/empty matches.
- **Behaviour:** without `VOYAGE_API_KEY` the **local lexical embedder** is used
  (offline). Re-embedding happens at first-run seed.
- **Action:** confirm key; to re-embed after a switch, truncate `facility_embeddings`
  and restart (idempotent seed) or run a re-embed task.

### Database (Postgres) down
- **Symptom:** startup fails to connect, or queries error.
- **Behaviour:** if `DATABASE_URL` is unset the app runs fully **in-memory** from the
  synthetic network (no durability). With it set, the app needs Postgres up.
- **Action:** check Render Postgres status; migrations are idempotent (advisory-lock
  protected) and re-applied on boot. As a stopgap, unset `DATABASE_URL` to run the
  in-memory demo.

### Deploy gone bad → rollback
- Render: redeploy the previous successful image from the dashboard (or `render
  deploys rollback`). The API is stateless; migrations are forward-only — a rollback
  of code is safe as long as no destructive migration shipped (none have).

### Auth broken (no logins)
- Check `JWT_SECRET` is present (required outside dev). Rotating it invalidates all
  live sessions (expected). Verify rate-limiter isn't tripping (per-IP window).

## Escalation
SEV1/2 → page the owner; capture `request_id`s + `/metrics` snapshot; open an incident
note (timeline, impact, fix, follow-ups).
