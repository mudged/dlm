# WI-17 — Backend minor cleanups (capture map, stop state, confirm limit, progress)

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none (light overlap with WI-16)
- **Area:** Backend Go (`backend/internal/capture`, `backend/internal/httpapi`, `backend/internal/reconstruct`)
- **Source:** Code review findings (Low). Each sub-task is independent; do as many as cleanly apply.

## Context & tasks

Four small, low-risk correctness/cleanliness fixes found in review:

1. **Capture sweep map retention** — `backend/internal/capture/capture.go` (~86–87, 134–151).
   Completed sweeps stay in `c.sweeps` indefinitely (only replaced on the next start for that device).
   With many unique device IDs over time the map grows. Fix: when a sweep goroutine finishes, either
   delete its entry or transition it to a small terminal record; ensure access stays mutex-guarded.

2. **Capture stop reports `idle` prematurely** — `backend/internal/httpapi/capture.go` (~71–74).
   `postCaptureStop` returns `"state":"idle"` immediately after signaling stop, before the sweep
   goroutine finishes turning LEDs off. Fix: return a `"stopping"` state (or wait briefly/poll the
   controller for the real state) so clients don't see `idle` while hardware is mid-sweep. Keep the
   endpoint non-blocking; prefer a distinct transitional state over blocking.

3. **Confirm endpoint missing body limit** — `backend/internal/httpapi/capture_models.go` (~131–134),
   `postModelsCaptureConfirm`. Wrap the JSON body in `http.MaxBytesReader` with a small limit, matching
   the other mutating handlers.

4. **Reconstruction progress never updates** — `backend/internal/reconstruct/reconstruct.go`
   (~55–64, 189–195). `Job.Progress` stays `0` until success. Either update `Progress` at a few
   coarse milestones during `runJob` (under the mutex) so clients can show partial progress, **or**
   remove `Progress` from the API and document that only status transitions are observable. Pick one
   and be consistent with the frontend (`web/app/models/new` video panel polling).

## Acceptance / tests

- Capture test: a finished sweep no longer retains a live entry (or retains only a bounded terminal
  record). Stop returns a transitional/accurate state.
- HTTP test: oversized confirm body → `400`/`413` standard envelope.
- Reconstruct test: progress is monotonic and reaches the documented terminal value (or progress field
  is gone and docs/tests reflect that).
- `cd backend && go test ./...` passes.

## Out of scope

- Job concurrency/retention caps (WI-16). Lifecycle shutdown (WI-11).

## Definition of done

The four cleanups are applied (or consciously skipped with a note), no bounded-map growth from
completed sweeps, accurate capture-stop state, a body limit on confirm, and consistent progress
reporting. Tests green.
