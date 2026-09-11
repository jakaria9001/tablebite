#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ ! -f "$ROOT/.env" ] && [ -f "$ROOT/.env.example" ]; then
  cp "$ROOT/.env.example" "$ROOT/.env"
  echo "Created .env from .env.example"
fi

set -a
source "$ROOT/.env"
set +a

command -v psql >/dev/null 2>&1 || { echo "PostgreSQL client (psql) is required."; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js is required."; exit 1; }
command -v go >/dev/null 2>&1 || { echo "Go is required."; exit 1; }

if [ ! -d "$ROOT/frontend/node_modules" ]; then
  (cd "$ROOT/frontend" && npm install)
fi

(cd "$ROOT/backend" && go mod download && go run ./cmd/migrate && go run ./cmd/seed-menu)

trap 'kill 0' EXIT
(cd "$ROOT/backend" && go run ./cmd/server) &
(cd "$ROOT/frontend" && npm run dev) &
wait
