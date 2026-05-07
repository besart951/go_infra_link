#!/usr/bin/env bash
set -euo pipefail

TARGET="${1:-all}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_TOOL_BIN="$REPO_ROOT/.cache/go-tools/bin"
RACE_PACKAGES=(
  ./internal/app
  ./internal/handler/middleware
  ./internal/infrastructure/exporting
  ./internal/infrastructure/realtime
  ./internal/service/exporting
  ./internal/service/notification
)

case "$TARGET" in
  all|backend|frontend) ;;
  *)
    echo "usage: scripts/ci.sh [all|backend|frontend]" >&2
    exit 2
    ;;
esac

install_go_tool() {
  local binary="$1"
  local package="$2"

  mkdir -p "$GO_TOOL_BIN"
  if [[ ! -x "$GO_TOOL_BIN/$binary" ]]; then
    echo "[ci] installing $package"
    GOBIN="$GO_TOOL_BIN" go install "$package"
  fi
}

run_backend() {
  echo "[ci] backend: tests"
  cd "$REPO_ROOT/backend"
  go test ./...

  echo "[ci] backend: race tests"
  go test -race "${RACE_PACKAGES[@]}"

  echo "[ci] backend: go vet"
  go vet ./...

  install_go_tool staticcheck honnef.co/go/tools/cmd/staticcheck@v0.7.0
  echo "[ci] backend: staticcheck"
  "$GO_TOOL_BIN/staticcheck" ./...

  install_go_tool govulncheck golang.org/x/vuln/cmd/govulncheck@v1.1.4
  echo "[ci] backend: govulncheck"
  "$GO_TOOL_BIN/govulncheck" ./...
}

ensure_pnpm() {
  if ! command -v node >/dev/null 2>&1; then
    echo "Node.js 24.x is required for frontend CI, but node was not found." >&2
    exit 1
  fi

  local node_major
  node_major="$(node -p "process.versions.node.split('.')[0]")"
  if [[ "$node_major" != "24" ]]; then
    echo "Node.js 24.x is required for frontend CI. Current version: $(node -p "process.versions.node")" >&2
    exit 1
  fi

  if command -v corepack >/dev/null 2>&1; then
    corepack enable
    corepack prepare pnpm@10.29.1 --activate
  fi
  pnpm --version | grep -qx '10.29.1'
}

run_frontend() {
  echo "[ci] frontend: install"
  cd "$REPO_ROOT/frontend"
  ensure_pnpm
  CI=true pnpm install --frozen-lockfile

  echo "[ci] frontend: check"
  pnpm check

  echo "[ci] frontend: test"
  pnpm test

  echo "[ci] frontend: build"
  pnpm build

  echo "[ci] frontend: lint"
  pnpm lint
}

if [[ "$TARGET" == "all" || "$TARGET" == "backend" ]]; then
  run_backend
fi

if [[ "$TARGET" == "all" || "$TARGET" == "frontend" ]]; then
  run_frontend
fi
