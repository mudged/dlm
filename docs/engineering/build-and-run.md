# Build and run (developers)

This is the developer-facing build and run reference. End users who just want to run a published
release should read the [user guide](../userguide/getting-started.md) instead.

## Prerequisites (versions)

- **Go ≥ 1.25** — `backend/go.mod` declares `go 1.25.0`, so an older system Go (e.g. 1.22) will
  refuse to build. Install a matching toolchain from [go.dev/dl](https://go.dev/dl/) and ensure
  `go version` matches (on Cursor Cloud, `/usr/local/go/bin` must be on `PATH`).
- **Node.js** — an **Active LTS** release (22.x is fine) for building `web/`. Node is only needed at
  build time; a small board can run just the compiled Go binary if you build elsewhere and copy the
  artifact.
- **Python 3** (optional) — only for **Python scene routines** on the host. Install Python 3 and
  ensure `python3` is on `PATH`. No other feature requires it.

## One-command build and run (repository standard)

From the **repository root**:

```bash
./scripts/run.sh
```

This installs web dependencies when needed (`npm ci` if `web/node_modules` is missing), runs
`npm run release:sync` in `web/` to build the Next.js static export into the Go embed tree, then
starts the Go server with `go run ./cmd/server` in `backend/` (the script **replaces** the shell via
`exec`). Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/).

### Environment variables for `run.sh`

| Variable | Effect |
|----------|--------|
| `DLM_SKIP_NPM_CI=1` | Skip `npm ci`; only run `release:sync` (fastest when `node_modules` is already correct). |
| `DLM_FORCE_NPM_CI=1` | Always run `npm ci` before `release:sync` (clean install, CI, or after `package-lock.json` changes). |

If `web/node_modules` already exists and neither skip nor force is set, the script skips `npm ci` and
only runs `release:sync`.

### Why `run.sh` can feel slow

1. **`npm ci`** — reinstalls `web/node_modules` from the lockfile (network + disk). Mitigate with
   the skip-when-present behavior or `DLM_SKIP_NPM_CI=1`.
2. **`next build`** (inside `release:sync`) — production compile and static export; tens of seconds
   to a few minutes depending on CPU. Next caches under `web/.next/`; the second build is faster.
3. **`go run ./cmd/server`** — compiles Go; subsequent runs reuse the build cache.

To measure: `time ./scripts/run.sh`, or time `npm ci`, `npm run release:sync`, and
`go run ./cmd/server` separately.

## Manual two-step run (same result as the script)

```bash
cd web && npm run release:sync
cd ../backend && go run ./cmd/server
```

HTML and `/api/v1/...` are served by the same process at
[http://127.0.0.1:8080/](http://127.0.0.1:8080/).

## Development: two processes (not what you ship)

1. **Go server** (API + embedded static UI after a sync, or rebuild after UI changes):

   ```bash
   cd backend && go run ./cmd/server
   ```

2. **Next.js dev server** (hot reload; proxies API to the Go port):

   ```bash
   cd web && npm run dev
   ```

Open [http://localhost:3000](http://localhost:3000) (or `http://127.0.0.1:3000`). The dev app
rewrites `/api/v1` to the Go backend (default `http://127.0.0.1:8080`).

**Server-Sent Events (live light updates):** `next dev` can buffer `text/event-stream`, so the UI
opens `EventSource` directly against the Go origin (`http://127.0.0.1:8080`, see `web/lib/sseUrl.ts`).
The Go server defaults `CORS_ALLOWED_ORIGINS` to common Next dev origins (`:3000` and `:8000`,
`localhost` and `127.0.0.1`). If the API is not on `http://127.0.0.1:8080`, set `DLM_BACKEND_ORIGIN`
and `NEXT_PUBLIC_DLM_API_ORIGIN` in `web/.env.local` (see `web/.env.local.example`). Override with
`CORS_ALLOWED_ORIGINS` for other page origins; `CORS_ALLOWED_ORIGINS=-` disables CORS headers
entirely. `./scripts/run.sh` serves UI and API on the same origin and needs none of this.

### If the 3D view does not update during a routine (dev)

1. In DevTools → **Network**, find the **`events`** request ending with
   `/api/v1/scenes/<scene-id>/lights/events`; in **EventStream**, **data** lines should look like
   `{"seq":1,"deltas":[...]}`.
2. **Go on a non-default port:** set `NEXT_PUBLIC_DLM_API_ORIGIN` and `DLM_BACKEND_ORIGIN` to that
   base URL.
3. **Embedded UI only:** after UI changes, run `cd web && npm run release:sync` and restart Go.

## Cross-compile for Raspberry Pi (arm64)

From `backend/`:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server
```

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

## Gotchas

- **`go.mod` says `go 1.25.0`** — the default system Go (1.22) will refuse to build. Ensure a Go
  1.25.x toolchain is on `PATH` (`/usr/local/go/bin` on Cursor Cloud).
- **No external services** — SQLite is embedded (pure Go, no cgo). The database file auto-creates at
  `data/dlm.db`. No Docker, Redis, or Postgres needed.
- **Sample data** — on first start with an empty DB, 3 sample models (sphere, cube, cone) are
  auto-seeded, along with three default Python routines. Deleting all models and restarting
  re-seeds them.
- **CSV upload field name** — `POST /api/v1/models` expects the CSV file under field name `file`
  (not `csv`), and the CSV header must be exactly `id,x,y,z`. Light IDs must be **0-based sequential**
  integers (`0`, `1`, `2`, …); 1-based IDs fail validation.
- **`next build` is the slow step** (~30–120 s). The `.next/` cache speeds up subsequent builds. Use
  `DLM_SKIP_NPM_CI=1 ./scripts/run.sh` to skip `npm ci` on repeat runs.

For environment variables, database, and API tuning see
[`environment-and-api.md`](environment-and-api.md). For system design and diagrams see
[`../design/architecture.md`](../design/architecture.md).
