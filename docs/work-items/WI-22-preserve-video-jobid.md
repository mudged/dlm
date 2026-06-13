# WI-22 — Preserve in-progress video job across tab switch

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** Frontend (`web/app/models/new/NewModelClient.tsx`)
- **Source:** Code review finding (Medium).

## Context

The "create model" page has CSV and video tabs. `switchTab`
(`web/app/models/new/NewModelClient.tsx` ~678–680) navigates to `/models/new?tab=…` **without
preserving `jobId`**, and `VideoPanel` unmounts on tab switch (~695–700). So switching CSV ↔ video
during an active reconstruction drops the job link; returning to the video tab shows the idle upload
form instead of the in-progress job (unless the old URL is still in history).

Key existing files:
- `web/app/models/new/NewModelClient.tsx` — `switchTab` (~678–680), tab rendering, `VideoPanel` (mount/unmount ~695–700, polling/cleanup ~195–200, 245–254), and how `jobId`/`tab` are read from the URL.

## Tasks

1. When switching tabs, **preserve the current `jobId`** in the URL query (e.g.
   `/models/new?tab=video&jobId=…`) so navigating back to the video tab rehydrates the in-progress job.
2. On mount/tab-activation, read `jobId` from the URL and resume polling that job (reuse the existing
   `VideoPanel` polling) rather than showing the idle form.
3. Keep the existing timeout cleanup on unmount/phase change intact (no double-polling, no leaks).
4. If lifting `jobId` to a parent that survives tab switches is simpler/cleaner than URL round-trips,
   that is acceptable — pick one approach and keep it consistent.

## Acceptance / tests

- Add/extend a test: with an active `jobId`, switching to CSV and back to video preserves the job
  (URL carries `jobId`, or state persists) and resumes polling rather than resetting to the upload form.
- `cd web && npm test && npm run lint` pass.

## Out of scope

- Backend job retention (WI-16) — assume the job still exists server-side while active.

## Definition of done

Switching tabs during reconstruction keeps the in-progress video job visible and polling; covered by a
test.
