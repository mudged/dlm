#!/usr/bin/env python3
"""
gen_fixtures.py — Generate synthetic video fixtures for reconstruct.py tests.

Creates two MP4 clips where known 3D light positions are swept in order (one
at a time) against a dark background.  Optionally embeds a perspective-correct
ArUco marker into each frame.

Usage
-----
    python3 gen_fixtures.py \\
        [--output-dir DIR]         # default: ./fixtures
        [--n-lights N]             # default: 6
        [--with-marker]            # embed ArUco marker in scene
        [--dwell-ms MS]            # blink dwell in ms, default: 333
        [--fps FPS]                # video frame rate, default: 30
        [--width W]                # frame width pixels, default: 320
        [--height H]               # frame height pixels, default: 240
        [--seed SEED]              # RNG seed, default: 42
        [--occlude-in-feed FEED_IDX LIGHT_IDX]  # suppress a light in one feed

Outputs
-------
    <output_dir>/feed_0.mp4
    <output_dir>/feed_1.mp4
    <output_dir>/ground_truth.json
        {
          "lights":   [{"id": N, "x": X, "y": Y, "z": Z}, ...],
          "dwell_ms": N,
          "marker":   {"dictionary": "DICT_4X4_50", "edge_length_m": 0.05} | null,
          "cameras":  [{"R": ..., "t": ..., "K": ...}, ...],
          "occlude":  {"feed": F, "light": L} | null
        }
"""

import argparse
import json
import math
import sys
from pathlib import Path
from typing import Optional

import cv2
import numpy as np


# ──────────────────────────────────────────────────────────────────────────────
# CLI
# ──────────────────────────────────────────────────────────────────────────────

def _parse_args(argv=None):
    p = argparse.ArgumentParser(
        description="Generate synthetic DLM test fixtures",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    p.add_argument("--output-dir",  default="fixtures",  help="Output directory")
    p.add_argument("--n-lights",    type=int, default=6,   help="Number of lights")
    p.add_argument("--with-marker", action="store_true",   help="Embed ArUco marker in scene")
    p.add_argument("--dwell-ms",    type=int, default=333, help="Blink dwell (ms)")
    p.add_argument("--fps",         type=int, default=30,  help="Video frame rate")
    p.add_argument("--width",       type=int, default=320, help="Frame width (px)")
    p.add_argument("--height",      type=int, default=240, help="Frame height (px)")
    p.add_argument("--seed",        type=int, default=42,  help="Random seed")
    p.add_argument(
        "--occlude-in-feed",
        nargs=2,
        type=int,
        metavar=("FEED_IDX", "LIGHT_IDX"),
        default=None,
        help="Suppress one light in one feed to test missing detection",
    )
    return p.parse_args(argv)


# ──────────────────────────────────────────────────────────────────────────────
# Geometry helpers
# ──────────────────────────────────────────────────────────────────────────────

def _make_K(w: int, h: int) -> np.ndarray:
    """Pinhole camera matrix matching reconstruct.py's _estimate_K."""
    f = float(max(w, h))
    return np.array([[f, 0.0, w / 2.0],
                     [0.0, f, h / 2.0],
                     [0.0, 0.0, 1.0]], dtype=np.float64)


def _project(pt3d: np.ndarray, K: np.ndarray, R: np.ndarray, t: np.ndarray):
    """Returns (px, py) or None if the point is behind the camera."""
    p_cam = R @ pt3d.reshape(3) + t.reshape(3)
    if p_cam[2] <= 0:
        return None
    px = K[0, 0] * p_cam[0] / p_cam[2] + K[0, 2]
    py = K[1, 1] * p_cam[1] / p_cam[2] + K[1, 2]
    return float(px), float(py)


def _rot_y(angle_rad: float) -> np.ndarray:
    c, s = math.cos(angle_rad), math.sin(angle_rad)
    return np.array([[c, 0, s], [0, 1, 0], [-s, 0, c]], dtype=np.float64)


# ──────────────────────────────────────────────────────────────────────────────
# Frame rendering
# ──────────────────────────────────────────────────────────────────────────────

def _blank_frame(w: int, h: int) -> np.ndarray:
    return np.zeros((h, w, 3), dtype=np.uint8)


def _draw_blob(frame: np.ndarray, pt2d, radius: int = 6) -> None:
    if pt2d is None:
        return
    cx, cy = int(round(pt2d[0])), int(round(pt2d[1]))
    h, w = frame.shape[:2]
    if 0 <= cx < w and 0 <= cy < h:
        cv2.circle(frame, (cx, cy), radius, (255, 255, 255), -1)


def _draw_marker_billboard(
    frame: np.ndarray,
    marker_img: np.ndarray,
    center_px: tuple[float, float],
    size_px: int,
    margin_px: int = 20,
) -> None:
    """
    Draws *marker_img* as a flat billboard at *center_px* in *frame*.

    A white rectangle (marker + *margin_px* quiet zone on each side) is
    drawn first so that the ArUco detector can reliably find the marker
    against the otherwise dark background.  The marker image is resized to
    *size_px* × *size_px* pixels.
    """
    cx, cy = int(round(center_px[0])), int(round(center_px[1]))
    half = size_px // 2
    h, w = frame.shape[:2]

    # White background region including the quiet zone.
    x0 = max(0, cx - half - margin_px)
    y0 = max(0, cy - half - margin_px)
    x1 = min(w, cx + half + margin_px)
    y1 = min(h, cy + half + margin_px)
    if x1 <= x0 or y1 <= y0:
        return  # off-screen
    frame[y0:y1, x0:x1] = (255, 255, 255)

    # Marker image.
    m = cv2.resize(marker_img, (size_px, size_px), interpolation=cv2.INTER_NEAREST)
    m_bgr = cv2.cvtColor(m, cv2.COLOR_GRAY2BGR)
    mx0 = max(0, cx - half)
    my0 = max(0, cy - half)
    mx1 = min(w, mx0 + size_px)
    my1 = min(h, my0 + size_px)
    sx0 = mx0 - (cx - half)
    sy0 = my0 - (cy - half)
    frame[my0:my1, mx0:mx1] = m_bgr[sy0:sy0 + (my1 - my0), sx0:sx0 + (mx1 - mx0)]


# ──────────────────────────────────────────────────────────────────────────────
# Main generator
# ──────────────────────────────────────────────────────────────────────────────

def generate(args) -> Path:
    rng = np.random.default_rng(args.seed)
    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    W, H, FPS = args.width, args.height, args.fps
    dwell_ms = args.dwell_ms
    n_lights = args.n_lights

    # Both cameras share the same intrinsic matrix, matching reconstruct.py.
    K = _make_K(W, H)

    # ── 3D light positions ────────────────────────────────────────────────────
    # Scattered in front of both cameras at z ≈ 0.8–1.2 m.
    #
    # X range is constrained so that all lights are visible in camera 1
    # (which is 0.30 m to the right).  Camera 1's left edge in world-X at the
    # minimum depth (0.80 m) is: 0.30 - 0.5*0.80 = -0.10 m.  Using [-0.08, 0.08]
    # gives a small safety margin.
    lights_3d: list[np.ndarray] = []
    for _ in range(n_lights):
        x = float(rng.uniform(-0.08, 0.08))
        y = float(rng.uniform(-0.08, 0.08))
        z = float(rng.uniform(0.80, 1.20))
        lights_3d.append(np.array([x, y, z]))

    # ── Camera poses ─────────────────────────────────────────────────────────
    # Camera 0: at world origin, looking down +Z.
    R0 = np.eye(3, dtype=np.float64)
    t0 = np.zeros((3, 1), dtype=np.float64)

    # Camera 1: 0.30 m to the right, no rotation (pure-translation stereo).
    # This makes recoverPose return the exact identity rotation, which keeps
    # the coordinate accuracy test well-conditioned.
    R1 = np.eye(3, dtype=np.float64)
    t1 = np.array([[-0.30], [0.0], [0.0]], dtype=np.float64)

    cameras = [(R0, t0), (R1, t1)]

    # ── ArUco marker ─────────────────────────────────────────────────────────
    # Placed at world position (0, 0.18, 0.55) m — above the light cluster,
    # facing the cameras (+Z towards origin).  This lets estimatePoseSingleMarkers
    # recover each camera's pose relative to the marker frame.
    marker_img: Optional[np.ndarray] = None  # type: ignore[name-defined]
    marker_spec = None
    marker_centre_world: Optional[np.ndarray] = None  # type: ignore[name-defined]
    MARKER_EDGE_M = 0.08  # physical edge length written into the spec

    if args.with_marker:
        dict_id = cv2.aruco.DICT_4X4_50
        adict = cv2.aruco.getPredefinedDictionary(dict_id)
        marker_img = cv2.aruco.generateImageMarker(adict, 0, 64)
        marker_spec = {"dictionary": "DICT_4X4_50", "edge_length_m": MARKER_EDGE_M}

        # 3D world position of the marker centre.
        # Placed midway between the two camera X positions (0 and 0.30 m) at a
        # y-level that puts it in the upper part of both frames and above the
        # light cluster (lights project to py≈93..160 in camera 0).
        # Camera 1 (pure X translation) sees the marker at the mirrored X offset.
        marker_centre_world = np.array([0.15, -0.10, 0.60])  # noqa: F821 world metres

    # ── Frame timing ─────────────────────────────────────────────────────────
    dwell_frames = max(1, round(dwell_ms * FPS / 1000))
    gap_frames   = max(1, round(50 * FPS / 1000))   # 50 ms dark gap

    # ── Occlusion spec ────────────────────────────────────────────────────────
    occlude_feed, occlude_light = (args.occlude_in_feed or (None, None))

    # ── Render videos ─────────────────────────────────────────────────────────
    for cam_idx, (R, t) in enumerate(cameras):
        frames: list[np.ndarray] = []

        # Pre-compute marker billboard centre for this camera.
        marker_center_px: Optional[tuple[float, float]] = None  # type: ignore[name-defined]
        if marker_img is not None and marker_centre_world is not None:
            mc_proj = _project(marker_centre_world, K, R, t)
            if mc_proj is not None:
                # Only draw if the centre is within the frame bounds.
                px, py = mc_proj
                if 0 <= px < W and 0 <= py < H:
                    marker_center_px = mc_proj

        # Billboard size scales roughly with depth for realism, capped to a
        # range that the ArUco detector can handle reliably (50–80 px).
        MARKER_BILLBOARD_PX = 60

        def _base_frame() -> np.ndarray:
            f = _blank_frame(W, H)
            if marker_center_px is not None and marker_img is not None:
                _draw_marker_billboard(f, marker_img, marker_center_px,
                                       MARKER_BILLBOARD_PX)
            return f

        # Leading dark gap.
        for _ in range(gap_frames):
            frames.append(_base_frame())

        for light_idx, pt3d in enumerate(lights_3d):
            pt2d = _project(pt3d, K, R, t)
            occluded = (cam_idx == occlude_feed and light_idx == occlude_light)

            for _ in range(dwell_frames):
                f = _base_frame()
                if not occluded:
                    _draw_blob(f, pt2d)
                frames.append(f)

            # Dark gap between lights.
            for _ in range(gap_frames):
                frames.append(_base_frame())

        # Write video.
        path = out_dir / f"feed_{cam_idx}.avi"
        _write_video(str(path), frames, FPS)
        print(f"Written: {path} ({len(frames)} frames, {len(frames)/FPS:.1f}s)")

    # ── Ground truth ──────────────────────────────────────────────────────────
    gt = {
        "lights": [
            {"id": i, "x": float(p[0]), "y": float(p[1]), "z": float(p[2])}
            for i, p in enumerate(lights_3d)
        ],
        "dwell_ms": dwell_ms,
        "marker": marker_spec,
        "cameras": [
            {"R": R.tolist(), "t": t.tolist(), "K": K.tolist()}
            for R, t in cameras
        ],
        "occlude": (
            {"feed": occlude_feed, "light": occlude_light}
            if occlude_feed is not None else None
        ),
    }
    gt_path = out_dir / "ground_truth.json"
    gt_path.write_text(json.dumps(gt, indent=2))
    print(f"Written: {gt_path}")
    return out_dir


def _write_video(path: str, frames: list[np.ndarray], fps: int) -> None:
    if not frames:
        return
    h, w = frames[0].shape[:2]
    # XVID in AVI is substantially lossless at this resolution and avoids the
    # DCT block artefacts that mp4v introduces; ArUco detection is sensitive to
    # border-pixel quality.
    fourcc = cv2.VideoWriter_fourcc(*"XVID")
    out = cv2.VideoWriter(path, fourcc, fps, (w, h))
    for f in frames:
        out.write(f)
    out.release()


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    args = _parse_args()
    generate(args)
