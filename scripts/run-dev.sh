#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEV_DIR="${ROOT_DIR}/.dev"
API_DIR="${ROOT_DIR}/what2cook-api"
UI_DIR="${ROOT_DIR}/what2cook-ui"
API_PID_FILE="${DEV_DIR}/api.pid"
UI_PID_FILE="${DEV_DIR}/ui.pid"
API_LOG="${DEV_DIR}/api.log"
UI_LOG="${DEV_DIR}/ui.log"

is_running() {
  local pid_file="$1"
  if [[ ! -f "${pid_file}" ]]; then
    return 1
  fi
  local pid
  pid="$(cat "${pid_file}")"
  if [[ -z "${pid}" ]]; then
    return 1
  fi
  kill -0 "${pid}" 2>/dev/null
}

if is_running "${API_PID_FILE}" || is_running "${UI_PID_FILE}"; then
  echo "Dev processes already running:"
  if is_running "${API_PID_FILE}"; then
    echo "  API pid=$(cat "${API_PID_FILE}") (log: ${API_LOG})"
  fi
  if is_running "${UI_PID_FILE}"; then
    echo "  UI  pid=$(cat "${UI_PID_FILE}") (log: ${UI_LOG})"
  fi
  echo "Stop them first with: task stop-dev (or ./scripts/stop-dev.sh)"
  exit 1
fi

mkdir -p "${DEV_DIR}"
rm -f "${API_PID_FILE}" "${UI_PID_FILE}"

echo "==> Starting API (go run . serve)"
(
  cd "${API_DIR}"
  nohup go run . serve >"${API_LOG}" 2>&1 &
  echo $! >"${API_PID_FILE}"
)

echo "==> Starting UI (bun run dev)"
(
  cd "${UI_DIR}"
  nohup bun run dev >"${UI_LOG}" 2>&1 &
  echo $! >"${UI_PID_FILE}"
)

echo "==> Dev started"
echo "  API:  http://localhost:8080  (pid=$(cat "${API_PID_FILE}"), log=${API_LOG})"
echo "  UI:   http://localhost:5173  (pid=$(cat "${UI_PID_FILE}"), log=${UI_LOG})"
echo "  Stop: task stop-dev"
