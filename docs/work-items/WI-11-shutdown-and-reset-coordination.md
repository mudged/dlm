# WI-11 — Stop background workers on factory reset & graceful shutdown

- **Model:** Medium (Sonnet)
- **Depends on:** WI-10 (clean run termination makes this simpler; not strictly required)
- **Area:** Backend Go (`backend/cmd/server`, `backend/internal/httpapi`, `backend/internal/routineengine`, `backend/internal/capture`, `backend/internal/reconstruct`)
- **Source:** Code review findings (High / Medium).

## Context

Background workers are not coordinated with two lifecycle events:

1. **Factory reset** — `backend/internal/httpapi/system.go` (~lines 8–23) wipes the DB
   (`store` reset, ~`store.go` 517–556) and reloads in-memory light state, but does **not** stop the
   routine engine, capture sweeps, or in-flight reconstruction jobs. Those goroutines keep mutating
   the old in-memory light state and reference scene/model IDs that no longer exist, producing
   inconsistent state until restart.
2. **Graceful shutdown** — `backend/cmd/server/main.go` (~lines 93–100) calls `srv.Shutdown` for the
   HTTP server only. The routine engine launches goroutines from `context.Background()`, `capture`
   has a `Shutdown` method (~`capture.go` 224–235) that is **never called**, and reconstruction jobs
   are not cancelled. On SIGTERM the HTTP server drains but lights keep changing and Python children
   keep running until the process is killed.

Key existing files:
- `backend/cmd/server/main.go` — wiring of `apiDeps`, server start, signal handling, `srv.Shutdown`.
- `backend/internal/routineengine/engine.go` — the `active` map of cancel funcs; add a `StopAll`/`Shutdown`.
- `backend/internal/capture/capture.go` — existing `Shutdown` (~224–235) and the `sweeps` map.
- `backend/internal/reconstruct/reconstruct.go` — job manager / worker goroutines; add cancellation.
- `backend/internal/httpapi/system.go` — factory-reset handler.

## Tasks

1. **Engine `StopAll`/`Shutdown`:** add a method to `routineengine.Engine` that cancels every cancel
   func in `active` (under its mutex), waits (bounded) for goroutines to exit, and marks their runs
   stopped via the store (reuse WI-10's cleanup). Make the engine own a root context created in
   `main` (not `context.Background()` inside `startPython`/`startShape`) so a single cancel tears
   everything down.
2. **Reconstruct manager `Shutdown`:** cancel running jobs (their `exec.CommandContext` contexts) and
   stop accepting new ones; safe to call once.
3. **Factory reset:** in `system.go`, **before** wiping the DB, stop the routine engine (`StopAll`),
   stop capture sweeps (`capture.Shutdown` or a lighter `StopAll`), and cancel reconstruction jobs.
   Then reset the DB, reload light state, and (if appropriate) allow the engine/capture to accept new
   work again. Ensure ordering avoids a goroutine writing light state *after* the reload. Document the
   chosen order in a comment.
4. **Graceful shutdown:** in `main.go` signal handler, after (or alongside) `srv.Shutdown`, call the
   engine, capture, and reconstruct shutdown hooks with a bounded timeout so SIGTERM stops all
   background light mutation and Python children.
5. Inject the necessary references (engine, capture controller, reconstruct manager) into `apiDeps`
   so `system.go` can reach them; if they are already there, just use them.

## Acceptance / tests

- Engine test: `StopAll` cancels active runs and they reach terminal store state.
- HTTP/integration test for factory reset: with a fake engine/capture/reconstruct, the reset handler
  invokes their stop hooks before resetting the store (assert call order via fakes/spies).
- Shutdown test (or a focused unit test of the shutdown sequence) verifying all hooks are invoked.
- `cd backend && go test ./...` passes; manual SIGTERM no longer leaves lights changing (note in PR).

## Out of scope

- Authentication / confirmation for factory reset (WI-15).
- Bounding reconstruction memory/job count (WI-16).

## Definition of done

Factory reset and SIGTERM both deterministically stop the routine engine, capture sweeps, and
reconstruction jobs (and their Python children) before tearing down or exiting; covered by tests.
