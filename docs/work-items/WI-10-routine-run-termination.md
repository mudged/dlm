# WI-10 — Routine engine: persist run termination on exit & init failure

- **Model:** Medium (Sonnet)
- **Depends on:** none
- **Area:** Backend Go (`backend/internal/routineengine`, `backend/internal/store`)
- **Source:** Code review findings (High). Repo root `/workspaces/dlm`.

## Context

The routine engine starts two kinds of routines on a scene: a **Python user script**
(`startPython`) and a built-in **shape animation** (`startShape`). When a routine starts, the
HTTP layer (`postSceneRoutineStart`) has already committed a `routine_runs` row with status
`running` and the engine records a cancel func in its in-memory `active` map.

Two paths leave that DB row stuck as `running` forever:

1. **Python process exit** — `backend/internal/routineengine/engine.go` (~lines 107–140). When the
   child process exits (success, error, or crash), the goroutine only logs and `defer e.unregister(runID)`s.
   It never calls `store.StopRoutineRun`. The `routine_runs` row stays `running`.
2. **Shape init failure** — `engine.go` (~lines 148–162). If `GetSceneDimensions` or
   `ParseAndInit` fails, the engine `cancel()`s and `unregister`s but again never marks the run
   stopped in SQLite.

Consequence: the UI/API believe a routine is still active, and any new start on that scene returns
`409 scene_routine_conflict` permanently — until the process is restarted (boot recovery via
`StopAllRunningRoutineRuns` clears it). This is a correctness/availability bug.

Key existing files:
- `backend/internal/routineengine/engine.go` — `startPython`, `startShape`, `registerCancel`, `unregister`, the shape ticker goroutine.
- `backend/internal/store/store.go` — `StartRoutineRun`, `StopRoutineRun`, `StopAllRunningRoutineRuns` (confirm exact names/signatures before using).
- `backend/internal/httpapi/*scene*routine*` — `postSceneRoutineStart` / stop handlers and the conflict code `scene_routine_conflict`.
- Tests near the engine package and store routine-run tests.

## Tasks

1. **On Python process termination** (`startPython` goroutine): after `cmd.Run()` returns, always
   mark the run stopped in the store (e.g. `store.StopRoutineRun(ctx, runID)`), regardless of whether
   the exit was clean, errored, or context-cancelled. Use a context that is **not** the cancelled
   per-run `ctx` for the DB write (e.g. `context.Background()` with a short timeout) so cancellation
   does not abort the cleanup write. Keep the existing log lines.
2. **On shape init failure** (`startShape`, the `GetSceneDimensions` / `ParseAndInit` error branches):
   before returning, call `store.StopRoutineRun` for `runID` so the committed `running` row is cleared.
3. **On normal shape termination** (the ticker goroutine when its context is cancelled / loop ends):
   verify it also marks the run stopped. If the stop handler already does this via the cancel path,
   document why and avoid a double-stop race; otherwise add it. Make `StopRoutineRun` idempotent
   (stopping an already-stopped run must not error) if it is not already.
4. Ensure the cleanup is **race-free** with the explicit stop endpoint (operator clicking "stop"):
   both may try to stop the same run; the store update should be a no-op the second time.

## Acceptance / tests

- Engine test (with a fake/stub store and a fast-exiting fake Python command, or by injecting the
  runner): starting a Python routine whose process exits immediately results in the run being marked
  stopped in the store; a subsequent start on the same scene succeeds (no lingering `409`).
- Engine test: shape routine with a definition that fails `ParseAndInit` (or a scene with missing
  dimensions) marks the run stopped; subsequent start succeeds.
- Store test: `StopRoutineRun` is idempotent.
- `cd backend && go test ./...` passes.

## Out of scope

- Coordinating workers on factory reset / shutdown (WI-11).
- Reworking the conflict-detection transaction itself (it is correct; only the cleanup is missing).

## Definition of done

Every routine run reaches a terminal `routine_runs` state when its process/animation ends or fails to
initialize, so scenes never get stuck in a permanent `scene_routine_conflict`. Covered by tests.
