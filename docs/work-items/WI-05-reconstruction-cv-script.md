# WI-05 — Reconstruction CV script: detect, pose, triangulate (REQ-048)

- **Model:** Medium (Sonnet)
- **Depends on:** WI-04 for the shared JSON contract (`JobSpec`/`Result`); the script itself can be authored and tested in parallel with system Python+OpenCV.
- **Area:** CV / Python (script shipped inside the `cvruntime` bundle)
- **Requirements:** REQ-048 (esp. BR 1–4, 7), REQ-047 (sweep ordering), REQ-005 (output shape)
- **Architecture:** `docs/architecture.md` §3.23 (pipeline), §3.23.1 (runtime)

## Context

A capture sweep (REQ-047) lights one bulb at a time, in index order `0 … n−1`, ~1 s each. The
operator records this from **two or more camera angles**. This script takes those videos and
produces **3D coordinates per light** in SI metres.

It is invoked as a child process by `internal/cvruntime` (WI-04). **Input/output is the JSON
contract defined in WI-04** (`JobSpec` on stdin/temp file → `Result` on stdout/result file). The
script lives in the runtime bundle source (suggested: `backend/internal/cvruntime/src/reconstruct.py`)
and uses `cv2` (`opencv-python-headless`) + `numpy` only.

## Tasks

Implement `reconstruct.py` with these stages:

1. **Read `JobSpec`**: feed paths, optional `marker` (dictionary + printed `edge_length_m`),
   optional `scale_hint_m`, and `dwell_ms`.
2. **Per-feed 2D blink detection:** For each video, find the **single bright blob** that is lit in
   each dwell window and record `(image_x, image_y)` plus the **ordinal** of that blink. Because the
   sweep lights one index at a time, the ordered sequence of detected blinks yields
   `ordinal → light_index` (ordinal `k` → light `k`, per REQ-047 / REQ-048 BR 2). Be robust to:
   - cameras starting/stopping at different times (align by the **sequence** of blinks, not wall clock);
   - frame-rate differences (window by detected on/off transitions, not a fixed frame count);
   - brief occlusions (a light missing in one feed is allowed if present in ≥ 2 feeds total).
   Downscale frames as needed for performance (the target includes a Raspberry Pi 4 — REQ-003).
3. **Camera pose / calibration:** Estimate each camera's pose. When fiducial **markers** are visible
   (ArUco/ChArUco via `cv2.aruco`), use them to recover pose, align feeds to a common frame, and set
   **metric scale** from the known `edge_length_m` (REQ-048 BR 4). Without markers, fall back to a
   documented relative-pose estimate and use `scale_hint_m` or a documented default for scale.
4. **Triangulation:** For each light index seen in **≥ 2** feeds, triangulate the 2D detections into a
   3D point (`cv2.triangulatePoints` / least-squares); optionally refine. Output metres.
5. **Report, never fabricate (REQ-048 BR 7):** indices not triangulable → `missing`; weak/uncertain
   → `low_confidence`. Do **not** invent coordinates for undetected lights.
6. **Emit `Result`** exactly per the WI-04 contract: `light_count` = highest detected index + 1 (or
   the sweep length if derivable), `lights` with `id` `0..n-1` for the ones you have, `missing`,
   `low_confidence`, and `status`/`error`. Output ids must be the **integer light indices**, suitable
   for REQ-005 validation downstream (WI-06 re-validates).

Keep the script's only external deps `cv2` + `numpy`. Log diagnostics to **stderr** (Go captures it);
put **only** the JSON `Result` on the agreed stdout/result channel.

## Acceptance / tests

- Provide a small **synthetic fixture generator** (Python, dev-only) that renders ≥ 2 short clips of a
  known 3D point set being swept (project known 3D points to two virtual cameras, blink them in order,
  optionally draw an ArUco marker). Add tests that run `reconstruct.py` on the synthetic clips and
  assert recovered coordinates are within a tolerance of ground truth, ordering is correct, and
  occluded points land in `missing`/`low_confidence`.
- Document how to run the script standalone for development (system `python3 -m venv` +
  `pip install opencv-python-headless numpy`), independent of the Go bundle.
- Keep fixtures tiny so tests stay fast.

## Out of scope

- Bundling/packaging the runtime (WI-04). Go orchestration and HTTP (WI-06).

## Definition of done

`reconstruct.py` turns ≥ 2 synthetic feeds of a swept light set into correctly-ordered 3D
coordinates within tolerance, reports undetected lights instead of fabricating them, uses markers
for scale when present, and conforms to the WI-04 JSON contract; covered by tests on synthetic data.
