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

## Restore drill (automated + tested)
`scripts/restore-drill.sh` proves a backup actually restores: it dumps a source
database, restores it into a throwaway scratch database, and verifies the restore
is faithful (identical tables with identical row counts). Exit 0 = verified.

```sh
scripts/restore-drill.sh <SOURCE_DATABASE_URL> [SCRATCH_DATABASE_URL]
```

- `SOURCE_DATABASE_URL` — a backup/replica or quiescent DB (defaults to `$DATABASE_URL`).
- `SCRATCH_DATABASE_URL` — optional empty DB to restore into; if omitted, a scratch
  database is created on the source server (needs `CREATEDB`) and dropped afterwards.

This drill is **verified** against a real Postgres 18 instance seeded by the API
(14 tables / 207 rows restored with matching counts; the empty-source and
count-mismatch paths both fail as expected).

### Periodic production drill
Schedule the script (e.g. a monthly Render cron or CI `workflow_dispatch`) against
the **live backup/replica** once the Render project is provisioned, pointing
`SOURCE_DATABASE_URL` at a restored snapshot. That exercises the real managed
backup end-to-end; the procedure + verification are already codified and tested here.
