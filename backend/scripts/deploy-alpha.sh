#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="epimethean.dev"
REMOTE_DIR="/var/lib/ec/alpha"

cd data || {
  echo "error: could not set def to alpha data"
  exit 2
}

[ -f "alpha/game.json" ] || {
    echo "error: could not find alpha data" >&2
    exit 1
}


echo "==> Deploying to ${REMOTE_HOST}:${REMOTE_DIR}/..."
rsync -avz --delete alpha/ "${REMOTE_HOST}:${REMOTE_DIR}/"
exit 2

echo "==> Done. Deployed alpha data ${REMOTE_HOST}:${REMOTE_DIR}."

exit 0
