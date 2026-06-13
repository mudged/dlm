# WI-19 — Capture progress polling after "Start capture"

- **Model:** Medium (Sonnet)
- **Depends on:** none
- **Area:** Frontend (`web/app/devices/detail/DeviceDetailClient.tsx`)
- **Source:** Code review finding (High).

## Context

On the device detail page, capture progress (`current_index` / `light_count`) is polled by a 1s
interval that is created **only inside the mount effect when the initial `getCaptureStatus` returns
`running`** (`web/app/devices/detail/DeviceDetailClient.tsx` ~82–120). When the user starts a capture
from the button (`onStartCapture` ~234–258), `captureStatus` updates once but **no interval starts**,
so the progress UI stays frozen at the initial value for the whole sweep.

Key existing files:
- `web/app/devices/detail/DeviceDetailClient.tsx` — the poll effect (~82–120), `onStartCapture` (~234–258), `onStopCapture`, the `captureStatus` state, and `getCaptureStatus` in the API client (`web/lib/*`).

## Tasks

1. Drive polling off the **capture state**, not just the initial fetch. Refactor so a single effect
   (or a small helper) starts a 1s interval whenever `captureStatus?.state === "running"` (or
   `"stopping"`, if WI-17 adds it) and clears it otherwise. Use the `running` state as the effect
   dependency so starting/stopping capture starts/stops polling automatically.
2. Ensure `onStartCapture` updates the state such that the polling effect engages (e.g. set state to
   `running` from the start response, then let the effect poll).
3. Stop polling when the sweep reaches a terminal state and on unmount; clear the interval in cleanup
   (no leaks, no setState-after-unmount).
4. Handle poll errors gracefully (don't spin forever on repeated failures; surface an error state).

## Acceptance / tests

- If a component test harness exists for this page, add a test: starting capture begins polling and
  `current_index`/progress advances as `getCaptureStatus` returns increasing values (fake timers + mocked
  API); polling stops at terminal state and on unmount.
- Otherwise add at least a unit test around the extracted polling helper.
- `cd web && npm test && npm run lint` pass.

## Out of scope

- Backend capture-stop state accuracy (WI-17 task 2) — consume whatever state the API returns.

## Definition of done

Capture progress updates live regardless of whether the sweep was already running on mount or started
from the button; the interval is correctly torn down. Covered by a test.
