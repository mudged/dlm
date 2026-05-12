# dlm

Domestic Light & Magic — **one Go binary** serves the **JSON API** and the **Next.js static export** (Tailwind) embedded at build time. See `docs/architecture.md` and `AGENTS.md`.

## One command (REQ-008)

From the **repository root**, after a clone (with **Go** and **Node.js** on `PATH`):

```bash
./scripts/run.sh
```

This runs **`npm ci`** and **`npm run release:sync`** in `web/`, then **`go run ./cmd/server`** in `backend/` (the Go process **replaces** the shell via `exec`). Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/).

- **Windows:** use **Git Bash**, **WSL**, or run the same steps manually.
- **Faster iteration** when `node_modules` is already present: `DLM_SKIP_NPM_CI=1 ./scripts/run.sh` runs **`npm run release:sync`** only (skips **`npm ci`**).
- **Repeat runs (default script):** if `web/node_modules` already exists, **`npm ci` is skipped** and only **`release:sync`** runs (still runs a full **`next build`**, which is usually the slow part). Use **`DLM_FORCE_NPM_CI=1 ./scripts/run.sh`** after **`package-lock.json`** changes or in **CI** for a clean install.

### Why `run.sh` can feel slow

Typical time on a cold machine is dominated by:

1. **`npm ci`** — wipes/reinstalls `node_modules` from the lockfile (network + disk). **Minutes** on a slow link. Mitigations: rely on the default **skip when `node_modules` exists**, **`DLM_SKIP_NPM_CI=1`**, or a persistent devcontainer layer with deps pre-installed.
2. **`next build`** (inside **`release:sync`**) — full production compile, lint, and static export. Often **~30–120 s** depending on CPU. Next caches under **`web/.next/`**; a second build is usually faster than the first.
3. **`go run ./cmd/server`** — compiles the Go module; **subsequent** runs reuse the build cache and are much quicker than the first.

To see where time goes: `time ./scripts/run.sh` or run `time npm ci`, `time npm run release:sync`, and `time go run ./cmd/server` separately under `web/` / `backend/`.

## Default samples (REQ-009)

On first start with an **empty** database, the server inserts three models — **Sample sphere**, **Sample cube**, and **Sample cone** — each with **500–1000** lights on the shape’s **outer surface**, consecutive spacing **5–10 cm** along the polyline (see `docs/architecture.md` §3.8). If you delete **all** models, the next process start seeds them again.

## Prerequisites

- **Go ≥ 1.25** (`backend/go.mod` declares `go 1.25.0`; the Cursor Cloud/devcontainer setup installs Go 1.25.x).
- **Node.js** — **Active LTS** in the devcontainer (`features/node` with **`version: "lts"`**); only needed to build `web/` (not on the Pi at runtime).
- **Python 3** (optional) — install on the **host** (e.g. Raspberry Pi) and ensure `python3` is on `PATH` if you run **Python scene routines**; see `docs/architecture.md` §3.17 and §6.2. Other features do not require it.

## Production-style run (manual two steps)

If you prefer not to use the script:

```bash
cd web && npm run release:sync
cd ../backend && go run ./cmd/server
```

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/) — HTML and `/api/v1/...` are served by the same program.

Cross-compile for Raspberry Pi (from `backend/`):

```text
GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server
```

## Development (two processes, not shipped)

1. **Go API + embedded UI** (re-sync after UI changes, or use `./scripts/run.sh` once):

   ```bash
   cd backend
   go run ./cmd/server
   ```

2. **Next dev** (hot reload; browser uses rewrites to `:8080` for `/api/v1`):

   ```bash
   cd web
   npm run dev
   ```

   Open [http://localhost:3000](http://localhost:3000) (or `http://127.0.0.1:3000`; **`next dev -p 8000`** is also allowed). The Go server **defaults** `CORS_ALLOWED_ORIGINS` to common Next dev origins (`:3000` and `:8000`, localhost + 127.0.0.1) so **Server-Sent Events** (`EventSource` to the Go port, usually `:8080`) work; `fetch` to `/api/v1` still uses Next rewrites. If the Go API is **not** on `http://127.0.0.1:8080`, set **`DLM_BACKEND_ORIGIN`** and **`NEXT_PUBLIC_DLM_API_ORIGIN`** to the same base URL in `web/.env.local` and restart Next (see `web/.env.local.example`). Override with `CORS_ALLOWED_ORIGINS` for other page origins; set `CORS_ALLOWED_ORIGINS=-` to disable CORS headers entirely.

### If the 3D view does not update during a routine (dev)

1. In DevTools → **Network**, select the **`events`** row whose **Request URL** ends with **`/api/v1/scenes/<scene-id>/lights/events`** (not a generic `events` name from another tool). Open **EventStream**: the **data** lines must look like `{"seq":1,"deltas":[...]}`. If you see something like `{"key":"started"}` only, that is **not** this app’s light stream — pick the request whose URL contains **`/lights/events`**.
2. **Go on a non-default port:** set **`NEXT_PUBLIC_DLM_API_ORIGIN`** and **`DLM_BACKEND_ORIGIN`** to that base URL (see `web/.env.local.example`).
3. **Embedded UI (`go run` only):** rebuild the web export after UI changes: `cd web && npm run release:sync`, then restart the Go server.

## Commands

| Area | Command |
|------|---------|
| **One-shot build + run (REQ-008)** | `./scripts/run.sh` from repo root (`DLM_FORCE_NPM_CI=1` for clean `npm ci`; `DLM_SKIP_NPM_CI=1` to skip install) |
| Sync export → embed tree | `cd web && npm run release:sync` |
| Go tests | `cd backend && go test ./...` |
| Web lint | `cd web && npm run lint` |
| Web export only | `cd web && npm run build` (output in `web/out/`) |

## High-throughput light updates (REQ-029)

- **Batch writes:** Prefer `PATCH /api/v1/models/{id}/lights/state/batch` and scene bulk routes (`PATCH /api/v1/scenes/{id}/lights/state/...` per `docs/architecture.md` §3.15) instead of one HTTP request per light when updating hundreds of lights frequently.
- **Connections:** Reuse HTTP keep-alive (default for Go’s server and typical browser `fetch` pools). For many parallel requests from a rich client, terminating **TLS + HTTP/2** in a reverse proxy while proxying HTTP/1.1 to the Go process is recommended in §3.18.
- **Push notifications:** After light state commits, the server emits **Server-Sent Events** on:
  - `GET /api/v1/models/{id}/lights/events` — `text/event-stream`, JSON `data:` lines `{ "seq": <uint64>, "deltas": [ { "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }` (`model_id` is implicit from the URL).
  - `GET /api/v1/scenes/{id}/lights/events` — `text/event-stream`, JSON `data:` lines `{ "seq": <uint64>, "deltas": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }`.
  - Clients merge `deltas[]` incrementally per **REQ-041** (`docs/requirements.md`) and `docs/architecture.md` §3.18; resync with `GET …/lights/state`, `GET /api/v1/models/{id}`, or `GET /api/v1/scenes/{id}` only on a sequence gap or `EventSource.onerror`.
- **Long-lived SSE:** If connections drop behind a proxy, set a generous **`HTTP_WRITE_TIMEOUT_SEC`** (or `0` for no write deadline in Go) when debugging; the handler extends write deadlines periodically.

## License

See [LICENSE](LICENSE).
