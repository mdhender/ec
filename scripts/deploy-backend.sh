#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist/linux"
REMOTE_HOST="epimethean.dev"
REMOTE_DIR="/var/www/app.epimethean.dev/backend/bin"

# Get the version, stripping any +dirty-* build metadata suffix.
VERSION="$(cd "${REPO_ROOT}/backend" && go run ./cmd/api show version)"
VERSION="${VERSION%%+dirty*}"

if [[ -z "${VERSION}" ]]; then
    echo "error: could not determine version" >&2
    exit 1
fi

echo "==> Building api v${VERSION} for linux/amd64 (no CGO)..."
mkdir -p "${DIST_DIR}"
(
    cd "${REPO_ROOT}/backend"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/api" ./cmd/api/
)

echo "==> Deploying to ${REMOTE_HOST}:${REMOTE_DIR}/api-${VERSION}..."
rsync -avz --delete "${DIST_DIR}/api" "${REMOTE_HOST}:${REMOTE_DIR}/api-${VERSION}"

echo "==> Done. Deployed api-${VERSION} to ${REMOTE_HOST}."
