#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"

curl -fsS "$BASE_URL/health" >/dev/null

curl -fsS -X POST "$BASE_URL/rag/ask" \
  -H "Content-Type: application/json" \
  -d '{"question":"What is this knowledge base about?","top_k":3}' >/dev/null

status="$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/rag/ingest")"
case "$status" in
  401|404) ;;
  *) echo "expected protected ingest to return 401 or 404, got $status" >&2; exit 1 ;;
esac

echo "smoke checks passed"
