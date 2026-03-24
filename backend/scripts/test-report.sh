#!/usr/bin/env bash
# test-report.sh — smoke test for report endpoints
# Usage: ./scripts/test-report.sh [port]
set -euo pipefail

PORT=${1:-18082}
DATA_PATH="./data/alpha"
JWT_SECRET="testsecret123"
MAGIC_LINK_1812="81ce2bb6-42fe-49b2-80c5-0558787c8471"
BASE="http://localhost:${PORT}"

# Build the server binary
echo "==> Building api server..."
go build -o ./tmp/api ./cmd/api/
echo ""

# Start the server
echo "==> Starting server on port ${PORT}..."
./tmp/api serve \
  --data-path "${DATA_PATH}" \
  --jwt-secret "${JWT_SECRET}" \
  --port "${PORT}" \
  2>&1 &
SERVER_PID=$!
trap 'echo ""; echo "==> Stopping server (pid ${SERVER_PID})..."; kill ${SERVER_PID} 2>/dev/null; wait ${SERVER_PID} 2>/dev/null' EXIT
sleep 1

# Login as empire 1812
echo "==> Logging in as empire 1812 (magic link: ${MAGIC_LINK_1812})..."
RESPONSE=$(curl -sf -X POST "${BASE}/api/login/${MAGIC_LINK_1812}")
JWT=$(echo "${RESPONSE}" | jq -r '.access_token')
echo "JWT: ${JWT}"
echo ""

# List reports for empire 1812 (should succeed — 200)
echo "==> GET /api/1812/reports  [expect 200 with report list]"
curl -sf -H "Authorization: Bearer ${JWT}" "${BASE}/api/1812/reports"
echo ""
echo ""

# Fetch report 0/0 for empire 1812 (should succeed — 200)
echo "==> GET /api/1812/reports/0/0  [expect 200 with report body]"
curl -sf -H "Authorization: Bearer ${JWT}" "${BASE}/api/1812/reports/0/0"
echo ""
echo ""

# Fetch report from empire 42 using empire 1812's JWT (should fail — 403)
echo "==> GET /api/42/reports/0/0  [expect 403 Forbidden — wrong empire]"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${JWT}" "${BASE}/api/42/reports/0/0")
if [ "${STATUS}" = "403" ]; then
  echo "PASS: got ${STATUS} Forbidden as expected"
else
  echo "FAIL: expected 403, got ${STATUS}"
  exit 1
fi
echo ""

echo "==> All checks passed."
