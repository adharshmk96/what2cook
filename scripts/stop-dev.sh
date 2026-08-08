#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEV_DIR="${ROOT_DIR}/.dev"
API_PID_FILE="${DEV_DIR}/api.pid"
UI_PID_FILE="${DEV_DIR}/ui.pid"

# Kill pid and all descendants (covers go run / bun child processes).
kill_tree() {
  local pid="$1"
  local kids
  kids="$(pgrep -P "${pid}" 2>/dev/null || true)"
  local kid
  for kid in ${kids}; do
    kill_tree "${kid}"
  done
  kill "${pid}" 2>/dev/null || true
}

force_kill_tree() {
  local pid="$1"
  local kids
  kids="$(pgrep -P "${pid}" 2>/dev/null || true)"
  local kid
  for kid in ${kids}; do
    force_kill_tree "${kid}"
  done
  kill -9 "${pid}" 2>/dev/null || true
}

stop_pid() {
  local name="$1"
  local pid_file="$2"

  if [[ ! -f "${pid_file}" ]]; then
    echo "==> ${name}: not running (no pid file)"
    return 0
  fi

  local pid
  pid="$(cat "${pid_file}")"
  if [[ -z "${pid}" ]]; then
    echo "==> ${name}: empty pid file, removing"
    rm -f "${pid_file}"
    return 0
  fi

  if ! kill -0 "${pid}" 2>/dev/null; then
    echo "==> ${name}: stale pid ${pid}, removing"
    rm -f "${pid_file}"
    return 0
  fi

  echo "==> ${name}: stopping pid ${pid}"
  kill_tree "${pid}"

  local i
  for i in 1 2 3 4 5; do
    if ! kill -0 "${pid}" 2>/dev/null; then
      rm -f "${pid_file}"
      echo "==> ${name}: stopped"
      return 0
    fi
    sleep 0.4
  done

  echo "==> ${name}: still alive, force killing"
  force_kill_tree "${pid}"
  rm -f "${pid_file}"
  echo "==> ${name}: force stopped"
}

stop_pid "API" "${API_PID_FILE}"
stop_pid "UI" "${UI_PID_FILE}"

# Free ports left behind by orphan binaries (e.g. ./what2cook from task build).
free_port() {
  local port="$1"
  local pids
  pids="$(lsof -nP -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -z "${pids}" ]]; then
    return 0
  fi
  echo "==> Freeing port ${port}: ${pids}"
  local pid
  for pid in ${pids}; do
    kill_tree "${pid}"
  done
  sleep 0.3
  pids="$(lsof -nP -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  for pid in ${pids}; do
    force_kill_tree "${pid}"
  done
}

free_port 8080
free_port 5173

echo "==> Dev processes stopped"
