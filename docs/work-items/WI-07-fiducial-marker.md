# WI-07 — Printable fiducial marker endpoint (REQ-049)

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none (router wiring only)
- **Area:** Backend Go (`backend/internal/httpapi`)
- **Requirements:** REQ-049 (business rule 5 — optional printable marker), REQ-048 (markers improve, never gate)
- **Architecture:** `docs/architecture.md` §3.2 (`/capture/marker`), §3.23.2

## Context

The "create from video" flow may offer an **optional printable fiducial marker** the operator
prints and places in shot to improve capture accuracy and metric scale. It is **optional** and
never required to create a model.

For the MVP, serve a **static, embedded marker asset** (an ArUco-style marker). Generating markers
dynamically from the CV bundle can come later; do not pull in heavy deps here.

Existing relevant files:
- `backend/internal/httpapi/router.go` — route registration.
- `backend/internal/httpapi/json.go` — response helpers.
- `backend/internal/webdist/embed.go` — example of `//go:embed`.

## Tasks

1. Add a static marker asset to the repo and embed it, e.g. `backend/internal/httpapi/assets/fiducial_marker.pdf`
   (PDF preferred for print fidelity; PNG acceptable). Include the **printed edge length** in the
   on-page text/filename and document it (this is the metric-scale reference used by WI-05). Choose a
   concrete dictionary + size (e.g. ArUco 4x4_50, id 0, 100 mm) and keep it consistent with what WI-05 detects.
2. Add handler `getCaptureMarker` and register `GET /capture/marker` in `router.go` (it is a distinct
   literal path; place it with the other literal routes). Serve the embedded bytes with the correct
   `Content-Type` (`application/pdf` or `image/png`) and a `Content-Disposition: inline; filename=…`.
   Optional query params `type` / `size` may select among assets if you add more than one; otherwise
   serve a sensible default and ignore unknown params.
3. Keep it dependency-light: a checked-in asset + `//go:embed` is sufficient. No image-generation libs.

## Acceptance / tests

- HTTP test: `GET /api/v1/capture/marker` returns `200` with the expected content-type and a non-empty body.
- `cd backend && go test ./...` passes.

## Out of scope

- Dynamic marker generation; UI download link (WI-08 links to this endpoint); the CV detection of the marker (WI-05).

## Definition of done

`GET /api/v1/capture/marker` returns a printable marker artifact with a documented edge length;
covered by a test.
