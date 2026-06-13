# WI-18 — SSE auto-reconnect with backoff (model & scene light streams)

- **Model:** Medium (Sonnet)
- **Depends on:** none (pairs well with WI-12 server-side)
- **Area:** Frontend (`web/lib/useModelLightsSSE.ts`, `web/lib/useSceneLightsSSE.ts`)
- **Source:** Code review finding (High). Repo root `/workspaces/dlm`.

## Context

Live light updates use `EventSource` via `useModelLightsSSE` / `useSceneLightsSSE`. On any error their
`onerror` handler calls `es.close()` and triggers `onReload`, but the `useEffect` that created the
`EventSource` does **not** re-run (its deps — `modelId`, `enabled`, `setModel`, `onReload`,
`sseLiveRef` — are unchanged). So after a transient network blip the connection is gone for good: live
updates freeze until the user navigates away or the id changes. The model detail page has **no polling
fallback**, so the 3D view and lights table can freeze while the server keeps updating.

Key existing files:
- `web/lib/useModelLightsSSE.ts` — `createModelSSEHandlers` (`onopen`/`onmessage`/`onerror`, ~75–82) and the `useEffect` that opens the `EventSource` (~111–137).
- `web/lib/useSceneLightsSSE.ts` — same pattern (~37–103, onerror ~89–96).
- `web/lib/sseUrl.ts` — `eventSourceUrl` (dev/prod origin handling).
- `web/lib/useModelLightsSSE.test.ts` — existing handler tests to extend.

## Tasks

1. Add **reconnect with bounded exponential backoff** to both hooks. On `onerror`: close the current
   `EventSource`, then schedule a reconnect (e.g. 1s → 2s → 5s → max ~15s, with small jitter) that
   opens a **new** `EventSource` to the same URL. Reset backoff to the base on a successful `onopen`.
2. Preserve correctness: on reconnect, the client must resync (it already resets `seqRef`/`sseLiveRef`
   and re-fetches via `onReload` on gap) so it can't apply stale deltas. Keep the existing
   sequence-gap → resync logic intact.
3. **Clean up on unmount and id change:** clear any pending reconnect timer in the effect cleanup and
   when `modelId`/`sceneId`/`enabled` changes (no reconnect storms, no timer leaks, no setState after
   unmount). Guard against opening a new connection after the effect has been torn down.
4. Avoid hammering: cap retries' frequency; if `enabled` is false, do not reconnect.
5. Consider extracting the shared reconnect logic so both hooks use one implementation.

## Acceptance / tests

- Extend `web/lib/useModelLightsSSE.test.ts` (and add a scene equivalent): simulate an `onerror`, then
  verify a new connection is attempted after the backoff delay (fake timers), backoff grows on repeated
  failures, and resets after `onopen`.
- Test that unmount / id-change clears the pending reconnect timer (no reconnect after teardown).
- `cd web && npm test && npm run lint` pass.

## Out of scope

- The server-side non-blocking fan-out (WI-12). Avoiding the "Loading…" flash on resync (WI-21).

## Definition of done

After a transient SSE error, both hooks reconnect automatically with backoff and resync correctly,
with all timers/connections cleaned up on unmount and id change; covered by tests.
