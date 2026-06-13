# WI-03 — Devices UI: light count + capture controls (REQ-047)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-01 (`light_count` in API), WI-02 (`/devices/{id}/capture/*` endpoints)
- **Area:** Frontend (`web/`, Next.js App Router + Tailwind + TypeScript)
- **Requirements:** REQ-047, REQ-002 (responsive), REQ-018 (icons/nav)
- **Architecture:** `docs/architecture.md` §4.15, §8.24

## Context

Repo `/workspaces/dlm`, frontend in `web/` (Next.js static export, client components,
`fetch` to same-origin `/api/v1/...`). Devices UI already exists:
- `web/lib/devices.ts` — typed fetchers (`Device`, `createDevice`, `patchDevice`, …).
- `web/app/devices/new/page.tsx` — add-device form.
- `web/app/devices/detail/DeviceDetailClient.tsx` (+ `page.tsx`) — device detail/edit, assign/unassign.

This item adds (a) a **light count** field to the add/edit forms and (b) **Start/Stop capture**
controls with status polling on the device detail page.

Backend endpoints (from WI-01/WI-02):
- `light_count` is a field on the device JSON (POST/PATCH accept it, GET returns it).
- `POST /api/v1/devices/{id}/capture/start` → `200 { device_id, state, light_count, current_index }`; errors `404`, `409 capture_conflict`, `422 capture_no_lights`.
- `POST /api/v1/devices/{id}/capture/stop` → `200 { device_id, state:"idle" }`.
- `GET /api/v1/devices/{id}/capture` → `{ state:"idle"|"running", light_count, current_index? }`.

## Tasks

1. **`web/lib/devices.ts`:**
   - Add `light_count: number` to the `Device` type.
   - Add `light_count` to `createDevice` input and `patchDevice` patch (send as a number).
   - Add fetchers: `startCapture(id)`, `stopCapture(id)`, `getCaptureStatus(id)` returning a typed
     `CaptureStatus = { state: "idle" | "running"; light_count: number; current_index?: number }`.
     Preserve the existing error-message extraction. `startCapture` must surface the server
     `error.code` (e.g. `capture_conflict`, `capture_no_lights`) so the UI can show tailored copy.
2. **Add/edit forms** (`devices/new/page.tsx`, `DeviceDetailClient.tsx`): add a **Light count**
   number input (integer, min 0, max 1000) bound to `light_count`. Show inline validation and the
   server `validation_failed` message on save errors.
3. **Capture controls** on `DeviceDetailClient.tsx`:
   - **Start capture** / **Stop capture** buttons, each with a visible Font Awesome icon + text
     label (REQ-018) and reachable without hover (REQ-002).
   - On mount and while running, **poll** `getCaptureStatus` (e.g. every ~1 s via `setInterval`,
     cleared on unmount) and display `state` and, when running, `current_index` / `light_count`
     (e.g. "Lighting 12 / 200"). Stop polling when idle.
   - **Disable** Start with explanatory copy when `light_count === 0`.
   - Handle errors: `capture_conflict` → "A capture is already running, or this device's model has
     an active routine."; `capture_no_lights` → "Set a light count first." Generic fallback otherwise.
   - Brief helper text: remind the operator to start recording from each camera angle before/while
     the sweep runs (videos are uploaded later — see Models "create from video", WI-08).
4. **Responsive:** controls stack full-width on mobile; no hover-only affordances (REQ-002).

## Acceptance / tests

- `cd web && npm run lint` and `cd web && npm test` pass.
- Add/extend tests for `web/lib/devices.ts` (mock `fetch`): `light_count` is sent on create/patch
  and parsed on read; `startCapture`/`stopCapture`/`getCaptureStatus` hit the right URLs and
  surface `error.code` on failure. Match the existing test style in `web/lib/*.test.ts`.
- Manual check (optional): with the backend running (`./scripts/run.sh`), the detail page shows the
  light-count field and capture buttons; starting reflects a running status.

## Out of scope

- Backend behaviour (WI-01/WI-02). Reconstruction / create-from-video (WI-06/WI-08).

## Definition of done

Operators can set a device's light count and start/stop the capture sequence from the device
detail screen, see live progress, and get clear messages on conflict/zero-count; lint + tests green.
