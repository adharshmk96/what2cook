#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVER_PID_FILE="${ROOT_DIR}/.dev/server.pid"
LEGACY_SERVER_PID_FILE="${ROOT_DIR}/.dev/api.pid"

stop_server() {
  local pid_file="$1"

  if [[ ! -f "${pid_file}" ]]; then
    return 1
  fi

  local pid
  pid="$(cat "${pid_file}")"
  if [[ -z "${pid}" ]] || ! kill -0 "${pid}" 2>/dev/null; then
    rm -f "${pid_file}"
    return 1
  fi

  echo "==> Stopping server pid ${pid}"

  # Legacy `go run` uses a child process for the compiled server.
  local child_pids
  child_pids="$(pgrep -P "${pid}" 2>/dev/null || true)"
  kill ${child_pids} "${pid}" 2>/dev/null || true

  for _ in 1 2 3 4 5; do
    if ! kill -0 "${pid}" 2>/dev/null; then
      rm -f "${pid_file}"
      echo "==> Server stopped"
      return 0
    fi
    sleep 0.4
  done

  echo "==> Server still running; force stopping pid ${pid}"
  kill -9 ${child_pids} "${pid}" 2>/dev/null || true
  rm -f "${pid_file}"
  echo "==> Server stopped"
}

stopped=0
stop_server "${SERVER_PID_FILE}" && stopped=1
stop_server "${LEGACY_SERVER_PID_FILE}" && stopped=1

if [[ "${stopped}" -eq 0 ]]; then
  echo "==> Server not running"
fi
