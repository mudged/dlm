# WI-28 — Use per-feed intrinsics in the essential-matrix path

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`)
- **Source:** Code review finding (Medium).

## Context

The essential-matrix fallback uses only camera 0's intrinsics for both feeds:
`reconstruct.py` ~314–332 calls `cv2.findEssentialMat(pts0, pts1, Ks[0], ...)` and `recoverPose` with
`Ks[0]`. If the two uploaded feeds have **different resolutions/intrinsics** (different phones, or one
downscaled differently), using `Ks[0]` for the second feed biases the recovered pose and therefore all
downstream triangulation.

Key files:
- `backend/internal/cvruntime/src/reconstruct.py` — `estimate_poses` essential-matrix branch (~304–332), how `Ks` is built per feed, and the intrinsics-scaling for the `MAX_PROC_DIM` downscale.
- Verify how `Ks[i]` is derived per feed (and whether the downscale already adjusts each `K`).

## Tasks

1. Use **per-feed intrinsics** when relating feed 0 and feed 1. Standard approaches:
   - Normalize each feed's points with its own `K` (`undistortPoints` / manual `K^-1`) and compute the
     essential matrix on normalized coordinates (identity `K`), then `recoverPose` accordingly; **or**
   - If staying with pixel-coordinate APIs, ensure both feeds' points are expressed under a consistent
     calibration rather than forcing feed 1 through `Ks[0]`.
2. Confirm the `MAX_PROC_DIM` downscale adjusts each feed's `K` consistently with the points it
   produces (scale fx, fy, cx, cy by the same factor used on that feed's frames). Fix if mismatched.
3. Keep the ArUco path and scale-hint logic unchanged.

## Acceptance / tests

- New/extended test with two feeds at **different** intrinsics (e.g. different focal length/resolution
  in a synthetic fixture): recovered relative pose and triangulated points are accurate within
  tolerance, whereas the old `Ks[0]`-for-both code would be biased.
- Same-intrinsics fixtures still pass.
- Run cvruntime Python tests per `backend/internal/cvruntime/src/README.md`.

## Out of scope

- ArUco scale (WI-24), correspondence (WI-25), geometry guards (WI-26).

## Definition of done

The essential-matrix path uses each feed's own intrinsics, so mixed-resolution captures reconstruct
correctly. Proven by a mixed-intrinsics test.
