#!/usr/bin/env python3
"""
reconstruct.py — DLM light-position reconstruction pipeline.

Reads a JobSpec JSON from the path given as sys.argv[1], writes a Result
JSON to stdout.  All diagnostic output goes to stderr.

External deps: cv2 (opencv-python-headless) + numpy only.

Standalone usage (dev):
    python3 reconstruct.py <spec_file.json>

JobSpec (stdin/temp file):
    {
        "feeds":        [{"path": "..."}, ...],
        "marker":       {"dictionary": "DICT_4X4_50", "edge_length_m": 0.05},  # optional
        "scale_hint_m": 0.30,   # optional — actual camera baseline in metres
        "dwell_ms":     1000
    }

Result (stdout):
    {
        "status":         "succeeded" | "failed",
        "light_count":    <int>,
        "lights":         [{"id": <int>, "x": <m>, "y": <m>, "z": <m>}, ...],
        "missing":        [<int>, ...],
        "low_confidence": [<int>, ...],
        "error":          <str> | null
    }
"""

import json
import math
import sys
import traceback
from collections import Counter
from typing import Optional

import cv2
import numpy as np


# ──────────────────────────────────────────────────────────────────────────────
# Tunables
# ──────────────────────────────────────────────────────────────────────────────

# Maximum length of the shorter frame dimension used during processing.
# Keeps CPU load manageable on a Raspberry Pi 4 (REQ-003).
MAX_PROC_DIM = 480

# After background subtraction, a pixel must exceed this to be considered lit.
BLOB_THRESHOLD = 30

# Minimum blob area (at processing resolution) to be treated as a real light.
MIN_BLOB_AREA = 4

# Reprojection error (pixels, full resolution) above which a point is flagged
# as low_confidence.
LOW_CONF_REPROJ_PX = 4.0

# Minimum camera-frame depth (Z_cam) for a triangulated point to be accepted as
# lying *in front of* a camera (cheirality).  A point whose depth is ≤ this in
# any contributing camera is geometrically behind/at that camera and is treated
# as missing rather than emitted as a valid coordinate.  Small positive value so
# numerical noise near a true zero does not flip the sign.
CHEIRALITY_MIN_DEPTH = 1e-6

# If no marker is visible and no scale_hint_m is provided the triangulated
# coordinates are expressed in normalised baseline units (distance between
# camera 0 and camera 1 = 1.0).  We use 1.0 so the output is at least
# self-consistent; callers should always supply scale_hint_m for real jobs.
DEFAULT_SCALE_M = 1.0

# Number of frames sampled from the video to build the background model.
BG_FRAMES = 10

# Maximum number of frames decoded during the background-sampling pass.
# Bounds CPU time on long videos while still spreading samples far enough
# apart to catch each light in its off-state (≈ 60 s at 30 fps).
BG_MAX_SCAN_FRAMES = 1800

# Stride used when the container reports no total frame count.
BG_STRIDE_FALLBACK = 30

# Seconds of video scanned when searching for an ArUco marker.
MARKER_SCAN_SECS = 5.0

# Blink dwell validation (when dwell_ms > 0): accept blinks whose on-duration is
# within this fraction of the expected dwell.  Rejects stray flashes and stuck-on
# segments outside the REQ-047 sweep window.
DWELL_MIN_FRAC = 0.4
DWELL_MAX_FRAC = 2.5


# ──────────────────────────────────────────────────────────────────────────────
# Utilities
# ──────────────────────────────────────────────────────────────────────────────

def _log(*args) -> None:
    print(*args, file=sys.stderr, flush=True)


def _downscale_factor(h: int, w: int) -> float:
    """Returns a factor ≤ 1.0 so that min(h,w)*factor ≤ MAX_PROC_DIM."""
    shorter = min(h, w)
    if shorter <= MAX_PROC_DIM:
        return 1.0
    return MAX_PROC_DIM / shorter


def _estimate_K(w: int, h: int) -> np.ndarray:
    """
    Estimates pinhole camera intrinsics from image dimensions.

    Assumes a ~60° horizontal field-of-view, which is a reasonable default for
    many wide-angle cameras used in maker/hobbyist setups.  Callers may supply
    calibration data in the spec to override this path (via ArUco).
    """
    f = float(max(w, h))
    return np.array([[f, 0.0, w / 2.0],
                     [0.0, f, h / 2.0],
                     [0.0, 0.0, 1.0]], dtype=np.float64)


def _rodrigues(rvec) -> np.ndarray:
    R, _ = cv2.Rodrigues(np.array(rvec, dtype=np.float64).ravel())
    return R


def _validate_positive_finite(value, name: str) -> Optional[str]:
    """Returns an error message when *value* is not a finite number > 0."""
    try:
        v = float(value)
    except (TypeError, ValueError):
        return f"{name} must be a finite number > 0"
    if not math.isfinite(v) or v <= 0:
        return f"{name} must be a finite number > 0"
    return None


def _accept_blink_duration(duration_s: float, dwell_ms: int) -> bool:
    """
    Returns True when *duration_s* is within the expected dwell window.

    When dwell_ms <= 0 validation is disabled (callers that omit dwell_ms still
    get the Python-side default of 1000 ms; Go forwards 1000 ms by default too).
    """
    if dwell_ms <= 0:
        return True
    expected_s = dwell_ms / 1000.0
    return (
        DWELL_MIN_FRAC * expected_s <= duration_s <= DWELL_MAX_FRAC * expected_s
    )


def _blink_duration_s(start_frame: int, end_frame_inclusive: int, fps: float) -> float:
    return (end_frame_inclusive - start_frame + 1) / fps


# ──────────────────────────────────────────────────────────────────────────────
# Stage 1 — Per-feed blink detection
# ──────────────────────────────────────────────────────────────────────────────

def detect_blinks(
    video_path: str,
    dwell_ms: int,
) -> tuple[list[tuple[float, float, float]], int, int, np.ndarray]:
    """
    Opens *video_path* and detects blink events via an on/off state machine.

    Frame-rate independence: blinks are delimited by brightness transitions,
    not by a fixed frame count, so cameras recording at different FPS or
    starting/stopping at different times all align correctly (REQ-048 BR 2).
    When *dwell_ms* > 0, on-duration must fall within
    ``[DWELL_MIN_FRAC, DWELL_MAX_FRAC] × dwell_ms`` or the blink is rejected
    (stray flashes / stuck-on segments).  Each accepted blink also records the
    *time* (seconds from the start of the clip) at which it began, so that
    which it began, so that cross-feed correspondence (`align_detections`) can
    use the sweep cadence to recover light indices robustly even when a feed
    misses or adds a blink — rather than blindly trusting positional ordinals.

    Returns
    -------
    blinks : list of (t_start_s, cx, cy)
        ``t_start_s`` is the time (in seconds, from the start of this clip) at
        which the blink turned on.  ``cx, cy`` is the image-space centroid at
        full resolution.
    frame_w, frame_h : int
        Full-resolution frame dimensions.
    K : np.ndarray
        Estimated 3×3 camera intrinsic matrix.
    """
    # ── Pass 1: metadata + background ────────────────────────────────────────
    # The background is built by sampling BG_FRAMES frames spread across the
    # full video (up to BG_MAX_SCAN_FRAMES decoded frames), then taking the
    # per-pixel minimum.  Spreading samples across the video means every light
    # appears dark in at least some samples (during other lights' blink
    # periods), so a light that is on throughout the first few consecutive
    # frames never contaminates its own background pixel.
    cap_bg = cv2.VideoCapture(video_path)
    try:
        if not cap_bg.isOpened():
            raise IOError(f"Cannot open video: {video_path!r}")

        frame_w = int(cap_bg.get(cv2.CAP_PROP_FRAME_WIDTH))
        frame_h = int(cap_bg.get(cv2.CAP_PROP_FRAME_HEIGHT))
        if frame_w == 0 or frame_h == 0:
            raise IOError(f"Video reports zero dimensions: {video_path!r}")

        # Frame rate is used only to convert blink frame indices into seconds for
        # cadence-based cross-feed alignment.  A bad/zero FPS falls back to a sane
        # default; alignment only needs a consistent per-feed time base.
        fps = cap_bg.get(cv2.CAP_PROP_FPS)
        if not fps or fps <= 0 or math.isnan(fps):
            fps = 30.0

        K = _estimate_K(frame_w, frame_h)
        scale = _downscale_factor(frame_h, frame_w)
        proc_w = max(1, int(frame_w * scale))
        proc_h = max(1, int(frame_h * scale))

        _log(f"  feed {video_path!r}: {frame_w}×{frame_h} → proc {proc_w}×{proc_h}")

        # Compute stride so that BG_FRAMES samples span the first
        # min(total_frames, BG_MAX_SCAN_FRAMES) frames of the clip.
        total_frames = int(cap_bg.get(cv2.CAP_PROP_FRAME_COUNT))
        scan_limit = (
            min(total_frames, BG_MAX_SCAN_FRAMES) if total_frames > 0
            else BG_MAX_SCAN_FRAMES
        )
        stride = max(1, scan_limit // BG_FRAMES)

        bg_acc: list[np.ndarray] = []
        fi = 0
        while len(bg_acc) < BG_FRAMES:
            ret, frame = cap_bg.read()
            if not ret:
                break
            if fi % stride == 0:
                small = cv2.resize(frame, (proc_w, proc_h), interpolation=cv2.INTER_AREA)
                bg_acc.append(cv2.cvtColor(small, cv2.COLOR_BGR2GRAY))
            fi += 1
    finally:
        cap_bg.release()

    if bg_acc:
        background = np.min(np.stack(bg_acc, axis=0), axis=0)
    else:
        background = np.zeros((proc_h, proc_w), dtype=np.uint8)

    # ── Pass 2: blink detection ───────────────────────────────────────────────
    # A fresh VideoCapture open restarts from frame 0 deterministically across
    # container backends (MP4, MKV, AVI, …).  CAP_PROP_POS_FRAMES = 0 seeking
    # is unreliable on some phone-upload containers and can misalign the blink
    # pass relative to the background frames, causing missed or spurious blinks.
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video (detection pass): {video_path!r}")

        blinks: list[tuple[float, float, float]] = []
        frame_idx = -1
        blink_start_frame = 0
        in_blink = False
        accum: list[tuple[float, float]] = []

        while True:
            ret, frame = cap.read()
            if not ret:
                break
            frame_idx += 1

            small = cv2.resize(frame, (proc_w, proc_h), interpolation=cv2.INTER_AREA)
            gray = cv2.cvtColor(small, cv2.COLOR_BGR2GRAY)
            diff = cv2.absdiff(gray, background)

            blob = _find_bright_blob(diff)

            if blob is not None:
                # Scale centroid back to full-resolution coordinates.
                cx_full = blob[0] / scale
                cy_full = blob[1] / scale
                if not in_blink:
                    in_blink = True
                    blink_start_frame = frame_idx
                    accum = [(cx_full, cy_full)]
                else:
                    accum.append((cx_full, cy_full))
            else:
                if in_blink:
                    duration_s = _blink_duration_s(
                        blink_start_frame, frame_idx - 1, fps
                    )
                    if _accept_blink_duration(duration_s, dwell_ms):
                        _finalise_blink(
                            blinks, blink_start_frame / fps, accum
                        )
                    else:
                        _log(
                            f"    rejected blink at frame {blink_start_frame}: "
                            f"duration {duration_s * 1000:.0f} ms outside "
                            f"dwell window ({dwell_ms} ms)"
                        )
                    accum = []
                    in_blink = False

        # Handle clip ending while still inside a blink.
        if in_blink and accum:
            duration_s = _blink_duration_s(blink_start_frame, frame_idx, fps)
            if _accept_blink_duration(duration_s, dwell_ms):
                _finalise_blink(blinks, blink_start_frame / fps, accum)
            else:
                _log(
                    f"    rejected trailing blink at frame {blink_start_frame}: "
                    f"duration {duration_s * 1000:.0f} ms outside "
                    f"dwell window ({dwell_ms} ms)"
                )
    finally:
        cap.release()
    _log(f"    → {len(blinks)} blink(s) detected")
    return blinks, frame_w, frame_h, K


def _find_bright_blob(
    diff_gray: np.ndarray,
) -> Optional[tuple[float, float]]:
    """Returns (cx, cy) centroid of the largest blob above threshold, or None."""
    _, binary = cv2.threshold(diff_gray, BLOB_THRESHOLD, 255, cv2.THRESH_BINARY)
    n, _, stats, centroids = cv2.connectedComponentsWithStats(
        binary, connectivity=8
    )
    best_area = MIN_BLOB_AREA - 1
    best: Optional[tuple[float, float]] = None
    for i in range(1, n):  # label 0 = background
        area = int(stats[i, cv2.CC_STAT_AREA])
        if area > best_area:
            best_area = area
            best = (float(centroids[i][0]), float(centroids[i][1]))
    return best


def _finalise_blink(
    blinks: list[tuple[float, float, float]],
    t_start_s: float,
    accum: list[tuple[float, float]],
) -> None:
    cx = sum(p[0] for p in accum) / len(accum)
    cy = sum(p[1] for p in accum) / len(accum)
    blinks.append((t_start_s, cx, cy))


# ──────────────────────────────────────────────────────────────────────────────
# Stage 2 — Align detections across feeds
# ──────────────────────────────────────────────────────────────────────────────

def align_detections(
    feed_blinks: list[list[tuple[float, float, float]]],
) -> dict[int, list[tuple[int, float, float]]]:
    """
    Merges per-feed blink lists into:
        light_index → [(feed_idx, cx, cy), …]

    Per REQ-047 / REQ-048 BR 2, the *k-th dwell window* of the capture sweep
    corresponds to light index ``k``.  The sweep lights each light for a fixed
    dwell and steps through them in order, so blinks arrive at a steady
    **cadence**.  Rather than blindly trusting each feed's positional ordinal
    (which silently drifts when a feed misses or adds a blink — occlusion,
    noise, thresholding), we recover each blink's *slot* (dwell-window index)
    from its **timing** relative to that cadence:

      * A **missed** blink leaves a ~2× cadence gap → the slot is skipped, so
        that light simply has one fewer view in this feed (it is not back-filled
        by a later blink).  Downstream lights keep their correct indices.
      * A **spurious** blink lands off-cadence and collides with a neighbouring
        slot → that slot becomes *ambiguous* for the feed and is dropped (so the
        feed contributes no detection there) rather than shifting every later
        index by one.

    Slot ``k`` is treated as light index ``k``; the first detected blink of
    each feed anchors slot 0 (the sweep is recorded from its start, REQ-047).
    A light only receives coordinates from 2D points that correspond to the
    **same** slot across feeds, so a drop/spurious in one feed can never
    mislabel another light (REQ-048 BR 2 / BR 7).
    """
    period = _estimate_period(feed_blinks)
    counts = [len(b) for b in feed_blinks]

    by_light: dict[int, list[tuple[int, float, float]]] = {}
    for feed_idx, blinks in enumerate(feed_blinks):
        if not blinks:
            _log(f"  feed {feed_idx}: 0 blinks")
            continue

        slots = _assign_slots([b[0] for b in blinks], period)

        # A feed must not contribute two blinks to the same slot; if it does
        # (e.g. a spurious blink collided with a real one) the slot is ambiguous
        # for this feed and is excluded — better one fewer view than a wrong one.
        ambiguous = {s for s, c in Counter(slots).items() if c > 1}

        for (_t, cx, cy), slot in zip(blinks, slots):
            if slot in ambiguous:
                continue
            by_light.setdefault(slot, []).append((feed_idx, cx, cy))

        note = f" (ambiguous slots dropped: {sorted(ambiguous)})" if ambiguous else ""
        _log(f"  feed {feed_idx}: {len(blinks)} blink(s) → slots {slots}{note}")

    if len(set(counts)) > 1:
        _log(
            f"  ⚠ feeds disagree on blink count {counts} "
            f"(cadence≈{period:.3f}s); using timing-based slot alignment so a "
            f"missed/spurious blink does not shift other light indices"
        )
    return by_light


def _estimate_period(
    feed_blinks: list[list[tuple[float, float, float]]],
) -> float:
    """
    Estimates the sweep cadence (seconds between consecutive dwell windows) as
    the **median** of all consecutive blink-start intervals pooled across feeds.

    The median is robust to the minority of intervals that are off-cadence: a
    dropped blink produces a ~2× interval and a spurious blink a short one, but
    the bulk of intervals equal one dwell+gap period.  Returns 0.0 when there is
    not enough data (fewer than one interval anywhere), signalling callers to
    fall back to positional ordering.
    """
    intervals: list[float] = []
    for blinks in feed_blinks:
        times = sorted(b[0] for b in blinks)
        intervals.extend(
            times[i + 1] - times[i] for i in range(len(times) - 1)
        )
    intervals = [d for d in intervals if d > 1e-6]
    if not intervals:
        return 0.0
    intervals.sort()
    mid = len(intervals) // 2
    if len(intervals) % 2:
        return intervals[mid]
    return 0.5 * (intervals[mid - 1] + intervals[mid])


def _assign_slots(times: list[float], period: float) -> list[int]:
    """
    Maps each blink-start *time* (chronological, seconds) to a 0-based dwell
    slot index using the sweep *period*.

    The first blink anchors slot 0.  Each subsequent blink advances the slot by
    the number of whole cadence periods that have elapsed since the last blink
    accepted into a *new* slot (``round(Δt / period)``):

      * ``steps >= 2`` → one or more skipped slots (missed blinks); those light
        indices are simply absent for this feed.
      * ``steps == 0`` → the blink falls inside the current slot's window
        (a spurious/duplicate detection); it is tagged with the current slot
        (making the slot ambiguous) and the cadence anchor is **not** advanced,
        so the next genuine blink is still measured from the real slot boundary.

    When *period* is non-positive (insufficient timing data) this degrades to
    plain positional indexing — identical to the legacy behaviour.
    """
    if not times:
        return []
    if period <= 0:
        return list(range(len(times)))

    slots = [0]
    slot = 0
    anchor = times[0]
    for t in times[1:]:
        steps = int(round((t - anchor) / period))
        if steps < 1:
            # Off-cadence extra blink within the current window: keep it in the
            # current slot (flagged ambiguous downstream) without moving anchor.
            slots.append(slot)
            continue
        slot += steps
        slots.append(slot)
        anchor = t
    return slots


# ──────────────────────────────────────────────────────────────────────────────
# Stage 3 — Camera pose estimation
# ──────────────────────────────────────────────────────────────────────────────

def estimate_poses(
    feed_paths: list[str],
    Ks: list[np.ndarray],
    marker_spec: Optional[dict],
    by_light: dict[int, list[tuple[int, float, float]]],
    scale_hint_m: Optional[float],
) -> tuple[list[Optional[tuple[np.ndarray, np.ndarray]]], float]:
    """
    Estimates extrinsic camera poses and the metric scale.

    Returns
    -------
    poses : list of (R, t) | None per feed
        Camera 0 is always at the world origin (R=I, t=0).
        R (3×3) and t (3×1) transform world points into camera frame:
            x_cam = R @ x_world + t
    metric_scale : float
        Multiplier applied to triangulated coordinates to convert them into
        SI metres.  When ArUco markers localise all cameras the poses are
        already in metres (solvePnP object points are defined in metres), so
        DLT output is already metric and this is 1.0.  When using scale_hint_m
        it is the caller-supplied baseline; otherwise 1.0 (output is in
        normalised baseline units).
    """
    n = len(feed_paths)
    poses: list[Optional[tuple[np.ndarray, np.ndarray]]] = [None] * n
    poses[0] = (np.eye(3, dtype=np.float64), np.zeros((3, 1), dtype=np.float64))

    # --- ArUco path (REQ-048 BR 4) ---
    if marker_spec is not None:
        _log("  trying ArUco marker pose estimation …")
        try:
            aruco_poses = _poses_from_aruco(feed_paths, Ks, marker_spec)
            n_found = sum(1 for p in aruco_poses if p is not None)
            if n_found == n:
                _log(f"  ArUco: all {n} cameras localised")
                # solvePnP uses object points defined in metres, so the
                # resulting poses are already metric.  DLT output is therefore
                # already in metres; no additional scaling is needed.
                return aruco_poses, 1.0
            else:
                _log(f"  ArUco: only {n_found}/{n} cameras; falling back to E-matrix")
        except Exception as exc:
            _log(f"  ArUco failed ({exc}); falling back to E-matrix")

    # --- Essential-matrix path (fallback) ---
    if n < 2:
        _log("  single feed: returning identity pose (no triangulation possible)")
        return poses, scale_hint_m or DEFAULT_SCALE_M

    pts0, pts1 = _common_2d_points(by_light, 0, 1)
    if len(pts0) < 5:
        _log(f"  only {len(pts0)} common points between feed 0 and 1 — skipping pose est.")
        return poses, scale_hint_m or DEFAULT_SCALE_M

    em_pose = _essential_matrix_pose(pts0, pts1, Ks[0], Ks[1], by_light)
    if em_pose is None:
        _log("  essential matrix estimation failed")
        return poses, scale_hint_m or DEFAULT_SCALE_M

    R1, t1, n_inliers = em_pose
    poses[1] = (R1, t1)
    _log(f"  E-matrix: feed 1 recovered, {n_inliers} inliers")

    # For feeds beyond the first two, PnP against points triangulated from 0+1.
    if n > 2:
        ref3d, ref_ids = _quick_triangulate_pair(by_light, poses, Ks, 0, 1)
        for fi in range(2, n):
            p = _pnp_pose(ref3d, ref_ids, by_light, fi, Ks[fi])
            if p is not None:
                poses[fi] = p
                _log(f"  PnP: feed {fi} localised")
            else:
                _log(f"  PnP: feed {fi} failed — skipping")

    return poses, scale_hint_m or DEFAULT_SCALE_M


# --- ArUco helpers -----------------------------------------------------------

def _aruco_dict(name: str):
    val = getattr(cv2.aruco, name, None)
    if val is None:
        raise ValueError(f"Unknown ArUco dictionary name: {name!r}")
    return cv2.aruco.getPredefinedDictionary(val)


def _poses_from_aruco(
    feed_paths: list[str],
    Ks: list[np.ndarray],
    marker_spec: dict,
) -> list[Optional[tuple[np.ndarray, np.ndarray]]]:
    adict = _aruco_dict(marker_spec["dictionary"])
    params = cv2.aruco.DetectorParameters()
    try:
        detector = cv2.aruco.ArucoDetector(adict, params)
        def _detect(gray):
            corners, ids, _ = detector.detectMarkers(gray)
            return corners, ids
    except AttributeError:
        # OpenCV < 4.7 fallback
        def _detect(gray):
            corners, ids, _ = cv2.aruco.detectMarkers(gray, adict, parameters=params)
            return corners, ids

    edge_m = float(marker_spec["edge_length_m"])
    dist = np.zeros((4, 1), dtype=np.float64)
    poses = []
    for fi, (path, K) in enumerate(zip(feed_paths, Ks)):
        pose = _aruco_pose_single_feed(path, K, dist, _detect, edge_m, fi)
        poses.append(pose)
    return poses


def _estimate_pose_single(corners, edge_m: float, K: np.ndarray, dist: np.ndarray):
    """
    Compatibility shim for estimatePoseSingleMarkers, which was removed in
    OpenCV 4.8+.  Falls back to solvePnP with the standard ArUco 3-D object
    points.
    """
    try:
        return cv2.aruco.estimatePoseSingleMarkers(corners, edge_m, K, dist)
    except AttributeError:
        pass
    # Manual implementation using solvePnP.
    half = edge_m / 2.0
    obj_pts = np.array([
        [-half,  half, 0.0],
        [ half,  half, 0.0],
        [ half, -half, 0.0],
        [-half, -half, 0.0],
    ], dtype=np.float64)
    rvecs, tvecs = [], []
    for c in corners:
        img_pts = c.reshape(4, 2).astype(np.float64)
        ok, rvec, tvec = cv2.solvePnP(obj_pts, img_pts, K, dist,
                                       flags=cv2.SOLVEPNP_IPPE_SQUARE)
        if not ok:
            raise RuntimeError("solvePnP failed for ArUco corner")
        rvecs.append(rvec)
        tvecs.append(tvec)
    return np.array(rvecs), np.array(tvecs), None


def _aruco_pose_single_feed(
    video_path: str,
    K: np.ndarray,
    dist: np.ndarray,
    detect_fn,
    edge_m: float,
    feed_idx: int,
) -> Optional[tuple[np.ndarray, np.ndarray]]:
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            return None
        fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
        max_frames = max(1, int(MARKER_SCAN_SECS * fps))

        pose: Optional[tuple[np.ndarray, np.ndarray]] = None
        for _ in range(max_frames):
            ret, frame = cap.read()
            if not ret:
                break
            gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            corners, ids = detect_fn(gray)
            if ids is not None and len(ids) > 0:
                rvecs, tvecs, _ = _estimate_pose_single(
                    corners[:1], edge_m, K, dist
                )
                R = _rodrigues(rvecs[0])
                t = tvecs[0].reshape(3, 1)
                pose = (R, t)
                _log(f"    feed {feed_idx}: marker found")
                break
    finally:
        cap.release()

    if pose is None:
        _log(f"    feed {feed_idx}: no marker in first {MARKER_SCAN_SECS:.0f}s")
    return pose


# --- Essential-matrix helpers ------------------------------------------------

def _normalize_image_points(
    pts: np.ndarray,
    K: np.ndarray,
) -> np.ndarray:
    """
    Map pixel coordinates to normalized camera coordinates (K applied).

    Used so findEssentialMat / recoverPose can relate two feeds with different
    intrinsics without forcing both through one camera matrix.
    """
    return cv2.undistortPoints(
        pts.reshape(-1, 1, 2).astype(np.float64),
        K,
        None,
        P=np.eye(3, dtype=np.float64),
    ).reshape(-1, 2).astype(np.float32)


def _essential_matrix_pose(
    pts0: np.ndarray,
    pts1: np.ndarray,
    K0: np.ndarray,
    K1: np.ndarray,
    by_light: dict[int, list[tuple[int, float, float]]],
) -> Optional[tuple[np.ndarray, np.ndarray, int]]:
    """
    Recover the relative pose of camera 1 w.r.t. camera 0 from matched pixels.

    Each feed's points are normalized with its own intrinsics.  OpenCV may return
    several E-matrix hypotheses; each is decomposed and ranked by cheirality
    (fewest missing lights) then mean reprojection error on the pair.
    """
    if len(pts0) < 5:
        return None

    pts0_n = _normalize_image_points(pts0, K0)
    pts1_n = _normalize_image_points(pts1, K1)
    I = np.eye(3, dtype=np.float64)
    norm_thresh = 1.0 / float(max(K0[0, 0], K1[0, 0], 1.0))

    candidates: list[tuple[tuple[int, float, int], int, np.ndarray, np.ndarray]] = []

    for pa, pb, invert in ((pts0_n, pts1_n, False), (pts1_n, pts0_n, True)):
        E, mask = cv2.findEssentialMat(
            pa, pb, I, method=cv2.RANSAC, prob=0.999, threshold=norm_thresh
        )
        if E is None:
            continue
        inlier_mask = mask.ravel() == 1
        n_inliers = int(inlier_mask.sum())
        if n_inliers < 5:
            continue
        pa_inl = pa[inlier_mask]
        pb_inl = pb[inlier_mask]

        for ci in range(E.shape[0] // 3):
            Ei = E[ci * 3:(ci + 1) * 3, :]
            _, R, t, _ = cv2.recoverPose(Ei, pa_inl, pb_inl, I)
            if invert:
                R1 = R.T
                t1 = (-R.T @ t.reshape(3, 1)).reshape(3)
            else:
                R1 = R
                t1 = t.reshape(3)
            score = _pose_pair_score(R1, t1, by_light, K0, K1)
            candidates.append((score, n_inliers, R1, t1))

    if not candidates:
        return None

    candidates.sort(key=lambda c: (c[0][0], c[0][1], -c[0][2], -c[1]))
    _, n_inliers, R1, t1 = candidates[0]
    return R1, t1.reshape(3, 1), n_inliers


def _pose_pair_score(
    R1: np.ndarray,
    t1: np.ndarray,
    by_light: dict[int, list[tuple[int, float, float]]],
    K0: np.ndarray,
    K1: np.ndarray,
) -> tuple[int, float, int]:
    """
    Rank an essential-matrix pose hypothesis.

    Returns (n_missing, mean_reproj_px, n_triangulated) — lower is better for
    the first two; higher is better for the third (tie-break).
    """
    poses: list[Optional[tuple[np.ndarray, np.ndarray]]] = [
        (np.eye(3, dtype=np.float64), np.zeros((3, 1), dtype=np.float64)),
        (R1, t1.reshape(3, 1)),
    ]
    ref3d, ref_ids = _quick_triangulate_pair(by_light, poses, [K0, K1], 0, 1)
    if len(ref_ids) < 3:
        return (999, 999.0, 0)

    Ps = {
        0: K0 @ np.hstack([np.eye(3), np.zeros((3, 1))]),
        1: K1 @ np.hstack([R1, t1.reshape(3, 1)]),
    }
    errs: list[float] = []
    for i, lid in enumerate(ref_ids):
        pt = ref3d[i]
        for fi in (0, 1):
            det = next((d for d in by_light.get(lid, []) if d[0] == fi), None)
            if det is None:
                continue
            cx, cy = det[1], det[2]
            proj = Ps[fi] @ np.append(pt, 1.0)
            if abs(proj[2]) < 1e-10:
                continue
            errs.append(math.hypot(proj[0] / proj[2] - cx, proj[1] / proj[2] - cy))

    mean_reproj = sum(errs) / len(errs) if errs else 999.0
    lights, miss, _ = triangulate_all(by_light, poses, [K0, K1], 1.0)
    return (len(miss), mean_reproj, len(lights))


def _common_2d_points(
    by_light: dict[int, list[tuple[int, float, float]]],
    feed_a: int,
    feed_b: int,
) -> tuple[np.ndarray, np.ndarray]:
    pts_a, pts_b = [], []
    for dets in by_light.values():
        da = next((d for d in dets if d[0] == feed_a), None)
        db = next((d for d in dets if d[0] == feed_b), None)
        if da and db:
            pts_a.append([da[1], da[2]])
            pts_b.append([db[1], db[2]])
    return (
        np.array(pts_a, dtype=np.float32),
        np.array(pts_b, dtype=np.float32),
    )


def _quick_triangulate_pair(
    by_light: dict[int, list[tuple[int, float, float]]],
    poses: list[Optional[tuple[np.ndarray, np.ndarray]]],
    Ks: list[np.ndarray],
    fi: int,
    fj: int,
) -> tuple[np.ndarray, list[int]]:
    """Returns (pts3d, light_ids) for lights visible in both fi and fj."""
    Ri, ti = poses[fi]  # type: ignore[misc]
    Rj, tj = poses[fj]  # type: ignore[misc]
    Pi = Ks[fi] @ np.hstack([Ri, ti])
    Pj = Ks[fj] @ np.hstack([Rj, tj])

    pts_i, pts_j, ids = [], [], []
    for lid, dets in sorted(by_light.items()):
        di = next((d for d in dets if d[0] == fi), None)
        dj = next((d for d in dets if d[0] == fj), None)
        if di and dj:
            pts_i.append([di[1], di[2]])
            pts_j.append([dj[1], dj[2]])
            ids.append(lid)

    if not ids:
        return np.zeros((0, 3)), []

    pts4d = cv2.triangulatePoints(
        Pi, Pj,
        np.array(pts_i, dtype=np.float64).T,
        np.array(pts_j, dtype=np.float64).T,
    )
    pts3d = (pts4d[:3] / pts4d[3]).T  # (N, 3)
    return pts3d, ids


def _pnp_pose(
    ref3d: np.ndarray,
    ref_ids: list[int],
    by_light: dict[int, list[tuple[int, float, float]]],
    feed_idx: int,
    K: np.ndarray,
) -> Optional[tuple[np.ndarray, np.ndarray]]:
    obj_pts, img_pts = [], []
    for i, lid in enumerate(ref_ids):
        d = next((x for x in by_light.get(lid, []) if x[0] == feed_idx), None)
        if d:
            obj_pts.append(ref3d[i])
            img_pts.append([d[1], d[2]])
    if len(obj_pts) < 4:
        return None
    ok, rvec, tvec, _ = cv2.solvePnPRansac(
        np.array(obj_pts, dtype=np.float64),
        np.array(img_pts, dtype=np.float64),
        K,
        np.zeros((4, 1)),
    )
    if not ok:
        return None
    return _rodrigues(rvec), tvec.reshape(3, 1)


# ──────────────────────────────────────────────────────────────────────────────
# Stage 4 — Triangulation
# ──────────────────────────────────────────────────────────────────────────────

def triangulate_all(
    by_light: dict[int, list[tuple[int, float, float]]],
    poses: list[Optional[tuple[np.ndarray, np.ndarray]]],
    Ks: list[np.ndarray],
    metric_scale: float,
) -> tuple[dict[int, tuple[float, float, float]], list[int], list[int]]:
    """
    Triangulates 3D coordinates for every light visible in ≥ 2 feeds.

    Lights visible in < 2 feeds are placed in *missing* (REQ-048 BR 7).
    Sanity guards (WI-26) also route a light to *missing* when its triangulated
    point fails **cheirality** (not in front of every contributing camera,
    ``Z_cam > 0``) or is **non-finite** (NaN/Inf) — bad geometry, degenerate
    correspondences or ill-conditioned poses must never be emitted as valid
    coordinates.  These are stronger than the reprojection-error check:
    lights with reprojection error > LOW_CONF_REPROJ_PX still get coordinates
    but go to *low_confidence*.
    Coordinates never fabricated for undetected lights.

    Returns
    -------
    lights_3d : dict  light_id → (x_m, y_m, z_m)
    missing   : list  light ids that could not be triangulated
    low_conf  : list  light ids with high reprojection error
    """
    # Build projection matrices for feeds with known poses.
    Ps: dict[int, np.ndarray] = {}
    for i, pose in enumerate(poses):
        if pose is not None:
            R, t = pose
            Ps[i] = Ks[i] @ np.hstack([R, t])

    lights_3d: dict[int, tuple[float, float, float]] = {}
    missing:   list[int] = []
    low_conf:  list[int] = []

    for light_id, dets in sorted(by_light.items()):
        valid = [(fi, cx, cy) for fi, cx, cy in dets if fi in Ps]
        if len(valid) < 2:
            _log(f"  light {light_id}: {len(valid)} usable view(s) → missing")
            missing.append(light_id)
            continue

        # DLT triangulation across all valid views (≥ 2).
        pt3d = _dlt_triangulate(valid, Ps)
        if pt3d is None:
            _log(f"  light {light_id}: degenerate triangulation → missing")
            missing.append(light_id)
            continue

        # Finite guard (geometry): an ill-conditioned SVD can yield NaN/Inf.
        # Reject before any further math so depth/reprojection stay meaningful.
        if not np.all(np.isfinite(pt3d)):
            _log(f"  light {light_id}: non-finite triangulation → missing")
            missing.append(light_id)
            continue

        # Cheirality: the point must lie in front of *every* contributing
        # camera.  A negative/zero depth means bad geometry produced a point
        # behind a camera — never emit it as a valid coordinate (WI-26).
        bad_depth = _behind_camera_depth(pt3d, valid, Ps)
        if bad_depth is not None:
            fi, depth = bad_depth
            _log(
                f"  light {light_id}: depth {depth:.4g} ≤ 0 in feed {fi} "
                f"(behind camera) → missing"
            )
            missing.append(light_id)
            continue

        x, y, z = pt3d * metric_scale

        # Finite guard: reject any NaN/Inf coordinate (ill-conditioned SVD or a
        # non-finite scale).  Routing to missing keeps the output valid JSON so
        # the Go side never fails to parse the whole job (WI-26).
        if not (math.isfinite(x) and math.isfinite(y) and math.isfinite(z)):
            _log(f"  light {light_id}: non-finite coordinate → missing")
            missing.append(light_id)
            continue

        reproj = _reprojection_error(pt3d, valid, Ps)
        if reproj > LOW_CONF_REPROJ_PX:
            _log(f"  light {light_id}: reproj={reproj:.1f}px → low_confidence")
            low_conf.append(light_id)

        lights_3d[light_id] = (float(x), float(y), float(z))

    return lights_3d, missing, low_conf


def _dlt_triangulate(
    valid: list[tuple[int, float, float]],
    Ps: dict[int, np.ndarray],
) -> Optional[np.ndarray]:
    """
    Linear (DLT) triangulation from ≥ 2 views.
    Returns a (3,) world point, or None on failure.
    """
    rows = []
    for fi, cx, cy in valid:
        P = Ps[fi]
        rows.append(cx * P[2] - P[0])
        rows.append(cy * P[2] - P[1])
    A = np.array(rows, dtype=np.float64)
    _, _, Vt = np.linalg.svd(A)
    X = Vt[-1]
    if abs(X[3]) < 1e-10:
        return None
    return X[:3] / X[3]


def _behind_camera_depth(
    pt3d: np.ndarray,
    valid: list[tuple[int, float, float]],
    Ps: dict[int, np.ndarray],
) -> Optional[tuple[int, float]]:
    """
    Cheirality check.  Returns ``(feed_idx, depth)`` for the first contributing
    camera in which *pt3d* does **not** lie in front (camera-frame depth
    ``Z_cam ≤ CHEIRALITY_MIN_DEPTH`` or non-finite), or ``None`` when the point
    is in front of every camera.

    The third row of a projection matrix ``P = K · [R | t]`` equals the third
    row of ``[R | t]`` (since ``K``'s last row is ``[0, 0, 1]``), so for a world
    point ``X`` the homogeneous product ``(P · [X, 1])[2]`` is exactly the
    point's depth in that camera's frame — no separate pose math is needed.
    """
    pt_h = np.append(pt3d, 1.0)
    for fi, _cx, _cy in valid:
        depth = float((Ps[fi] @ pt_h)[2])
        if not math.isfinite(depth) or depth <= CHEIRALITY_MIN_DEPTH:
            return fi, depth
    return None


def _reprojection_error(
    pt3d: np.ndarray,
    valid: list[tuple[int, float, float]],
    Ps: dict[int, np.ndarray],
) -> float:
    errs = []
    for fi, cx, cy in valid:
        proj = Ps[fi] @ np.append(pt3d, 1.0)
        if abs(proj[2]) < 1e-10:
            continue
        errs.append(math.hypot(proj[0] / proj[2] - cx, proj[1] / proj[2] - cy))
    return max(errs) if errs else 0.0


# ──────────────────────────────────────────────────────────────────────────────
# Stage 5 — Emit Result
# ──────────────────────────────────────────────────────────────────────────────

def _make_result(
    lights_3d: dict[int, tuple[float, float, float]],
    missing: list[int],
    low_conf: list[int],
    all_ids: set[int],
) -> dict:
    light_count = max(all_ids) + 1 if all_ids else 0
    return {
        "status":         "succeeded",
        "light_count":    light_count,
        "lights":         [
            {"id": lid, "x": x, "y": y, "z": z}
            for lid, (x, y, z) in sorted(lights_3d.items())
        ],
        "missing":        sorted(missing),
        "low_confidence": sorted(set(low_conf)),
        "error":          None,
    }


def _make_error(message: str) -> dict:
    return {
        "status":         "failed",
        "light_count":    0,
        "lights":         [],
        "missing":        [],
        "low_confidence": [],
        "error":          message,
    }


def _serialise_result(result: dict) -> str:
    """
    Serialise *result* to RFC8259-valid JSON.

    ``allow_nan=False`` guarantees the output never contains the non-standard
    ``NaN``/``Infinity`` tokens that Python's ``json.dumps`` emits by default —
    those tokens make the Go side's ``json.Unmarshal`` fail the *whole* job.
    The per-light finite guard in ``triangulate_all`` should already keep
    non-finite values out of the result; this is the final safety net.  If one
    somehow survives we fail cleanly with ``status:"failed"`` rather than
    emitting output Go cannot parse.
    """
    try:
        return json.dumps(result, allow_nan=False)
    except ValueError as exc:
        return json.dumps(
            _make_error(f"non-finite value in reconstruction result: {exc}")
        )


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

def _emit_failure(message: str) -> None:
    """Write a failed Result JSON to stdout and exit non-zero."""
    print(json.dumps(_make_error(message)))
    sys.exit(1)


def main() -> None:
    if len(sys.argv) < 2:
        _emit_failure("usage: reconstruct.py <spec_file.json>")

    try:
        with open(sys.argv[1]) as fh:
            spec = json.load(fh)
    except Exception as exc:
        _emit_failure(f"cannot read spec: {exc}")

    feeds = spec.get("feeds", [])
    if len(feeds) < 2:
        _emit_failure(
            f"at least 2 feeds are required for triangulation; got {len(feeds)}"
        )

    dwell_ms      = int(spec.get("dwell_ms", 1000))
    marker_spec   = spec.get("marker")
    scale_hint_m  = spec.get("scale_hint_m")

    if scale_hint_m is not None:
        scale_err = _validate_positive_finite(scale_hint_m, "scale_hint_m")
        if scale_err:
            _emit_failure(scale_err)

    if marker_spec is not None:
        edge_err = _validate_positive_finite(
            marker_spec.get("edge_length_m"), "marker.edge_length_m"
        )
        if edge_err:
            _emit_failure(edge_err)

    try:
        _log(f"DLM reconstruct: {len(feeds)} feed(s), dwell={dwell_ms} ms")

        # Stage 1 — blink detection.
        feed_blinks: list[list[tuple[float, float, float]]] = []
        Ks: list[np.ndarray] = []
        for feed in feeds:
            blinks, _fw, _fh, K = detect_blinks(feed["path"], dwell_ms)
            feed_blinks.append(blinks)
            Ks.append(K)

        # Stage 2 — align across feeds.
        by_light = align_detections(feed_blinks)
        all_ids  = set(by_light.keys())
        _log(f"  detected light indices: {sorted(all_ids) if all_ids else '(none)'}")

        if not all_ids:
            _emit_failure("no blink events detected in any feed")

        # Stage 3 — camera pose.
        poses, metric_scale = estimate_poses(
            [f["path"] for f in feeds],
            Ks,
            marker_spec,
            by_light,
            scale_hint_m,
        )

        # Stage 4 — triangulation.
        lights_3d, missing, low_conf = triangulate_all(
            by_light, poses, Ks, metric_scale
        )

        # Any integer in [0, max_id] that was never seen in any feed is missing.
        max_id = max(all_ids)
        for lid in range(max_id + 1):
            if lid not in all_ids and lid not in missing:
                missing.append(lid)

        result = _make_result(lights_3d, missing, low_conf, all_ids)
        _log(
            f"  done: {len(lights_3d)} triangulated, "
            f"{len(missing)} missing, {len(low_conf)} low_confidence"
        )
        print(_serialise_result(result))

    except Exception as exc:
        _log(f"ERROR: {exc}")
        _log(traceback.format_exc())
        print(json.dumps(_make_error(str(exc))))
        sys.exit(1)


if __name__ == "__main__":
    main()
