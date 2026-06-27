# Backups & Disaster Recovery (GEC-95)

## What holds state
Only Postgres (facilities, metrics, approvals/tasks, users/credentials, refresh
tokens, embeddings). The API is **stateless** and Redis (when enabled) holds only
ephemeral cache/rate-limit state. So DR == Postgres backup/restore + redeploy.

## Backups
- **Render managed Postgres** takes automated daily backups with point-in-time
  recovery on paid plans — enable it on the database and set the retention window.
- For an extra off-platform copy, schedule `pg_dump` (e.g. a Render cron) to object
  storage.

## Targets
- **RPO** ≤ 24h (daily backup) — tighten with PITR if the data warrants.
- **RTO** ≤ 1h (provision DB from backup + redeploy the stateless API).

## Restore procedure
1. Provision/restore Postgres from the chosen backup (Render dashboard → Restore, or
   `pg_restore` of the dump).
2. Point `DATABASE_URL` at the restored instance.
3. Redeploy the API — the migration runner reconciles schema on boot
   (advisory-locked, idempotent); first-run seed is skipped when data exists.
4. Verify `/readyz` (DB ping) and the acceptance smoke matrix
   (`docs/acceptance-handover.md`).

## Needs a human / live infra
A **tested restore drill** (restore into a scratch DB and verify) must be run on the
real Render project once provisioned; it can't be exercised from the repo alone.
