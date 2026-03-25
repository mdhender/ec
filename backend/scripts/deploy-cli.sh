#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist/linux"
REMOTE_HOST="epimethean.dev"
REMOTE_DIR="/opt/ec"

# Get the version, stripping any +dirty-* build metadata suffix.
VERSION="$(go run ./cmd/cli version)"
VERSION="${VERSION%%+dirty*}"

if [[ -z "${VERSION}" ]]; then
    echo "error: could not determine version" >&2
    exit 1
fi

echo "==> Building cli v${VERSION} for linux/amd64 (no CGO)..."
mkdir -p "${DIST_DIR}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/cli" ./cmd/cli/

echo "==> Deploying to ${REMOTE_HOST}:${REMOTE_DIR}/cli-${VERSION}..."
# rsync -avz --delete "${DIST_DIR}/cli" "${REMOTE_HOST}:${REMOTE_DIR}/cli-${VERSION}"
scp "${DIST_DIR}/cli" "${REMOTE_HOST}:${REMOTE_DIR}/cli-${VERSION}"

echo "==> Done. Deployed cli-${VERSION} to ${REMOTE_HOST}."
