# WI-06 — Reconstruction jobs + `/models/capture*` API (REQ-048, REQ-049)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-04 (`cvruntime.Run` + JSON contract), WI-05 (the script the runtime ships). You can develop against a **fake CV runner** without WI-05 finished.
- **Area:** Backend Go (`backend/internal/reconstruct`, `backend/internal/httpapi`)
- **Requirements:** REQ-048, REQ-049 (esp. review-then-confirm), REQ-005/REQ-007 (output validation)
- **Architecture:** `docs/architecture.md` §3.1, §3.2 (`/models/capture*`), §3.23, §8.25

## Context

Turn uploaded videos into a model. This item builds the **async job** orchestration and HTTP API
that sits between the UI (WI-08) and the CV runtime (WI-04/WI-05). A model is **persisted only on
explicit confirm** after the user reviews the detected lights.

Existing relevant files:
- `backend/internal/httpapi/models.go` — `createModel` shows the multipart + `store.Create(ctx, name, []wiremodel.Light)` pattern and the error envelope.
- `backend/internal/wiremodel/csv.go` — `ParseLightsCSV`, `ParseError`, and the REQ-005/007 validation rules; `wiremodel.Light` type.
- `backend/internal/store/store.go` — `Create`, `ErrDuplicateName`.
- `backend/internal/cvruntime` (WI-04) — `Run(ctx, JobSpec) (Result, error)` and the contract types.
- `backend/internal/httpapi/router.go` — register routes; `apiDeps`; `NewSiteHandler`.
- `backend/internal/config/config.go` — for a work-dir base (use `DLM_DATA_DIR`).

## Tasks

### 1. `internal/reconstruct` package

- A **`CVRunner` interface** wrapping `cvruntime.Run` so tests inject a fake.
- An in-memory **job store**: `map[jobID]*Job` guarded by a mutex. `Job` holds `Status`
  (`pending|running|succeeded|failed`), `Progress` (0..1), `Result` (candidate lights + `missing` +
  `low_confidence`), `Error`, the work-dir path, and a created timestamp. **No SQLite table** — nothing
  is persisted until confirm (a job lost to restart is acceptable; document it).
- `Manager`:
  - `Create(ctx, files [][]byte or io.Readers, params) (jobID, error)` — write uploads to a per-job
    work dir under `DLM_DATA_DIR/runtime/capture/<jobID>/`, enqueue a worker (one job at a time by
    default on the Pi). Reject `< 2` files.
  - worker runs `CVRunner.Run`, updates status/progress, stores `Result`.
  - `Get(jobID) (*Job, bool)`.
  - `Confirm(ctx, jobID, name) (store.Summary, error)` — convert the candidate `Result.lights` to
    `[]wiremodel.Light`, **re-validate with the same REQ-005/REQ-007 rules**, then `store.Create`.
    Reuse validation: either build CSV bytes and call `wiremodel.ParseLightsCSV` (guarantees identical
    rules) **or** extract a shared `wiremodel.ValidateLights([]Light) error` and call it from both paths.
    Only `succeeded` jobs are confirmable.
  - `Discard(jobID)` — cancel if running, delete the work dir, drop the job.
  - Startup/janitor: remove stale work dirs on boot and on terminal states.

### 2. HTTP endpoints (`backend/internal/httpapi/capture_models.go`) + router

Register in `router.go`. **Order matters**: register `POST /models/capture`,
`GET /models/capture/{jobId}`, `POST /models/capture/{jobId}/confirm`,
`DELETE /models/capture/{jobId}` **before** the existing `/models/{id}` patterns so `capture` is not
parsed as a model id.

- `POST /models/capture` — `multipart/form-data` with **≥ 2** `files` (video) + optional
  `marker`/`scale_hint`. Stream with `http.MaxBytesReader` using a **large** capture limit (new
  const, e.g. a few hundred MB — *not* the 1 MiB CSV limit). Validate file count and an **allowed
  container/extension** allowlist (e.g. `.mp4`, `.mov`, `.mkv`, `.webm` — document). `202
  { job_id, status:"pending" }`; `400` on `<2` files / bad container / oversize.
- `GET /models/capture/{jobId}` — `{ status, progress, result?, error? }`. `404` unknown.
- `POST /models/capture/{jobId}/confirm` — body `{ name }`. On success `201` model summary. Errors:
  `400 validation_failed` (REQ-005/007 / empty name), `404` unknown or non-`succeeded` job,
  `409 conflict` duplicate name (mirror `createModel`).
- `DELETE /models/capture/{jobId}` — `204` / `404`.

Wire the `reconstruct.Manager` into `apiDeps`/`NewSiteHandler` (inject a `CVRunner` backed by
`cvruntime`; allow a nil/fake for API-only construction in tests).

### 3. Security / limits (see architecture §9)

- Enforce per-request and per-file size limits; reject early. Stream to disk (do **not** buffer whole
  videos in memory). Clean up work dirs on terminal states and on startup. Treat uploads as untrusted
  input to the child (the child has its own timeout via WI-04).

## Acceptance / tests

- `internal/reconstruct` tests with a **fake `CVRunner`**:
  - happy path: create job (2 fake files) → worker → `succeeded` with a candidate result; `Confirm`
    persists a model (assert via the store) whose lights match; `Get` reflects status transitions.
  - `<2` files rejected; confirm on a non-`succeeded` job → error; invalid candidate (e.g. non-sequential
    ids) → confirm returns a validation error (not a panic); `Discard` removes the work dir.
- HTTP tests: `POST /models/capture` with 2 small fake files → `202`; status endpoint; confirm →
  `201`; duplicate name → `409`; `<2` files → `400`; unknown job → `404`.
- `cd backend && go test ./...` passes (no real CV runtime needed — use the fake runner; gate any
  real-runtime integration test behind `DLM_CV_RUNTIME_DIR`).

## Out of scope

- The CV algorithm (WI-05) and runtime packaging (WI-04). The marker endpoint (WI-07). UI (WI-08).

## Definition of done

Uploading ≥ 2 videos starts an async job; clients can poll status/result; confirming persists a
validated model and cancelling discards it; covered by unit + HTTP tests with a fake CV runner.
