# WI-26 — Triangulation sanity guards (cheirality + NaN/Inf)

- **Model:** Hard (Opus) — geometry reasoning
- **Depends on:** none (complements WI-25)
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`)
- **Source:** Code review findings (High / Medium). Repo root `/workspaces/dlm`.

## Context

Two robustness gaps let bad geometry escape as "valid" coordinates:

1. **No cheirality / positive-depth check** — `reconstruct.py` `_dlt_triangulate` (~593–611) and
   `triangulate_all` (~538–588) accept any homogeneous SVD solution. There is no check that the point
   lies **in front of** all contributing cameras (`Z_cam > 0`). Degenerate geometry, bad
   correspondences, or bad poses can yield points behind the camera that are still scaled and emitted as
   valid `(x, y, z)`.
2. **No finite check** — after SVD triangulation (~607–611) and before `json.dumps` (~735) there is no
   NaN/Inf guard. Ill-conditioned systems can produce NaN; Python's `json.dumps` emits non-RFC8259
   `NaN`/`Infinity` tokens by default, which can make the Go side's `json.Unmarshal` fail the **whole
   job** rather than cleanly reporting `status:"failed"` (or dropping the bad light).

Key files:
- `backend/internal/cvruntime/src/reconstruct.py` — `_dlt_triangulate` (~593–611), `_reprojection_error` (~614–...), `triangulate_all` (~538–590), JSON assembly/`json.dumps` (~632–648, ~735), `missing`/`low_confidence` lists.
- Go contract: `backend/internal/cvruntime/contract.go`, `run.go` (~66–83) result parsing.

## Tasks

1. **Cheirality:** after triangulating a point, verify it has positive depth in **every** contributing
   camera (transform the world point into each camera frame using its `Ps`/pose and check `Z > 0`,
   within a small epsilon). If it fails, treat the light as **missing** (don't emit coordinates) and log
   to stderr. Coordinate with WI-25's "missing not fabricated" policy.
2. **Finite check:** reject any triangulated point containing NaN/Inf — route to `missing`. As a final
   safety net, ensure `json.dumps` cannot emit non-finite tokens (e.g. `json.dumps(..., allow_nan=False)`
   and handle the resulting error by failing the job cleanly with `status:"failed"`, or sanitize
   beforehand so output is always valid RFC8259 JSON).
3. Keep existing reprojection-error → `low_confidence` behavior; cheirality/NaN failures are stronger
   (→ `missing`), reprojection threshold stays `low_confidence` (still emitted).

## Acceptance / tests

- New test: a correspondence that triangulates to a point behind a camera → light reported `missing`,
  not in `lights`.
- New test: forcing a degenerate/NaN triangulation → light is `missing` and the emitted JSON is valid
  (Go-side parse succeeds with a clean result; no `NaN` token).
- Existing tests still pass; run cvruntime Python tests per the src `README.md`.

## Out of scope

- Cross-feed correspondence (WI-25). Per-feed intrinsics (WI-28).

## Definition of done

Points behind any camera or with non-finite coordinates are never emitted as valid lights; output JSON
is always RFC8259-valid so Go never fails to parse. Proven by new tests.
