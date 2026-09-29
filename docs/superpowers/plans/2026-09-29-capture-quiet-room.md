# Capture in a dim room — implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let video capture ignore a lamp or window that is already in a dim room before the sweep, without treating the phone's exposure change as a bulb.

**Architecture:** The change stays inside `backend/internal/cvruntime/src/reconstruct.py`. A clip with a red–blue–green bookend builds its background from quiet frames outside that bookend (median, brighter-only, frame-wide lift removed). A clip with no bookend keeps today's darkest-sample background. Go and the web app do not change; the new sentence rides the existing `rejected_feeds` reason and job `error`.

**Tech Stack:** Python 3, OpenCV (`cv2`), NumPy, `unittest`. Product docs in Markdown.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-29-capture-quiet-room-design.md` (REQ-051). REQ-050 numbering does not change.
- Quiet minimum is 0.3 s. Provisional window is the first 1.0 s, then the last 1.0 s only if that found no bookend.
- Room picture is the per-pixel median of up to 15 frames. Do not mix the before side and the after side. The sweep between bookends is never the room picture.
- One bookend with more frames after it than before it is the opening flash: the only candidate is `before`. One bookend with at least as many frames before it as after it is the closing flash: the only candidate is `after`.
- A pixel counts only when `max(0, frame − room) − median(that difference)` is greater than `BLOB_THRESHOLD` (30). Darkening does not count.
- `QUIET_ROOM_REASON` is exactly `the recording needs a moment of the room with the bulbs off, before the opening flash or after the closing flash`.
- When any dropped reason is that sentence and fewer than two clips remain, the job error is exactly `Fewer than two clips could be used. Dropped: {dropped}.` with `{dropped}` as `file (reason)` joined by `"; "`.
- When no dropped reason is that sentence, keep `The recordings missed the start and end flashes, so fewer than two clips could be used.`
- The stays-bright paragraph ends at `Keep the camera still.` Delete `and avoid a bright window behind the lights.`
- No new API fields. Do not revert unrelated edits already in the files this plan touches.
- Rebuild the room picture at most once, and only when the bookend count changes or the quiet boundary moves by more than 2 frames.

---

### Task 1: Quiet-side and excess helpers

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (tunables near `CUE_TAIL_FRAC`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: nothing
- Produces:
  - `QUIET_MIN_S: float = 0.3`
  - `QUIET_SAMPLE_FRAMES: int = 15`
  - `QUIET_BOUNDARY_FRAMES: int = 2`
  - `PROVISIONAL_S: float = 1.0`
  - `QUIET_ROOM_REASON: str` (the sentence in Global Constraints)
  - `_quiet_interval(cue_frames: list[tuple[int, int]], n_frames: int, fps: float) -> tuple[Optional[tuple[str, int, int]], Optional[str]]`
  - `_positive_excess(frame: np.ndarray, baseline: np.ndarray) -> np.ndarray`
  - `_boundary_moved(old: list[tuple[int, int]], new: list[tuple[int, int]], side: str) -> bool`
  - `_sample_indexes(start: int, end: int, k: int) -> list[int]`

`_quiet_interval` returns `((side, start, end_exclusive), None)` or `(None, QUIET_ROOM_REASON)` or `(None, None)` when `cue_frames` is empty (caller uses the legacy background). `side` is `"before"` or `"after"`. Cue tuples are `(first_red_frame, last_green_frame)` inclusive, matching `_cue_frames`.

- [ ] **Step 1: Write the failing test**

Append this class at the end of `test_reconstruct.py`:

```python
class TestQuietRoomHelpers(unittest.TestCase):
    """REQ-051: which frames are the room, and what counts as brighter."""

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestQuietRoomHelpers -v`

Expected: FAIL with `AttributeError` (`_quiet_interval` missing).

- [ ] **Step 3: Write minimal implementation**

Add these tunables next to `CUE_TAIL_FRAC` in `reconstruct.py`:

```python
# REQ-051: quiet room outside the bookend. A side shorter than this is not
# the room with the bulbs off.
QUIET_MIN_S = 0.3
QUIET_SAMPLE_FRAMES = 15
QUIET_BOUNDARY_FRAMES = 2
PROVISIONAL_S = 1.0
QUIET_ROOM_REASON = (
    "the recording needs a moment of the room with the bulbs off, "
    "before the opening flash or after the closing flash"
)
```

Add these functions after `_cue_frames`:

```python
def _quiet_interval(
    cue_frames: list[tuple[int, int]],
    n_frames: int,
    fps: float,
) -> tuple[Optional[tuple[str, int, int]], Optional[str]]:
    """
    Picks the quiet side for the room picture (REQ-051).

    Returns ``((side, start, end_exclusive), None)``, ``(None, QUIET_ROOM_REASON)``,
    or ``(None, None)`` when *cue_frames* is empty. *side* is ``"before"`` or
    ``"after"``. Cue tuples are ``(first_red, last_green)`` inclusive.
    """
    if not cue_frames:
        return None, None
    if not fps or fps <= 0:
        fps = 30.0
    first_red = cue_frames[0][0]
    after_start = cue_frames[-1][1] + 1
    before_n = first_red
    after_n = max(0, n_frames - after_start)
    if len(cue_frames) == 1:
        if after_n > before_n:
            start, end, side = 0, first_red, "before"
        else:
            start, end, side = after_start, n_frames, "after"
    elif before_n / fps >= QUIET_MIN_S:
        start, end, side = 0, first_red, "before"
    else:
        start, end, side = after_start, n_frames, "after"
    if (end - start) / fps + 1e-9 < QUIET_MIN_S:
        return None, QUIET_ROOM_REASON
    return (side, start, end), None


def _positive_excess(frame: np.ndarray, baseline: np.ndarray) -> np.ndarray:
    """Brightening of *frame* over *baseline*, minus the frame-wide median lift."""
    diff = cv2.subtract(frame, baseline)
    lift = int(np.median(diff))
    if lift <= 0:
        return diff
    return cv2.subtract(diff, np.full_like(diff, np.uint8(lift)))


def _boundary_moved(
    old: list[tuple[int, int]],
    new: list[tuple[int, int]],
    side: str,
) -> bool:
    """True when the quiet boundary moved by more than QUIET_BOUNDARY_FRAMES."""
    if not new or len(old) != len(new):
        return True
    if side == "before":
        return abs(old[0][0] - new[0][0]) > QUIET_BOUNDARY_FRAMES
    return abs(old[-1][1] - new[-1][1]) > QUIET_BOUNDARY_FRAMES


def _sample_indexes(start: int, end: int, k: int) -> list[int]:
    """Up to *k* unique indexes in ``[start, end)``, spread across the span."""
    span = end - start
    if span <= 0 or k <= 0:
        return []
    if span <= k:
        return list(range(start, end))
    return [start + (i * span) // k for i in range(k)]
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestQuietRoomHelpers -v`

Expected: PASS (11 tests).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
Pick the quiet side of a capture clip before measuring bulbs.

A closing-only recording must use the frames after the colour flash, so the sweep itself is not treated as the room.
EOF
)"
```

---

### Task 2: Measure bulbs against the quiet room

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_scan_feed`, `analyse_feed`)
- Modify: `backend/internal/cvruntime/src/gen_fixtures.py` (leading gap default, trailing pad)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: Task 1 helpers and `QUIET_ROOM_REASON`
- Produces:
  - `_scan_feed(video_path: str, problems: Optional[list[str]] = None) -> tuple[list, list, float, int, int, np.ndarray]` (same tuple as today: blobs, colours, fps, frame_w, frame_h, K)
  - `analyse_feed(video_path: str, dwell_ms: int, rejections: Optional[list[float]] = None, problems: Optional[list[str]] = None)` with the same return tuple as today
  - `_scan_feed_legacy(video_path: str)` — today's `_scan_feed` body, unchanged

- [ ] **Step 1: Write the failing test**

Append `TestQuietRoom` to `test_reconstruct.py`:

```python
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestQuietRoom -v`

Expected: FAIL. The window clip reports the wrong blink count or centroids on the bright rectangle, and `analyse_feed` does not accept a fourth argument.

- [ ] **Step 3: Lengthen fixture quiet pads**

In `gen_fixtures.py`, change `--leading-gap-ms` default from `50` to `300`, and its help to `Dark gap before the opening bookend (ms); default 300`.

After the closing `_bookend()` call, replace:

```python
frames.extend(_base_frame() for _ in range(cue_frames))
```

with:

```python
# REQ-051: at least 0.3 s of room after the closing bookend.
frames.extend(_base_frame() for _ in range(round(0.3 * FPS)))
```

Leave callers that pass `leading_gap_ms=0` alone. They use this trailing pad.

- [ ] **Step 4: Write the scan**

In `reconstruct.py`:

1. Change `from collections import Counter` to `from collections import Counter, deque`.
2. Rename the existing `_scan_feed` to `_scan_feed_legacy`. Do not change its body.
3. Add `problems: Optional[list[str]] = None` as the last parameter of `analyse_feed`, and pass it into `_scan_feed`. Update the `analyse_feed` docstring with one sentence: a too-short quiet side appends `QUIET_ROOM_REASON` to `problems` when that list is passed, returns the bookends, and returns no blinks.
4. Insert the following above `_scan_feed_legacy`.

```python
def _video_meta(cap: cv2.VideoCapture, video_path: str) -> dict:
    """Width, height, fps, intrinsics, and process resolution for an open capture."""
    frame_w = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    frame_h = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    if frame_w == 0 or frame_h == 0:
        raise IOError(f"Video reports zero dimensions: {video_path!r}")
    fps = cap.get(cv2.CAP_PROP_FPS)
    if not fps or fps <= 0 or math.isnan(fps):
        fps = 30.0
    scale = _downscale_factor(frame_h, frame_w)
    return {
        "frame_w": frame_w,
        "frame_h": frame_h,
        "fps": float(fps),
        "K": _estimate_K(frame_w, frame_h),
        "scale": scale,
        "proc_w": max(1, int(frame_w * scale)),
        "proc_h": max(1, int(frame_h * scale)),
    }


def _median_u8(frames: list[np.ndarray]) -> np.ndarray:
    return np.median(np.stack(frames, axis=0), axis=0).astype(np.uint8)


def _median_baselines(frames: list[np.ndarray]) -> tuple[np.ndarray, np.ndarray]:
    grays = [cv2.cvtColor(f, cv2.COLOR_BGR2GRAY) for f in frames]
    maxes = [f.max(axis=2) for f in frames]
    return _median_u8(grays), _median_u8(maxes)


def _resize_small(frame: np.ndarray, meta: dict) -> np.ndarray:
    return cv2.resize(
        frame, (meta["proc_w"], meta["proc_h"]), interpolation=cv2.INTER_AREA
    )


def _colour_of(small_bgr: np.ndarray, baseline_gray, baseline_max):
    excess_max = _positive_excess(small_bgr.max(axis=2), baseline_max)
    return _frame_colour(small_bgr, excess_max)


def _classify_against_head(video_path: str) -> tuple[dict, list, list]:
    """
    One straight read. Returns ``(meta, colours, tail_frames)``.

    *tail_frames* is the last ``PROVISIONAL_S`` of small BGR frames, used only
    when this pass finds no bookend.
    """
    head: list[np.ndarray] = []
    tail: deque = deque()
    colours: list = []
    baseline = None
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video: {video_path!r}")
        meta = _video_meta(cap, video_path)
        n_prov = max(1, int(round(meta["fps"] * PROVISIONAL_S)))
        tail = deque(maxlen=n_prov)
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            small = _resize_small(frame, meta)
            tail.append(small)
            if baseline is None:
                head.append(small)
                if len(head) < n_prov:
                    continue
                baseline = _median_baselines(head)
                colours.extend(_colour_of(stored, *baseline) for stored in head)
                continue
            colours.append(_colour_of(small, *baseline))
        if baseline is None and head:
            baseline = _median_baselines(head)
            colours = [_colour_of(stored, *baseline) for stored in head]
    finally:
        cap.release()
    return meta, colours, list(tail)


def _classify_against_stored(video_path: str, meta: dict, stored: list) -> list:
    """One straight read. Colours of every frame against the median of *stored*."""
    baseline_gray, baseline_max = _median_baselines(stored)
    colours: list = []
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video: {video_path!r}")
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            colours.append(
                _colour_of(_resize_small(frame, meta), baseline_gray, baseline_max)
            )
    finally:
        cap.release()
    return colours


def _room_baselines(video_path: str, meta: dict, start: int, end: int):
    """One straight read. Median grey and max-channel of samples in ``[start, end)``."""
    wanted = set(_sample_indexes(start, end, QUIET_SAMPLE_FRAMES))
    grays: list[np.ndarray] = []
    maxes: list[np.ndarray] = []
    fi = 0
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video: {video_path!r}")
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            if fi in wanted:
                small = _resize_small(frame, meta)
                grays.append(cv2.cvtColor(small, cv2.COLOR_BGR2GRAY))
                maxes.append(small.max(axis=2))
            fi += 1
            if fi >= end and len(grays) == len(wanted):
                break
    finally:
        cap.release()
    if not grays:
        raise IOError(f"No quiet frames in {video_path!r}")
    return _median_u8(grays), _median_u8(maxes)


def _measure_feed(video_path: str, meta: dict, baseline_gray, baseline_max):
    """One straight read. Blobs and colours against the real room picture."""
    blobs: list = []
    colours: list = []
    scale = meta["scale"]
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video (detection pass): {video_path!r}")
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            small = _resize_small(frame, meta)
            gray = cv2.cvtColor(small, cv2.COLOR_BGR2GRAY)
            blob = _find_bright_blob(_positive_excess(gray, baseline_gray))
            blobs.append(None if blob is None else (blob[0] / scale, blob[1] / scale))
            colours.append(
                _frame_colour(
                    small, _positive_excess(small.max(axis=2), baseline_max)
                )
            )
    finally:
        cap.release()
    return blobs, colours


def _scan_with_room(video_path, meta, colours, cues, problems):
    """Blobs from the quiet-room picture. Appends QUIET_ROOM_REASON when needed."""
    n = len(colours)
    interval, reason = _quiet_interval(cues, n, meta["fps"])
    if reason:
        if problems is not None:
            problems.append(reason)
        return (
            [None] * n, colours, meta["fps"], meta["frame_w"], meta["frame_h"], meta["K"]
        )
    side, start, end = interval
    baseline_gray, baseline_max = _room_baselines(video_path, meta, start, end)
    blobs, measured = _measure_feed(video_path, meta, baseline_gray, baseline_max)
    cues2 = _cue_frames(measured, meta["fps"])
    if not _boundary_moved(cues, cues2, side):
        return (
            blobs, measured, meta["fps"], meta["frame_w"], meta["frame_h"], meta["K"]
        )
    interval2, reason2 = _quiet_interval(cues2, len(measured), meta["fps"])
    if interval2 is None:
        if reason2 and problems is not None:
            problems.append(reason2)
            return (
                [None] * len(measured),
                measured,
                meta["fps"],
                meta["frame_w"],
                meta["frame_h"],
                meta["K"],
            )
        return (
            blobs, measured, meta["fps"], meta["frame_w"], meta["frame_h"], meta["K"]
        )
    _side2, start2, end2 = interval2
    baseline_gray, baseline_max = _room_baselines(video_path, meta, start2, end2)
    blobs, measured = _measure_feed(video_path, meta, baseline_gray, baseline_max)
    return blobs, measured, meta["fps"], meta["frame_w"], meta["frame_h"], meta["K"]


def _scan_feed(video_path: str, problems: Optional[list[str]] = None):
    """
    Blobs and flash colours for one clip (REQ-051).

    A bookend selects the quiet-room picture. No bookend uses `_scan_feed_legacy`.
    """
    meta, colours, tail = _classify_against_head(video_path)
    cues = _cue_frames(colours, meta["fps"])
    if not cues and tail:
        colours = _classify_against_stored(video_path, meta, tail)
        cues = _cue_frames(colours, meta["fps"])
    if not cues:
        return _scan_feed_legacy(video_path)
    return _scan_with_room(video_path, meta, colours, cues, problems)
```

5. In `test_video_capture_released_on_read_error`, the fake capture never returns end-of-file, and the old background pass stopped after a fixed sample count. The new first pass reads until end-of-file, so make `read` stop after 120 frames (the count `get` already reports). Keep the throw on the second capture's first read:

```python
            def read(self):
                self.read_calls += 1
                if self.instance_id >= 1 and self.read_calls == 1:
                    raise RuntimeError("simulated decode failure")
                if self.read_calls > 120:
                    return False, None
                frame = np.zeros((120, 160, 3), dtype=np.uint8)
                return True, frame
```

Replace the hardcoded `2` with the number of captures constructed:

```python
        self.assertEqual(
            len(released),
            FakeCap._seq,
            f"expected every capture released, got {len(released)} release(s)",
        )
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py -v`

Expected: PASS, including `TestQuietRoom`, `TestQuietRoomHelpers`, the existing bookend tests, and `test_continuously_bright_clip_explains_missing_sweep`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py backend/internal/cvruntime/src/gen_fixtures.py
git commit -m "$(cat <<'EOF'
Ignore a lamp or window that is already in the shot before the sweep.

Bulbs are measured against the quiet frames outside the colour flash, so a dim room does not have to be pitch black.
EOF
)"
```

---

### Task 3: Name a clip that has no quiet room

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`main`, `_missed_flashes_message`, `_no_blinks_message`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: `analyse_feed(..., problems)`, `QUIET_ROOM_REASON`, `_number_from_bookends`, `_rejected_feeds`
- Produces: job `error` and `rejected_feeds` behaviour in Global Constraints. No new JSON fields.

- [ ] **Step 1: Write the failing test**

Add these methods to `TestQuietRoom`:

```python
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
```

In `test_continuously_bright_clip_explains_missing_sweep`, after `self.assertIn("capture sweep", err)`, add:

```python
        self.assertNotIn("window", err)
```

- [ ] **Step 2: Run test to verify it fails**

Run these two commands:

`cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestQuietRoom.test_two_short_clips_do_not_claim_the_flashes_were_missed TestQuietRoom.test_good_feeds_still_succeed_when_one_clip_is_short -v`

`cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestRobustBlinkDetection.test_continuously_bright_clip_explains_missing_sweep -v`

Expected: FAIL. The short-clip job still says the flashes were missed, or `rejected_feeds` uses `no light blinks found next to the start or end signal`. The bright-clip error still contains `window`.

- [ ] **Step 3: Write minimal implementation**

Replace `_missed_flashes_message` with:

```python
def _missed_flashes_message(rejected_feeds: list[dict]) -> str:
    dropped = "; ".join(f"{r['file']} ({r['reason']})" for r in rejected_feeds)
    if any(r["reason"] == QUIET_ROOM_REASON for r in rejected_feeds):
        return f"Fewer than two clips could be used. Dropped: {dropped}."
    return (
        "The recordings missed the start and end flashes, so fewer than two "
        f"clips could be used. Dropped: {dropped}."
    )
```

In `_no_blinks_message`, delete `and avoid a bright window behind the lights` so the stays-bright string ends with `Keep the camera still.` Change the docstring sentence `A segment that stays bright for many seconds is the string (or a window) left on, not that sweep.` to `A segment that stays bright for many seconds is the string left on, not that sweep.`

In `main`, replace the stage-1 loop and the numbering hand-off with:

```python
        scans: list[dict] = []
        all_Ks: list[np.ndarray] = []
        rejections: list[float] = []
        quiet_reasons: dict[int, str] = {}
        for feed in feeds:
            problems: list[str] = []
            blinks, cues, _fw, _fh, K = analyse_feed(
                feed["path"], dwell_ms, rejections, problems
            )
            name = feed.get("name") or os.path.basename(feed["path"])
            if problems:
                quiet_reasons[len(scans)] = problems[0]
            scans.append({"name": name, "blinks": blinks, "cues": cues})
            all_Ks.append(K)

        if not any(s["blinks"] for s in scans) and not quiet_reasons:
            _emit_failure(_no_blinks_message(rejections, dwell_ms))

        # Stage 2 — number each feed from its bookend, then keep usable feeds.
        numbered, rejected = _number_from_bookends(scans, light_count)
        for fi, reason in quiet_reasons.items():
            rejected[fi] = reason
        rejected_feeds = _rejected_feeds(scans, rejected)
```

Leave the rest of `main` (usable count, `_emit_failure(_missed_flashes_message(...))`, triangulation) as it is.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
Name a capture clip that never shows the room with the bulbs off.

The job no longer claims the colour flashes were missed when the recording has no quiet moment to compare with.
EOF
)"
```

---

### Task 4: Document the dim-room rule

**Files:**
- Modify: `docs/requirements/requirements.md`
- Modify: `docs/requirements/acceptance-criteria.md`
- Modify: `docs/design/overview.md`
- Modify: `docs/design/appendix-traceability.md`
- Modify: `docs/design/backend-lights-and-automation.md`
- Modify: `docs/userguide/build-model-from-video.md`

**Interfaces:**
- Consumes: the reason sentence and the 0.3 s rule from Task 1
- Produces: REQ-051 in the feature index and the design traceability tables. No new `§` number.

- [ ] **Step 1: Update the requirements tour and checklist**

In `docs/requirements/requirements.md` §11, at the end of item 4 ("The app figures it out"), add:

`In a dim room, a light that is already in the picture before the opening flash is ignored. If that opening flash was missed, a light that is still there after the closing flash is ignored instead. Each bulb still has to be the brightest new thing when it turns on.`

Append this index row (do not renumber):

```markdown
| REQ-051 | Dim-room capture ignores a light already in the shot | §11 Building a model from video |
```

In `docs/requirements/acceptance-criteria.md`, under "Building a model from video", after the "clip that missed the colour flashes" item, add:

```markdown
**A dim room is enough.**
- Try: film the capture sweep in a dim room. A lamp or a bright window may already be in the shot before you press Start capture. Keep the camera still. If you are using the printed marker, leave enough light to see it.
- You should see: the review finds the bulbs. That lamp or window is not treated as a light that stayed on.

**A clip with no moment of the room is named.**
- Try: upload clips that include the colour flashes but have no moment with the bulbs off before the opening flash or after the closing flash.
- You should see: each of those clips named, with a reason that the recording needs a moment of the room with the bulbs off. The app does not say the flashes were missed.
```

- [ ] **Step 2: Update the design and the user guide**

In `docs/design/overview.md`, add a row after REQ-050:

```markdown
| REQ-051 | §3.23: with a bookend, blink detection uses the per-pixel median of the quiet side (before the opening bookend when that stretch is at least 0.3 s, otherwise after the closing bookend; a single bookend uses only the side opposite the sweep); a pixel counts only when it brightens by more than the frame-wide median lift, above the existing 30-level cutoff; a bookend clip with under 0.3 s of quiet room is listed in `rejected_feeds`; a clip with no bookend keeps the darkest-sample background. |
```

In `docs/design/appendix-traceability.md`, after the REQ-050 bullet, add:

```markdown
- **REQ-051** — dim-room capture. A lamp or window already visible before the opening flash is ignored; if that opening flash was missed, a light still visible after the closing flash is ignored instead. Each bulb must still be the brightest new thing. A clip with a bookend but no quiet moment of the room is named and dropped. A clip with no bookend still uses the darkest-sample background. See §3.23.
```

In `docs/design/backend-lights-and-automation.md`, change the §3.23 heading to include REQ-051, and replace the "Per-feed 2D blink detection" bullet with:

```markdown
- **Per-feed 2D blink detection (REQ-051):** when a clip has a red–blue–green bookend, the background is the per-pixel median of up to 15 frames from the quiet side. Two or more bookends use the frames before the opening bookend when that stretch is at least 0.3 s, otherwise the frames after the closing bookend. One bookend is the opening when more frames follow it than precede it (only the before side is a candidate) and the closing when at least as many frames precede it (only the after side is a candidate). The sweep between bookends is never the background. A pixel is lit only when it is brighter than that picture by more than the median brightening of the whole frame; darkening does not count. The cutoff stays 30 levels (`BLOB_THRESHOLD`). The largest lit blob is the bulb. If the chosen quiet side is under 0.3 s, the clip is dropped with reason `the recording needs a moment of the room with the bulbs off, before the opening flash or after the closing flash` and is not numbered. A clip with no bookend keeps the darkest-sample background (per-pixel minimum of `BG_FRAMES` spread samples and an absolute difference) so a string left on still reports that the clip stays bright. Slot assignment (§3.23.3) is unchanged. Frames are downscaled as needed for Pi performance.
```

In `docs/userguide/build-model-from-video.md`, replace the stays-bright bullet under "Good to know" with these three bullets, in order:

```markdown
- A dim room is enough. A lamp or a window that is already in the shot before the opening colour flash is ignored. The bulbs still need to be the brightest new thing when they turn on, and the camera needs to stay still. Light on the printed marker is fine.
- If it says the clip **stays bright** and no individual blinks were found, the bulbs were on together, or something lit up during the sweep and stayed on. Film again with **Start capture**, so each bulb lights by itself for about a second, and keep the camera still.
- If a clip is set aside because the recording needs a moment of the room with the bulbs off, start filming before **Start capture** and keep filming until after the closing colour flash, with the bulbs off in those moments.
```

- [ ] **Step 3: Check the sentences are present**

Run:

```bash
rg -n "REQ-051" docs/requirements/requirements.md docs/design/overview.md docs/design/appendix-traceability.md docs/design/backend-lights-and-automation.md
rg -n "moment of the room" docs/requirements/acceptance-criteria.md docs/userguide/build-model-from-video.md
rg -n "avoid a bright window" docs/userguide/build-model-from-video.md backend/internal/cvruntime/src/reconstruct.py
```

Expected: REQ-051 appears in all four doc files. "moment of the room" appears in the checklist and the user guide. The last search prints no matches.

- [ ] **Step 4: Commit**

```bash
git add docs/requirements/requirements.md docs/requirements/acceptance-criteria.md docs/design/overview.md docs/design/appendix-traceability.md docs/design/backend-lights-and-automation.md docs/userguide/build-model-from-video.md
git commit -m "$(cat <<'EOF'
Document that a dim room is enough for capture.

A light already in the shot before the colour flash is ignored, and a clip with no quiet moment of the room is named.
EOF
)"
```
