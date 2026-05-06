#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$REPO_ROOT/tmp/dev"
LOG_DIR="$STATE_DIR/logs"

mkdir -p "$LOG_DIR"

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

strip_wrapping_quotes() {
  local value="$1"

  if [[ ${#value} -ge 2 ]]; then
    if [[ ${value:0:1} == '"' && ${value: -1} == '"' ]]; then
      printf '%s' "${value:1:${#value}-2}"
      return 0
    fi

    if [[ ${value:0:1} == "'" && ${value: -1} == "'" ]]; then
      printf '%s' "${value:1:${#value}-2}"
      return 0
    fi
  fi

  printf '%s' "$value"
}

import_dotenv() {
  local path="$1"
  [[ -f "$path" ]] || return 0

  while IFS= read -r line || [[ -n "$line" ]]; do
    line="$(trim "$line")"
    [[ -z "$line" || ${line:0:1} == "#" ]] && continue
    [[ "$line" == *=* ]] || continue

    local key="${line%%=*}"
    local raw="${line#*=}"
    local comment_marker=' #'

    key="$(trim "$key")"
    raw="$(trim "$raw")"

    if [[ "$raw" == *"$comment_marker"* ]]; then
      raw="${raw%%"$comment_marker"*}"
      raw="$(trim "$raw")"
    fi

    raw="$(strip_wrapping_quotes "$raw")"

    [[ -n "$key" ]] && export "$key=$raw"
  done <"$path"
}

import_dotenv "$REPO_ROOT/.env"

step() {
  printf '[dev] %s\n' "$1"
}

warn() {
  printf '[dev] WARNING: %s\n' "$1" >&2
}

fail() {
  printf '[dev] ERROR: %s\n' "$1" >&2
  exit 1
}

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
    return
  fi

  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
    return
  fi

  fail 'Docker Compose is not installed.'
}

require_command() {
  local command_name="$1"
  local message="$2"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    fail "$message"
  fi
}

is_process_running() {
  local pid="$1"
  [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1
}

start_detached_process() {
  local name="$1"
  local workdir="$2"
  shift 2

  local pid_file="$STATE_DIR/$name.pid"
  local log_file="$LOG_DIR/$name.log"

  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(<"$pid_file")"
    if is_process_running "$pid"; then
      step "$name is already running (pid $pid)."
      step "Logs: $log_file"
      return 0
    fi
    rm -f "$pid_file"
  fi

  : >"$log_file"
  (
    cd "$workdir"
    nohup "$@" >"$log_file" 2>&1 &
    echo $! >"$pid_file"
  )

  local pid
  pid="$(<"$pid_file")"
  step "Started $name (pid $pid)."
  step "Logs: $log_file"
}

stop_local_process() {
  local name="$1"
  local pid_file="$STATE_DIR/$name.pid"

  [[ -f "$pid_file" ]] || return 0

  local pid
  pid="$(<"$pid_file")"
  if is_process_running "$pid"; then
    step "Stopping $name (pid $pid)..."
    kill "$pid" >/dev/null 2>&1 || true
  fi

  rm -f "$pid_file"
}

start_postgres() {
  step 'Starting postgres + pgAdmin via docker compose...'
  (
    cd "$REPO_ROOT"
    compose up -d postgres pgadmin
  )
}

wait_for_postgres() {
  local attempts="${1:-30}"
  local delay_seconds="${2:-2}"

  step 'Waiting for postgres to become ready...'
  (
    cd "$REPO_ROOT"
    for ((attempt = 1; attempt <= attempts; attempt++)); do
      if compose exec -T postgres sh -lc 'pg_isready -U "${POSTGRES_USER:-postgres}" -p "${POSTGRES_CONTAINER_PORT:-5432}"' >/dev/null 2>&1; then
        return 0
      fi
      sleep "$delay_seconds"
    done
    return 1
  ) || fail 'PostgreSQL did not become ready in time.'
}

run_db_bootstrap() {
  require_command go 'Go is required to run backend bootstrap.'
  step 'Running database bootstrap...'
  (
    cd "$REPO_ROOT/backend"
    go run ./cmd/db-bootstrap/
  )
}

get_go_bin() {
  local go_bin
  go_bin="$(go env GOBIN)"
  if [[ -n "$go_bin" ]]; then
    printf '%s\n' "$go_bin"
    return 0
  fi

  local go_path
  go_path="$(go env GOPATH)"
  printf '%s/bin\n' "$go_path"
}

ensure_air() {
  if command -v air >/dev/null 2>&1; then
    command -v air
    return 0
  fi

  local air_bin
  air_bin="$(get_go_bin)/air"
  if [[ -x "$air_bin" ]]; then
    printf '%s\n' "$air_bin"
    return 0
  fi

  step 'Installing Air for backend hot reload...' >&2
  go install github.com/air-verse/air@latest

  [[ -x "$air_bin" ]] || fail "Air was installed, but '$air_bin' was not found. Check 'go env GOPATH' and 'go env GOBIN'."
  printf '%s\n' "$air_bin"
}

start_backend() {
  require_command go 'Go is required to start the backend.'

  local air_command
  air_command="$(ensure_air)"

  step 'Starting backend with Air hot reload in the background...'
  start_detached_process backend "$REPO_ROOT/backend" "$air_command"
}

start_frontend() {
  require_command pnpm 'pnpm is required to start the frontend.'
  step 'Starting frontend in the background...'
  start_detached_process frontend "$REPO_ROOT/frontend" pnpm dev
}

run_seed() {
  require_command go 'Go is required to run the seeder.'
  step 'Running seeder...'
  (
    cd "$REPO_ROOT/backend"
    go run ./cmd/seeder/
  )
}

reset_database() {
  if [[ "$FORCE" != 'true' ]]; then
    printf '\n'
    warn 'This will delete all data in PostgreSQL.'
    read -r -p 'Type RESET to continue: ' confirmation
    if [[ "$confirmation" != 'RESET' ]]; then
      step 'Cancelled reset.'
      return 0
    fi
  fi

  step 'Resetting PostgreSQL schema public...'
  (
    cd "$REPO_ROOT"
    compose exec -T postgres sh -lc 'psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-go_infra_link}" -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO ${POSTGRES_USER:-postgres}; GRANT ALL ON SCHEMA public TO public;"'
  )
}

stop_all() {
  stop_local_process frontend
  stop_local_process backend

  step 'Stopping containers...'
  (
    cd "$REPO_ROOT"
    compose stop
  )
}

show_help() {
  cat <<'EOF'
Usage:
  ./scripts/dev.sh <action> [--force]

Actions:
  start      Start postgres+pgAdmin, run db bootstrap, then backend with Air hot reload and frontend in the background
  postgres   Start only postgres + pgAdmin
  pgadmin    Start only pgAdmin
  backend    Run db bootstrap, then start only backend with Air hot reload in the background
  frontend   Start only frontend in the background
  bootstrap  Run backend db bootstrap once
  seed       Run db bootstrap, then backend seeder once
  reset-db   Drop & recreate public schema, then run db bootstrap
  reseed     reset-db + seed
  stop       Stop docker compose services and background processes started by this script
  format     Run gofmt -w on the backend
  help       Show this help

Examples:
  ./scripts/dev.sh start
  ./scripts/dev.sh reset-db --force
  ./scripts/dev.sh reseed --force

Logs:
  Backend log : ./tmp/dev/logs/backend.log
  Frontend log: ./tmp/dev/logs/frontend.log
EOF
}

ACTION='start'
FORCE='false'

while [[ $# -gt 0 ]]; do
  case "$1" in
    start|postgres|pgadmin|backend|frontend|bootstrap|seed|reset-db|reseed|stop|format|help)
      ACTION="$1"
      ;;
    -f|--force)
      FORCE='true'
      ;;
    -h|--help)
      ACTION='help'
      ;;
    *)
      fail "Unknown argument: $1"
      ;;
  esac
  shift
done

case "$ACTION" in
  start)
    start_postgres
    wait_for_postgres
    run_db_bootstrap
    start_backend
    start_frontend
    step "Backend health: http://localhost:${BACKEND_PORT:-8080}/health"
    step "Frontend: http://localhost:${FRONTEND_PORT:-5173}"
    ;;
  postgres)
    start_postgres
    ;;
  pgadmin)
    step 'Starting pgAdmin via docker compose...'
    (
      cd "$REPO_ROOT"
      compose up -d pgadmin
    )
    ;;
  backend)
    run_db_bootstrap
    start_backend
    ;;
  frontend)
    start_frontend
    ;;
  bootstrap)
    run_db_bootstrap
    ;;
  seed)
    run_db_bootstrap
    run_seed
    ;;
  reset-db)
    reset_database
    run_db_bootstrap
    ;;
  reseed)
    reset_database
    run_db_bootstrap
    run_seed
    ;;
  stop)
    stop_all
    ;;
  format)
    require_command gofmt 'gofmt is required to format the backend.'
    step 'Formatting Go code (gofmt -w)...'
    (
      cd "$REPO_ROOT/backend"
      gofmt -w .
    )
    step 'Done. Run: gofmt -l . to list any remaining differences.'
    ;;
  help)
    show_help
    ;;
esac