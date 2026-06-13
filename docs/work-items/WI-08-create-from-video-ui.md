# WI-08 — Models "create from video" UI (REQ-049)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-06 (`/models/capture*` API), WI-07 (marker endpoint). Reuses the existing model 3D canvas for the preview.
- **Area:** Frontend (`web/`, Next.js App Router + Tailwind + TypeScript)
- **Requirements:** REQ-049, REQ-048, REQ-002 (responsive), REQ-010 (3D preview reuse)
- **Architecture:** `docs/architecture.md` §4.6, §4.7, §4.17, §8.25

## Context

Add a second model-creation path alongside CSV upload: **create a model from uploaded videos**.
The flow is: upload ≥ 2 videos → poll the async reconstruction job → **review** the detected lights →
**confirm** (with a name) to persist a normal model, or **cancel** to discard. Processing is
server-side, so navigating away does not abort it.

Backend (WI-06/WI-07):
- `POST /api/v1/models/capture` — `multipart` `files` (≥ 2) + optional `marker`/`scale_hint` → `202 { job_id, status }`.
- `GET /api/v1/models/capture/{jobId}` → `{ status, progress, result?: { light_count, lights:[{id,x,y,z}], missing:[id], low_confidence:[id] }, error? }`.
- `POST /api/v1/models/capture/{jobId}/confirm` — `{ name }` → `201` model (then route to it). `400`/`404`/`409`.
- `DELETE /api/v1/models/capture/{jobId}` → `204`.
- `GET /api/v1/capture/marker` — printable marker (link/download).

Existing relevant files:
- `web/lib/models.ts` — model fetchers (extend here).
- `web/app/models/new/page.tsx` — current CSV upload form (add the video path / tab).
- `web/app/models/detail/ModelDetailClient.tsx` — contains the three.js model canvas; reuse the canvas component for the review preview (extract/share it if needed, keeping the detail page working).

## Tasks

1. **`web/lib/models.ts`:** add typed fetchers `createCaptureJob(files, params)`,
   `getCaptureJob(jobId)`, `confirmCaptureJob(jobId, name)`, `discardCaptureJob(jobId)`, with a
   `CaptureJob` type matching the API. Surface server `error.message` / `error.code`.
2. **Entry point:** on `web/app/models/new` (a tab/segmented control, or a sibling route like
   `/models/new/video`), present **CSV upload** and **Create from video** clearly distinguished
   (REQ-049 BR1). The video panel has a **multi-file** input (require ≥ 2; client-side hint, server
   enforces), an optional **fiducial marker** selector, and a **Download printable marker** link to
   `GET /api/v1/capture/marker` (optional — never blocks submit, REQ-049 BR5).
3. **Submit + progress:** submit → `job_id`; show a **pending/progress** state and poll
   `getCaptureJob` until terminal. Make the job resumable by URL (e.g. include `jobId` in the route or
   query) so closing/reopening the tab returns to the review (processing continues server-side).
4. **Review (REQ-049 BR3):** on `succeeded`, show the **detected light count** and any
   **missing / low-confidence** ids, plus an optional **read-only 3D preview** reusing the model
   canvas (REQ-010) over the candidate coordinates. Require an explicit **Confirm** with a **name**
   field, or **Cancel**.
5. **Confirm/cancel:** Confirm → `confirmCaptureJob`; on `201` navigate to the new `/models/[id]`.
   Show `400`/`409` messages inline (validation, duplicate name). Cancel → `discardCaptureJob` and
   return to the models list / CSV tab.
6. **Responsive (REQ-002):** mobile stacks uploader → progress → review with full-width controls and
   a non-hover-only Confirm; tablet/desktop may split upload and review.

## Acceptance / tests

- `cd web && npm run lint` and `cd web && npm test` pass.
- Add `web/lib/models.ts` tests (mock `fetch`): the capture fetchers hit the right URLs/methods,
  send `FormData` for the multipart create, and surface error messages. Match existing `web/lib/*.test.ts` style.
- Manual check (optional) with `./scripts/run.sh`: the new-model screen offers the video path; a
  submitted job shows progress, then a review with confirm/cancel.

## Out of scope

- Per-light manual editing of detected coordinates (future). Backend behaviour (WI-06/WI-07).

## Definition of done

Users can create a model from uploaded videos: choose the video path, upload ≥ 2 files, watch
progress, review detected vs missing lights (with an optional 3D preview), then confirm to create a
normal model or cancel to discard; lint + tests green.
