# Wire spacing gap fill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fill a scanned hole of up to three lights from the string's own spacing, move a low-confidence light onto that line, and name both on the review screen before save.

**Architecture:** One pure function in the CV child runs after unseen ids are marked missing. It does not change blink detection, bookends, pose, or triangulation. Go forwards `max_gap_cm` as `max_gap_m` and parses two new id lists. The review screen shows those lists. Confirm still refuses a non-empty `missing` list.

**Tech Stack:** Python 3 (`unittest`, no new packages), Go 1.25, Next.js / TypeScript (Vitest).

## Global Constraints

- Confident triangulated lights are the only anchors. Their coordinates stay as scanned. Guesses do not become anchors.
- `max_step` is the longest Euclidean distance among anchors whose ids differ by 1 and whose distance is greater than 0. One such pair is enough. With none, `max_step` is `max_gap_m` (default `0.10`).
- An interior run of 1, 2, or 3 ids is placed evenly on the straight line between its anchors, even when that spacing is shorter or longer than `max_step`. A run of 4 or more is left whole.
- An end run of 1, 2, or 3 repeats the neighbour step's direction. Each hop is `max_step` long. The neighbour must be an anchor and that step's length must be greater than 0.
- Fallback runs only when there is no usable pair. A low-confidence light within 3 ids of an anchor, at a distance greater than 0, moves to `id distance × max_gap_m` along the scanned direction. Missing lights are not invented from one anchor.
- With no anchors, `guessed` and `adjusted` stay empty.
- Form field `max_gap_cm`, prefilled with `10`. Blank or omitted means 10 cm. **400** message, exact: `maximum gap must be a positive number of centimetres up to 100`. Child message, exact: `max_gap_m must be a positive finite number`.
- Result keys `guessed` and `adjusted` are sorted arrays, `[]` when empty, never JSON `null`. Each id is in at most one of `missing`, `low_confidence`, `guessed`, and `adjusted`.
- Confirm still refuses a non-empty `missing` list. The saved model stores `id, x, y, z` only.
- New requirement **REQ-053**. Do not renumber or reuse a `REQ-*` code or a `§N.N` section. Add §3.23.4.
- Do not fit a curve, do not add per-light accept/reject, and do not colour guessed lights in the preview.
- Tests compare coordinates within `1e-6` m. Python tests call `apply_wire_spacing` with positions and do not decode video, except the three existing pipeline tests this plan updates.

```bash
cd /workspaces/dlm/backend/internal/cvruntime/src
PY=python3
python3 -c "import cv2" 2>/dev/null || PY=/workspaces/dlm/dist/cvruntime/linux_amd64/python/bin/python
```

Every later Python command assumes that `PY` and that working directory.

---

### Task 1: Spacing function

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (new function before Stage 5)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py` (new `TestWireSpacing`)

**Interfaces:**
- Consumes: triangulated points, `missing`, `low_confidence`, `span`, `max_gap_m`
- Produces: `apply_wire_spacing(lights_3d, missing, low_conf, span, max_gap_m) -> tuple[dict[int, tuple[float, float, float]], list[int], list[int], list[int], list[int]]` which is `(points, missing, low_confidence, guessed, adjusted)`. Lists are sorted. The input dict is not mutated.

- [ ] **Step 1: Write the failing tests**

Append this class at the end of `test_reconstruct.py`:

```python
class TestWireSpacing(unittest.TestCase):
    """REQ-053: short holes are filled from the wire spacing."""

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def _near(self, got, want):
        self.assertEqual(len(got), 3)
        for a, b in zip(got, want):
            self.assertAlmostEqual(a, b, delta=1e-6)

    def _apply(self, points, missing, low, span, gap=0.10):
        return self.m.apply_wire_spacing(
            dict(points), list(missing), list(low), span, gap
        )

    def test_end_uses_longest_confident_step(self):
        points, missing, low, guessed, adjusted = self._apply(
            {2: (0.0, 0.0, 0.0), 3: (0.08, 0.0, 0.0), 4: (0.18, 0.0, 0.0)},
            [0, 1], [], 5,
        )
        self.assertEqual(guessed, [0, 1])
        self.assertEqual(missing, [])
        self.assertEqual(adjusted, [])
        self.assertEqual(low, [])
        self._near(points[0], (-0.20, 0.0, 0.0))
        self._near(points[1], (-0.10, 0.0, 0.0))
        self._near(points[2], (0.0, 0.0, 0.0))
        self._near(points[3], (0.08, 0.0, 0.0))
        self._near(points[4], (0.18, 0.0, 0.0))

    def test_suffix_uses_longest_step_not_the_short_neighbour(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 1: (0.10, 0.0, 0.0), 2: (0.18, 0.0, 0.0)},
            [3], [], 4,
        )
        self.assertEqual(guessed, [3])
        self.assertEqual(missing, [])
        self.assertEqual(adjusted, [])
        self._near(points[3], (0.28, 0.0, 0.0))

    def test_interior_low_confidence_moves_to_the_midpoint(self):
        points, missing, low, guessed, adjusted = self._apply(
            {3: (0.08, 0.0, 0.0), 4: (9.0, 9.0, 9.0), 5: (0.24, 0.0, 0.0)},
            [], [4], 6,
        )
        self.assertEqual(adjusted, [4])
        self.assertEqual(low, [])
        self.assertEqual(guessed, [])
        self.assertEqual(missing, [])
        self._near(points[4], (0.16, 0.0, 0.0))
        self._near(points[3], (0.08, 0.0, 0.0))

    def test_three_missing_ids_are_evenly_spaced(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 4: (0.40, 0.0, 0.0)},
            [1, 2, 3], [], 5,
        )
        self.assertEqual(guessed, [1, 2, 3])
        self.assertEqual(missing, [])
        self._near(points[1], (0.10, 0.0, 0.0))
        self._near(points[2], (0.20, 0.0, 0.0))
        self._near(points[3], (0.30, 0.0, 0.0))

    def test_interior_fills_even_when_the_step_is_longer_than_the_max(self):
        points, missing, low, guessed, adjusted = self._apply(
            {
                0: (0.0, 0.0, 0.0),
                2: (1.0, 0.0, 0.0),
                4: (2.0, 0.0, 0.0),
                5: (2.08, 0.0, 0.0),
            },
            [1, 3], [], 6,
        )
        self.assertEqual(guessed, [1, 3])
        self._near(points[1], (0.50, 0.0, 0.0))
        self._near(points[3], (1.50, 0.0, 0.0))

    def test_four_missing_ids_stay_missing(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 5: (0.50, 0.0, 0.0)},
            [1, 2, 3, 4], [], 6,
        )
        self.assertEqual(missing, [1, 2, 3, 4])
        self.assertEqual(guessed, [])
        self.assertNotIn(1, points)

    def test_low_confidence_inside_a_run_of_four_stays(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 2: (1.0, 1.0, 1.0), 5: (0.50, 0.0, 0.0)},
            [1, 3, 4], [2], 6,
        )
        self.assertEqual(low, [2])
        self.assertEqual(adjusted, [])
        self.assertEqual(guessed, [])
        self.assertEqual(missing, [1, 3, 4])
        self._near(points[2], (1.0, 1.0, 1.0))

    def test_end_without_a_confident_neighbour_stays_missing(self):
        points, missing, low, guessed, adjusted = self._apply(
            {2: (0.0, 0.0, 0.0)},
            [0, 1, 3], [], 4,
        )
        self.assertEqual(missing, [0, 1, 3])
        self.assertEqual(guessed, [])

    def test_zero_length_neighbour_does_not_grow_the_end(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 1: (0.0, 0.0, 0.0)},
            [2], [], 3,
        )
        self.assertEqual(missing, [2])
        self.assertEqual(guessed, [])
        self.assertNotIn(2, points)

    def test_fallback_slides_a_low_confidence_light_and_leaves_the_hole(self):
        points, missing, low, guessed, adjusted = self._apply(
            {2: (0.0, 0.0, 0.0), 4: (0.0, 0.50, 0.0)},
            [3], [4], 5, gap=0.10,
        )
        self.assertEqual(adjusted, [4])
        self.assertEqual(low, [])
        self.assertEqual(missing, [3])
        self.assertEqual(guessed, [])
        self._near(points[4], (0.0, 0.20, 0.0))

    def test_low_confidence_on_the_anchor_stays(self):
        points, missing, low, guessed, adjusted = self._apply(
            {0: (0.0, 0.0, 0.0), 1: (0.0, 0.0, 0.0)},
            [], [1], 2,
        )
        self.assertEqual(low, [1])
        self.assertEqual(adjusted, [])
        self._near(points[1], (0.0, 0.0, 0.0))

    def test_no_anchors_leaves_the_lists_empty(self):
        points, missing, low, guessed, adjusted = self._apply(
            {}, [0, 1], [], 2,
        )
        self.assertEqual(points, {})
        self.assertEqual(missing, [0, 1])
        self.assertEqual(guessed, [])
        self.assertEqual(adjusted, [])

    def test_input_dict_is_not_mutated(self):
        src = {2: (0.0, 0.0, 0.0), 3: (0.08, 0.0, 0.0)}
        self._apply(src, [0, 1], [], 4)
        self.assertEqual(src, {2: (0.0, 0.0, 0.0), 3: (0.08, 0.0, 0.0)})
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `"$PY" test_reconstruct.py TestWireSpacing -v`

Expected: FAIL with `AttributeError` (`apply_wire_spacing` is missing).

- [ ] **Step 3: Implement `apply_wire_spacing`**

Insert this function in `reconstruct.py` immediately above the `# Stage 5 — Emit Result` comment:

```python
def apply_wire_spacing(
    lights_3d: dict[int, tuple[float, float, float]],
    missing: list[int],
    low_conf: list[int],
    span: int,
    max_gap_m: float,
) -> tuple[
    dict[int, tuple[float, float, float]],
    list[int],
    list[int],
    list[int],
    list[int],
]:
    """Fill holes of 1..3 lights and move low-confidence lights (REQ-053).

    Anchors are triangulated ids that are not low-confidence. Their
    coordinates are copied through unchanged. Returns new lists; the
    caller's dict is not mutated.
    """
    points = dict(lights_3d)
    missing_set = set(missing)
    low_set = set(low_conf)
    guessed: list[int] = []
    adjusted: list[int] = []

    anchors = sorted(lid for lid in points if lid not in low_set)
    anchor_set = set(anchors)

    def _dist(a: int, b: int) -> float:
        ax, ay, az = points[a]
        bx, by, bz = points[b]
        return math.dist((ax, ay, az), (bx, by, bz))

    usable: list[float] = []
    for lid in anchors:
        if lid + 1 in anchor_set:
            d = _dist(lid, lid + 1)
            if d > 0:
                usable.append(d)
    max_step = max(usable) if usable else max_gap_m

    def _place(lid: int, xyz: tuple[float, float, float]) -> None:
        if lid in low_set:
            points[lid] = xyz
            low_set.discard(lid)
            adjusted.append(lid)
        elif lid in missing_set:
            points[lid] = xyz
            missing_set.discard(lid)
            guessed.append(lid)

    if anchors:
        for i in range(len(anchors) - 1):
            prev, nxt = anchors[i], anchors[i + 1]
            k = nxt - prev - 1
            if not 1 <= k <= 3:
                continue
            px, py, pz = points[prev]
            nx, ny, nz = points[nxt]
            for j in range(1, k + 1):
                t = j / (k + 1)
                _place(prev + j, (
                    px + t * (nx - px),
                    py + t * (ny - py),
                    pz + t * (nz - pz),
                ))

        s = anchors[0]
        if 1 <= s <= 3 and (s + 1) in anchor_set:
            vx = points[s + 1][0] - points[s][0]
            vy = points[s + 1][1] - points[s][1]
            vz = points[s + 1][2] - points[s][2]
            length = math.sqrt(vx * vx + vy * vy + vz * vz)
            if length > 0:
                scale = max_step / length
                for j in range(1, s + 1):
                    _place(s - j, (
                        points[s][0] - j * scale * vx,
                        points[s][1] - j * scale * vy,
                        points[s][2] - j * scale * vz,
                    ))

        t_id = anchors[-1]
        k_suf = span - 1 - t_id
        if 1 <= k_suf <= 3 and (t_id - 1) in anchor_set:
            vx = points[t_id][0] - points[t_id - 1][0]
            vy = points[t_id][1] - points[t_id - 1][1]
            vz = points[t_id][2] - points[t_id - 1][2]
            length = math.sqrt(vx * vx + vy * vy + vz * vz)
            if length > 0:
                scale = max_step / length
                for j in range(1, k_suf + 1):
                    _place(t_id + j, (
                        points[t_id][0] + j * scale * vx,
                        points[t_id][1] + j * scale * vy,
                        points[t_id][2] + j * scale * vz,
                    ))

    if not usable and anchors:
        for lid in sorted(low_set):
            # A low-confidence light inside a run of 4 or more stays put.
            # That run was left unchanged in the interior phase.
            prevs = [a for a in anchors if a < lid]
            nexts = [a for a in anchors if a > lid]
            if prevs and nexts and (min(nexts) - max(prevs) - 1) >= 4:
                continue
            best = min(anchors, key=lambda a: (abs(a - lid), a))
            gap = abs(best - lid)
            if not 1 <= gap <= 3:
                continue
            ax, ay, az = points[best]
            lx, ly, lz = points[lid]
            vx, vy, vz = lx - ax, ly - ay, lz - az
            length = math.sqrt(vx * vx + vy * vy + vz * vz)
            if length <= 0:
                continue
            scale = (gap * max_gap_m) / length
            points[lid] = (ax + scale * vx, ay + scale * vy, az + scale * vz)
            low_set.discard(lid)
            adjusted.append(lid)

    return (
        points,
        sorted(missing_set),
        sorted(low_set),
        sorted(guessed),
        sorted(adjusted),
    )
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `"$PY" test_reconstruct.py TestWireSpacing -v`

Expected: PASS (12 tests).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
feat(capture): fill short light holes from wire spacing

EOF
)"
```

---

### Task 2: Call the spacing step from the scan

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_make_result`, `_make_error`, `main`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: `apply_wire_spacing` from Task 1. `span` is the value already computed in `main` (`light_count` when set, otherwise `max(all_ids) + 1`).
- Produces: success and failure JSON always include `guessed` and `adjusted`. `main` reads `max_gap_m` and fails with `max_gap_m must be a positive finite number` when the key is present and not a positive finite number. An absent key uses `0.10`.

- [ ] **Step 1: Write the failing parser and schema tests**

Add these methods to `TestWireSpacing`:

```python
    def test_absent_max_gap_defaults(self):
        self.assertEqual(self.m._parse_max_gap_m(None), (0.10, None))

    def test_present_max_gap_is_kept(self):
        self.assertEqual(self.m._parse_max_gap_m(0.25), (0.25, None))

    def test_invalid_max_gap_message(self):
        for bad in (0, -0.1, float("nan"), float("inf"), "10", True):
            value, err = self.m._parse_max_gap_m(bad)
            self.assertEqual(err, "max_gap_m must be a positive finite number", bad)
            self.assertEqual(value, 0.10)

    def test_make_error_includes_empty_guess_lists(self):
        err = self.m._make_error("boom")
        self.assertEqual(err["guessed"], [])
        self.assertEqual(err["adjusted"], [])

    def test_make_result_includes_guess_lists(self):
        result = self.m._make_result(
            {0: (0.0, 0.0, 0.0)}, [], [], {0}, guessed=[0], adjusted=[]
        )
        self.assertEqual(result["guessed"], [0])
        self.assertEqual(result["adjusted"], [])
```

In `test_result_schema`, extend the `required` tuple to:

```python
        required = ("status", "light_count", "lights", "missing",
                    "low_confidence", "guessed", "adjusted", "error")
```

and after the `low_confidence` type assert add:

```python
        self.assertIsInstance(res["guessed"], list)
        self.assertIsInstance(res["adjusted"], list)
```

In `test_occluded_light_in_missing_not_lights`, replace the two assertions that the occluded light has no coordinates and is in `missing` with:

```python
        self.assertIn(
            occlude_light, recovered_ids,
            f"Light {occlude_light} between confident neighbours must be guessed",
        )
        self.assertIn(occlude_light, res["guessed"])
        self.assertNotIn(occlude_light, res["missing"])
```

In `test_dropped_blink_does_not_mislabel_other_lights`, replace the two assertions that the occluded light must not get coordinates and must be in `missing` with:

```python
        self.assertIn(occlude_light, recovered_ids)
        self.assertIn(occlude_light, res["guessed"])
        self.assertNotIn(occlude_light, res["missing"])
```

Leave the loop that checks lights after the occluded id against ground truth as it is. It already skips the occluded id.

In `test_success_reports_rejected_feeds_and_light_count`, replace the two `assertIn` calls for ids 6 and 7 with:

```python
        self.assertIn(6, res["guessed"])
        self.assertIn(7, res["guessed"])
        self.assertNotIn(6, res["missing"])
        self.assertNotIn(7, res["missing"])
```

Leave `light_count == 8` and the empty `rejected_feeds` assert.

In `test_no_fabrication_guarantee`, add after the missing-disjoint assert:

```python
        self.assertEqual(set(res["guessed"]) & miss, set())
        self.assertEqual(set(res["adjusted"]) & set(res["low_confidence"]), set())
        self.assertTrue(set(res["guessed"]) <= recovered)
        self.assertTrue(set(res["adjusted"]) <= recovered)
```

- [ ] **Step 2: Run the new unit tests to verify they fail**

Run: `"$PY" test_reconstruct.py TestWireSpacing.test_absent_max_gap_defaults TestWireSpacing.test_make_error_includes_empty_guess_lists -v`

Expected: FAIL with `AttributeError` (`_parse_max_gap_m` is missing) or a missing `guessed` key.

- [ ] **Step 3: Parse the gap, emit the lists, and call the function**

Add next to `LIGHT_COUNT_ERROR`:

```python
MAX_GAP_ERROR = "max_gap_m must be a positive finite number"
```

Add next to `_parse_light_count`:

```python
def _parse_max_gap_m(value) -> tuple[float, Optional[str]]:
    """Returns ``(max_gap_m, error)``. An absent value is 0.10 m."""
    if value is None:
        return 0.10, None
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        return 0.10, MAX_GAP_ERROR
    v = float(value)
    if not math.isfinite(v) or v <= 0:
        return 0.10, MAX_GAP_ERROR
    return v, None
```

In `_make_result`, add parameters `guessed: Optional[list[int]] = None` and `adjusted: Optional[list[int]] = None` after `light_count`, and add these keys to the returned dict:

```python
        "guessed":        sorted(guessed or []),
        "adjusted":       sorted(adjusted or []),
```

In `_make_error`, add `"guessed": []` and `"adjusted": []`.

In `main`, after the `light_count` parse:

```python
    max_gap_m, gap_err = _parse_max_gap_m(spec.get("max_gap_m"))
    if gap_err:
        _emit_failure(gap_err)
```

After the span loop and before `_make_result`:

```python
        lights_3d, missing, low_conf, guessed, adjusted = apply_wire_spacing(
            lights_3d, missing, low_conf, span, max_gap_m
        )
```

Pass `guessed=guessed, adjusted=adjusted` into `_make_result`. Extend the done log with the guessed and adjusted counts.

- [ ] **Step 4: Run the spacing tests and the updated pipeline tests**

Run: `"$PY" test_reconstruct.py TestWireSpacing.test_absent_max_gap_defaults TestWireSpacing.test_present_max_gap_is_kept TestWireSpacing.test_invalid_max_gap_message TestWireSpacing.test_make_error_includes_empty_guess_lists TestWireSpacing.test_make_result_includes_guess_lists TestReconstructSynthetic.test_occluded_light_in_missing_not_lights TestReconstructSynthetic.test_dropped_blink_does_not_mislabel_other_lights TestReconstructSynthetic.test_result_schema TestReconstructSynthetic.test_no_fabrication_guarantee TestBookends.test_success_reports_rejected_feeds_and_light_count -v`

Expected: PASS. If `test_success_reports_rejected_feeds_and_light_count` leaves 6 and 7 in `missing`, the fixture's last two lights are not a confident adjacent pair. In that case keep them in `missing` and assert `guessed` does not contain them. Do not weaken the unit tests in `TestWireSpacing`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
feat(capture): apply wire spacing before the scan result

EOF
)"
```

---

### Task 3: Forward the gap and the new lists through Go

**Files:**
- Modify: `backend/internal/cvruntime/contract.go`
- Modify: `backend/internal/cvruntime/run.go`
- Modify: `backend/internal/cvruntime/cvruntime_test.go`
- Modify: `backend/internal/reconstruct/reconstruct.go`
- Modify: `backend/internal/reconstruct/reconstruct_test.go`
- Modify: `backend/internal/httpapi/capture_models.go`
- Modify: `backend/internal/httpapi/capture_models_test.go`

**Interfaces:**
- Consumes: the child JSON keys `guessed`, `adjusted`, and `max_gap_m`
- Produces: `CreateParams.MaxGapM float64`, `JobSpec.MaxGapM float64 \`json:"max_gap_m,omitempty"\``, `Result.Guessed` and `Result.Adjusted` as `[]int \`json:"guessed"\`` and `[]int \`json:"adjusted"\``. `(*Result).normalizeLists()` replaces nil `Lights`, `Missing`, `LowConfidence`, `Guessed`, `Adjusted`, and `RejectedFeeds` with empty slices. `Run` calls it before every successful return of a parsed result, including a `status:"failed"` payload.

- [ ] **Step 1: Write the failing Go tests**

In `cvruntime_test.go` `TestRun_StubInterpreter`, after the light-id check, add:

```go
	if result.Guessed == nil || result.Adjusted == nil {
		t.Fatalf("Guessed=%v Adjusted=%v, want empty slices", result.Guessed, result.Adjusted)
	}
```

The canned JSON in that test stays without those keys, so this fails until `normalizeLists` runs.

In `capture_models_test.go`, add:

```go
func TestAPIv1CaptureModels_postMaxGap_forwardsMetres(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)
	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	}, "max_gap_cm", "100")
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	if math.Abs(fake.last.MaxGapM-1.0) > 1e-9 {
		t.Fatalf("MaxGapM = %v, want 1", fake.last.MaxGapM)
	}
}

func TestAPIv1CaptureModels_postWithoutMaxGap_defaultsToTenCm(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)
	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	})
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	if math.Abs(fake.last.MaxGapM-0.10) > 1e-9 {
		t.Fatalf("MaxGapM = %v, want 0.10", fake.last.MaxGapM)
	}
}

func TestAPIv1CaptureModels_postWithInvalidMaxGap_returns400(t *testing.T) {
	const wantMsg = "maximum gap must be a positive number of centimetres up to 100"
	for _, val := range []string{"0", "-1", "101", "nope"} {
		t.Run(val, func(t *testing.T) {
			srv, _ := newCaptureModelsTestServer(t)
			body, ct := multipartBody(t, map[string]string{
				"feed_a.mp4": "fake-video-a",
				"feed_b.mp4": "fake-video-b",
			}, "max_gap_cm", val)
			res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusBadRequest {
				b, _ := io.ReadAll(res.Body)
				t.Fatalf("status = %d, want 400, body = %s", res.StatusCode, b)
			}
			var env struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if env.Error.Message != wantMsg {
				t.Fatalf("message = %q, want %q", env.Error.Message, wantMsg)
			}
		})
	}
}
```

Add `"math"` to that test file's imports if it is not already imported.

In `reconstruct_test.go`, add:

```go
func TestManager_confirmGuessedLights_persists(t *testing.T) {
	st := newTestStore(t)
	filled := cvruntime.Result{
		Status:     "succeeded",
		LightCount: 3,
		Lights: []cvruntime.LightPoint{
			{ID: 0, X: 0}, {ID: 1, X: 0.1}, {ID: 2, X: 0.2},
		},
		Missing: []int{},
		Guessed: []int{1},
	}
	m := reconstruct.New(&fakeRunner{result: filled}, st, t.TempDir())
	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitFor(t, m, id, reconstruct.StatusSucceeded)
	if _, err := m.Confirm(context.Background(), id, "guessed-model"); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	all, err := st.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range all {
		if s.Name == "guessed-model" {
			found = true
		}
	}
	if !found {
		t.Fatal("guessed model was not persisted")
	}
}

func TestManager_confirmGuessedAndMissing_stillNamesTheHole(t *testing.T) {
	st := newTestStore(t)
	partial := cvruntime.Result{
		Status:     "succeeded",
		LightCount: 4,
		Lights: []cvruntime.LightPoint{
			{ID: 0, X: 0}, {ID: 1, X: 0.1}, {ID: 3, X: 0.3},
		},
		Missing: []int{2},
		Guessed: []int{1},
	}
	m := reconstruct.New(&fakeRunner{result: partial}, st, t.TempDir())
	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitFor(t, m, id, reconstruct.StatusSucceeded)
	_, err = m.Confirm(context.Background(), id, "partial-model")
	var pe *wiremodel.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wiremodel.ParseError, got %v", err)
	}
	if !strings.Contains(pe.Message, "2") {
		t.Fatalf("error message should name missing light id 2, got: %q", pe.Message)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /workspaces/dlm/backend && go test ./internal/cvruntime/ ./internal/httpapi/ ./internal/reconstruct/ -count=1 -run 'TestRun_StubInterpreter|TestAPIv1CaptureModels_postMaxGap_forwardsMetres|TestAPIv1CaptureModels_postWithoutMaxGap_defaultsToTenCm|TestAPIv1CaptureModels_postWithInvalidMaxGap_returns400|TestManager_confirmGuessedLights_persists|TestManager_confirmGuessedAndMissing_stillNamesTheHole'`

Expected: FAIL to compile (`MaxGapM`, `Guessed`, or `Adjusted` is unknown).

- [ ] **Step 3: Implement the Go fields**

In `contract.go` `JobSpec`, add:

```go
	// MaxGapM is the fallback step in metres when the scan has no confident
	// adjacent pair (REQ-053). Zero means the child uses 0.10.
	MaxGapM float64 `json:"max_gap_m,omitempty"`
```

In `Result`, add:

```go
	Guessed       []int          `json:"guessed"`
	Adjusted      []int          `json:"adjusted"`
```

Add:

```go
// normalizeLists makes JSON omit no id list. A nil slice would marshal as null.
func (r *Result) normalizeLists() {
	if r.Lights == nil {
		r.Lights = []LightPoint{}
	}
	if r.Missing == nil {
		r.Missing = []int{}
	}
	if r.LowConfidence == nil {
		r.LowConfidence = []int{}
	}
	if r.Guessed == nil {
		r.Guessed = []int{}
	}
	if r.Adjusted == nil {
		r.Adjusted = []int{}
	}
	if r.RejectedFeeds == nil {
		r.RejectedFeeds = []RejectedFeed{}
	}
}
```

In `run.go`, call `result.normalizeLists()` immediately after each `json.Unmarshal` that returns a result (the failed-status path and the success path).

In `reconstruct.CreateParams`, add `MaxGapM float64`. In the `cvruntime.JobSpec` literal inside `Create`, set `MaxGapM: params.MaxGapM`.

In `capture_models.go`, after the `light_count` parse:

```go
	maxGapM := 0.10
	if raw := strings.TrimSpace(r.FormValue("max_gap_cm")); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 100 {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "maximum gap must be a positive number of centimetres up to 100")
			return
		}
		maxGapM = v / 100
	}
	params.MaxGapM = maxGapM
```

`math` is already imported in that file.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd /workspaces/dlm/backend && go test ./internal/cvruntime/ ./internal/httpapi/ ./internal/reconstruct/ -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/contract.go backend/internal/cvruntime/run.go backend/internal/cvruntime/cvruntime_test.go backend/internal/reconstruct/reconstruct.go backend/internal/reconstruct/reconstruct_test.go backend/internal/httpapi/capture_models.go backend/internal/httpapi/capture_models_test.go
git commit -m "$(cat <<'EOF'
feat(capture): accept a maximum gap and return guessed lights

EOF
)"
```

---

### Task 4: Review screen

**Files:**
- Modify: `web/lib/models.ts`
- Modify: `web/lib/models.test.ts`
- Modify: `web/app/models/new/NewModelClient.tsx`

**Interfaces:**
- Consumes: `max_gap_cm` on `POST /api/v1/models/capture`, and `guessed` / `adjusted` on `CaptureJobResult`
- Produces: `parseMaxGapCm(raw: string): number | undefined` and `maxGapCmMessage`. Blank returns `undefined`. Any other value that is not a finite number greater than 0 and at most 100 throws `Error` with `maximum gap must be a positive number of centimetres up to 100`. `createCaptureJob` sets form field `max_gap_cm` only when `params.max_gap_cm` is a number.

- [ ] **Step 1: Write the failing Vitest tests**

In `web/lib/models.test.ts`, import `parseMaxGapCm` and add:

```typescript
describe("parseMaxGapCm (REQ-053)", () => {
  it("returns undefined for a blank box", () => {
    expect(parseMaxGapCm("")).toBeUndefined();
    expect(parseMaxGapCm("  ")).toBeUndefined();
  });

  it("parses 10 and 100", () => {
    expect(parseMaxGapCm("10")).toBe(10);
    expect(parseMaxGapCm("100")).toBe(100);
  });

  it("rejects zero, above 100, and words", () => {
    for (const raw of ["0", "-1", "101", "nope"]) {
      expect(() => parseMaxGapCm(raw)).toThrow(
        "maximum gap must be a positive number of centimetres up to 100",
      );
    }
  });
});
```

Next to the existing `light_count` form tests, add:

```typescript
  it("sets max_gap_cm on the form when the caller passes it", async () => {
    let captured: FormData | null = null;
    globalThis.fetch = vi.fn(async (_url, init) => {
      captured = init?.body as FormData;
      return new Response(JSON.stringify(MOCK_JOB), {
        status: 202,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const f1 = new File(["a"], "a.mp4", { type: "video/mp4" });
    const f2 = new File(["b"], "b.mp4", { type: "video/mp4" });
    await createCaptureJob([f1, f2], { max_gap_cm: 10 });

    expect((captured as FormData).get("max_gap_cm")).toBe("10");
  });

  it("omits max_gap_cm when the caller does not pass it", async () => {
    let captured: FormData | null = null;
    globalThis.fetch = vi.fn(async (_url, init) => {
      captured = init?.body as FormData;
      return new Response(JSON.stringify(MOCK_JOB), {
        status: 202,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const f1 = new File(["a"], "a.mp4", { type: "video/mp4" });
    const f2 = new File(["b"], "b.mp4", { type: "video/mp4" });
    await createCaptureJob([f1, f2]);

    expect((captured as FormData).has("max_gap_cm")).toBe(false);
  });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /workspaces/dlm/web && npm test -- lib/models.test.ts`

Expected: FAIL (`parseMaxGapCm` is not exported, or `max_gap_cm` is not on the form).

- [ ] **Step 3: Implement the client and the review copy**

In `web/lib/models.ts`, add `guessed: number[]` and `adjusted: number[]` to `CaptureJobResult`.

Add:

```typescript
export const maxGapCmMessage =
  "maximum gap must be a positive number of centimetres up to 100";

/** Blank means omit max_gap_cm. Otherwise a number of centimetres in (0, 100]. */
export function parseMaxGapCm(raw: string): number | undefined {
  const trimmed = raw.trim();
  if (trimmed === "") return undefined;
  if (!/^(?:\d+\.?\d*|\.\d+)$/.test(trimmed)) {
    throw new Error(maxGapCmMessage);
  }
  const n = Number(trimmed);
  if (!Number.isFinite(n) || n <= 0 || n > 100) {
    throw new Error(maxGapCmMessage);
  }
  return n;
}
```

Extend `createCaptureJob`'s params type with `max_gap_cm?: number`, and after the light-count block:

```typescript
  if (params?.max_gap_cm !== undefined) {
    fd.set("max_gap_cm", String(params.max_gap_cm));
  }
```

In `NewModelClient.tsx`, import `parseMaxGapCm`. In `VideoPanel`, add `const [maxGapCm, setMaxGapCm] = useState("10");`. In `onSubmit`, parse it the same way as light count, and pass `max_gap_cm: parsedMaxGap` to `createCaptureJob`. In `onCancel`, `setMaxGapCm("10")`.

After the light-count field, add:

```tsx
        <div className="flex flex-col gap-1">
          <label htmlFor="capture-max-gap" className="text-sm font-medium">
            Maximum gap (cm)
          </label>
          <p className="text-xs text-slate-500 dark:text-slate-400">
            This number is used only when the scan cannot measure a step from
            lights it is sure about.
          </p>
          <input
            id="capture-max-gap"
            name="max_gap_cm"
            type="text"
            inputMode="decimal"
            autoComplete="off"
            value={maxGapCm}
            onChange={(e) => setMaxGapCm(e.target.value)}
            disabled={phase.kind === "submitting"}
            className="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </div>
```

In the review block, when `result.missing.length > 0`, keep the missing-ids line and add a second line: `Confirm will not save a model while any light is missing.`

After the low-confidence block, when `result.guessed.length > 0`, show `Guessed IDs: …` and `These positions were filled in along the string. They were not seen in the videos.`

When `result.adjusted.length > 0`, show `Adjusted IDs: …` and `These were seen, then moved onto the usual step.`

Show `All lights detected with high confidence.` only when `missing`, `low_confidence`, `guessed`, and `adjusted` are all empty.

The preview keeps using `captureToPreviewLights(result.lights)` with no extra colour.

- [ ] **Step 4: Run the Vitest file**

Run: `cd /workspaces/dlm/web && npm test -- lib/models.test.ts`

Expected: PASS.

- [ ] **Step 5: Check the upload form in the browser**

Open `http://127.0.0.1:8080/models/new?tab=video` (start `./scripts/run.sh` from the repo root first if nothing is listening). Confirm the Maximum gap field is prefilled with `10`, the helper sentence is the one above, and Confirm is still on the page only after a successful job. A guessed or adjusted line cannot be forced without a scan result; the sentences are covered by the component source in Step 3.

- [ ] **Step 6: Commit**

```bash
git add web/lib/models.ts web/lib/models.test.ts web/app/models/new/NewModelClient.tsx
git commit -m "$(cat <<'EOF'
feat(capture): show guessed and adjusted lights before save

EOF
)"
```

---

### Task 5: Write the requirement and the design section

**Files:**
- Modify: `docs/requirements/requirements.md`
- Modify: `docs/requirements/acceptance-criteria.md`
- Modify: `docs/userguide/build-model-from-video.md`
- Modify: `docs/design/backend-lights-and-automation.md`
- Modify: `docs/design/architecture.md`
- Modify: `docs/design/backend-service.md`
- Modify: `docs/design/frontend.md`
- Modify: `docs/design/overview.md`
- Modify: `docs/design/appendix-traceability.md`

**Interfaces:**
- Consumes: the behaviour from Tasks 1–4
- Produces: REQ-053 and §3.23.4. No source-code change.

- [ ] **Step 1: Update the requirements and the user guide**

In `docs/requirements/requirements.md` §11, replace `Lights it couldn't place are honestly reported, never made up.` with:

```markdown
A hole of up to three lights is filled in along the string and named before you save. A longer hole is still reported, and the model is not saved while any light is missing. A light the scan was unsure about can be moved onto the usual spacing and is named as adjusted.
```

Append this index row after REQ-052. Keep the code out of the section prose.

```markdown
| REQ-053 | Short holes in a scanned model are filled and named | §11 Building a model from video |
```

In `docs/requirements/acceptance-criteria.md`, under "Building a model from video", add:

```markdown
**A short hole in the scan is filled in and named.**
- Try: create a model from video where the scan missed three lights in a row between lights it saw.
- You should see: those three ids named as guessed, and Confirm can save the model.

**A long hole still blocks the save.**
- Try: create a model from video where the scan missed four lights in a row.
- You should see: those ids named as missing, a line that Confirm will not save while any light is missing, and the save is refused.
```

In `docs/userguide/build-model-from-video.md`, in section 4, add a sentence that the upload page has a **Maximum gap (cm)** box prefilled with 10, used only when the scan cannot measure a step from lights it is sure about. In the review list, add a guessed line (filled in along the string, not seen in the videos) and an adjusted line (seen, then moved onto the usual step). Do not write a `REQ-*` code.

- [ ] **Step 2: Update the design docs**

In `docs/design/backend-lights-and-automation.md` §3.23, replace the sentence that begins `The pipeline never fabricates coordinates` with:

```markdown
A hole of one, two, or three lights is filled by §3.23.4, and a low-confidence light inside that hole is moved onto the filled line. A longer hole stays in `missing`. The review step (§4.17) shows guessed, adjusted, missing, and low-confidence ids so the user can re-record before confirming.
```

In the result-contract sentence, add `guessed` and `adjusted` to the JSON sketch. In the mermaid diagram, insert a node between triangulation and the result:

```mermaid
  TRI["triangulate index seen in ≥ 2 feeds → x,y,z (metres)"] --> GAP
  GAP["wire spacing: fill a hole of 1–3, move a low-confidence light (§3.23.4)"] --> REP
```

Change `REP` so it reports `guessed` and `adjusted` as well. Leave `TRI`'s other incoming edge as it is.

Add this section after §3.23.3. Do not renumber §3.23 or §3.23.1–§3.23.3.

```markdown
#### 3.23.4 Wire spacing (REQ-053)

**In plain terms:** after the scan has placed the lights it is sure about, a short gap in the string is filled in a straight line so the model can be saved. A light the scan was unsure about is moved onto that same line. A gap of four or more is left missing, and Confirm still refuses the save.

- Anchors are triangulated lights that are not in `low_confidence`. Their coordinates stay as scanned.
- `max_step` is the longest distance between anchors whose ids differ by 1, ignoring a pair whose distance is 0. With no such pair, `max_step` is `max_gap_m` (default 0.10). The upload field is `max_gap_cm`, default 10. A value that is not a finite number greater than 0 and at most 100 is **400** `maximum gap must be a positive number of centimetres up to 100`.
- A run of 1, 2, or 3 ids between two anchors is placed evenly on the segment. A low-confidence id in that run is `adjusted`. A missing id is `guessed`. A run of 4 or more is left as it was.
- A run of 1, 2, or 3 ids before the first anchor, or after the last, is filled by repeating that end's neighbour step. The direction is the neighbour step. Each hop is `max_step` long. This runs only when that neighbour is an anchor and the step length is greater than 0.
- When no adjacent anchor pair has length greater than 0, a low-confidence light within 3 ids of an anchor moves along the scanned direction to `id distance × max_gap_m` and is `adjusted`. A missing light is not invented from one anchor. The nearest anchor wins; an id-distance tie uses the lower id.
- With no anchors, nothing is guessed or adjusted.
- `guessed` and `adjusted` are sorted, disjoint from `missing` and `low_confidence`, and present on the result as arrays. Confirm still rejects a non-empty `missing` list. The saved model does not store these labels.
```

In `docs/design/architecture.md`, change the index range `§3.15–§3.23.3` to `§3.15–§3.23.4`.

In `docs/design/backend-service.md`, on the `POST /api/v1/models/capture` row, add optional `max_gap_cm` (centimetres, default 10, stored as `max_gap_m`) and add invalid `max_gap_cm` to the **400** list with the message above. On the `GET /api/v1/models/capture/{jobId}` row, add `guessed` and `adjusted` id arrays to `result`.

In `docs/design/frontend.md` §4.17, add the Maximum gap field (label, helper, prefilled `10`) to the entry bullet, and add the guessed line, the adjusted line, the confirm-refusal sentence, and the four-list rule for the green line to the review bullet.

In `docs/design/overview.md`, add:

```markdown
| REQ-053 | §3.23.4, §4.17: after triangulation, a hole of 1–3 lights is filled along the string and listed in `guessed`; a low-confidence light in that hole is moved and listed in `adjusted`; a run of 4 or more stays in `missing` and blocks confirm; `max_gap_cm` defaults to 10 and is the fallback step only when no confident adjacent pair exists. |
```

In the REQ-048 row, replace `undetected lights are reported, never fabricated` with `a hole of up to three lights is filled by REQ-053, and a longer hole is reported and blocks confirm`. In the REQ-049 row, include guessed and adjusted in the review clause.

In `docs/design/appendix-traceability.md`, add a REQ-053 bullet after REQ-052: short holes in a scanned string are filled and named; a run of four or more stays missing and blocks confirm. Cite §3.23.4 and §4.17.

- [ ] **Step 3: Commit**

```bash
git add docs/requirements/requirements.md docs/requirements/acceptance-criteria.md docs/userguide/build-model-from-video.md docs/design/backend-lights-and-automation.md docs/design/architecture.md docs/design/backend-service.md docs/design/frontend.md docs/design/overview.md docs/design/appendix-traceability.md
git commit -m "$(cat <<'EOF'
docs: describe filling short holes in a scanned light string

EOF
)"
```
