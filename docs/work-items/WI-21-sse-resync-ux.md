# WI-21 — Lightweight SSE resync (no full-page flash, no lost edits, no refetch race)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-18 (reconnect) and WI-20 (load guards) ideally land first; not strictly required
- **Area:** Frontend (`web/app/models/detail`, `web/app/scenes/detail`)
- **Source:** Code review findings (Medium).

## Context

Both detail pages pass their full `load` as the SSE `onReload`. That causes three UX problems on a
sequence gap or SSE error:

1. **Full-page "Loading…" flash** — `web/app/models/detail/ModelDetailClient.tsx` (~417–418, 446,
   609–610): `load` sets `loading = true`, unmounting the 3D view and lights table even though data is
   already in state.
2. **Lost in-progress edits** — `web/app/scenes/detail/SceneDetailClient.tsx` (~83–84, 104): scene
   `load` always `setPendingMarginM(null)`, so an SSE gap/error while the user is previewing a new
   boundary margin wipes their unsaved preview.
3. **SSE-vs-refetch race** — `SceneDetailClient.tsx` (~108–127): while a routine runs, polling every
   1.5s can call `fetchScene` when `sceneSseLiveRef.current` is false; a slower GET can return an older
   snapshot and overwrite newer SSE-merged state (flicker/stale lights). Related: initial GET vs SSE
   race on mount (`ModelDetailClient.tsx` ~441–446 + 411–432) where `load()` completing can revert
   fresher SSE merges.

Key existing files:
- `web/app/models/detail/ModelDetailClient.tsx`, `web/app/scenes/detail/SceneDetailClient.tsx`.
- `web/lib/useModelLightsSSE.ts`, `web/lib/useSceneLightsSSE.ts` (the `onReload` contract, `seqRef`, `sseLiveRef`).

## Tasks

1. **Resync without the loading flash:** add a separate lightweight resync fetch used as `onReload`
   that updates `model`/`scene` state **without** toggling the page-level `loading` flag (keep the 3D
   view/table mounted). Reserve `loading` for the genuine initial load.
2. **Preserve in-progress edits:** the resync path must not clobber transient UI edit state. Don't
   reset `pendingMarginM` (and similar unsaved previews) on an SSE-triggered resync; only reset on an
   explicit user-initiated reload or successful save.
3. **Fix the refetch race:** coordinate full GETs with SSE sequence numbers. When applying a full
   fetch result, reconcile with `seqRef` so a stale GET cannot overwrite a newer SSE-merged state
   (e.g. ignore/merge a fetch whose snapshot predates the latest applied sequence, or pause the
   periodic poll while SSE is live and only poll when `sseLiveRef` is false **and** no newer SSE
   sequence has been applied since the GET started).
4. Keep behavior correct: a real gap still triggers an actual resync so lights converge to truth.

## Acceptance / tests

- Test (or manual note + unit coverage where feasible): an SSE `onReload` updates state without setting
  `loading` true; the 3D canvas/table stay mounted.
- Scene test: an SSE-triggered resync preserves `pendingMarginM`.
- Test the race fix: a full fetch resolving with an older snapshot does not overwrite newer
  SSE-applied state.
- `cd web && npm test && npm run lint` pass.

## Out of scope

- The reconnect mechanism itself (WI-18). Stale-load guards on id change (WI-20).

## Definition of done

SSE gaps/errors resync silently without a full-page flash, without discarding unsaved margin edits, and
without stale GETs clobbering newer live state. Covered by tests where feasible.
