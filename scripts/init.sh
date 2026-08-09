#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="${ROOT_DIR}/what2cook-ui"
API_DIR="${ROOT_DIR}/what2cook-api"
CONFIG_EXAMPLE="${API_DIR}/config.yaml.example"
CONFIG_FILE="${API_DIR}/config.yaml"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: missing required command: $1"
    echo "  Install it, then re-run: task init"
    exit 1
  fi
}

echo "==> Checking prerequisites"
need_cmd go
need_cmd bun

go_version="$(go env GOVERSION 2>/dev/null || go version)"
bun_version="$(bun --version)"
echo "  go:  ${go_version}"
echo "  bun: ${bun_version}"

if ! command -v task >/dev/null 2>&1; then
  echo "WARN: task not found on PATH (optional after init; used for build/run-dev)"
fi

echo "==> Installing Go modules"
(
  cd "${API_DIR}"
  go mod download
)

echo "==> Installing UI dependencies"
(
  cd "${UI_DIR}"
  bun install
)

if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "==> Creating local config from example"
  cp "${CONFIG_EXAMPLE}" "${CONFIG_FILE}"
  echo "  Wrote ${CONFIG_FILE}"
  echo "  Edit auth.token_secret (and SMTP if needed) before production use"
else
  echo "==> Config already exists: ${CONFIG_FILE}"
fi

echo "==> Init complete — ready to run"
echo "  Dev:   task run-dev"
echo "  Build: task build"
echo "  App:   http://localhost:8080/app/"
