#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
IAM_BASE_URL="${IAM_BASE_URL:-http://localhost:8081}"
ORCH_BASE_URL="${ORCH_BASE_URL:-http://localhost:8084}"
LOGIN="${LOGIN:-admin}"
PASSWORD="${PASSWORD:-admin123}"
COOKIE_JAR="$ROOT_DIR/.metrics_export.cookies"

curl -sS -X POST "$IAM_BASE_URL/api/iam/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"login\":\"$LOGIN\",\"password\":\"$PASSWORD\"}" \
  -c "$COOKIE_JAR" >/dev/null

curl -sS -X POST "$ORCH_BASE_URL/api/orchestrator/metrics/export-csv" \
  -b "$COOKIE_JAR"

echo
