# WI-02 — Capture light-sequence controller + endpoints (REQ-047)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-01 (`devices.light_count` must exist)
- **Area:** Backend Go (`backend/`)
- **Requirements:** REQ-047 (see `docs/requirements.md`)
- **Architecture:** `docs/architecture.md` §3.1 (`internal/capture`), §3.2 (`/devices/{id}/capture/*`), §3.22, §8.24

## Context

Repo `/workspaces/dlm`, Go module in `backend/`. We need a **built-in capture sweep**: when
started for a device, it turns **one light on at a time** in ascending index order
`0 … n−1`, each for **≈ 1 second**, then off, where `n` is the device's configured
`light_count` (WI-01). It drives the device **directly by LED index** and must work even when
the device is **not assigned** to a model — so it does **not** go through `LightStateStore`
(which is per-model). It runs **server-side** (no browser needed) and on stop/completion all
swept LEDs go off within **2 seconds** (REQ-040 bound).

This is **not** a scene routine (REQ-021/023) — do not involve `python3` or the shape ticker.

Key existing files:
- `backend/internal/devices/pusher.go` — `Pusher` (has `Store`, `Client *http.Client`).
- `backend/internal/devices/wled.go` — `buildWLEDState`, `postJSONState`, `rgbFromLogical` (unexported). WLED `/json/state` with `seg[0].i = [[idx, r,g,b], …]`.
- `backend/internal/store/devices.go` — `GetDevice`, `Device.LightCount` (WI-01).
- `backend/internal/httpapi/router.go` — route registration + `apiDeps` + `NewSiteHandler`.
- `backend/internal/httpapi/json.go` — `writeJSON`, `writeAPIError`.

## Tasks

### 1. Device frame-drive helpers (`backend/internal/devices`)

Add **exported** helpers to drive raw LED frames (the capture sweep needs them; the existing
`Pusher.PushModel` is model-state-based and unsuitable). Suggested in a new
`backend/internal/devices/capture_drive.go`:

- `func (p *Pusher) DriveSingleLED(ctx context.Context, d store.Device, litIdx, n int) error` —
  build a frame of length `n` where only `litIdx` is on (white at full brightness, e.g.
  `#ffffff`, `100`) and all others off, reuse `buildWLEDState`, POST via `postJSONState` to `d.BaseURL`/`d.WLEDPassword`.
- `func (p *Pusher) DriveAllOff(ctx context.Context, d store.Device, n int) error` — all `n` LEDs off.

(You may instead expose a small package function taking `(*http.Client, baseURL, password, …)`;
either way keep the WLED specifics in `internal/devices`.)

### 2. Capture controller (`backend/internal/capture/capture.go`)

Create package `capture` with a thread-safe controller managing **at most one active sweep per device**:

- A `driver` interface (so tests use a fake) with `DriveSingleLED(ctx, d, idx, n) error` and `DriveAllOff(ctx, d, n) error`.
- A `deviceGetter` interface: `GetDevice(ctx, id) (store.Device, error)`.
- Optional `routineChecker` interface: `ModelHasActiveRoutineRun(ctx, modelID string) (bool, error)` — see task 4.
- `type Controller struct { … }` with `New(getter, driver, checker, opts) *Controller`.
- `Start(ctx, deviceID) (Status, error)`:
  - Load device; `ErrDeviceNotFound` propagates.
  - If `LightCount == 0` → `ErrCaptureNoLights`.
  - If a sweep is already running for this device → `ErrCaptureConflict`.
  - If the device is assigned (`ModelID != nil`) and the routine checker reports an active run for that model → `ErrCaptureConflict`.
  - Otherwise launch a goroutine: a `time.Ticker` with period = **dwell** (default **1000 ms**, configurable via `DLM_CAPTURE_DWELL_MS`) iterates `idx = 0 … n−1` calling `DriveSingleLED`; after the last dwell call `DriveAllOff` and set state `idle`. A `stop` channel cancels.
  - Return current `Status` with `state=running, light_count=n, current_index=0`.
- `Stop(deviceID)`: signal the goroutine to stop, call `DriveAllOff`, set `idle`. Guarantee no further drive calls and LEDs off **within 2 s** (REQ-040). Idempotent (stopping an idle device is a no-op success).
- `Status(deviceID) Status`: `{ State string; LightCount int; CurrentIndex int }` where `State ∈ {"idle","running"}`.
- Exported sentinel errors: `ErrCaptureNoLights`, `ErrCaptureConflict` (plus surface `store.ErrDeviceNotFound`).
- On process shutdown, stop all sweeps (expose a `Shutdown()` / honor a context).

### 3. HTTP endpoints (`backend/internal/httpapi/capture.go`) + router

Add handlers and register in `router.go` (literal `capture` segment is safe before `{id}` patterns):

- `POST /devices/{id}/capture/start` → `Start`. `200` `{ "device_id", "state":"running", "light_count", "current_index":0 }`. Map errors: `ErrDeviceNotFound`→`404 not_found`; `ErrCaptureNoLights`→`422 capture_no_lights`; `ErrCaptureConflict`→`409 capture_conflict`.
- `POST /devices/{id}/capture/stop` → `Stop`. `200` `{ "device_id", "state":"idle" }`. `404` unknown device.
- `GET /devices/{id}/capture` → `Status`. `200` `{ "state", "light_count", "current_index"? }`. `404` unknown device.

Wire the controller into `apiDeps` and construct it in `NewSiteHandler` (driver = the `*devices.Pusher` already threaded in; getter = `store`). When `pusher` is nil (API-only construction), the start handler may return a clear `503`/`409` or use a no-op driver — pick one and document in a comment.

### 4. Routine conflict check (best-effort)

Implement `ModelHasActiveRoutineRun(ctx, modelID)` on `*store.Store`: a model is "busy" if it
belongs to a scene that has a `routine_runs` row with `status='running'`. Use existing
`scene_models` and `routine_runs` tables (inspect `backend/internal/store/routines.go` and
`scenes.go`). **If this cross-table query proves complex, ship the `ErrCaptureConflict` guard
for "already running on this device" only**, and leave a `// TODO(REQ-047): routine-run cross-check`
plus pass a `nil` checker — do not block the work item on it.

## Acceptance / tests

- `backend/internal/capture/capture_test.go` with a **fake driver** recording calls:
  - Sweep with `n=3` calls `DriveSingleLED(idx=0,1,2)` in order, then `DriveAllOff` on completion.
  - `Stop` mid-sweep calls `DriveAllOff` and stops further `DriveSingleLED` calls quickly (assert within a generous test bound; you may inject a short dwell via opts).
  - `Start` twice → second returns `ErrCaptureConflict`.
  - `LightCount=0` → `ErrCaptureNoLights`.
- HTTP test (extend or add near `devices_api_test.go`): start `200`, status reflects running, stop `200`; unknown device `404`; zero-count `422`.
- `cd backend && go test ./...` passes.

## Out of scope

- Devices UI (WI-03). Reconstruction (WI-04..WI-08).

## Definition of done

A server-side sweep can be started/stopped per device via the API, lights one LED at a time in
order for ~1 s each, turns everything off on stop/completion within 2 s, enforces one-per-device,
and is covered by unit + HTTP tests.
