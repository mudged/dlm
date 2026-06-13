# WI-12 — Non-blocking SSE fan-out (don't block API handlers on slow clients)

- **Model:** Medium (Sonnet)
- **Depends on:** none
- **Area:** Backend Go (`backend/internal/httpapi/revision_hub.go`)
- **Source:** Code review finding (Medium, concurrency).

## Context

The SSE revision hub fans out light updates to subscribers. In
`backend/internal/httpapi/revision_hub.go` (~lines 98–145) the `emit` path sends on each subscriber's
channel with a **blocking** send (`ch <- msg`) on a 1024-buffered channel. `emit` is invoked
synchronously from light-patch HTTP handlers via `notifyModelLightsChanged` /
`notifyAfterSceneLightPatch`. A slow, stalled, or backpressured SSE client whose buffer fills will
block `emit`, which in turn delays or times out unrelated API requests. On a Pi under bursty light
updates this can stall the whole API.

Key existing files:
- `backend/internal/httpapi/revision_hub.go` — `subscribe`, `unsubscribe`, `emit`, the per-subscriber channel and buffer size; the SSE handler goroutine that reads the channel and writes to the `http.ResponseWriter`.
- Callers: `notifyModelLightsChanged`, `notifyAfterSceneLightPatch` (and any scene equivalent).
- Existing SSE tests (search for `revision_hub` / `events` tests).

## Tasks

1. Change the fan-out to a **non-blocking send with drop/coalesce semantics**: in `emit`, use a
   `select { case ch <- msg: default: ... }` so a full subscriber buffer never blocks the publisher.
2. Decide and implement the drop policy so clients stay correct:
   - Preferred: when a subscriber would block, mark it "behind" (drop the delta) and signal it to do a
     **full resync** (the client already supports resync on sequence gap — bump/break the sequence so
     the client detects the gap and re-fetches). This keeps lights correct without unbounded buffering.
   - Alternatively, close/disconnect a persistently-behind subscriber so it reconnects (works with the
     client reconnect added in WI-18).
3. Ensure `unsubscribe`/cleanup is race-free with `emit` (don't send on a closed channel; guard with
   the hub mutex or a `done` channel).
4. Keep `emit` O(subscribers) and non-blocking; do not hold the hub lock while writing to sockets.

## Acceptance / tests

- Unit test: a subscriber that never drains its channel does **not** block `emit`; other subscribers
  still receive messages; the slow subscriber is either marked for resync or disconnected.
- Unit test: normal subscribers receive deltas in order with correct sequence numbers.
- Race detector clean: `cd backend && go test -race ./internal/httpapi/...`.
- `cd backend && go test ./...` passes.

## Out of scope

- Client-side reconnect/resync behavior (WI-18 / WI-21) — though this item should leave the protocol
  compatible with a client that resyncs on gap.

## Definition of done

A slow or stalled SSE client can never block light-patch API handlers; behind subscribers are dropped
or resynced rather than buffered unboundedly. Verified under `-race`.
