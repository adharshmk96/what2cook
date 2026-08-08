#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEV_DIR="${ROOT_DIR}/.dev"
API_DIR="${ROOT_DIR}/what2cook-api"
SERVER_PID_FILE="${DEV_DIR}/server.pid"
SERVER_LOG="${DEV_DIR}/server.log"

echo "==> Stopping previous server"
"${ROOT_DIR}/scripts/stop-dev.sh"

echo "==> Building UI and server"
"${ROOT_DIR}/scripts/build.sh"

mkdir -p "${DEV_DIR}"

echo "==> Starting server"
(
  cd "${API_DIR}"
  nohup ./what2cook serve >"${SERVER_LOG}" 2>&1 &
  echo $! >"${SERVER_PID_FILE}"
)

server_ok=0
for _ in $(seq 1 40); do
  if ! kill -0 "$(cat "${SERVER_PID_FILE}")" 2>/dev/null; then
    break
  fi

  if curl -sf "http://127.0.0.1:8080/healthz" >/dev/null 2>&1; then
    server_ok=1
    break
  fi

  sleep 0.25
done

if [[ "${server_ok}" -ne 1 ]]; then
  echo "ERROR: Server failed to start on :8080. Last log lines:"
  tail -n 20 "${SERVER_LOG}" || true
  "${ROOT_DIR}/scripts/stop-dev.sh"
  exit 1
fi

echo "==> Server started"
echo "  App:  http://localhost:8080/app/"
echo "  API:  http://localhost:8080/api/v1/"
echo "  PID:  $(cat "${SERVER_PID_FILE}")"
echo "  Log:  ${SERVER_LOG}"
echo "  Stop: task stop-dev"
