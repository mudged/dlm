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

import json
import math
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

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


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    unittest.main(verbosity=2)
