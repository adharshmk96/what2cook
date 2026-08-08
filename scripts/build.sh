#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="${ROOT_DIR}/what2cook-ui"
API_DIR="${ROOT_DIR}/what2cook-api"
DIST_SRC="${UI_DIR}/dist"
DIST_DST="${API_DIR}/web/dist"

echo "==> Building UI"
(
  cd "${UI_DIR}"
  bun run build
)

echo "==> Syncing UI into API embed path"
rm -rf "${DIST_DST}"
mkdir -p "${DIST_DST}"
cp -R "${DIST_SRC}/." "${DIST_DST}/"
touch "${DIST_DST}/.gitkeep"

echo "==> Building API binary"
(
  cd "${API_DIR}"
  go build -o what2cook .
)

echo "==> Build complete: ${API_DIR}/what2cook"
