#!/usr/bin/env bash
# REQ-008: build Next static export into backend embed tree, then run the Go server.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/web"

# Install strategy (performance):
# - DLM_SKIP_NPM_CI=1 — never run npm ci; use existing node_modules (fastest if deps already OK).
# - DLM_FORCE_NPM_CI=1 — always npm ci (CI, fresh clone, or after package-lock.json changed).
# - Otherwise: npm ci only when node_modules is missing; if present, go straight to release:sync
#   (saves minutes on repeat runs; use DLM_FORCE_NPM_CI=1 after pulling lockfile updates).
if [[ "${DLM_SKIP_NPM_CI:-}" == "1" ]]; then
  npm run release:sync
elif [[ "${DLM_FORCE_NPM_CI:-}" == "1" ]] || [[ ! -d node_modules ]]; then
  npm ci
  npm run release:sync
else
  npm run release:sync
fi

cd "$ROOT/backend"
exec go run ./cmd/server
