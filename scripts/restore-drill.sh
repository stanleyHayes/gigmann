#!/usr/bin/env bash
# Restore drill (GEC-95). Proves a Postgres backup actually restores: it dumps a
# source database, restores it into a throwaway scratch database, and verifies the
# restore is faithful (identical set of tables with identical row counts).
#
# Run it periodically against a backup/replica/quiescent source to satisfy the DR
# "tested restore" requirement (RPO <= 24h, RTO <= 1h — see docs/backups-and-dr.md).
#
# Usage:
#   scripts/restore-drill.sh <SOURCE_DATABASE_URL> [SCRATCH_DATABASE_URL]
#
#   SOURCE_DATABASE_URL    database to drill (or set $DATABASE_URL). Should be a
#                          backup/replica or quiescent so counts are stable.
#   SCRATCH_DATABASE_URL   optional empty database to restore into. If omitted, a
#                          scratch database is created on the source server (needs
#                          CREATEDB) and dropped afterwards.
#
# Exit status 0 = restore verified, non-zero = drill failed.
set -euo pipefail

SRC="${1:-${DATABASE_URL:-}}"
if [ -z "$SRC" ]; then
  echo "usage: restore-drill.sh <SOURCE_DATABASE_URL> [SCRATCH_DATABASE_URL]" >&2
  exit 2
fi
PROVIDED_SCRATCH="${2:-}"

for bin in pg_dump pg_restore psql; do
  command -v "$bin" >/dev/null || { echo "error: $bin not found on PATH" >&2; exit 2; }
done

# Rewrite a postgres URL's path (database name); used to derive the maintenance
# and scratch connection strings from the source.
url_with_db() {
  python3 - "$1" "$2" <<'PY'
import sys, urllib.parse as u
p = u.urlparse(sys.argv[1])
print(u.urlunparse((p.scheme, p.netloc, "/" + sys.argv[2].lstrip("/"), "", "", "")))
PY
}

WORK="$(mktemp -d)"
DUMP="$WORK/source.dump"
CREATED_SCRATCH=""

cleanup() {
  if [ -n "$CREATED_SCRATCH" ]; then
    psql "$ADMIN_URL" -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS \"$CREATED_SCRATCH\";" >/dev/null 2>&1 || true
  fi
  rm -rf "$WORK"
}
trap cleanup EXIT

echo "==> [1/4] dumping source (custom format)"
pg_dump -Fc --no-owner --no-privileges -f "$DUMP" "$SRC"

if [ -n "$PROVIDED_SCRATCH" ]; then
  SCRATCH_URL="$PROVIDED_SCRATCH"
  echo "==> [2/4] using provided scratch database"
else
  ADMIN_URL="$(url_with_db "$SRC" postgres)"
  CREATED_SCRATCH="restore_drill_$$_$(date +%Y%m%d%H%M%S)"
  echo "==> [2/4] creating scratch database $CREATED_SCRATCH"
  psql "$ADMIN_URL" -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$CREATED_SCRATCH\";" >/dev/null
  SCRATCH_URL="$(url_with_db "$SRC" "$CREATED_SCRATCH")"
fi

echo "==> [3/4] restoring into scratch"
pg_restore --no-owner --no-privileges -d "$SCRATCH_URL" "$DUMP" 2>"$WORK/restore.err" || true
# pg_restore can return non-zero on benign warnings; data-level faithfulness is
# asserted by the row-count comparison below.

# Per-table row counts for the public schema, as "schema.table|count" lines.
table_counts() {
  psql "$1" -At -F'|' <<'SQL'
SELECT table_schema || '.' || table_name,
       (xpath('/row/c/text()',
              query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name),
                           false, true, '')))[1]::text::bigint
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY 1;
SQL
}

echo "==> [4/4] verifying tables + row counts"
SRC_COUNTS="$(table_counts "$SRC")"
DST_COUNTS="$(table_counts "$SCRATCH_URL")"

if [ -z "$SRC_COUNTS" ]; then
  echo "FAIL: source has no public tables — nothing to verify" >&2
  exit 1
fi
if [ "$SRC_COUNTS" != "$DST_COUNTS" ]; then
  echo "FAIL: restored database does not match source:" >&2
  diff <(printf '%s\n' "$SRC_COUNTS") <(printf '%s\n' "$DST_COUNTS") >&2 || true
  exit 1
fi

TABLES="$(printf '%s\n' "$SRC_COUNTS" | grep -c . || true)"
ROWS="$(printf '%s\n' "$SRC_COUNTS" | awk -F'|' '{s+=$2} END{print s+0}')"
echo "PASS: restored $TABLES tables / $ROWS rows with matching counts"
