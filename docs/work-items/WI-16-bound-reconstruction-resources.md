# WI-16 — Bound reconstruction jobs & child output buffers

- **Model:** Medium (Sonnet)
- **Depends on:** none (touches the same package as WI-11)
- **Area:** Backend Go (`backend/internal/reconstruct`, `backend/internal/cvruntime`)
- **Source:** Code review findings (Medium, resource exhaustion on Pi).

## Context

Two unbounded-resource issues in the reconstruction path are risky on a Pi 4:

1. **Unbounded jobs** — `backend/internal/reconstruct/reconstruct.go` (~80–82, 98–157): each
   `POST /models/capture` spawns a goroutine + a Python child, and job metadata lives in an in-memory
   map until confirm/discard. Repeated uploads without cleanup can exhaust RAM, disk
   (`runtime/capture/<jobID>`), and CPU.
2. **Unbounded child output** — `backend/internal/cvruntime/run.go` (~66–68): the child's
   stdout/stderr are captured into `bytes.Buffer`s with no cap. A verbose or misbehaving CV child can
   grow memory without limit before it exits.

Key existing files:
- `backend/internal/reconstruct/reconstruct.go` — `Manager`, job map, worker, `Create`/`Confirm`/`Discard`, any janitor.
- `backend/internal/cvruntime/run.go` — `exec.CommandContext`, stdout/stderr capture, result parsing (~66–83).
- `backend/internal/config/config.go` — for data-dir/limits config.

## Tasks

1. **Concurrency cap:** allow at most **one** reconstruction job running at a time (Pi default), with
   additional `POST /models/capture` either queued (bounded queue) or rejected with `429`/`503`
   standard envelope. Make the cap a small config/const.
2. **Job retention / janitor:** cap the number of retained jobs and evict the oldest terminal jobs
   (delete their work dirs). Add a TTL so abandoned `succeeded`/`failed` jobs (never confirmed/discarded)
   are cleaned up after a bounded time. Keep the existing boot-time stale-dir cleanup.
3. **Bounded child output:** wrap the child stdout (and stderr) capture in a size-limited writer
   (e.g. a custom `limitedBuffer` or `io.LimitWriter`-style cap of a few MB). If the cap is exceeded,
   stop buffering and treat as a failed job with a clear error rather than OOM. Keep enough of stderr
   for diagnostics (e.g. last N KB).
4. Confirm the child already has a timeout (via `exec.CommandContext` + a deadline). If not, add one.

## Acceptance / tests

- Reconstruct test (fake `CVRunner`): submitting more jobs than the cap queues or rejects; terminal
  jobs are evicted and their work dirs removed; TTL eviction removes abandoned jobs (can fake the clock
  or call the janitor directly).
- `cvruntime` test: a fake child that prints more than the cap is failed cleanly (no unbounded buffer);
  a normal child still parses its JSON result.
- `cd backend && go test ./...` passes.

## Out of scope

- Cancelling jobs on shutdown/reset (WI-11) — keep the manager's `Shutdown` compatible.
- The CV algorithm itself.

## Definition of done

Reconstruction has a bounded number of concurrent/retained jobs with work-dir cleanup, and child
output cannot grow memory without limit; covered by tests.
