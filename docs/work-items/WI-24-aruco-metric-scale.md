# WI-24 — Fix ArUco metric double-scaling

- **Model:** Medium (Sonnet)
- **Depends on:** none
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`)
- **Source:** Code review finding (**Critical**). Repo root `/workspaces/dlm`.

## Context

When all cameras localize via ArUco markers, `estimate_poses` returns
`metric_scale = marker_spec["edge_length_m"]` (`reconstruct.py` ~298), and `triangulate_all`
multiplies every triangulated 3D point by that value (~587: `x, y, z = pt3d * metric_scale`).

But the ArUco pose path already produces **metric** results: `solvePnP` uses object points defined in
metres (`_estimate_pose_single`, ~398: `half = edge_m / 2.0`; object points at `±half`). So the camera
poses (and hence DLT-triangulated points) are already in metres. Multiplying again by the marker edge
length shrinks every coordinate by the marker size (e.g. a 0.05 m marker → coordinates ~20× too
small). This breaks the **primary** marker-based scale path (REQ-048 BR 4). Existing tests never hit a
successful ArUco path (they exercise the essential-matrix fallback), so this is uncaught.

Key files:
- `backend/internal/cvruntime/src/reconstruct.py` — `estimate_poses` (~286–384, esp. the ArUco success return ~296–298), `_estimate_pose_single` (~387–414), `triangulate_all` (~538–590), and the docstring at ~280–284 claiming the multiplier converts to metres.
- `backend/internal/cvruntime/src/test_reconstruct.py`, `gen_fixtures.py` — test scaffolding.
- README: `backend/internal/cvruntime/src/README.md` (marker examples) and `docs/requirements.md` REQ-048 BR 4.

## Tasks

1. **Correct the scale for the ArUco success path:** since ArUco poses are already metric, the ArUco
   path must return `metric_scale = 1.0` (and triangulation already yields metres). Confirm the units
   end-to-end: the camera-0-relative poses from ArUco are in metres, so DLT output is metres; do **not**
   multiply by `edge_length_m`.
2. Verify the **essential-matrix + `scale_hint_m`** path is unaffected (that path legitimately scales a
   unit-baseline reconstruction by the supplied baseline). Only the ArUco branch is wrong.
3. Update the misleading docstring (~280–284) to describe the corrected semantics.
4. **Add a test that actually exercises a successful ArUco reconstruction** (this is the gap that hid
   the bug): generate or stub a synthetic multi-feed scene where ArUco localizes all cameras (extend
   `gen_fixtures.py` if needed) and assert reconstructed coordinates are in metres at the correct scale
   (e.g. a known 1 m separation reconstructs as ~1 m, not ~0.05 m).

## Acceptance / tests

- New ArUco-success test: reconstructed inter-light distances match ground truth in metres within a
  small tolerance.
- Existing essential-matrix/scale-hint tests still pass unchanged.
- Run the Python tests the way the repo does (see `backend/internal/cvruntime/src/README.md`; e.g.
  `python -m pytest` within the cvruntime src, or the documented runner). `cd backend && go test ./...`
  still passes (Go contract unchanged).

## Out of scope

- Ordinal drift (WI-25), geometry guards (WI-26), per-feed intrinsics (WI-28).

## Definition of done

Successful ArUco reconstruction produces correctly-scaled metric coordinates (no double-scaling),
proven by a new test that exercises the ArUco path.
