# Environment, database, and API tuning

Reference for runtime configuration and integrator-facing API behavior. For the overall system
design and diagrams see [`../design/architecture.md`](../design/architecture.md).

## Environment variables

| Variable | Default | Effect |
|----------|---------|--------|
| `HTTP_LISTEN` | `:8080` | Address/port the server listens on (e.g. `:80`). |
| `DLM_DATA_DIR` | `data` | Root data directory (database + temporary capture uploads). |
| `DLM_DB_PATH` | `data/dlm.db` | SQLite database file path (overrides the default under `DLM_DATA_DIR`). |
| `DLM_PYTHON3` | *(auto)* | Path to the `python3` interpreter for **user** Python routines. |
| `CORS_ALLOWED_ORIGINS` | Next dev origins on `:3000`/`:8000` | Comma-separated allowed origins; `-` disables CORS headers entirely. |
| `HTTP_WRITE_TIMEOUT_SEC` | *(server default)* | HTTP write timeout; increase or set `0` (no deadline) when debugging SSE behind a proxy. |
| `DLM_CV_RUNTIME_DIR` | *(unset)* | Absolute path to a CV runtime bundle (see [`cv-runtime.md`](cv-runtime.md)). |
| `DLM_CAPTURE_DWELL_MS` | `1000` | Milliseconds each light stays on during a device capture sweep. |

Dev-only (Next.js): `DLM_BACKEND_ORIGIN` and `NEXT_PUBLIC_DLM_API_ORIGIN` (see `web/.env.local.example`
and [`build-and-run.md`](build-and-run.md)).

## Embedded database and CSV upload

- **SQLite** is embedded (pure Go, `modernc.org/sqlite`, no cgo, no separate database server). The
  database file is created automatically at **`data/dlm.db`** by default (override with `DLM_DATA_DIR`
  / `DLM_DB_PATH`). Treat it as persistent storage and back it up with normal file backups.
- **CSV upload** for models: `POST /api/v1/models` expects the file field name **`file`** (not `csv`).
  The header must be exactly `id,x,y,z`. Light IDs must be **0-based** sequential integers
  (`0`, `1`, `2`, …).

## High-throughput light updates (integrators)

- Prefer **batch** routes when updating many lights (`PATCH /api/v1/models/{id}/lights/state/batch`
  and the scene bulk routes documented in [`../design/architecture.md`](../design/architecture.md))
  instead of one HTTP request per light. This keeps performance reasonable on small hardware.
- Reuse HTTP keep-alive. For many parallel clients, terminating TLS + HTTP/2 in a reverse proxy while
  proxying HTTP/1.1 to Go is a common pattern (see the architecture doc).
- After commits, the server emits **Server-Sent Events** so clients do not need to poll rapidly:
  - **Model stream:** `GET /api/v1/models/{id}/lights/events` (`text/event-stream`) sends JSON
    `data:` lines shaped like `{ "seq": <uint64>, "deltas": [ { "light_id", "on"?, "color"?, "brightness_pct"? }, ... ] }`; `model_id` is implicit from the URL.
  - **Scene stream:** `GET /api/v1/scenes/{id}/lights/events` sends `{ "seq": <uint64>, "deltas": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, ... ] }`.
- Clients should apply each event’s `deltas[]` incrementally (architecture §3.18). Resync with
  `GET .../lights/state`, `GET /api/v1/models/{id}`, or `GET /api/v1/scenes/{id}` only when the
  sequence number skips or `EventSource.onerror` fires.
- If SSE drops behind a proxy, increase **`HTTP_WRITE_TIMEOUT_SEC`** or set `0` for no write deadline
  while debugging; the handler extends deadlines periodically.
