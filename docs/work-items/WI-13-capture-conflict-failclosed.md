# WI-13 — Capture conflict guard must fail closed on DB error

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** Backend Go (`backend/internal/capture/capture.go`)
- **Source:** Code review finding (Medium).

## Context

Before starting a capture sweep, the controller checks whether the model has an active routine run so
a sweep and a routine don't drive the same WLED device at once. In
`backend/internal/capture/capture.go` (~lines 125–129) the guard only blocks when
`ModelHasActiveRoutineRun` returns busy **and** `cherr == nil`. If the DB query returns an error, the
check is effectively skipped and the capture starts anyway — a fail-**open** bug that allows a sweep
and an active routine to fight over the device.

Key existing files:
- `backend/internal/capture/capture.go` — the `StartSweep`/start path with the `ModelHasActiveRoutineRun` call (~125–129).
- `backend/internal/store` — `ModelHasActiveRoutineRun` signature and error semantics.
- Capture tests (`backend/internal/capture/*_test.go`) and/or `backend/internal/httpapi/capture*` tests.

## Tasks

1. Treat a query **error** as a refusal: if `cherr != nil`, do not start the sweep — return an error
   that the HTTP layer maps to a sensible status (e.g. `409`/`503` with the standard error envelope),
   rather than proceeding.
2. Keep the existing busy behavior (active routine → conflict) unchanged.
3. Confirm there are no other fail-open guards in the same start path (apply the same fix if found).

## Acceptance / tests

- Capture test with a stubbed store whose `ModelHasActiveRoutineRun` returns an error: `StartSweep`
  refuses (no sweep goroutine started); error surfaced to caller.
- Existing "busy → conflict" and "free → starts" tests still pass.
- `cd backend && go test ./...` passes.

## Out of scope

- The conflict detection logic itself; broader auth.

## Definition of done

A DB error while checking for an active routine prevents the capture from starting (fail closed),
covered by a test.
