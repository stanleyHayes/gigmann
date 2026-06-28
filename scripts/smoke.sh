#!/usr/bin/env bash
# Post-deploy smoke (GEC-111). Exercises the critical path of a running deployment
# over HTTP — health, readiness, login, a grounded Daily Brief, and metrics. It is
# lightweight (no browser); the full browser journey is the Playwright e2e
# (GEC-53/101). Exit 0 = the deployment is serving its core path.
#
# Usage: scripts/smoke.sh <BASE_URL> [EMAIL] [PASSWORD]
#   BASE_URL   e.g. https://gigmann-staging.onrender.com (or set $BASE_URL)
#   EMAIL      login email    (default ceo@gigmann.health / $SMOKE_EMAIL)
#   PASSWORD   login password (default demo-pass-1234     / $SMOKE_PASSWORD)
set -euo pipefail

BASE="${1:-${BASE_URL:-}}"
if [ -z "$BASE" ]; then
  echo "usage: smoke.sh <BASE_URL> [EMAIL] [PASSWORD]" >&2
  exit 2
fi
BASE="${BASE%/}"
EMAIL="${2:-${SMOKE_EMAIL:-ceo@gigmann.health}}"
PASSWORD="${3:-${SMOKE_PASSWORD:-ahenfie-demo}}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

fail() { echo "SMOKE FAIL: $*" >&2; exit 1; }

# req METHOD PATH [BEARER] [JSON_BODY] -> sets CODE and writes the body to $TMP/body
req() {
  local method="$1" path="$2" bearer="${3:-}" data="${4:-}"
  local args=(-sS -o "$TMP/body" -w '%{http_code}' -X "$method" "$BASE$path" --max-time 20)
  [ -n "$bearer" ] && args+=(-H "Authorization: Bearer $bearer")
  [ -n "$data" ] && args+=(-H 'Content-Type: application/json' -d "$data")
  CODE="$(curl "${args[@]}")" || fail "request error: $method $path"
}

jget() { python3 -c 'import sys,json;print(json.load(sys.stdin).get(sys.argv[1],""))' "$1" <"$TMP/body"; }

echo "==> health";  req GET /healthz; [ "$CODE" = 200 ] || fail "/healthz -> $CODE"
echo "==> ready";   req GET /readyz;  [ "$CODE" = 200 ] || fail "/readyz -> $CODE"

echo "==> login"
req POST /api/v1/auth/login "" "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
[ "$CODE" = 200 ] || fail "login -> $CODE"
TOKEN="$(jget token)"
[ -n "$TOKEN" ] || fail "login returned no token"

echo "==> brief"
req GET /api/v1/brief "$TOKEN"
[ "$CODE" = 200 ] || fail "/brief -> $CODE"
python3 -c 'import sys,json;d=json.load(sys.stdin);sys.exit(0 if isinstance(d.get("items"),list) else 1)' <"$TMP/body" \
  || fail "/brief response missing items[]"

echo "==> metrics"
req GET /api/v1/metrics "$TOKEN"
[ "$CODE" = 200 ] || fail "/metrics -> $CODE"

echo "SMOKE PASS: health, ready, login, brief, metrics OK against $BASE"
