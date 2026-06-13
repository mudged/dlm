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

# If no marker is visible and no scale_hint_m is provided the triangulated
# coordinates are expressed in normalised baseline units (distance between
# camera 0 and camera 1 = 1.0).  We use 1.0 so the output is at least
# self-consistent; callers should always supply scale_hint_m for real jobs.
DEFAULT_SCALE_M = 1.0

# Frames read at the start of each feed to build the background model.
BG_FRAMES = 10

# Seconds of video scanned when searching for an ArUco marker.
MARKER_SCAN_SECS = 5.0


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


# ──────────────────────────────────────────────────────────────────────────────
# Stage 1 — Per-feed blink detection
# ──────────────────────────────────────────────────────────────────────────────

def detect_blinks(
    video_path: str,
    dwell_ms: int,
) -> tuple[list[tuple[int, float, float]], int, int, np.ndarray]:
    """
    Opens *video_path* and detects blink events via an on/off state machine.

    Frame-rate independence: blinks are delimited by brightness transitions,
    not by a fixed frame count, so cameras recording at different FPS or
    starting/stopping at different times all align correctly (REQ-048 BR 2).

    Returns
    -------
    blinks : list of (ordinal, cx, cy)
        Image-space centroid at full resolution; 0-based ordinals.
    frame_w, frame_h : int
        Full-resolution frame dimensions.
    K : np.ndarray
        Estimated 3×3 camera intrinsic matrix.
    """
    cap = cv2.VideoCapture(video_path)
    if not cap.isOpened():
        raise IOError(f"Cannot open video: {video_path!r}")

    frame_w = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    frame_h = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    if frame_w == 0 or frame_h == 0:
        cap.release()
        raise IOError(f"Video reports zero dimensions: {video_path!r}")

    K = _estimate_K(frame_w, frame_h)
    scale = _downscale_factor(frame_h, frame_w)
    proc_w = max(1, int(frame_w * scale))
    proc_h = max(1, int(frame_h * scale))

    _log(f"  feed {video_path!r}: {frame_w}×{frame_h} → proc {proc_w}×{proc_h}")

    # Build background model from the darkest pixel across BG_FRAMES.
    # Using the per-pixel minimum is robust against a light being on in the
    # very first frame.
    bg_acc: list[np.ndarray] = []
    for _ in range(BG_FRAMES):
        ret, frame = cap.read()
        if not ret:
            break
        small = cv2.resize(frame, (proc_w, proc_h), interpolation=cv2.INTER_AREA)
        bg_acc.append(cv2.cvtColor(small, cv2.COLOR_BGR2GRAY))

    if bg_acc:
        background = np.min(np.stack(bg_acc, axis=0), axis=0)
    else:
        background = np.zeros((proc_h, proc_w), dtype=np.uint8)

    # Reset to frame 0 and detect blinks over the full clip.
    cap.set(cv2.CAP_PROP_POS_FRAMES, 0)

    blinks: list[tuple[int, float, float]] = []
    ordinal = 0
    in_blink = False
    accum: list[tuple[float, float]] = []

    while True:
        ret, frame = cap.read()
        if not ret:
            break

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
                accum = [(cx_full, cy_full)]
            else:
                accum.append((cx_full, cy_full))
        else:
            if in_blink:
                _finalise_blink(blinks, ordinal, accum)
                ordinal += 1
                accum = []
                in_blink = False

    # Handle clip ending while still inside a blink.
    if in_blink and accum:
        _finalise_blink(blinks, ordinal, accum)

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
    blinks: list[tuple[int, float, float]],
    ordinal: int,
    accum: list[tuple[float, float]],
) -> None:
    cx = sum(p[0] for p in accum) / len(accum)
    cy = sum(p[1] for p in accum) / len(accum)
    blinks.append((ordinal, cx, cy))


# ──────────────────────────────────────────────────────────────────────────────
# Stage 2 — Align detections across feeds
# ──────────────────────────────────────────────────────────────────────────────

def align_detections(
    feed_blinks: list[list[tuple[int, float, float]]],
) -> dict[int, list[tuple[int, float, float]]]:
    """
    Merges per-feed blink lists into:
        light_index → [(feed_idx, cx, cy), …]

    Per REQ-047 / REQ-048 BR 2, ordinal k in a feed maps directly to light
    index k.  Alignment is by ordinal sequence, not wall-clock time, so
    cameras that start or stop at different moments still agree.
    """
    by_light: dict[int, list[tuple[int, float, float]]] = {}
    for feed_idx, blinks in enumerate(feed_blinks):
        for ordinal, cx, cy in blinks:
            light_idx = ordinal
            by_light.setdefault(light_idx, []).append((feed_idx, cx, cy))
    return by_light


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
        SI metres.  When markers are used this is exact; when using
        scale_hint_m it is the caller-supplied baseline; otherwise 1.0
        (output is in normalised baseline units).
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
                return aruco_poses, float(marker_spec["edge_length_m"])
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

    E, mask = cv2.findEssentialMat(
        pts0, pts1, Ks[0], method=cv2.RANSAC, prob=0.999, threshold=1.0
    )
    if E is None:
        _log("  essential matrix estimation failed")
        return poses, scale_hint_m or DEFAULT_SCALE_M

    # findEssentialMat may return multiple stacked 3×3 solutions (shape 9×3);
    # recoverPose requires exactly one 3×3 matrix.
    if E.shape != (3, 3):
        E = E[:3, :]

    inlier_mask = mask.ravel() == 1
    if inlier_mask.sum() < 5:
        _log(f"  too few E-matrix inliers ({inlier_mask.sum()}) — skipping pose est.")
        return poses, scale_hint_m or DEFAULT_SCALE_M

    _, R1, t1, _ = cv2.recoverPose(
        E, pts0[inlier_mask], pts1[inlier_mask], Ks[0]
    )
    poses[1] = (R1, t1)
    _log(f"  E-matrix: feed 1 recovered, {inlier_mask.sum()} inliers")

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

    cap.release()
    if pose is None:
        _log(f"    feed {feed_idx}: no marker in first {MARKER_SCAN_SECS:.0f}s")
    return pose


# --- Essential-matrix helpers ------------------------------------------------

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
    Lights with reprojection error > LOW_CONF_REPROJ_PX go to *low_confidence*.
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

        reproj = _reprojection_error(pt3d, valid, Ps)
        if reproj > LOW_CONF_REPROJ_PX:
            _log(f"  light {light_id}: reproj={reproj:.1f}px → low_confidence")
            low_conf.append(light_id)

        x, y, z = pt3d * metric_scale
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


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

def main() -> None:
    if len(sys.argv) < 2:
        print(json.dumps(_make_error("usage: reconstruct.py <spec_file.json>")))
        sys.exit(1)

    try:
        with open(sys.argv[1]) as fh:
            spec = json.load(fh)
    except Exception as exc:
        print(json.dumps(_make_error(f"cannot read spec: {exc}")))
        sys.exit(1)

    feeds = spec.get("feeds", [])
    if len(feeds) < 2:
        print(json.dumps(_make_error(
            f"at least 2 feeds are required for triangulation; got {len(feeds)}"
        )))
        sys.exit(1)

    dwell_ms      = int(spec.get("dwell_ms", 1000))
    marker_spec   = spec.get("marker")
    scale_hint_m  = spec.get("scale_hint_m")

    try:
        _log(f"DLM reconstruct: {len(feeds)} feed(s), dwell={dwell_ms} ms")

        # Stage 1 — blink detection.
        feed_blinks: list[list[tuple[int, float, float]]] = []
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
            print(json.dumps(_make_error("no blink events detected in any feed")))
            return

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
        print(json.dumps(result))

    except Exception as exc:
        _log(f"ERROR: {exc}")
        _log(traceback.format_exc())
        print(json.dumps(_make_error(str(exc))))
        sys.exit(1)


if __name__ == "__main__":
    main()
