# WI-20 — Guard async loads against stale responses on id change

- **Model:** Medium (Sonnet)
- **Depends on:** none
- **Area:** Frontend (`web/app/models/detail`, `web/app/scenes/detail`, `web/app/routines/python`, `web/app/routines/shape`)
- **Source:** Code review finding (High).

## Context

Detail/editor pages read a route query param (`id`) and `load()` the corresponding resource. These
async `load` handlers apply their results with **no `AbortController`, generation token, or
post-await check** that the response still matches the current id. Fast navigation (model A → model B)
can let A's slower response land **after** B's and overwrite state, showing the wrong resource.

Affected files (verify exact line numbers before editing):
- `web/app/models/detail/ModelDetailClient.tsx` (~411–438)
- `web/app/scenes/detail/SceneDetailClient.tsx` (~70–96)
- `web/app/routines/python/PythonRoutineEditorClient.tsx` (~101–126)
- `web/app/routines/shape/ShapeRoutineEditorClient.tsx` (~101–133)

Note: the **target-scene** fetches in the routine editors already use a `cancelled` flag
(`PythonRoutineEditorClient.tsx` ~153–168; `ShapeRoutineEditorClient.tsx` ~161–186) — mirror that
proven pattern for the main `load`.

## Tasks

1. For each `load`, add a guard so a response is only applied if it still corresponds to the current
   request. Use **one** consistent approach across the four files:
   - a `let cancelled = false` flag set in the effect cleanup (matching the existing routine-editor
     pattern), and/or
   - an `AbortController` passed to `fetch` and aborted on cleanup / id change, plus a check that the
     resolved id matches the current `id` before `setState`.
2. Ensure the guard also prevents `setState` after unmount.
3. Keep loading/error UX correct (don't leave a stuck spinner if a request is aborted intentionally —
   an abort should be treated as "superseded", not an error to display).

## Acceptance / tests

- Where feasible, add a test simulating out-of-order resolution (id A request resolves after id B):
  state reflects B, not A. At minimum cover the models detail page.
- `cd web && npm test && npm run lint` pass.

## Out of scope

- SSE resync behavior (WI-21). PATCH/save unmount guards (WI-23 task).

## Definition of done

Switching ids quickly never shows a stale resource; aborted/superseded loads don't overwrite newer
state or update unmounted components. Covered by at least one test.
