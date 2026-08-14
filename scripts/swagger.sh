#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

(
  cd "$REPO_ROOT/backend"
  go run ./cmd/swagger
)

(
  cd "$REPO_ROOT/frontend"
  pnpm run api:generate
)
