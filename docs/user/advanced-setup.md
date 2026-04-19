# Advanced setup and developer notes

This document is for people who want more detail than the main [README](../../README.md): performance tuning, development workflows, cross-compilation, and API-oriented tips.

## One-command build and run (repository standard)

From the **repository root** (same as README), the supported path is:

```bash
./scripts/run.sh
```

This installs web dependencies when needed (`npm ci` if `web/node_modules` is missing), runs `npm run release:sync` in `web/` to build the Next.js static export into the Go embed tree, then starts the Go server with `go run ./cmd/server` in `backend/` (the script **replaces** the shell via `exec`).

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/).

### Environment variables for `run.sh`

| Variable | Effect |
|----------|--------|
| `DLM_SKIP_NPM_CI=1` | Skip `npm ci`; only runs `release:sync` (fastest when `node_modules` is already correct). |
| `DLM_FORCE_NPM_CI=1` | Always run `npm ci` before `release:sync` (clean install, CI, or after `package-lock.json` changes). |

If `web/node_modules` already exists and neither skip nor force is set, the script skips `npm ci` and only runs `release:sync`.

### Why `run.sh` can feel slow

Typical bottlenecks:

1. **`npm ci`** — reinstalls `web/node_modules` from the lockfile (network + disk). Mitigations: rely on skip-when-present behavior, `DLM_SKIP_NPM_CI=1`, or a dev environment with dependencies pre-installed.
2. **`next build`** (inside `release:sync`) — production compile and static export; often on the order of tens of seconds to a few minutes depending on CPU. Next caches under `web/.next/`; the second build is usually faster.
3. **`go run ./cmd/server`** — compiles Go; subsequent runs reuse the build cache.

To measure: `time ./scripts/run.sh` or run `time npm ci`, `time npm run release:sync`, and `time go run ./cmd/server` separately under `web/` and `backend/`.

## Prerequisites (versions)

- **Go:** The module in `backend/go.mod` declares the required language version (currently **1.25.x**). Install a matching toolchain from [go.dev/dl](https://go.dev/dl/). If your system ships an older Go, install a newer release and ensure `go version` matches.
- **Node.js:** Use an **Active LTS** release for building `web/` (see [nodejs.org](https://nodejs.org/)). Node is for building the UI; a small board can run only the compiled Go binary if you build elsewhere and copy artifacts.
- **Python 3** (optional): For **Python scene routines** on the host, install Python 3 and ensure `python3` is on `PATH`. Other features do not require it.

## Manual two-step run (same result as the script)

```bash
cd web && npm run release:sync
cd ../backend && go run ./cmd/server
```

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/) — HTML and `/api/v1/...` are served by the same process.

### Cross-compile for Raspberry Pi (arm64)

From `backend/`:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server
```

## Development: two processes (not what you ship)

1. **Go server** (API + embedded static UI after a sync, or rebuild after UI changes):

   ```bash
   cd backend
   go run ./cmd/server
   ```

2. **Next.js dev server** (hot reload; proxies API to the Go port):

   ```bash
   cd web
   npm run dev
   ```

Open [http://localhost:3000](http://localhost:3000) (or `http://127.0.0.1:3000`). The dev app rewrites `/api/v1` to the Go backend (default `http://127.0.0.1:8080`).

**Server-Sent Events:** `next dev` can buffer `text/event-stream`, so the UI may open `EventSource` directly against the Go origin (`http://127.0.0.1:8080`). The Go server defaults `CORS_ALLOWED_ORIGINS` to common Next dev origins (`:3000` and `:8000`, localhost and 127.0.0.1). If the API is not on `http://127.0.0.1:8080`, set `DLM_BACKEND_ORIGIN` and `NEXT_PUBLIC_DLM_API_ORIGIN` in `web/.env.local` (see `web/.env.local.example`). Override with `CORS_ALLOWED_ORIGINS` for other page origins; `CORS_ALLOWED_ORIGINS=-` disables CORS headers entirely.

### If the 3D view does not update during a routine (dev)

1. In DevTools → **Network**, find the **`events`** request whose URL ends with `/api/v1/scenes/<scene-id>/lights/events`. In **EventStream**, **data** lines should look like `{"seq":1,"deltas":[...]}`. If you only see unrelated payloads, select the request whose URL contains `/lights/events`.
2. **Go on a non-default port:** set `NEXT_PUBLIC_DLM_API_ORIGIN` and `DLM_BACKEND_ORIGIN` to that base URL.
3. **Embedded UI only:** after UI changes, run `cd web && npm run release:sync` and restart Go.

## Command reference

| Area | Command |
|------|---------|
| One-shot build + run | `./scripts/run.sh` from repo root (`DLM_FORCE_NPM_CI=1` or `DLM_SKIP_NPM_CI=1` as above) |
| Sync export → embed tree | `cd web && npm run release:sync` |
| Go tests | `cd backend && go test ./...` |
| Web lint | `cd web && npm run lint` |
| Web tests | `cd web && npm test` |
| Web export only | `cd web && npm run build` (output in `web/out/`) |
| Download Go deps | `cd backend && go mod download` |
| Install frontend deps | `cd web && npm ci` |

## High-throughput light updates (API-oriented)

- Prefer **batch** routes when updating many lights (`PATCH /api/v1/models/{id}/lights/state/batch` and scene bulk routes in [`docs/agentic-development/architecture.md`](../agentic-development/architecture.md)) instead of one HTTP request per light.
- Reuse HTTP keep-alive. For many parallel clients, terminating TLS + HTTP/2 in a reverse proxy while proxying HTTP/1.1 to Go is a common pattern (see [architecture](../agentic-development/architecture.md)).
- After commits, the server emits **SSE** on `GET /api/v1/models/{id}/lights/events` and `GET /api/v1/scenes/{id}/lights/events` (`text/event-stream`). Subscribers typically refetch authoritative state after each event.
- If SSE drops behind a proxy, increase **`HTTP_WRITE_TIMEOUT_SEC`** or set `0` for no write deadline when debugging; the handler extends deadlines periodically.

## Embedded database and CSV upload

- **SQLite** is embedded (pure Go, no separate database server). The database file is created automatically at **`data/dlm.db`** by default (see server configuration if you override the data directory).
- **CSV upload** for models: `POST /api/v1/models` expects the file field name **`file`** (not `csv`). Header must be exactly `id,x,y,z`. Light IDs must be **0-based** sequential integers (`0`, `1`, `2`, …).

For system design and diagrams, see **[architecture.md](../agentic-development/architecture.md)**.
