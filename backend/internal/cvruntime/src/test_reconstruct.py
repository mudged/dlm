#!/usr/bin/env python3
"""
test_reconstruct.py — Integration tests for reconstruct.py using synthetic fixtures.

Each test runs the full pipeline (gen_fixtures → reconstruct) on tiny
in-memory-rendered clips, keeping the suite fast.

Run with:
    python3 test_reconstruct.py [-v]

Requirements:
    pip install opencv-python-headless numpy
    (same deps as reconstruct.py / gen_fixtures.py)
"""

import importlib.util
import json
import math
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import numpy as np

import cv2

SCRIPT_DIR   = Path(__file__).parent
RECONSTRUCT  = SCRIPT_DIR / "reconstruct.py"
GEN_FIXTURES = SCRIPT_DIR / "gen_fixtures.py"


# ──────────────────────────────────────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────────────────────────────────────

def _gen(outdir: str, **kwargs) -> dict:
    """Run gen_fixtures.py with *kwargs* mapped to CLI flags."""
    cmd = [sys.executable, str(GEN_FIXTURES), "--output-dir", outdir]
    for k, v in kwargs.items():
        flag = "--" + k.replace("_", "-")
        if isinstance(v, bool):
            if v:
                cmd.append(flag)
        elif isinstance(v, (list, tuple)):
            cmd += [flag] + [str(x) for x in v]
        else:
            cmd += [flag, str(v)]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError(f"gen_fixtures failed:\n{r.stdout}\n{r.stderr}")
    gt_path = Path(outdir) / "ground_truth.json"
    return json.loads(gt_path.read_text())


def _reconstruct(spec: dict) -> dict:
    """Write *spec* to a temp file, invoke reconstruct.py, return parsed Result."""
    with tempfile.NamedTemporaryFile(
        mode="w", suffix=".json", delete=False
    ) as fh:
        json.dump(spec, fh)
        spec_path = fh.name
    try:
        r = subprocess.run(
            [sys.executable, str(RECONSTRUCT), spec_path],
            capture_output=True,
            text=True,
        )
    finally:
        os.unlink(spec_path)

    sys.stderr.write(r.stderr)  # surface diagnostic output on failure
    if r.returncode != 0 and not r.stdout.strip():
        raise RuntimeError(
            f"reconstruct.py exited {r.returncode} with no JSON output.\n"
            f"stderr: {r.stderr}"
        )
    return json.loads(r.stdout)


def _reconstruct_with_exit(spec: dict) -> tuple[dict, int]:
    """Like _reconstruct but also return the child process exit code."""
    with tempfile.NamedTemporaryFile(
        mode="w", suffix=".json", delete=False
    ) as fh:
        json.dump(spec, fh)
        spec_path = fh.name
    try:
        r = subprocess.run(
            [sys.executable, str(RECONSTRUCT), spec_path],
            capture_output=True,
            text=True,
        )
    finally:
        os.unlink(spec_path)

    sys.stderr.write(r.stderr)
    if r.returncode != 0 and not r.stdout.strip():
        raise RuntimeError(
            f"reconstruct.py exited {r.returncode} with no JSON output.\n"
            f"stderr: {r.stderr}"
        )
    return json.loads(r.stdout), r.returncode


def _spec(outdir: str, gt: dict, extra: dict | None = None) -> dict:
    """Build a minimal JobSpec for two feeds in *outdir*."""
    s: dict = {
        "feeds":    [{"path": str(Path(outdir) / f"feed_{i}.avi")} for i in range(2)],
        "dwell_ms": gt["dwell_ms"],
    }
    if extra:
        s.update(extra)
    return s


# ──────────────────────────────────────────────────────────────────────────────
# Test cases
# ──────────────────────────────────────────────────────────────────────────────

class TestReconstructSynthetic(unittest.TestCase):

    # ── ordering and basic structure ─────────────────────────────────────────

    def test_ordering_and_id_type(self):
        """Recovered IDs are 0-based integers in ascending order."""
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=5, seed=1)
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        ids = [lp["id"] for lp in res["lights"]]
        self.assertEqual(sorted(ids), ids, "IDs not in ascending order")
        for lid in ids:
            self.assertIsInstance(lid, int)
            self.assertGreaterEqual(lid, 0)

    def test_light_count_field(self):
        """light_count == highest seen index + 1."""
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=4, seed=2)
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        all_ids = (
            {lp["id"] for lp in res["lights"]}
            | set(res["missing"])
            | set(res["low_confidence"])
        )
        if all_ids:
            self.assertEqual(res["light_count"], max(all_ids) + 1)

    def test_most_lights_recovered(self):
        """At least N-1 of N lights must be triangulated (basic no-marker run)."""
        n = 6
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=n, seed=3)
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertGreaterEqual(
            len(res["lights"]),
            n - 1,
            f"Expected ≥{n - 1} triangulated lights, got {len(res['lights'])}",
        )

    # ── occlusion / missing detection (REQ-048 BR 7) ────────────────────────

    def test_occluded_light_in_missing_not_lights(self):
        """A light visible in only 1 feed must appear in missing, never in lights."""
        occlude_feed, occlude_light = 1, 2
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=5, seed=4,
                       occlude_in_feed=[occlude_feed, occlude_light])
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        recovered_ids = {lp["id"] for lp in res["lights"]}
        self.assertNotIn(
            occlude_light, recovered_ids,
            f"Light {occlude_light} occluded in feed {occlude_feed} "
            f"must NOT have coordinates",
        )
        self.assertIn(
            occlude_light, res["missing"],
            f"Light {occlude_light} must appear in 'missing'",
        )

    def test_no_fabrication_guarantee(self):
        """lights + missing + low_confidence are mutually disjoint (no duplication)."""
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=5, seed=5)
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        recovered = {lp["id"] for lp in res["lights"]}
        miss      = set(res["missing"])
        low       = set(res["low_confidence"])

        self.assertEqual(recovered & miss, set(),
                         "A light ID appears in both lights and missing")
        # low_confidence is a subset of lights (it has coordinates, just uncertain)
        self.assertTrue(low <= recovered or low == set(),
                        "low_confidence IDs should be a subset of lights IDs")

    # ── scale_hint_m coordinate accuracy ─────────────────────────────────────

    def test_coordinate_accuracy_with_scale_hint(self):
        """
        Supplying scale_hint_m = 0.30 m (the camera baseline used by gen_fixtures)
        should yield recovered coordinates within 8 cm of ground truth for the
        lights that get triangulated.

        The 8 cm tolerance accounts for blob-centroid discretisation at the
        320×240 resolution used in the fixtures.
        """
        TOLERANCE_M = 0.08
        BASELINE_M  = 0.30   # matches gen_fixtures camera 1 t[0]

        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=5, seed=6)
            res = _reconstruct(_spec(d, gt, {"scale_hint_m": BASELINE_M}))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        gt_by_id = {lp["id"]: lp for lp in gt["lights"]}
        checked  = 0
        for lp in res["lights"]:
            lid = lp["id"]
            if lid not in gt_by_id:
                continue
            g = gt_by_id[lid]
            dist = math.sqrt(
                (lp["x"] - g["x"]) ** 2 +
                (lp["y"] - g["y"]) ** 2 +
                (lp["z"] - g["z"]) ** 2
            )
            self.assertLessEqual(
                dist, TOLERANCE_M,
                f"Light {lid}: dist={dist:.4f} m > tolerance {TOLERANCE_M} m\n"
                f"  recovered=({lp['x']:.3f},{lp['y']:.3f},{lp['z']:.3f})\n"
                f"  truth    =({g['x']:.3f},{g['y']:.3f},{g['z']:.3f})",
            )
            checked += 1
        self.assertGreater(checked, 0, "No recovered lights matched GT IDs")

    # ── ArUco marker path ────────────────────────────────────────────────────

    def test_with_marker_spec_succeeds_and_finds_lights(self):
        """
        When a marker spec is included in the JobSpec, reconstruction must
        succeed and recover most lights.

        Two sub-scenarios are meaningful in practice:
          (a) Marker visible in both feeds → ArUco poses used.
          (b) Marker not visible (spec provided but frames don't contain it)
              → graceful fallback to the essential-matrix path.

        This test covers scenario (b), which also exercises the full code path
        through the ArUco attempt and fallback.  Scenario (a) requires either
        a physical camera setup or a more sophisticated synthetic scene that can
        render a perspective-correct marker without occluding the lights.
        """
        n = 6  # ≥ 6 ensures E-matrix has enough correspondences
        with tempfile.TemporaryDirectory() as d:
            # Generate fixture WITHOUT an embedded marker so blink detection
            # is not disturbed.  Pass the marker spec anyway to exercise the
            # ArUco-attempt → E-matrix fallback code path.
            gt  = _gen(d, n_lights=n, seed=7)
            spec = _spec(d, gt)
            spec["marker"] = {"dictionary": "DICT_4X4_50", "edge_length_m": 0.05}
            res = _reconstruct(spec)

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertGreaterEqual(
            len(res["lights"]),
            n - 1,
            f"Expected ≥{n-1} lights (via E-matrix fallback), got {len(res['lights'])}",
        )

    def test_aruco_no_double_scaling(self):
        """
        ArUco success path must return metric_scale=1.0, not edge_length_m.

        solvePnP uses object points defined in metres, so the resulting poses
        (and hence DLT-triangulated points) are already metric.  Multiplying
        again by edge_length_m shrinks every coordinate by the marker size —
        e.g. a 0.05 m marker makes coordinates 20× too small.

        This test imports reconstruct.py directly, stubs _poses_from_aruco with
        two known metric camera poses (baseline 0.30 m), then:
          (a) asserts estimate_poses returns metric_scale == 1.0
          (b) runs triangulate_all with that scale and asserts the recovered
              world point is within 1 mm of the known ground-truth position

        Both assertions would fail with the old code (scale = edge_length_m =
        0.05 → recovered z ≈ 0.05 m instead of 1.0 m).
        """
        spec_mod = importlib.util.spec_from_file_location(
            "_reconstruct_impl", str(RECONSTRUCT)
        )
        m = importlib.util.module_from_spec(spec_mod)
        spec_mod.loader.exec_module(m)

        # Synthetic two-camera stereo rig: camera 0 at origin, camera 1
        # displaced 0.30 m along X.  Both cameras look down +Z.
        BASELINE_M = 0.30
        DEPTH_M    = 1.0    # known light depth in metres
        f          = 320.0
        W, H       = 320, 240

        K = np.array([[f, 0.0, W / 2.0],
                      [0.0, f, H / 2.0],
                      [0.0, 0.0, 1.0  ]], dtype=np.float64)

        R0 = np.eye(3,  dtype=np.float64)
        t0 = np.zeros((3, 1), dtype=np.float64)
        R1 = np.eye(3,  dtype=np.float64)
        t1 = np.array([[-BASELINE_M], [0.0], [0.0]], dtype=np.float64)

        # --- (a) metric_scale assertion ---
        def _stub_aruco(feed_paths, Ks, marker_spec):
            return [(R0, t0), (R1, t1)]

        original = m._poses_from_aruco
        m._poses_from_aruco = _stub_aruco
        try:
            poses, scale = m.estimate_poses(
                feed_paths=["a.mp4", "b.mp4"],
                Ks=[K, K],
                by_light={},
                marker_spec={"dictionary": "DICT_4X4_50", "edge_length_m": 0.05},
                scale_hint_m=None,
            )
        finally:
            m._poses_from_aruco = original

        self.assertEqual(
            scale, 1.0,
            f"ArUco success path returned metric_scale={scale}; expected 1.0. "
            f"(The old bug returned edge_length_m and caused double-scaling.)",
        )

        # --- (b) triangulation accuracy assertion ---
        # Known light at world origin (0, 0, DEPTH_M) in metres.
        pt_world = np.array([0.0, 0.0, DEPTH_M])

        def _proj(R, t, pt):
            c = R @ pt + t.ravel()
            return f * c[0] / c[2] + W / 2.0, f * c[1] / c[2] + H / 2.0

        cx0, cy0 = _proj(R0, t0, pt_world)
        cx1, cy1 = _proj(R1, t1, pt_world)
        by_light = {0: [(0, cx0, cy0), (1, cx1, cy1)]}

        lights_3d, missing, _ = m.triangulate_all(by_light, poses, [K, K], scale)

        self.assertIn(0, lights_3d, "Light 0 must be triangulated")
        self.assertEqual(missing, [], "No lights should be in missing")

        x, y, z = lights_3d[0]
        self.assertAlmostEqual(
            z, DEPTH_M, delta=0.001,
            msg=f"z={z:.4f} m should be ~{DEPTH_M} m (ArUco double-scaling would give ~{0.05 * DEPTH_M:.4f} m)",
        )
        self.assertAlmostEqual(x, 0.0, delta=0.001, msg=f"x={x:.4f} m should be ~0.0 m")
        self.assertAlmostEqual(y, 0.0, delta=0.001, msg=f"y={y:.4f} m should be ~0.0 m")

    # ── error handling ───────────────────────────────────────────────────────

    def test_error_on_single_feed(self):
        """A single-feed spec must return status=failed (triangulation impossible)."""
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=3, seed=8)
            spec = {
                "feeds":    [{"path": str(Path(d) / "feed_0.avi")}],
                "dwell_ms": gt["dwell_ms"],
            }
            res = _reconstruct(spec)

        self.assertEqual(res["status"], "failed")
        self.assertIsNotNone(res["error"])
        self.assertIsInstance(res["error"], str)
        self.assertGreater(len(res["error"]), 0)

    def test_result_schema(self):
        """Result JSON must contain all required fields with correct types."""
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=3, seed=9)
            res = _reconstruct(_spec(d, gt))

        required = ("status", "light_count", "lights", "missing",
                    "low_confidence", "error")
        for key in required:
            self.assertIn(key, res, f"Result missing field '{key}'")

        self.assertIsInstance(res["status"],         str)
        self.assertIsInstance(res["light_count"],    int)
        self.assertIsInstance(res["lights"],         list)
        self.assertIsInstance(res["missing"],        list)
        self.assertIsInstance(res["low_confidence"], list)

        for lp in res["lights"]:
            for field in ("id", "x", "y", "z"):
                self.assertIn(field, lp, f"LightPoint missing field '{field}'")
            self.assertIsInstance(lp["id"], int)
            self.assertIsInstance(lp["x"],  float)
            self.assertIsInstance(lp["y"],  float)
            self.assertIsInstance(lp["z"],  float)

    # ── multi-feed occlusion: re-alignment (WI-25) ──────────────────────────

    def test_dropped_blink_does_not_mislabel_other_lights(self):
        """
        A feed that misses a single blink (occlusion) must NOT shift the indices
        of the lights that blink after it.

        With naive ordinal→index mapping, occluding light 2 in feed 1 would make
        feed 1's later blinks slide into slots 2, 3, … so light 2 gets
        mismatched 2D points (wrong coordinates) and the lights after it are
        mislabeled.  Timing-based correspondence (WI-25) must instead leave the
        occluded slot empty (light 2 → missing) while lights 3, 4, 5 keep their
        correct IDs and positions.

        Uses n_lights=6 so that even after the drop there are ≥5 common points
        between the two feeds, letting pose estimation succeed and the unaffected
        lights actually triangulate (a stronger check than the all-missing
        degenerate case).
        """
        TOLERANCE_M = 0.08
        BASELINE_M  = 0.30
        occlude_feed, occlude_light = 1, 2
        n = 6

        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=n, seed=21,
                       occlude_in_feed=[occlude_feed, occlude_light])
            res = _reconstruct(_spec(d, gt, {"scale_hint_m": BASELINE_M}))

        self.assertEqual(res["status"], "succeeded", res.get("error"))

        recovered_ids = {lp["id"] for lp in res["lights"]}
        # The occluded light has only one usable view → must be reported missing,
        # never fabricated with coordinates from a mismatched correspondence.
        self.assertNotIn(occlude_light, recovered_ids,
                         f"Occluded light {occlude_light} must not get coordinates")
        self.assertIn(occlude_light, res["missing"],
                      f"Occluded light {occlude_light} must be in 'missing'")

        # Lights that blink AFTER the dropped one must keep their correct IDs and
        # land near ground truth (i.e. not be shifted into the wrong slot).
        gt_by_id = {lp["id"]: lp for lp in gt["lights"]}
        checked_after = 0
        for lp in res["lights"]:
            lid = lp["id"]
            if lid <= occlude_light or lid not in gt_by_id:
                continue
            g = gt_by_id[lid]
            dist = math.sqrt(
                (lp["x"] - g["x"]) ** 2 +
                (lp["y"] - g["y"]) ** 2 +
                (lp["z"] - g["z"]) ** 2
            )
            self.assertLessEqual(
                dist, TOLERANCE_M,
                f"Light {lid} (after the dropped blink) drifted {dist:.4f} m "
                f"> {TOLERANCE_M} m — correspondence likely shifted",
            )
            checked_after += 1
        self.assertGreater(
            checked_after, 0,
            "Expected at least one light after the occluded one to triangulate",
        )

    def test_mixed_intrinsics_pipeline_succeeds(self):
        """
        WI-28: the full pipeline completes on mixed-resolution feeds.

        Coordinate accuracy with differing intrinsics is asserted in
        TestPerFeedIntrinsics (synthetic projections); video fixtures at
        different resolutions carry dissimilar blob discretisation that makes
        end-to-end metric accuracy a poor regression signal here.
        """
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(
                d, n_lights=6, seed=42,
                feed1_width=480, feed1_height=360,
            )
            res = _reconstruct(_spec(d, gt, {"scale_hint_m": 0.30}))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertGreaterEqual(
            len(res["lights"]), 5,
            "expected most lights triangulated on mixed-resolution feeds",
        )

    def test_missing_cover_full_range(self):
        """
        The union of lights + missing + low_confidence IDs must cover 0 … light_count-1
        (every index is either triangulated, missing, or flagged low-confidence).
        """
        with tempfile.TemporaryDirectory() as d:
            gt  = _gen(d, n_lights=5, seed=10)
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        if res["light_count"] == 0:
            return

        all_seen = (
            {lp["id"] for lp in res["lights"]}
            | set(res["missing"])
            | set(res["low_confidence"])
        )
        expected = set(range(res["light_count"]))
        self.assertEqual(
            all_seen,
            expected,
            f"IDs not covering 0…{res['light_count']-1}: "
            f"seen={sorted(all_seen)}, expected={sorted(expected)}",
        )


def _load_reconstruct_module():
    """Imports reconstruct.py as a module so its internals can be unit-tested."""
    spec_mod = importlib.util.spec_from_file_location(
        "_reconstruct_impl", str(RECONSTRUCT)
    )
    m = importlib.util.module_from_spec(spec_mod)
    spec_mod.loader.exec_module(m)
    return m


class TestCrossFeedCorrespondence(unittest.TestCase):
    """
    Unit tests for WI-25 timing-based blink correspondence.

    These exercise `align_detections` / `_assign_slots` / `_estimate_period`
    directly with synthetic, timestamped blink lists so the slot→light mapping
    can be asserted precisely (independent of video rendering).  Each light is
    given a distinct centroid x = light_index * 10 so a mislabel is detectable.
    """

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    # Build one feed's blink list: (t, cx, cy) with cx encoding the light id.
    @staticmethod
    def _feed(items):
        # items: list of (time, light_id)  → (time, light_id*10, 0.0)
        return [(float(t), float(lid * 10), 0.0) for t, lid in items]

    # ── _assign_slots ────────────────────────────────────────────────────────

    def test_assign_slots_happy_path(self):
        self.assertEqual(self.m._assign_slots([0.0, 1.0, 2.0, 3.0], 1.0),
                         [0, 1, 2, 3])

    def test_assign_slots_dropped_blink_skips_slot(self):
        # A ~2× gap (missing blink) must skip a slot, not back-fill it.
        self.assertEqual(self.m._assign_slots([0.0, 1.0, 3.0, 4.0], 1.0),
                         [0, 1, 3, 4])

    def test_assign_slots_spurious_collides_not_shifts(self):
        # An off-cadence extra blink (t=1.4) stays in the current slot and does
        # NOT advance the cadence anchor, so later real blinks keep their slots.
        self.assertEqual(self.m._assign_slots([0.0, 1.0, 1.4, 2.0, 3.0], 1.0),
                         [0, 1, 1, 2, 3])

    def test_assign_slots_zero_period_falls_back_to_positional(self):
        self.assertEqual(self.m._assign_slots([0.0, 5.0, 10.0], 0.0),
                         [0, 1, 2])

    # ── _estimate_period ──────────────────────────────────────────────────────

    def test_estimate_period_robust_to_outliers(self):
        # Majority of intervals are 1.0; a drop (2.0) and a spurious (0.4) are
        # outliers the median ignores.
        feed_a = self._feed([(0, 0), (1, 1), (2, 2), (3, 3), (4, 4)])
        feed_b = self._feed([(0, 0), (1, 1), (1.4, 99), (3, 3), (4, 4)])
        self.assertAlmostEqual(self.m._estimate_period([feed_a, feed_b]), 1.0,
                               places=6)

    def test_estimate_period_no_intervals(self):
        # Single blink per feed → no intervals → 0.0 (caller falls back).
        self.assertEqual(
            self.m._estimate_period([self._feed([(0, 0)]), self._feed([(0, 0)])]),
            0.0,
        )

    # ── align_detections ──────────────────────────────────────────────────────

    def test_align_happy_path_matches_by_slot(self):
        feed_a = self._feed([(0, 0), (1, 1), (2, 2), (3, 3)])
        feed_b = self._feed([(0, 0), (1, 1), (2, 2), (3, 3)])
        by_light = self.m.align_detections([feed_a, feed_b])
        self.assertEqual(sorted(by_light), [0, 1, 2, 3])
        for slot in range(4):
            feeds = {fi: cx for fi, cx, _ in by_light[slot]}
            self.assertEqual(feeds, {0: slot * 10, 1: slot * 10},
                             f"slot {slot} mismatched correspondence")

    def test_align_dropped_blink_no_mislabel(self):
        # feed 1 misses light 2; lights 3,4 must still correspond to lights 3,4
        # (not slide up), and slot 2 must contain only feed 0.
        feed_a = self._feed([(0, 0), (1, 1), (2, 2), (3, 3), (4, 4)])
        feed_b = self._feed([(0, 0), (1, 1), (3, 3), (4, 4)])  # light 2 dropped
        by_light = self.m.align_detections([feed_a, feed_b])

        self.assertEqual({fi for fi, _, _ in by_light[2]}, {0},
                         "occluded slot 2 must have only feed 0")
        # Light 3 and 4: both feeds present and centroids agree (no shift).
        for slot in (3, 4):
            feeds = {fi: cx for fi, cx, _ in by_light[slot]}
            self.assertEqual(
                feeds, {0: slot * 10, 1: slot * 10},
                f"slot {slot} got shifted/mismatched correspondence: {feeds}",
            )

    def test_align_spurious_blink_no_downstream_shift(self):
        # feed 1 has a spurious blink between lights 1 and 2; downstream lights
        # 2 and 3 must keep correct correspondence (slot 1 becomes ambiguous and
        # is dropped for that feed rather than shifting everything).
        feed_a = self._feed([(0, 0), (1, 1), (2, 2), (3, 3)])
        feed_b = [(0.0, 0.0, 0.0), (1.0, 10.0, 0.0), (1.4, 999.0, 0.0),
                  (2.0, 20.0, 0.0), (3.0, 30.0, 0.0)]
        by_light = self.m.align_detections([feed_a, feed_b])

        # Slot 1 ambiguous for feed 1 → only feed 0 contributes there.
        self.assertEqual({fi for fi, _, _ in by_light[1]}, {0},
                         "ambiguous slot 1 must drop feed 1's contribution")
        # The spurious centroid (999) must never end up labelling any light.
        all_cx = {cx for dets in by_light.values() for _, cx, _ in dets}
        self.assertNotIn(999.0, all_cx, "spurious blink leaked into a light slot")
        # Lights 2 and 3 keep correct cross-feed correspondence.
        for slot in (2, 3):
            feeds = {fi: cx for fi, cx, _ in by_light[slot]}
            self.assertEqual(
                feeds, {0: slot * 10, 1: slot * 10},
                f"slot {slot} shifted after spurious blink: {feeds}",
            )


class TestTriangulationSanityGuards(unittest.TestCase):
    """
    Unit tests for WI-26 triangulation sanity guards.

    A two-camera stereo rig (camera 0 at the origin, camera 1 displaced 0.30 m
    along X, both looking down +Z) is built directly so crafted correspondences
    can drive `triangulate_all` past the geometry/finite guards without rendering
    video.  Light centroids are the *exact* projections of a chosen world point,
    so the only variable under test is the guard behaviour.
    """

    BASELINE_M = 0.30
    F          = 320.0
    W, H       = 320, 240

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()
        cls.K = np.array([[cls.F, 0.0,   cls.W / 2.0],
                          [0.0,   cls.F, cls.H / 2.0],
                          [0.0,   0.0,   1.0        ]], dtype=np.float64)
        R0 = np.eye(3, dtype=np.float64)
        t0 = np.zeros((3, 1), dtype=np.float64)
        R1 = np.eye(3, dtype=np.float64)
        t1 = np.array([[-cls.BASELINE_M], [0.0], [0.0]], dtype=np.float64)
        cls.poses = [(R0, t0), (R1, t1)]

    def _project(self, pose, pt_world) -> tuple[float, float]:
        R, t = pose
        c = R @ np.asarray(pt_world, dtype=np.float64) + t.ravel()
        return (self.F * c[0] / c[2] + self.W / 2.0,
                self.F * c[1] / c[2] + self.H / 2.0)

    def _by_light_for(self, pt_world) -> dict:
        cx0, cy0 = self._project(self.poses[0], pt_world)
        cx1, cy1 = self._project(self.poses[1], pt_world)
        return {0: [(0, cx0, cy0), (1, cx1, cy1)]}

    def test_point_in_front_is_triangulated(self):
        """Sanity baseline: a point in front of both cameras is recovered."""
        by_light = self._by_light_for((0.0, 0.0, 1.0))
        lights_3d, missing, _ = self.m.triangulate_all(
            by_light, self.poses, [self.K, self.K], 1.0
        )
        self.assertIn(0, lights_3d, "in-front point should triangulate")
        self.assertEqual(missing, [])
        self.assertAlmostEqual(lights_3d[0][2], 1.0, delta=1e-3)

    def test_point_behind_camera_is_missing(self):
        """
        A correspondence that triangulates to a point *behind* the cameras
        (negative depth) must be reported missing — never emitted as a light.
        """
        # (0, 0, -1.0) is behind both cameras (they look down +Z) but still
        # projects to valid pixel coordinates, so DLT recovers it.
        by_light = self._by_light_for((0.0, 0.0, -1.0))
        lights_3d, missing, low_conf = self.m.triangulate_all(
            by_light, self.poses, [self.K, self.K], 1.0
        )
        self.assertNotIn(0, lights_3d,
                         "point behind camera must NOT get coordinates")
        self.assertIn(0, missing, "point behind camera must be in 'missing'")
        self.assertNotIn(0, low_conf)

    def test_non_finite_triangulation_is_missing_and_json_valid(self):
        """
        Forcing a NaN triangulation routes the light to missing, and the emitted
        result is valid RFC8259 JSON (no `NaN` token; Go-side parse succeeds).
        """
        by_light = self._by_light_for((0.0, 0.0, 1.0))

        original = self.m._dlt_triangulate
        self.m._dlt_triangulate = lambda valid, Ps: np.array(
            [np.nan, 0.0, 1.0], dtype=np.float64
        )
        try:
            lights_3d, missing, _ = self.m.triangulate_all(
                by_light, self.poses, [self.K, self.K], 1.0
            )
        finally:
            self.m._dlt_triangulate = original

        self.assertNotIn(0, lights_3d, "NaN point must NOT get coordinates")
        self.assertIn(0, missing, "NaN point must be in 'missing'")

        result = self.m._make_result(lights_3d, missing, [], {0})
        out = self.m._serialise_result(result)
        self.assertNotIn("NaN", out, "emitted JSON contains a non-RFC8259 token")
        parsed = json.loads(out)  # would still parse NaN; assertion above is real
        self.assertEqual(parsed["status"], "succeeded")
        self.assertIn(0, parsed["missing"])

    def test_serialise_result_rejects_non_finite(self):
        """
        Final safety net: if a non-finite coordinate somehow reaches the result
        dict, `_serialise_result` must NOT emit a `NaN`/`Infinity` token; it
        fails cleanly with status:"failed" so the Go parser never chokes.
        """
        bad = self.m._make_result({0: (float("nan"), 0.0, 1.0)}, [], [], {0})
        out = self.m._serialise_result(bad)
        self.assertNotIn("NaN", out)
        self.assertNotIn("Infinity", out)
        parsed = json.loads(out)
        self.assertEqual(parsed["status"], "failed")
        self.assertIsInstance(parsed["error"], str)


class TestPerFeedIntrinsics(unittest.TestCase):
    """
    Unit tests for WI-28 per-feed intrinsics in the essential-matrix path.

    Feed 0 uses 320×240 intrinsics; feed 1 uses 480×360 (different focal length
    and principal point).  The old code passed Ks[0] into findEssentialMat and
    recoverPose for both feeds, which biases pose recovery when resolutions differ.
    """

    BASELINE_M = 0.30
    W0, H0 = 320, 240
    W1, H1 = 480, 360

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()
        cls.K0 = cls.m._estimate_K(cls.W0, cls.H0)
        cls.K1 = cls.m._estimate_K(cls.W1, cls.H1)
        R0 = np.eye(3, dtype=np.float64)
        t0 = np.zeros((3, 1), dtype=np.float64)
        R1 = np.eye(3, dtype=np.float64)
        t1 = np.array([[-cls.BASELINE_M], [0.0], [0.0]], dtype=np.float64)
        cls.gt_poses = [(R0, t0), (R1, t1)]

    def _project(self, K, pose, pt_world) -> tuple[float, float]:
        R, t = pose
        c = R @ np.asarray(pt_world, dtype=np.float64) + t.ravel()
        return (
            K[0, 0] * c[0] / c[2] + K[0, 2],
            K[1, 1] * c[1] / c[2] + K[1, 2],
        )

    def _synthetic_by_light(self, world_pts: list[tuple[float, float, float]]) -> dict:
        by_light: dict[int, list[tuple[int, float, float]]] = {}
        Ks = [self.K0, self.K1]
        for lid, pt in enumerate(world_pts):
            dets = []
            for fi, pose in enumerate(self.gt_poses):
                cx, cy = self._project(Ks[fi], pose, pt)
                dets.append((fi, cx, cy))
            by_light[lid] = dets
        return by_light

    def test_mixed_intrinsics_pose_recovery(self):
        """E-matrix path with per-feed K recovers the known relative pose."""
        world_pts = [
            (0.0, 0.0, 1.0),
            (0.05, 0.03, 0.9),
            (-0.04, 0.06, 1.1),
            (0.02, -0.05, 0.95),
            (-0.03, -0.02, 1.05),
            (0.06, 0.0, 1.0),
        ]
        by_light = self._synthetic_by_light(world_pts)

        poses, scale = self.m.estimate_poses(
            feed_paths=["a.avi", "b.avi"],
            Ks=[self.K0, self.K1],
            marker_spec=None,
            by_light=by_light,
            scale_hint_m=self.BASELINE_M,
        )

        self.assertEqual(scale, self.BASELINE_M)
        self.assertIsNotNone(poses[1])
        R1, t1 = poses[1]  # type: ignore[misc]
        _, gt_t1 = self.gt_poses[1]
        # Translation direction and magnitude (E-matrix is up to scale).
        gt_dir = gt_t1.ravel() / np.linalg.norm(gt_t1)
        est_dir = t1.ravel() / np.linalg.norm(t1)
        self.assertGreater(np.dot(gt_dir, est_dir), 0.99,
                           "recovered translation direction is wrong")
        np.testing.assert_allclose(R1, self.gt_poses[1][0], atol=0.05)

    def test_mixed_intrinsics_triangulation_accuracy(self):
        """Triangulation after mixed-intrinsics pose recovery is near ground truth."""
        world_pts = [
            (0.0, 0.0, 1.0),
            (0.05, 0.03, 0.9),
            (-0.04, 0.06, 1.1),
            (0.02, -0.05, 0.95),
            (-0.03, -0.02, 1.05),
            (0.06, 0.0, 1.0),
        ]
        by_light = self._synthetic_by_light(world_pts)

        poses, scale = self.m.estimate_poses(
            feed_paths=["a.avi", "b.avi"],
            Ks=[self.K0, self.K1],
            marker_spec=None,
            by_light=by_light,
            scale_hint_m=self.BASELINE_M,
        )
        lights_3d, missing, _ = self.m.triangulate_all(
            by_light, poses, [self.K0, self.K1], scale
        )

        self.assertEqual(missing, [], f"unexpected missing: {missing}")
        for lid, pt in enumerate(world_pts):
            self.assertIn(lid, lights_3d)
            x, y, z = lights_3d[lid]
            dist = math.sqrt(
                (x - pt[0]) ** 2 + (y - pt[1]) ** 2 + (z - pt[2]) ** 2
            )
            self.assertLess(
                dist, 0.02,
                f"light {lid}: dist={dist:.4f} m with per-feed intrinsics",
            )

    def test_single_k_for_both_feeds_is_biased(self):
        """Using K0 for both feeds (old bug) fails cheirality or is materially wrong."""
        world_pts = [
            (0.0, 0.0, 1.0),
            (0.05, 0.03, 0.9),
            (-0.04, 0.06, 1.1),
            (0.02, -0.05, 0.95),
            (-0.03, -0.02, 1.05),
            (0.06, 0.0, 1.0),
        ]
        by_light = self._synthetic_by_light(world_pts)
        pts0, pts1 = self.m._common_2d_points(by_light, 0, 1)

        # Old behaviour: both point sets forced through K0.
        E, mask = cv2.findEssentialMat(
            pts0, pts1, self.K0, method=cv2.RANSAC, prob=0.999, threshold=1.0
        )
        self.assertIsNotNone(E)
        if E.shape != (3, 3):
            E = E[:3, :]
        inliers = mask.ravel() == 1
        _, R_wrong, t_wrong, _ = cv2.recoverPose(
            E, pts0[inliers], pts1[inliers], self.K0
        )
        wrong_poses = [
            self.gt_poses[0],
            (R_wrong, t_wrong),
        ]
        lights_wrong, miss_wrong, _ = self.m.triangulate_all(
            by_light, wrong_poses, [self.K0, self.K1], self.BASELINE_M
        )

        em = self.m._essential_matrix_pose(
            pts0, pts1, self.K0, self.K1, by_light
        )
        self.assertIsNotNone(em)
        R1, t1, _ = em
        poses_ok = [
            self.gt_poses[0],
            (R1, t1),
        ]
        lights_ok, miss_ok, _ = self.m.triangulate_all(
            by_light, poses_ok, [self.K0, self.K1], self.BASELINE_M
        )

        self.assertEqual(miss_ok, [])
        max_ok_err = max(
            math.sqrt(sum((lights_ok[j][k] - world_pts[j][k]) ** 2 for k in range(3)))
            for j in lights_ok
        )
        self.assertLess(max_ok_err, 0.02)

        # Old path must either miss lights (cheirality failure) or land far off.
        if lights_wrong:
            max_wrong_err = max(
                math.sqrt(sum((lights_wrong[j][k] - world_pts[j][k]) ** 2 for k in range(3)))
                for j in lights_wrong
            )
            self.assertGreater(max_wrong_err, 0.05)
        else:
            self.assertGreater(len(miss_wrong), 0)


class TestRobustBlinkDetection(unittest.TestCase):
    """
    WI-29: robust background model and reliable frame rewind.

    Two issues addressed:
      1. CAP_PROP_POS_FRAMES = 0 seeking is unreliable on some MP4/MKV
         backends; the new code re-opens a fresh VideoCapture for the
         detection pass so frame alignment is deterministic.
      2. The old background was the per-pixel minimum of the first 10
         *consecutive* frames.  If a light is on throughout that window
         its pixel never appears dark in the background → missed blink.
         The new background samples BG_FRAMES frames spread across the
         full video, so every light is captured in its off-state in at
         least one sample.
    """

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    # ── helper: write a minimal single-light video ────────────────────────────

    @staticmethod
    def _write_test_video(
        path: str,
        *,
        width: int = 160,
        height: int = 120,
        fps: int = 30,
        leading_gap_frames: int = 0,
        dwell_frames: int = 10,
        gap_frames: int = 2,
        n_lights: int = 1,
        blob_cx: int = 80,
        blob_cy: int = 60,
        blob_radius: int = 5,
    ) -> None:
        """
        Writes a simple AVI to *path* with *n_lights* sequential blink events.
        Each light is a white circle at (blob_cx, blob_cy) for *dwell_frames*
        followed by *gap_frames* of darkness.  A *leading_gap_frames* dark
        prefix is prepended before light 0.
        """
        fourcc = cv2.VideoWriter_fourcc(*"XVID")
        out = cv2.VideoWriter(path, fourcc, fps, (width, height))
        blank = np.zeros((height, width, 3), dtype=np.uint8)

        def _dark():
            out.write(blank.copy())

        def _lit():
            f = blank.copy()
            cv2.circle(f, (blob_cx, blob_cy), blob_radius, (255, 255, 255), -1)
            out.write(f)

        for _ in range(leading_gap_frames):
            _dark()
        for _ in range(n_lights):
            for _ in range(dwell_frames):
                _lit()
            for _ in range(gap_frames):
                _dark()

        out.release()

    # ── test 1: always-on during the initial background window ───────────────

    def test_always_on_during_background_window_still_detected(self):
        """
        WI-29 issue 2: a light that is on throughout the first BG_FRAMES
        consecutive frames must still be detected after the fix.

        Fixture: leading_gap_ms=0 so light 0 starts at frame 0 and is on
        for all of the first BG_FRAMES=10 frames.  With the old code (min
        of first 10 frames), every sample for that pixel is lit → background
        is bright → diff ≈ 0 → blink missed.  With the new code (spread
        samples across the full video), the samples taken during later lights'
        on-periods catch light 0 in its off-state → background is dark →
        blink detected.
        """
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(
                d,
                n_lights=5,
                seed=29,
                leading_gap_ms=0,   # light 0 starts at frame 0
            )
            res = _reconstruct(_spec(d, gt))

        self.assertEqual(res["status"], "succeeded", res.get("error"))

        recovered_ids = {lp["id"] for lp in res["lights"]}
        missing_ids   = set(res["missing"])

        # Light 0 is the one at risk (it blinks from frame 0, contaminating
        # the old consecutive-frame background).  It must be detected.
        self.assertIn(
            0, recovered_ids,
            f"Light 0 (on at frame 0, no initial gap) was not detected — "
            f"background contamination fix may not be working. "
            f"recovered={sorted(recovered_ids)}, missing={sorted(missing_ids)}",
        )

        # At least N-1 lights must be recovered overall.
        n = gt["light_count"] if "light_count" in gt else len(gt["lights"])
        self.assertGreaterEqual(
            len(res["lights"]),
            n - 1,
            f"Expected ≥{n - 1} lights; got {len(res['lights'])}",
        )

    # ── test 2: re-open path / unit test of detect_blinks ───────────────────

    def test_detect_blinks_unit_correct_count_and_position(self):
        """
        WI-29 issue 1: detect_blinks called directly must return the correct
        blink count and centroid even though the implementation no longer uses
        CAP_PROP_POS_FRAMES = 0 to rewind.

        A synthetic AVI is written with a known blink pattern (3 lights, each
        lit for 10 frames, separated by 2-frame dark gaps, with a 3-frame
        leading gap so the background pass sees all lights dark at least once).
        detect_blinks must find all 3 blinks at the expected centroid.
        """
        m = self.m
        BG_FRAMES = m.BG_FRAMES

        WIDTH, HEIGHT = 160, 120
        DWELL   = 10
        GAP     = 3
        LEADING = BG_FRAMES + 2    # guarantee at least one fully dark sample
        N_LIGHTS = 3
        BLOB_CX, BLOB_CY = 80, 60
        FPS = 30

        with tempfile.TemporaryDirectory() as d:
            video_path = str(Path(d) / "test_blinks.avi")
            self._write_test_video(
                video_path,
                width=WIDTH,
                height=HEIGHT,
                fps=FPS,
                leading_gap_frames=LEADING,
                dwell_frames=DWELL,
                gap_frames=GAP,
                n_lights=N_LIGHTS,
                blob_cx=BLOB_CX,
                blob_cy=BLOB_CY,
            )

            blinks, fw, fh, K = m.detect_blinks(
                video_path,
                dwell_ms=round(DWELL * 1000 / FPS),
            )

        self.assertEqual(fw, WIDTH)
        self.assertEqual(fh, HEIGHT)
        self.assertEqual(
            len(blinks), N_LIGHTS,
            f"Expected {N_LIGHTS} blinks, got {len(blinks)}: {blinks}",
        )
        for t_s, cx, cy in blinks:
            self.assertAlmostEqual(cx, BLOB_CX, delta=4,
                                   msg=f"centroid x={cx:.1f} off from {BLOB_CX}")
            self.assertAlmostEqual(cy, BLOB_CY, delta=4,
                                   msg=f"centroid y={cy:.1f} off from {BLOB_CY}")

    def test_detect_blinks_unit_no_leading_gap(self):
        """
        WI-29: detect_blinks must find a blink that starts at frame 0 (no
        leading gap) when the spread-background fix is in place.

        The background is built from samples spread across the whole clip.
        Since the light is only on for the first dwell_frames frames and dark
        afterwards, at least one spread sample will be dark → background is
        dark at that pixel → blink detected.
        """
        m = self.m
        WIDTH, HEIGHT = 160, 120
        DWELL = 10
        GAP   = 3
        FPS   = 30

        with tempfile.TemporaryDirectory() as d:
            video_path = str(Path(d) / "no_gap.avi")
            # 3 lights, no leading gap: light 0 is on from frame 0.
            self._write_test_video(
                video_path,
                width=WIDTH,
                height=HEIGHT,
                fps=FPS,
                leading_gap_frames=0,
                dwell_frames=DWELL,
                gap_frames=GAP,
                n_lights=3,
            )

            blinks, _, _, _ = m.detect_blinks(
                video_path,
                dwell_ms=round(DWELL * 1000 / FPS),
            )

        # All 3 blinks (including light 0 that starts at frame 0) must be
        # found.  The old code (min of first BG_FRAMES consecutive frames)
        # would miss light 0 because every early sample sees it on.
        self.assertEqual(
            len(blinks), 3,
            f"Expected 3 blinks (incl. one at frame 0), got {len(blinks)}: {blinks}",
        )


class TestPythonRobustnessMisc(unittest.TestCase):
    """WI-30: scale validation, dwell filtering, exit codes, capture release."""

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def test_invalid_scale_hint_zero(self):
        with tempfile.TemporaryDirectory() as d:
            res, code = _reconstruct_with_exit(
                _spec(d, {"dwell_ms": 1000}, {"scale_hint_m": 0})
            )
        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertIn("scale_hint_m", res["error"])

    def test_invalid_scale_hint_negative(self):
        with tempfile.TemporaryDirectory() as d:
            res, code = _reconstruct_with_exit(
                _spec(d, {"dwell_ms": 1000}, {"scale_hint_m": -0.3})
            )
        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertIn("scale_hint_m", res["error"])

    def test_invalid_marker_edge_length(self):
        with tempfile.TemporaryDirectory() as d:
            res, code = _reconstruct_with_exit(
                _spec(
                    d,
                    {"dwell_ms": 1000},
                    {
                        "marker": {
                            "dictionary": "DICT_4X4_50",
                            "edge_length_m": 0,
                        }
                    },
                )
            )
        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertIn("edge_length_m", res["error"])

    def test_no_blinks_exits_nonzero(self):
        """Logical failure must emit status failed JSON and a non-zero exit code."""
        with tempfile.TemporaryDirectory() as d:
            gt = {"dwell_ms": 1000}
            for i in range(2):
                path = Path(d) / f"feed_{i}.avi"
                fourcc = cv2.VideoWriter_fourcc(*"XVID")
                out = cv2.VideoWriter(str(path), fourcc, 30, (160, 120))
                blank = np.zeros((120, 160, 3), dtype=np.uint8)
                for _ in range(60):
                    out.write(blank)
                out.release()

            res, code = _reconstruct_with_exit(_spec(d, gt))

        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertIn("no blink", res["error"].lower())

    def test_continuously_bright_clip_explains_missing_sweep(self):
        """A clip that stays bright for the whole recording must say so.

        A capture sweep turns one bulb on for about a second. If the string
        stays lit, the detector sees one long flash and must tell the operator
        to film the sweep instead of the generic "no blink events" line.
        """
        with tempfile.TemporaryDirectory() as d:
            gt = {"dwell_ms": 1000}
            for i in range(2):
                path = Path(d) / f"feed_{i}.avi"
                fourcc = cv2.VideoWriter_fourcc(*"XVID")
                out = cv2.VideoWriter(str(path), fourcc, 30, (160, 120))
                dark = np.zeros((120, 160, 3), dtype=np.uint8)
                bright = np.full((120, 160, 3), 200, dtype=np.uint8)
                out.write(dark)
                for _ in range(149):
                    out.write(bright)
                out.release()

            res, code = _reconstruct_with_exit(_spec(d, gt))

        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        err = res["error"].lower()
        self.assertIn("stays bright", err)
        self.assertIn("capture sweep", err)
        self.assertNotIn("window", err)

    def test_short_flash_rejected_by_dwell_validation(self):
        """A one-frame flash must not count as a blink when dwell_ms is set."""
        m = self.m
        fps = 30
        dwell_frames = 10
        dwell_ms = round(dwell_frames * 1000 / fps)

        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "flash.avi")
            fourcc = cv2.VideoWriter_fourcc(*"XVID")
            out = cv2.VideoWriter(path, fourcc, fps, (160, 120))
            blank = np.zeros((120, 160, 3), dtype=np.uint8)
            lit = blank.copy()
            cv2.circle(lit, (80, 60), 5, (255, 255, 255), -1)

            for _ in range(m.BG_FRAMES + 5):
                out.write(blank)
            out.write(lit)  # ~33 ms flash — below DWELL_MIN_FRAC × dwell_ms
            for _ in range(5):
                out.write(blank)
            for _ in range(dwell_frames):
                out.write(lit)
            for _ in range(5):
                out.write(blank)
            out.release()

            blinks, _, _, _ = m.detect_blinks(path, dwell_ms=dwell_ms)

        self.assertEqual(
            len(blinks),
            1,
            f"short flash should be rejected; got {len(blinks)} blink(s): {blinks}",
        )

    def test_video_capture_released_on_read_error(self):
        """VideoCapture.release() must run even when detection raises."""
        m = self.m
        released: list[str] = []

        class FakeCap:
            _seq = 0

            def __init__(self, path):
                self.path = path
                self.instance_id = FakeCap._seq
                FakeCap._seq += 1
                self.read_calls = 0

            def isOpened(self):
                return True

            def get(self, prop):
                if prop == cv2.CAP_PROP_FRAME_WIDTH:
                    return 160
                if prop == cv2.CAP_PROP_FRAME_HEIGHT:
                    return 120
                if prop == cv2.CAP_PROP_FPS:
                    return 30
                if prop == cv2.CAP_PROP_FRAME_COUNT:
                    return 120
                return 0

            def read(self):
                self.read_calls += 1
                if self.instance_id >= 1 and self.read_calls == 1:
                    raise RuntimeError("simulated decode failure")
                if self.read_calls > 120:
                    return False, None
                frame = np.zeros((120, 160, 3), dtype=np.uint8)
                return True, frame

            def release(self):
                released.append(self.path)

        FakeCap._seq = 0
        with patch.object(cv2, "VideoCapture", FakeCap):
            with self.assertRaises(RuntimeError):
                m.detect_blinks("/fake/video.avi", dwell_ms=333)

        self.assertEqual(
            len(released),
            FakeCap._seq,
            f"expected every capture released, got {len(released)} release(s)",
        )


class TestBookends(unittest.TestCase):
    """Red-blue-green start and end flashes anchor light numbering."""

    FPS = 30
    W, H = 160, 120
    # BGR colours of every LED during a flash.
    RED, BLUE, GREEN = (0, 0, 255), (255, 0, 0), (0, 255, 0)
    STRING = [(30, 40), (80, 40), (130, 40), (30, 90), (80, 90), (130, 90)]

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def _frame(self, spots=(), colour=(255, 255, 255)):
        f = np.zeros((self.H, self.W, 3), dtype=np.uint8)
        for cx, cy in spots:
            cv2.circle(f, (cx, cy), 6, colour, -1)
        return f

    def _dark(self, n):
        return [self._frame() for _ in range(n)]

    def _cue(self, per_colour=6, gap=6):
        frames = []
        for i, colour in enumerate((self.RED, self.BLUE, self.GREEN)):
            if i:
                frames += self._dark(gap)
            frames += [self._frame(self.STRING, colour) for _ in range(per_colour)]
        return frames

    def _write(self, path, frames):
        out = cv2.VideoWriter(
            str(path), cv2.VideoWriter_fourcc(*"XVID"), self.FPS, (self.W, self.H)
        )
        for f in frames:
            out.write(f)
        out.release()

    @staticmethod
    def _feed(name, blinks, cues):
        return {
            "name": name,
            "blinks": [(float(t), float(cx), 0.0) for t, cx in blinks],
            "cues": [(float(a), float(b)) for a, b in cues],
        }

    def test_find_cues_finds_one_red_blue_green_cue(self):
        lead = 15
        frames = self._dark(lead) + self._cue() + self._dark(15)
        with tempfile.TemporaryDirectory() as d:
            path = Path(d) / "cue.avi"
            self._write(path, frames)
            cues = self.m.find_cues(str(path))

        self.assertEqual(len(cues), 1, f"cues={cues}")
        start, end = cues[0]
        self.assertAlmostEqual(start, lead / self.FPS, delta=1.5 / self.FPS)
        self.assertAlmostEqual(end, (lead + 30) / self.FPS, delta=1.5 / self.FPS)

    def test_find_cues_tolerates_one_frame_smear_in_dark_gaps(self):
        lead = 15

        def cue_with_gap_smears():
            frames = []
            for i, colour in enumerate((self.RED, self.BLUE, self.GREEN)):
                if i:
                    gap = self._dark(6)
                    gap[len(gap) // 2] = self._frame(self.STRING, (255, 255, 255))
                    frames += gap
                frames += [self._frame(self.STRING, colour) for _ in range(6)]
            return frames

        frames = self._dark(lead) + cue_with_gap_smears() + self._dark(15)
        with tempfile.TemporaryDirectory() as d:
            path = Path(d) / "cue_smear.avi"
            self._write(path, frames)
            cues = self.m.find_cues(str(path))

        self.assertEqual(len(cues), 1, f"cues={cues}")

    def test_blinks_outside_both_cues_ignored_and_first_is_slot_zero(self):
        feed = self._feed(
            "a.mp4",
            [(1.0, 999), (4.0, 0), (5.0, 10), (6.0, 20), (12.0, 888)],
            [(2.0, 3.0), (10.0, 11.0)],
        )
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(rejected, [])
        self.assertEqual(sorted(by_light), [0, 1, 2])
        for slot in range(3):
            self.assertEqual(by_light[slot], [(0, slot * 10.0, 0.0)])

    def test_closing_only_counts_back_from_light_count(self):
        feed = self._feed("late.mp4", [(6.0, 20), (7.0, 30), (8.0, 40)], [(10.0, 11.0)])
        by_light, rejected = self.m.number_from_bookends([feed], 5)

        self.assertEqual(rejected, [])
        self.assertEqual(sorted(by_light), [2, 3, 4])
        self.assertEqual(by_light[2], [(0, 20.0, 0.0)])

    def test_closing_only_without_count_copies_last_index_from_opening_clip(self):
        full = self._feed(
            "full.mp4",
            [(4.0, 0), (5.0, 10), (6.0, 20), (7.0, 30), (8.0, 40)],
            [(2.0, 3.0), (10.0, 11.0)],
        )
        late = self._feed("late.mp4", [(6.2, 20), (7.2, 30), (8.2, 40)], [(10.2, 11.2)])
        by_light, rejected = self.m.number_from_bookends([full, late], None)

        self.assertEqual(rejected, [])
        self.assertEqual({fi for fi, _, _ in by_light[4]}, {0, 1})
        self.assertEqual({fi for fi, _, _ in by_light[2]}, {0, 1})
        self.assertEqual({fi for fi, _, _ in by_light[0]}, {0})

    def test_closing_only_without_count_or_opening_clip_is_dropped(self):
        late = self._feed("late.mp4", [(6.0, 20), (7.0, 30)], [(10.0, 11.0)])
        by_light, rejected = self.m.number_from_bookends([late], None)

        self.assertEqual(by_light, {})
        self.assertEqual(len(rejected), 1)
        reason = rejected[0]["reason"].lower()
        self.assertIn("opening flash", reason)
        self.assertIn("light count", reason)

    def test_clip_without_cue_is_rejected(self):
        feed = self._feed("nocue.mp4", [(1.0, 0), (2.0, 10)], [])
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(by_light, {})
        self.assertEqual(rejected, [{"file": "nocue.mp4", "reason": "no start or end signal found"}])

    def test_one_cue_with_blinks_on_both_sides_is_ambiguous(self):
        feed = self._feed("amb.mp4", [(1.0, 0), (5.0, 10), (6.0, 20)], [(2.0, 3.0)])
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(by_light, {})
        self.assertIn("ambiguous", rejected[0]["reason"])

    def test_three_cues_are_ambiguous(self):
        feed = self._feed(
            "three.mp4", [(4.0, 0), (5.0, 10)], [(2.0, 3.0), (6.0, 7.0), (9.0, 10.0)]
        )
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(by_light, {})
        self.assertIn("ambiguous", rejected[0]["reason"])

    def test_light_count_drops_high_indexes(self):
        feed = self._feed(
            "a.mp4", [(4.0, 0), (5.0, 10), (6.0, 20)], [(2.0, 3.0), (10.0, 11.0)]
        )
        by_light, _ = self.m.number_from_bookends([feed], 2)
        self.assertEqual(sorted(by_light), [0, 1])

    def test_one_second_flash_before_opening_cue_is_ignored(self):
        spots = self.STRING
        flash_spot = (80, 65)
        frames = (
            self._dark(12)
            + [self._frame([flash_spot]) for _ in range(30)]
            + self._dark(12)
            + self._cue()
            + self._dark(15)
        )
        for k in range(3):
            frames += [self._frame([spots[k]]) for _ in range(30)]
            frames += self._dark(3)
        frames += self._dark(15) + self._cue() + self._dark(12)

        with tempfile.TemporaryDirectory() as d:
            path = Path(d) / "flash.avi"
            self._write(path, frames)
            blinks, cues, _fw, _fh, _K = self.m.analyse_feed(str(path), 1000)

        # A pre-cue flash is not a bulb. The quiet-room picture may absorb it,
        # and numbering still drops anything before the opening cue.
        self.assertEqual(len(cues), 2, f"cues={cues}")
        inside = [b for b in blinks if cues[0][1] <= b[0] < cues[1][0]]
        self.assertEqual(len(inside), 3, f"blinks={blinks}")
        for k, (_t, cx, cy) in enumerate(inside):
            self.assertAlmostEqual(cx, spots[k][0], delta=4)
            self.assertAlmostEqual(cy, spots[k][1], delta=4)
        by_light, rejected = self.m.number_from_bookends(
            [{"name": "flash.avi", "blinks": blinks, "cues": cues}], None
        )
        self.assertEqual(rejected, [])
        self.assertEqual(sorted(by_light), [0, 1, 2])
        for k in range(3):
            (_fi, cx, cy), = by_light[k]
            self.assertAlmostEqual(cx, spots[k][0], delta=4)
            self.assertAlmostEqual(cy, spots[k][1], delta=4)
            self.assertGreater(
                math.hypot(cx - flash_spot[0], cy - flash_spot[1]), 4
            )

    def test_job_fails_when_clips_have_no_signal(self):
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=5, seed=11, no_bookend=True)
            res, code = _reconstruct_with_exit(_spec(d, gt))

        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        err = res["error"]
        self.assertIn("start and end flashes", err)
        self.assertIn("feed_0.avi", err)
        self.assertIn("feed_1.avi", err)
        self.assertIn("no start or end signal found", err)
        self.assertEqual(res["rejected_feeds"], [
            {"file": "feed_0.avi", "reason": "no start or end signal found"},
            {"file": "feed_1.avi", "reason": "no start or end signal found"},
        ])

    def test_invalid_light_count_fails_job(self):
        for bad in (0, 1001, 2.5, "5", True):
            with tempfile.TemporaryDirectory() as d:
                res, code = _reconstruct_with_exit(
                    _spec(d, {"dwell_ms": 1000}, {"light_count": bad})
                )
            self.assertEqual(code, 1, bad)
            self.assertEqual(
                res["error"], "light_count must be a whole number from 1 to 1000"
            )
            self.assertEqual(res["rejected_feeds"], [])

    def test_success_reports_rejected_feeds_and_light_count(self):
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=6, seed=12)
            res = _reconstruct(_spec(d, gt, {"light_count": 8}))

        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(res["rejected_feeds"], [])
        self.assertEqual(res["light_count"], 8)
        self.assertIn(6, res["missing"])
        self.assertIn(7, res["missing"])


class TestQuietRoomHelpers(unittest.TestCase):
    """REQ-051: which frames are the room, and what counts as brighter."""

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def test_quiet_room_reason_sentence(self):
        self.assertEqual(
            self.m.QUIET_ROOM_REASON,
            "the recording needs a moment of the room with the bulbs off, "
            "before the opening flash or after the closing flash",
        )

    def test_empty_cues_mean_legacy_background(self):
        interval, reason = self.m._quiet_interval([], 100, 30.0)
        self.assertIsNone(interval)
        self.assertIsNone(reason)

    def test_two_bookends_prefer_before_when_long_enough(self):
        interval, reason = self.m._quiet_interval([(15, 30), (100, 115)], 130, 30.0)
        self.assertIsNone(reason)
        self.assertEqual(interval, ("before", 0, 15))

    def test_two_bookends_fall_back_to_after(self):
        interval, reason = self.m._quiet_interval([(3, 20), (40, 55)], 80, 30.0)
        self.assertIsNone(reason)
        self.assertEqual(interval, ("after", 56, 80))

    def test_two_bookends_reject_when_both_sides_are_short(self):
        interval, reason = self.m._quiet_interval([(3, 20), (23, 40)], 43, 30.0)
        self.assertIsNone(interval)
        self.assertEqual(reason, self.m.QUIET_ROOM_REASON)

    def test_opening_only_uses_before_not_the_sweep(self):
        interval, reason = self.m._quiet_interval([(15, 40)], 200, 30.0)
        self.assertIsNone(reason)
        self.assertEqual(interval, ("before", 0, 15))

    def test_closing_only_uses_after_not_the_sweep(self):
        interval, reason = self.m._quiet_interval([(90, 110)], 120, 30.0)
        self.assertIsNone(reason)
        self.assertEqual(interval, ("after", 111, 120))

    def test_single_bookend_tie_counts_as_closing(self):
        interval, reason = self.m._quiet_interval([(10, 19)], 30, 30.0)
        self.assertIsNone(reason)
        self.assertEqual(interval, ("after", 20, 30))

    def test_opening_only_short_lead_does_not_use_the_sweep(self):
        interval, reason = self.m._quiet_interval([(3, 20)], 200, 30.0)
        self.assertIsNone(interval)
        self.assertEqual(reason, self.m.QUIET_ROOM_REASON)

    def test_positive_excess_removes_frame_wide_lift_and_ignores_darkening(self):
        frame = np.full((4, 4), 40, np.uint8)
        frame[0, 0] = 255
        excess = self.m._positive_excess(frame, np.zeros((4, 4), np.uint8))
        self.assertEqual(int(excess[0, 0]), 215)
        self.assertEqual(int(excess[1, 1]), 0)
        dark = self.m._positive_excess(
            np.zeros((2, 2), np.uint8), np.full((2, 2), 180, np.uint8)
        )
        self.assertTrue(np.all(dark == 0))

    def test_boundary_moved_tolerates_two_frames(self):
        old = [(10, 40), (80, 100)]
        self.assertFalse(self.m._boundary_moved(old, [(12, 40), (80, 100)], "before"))
        self.assertTrue(self.m._boundary_moved(old, [(13, 40), (80, 100)], "before"))
        self.assertFalse(self.m._boundary_moved(old, [(10, 40), (80, 102)], "after"))
        self.assertTrue(self.m._boundary_moved(old, [(10, 40), (80, 103)], "after"))
        self.assertTrue(self.m._boundary_moved(old, [(10, 40)], "before"))

    def test_sample_indexes_cover_short_and_long_spans(self):
        self.assertEqual(self.m._sample_indexes(0, 9, 15), list(range(9)))
        indexes = self.m._sample_indexes(5, 35, 15)
        self.assertEqual(len(indexes), 15)
        self.assertEqual(len(set(indexes)), 15)
        self.assertTrue(all(5 <= i < 35 for i in indexes))


class TestQuietRoom(unittest.TestCase):
    """REQ-051: a steady bright patch in a dim room is not a bulb."""

    FPS = 30
    W, H = 160, 120
    BULBS = [(40, 60), (80, 60), (40, 100)]
    RED, BLUE, GREEN = (0, 0, 255), (255, 0, 0), (0, 255, 0)

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def _frame(self, spots=(), colour=(255, 255, 255), lift=0):
        f = np.zeros((self.H, self.W, 3), dtype=np.uint8)
        f[0:24, 110:150] = 180
        for cx, cy in spots:
            cv2.circle(f, (cx, cy), 6, colour, -1)
        if lift:
            f = np.clip(f.astype(np.int16) + lift, 0, 255).astype(np.uint8)
        return f

    def _cue(self):
        frames = []
        for i, colour in enumerate((self.RED, self.BLUE, self.GREEN)):
            if i:
                frames += [self._frame() for _ in range(6)]
            frames += [self._frame(self.BULBS, colour) for _ in range(6)]
        return frames

    def _sweep(self, lift=0):
        frames = []
        for i, bulb in enumerate(self.BULBS):
            frames += [self._frame([bulb], lift=lift) for _ in range(15)]
            if i != len(self.BULBS) - 1:
                frames += [self._frame(lift=lift) for _ in range(6)]
        return frames

    def _write(self, path, frames):
        out = cv2.VideoWriter(
            str(path), cv2.VideoWriter_fourcc(*"XVID"), self.FPS, (self.W, self.H)
        )
        for f in frames:
            out.write(f)
        out.release()

    def _assert_on_bulbs(self, blinks):
        self.assertEqual(len(blinks), len(self.BULBS), blinks)
        for _t, cx, cy in blinks:
            bx, by = min(
                self.BULBS, key=lambda p: (p[0] - cx) ** 2 + (p[1] - cy) ** 2
            )
            self.assertLess(
                math.hypot(cx - bx, cy - by), math.hypot(cx - 130, cy - 12)
            )

    def test_window_present_before_opening_is_not_a_bulb(self):
        frames = (
            [self._frame() for _ in range(15)]
            + self._cue()
            + [self._frame() for _ in range(6)]
            + self._sweep()
            + self._cue()
            + [self._frame() for _ in range(15)]
        )
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "window.avi")
            self._write(path, frames)
            rejections: list[float] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 500, rejections)
        self.assertEqual(len(cues), 2, cues)
        self._assert_on_bulbs(blinks)
        self.assertFalse(any(d >= 1.5 for d in rejections), rejections)

    def test_frame_wide_lift_during_the_sweep_is_not_a_bulb(self):
        frames = (
            [self._frame() for _ in range(15)]
            + self._cue()
            + [self._frame() for _ in range(6)]
            + self._sweep(lift=40)
            + self._cue()
            + [self._frame() for _ in range(15)]
        )
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "lift.avi")
            self._write(path, frames)
            rejections: list[float] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 500, rejections)
        self.assertEqual(len(cues), 2, cues)
        self._assert_on_bulbs(blinks)
        self.assertFalse(any(d >= 1.5 for d in rejections), rejections)

    def test_closing_only_uses_the_tail_with_a_window(self):
        frames = self._sweep() + self._cue() + [self._frame() for _ in range(15)]
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "closing.avi")
            self._write(path, frames)
            blinks, cues, *_ = self.m.analyse_feed(path, 500, [])
        self.assertEqual(len(cues), 1, cues)
        self._assert_on_bulbs(blinks)
        self.assertTrue(all(t < cues[0][0] for t, _cx, _cy in blinks))

    def test_short_quiet_room_sets_the_reason_and_keeps_the_bookends(self):
        frames = (
            [self._frame() for _ in range(3)]
            + self._cue()
            + self._sweep()
            + self._cue()
            + [self._frame() for _ in range(3)]
        )
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "short.avi")
            self._write(path, frames)
            problems: list[str] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 500, None, problems)
            found = self.m.find_cues(path)
        self.assertEqual(blinks, [])
        self.assertEqual(problems, [self.m.QUIET_ROOM_REASON])
        self.assertEqual(len(cues), 2)
        self.assertEqual(len(found), 2)

    def test_rebuilt_short_quiet_room_drops_bulbs_without_a_problems_list(self):
        # A lamp on the opening red spot for most of the quiet frames hides that
        # red pulse from the room picture, so the rebuild finds only the closing
        # bookend, whose 3-frame tail is too short.
        lamp = (120, 90)
        lit = self._frame()
        cv2.circle(lit, lamp, 12, (255, 255, 255), -1)
        frames = [lit.copy() for _ in range(6)]
        frames += [self._frame() for _ in range(4)]
        for i, (spot, colour) in enumerate(
            ((lamp, self.RED), ((140, 90), self.BLUE), ((120, 110), self.GREEN))
        ):
            if i:
                frames += [self._frame() for _ in range(6)]
            frames += [self._frame([spot], colour) for _ in range(6)]
        frames += (
            [self._frame() for _ in range(6)]
            + self._sweep()
            + self._cue()
            + [self._frame() for _ in range(3)]
        )
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "rebuild.avi")
            self._write(path, frames)
            meta, colours, _tail = self.m._classify_against_head(path)
            first_cues = self.m._cue_frames(colours, meta["fps"])
            blinks, cues, *_ = self.m.analyse_feed(path, 500)
            problems: list[str] = []
            listed, _cues, *_ = self.m.analyse_feed(path, 500, None, problems)
        self.assertEqual(len(first_cues), 2, first_cues)
        self.assertEqual(len(cues), 1, cues)
        self.assertEqual(blinks, [])
        self.assertEqual(listed, [])
        self.assertEqual(problems, [self.m.QUIET_ROOM_REASON])

    def _short_clip(self, path: str) -> None:
        frames = (
            [self._frame() for _ in range(3)]
            + self._cue()
            + self._sweep()
            + self._cue()
            + [self._frame() for _ in range(3)]
        )
        self._write(path, frames)

    def test_two_short_clips_do_not_claim_the_flashes_were_missed(self):
        with tempfile.TemporaryDirectory() as d:
            a = str(Path(d) / "a.avi")
            b = str(Path(d) / "b.avi")
            self._short_clip(a)
            self._short_clip(b)
            spec = {
                "feeds": [{"path": a, "name": "a.avi"}, {"path": b, "name": "b.avi"}],
                "dwell_ms": 500,
            }
            res, code = _reconstruct_with_exit(spec)
        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertTrue(
            res["error"].startswith("Fewer than two clips could be used.")
        )
        self.assertNotIn("missed the start and end flashes", res["error"])
        self.assertEqual(
            res["rejected_feeds"],
            [
                {"file": "a.avi", "reason": self.m.QUIET_ROOM_REASON},
                {"file": "b.avi", "reason": self.m.QUIET_ROOM_REASON},
            ],
        )

    def test_bright_clips_keep_the_stays_bright_hint_beside_a_short_clip(self):
        with tempfile.TemporaryDirectory() as d:
            dark = np.zeros((self.H, self.W, 3), dtype=np.uint8)
            bright = np.full((self.H, self.W, 3), 200, dtype=np.uint8)
            feeds = []
            for i in range(2):
                path = str(Path(d) / f"bright_{i}.avi")
                self._write(path, [dark] + [bright] * 149)
                feeds.append({"path": path, "name": f"bright_{i}.avi"})
            short = str(Path(d) / "short.avi")
            self._short_clip(short)
            feeds.append({"path": short, "name": "short.avi"})
            res, code = _reconstruct_with_exit({"feeds": feeds, "dwell_ms": 1000})
        self.assertEqual(code, 1)
        self.assertEqual(res["status"], "failed")
        self.assertIn("stays bright", res["error"].lower())
        self.assertEqual(
            res["rejected_feeds"],
            [{"file": "short.avi", "reason": self.m.QUIET_ROOM_REASON}],
        )

    def test_good_feeds_still_succeed_when_one_clip_is_short(self):
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=4, seed=7)
            short = str(Path(d) / "short.avi")
            self._short_clip(short)
            spec = _spec(d, gt)
            spec["feeds"].append({"path": short, "name": "short.avi"})
            res = _reconstruct(spec)
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(
            res["rejected_feeds"],
            [{"file": "short.avi", "reason": self.m.QUIET_ROOM_REASON}],
        )


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    unittest.main(verbosity=2)
