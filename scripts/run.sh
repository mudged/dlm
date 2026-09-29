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

# go run places the binary in a temp directory, so the release-layout lookup
# (sibling runtime/cv/) does not see a bundle built into dist/. Point at that
# bundle when the operator has not chosen one.
if [[ -z "${DLM_CV_RUNTIME_DIR:-}" ]]; then
  cv_arch=""
  if command -v go >/dev/null 2>&1; then
    cv_arch="$(go env GOARCH 2>/dev/null || true)"
  fi
  if [[ -z "$cv_arch" ]]; then
    case "$(uname -m)" in
      x86_64) cv_arch=amd64 ;;
      aarch64 | arm64) cv_arch=arm64 ;;
    esac
  fi
  cv_dir="${ROOT}/dist/cvruntime/linux_${cv_arch}"
  if [[ -n "$cv_arch" && -d "${cv_dir}/python" && -f "${cv_dir}/reconstruct.py" ]]; then
    export DLM_CV_RUNTIME_DIR="$cv_dir"
  fi
fi

exec go run ./cmd/server
