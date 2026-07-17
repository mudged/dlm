# Backend service: HTTP surface, data, and models (§3.1–§3.14)

This file explains the Go side of dlm — the single binary that serves the JSON HTTP API and the embedded web UI. It covers how the code is organized, every HTTP endpoint, how data is stored in SQLite, how release builds are cross-compiled, and how the built-in sample models and routines are seeded. dlm is a hobbyist app for controlling home LED light installations: users import 3D "models" of where their lights physically sit, group them into "scenes", and run "routines" that animate them.

Part of the [dlm architecture](architecture.md); see the [glossary](glossary.md) for unfamiliar terms.

## 3. Golang service (runtime)

**In plain terms:** The whole backend is one Go program. It exposes a versioned JSON API under `/api/v1/…` and also serves the static web UI. This part of the doc walks through the internal packages, the endpoints, and the database.

A few domain terms used throughout:

- **light state triple** — the three per-light values `(on, color, brightness_pct)` that describe how a single LED currently looks.
- **LightStateStore** — the in-memory, authoritative store of every light's triple while the process runs. It is *not* saved to SQLite (REQ-039).
- **WLED** — popular open-source firmware that runs on an ESP32/ESP8266 and drives addressable LED strips over the network; dlm talks to it over HTTP to physically light your LEDs.

### 3.1 Module and packages

**In plain terms:** The backend is split into small packages with clear jobs: one wires up the server, others handle config, HTTP, the database, the in-memory light state, devices, and the routine engine.

- **Module path:** `example.com/dlm/backend` (or successor); unchanged conceptually.
- **`cmd/server`:** Loads config, constructs the `http.Server`, mounts the API sub-router and the static file server (served from `embed`), and does a graceful shutdown on SIGINT/SIGTERM (the signals sent by Ctrl-C or a service stop).
- **`internal/config`:** Env-based settings — listen address, timeouts, and optional CORS (Cross-Origin Resource Sharing, the browser rule about which origins may call the API). CORS is mainly for dev, when the UI dev server runs on a different origin; in production the UI and API share an origin, so CORS is rarely needed. The SQLite path comes from `DLM_DB_PATH` and/or `DLM_DATA_DIR` (§3.3).
- **`internal/httpapi`:** Middleware (request ID, `slog` structured logging, panic recover, optional CORS) plus the JSON handlers. All errors use one envelope: `{ "error": { "code", "message", "details"? } }`. Handlers for models, light-state (via `internal/lightstate`), scenes, scene-region routes (§3.15), devices (§3.20), and routines / routine-runs (§3.16) delegate to `internal/store`, `internal/lightstate`, `internal/devices`, `internal/wiremodel`, and `internal/routineengine` for start/stop supervision. Routine ticks run in `routineengine`; the HTTP handlers for §3.15 stay the single validation surface, whether called from loopback (the server calling itself on `127.0.0.1`) or an external client (REQ-038 / REQ-039). SSE (Server-Sent Events, a long-lived HTTP stream the server pushes updates on) at `…/lights/events` emit REQ-041 `deltas[]`. See §3.18 (REQ-029, REQ-041).
- **`internal/lightstate`:** The authoritative in-memory per-model maps `light idx → { on, color, brightness_pct }` (REQ-039, §3.9, §3.21). All writes from §3.9 / §3.10 / §3.11 / §3.15 mutate here first, then optionally push to WLED (§3.20). On a real change it emits SSE `data:` lines with a monotonic (always-increasing) `seq` counter and a per-commit `deltas[]` listing only the lights that changed (REQ-041, §3.18).
- **`internal/devices`:** CRUD for the `devices` table (§3.20); a WLED HTTP JSON API client; an optional discovery adapter (mDNS / subnet probe — not required for the MVP, the minimum viable product); and startup / post-assign sync with `LightStateStore` (§3.21).
- **`internal/capture`:** The REQ-047 capture-sweep controller (§3.22). It runs one `time.Ticker`-driven server-side sweep per device that drives the `internal/devices` WLED client to light exactly one LED (`idx` `0 … light_count−1`) for a fixed ≈ 1 s dwell, then off. It does not route through `LightStateStore` (the device may be unassigned). Start/stop status is exposed via §3.2; stop or completion turns all swept LEDs off within REQ-040's 2 s.
- **`internal/reconstruct`:** The REQ-048 / REQ-049 camera-capture orchestration (§3.23). It accepts uploaded video files into a work dir under `DLM_DATA_DIR`, runs an async job that invokes the bundled OpenCV child (`internal/cvruntime` — §3.23.1), parses its JSON result (per-light `x,y,z` plus missing / low-confidence lists), holds the candidate set in memory for review, and on confirm hands a validated `[]wiremodel.Light` to `internal/store` for the normal create transaction (§3.3, REQ-005 / REQ-007). It never persists a model before explicit confirm (REQ-049).
- **`internal/cvruntime`:** The REQ-048 self-contained OpenCV + Python runtime bundle (`//go:embed` compressed, extracted on first use to `DLM_DATA_DIR/runtime/cv/<version>/`, or shipped as a sibling asset in the release archive — §3.23.1) plus the frozen `reconstruct` CV script. It resolves / extracts the interpreter and exposes a `Run(ctx, args) (result, error)` child-process wrapper. No operator `python3` install is required (contrast with §3.17 / REQ-045).
- **`internal/routineengine`:** One supervisor per active `routine_runs` row (§3.16): a `python3` child for `python_scene_script` (§3.17), or a Go `time.Ticker` for `shape_animation` (§3.17.2). It mutates lights via internal calls equivalent to §3.15 (or loopback `127.0.0.1` — implementor's choice) so REQ-021 / REQ-039 semantics still hold. `cmd/server` starts the scheduler on process boot (either resuming `running` rows or marking orphaned runs stopped — document the chosen policy).
- **`internal/wiremodel`:** Parses and validates uploaded CSV (comma-separated values) per §3.6; returns structured errors for HTTP 400 responses (REQ-007).
- **`internal/store`:** The SQLite repository (see §3.3, §3.12, §3.16). It persists models (geometry only for `lights` rows — §3.3), scenes, scene_models, routines, routine_runs, and device metadata. It does not store per-light `on` / `color` / `brightness` (REQ-039). It handles the idempotent seed (§3.8, §3.8.1) and factory reset (§3.14): factory reset MUST `DELETE` all `devices` rows (REQ-017 / REQ-035) before re-seeding.
- **`internal/samples`:** Pure functions that return `[]wiremodel.Light` (sequential `id`s from 0) for the three canonical shapes: on-surface positions, even coverage per §3.8, consecutive dᵢ in band; no I/O (REQ-009).

### 3.2 HTTP surface

**In plain terms:** This is the complete list of endpoints the server exposes. The `Kind` column separates the JSON API from the static UI files. Read the ordering note right after the table — the order routes are registered in matters, because literal paths like `system` must be matched before catch-all `{id}` patterns.

| Kind | Path | Purpose |
|------|------|--------|
| API | `GET /health` | Liveness; **no auth**. |
| API | `GET /api/v1/models` | List models (**metadata** + **light_count**). REQ-006 |
| API | `POST /api/v1/models` | Create model: `multipart/form-data` with `name` (string) + `file` (CSV). REQ-006/007 |
| API | `GET /api/v1/models/{id}` | Model **metadata** + ordered **lights** `{ id, x, y, z, on, color, brightness_pct }[]` (§3.9). REQ-006, REQ-011 |
| API | `DELETE /api/v1/models/{id}` | Delete model (**204** / **404**). **409** if the model is **assigned** to **any** scene (REQ-006 rule 5, §3.13). REQ-006 |
| API | `GET /api/v1/scenes` | List scenes (**id**, **name**, **created_at**, **model_count**). REQ-015 |
| API | `POST /api/v1/scenes` | Create scene: JSON `{ "name", "models": [ { "model_id": "<uuid>" }, … ] }` — **≥ 1** element; array order is placement order (REQ-015 rule 2). Server computes all `offset_x/y/z` (§3.12 create-time algorithm); client MUST not send offsets on create (reject or ignore unknown offset fields per §3.13). **201** + scene summary. REQ-015 |
| API | `GET /api/v1/scenes/{id}` | Scene **metadata** + **placements** + **full** **lights** (positions + state) per model for one round-trip (§3.13). REQ-015 |
| API | `DELETE /api/v1/scenes/{id}` | Delete scene and **all** placements (**204** / **404**). Used after user confirms deleting the whole scene when removing the last model (REQ-015). |
| API | `POST /api/v1/scenes/{id}/models` | **Add** a model to an existing scene: `{ "model_id", "offset_x"?, "offset_y"?, "offset_z"? }` — if offsets omitted, server computes a default "to the right" placement (§3.13). **201** / **400** / **404** / **409** (duplicate model in scene). |
| API | `PATCH /api/v1/scenes/{id}` | **Update** scene-level metadata (currently `margin_m` only; future fields like `name` MAY be added per §3.13). Body `{ "margin_m": <number> }` — must be finite, `0 ≤ margin_m ≤ 5` SI metres; out-of-range, `NaN`, `Inf`, non-numeric, or missing field → `400` `{ "error": { "code": "invalid_margin_m", "message": <human-readable> } }`. **404** if scene id unknown. **200** returns the updated scene summary including `margin_m`. REQ-015 BR 13 / REQ-034 |
| API | `PATCH /api/v1/scenes/{id}/models/{modelId}` | **Update** integer offsets `{ "offset_x", "offset_y", "offset_z" }` (all required or PATCH semantics per implementor — must re-validate containment). **200** with updated placement + optional derived bounds metadata. |
| API | `DELETE /api/v1/scenes/{id}/models/{modelId}` | **Remove** model from scene. If **> 1** models: **204**. If this is the last model: **409** `scene_last_model` — no mutation; message states that confirming will delete the entire scene; client shows modal then calls `DELETE /api/v1/scenes/{id}` (§3.13). |
| API | `GET /api/v1/scenes/{id}/dimensions` | Return scene-space dimensions used for region queries: `origin`, `size`, `max`, and `margin_m` (see §3.15). REQ-020 |
| API | `GET /api/v1/scenes/{id}/lights/events` | **SSE** (`text/event-stream`): JSON `data:` lines `{ "seq": <uint64>, "deltas": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }` after scene-relevant light state commits — `deltas` lists only lights whose triple changed in that commit (REQ-041, §3.18). Legacy `{ "seq" }` only is insufficient for new implementations. |
| API | `GET /api/v1/scenes/{id}/lights` | Return all scene lights as a flattened list in scene coordinates (`sx/sy/sz`) plus model/light identity. REQ-020 |
| API | `POST /api/v1/scenes/{id}/lights/query/cuboid` | Return only lights inside a caller-supplied cuboid in scene space. REQ-020 |
| API | `POST /api/v1/scenes/{id}/lights/query/sphere` | Return only lights inside a caller-supplied sphere in scene space. REQ-020 |
| API | `PATCH /api/v1/scenes/{id}/lights/state/cuboid` | Transactional bulk state update for all lights inside a cuboid in scene space; state semantics match REQ-011. REQ-020 |
| API | `PATCH /api/v1/scenes/{id}/lights/state/sphere` | Transactional bulk state update for all lights inside a sphere in scene space; state semantics match REQ-011. REQ-020 |
| API | `PATCH /api/v1/scenes/{id}/lights/state/scene` | Transactional bulk state update for **every** light currently in the scene (all derived `sx/sy/sz`); body is **only** `on`, `color`, `brightness_pct` (at least one required); **no** geometry object. REQ-020 (uniform whole-scene patch for routines §3.16) |
| API | `PATCH /api/v1/scenes/{id}/lights/state/batch` | Transactional **per-light** updates within one scene: body `{ "updates": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }` — each object identifies one light by `model_id` + `light_id` (idx) and must be a member of the scene at request time else `400`; at least one state field per element; partial merge per row like §3.9; all or nothing in one transaction. REQ-020 (scene-level batch for routines and integrators §3.15) |
| API | `GET /api/v1/routines` | List routine definitions (**id**, **name**, **description**, `type`, `python_source`, `definition_json`, **created_at**). `python_scene_script` rows include `python_source` and `definition_json` null or omitted; `shape_animation` rows include `definition_json` (object — §3.17.2) and `python_source` as empty string or omitted. REQ-021, REQ-022, REQ-033 |
| API | `GET /api/v1/routines/{id}` | **200** one definition (same shape as list row). **404** if unknown. REQ-022, REQ-033 |
| API | `POST /api/v1/routines` | Create definition. Body `name` (required, trimmed non-empty), `description` (string, may be `""`), `type`: `python_scene_script` \| `shape_animation`. If `python_scene_script`: `python_source` string (client MAY substitute `PYTHON_ROUTINE_DEFAULT_SOURCE` per §4.13); `definition_json` omitted or null. If `shape_animation`: `definition_json` required object — server validates per §3.17.2 (reject `400` on schema / range / shape count errors); `python_source` ignored or stored as `""`. Unknown `type` → **400**. **201** + full row. REQ-021, REQ-022, REQ-023, REQ-033 |
| API | `PATCH /api/v1/routines/{id}` | JSON with **≥ 1** field among `name`, `description`, `python_source` (only when `type` is `python_scene_script`), `definition_json` (only when `type` is `shape_animation`). **200** updated row. **404** if missing. **409** `routine_not_editable` for legacy unsupported `type` rows only. **409** `routine_run_active` if a `running` run exists for this `routine_id`. REQ-021, REQ-022, REQ-033 |
| API | `DELETE /api/v1/routines/{id}` | Delete definition. **204** / **404**. **409** `routine_run_active` if a persisted run row exists with `status=running` for this `routine_id` (§3.16). REQ-021 |
| API | `POST /api/v1/scenes/{id}/routines/{routineId}/start` | Start an automated run of `routineId` on scene `id`. **201** `{ "run_id", "scene_id", "routine_id", "status" }`. **404** missing scene or routine; **409** `scene_routine_conflict` if this scene already has any `status=running` row (REQ-021 BR 5 — including when the caller re-posts start for the same routine while it is still running). REQ-021 |
| API | `POST /api/v1/scenes/{id}/routines/runs/{runId}/stop` | Stop run `runId` if it belongs to scene `id`. **200** `{ "run_id", "status": "stopped" }` or **204**. **404** if mismatch or unknown. REQ-021 |
| API | `GET /api/v1/scenes/{id}/routines/runs` | **200**: `{ "runs": [ { "id", "routine_id", "routine_name", "status" } ] }` — only `status=running` rows for this scene (empty array or one element per §3.16 concurrency rule). REQ-021 |
| API | `GET /api/v1/models/{id}/lights/state` | **All** lights' state for the model (ordered by `id`). REQ-011 |
| API | `GET /api/v1/models/{id}/lights/events` | **SSE** (`text/event-stream`): `data:` lines `{ "seq": <uint64>, "deltas": [ { "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }` (`model_id` implicit from URL) after successful light-state writes for this model (REQ-041, §3.18). |
| API | `GET /api/v1/models/{id}/lights/{lightId}/state` | **One** light's state (**404** if model or `lightId` missing). REQ-011 |
| API | `PATCH /api/v1/models/{id}/lights/{lightId}/state` | **Partial** update of `on`, `color`, `brightness_pct` (JSON body; omitted fields unchanged). **200** returns the full updated state object. REQ-011 |
| API | `PATCH /api/v1/models/{id}/lights/state/batch` | **Atomic** partial update of many lights: JSON body `{ "ids": [<int>, …], "on"?, "color"?, "brightness_pct"? }` with at least one of `on`, `color`, `brightness_pct` present; omitted fields unchanged per row. **200** returns `{ "states": [ { "id", "on", "color", "brightness_pct" }, … ] }` sorted by `id`. **400** if `ids` empty, duplicate ids, any id **∉ [0, n−1]**, or merged values invalid. REQ-013 (§3.10). |
| API | `POST /api/v1/models/{id}/lights/state/reset` | REQ-014: no body (or empty JSON object). One transaction sets every light in the model to `on=false`, `color="#ffffff"`, `brightness_pct=100`. **200** returns `{ "states": [ … ] }` for all n lights, `id` ascending. **404** if model missing. |
| API | `POST /api/v1/system/factory-reset` | REQ-017: no body (or `{}`). One transaction per §3.14 (deletes `routine_runs`, `routines`, `devices`, then scenes / models / lights in FK-safe order) then `SeedDefaultSamples` (§3.8) and `SeedDefaultPythonRoutines` (§3.8.1). **200** `{ "ok": true }`. **405** on non-`POST` MUST use the §3.2 JSON envelope `{ "error": { "code": "method_not_allowed", "message": "method not allowed" } }` (not `http.Error` `text/plain`). **500** only on unexpected store failure (no partial wipe visible after commit). Register **before** any `{id}` catch-alls so `system` is not parsed as an id. |
| API | `GET /api/v1/devices` | REQ-035 / REQ-037: List registered devices (see §3.20). |
| API | `POST /api/v1/devices` | REQ-035 / REQ-037: Register device (manual `base_url` + `name` + `type`) — MVP path. `base_url` MUST pass the §3.20 SSRF allowlist (REQ-035 BR 6); violations return `400` `invalid_base_url` (§3.2 envelope). |
| API | `POST /api/v1/devices/discover` | Optional: LAN candidate discovery (§3.20) — may return `501` or omit route until implemented. |
| API | `GET /api/v1/devices/{id}` | Device detail (§3.20). |
| API | `PATCH /api/v1/devices/{id}` | Edit metadata / connection fields (§3.20). When `base_url` is patched, the same REQ-035 BR 6 allowlist applies and rejection returns `400` `invalid_base_url`. |
| API | `DELETE /api/v1/devices/{id}` | Remove device; MUST clear `model_id` assignment in same transaction (REQ-037) then delete row (§3.20). |
| API | `POST /api/v1/devices/{id}/assign` | Body `{ "model_id" }`; REQ-036 enforcement; post-commit §3.21 sync (§3.20). |
| API | `POST /api/v1/devices/{id}/unassign` | Clears `model_id`; physical output policy §3.21. |
| API | `POST /api/v1/devices/{id}/capture/start` | REQ-047: Start the capture light sequence for the device (§3.22). **202** / **200** `{ "device_id", "state":"running", "light_count", "current_index":0 }`. **404** unknown device; **409** `capture_conflict` if a sweep is already running for this device or the device's assigned model is in an active routine run (§3.22); **422** `capture_no_lights` if `light_count = 0`. |
| API | `POST /api/v1/devices/{id}/capture/stop` | REQ-047: Stop the running sweep; all swept LEDs off within REQ-040's 2 s. **200** `{ "device_id", "state":"idle" }` / **204**. **404** unknown device. |
| API | `GET /api/v1/devices/{id}/capture` | REQ-047: Sweep status `{ "state":"idle"|"running", "light_count", "current_index"? }` (poll while recording). **404** unknown device. |
| API | `POST /api/v1/models/capture` | REQ-048 / REQ-049: Create a reconstruction job from `multipart/form-data` with **two or more** `files` (video) + optional `marker` (fiducial dictionary/board id) and optional scale hints. **202** `{ "job_id", "status":"pending" }`. **400** if **< 2** files or unsupported container (§3.23). |
| API | `GET /api/v1/models/capture/{jobId}` | REQ-048 / REQ-049: Job progress / result: `{ "status":"pending"|"running"|"succeeded"|"failed", "progress":0..1, "result"?: { "light_count", "lights":[{"id","x","y","z"}], "missing":[<id>], "low_confidence":[<id>] }, "error"? }` (§3.23). **404** unknown job. |
| API | `POST /api/v1/models/capture/{jobId}/confirm` | REQ-049: Persist the reviewed result as a new model. Body `{ "name" }` (required, trimmed). Server re-validates the candidate lights with the same REQ-005 / REQ-007 rules and creates the model transactionally (§3.3). **201** model summary; **400** invalid name / validation failure; **404** unknown or non-`succeeded` job; **409** duplicate `name` (§3.3). |
| API | `DELETE /api/v1/models/capture/{jobId}` | REQ-049: Cancel / discard the job and delete its uploaded work files. **204** / **404**. |
| API | `GET /api/v1/capture/marker` | REQ-049: Return an optional printable fiducial marker artifact (PDF or PNG) with brief placement guidance; query selects `type` / `size` where offered (§3.23.2). Optional: never required to create a model. |
| API | `/api/v1/*` | Other versioned JSON endpoints as the product grows. |
| Static | `/`, `/*.html`, `/_next/**`, other export assets | Next static export tree from embed; SPA (single-page app) / HTML5 fallback policy: serve `index.html` for unmatched non-API GET if needed (implementor defines exact fallback rules). |

**Ordering (the gotcha):** Register API routes **before** static / `NotFound` handling so `/api/v1` is never swallowed by the UI fallback. Register **literal** path segments `routines`, `start`, `stop`, `factory-reset`, `reset`, `state`, `events`, `batch`, `system`, `devices`, `discover`, `assign`, `unassign` **before** any `{id}` pattern that could swallow them (`lights/events` before `lights/{lightId}`; `POST …/devices/discover` before `GET …/devices/{id}`). Register `PATCH /api/v1/routines/{id}` so `routines` is not matched as an id elsewhere. Register literal `capture`, `start`, `stop`, `confirm`, `marker` segments so `/devices/{id}/capture/*` and `/models/capture`, `/models/capture/{jobId}/confirm` are not swallowed by the `/models/{id}` or `/devices/{id}` patterns (`/models/capture` before `/models/{id}`; `/capture/marker` before any `/capture/{jobId}`).

### 3.3 Persistence (REQ-006)

**In plain terms:** dlm stores its catalog data (models, scenes, routines, devices) in a single embedded SQLite file. Note the deliberate split: geometry (`x,y,z`) lives in SQLite, but per-light on/color/brightness never does — that lives only in memory (REQ-039).

- **Engine:** SQLite accessed from the same Go process (prefer a **pure Go** driver such as `modernc.org/sqlite` to avoid cgo — the mechanism that lets Go call C code — because cgo makes cross-compiling to `linux/arm64` painful).
- **Location:** Configurable path via environment (e.g. `DLM_DB_PATH`), defaulting to a file under an application data directory (e.g. `DLM_DATA_DIR` + `dlm.db`). The DB file is runtime state, not embedded in the binary; creating/opening it at startup satisfies REQ-004 (no extra shipped daemon).
- **Schema (logical):**
  - **`models`:** `id` (TEXT UUID — universally unique identifier — primary key), `name` (TEXT NOT NULL, UNIQUE), `created_at` (TEXT RFC3339 UTC — the standard timestamp format like `2026-07-17T16:24:00Z`).
  - **`lights`:** `model_id` (TEXT FK — foreign key — → `models.id`), `idx` (INTEGER light index), `x`, `y`, `z` (REAL) **only** — no `on`, `color`, or `brightness_pct` columns (REQ-039). Per-light output state lives only in `LightStateStore` (§3.21) for the running process.
  - **`devices`:** `id` (TEXT UUID PK), `type` (TEXT NOT NULL, first value `wled`), `name` (TEXT NOT NULL), `base_url` (TEXT NOT NULL — e.g. `http://192.168.1.50`), optional `bearer` / `password` storage per §3.20, `light_count` (INTEGER NOT NULL DEFAULT 0 — REQ-047: number of addressable lights the device drives, used as the sweep length even when unassigned; `0 ≤ light_count ≤ 1000` consistent with REQ-005; idempotent `ALTER TABLE devices ADD COLUMN light_count INTEGER NOT NULL DEFAULT 0` on startup if missing), `model_id` (TEXT NULL UNIQUE FK → `models.id` ON DELETE SET NULL or RESTRICT — implementor documents unassign semantics (REQ-036)), `created_at` (TEXT RFC3339). At most one non-null `model_id` across rows for a given model (UNIQUE on `model_id` where not null). `light_count` is exposed in `GET` / `POST` / `PATCH /api/v1/devices…` and editable in §4.15.
  - **`scenes`:** `id` TEXT PRIMARY KEY, `name` TEXT NOT NULL UNIQUE, `created_at` TEXT NOT NULL (RFC3339), `margin_m` REAL NOT NULL DEFAULT 0.3 (REQ-015 BR 12 / REQ-034 rule 3 — SI metres, symmetric padding applied on all six faces of the scene-space light AABB — axis-aligned bounding box — when drawing the faint boundary cuboid). Idempotent migration (a schema change applied safely, whether or not it ran before): `ALTER TABLE scenes ADD COLUMN margin_m REAL NOT NULL DEFAULT 0.3` on startup if the column is missing, so legacy rows inherit the same `0.3` default (REQ-015 BR 12 migration fallback).
- **Transactions:** `POST /api/v1/models` MUST parse + validate CSV, then insert `models` + geometry `lights` rows in one transaction (an all-or-nothing DB write). On commit, `LightStateStore` MUST allocate default triples for every `idx` (REQ-014 / REQ-039) in memory (all off, `#ffffff`, `brightness_pct=100`). Rollback on failure clears any in-memory entries created for that attempt (REQ-007).
- **Migration from legacy schemas** that stored `on` / `color` / `brightness_pct` on `lights`: a one-time migration drops those columns (or ignores them) and initializes `LightStateStore` at process start from defaults + §3.21 sync rules (do not reload dropped values as authoritative — REQ-039).
- **Migrations:** Implementor MAY use a minimal migration hook or `CREATE TABLE IF NOT EXISTS` at startup; document schema version if evolved.

**Architectural resolutions (requirements open questions):**

- **Model name uniqueness:** UNIQUE on `models.name` so the list UI stays unambiguous; duplicate names return **409** with a clear message (product may relax later if requirements change).
- **Creation date:** Set by the server at successful create time, stored and exposed as UTC RFC3339 (no client clock trust).
- **Upload naming:** `name` is a required multipart field (non-empty after trim); not inferred from filename alone.
- **Empty model:** Allowed — CSV may contain only the header row (0 lights); still creates a model row with `light_count = 0`.

### 3.4 Build and cross-compile (release artifacts, REQ-043)

**In plain terms:** Every release bakes the same web UI into the binary once, then cross-compiles that one Go program for each target OS/CPU. The canonical targets and exact `go build` commands are below.

**Single embed step (all targets):** Every release build MUST bake the same Next.js static export into `backend/internal/webdist/` (or the package `go:embed` reads from — see §3.5) before invoking `go build`. The canonical sequence matches `npm run release:sync` from `web/` (same contract as `scripts/run.sh` / REQ-008), normally executed once per CI job before cross-compilation so all `GOOS` / `GOARCH` binaries embed identical UI bytes.

**Canonical release targets** (REQ-043) — normative `GOOS` / `GOARCH` triple:

| GOOS | GOARCH | Primary use | Example artifact basename |
|------|--------|-------------|---------------------------|
| `linux` | `arm64` | Raspberry Pi 4 / 5 (64-bit OS), REQ-003 | `dlm_linux_arm64` |
| `linux` | `amd64` | Desktop / server Linux x86_64 | `dlm_linux_amd64` |
| `windows` | `amd64` | Windows 10/11 x86_64 | `dlm_windows_amd64.exe` |

**Optional:** `linux/arm` (ARMv7) MAY be added later for older 32-bit Pi OS installs; it is not part of the REQ-043 MUST set unless requirements change.

**Cross-compile commands** (run from `backend/` after `internal/webdist/` is populated):

```bash
GOOS=linux  GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_linux_arm64 ./cmd/server
GOOS=linux  GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_linux_amd64 ./cmd/server
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_windows_amd64.exe ./cmd/server
```

**Notes:**

- Pure Go SQLite (`modernc.org/sqlite`) avoids cgo so `linux/arm64` builds can run on `ubuntu-latest` runners without an ARM builder for most configurations; if the implementation enables cgo, architecture MUST document Zig / `musl` or native ARM64 runners for Pi artifacts.
- `-trimpath` and `-ldflags="-s -w"` reduce binary size (recommended for Pi SD cards); exact flags are implementor choice.
- **Single file per platform:** each row is one executable per REQ-004; systemd / Windows Service wrappers are OS infrastructure, not a second product binary.

**CI packaging:** Built files SHOULD be uploaded from a `dist/` (or `artifacts/`) directory at repo root with the stable names above so GitHub Releases assets are predictable (§6.8).

### 3.5 Embedding static UI

**In plain terms:** The web UI is compiled to static files and embedded into the Go binary with `//go:embed`, so the single binary serves the whole app with no separate web server.

- **Mechanism:** `//go:embed` on a package (e.g. `internal/webdist`), embedding the full `out/` tree from `next export` (path names preserved: `_next/static/...`).
- **Build contract:** A documented step (Make, `task`, or CI script) runs `npm ci && npm run build` in `web/` (with `output: 'export'`), then syncs `web/out/` → `backend/internal/webdist/` (or `backend/web/build/`).
- **`embed` empty-dir issue:** the repository may keep a small placeholder file under `webdist/` so `go test ./...` works before the first UI bake; release builds replace the directory contents with the real export.

### 3.6 CSV interchange and validation (REQ-005, REQ-007)

**In plain terms:** Users import light positions as a CSV file. This section defines the exact format (header, columns, limits) and the validation rules the Go code enforces on upload.

- **Encoding:** UTF-8 text; parse with Go `encoding/csv`.
- **Delimiter:** Comma (`,`). RFC 4180 quoting rules SHOULD be supported so numeric fields may be quoted.
- **Header row:** Required first record; must be exactly the four fields `id`, `x`, `y`, `z` in that order (case-sensitive, no extra columns). Wrong or missing header → **400** with a format error (matches the acceptance scenario for `idx,x,y,z`).
- **Data rows:** Each row provides one light; `id` must parse as an integer; `x`, `y`, `z` must parse as finite floating-point numbers (reject NaN / Inf).
- **Row count:** After the header, **0 ≤ n ≤ 1000** data rows; n > 1000 → **400** referencing the cap.
- **Id sequence:** Let n be the number of data rows. Sorting rows by `id`, the multiset of ids MUST equal `{0,1,…,n−1}` exactly once each (contiguous from 0, no gaps, no duplicates). Implementor MAY require rows to appear sorted by `id` ascending for simpler validation, or accept any order and validate the set — behavior MUST match tests derived from `docs/requirements/acceptance-criteria.md`.
- **Authoritative checks:** All rules above are enforced in Go on upload; the UI MAY add client-side hints but MUST NOT trust them for security or integrity.

**Chain topology (REQ-005 rule 6):** The CSV does not encode edges explicitly; logical adjacency is only (i, i+1) for i = 0 … n−2. Visualization and any future features MUST not introduce extra adjacencies between non-consecutive ids.

**JSON API shapes (illustrative):**

- **List item:** `{ "id": "<uuid>", "name": "…", "created_at": "<RFC3339>", "light_count": <int> }`
- **Detail:** same metadata plus `"lights": [ { "id": 0, "x": 0, "y": 0, "z": 0, "on": true, "color": "#RRGGBB", "brightness_pct": 100 }, … ]` ordered by `id` ascending (§3.9).

**HTTP errors:** Use the existing `{ "error": { "code", "message" } }` envelope; optional `details` field for row/column hints (REQ-007). **409 Conflict** for duplicate `name`.

### 3.7 Local build-and-run script (REQ-008)

**In plain terms:** `scripts/run.sh` is the one command that builds the UI, embeds it, and starts the Go server. This section describes exactly what it does.

- **Canonical path:** `scripts/run.sh` at the repository root, executable (`chmod +x`), `#!/usr/bin/env bash` with `set -euo pipefail` (or equivalent strictness).
- **Behavior (single invocation):**
  1. Resolve repo root (e.g. `ROOT="$(cd "$(dirname "$0")/.." && pwd)"`).
  2. Under `cwd = $ROOT/web`: run `npm ci` when `node_modules` is missing or `DLM_FORCE_NPM_CI=1`; skip `npm ci` when `DLM_SKIP_NPM_CI=1` or when `node_modules` already exists (see README for trade-offs). Then `npm run release:sync` so `web/out/` is copied to `backend/internal/webdist/dist/`.
  3. Run `go run ./cmd/server` with `cwd = $ROOT/backend` (or `go build -o … && exec ./dlm` if preferred — document in README). Use `exec` for the final Go process so SIGINT reaches the server.
- **Prerequisites:** Node.js + npm on PATH, Go on PATH; network for `npm ci` on cold clone or when forcing a clean install.
- **README.md** MUST show the exact line, e.g. `./scripts/run.sh` or `bash scripts/run.sh`, and note WSL / Git Bash on Windows if applicable.
- **AGENTS.md** already references REQ-008; keep README and this section aligned when the script changes.
- Not a substitute for a Pi production install (still one binary + optional systemd); operators may build off-device and copy `dlm-arm64` per §3.4.

### 3.8 Default sample models (REQ-009)

**In plain terms:** A fresh database seeds three sample models — a sphere, a cube, and a cone — so the app is not empty on first run. This section is the precise recipe: how many lights, how far apart, and how they must cover each shape's surface.

**Units:** SI meters. Each sample returns `[]wiremodel.Light` with sequential `id` 0 … n−1 (this order is the polyline order for REQ-010 segment drawing).

**Counts:** Each of the three samples MUST have **500 ≤ n ≤ 1000** vertices (REQ-009 / REQ-005). Choose a deterministic n per shape (fixed constant or derived from a reproducible rule) so tests and seeds are stable.

**Consecutive spacing:** For every i ∈ {0,…,n−2}, let dᵢ be the Euclidean distance between light i and i+1. Require **0.05 ≤ dᵢ ≤ 0.10** (5–10 cm). Tests SHOULD use a small tolerance (e.g. `1e-3` m) on the bounds to absorb floating-point error; document the chosen ε.

**Surface placement (nominal solids, centered/aligned as today):**

- **Nominal boundary:** the analytic surface of the intended sphere (radius R = 1.0 m, diameter 2 m, center origin), axis-aligned cube (edge 2 m, center origin, faces at ±1 m), or right circular cone (height 2 m, base radius 1 m, base in z = 0 plane, apex at (0,0,2) — or an equivalent consistent pose documented in code).
- **Exterior only:** Each vertex MUST lie on the nominal boundary or in the thin outer shell allowed by requirements: the closest distance from the point to the nominal surface MUST be ≤ 0.03 m, and the point MUST not lie in the interior of the solid (e.g. sphere: ‖p‖ ≥ R − δ with δ tiny for float; cube: not strictly inside the 2×2×2 volume; cone: on or outside the lateral surface + base disk region as defined by the implementor's half-space test).

**Coverage intent (REQ-009 "not edge-only"):**

- **Cube:** Lights MUST be placed on all six exterior square face planes. A wireframe layout (vertices and edges only) is not sufficient. The majority of lights MUST lie in the interior of each face's square (parameterized patch), not only on edges or corners — architecture sets a quota: at least ⌊0.85 n⌋ lights MUST lie strictly in the open face patches (each coordinate on a face is inside the (−1,1)² parameter rectangle for that face, i.e. not on the rim of the square in face-local (u,v)). The remaining lights MAY be used for transitions between faces or to satisfy dᵢ if needed. **Evenness (cube):** Partition n across the six faces with counts differing by at most one (e.g. base q = ⌊n/6⌋, r = n mod 6 faces get q+1 lights). On each face, place that many points on a deterministic quasi-uniform 2D pattern in (u,v) (e.g. regular staggered grid, 2D Halton pairs, or boustrophedon rows) so area coverage is even within the face; document the chosen pattern in code comments.
- **Sphere:** Lights MUST lie on ‖p‖ = R and MUST not read as concentrated on a single narrow strip or one 1D curve only. **Evenness (sphere):** Use a point set known for approximate equal area per point on the sphere (e.g. Fibonacci / golden-angle lattice, HEALPix-style construction, or subdivided icosahedron vertices). After placement, order into the final polyline (below) so REQ-010 still shows a connected path; the underlying set MUST pass tests that no hemisphere (or other fixed cap of bounded area) contains more than a documented fraction of lights (e.g. ≤ 55% of n in any closed hemisphere through the center) — implementor picks caps and thresholds to match acceptance tests.
- **Cone:** Lights MUST cover both the lateral (curved) surface and the flat base disk (z = 0, ρ ≤ r). **Evenness (cone):** Split n between lateral and base in proportion to surface areas (lateral π r ℓ, base π r² with slant height ℓ), rounding with a deterministic rule so the two counts sum to n. On the lateral patch, use a (height, azimuth) quasi-uniform grid or spiral; on the base, use a polar or 2D quasi-uniform pattern in the disk. The same ≥ 85% rule as the cube applies within each part: at least ⌊0.85 n_lat⌋ lateral lights MUST lie in the interior of the lateral patch (not only on the rim or apex), and at least ⌊0.85 n_base⌋ base lights in the interior of the disk (ρ strictly between 0 and r), unless n for that part is too small — in which case document the degenerate exception in tests.

**Ordering vs spacing (two-step design):**

Requirements demand both even 2D/area placement and consecutive dᵢ in [0.05, 0.10]. Architecturally, treat this as two steps:

1. **Target positions:** Produce an unordered (or weakly ordered) multiset of on-surface points meeting the evenness and per-face/per-part quotas above.
2. **Path construction:** Build a single open polyline P₀,…,P_{n−1} through exactly those points (or through refined points after subdivision) such that every dᵢ lies in the band. Allowed techniques: (a) generate points along a space-filling or serpentine path on the unfolded or parameterized surface so consecutive samples are naturally ~0.075 m apart; (b) sort / chain with greedy nearest-neighbor from a fixed seed and then split long edges / merge short jumps by inserting or removing intermediates on the same surface; (c) walk face patches in a fixed order (cube) with boustrophedon rows at ~0.075 m step, using short surface segments across edges only where needed. Subdivide any segment longer than 0.10 m with collinear on-surface inserts; avoid interior shortcuts that leave the boundary.

**Spacing algorithm (implementor):** After the polyline exists, verify all dᵢ; adjust with subdivision or local reordering while preserving evenness tests. If n falls outside [500,1000], retune the step size or pattern density and regenerate — do not relax dᵢ or surface rules.

**Architectural resolutions (requirements open questions):**

- **Open vs closed path:** Open polyline is the default; REQ-010 draws segments (i, i+1) only, so no automatic (n−1,0) unless requirements change.
- **User deletes samples:** Seed only when `SELECT COUNT(*) FROM models` is 0 immediately after migrations on process startup. Deleting a subset does not re-seed during that run; if the user deletes all models, the next process start seeds again.
- **Fixed English names:** `Sample sphere`, `Sample cube`, `Sample cone`.

**Per-shape summary (`internal/samples`):**

| Shape | Characteristic size | Coverage + ordering |
|-------|---------------------|---------------------|
| **Sphere** | R = 1 m | Area-even point set on ‖p‖ = R; ordered into a polyline with dᵢ ∈ [0.05,0.10]; n ∈ [500,1000]; hemisphere / cap tests to block single-strip dominance. |
| **Cube** | Edge 2 m | Six faces, ~n/6 lights per face (±1); interior (u,v) patterns; open polyline visiting faces with dᵢ in band — not an edge-only walk. |
| **Cone** | h = 2 m, r = 1 m | Area-split between lateral and base; quasi-uniform on each; open polyline with dᵢ in band covering both parts. |

**Integration:**

- After `store.Open` + `migrate`, if the model count is 0, `store.SeedDefaultSamples(ctx)` runs one transaction inserting three models and lights via `samples.SphereLights()`, `samples.CubeLights()`, `samples.ConeLights()`.
- Samples use the same persistence path as user CSV so REQ-005 invariants hold; `LightStateStore` MUST receive default triples per §3.9 for each inserted `idx` after the seed transaction commits.

### 3.8.1 Default sample Python routines (REQ-032, REQ-017)

**In plain terms:** A fresh database (or one after factory reset) also seeds three ready-made Python routines. To avoid drift, the same source text feeds both the editor UI and the Go seed; pick one syncing approach and document it.

**Purpose:** When the `routines` table has zero rows (a fresh DB after migrations) or after a factory reset (§3.14), insert exactly three `python_scene_script` definitions whose `python_source` text implements the normative behaviors in §3.17.1 (growing sphere, sweeping cuboid, random colour cycle — all scene lights).

**Single source of truth:** `web/lib/pythonRoutineSamples.ts` exports three string constants (e.g. `PYTHON_SAMPLE_GROWING_SPHERE_SOURCE`, `PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE`, `PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE` — exact export names are implementor choice if documented). The Go binary MUST receive the same UTF-8 text at seed time (no accidental drift from the editor experience) via one of: (a) `//go:generate` or a release script that copies or embeds from the TS module into `internal/seed` / `embed`; (b) hand-maintained mirrored `.go` string literals with a CI check that diffs against the TS exports; (c) a JSON bundle written during `npm run release:sync`. Pick one approach and document it in `README.md`.

**Store API:** `SeedDefaultPythonRoutines(ctx, tx)` (or equivalent): `INSERT` three rows with `type = 'python_scene_script'`, distinct `name` / `description` recognizable per REQ-032, `python_source` set to each sample string, `description` may be `""`.

**Ordering:** Call after `SeedDefaultSamples` in `Open` when `SELECT COUNT(*) FROM routines` = 0 (same process-start transaction or an adjacent transaction — avoid a partial seed without documented recovery).

**User deletes defaults:** No automatic re-seed until the next factory reset or a fresh DB (REQ-032).

**Tests (REQ-032):** Integration or store tests assert that after a fresh `Open` (empty `routines` table), `GET /api/v1/routines` returns exactly three definitions with `type` `python_scene_script`, distinct recognizable names, and non-empty `python_source` that parses (e.g. `ast.parse` in CI optional).

**Tests (REQ-009, §3.8 samples only):** `internal/samples`: assert 500 ≤ n ≤ 1000; 0.05 − ε ≤ dᵢ ≤ 0.10 + ε for all consecutive pairs; surface predicates (distance-to-nominal-surface ≤ 0.03, not interior) per shape; characteristic ~2 m extent within documented tolerance; cube: each of the six faces has the expected light count (±1 of ⌊n/6⌋ / ⌈n/6⌉) and ≥ ⌊0.85 n⌋ lights in open face interiors; sphere: cap / hemisphere bound per §3.8; cone: both lateral and base receive lights per area split, with interior quotas analogous to the cube where n per part allows. Integration: fresh DB list shows three expected model names.

### 3.9 Per-light state API and authoritative memory (REQ-011, REQ-012, REQ-031, REQ-039)

**In plain terms:** A light's geometry (`x,y,z`) is in SQLite, but its state triple (`on, color, brightness_pct`) lives only in the in-memory `LightStateStore`. Read/write endpoints merge partial updates onto the current triple and, on real change, push an SSE event and sync any assigned WLED device.

**Storage split:** SQLite `lights` rows hold only `x,y,z` (§3.3). The operational triples `(on, color, brightness_pct)` are authoritative in `LightStateStore` (§3.21) for the lifetime of the process; they MUST not be written to SQLite for durability (REQ-039).

**Canonical JSON field names (API):**

| Field | Type | Rules |
|-------|------|--------|
| `on` | boolean | `true` = lit appearance (REQ-012); `false` = off (REQ-012 `#D0D0D0` at 85% transparency in three.js, §4.7). |
| `color` | string | `#` + 6 hex digits `[0-9A-Fa-f]` (normalize lowercase on output). |
| `brightness_pct` | number | [0, 100]; when `on` is false, the value is stored in memory but does not produce a lit appearance. |

**Default state (REQ-014):** Whenever a model gains n lights (create or seed), `LightStateStore` MUST contain exactly n entries `idx` 0…n−1 with `on=false`, `color="#ffffff"`, `brightness_pct=100`.

**Endpoints (unchanged URLs — see §3.2):** Handlers read / merge / write `LightStateStore`; after any successful mutation that changes the effective triple for any light (§3.19), emit an SSE `seq` (§3.18) and invoke `devices.SyncModelLights` (§3.20) for models with an assigned WLED device (REQ-038).

`GET /api/v1/models/{id}` and `GET /api/v1/scenes/{id}` compose responses by joining geometry from SQLite with triples from `LightStateStore`.

**REQ-031:** Compare the merged vs current in-memory triple before recording a change; skip device I/O when they are equivalent (§3.19).

**Upload limits:** Unchanged from prior text (small `PATCH`, larger batch cap §3.10).

### 3.10 Batch light state update (REQ-013, extends REQ-011)

**In plain terms:** One request can update many lights of a model at once. The server merges the fields onto each light's current triple atomically in memory (no SQLite writes for state), then fans out SSE and WLED pushes only for lights that actually changed.

**Rationale:** One round-trip applies many ids; the server updates `LightStateStore` atomically (single mutex / per-model lock — document in code) without SQLite writes for state (REQ-039).

**Endpoint:** `PATCH /api/v1/models/{id}/lights/state/batch`

**Request body (JSON):**

| Field | Type | Rules |
|-------|------|--------|
| `ids` | array of int | Required; non-empty; no duplicates; each id MUST satisfy 0 ≤ id ≤ n−1 for the model's light count n (REQ-005). |
| `on` | boolean | Optional; if present, applied to every listed light. |
| `color` | string | Optional; if present, canonical `#RRGGBB` for all listed lights (same validation as §3.9). |
| `brightness_pct` | number | Optional; if present, in [0, 100] for all listed lights. |

**Rule:** At least one of `on`, `color`, `brightness_pct` MUST appear in the body (otherwise **400** — an empty patch object is rejected to catch client bugs).

**REQ-031 note:** When every listed light already matches the merged triple, return **200** with `states[]` and perform no memory writes / no device push (optional `unchanged_all: true`) — §3.19.

**Behavior:**

1. Validate the model exists (**404** if not).
2. Validate `ids` (non-empty, unique, in range); on failure **400** with `error.message` naming out-of-range or duplicate ids where practical.
3. For each id, merge the provided fields onto the current in-memory triple (same semantics as a single `PATCH`).
4. Reject with **400** if a merged `color` or `brightness_pct` would be invalid (same rules as §3.9).
5. Commit all changes under the `LightStateStore` lock; then fan out SSE and WLED (§3.20) for actually changed indices only (§3.19); **200** with `{ "states": [ … ] }`.

**Implementation:** `internal/lightstate.BatchPatch` (or equivalent); handlers in `internal/httpapi`.

**Client:** After **200**, merge `states` into the in-memory `lights` array used by §4.7 and the paginated table (§4.8) in one React `setState` (or equivalent) so REQ-012 timeliness holds for bulk applies (no indefinite staleness).

**Compatibility:** The single-light `PATCH …/lights/{lightId}/state` remains available for integrators and simple UI paths (REQ-011).

### 3.11 Reset all lights to default state (REQ-014)

**In plain terms:** One POST returns every light in a model to its default look (off, white, 100%). It only touches the in-memory store, not SQLite.

**Endpoint:** `POST /api/v1/models/{id}/lights/state/reset`

**Request:** No JSON body required (`Content-Length: 0` or `{}` accepted and ignored).

**Behavior:**

1. **404** if the model does not exist.
2. Under the `LightStateStore` lock, set every `idx` 0…n−1 to `on=false`, `color="#ffffff"`, `brightness_pct=100` (REQ-014).
3. **200** with `{ "states": [ … ] }`; SSE + device sync per §3.9 (may no-op if already default — §3.19).

**Implementation:** `internal/lightstate.ResetAll`; not an `internal/store` SQL `UPDATE` for state fields.

**Client:** The model detail page exposes a Reset (or Reset lights) button (§4.6). On **200**, merge `states` into `lights` and refresh §4.7 + §4.8 in one update (same timeliness expectation as §8.7).

**Integrators:** Idempotent aside from touching `updated` semantics if added later; safe to call repeatedly.

### 3.12 Scene persistence and containment (REQ-015)

**In plain terms:** A scene groups one or more models, each shifted by an integer offset. Canonical light coordinates (`x,y,z`) never change; the scene view uses derived coordinates `sx = x + ox`, etc. All lights must stay in the non-negative octant, and the server computes offsets automatically.

**Tables (logical, SQLite):**

| Table | Columns | Notes |
|-------|---------|--------|
| `scenes` | `id` (TEXT UUID PK), `name` (TEXT NOT NULL, UNIQUE), `created_at` (TEXT RFC3339 UTC) | One row per scene. No row without ≥ 1 `scene_models` row after a successful create (transaction). |
| `scene_models` | `scene_id` (TEXT FK → `scenes.id` ON DELETE CASCADE), `model_id` (TEXT FK → `models.id` ON DELETE RESTRICT or check before model delete), `offset_x`, `offset_y`, `offset_z` (INTEGER NOT NULL) | PK `(scene_id, model_id)` — each model at most once per scene. Offsets are signed integers interpreted as whole SI meters added to each light's `x`, `y`, `z` (float64) from `lights`. The API accepts only non-negative offsets; invalid combinations fail the containment checks below. |

**Referential integrity:** Deleting a model MUST fail with **409** if any `scene_models` row references `model_id` (REQ-006 rule 5). Use `ON DELETE RESTRICT` on `scene_models.model_id` and/or an explicit pre-check in the handler so the JSON error matches `model_in_scenes` (§3.13).

**Scene-space position:** For light `L` with model-space `(x, y, z)` and placement `(ox, oy, oz)`:

`sx = x + ox`, `sy = y + oy`, `sz = z + oz` (all float64 arithmetic).

**Canonical vs derived (REQ-015 rule 4):** `lights` rows and `GET /api/v1/models/{id}` always expose the stored `x`, `y`, `z` (REQ-005) — never rewritten when a model joins or leaves a scene or when `scene_models` offsets change. Only `scene_models` stores `offset_*`. `GET /api/v1/scenes/{id}` and the scene three.js view use derived `sx`, `sy`, `sz` (plus canonical `x`, `y`, `z` on each light object for labels / comparison).

**Containment (authoritative in Go):** For every light in every model in the scene, `sx`, `sy`, `sz` MUST be ≥ 0 (treat tiny negative float noise < 1e-9 as 0 if needed — document ε). Violations → **400** on create / add / patch with `error.message` naming the model and axis where possible.

**Automatic bounds / margin for rendering:** Let `Mmax_x`, `Mmax_y`, `Mmax_z` be the maxima of `sx`, `sy`, `sz` respectively over all lights in the scene. The visual / framing AABB (axis-aligned bounding box) uses origin `(0,0,0)` and upper corner `(Mmax_x + 1, Mmax_y + 1, Mmax_z + 1)` meters (≥ 1 m padding beyond the tight max per axis; mins at 0 per REQ-015). Camera framing (§4.9) uses this box (+ sphere radius for 2 cm markers).

**Canonical "right" axis (+X):** three.js's default camera looks toward −Z with +Y up; "to the right" of the existing layout means increasing scene `+X`. Document in UI copy if needed for the user mental model.

**Default placement for `POST …/scenes/{id}/models` (omit offsets):**

1. Compute `M_x` = the maximum `sx` over all lights already in the scene (if no lights, `M_x = 0`).
2. Gap: `gap_m = 0` meters abutting (architecture default; implementor MAY use a small positive constant e.g. `0.1` if tests prefer visual separation — document).
3. For the incoming model, load all lights; let `m_min_x` = min(x), `m_min_y` = min(y), `m_min_z` = min(z) over its lights.
4. `offset_x` = the smallest integer ≥ `M_x + gap_m - m_min_x` (ceil to int). `offset_y` = the smallest integer ≥ `−m_min_y` so that `y + offset_y ≥ 0` for all lights (typically `max(0, ⌈−m_min_y⌉)` if all `y` share an offset). `offset_z` analogous for `z`. Re-run the full containment check after rounding.

**Create-time placement for `POST /api/v1/scenes` (REQ-015 rule 2):** The input is an ordered list `[model_id₁, …, model_idₙ]` (JSON array order). Store implementation `CreateScene(ctx, name, modelIDs []string)` (or equivalent):

1. Validate unique ids, existence, `len ≥ 1`, scene name.
2. `placed := []scenePlacementLights{}` (empty accumulator).
3. For each `mid` in `modelIDs` order:
   - Load `Detail(mid)` lights.
   - If `len(placed) == 0`: first model — compute minimal non-negative integers `ox`, `oy`, `oz` so every light has `x+ox ≥ 0`, `y+oy ≥ 0`, `z+oz ≥ 0` (e.g. `ox = max(0, ⌈−min_x(lights) − ε⌉)`, same for `y`, `z`). This anchors the first footprint in the non-negative octant without user input.
   - Else: `ox`, `oy`, `oz` ← `DefaultOffsetsForNewModel(placed, lights)` (same routine as `POST …/scenes/{id}/models` with omitted offsets).
   - Append `{model_id: mid, offsetX/Y/Z: ox/oy/oz, lights}` to `placed`; run `validateScenePlacements(placed)`; on failure abort the whole transaction with **400**.
4. BEGIN; insert `scenes` + all `scene_models` rows; COMMIT.

**Optional REQ-015 advanced override (not MVP):** If the product later allows client-supplied offsets on create, document a separate flag or endpoint; the default path remains fully automatic.

**Scene name uniqueness:** Reuse the same policy as models: UNIQUE on `scenes.name`; duplicate → **409**.

### 3.13 Scenes HTTP API behavior (REQ-015)

**In plain terms:** This section pins down the request/response bodies and status codes for every scene endpoint, including the "delete the last model" flow (409 `scene_last_model`) and the `margin_m` PATCH rules.

**`POST /api/v1/scenes`** — Body `{ "name": string, "models": [ { "model_id": "<uuid>" }, … ] }`. Each object SHOULD contain only `model_id`; if `offset_x`, `offset_y`, `offset_z` appear, return `400` `validation_failed` (do not accept client offsets on create — REQ-015 rule 2) or strict JSON schema rejection. Validate: `name` non-empty trimmed; `models.length ≥ 1`; no duplicate `model_id`; every `model_id` exists. Server runs the §3.12 create-time placement algorithm (no offsets from client); then a transaction inserts `scenes` + `scene_models`. Response `201`: `{ "id", "name", "created_at", "model_count" }`. Canonical `lights` rows are never updated by this handler.

**`GET /api/v1/scenes/{id}`** — `200`: `{ "id", "name", "created_at", "items": [ { "model_id", "name", "offset_x", "offset_y", "offset_z", "lights": [ { "id", "x", "y", "z", "sx", "sy", "sz", "on", "color", "brightness_pct" }, … ] } ] }`. `x`, `y`, `z` MUST match the persisted `lights` (canonical REQ-005); `sx`, `sy`, `sz` MUST equal canonical + offsets (derived). Scene tooltips / labels SHOULD show both scene and model-local coordinates where space allows (§4.9). `PATCH /api/v1/scenes/{id}/models/{modelId}` updates only `scene_models` offsets — never `lights`.

**`POST /api/v1/scenes/{id}/models`** — Add one model. Optional offsets → default §3.12. **409** if `model_id` already in the scene.

**`PATCH /api/v1/scenes/{id}/models/{modelId}`** — Update offsets; re-validate full scene containment.

**`DELETE /api/v1/scenes/{id}/models/{modelId}`** — If the remaining model count after delete ≥ 1: delete the row, **204**. If this is the only model: **409** `{ "error": { "code": "scene_last_model", "message": "…", "details": { "scene_id": "…" } } }` — no row deleted. The client shows a confirm dialog then calls `DELETE /api/v1/scenes/{id}`.

**`DELETE /api/v1/models/{id}`** — Before delete, `SELECT` from `scene_models`. If any row: `409` `{ "error": { "code": "model_in_scenes", "message": "…", "details": { "scenes": [ { "id", "name" }, … ] } } }` (REQ-006 rule 5).

**`PATCH /api/v1/scenes/{id}`** — Update scene-level metadata (REQ-015 BR 13 / REQ-034).

- **Body (JSON):** `{ "margin_m": <number> }`. Strict schema: `margin_m` is required; unknown fields → `400` `bad_request` (`DisallowUnknownFields`). Future fields (e.g. `name`) MAY be added here with their own validation — do not break existing callers.
- **Validation of `margin_m`:** finite `float64` (reject `NaN` / `±Inf`), `0 ≤ margin_m ≤ 5` SI metres. Out-of-range or non-finite → `400` `{ "error": { "code": "invalid_margin_m", "message": <human-readable>, "details": { "min": 0, "max": 5 } } }`.
- **Persistence:** A single `UPDATE scenes SET margin_m = ? WHERE id = ?` outside a multi-statement transaction (no cross-table writes). Light state (REQ-011 / REQ-039), canonical light rows (REQ-005), and placements (`scene_models`) MUST NOT be mutated by this handler.
- **Response 200:** `{ "id", "name", "created_at", "model_count", "margin_m" }` so the client can mirror the persisted value without a follow-up `GET`. **404** `{ "error": { "code": "not_found" } }` if the scene id does not exist.
- **Push (REQ-041):** `margin_m` changes affect the viewport geometry but not per-light `on` / `color` / `brightness_pct` — do NOT emit a `/lights/events` delta for this PATCH. Clients that edit `margin_m` MUST echo the server response (or re-fetch `GET …/scenes/{id}` / `…/dimensions`) before the next `requestAnimationFrame` paint so the `SceneLightsCanvas` boundary wire rebuilds (§4.7 / §4.9).
- `GET /api/v1/scenes/{id}` and `GET /api/v1/scenes` (list) also include `margin_m` in their JSON payloads so clients do not need a second round-trip to render the boundary cuboid on initial load.

**Light state in scene view:** Same persistence as model detail; the composite view uses the `GET /api/v1/scenes/{id}` payload or parallel `PATCH` to `/models/{id}/lights/…` after picking a model + light (implementor chooses the simplest option consistent with REQ-012 timeliness). Recommended: the scene detail response includes current state; a per-light PATCH from the scene page reuses the existing endpoints by `model_id` + `lightId` then merges into local state for that model's group in three.js.

### 3.14 Factory reset (REQ-017)

**In plain terms:** One endpoint wipes all catalog data and devices, then re-seeds the three sample models and three sample routines — the same state as a brand-new install. It runs as one transaction in a foreign-key-safe delete order, then rebuilds the in-memory light state.

**Purpose:** Atomically return persisted catalog data (model geometry, scenes, routines) and the device registry to the same logical outcome as first startup (REQ-009 + §3.8.1) with no registered devices (REQ-017 / REQ-035 / REQ-036); rebuild in-memory light state after commit (REQ-039).

**Store API (conceptual):** `FactoryReset(ctx) error` in `internal/store` + `LightStateStore.LoadLightStateFromDB()` (or an equivalent full rebuild) after commit:

1. `BEGIN IMMEDIATE`; (the shipped build has no Go routine supervisor — clients with open tabs SHOULD receive errors on their next `fetch` after `routine_runs` are deleted).
2. `DELETE FROM routine_runs`.
3. `DELETE FROM routines`.
4. `DELETE FROM scene_models`.
5. `DELETE FROM scenes`.
6. `DELETE FROM devices` (REQUIRED — REQ-017 / REQ-035).
7. `DELETE FROM lights`.
8. `DELETE FROM models`.
9. `SeedDefaultSamples` + `SeedDefaultPythonRoutines` (§3.8 / §3.8.1) in the same transaction.
10. `COMMIT`.
11. `LightStateStore`: drop all prior maps; rebuild from current models (defaults per REQ-014 for each model) (REQ-017 / REQ-039).

**Invariants:** After **200**, `GET /api/v1/models` lists only the three samples; `GET /api/v1/scenes` is empty; `GET /api/v1/routines` lists exactly the three default Python sample routines; `GET /api/v1/devices` is empty.

**Security / abuse:** The endpoint is unauthenticated (same as the rest of the MVP API). Operator documentation should state that any client that can reach the server can invoke factory reset; future auth may protect this route first.

**HTTP:** `POST /api/v1/system/factory-reset` — empty body or `{}` accepted; **200** `{ "ok": true }`. Idempotent in effect: repeated calls still yield three deterministic samples per §3.8.
