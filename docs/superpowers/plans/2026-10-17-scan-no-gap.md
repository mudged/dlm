# Scan bulbs with no dark gap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make video reconstruction separate one-second bulbs that follow each other with no all-off frame, while ignoring a window or a leg that stays bright for the whole clip.

**Architecture:** Three changes inside the existing scanner, all in `reconstruct.py`. A faint white smear between the red, blue, and green pulses still counts as the bookend gap, so the quiet-room picture is used. Pixels lit for longer than three dwells are room, not a bulb. A blink ends when the lit spot goes dark or when it stays jumped to a new place for two frames. If the room mask leaves no dwell-length blink, the scan falls back to the unmasked picture so a string left on still says the clip stays bright.

**Tech Stack:** Python 3.12, OpenCV, NumPy, `unittest`. The Go server is unchanged.

## Global Constraints

- The Go server stays pure Go with no cgo. This change is the Python scanner plus a copy into the existing CV bundle.
- Raspberry Pi 4: the room mask is one extra sequential decode. Keep it at processing resolution (shorter side ≤ 480) and one `uint16` count image. Do not store frames.
- Do not renumber or reuse `§3.23` or `REQ-051`.
- Do not add a requirement and do not edit `docs/requirements/` or `docs/userguide/`. REQ-051 already says a light already in the picture is ignored and each bulb is the brightest new thing when it turns on.
- Dwell acceptance stays `[0.4, 2.5] × dwell_ms`. `BLOB_THRESHOLD` stays 30. `CUE_TAIL_FRAC` stays 0.25. Cue pulse and gap ranges stay as they are.
- New constants, exact values: `ROOM_LIT_FACTOR = 3`, `BLINK_JUMP_FRAC = 0.025`, `BLINK_JUMP_FRAMES = 2`.
- Do not change `gen_fixtures.py`. Its 50 ms gap stays. New tests build their own frames.
- Tests run from `backend/internal/cvruntime/src`. Use `python3` when `import cv2` works; otherwise the bundle interpreter:

```bash
cd /workspaces/dlm/backend/internal/cvruntime/src
PY=python3
python3 -c "import cv2" 2>/dev/null || PY=/workspaces/dlm/dist/cvruntime/linux_amd64/python/bin/python
```

Every later test command assumes that `PY` and that working directory.

---

### Task 1: Bookend gap may be a faint white smear

On `20261017_180459.mp4` the opening flash is red, then white, then blue, then white, then green. The white intervals last 200 ms and their bright area is a few percent of the colour pulses, so today's matcher (gap must be fully dark) finds no bookend and never uses the quiet-room picture.

**Files:**
- Modify: `docs/design/backend-lights-and-automation.md` (the REQ-051 bullet in §3.23)
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_cue_frames`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py` (`TestQuietRoom`)

**Interfaces:**
- Consumes: `CUE_SEQUENCE`, `CUE_GAP_MIN_S`, `CUE_GAP_MAX_S`, `CUE_TAIL_FRAC`, `_runs`, `_cue_frames(colours, fps) -> list[tuple[int, int]]`
- Produces: `_cue_frames` treats a non-colour run as dark when it lasts a cue gap and its peak area is `< CUE_TAIL_FRAC` times the smaller neighbouring colour peak. Later tasks still call `_cue_frames` the same way.

- [ ] **Step 1: Update the §3.23 blink bullet**

In `docs/design/backend-lights-and-automation.md`, in the bullet that starts `**Per-feed 2D blink detection (REQ-051):**`, replace the sentence `The largest lit blob is the bulb.` with:

```markdown
The largest lit blob is the bulb. A pixel lit for longer than three dwells is part of the room and is not a bulb, even when exposure makes it brighter than the quiet picture. If that mask leaves no blink that fits the dwell, the scan uses the unmasked picture, so a string left on still reports that the clip stays bright. A blink also ends when the lit spot stays more than 0.025 of the shorter frame side away from where it has been, for two frames in a row, so the next bulb can follow with no dark frame. A short bright smear between the red, blue, and green pulses still counts as the dark gap when its area is under a quarter of both neighbouring pulses.
```

Leave the rest of the bullet, including the quiet-side rules and the no-bookend sentence, unchanged. Do not renumber the section.

- [ ] **Step 2: Write the failing test**

Add this method to `TestQuietRoom` in `test_reconstruct.py`:

```python
    def test_white_smear_between_colours_is_still_a_bookend(self):
        frames = [self._frame() for _ in range(15)]
        for i, colour in enumerate((self.RED, self.BLUE, self.GREEN)):
            if i:
                smear = self._frame()
                cv2.circle(smear, self.BULBS[0], 2, (255, 255, 255), -1)
                frames += [smear for _ in range(6)]
            frames += [self._frame(self.BULBS, colour) for _ in range(6)]
        frames += [self._frame() for _ in range(15)]
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "smear.avi")
            self._write(path, frames)
            cues = self.m.find_cues(path)
        self.assertEqual(len(cues), 1, cues)
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `"$PY" test_reconstruct.py TestQuietRoom.test_white_smear_between_colours_is_still_a_bookend -v`

Expected: FAIL with `0 != 1` (no cue).

- [ ] **Step 4: Treat a small non-colour gap as dark**

In `_cue_frames`, after the short-pulse nulling and after `runs = _runs(labels)`, relabel those gap frames and rebuild `runs` before the `while` loop. `lasts` must already be defined; move the existing `lasts` helper above this block.

```python
    def lasts(run, lo: float, hi: float) -> bool:
        d = (run[2] - run[1] + 1) / fps
        return lo - 1e-9 <= d <= hi + 1e-9

    runs = _runs(labels)
    for idx in range(1, len(runs) - 1):
        label, first, last = runs[idx]
        if label is None or label in CUE_SEQUENCE:
            continue
        if not lasts(runs[idx], CUE_GAP_MIN_S, CUE_GAP_MAX_S):
            continue
        prev_l, prev_a, prev_b = runs[idx - 1]
        next_l, next_a, next_b = runs[idx + 1]
        if prev_l not in CUE_SEQUENCE or next_l not in CUE_SEQUENCE:
            continue
        peak = max(colours[i][1] for i in range(first, last + 1))
        prev_peak = max(colours[i][1] for i in range(prev_a, prev_b + 1))
        next_peak = max(colours[i][1] for i in range(next_a, next_b + 1))
        if peak < CUE_TAIL_FRAC * min(prev_peak, next_peak):
            for i in range(first, last + 1):
                labels[i] = None
    runs = _runs(labels)
```

Delete the later duplicate `def lasts` so only one remains.

- [ ] **Step 5: Run the tests**

Run: `"$PY" test_reconstruct.py TestQuietRoom.test_white_smear_between_colours_is_still_a_bookend TestQuietRoom -v`

Expected: PASS, including the existing quiet-room tests.

- [ ] **Step 6: Commit**

```bash
git add docs/design/backend-lights-and-automation.md backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
fix: accept a faint white smear between bookend colours

The phone clips leave a short white gap between red, blue, and green, so the bookend was missed and the quiet-room picture was never used.
EOF
)"
```

---

### Task 2: Pixels lit for longer than three dwells are the room

The quiet median still leaves the leg and the window above the threshold, because exposure lifts them during the sweep. They are the largest blob, so the bulb never wins. A one-second bulb is lit for about one dwell plus the short bookend pulses. A leg or a window is lit for the whole sweep.

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_room_pixel_mask`, `_measure_feed`, `_scan_with_room`, `_scan_feed`, `_scan_feed_legacy`, `analyse_feed`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py` (`TestQuietRoom`)

**Interfaces:**
- Consumes: `_positive_excess`, `_find_bright_blob`, `BLOB_THRESHOLD`, `analyse_feed(video_path, dwell_ms, rejections=None, problems=None)`
- Produces:
  - `ROOM_LIT_FACTOR = 3`
  - `_room_pixel_mask(video_path: str, meta: dict, baseline_gray: np.ndarray, dwell_ms: int) -> np.ndarray` boolean room mask at processing resolution
  - `_scan_feed(video_path, problems=None, dwell_ms=1000, mask_room=True)` same return tuple as today. `mask_room=False` is today's unmasked picture.
  - `analyse_feed` scans with the mask. When that yields no dwell-length blink and `problems` is empty, it scans once more with `mask_room=False` and uses that blink list and its rejection durations. A short quiet room (`problems` already set) is not retried.

- [ ] **Step 1: Write the failing test**

Add `_frame`'s glare flag and this test to `TestQuietRoom`. Keep the existing `f[0:24, 110:150] = 180` line; only raise it when `glare` is true.

```python
    def _frame(self, spots=(), colour=(255, 255, 255), lift=0, glare=False):
        f = np.zeros((self.H, self.W, 3), dtype=np.uint8)
        f[0:24, 110:150] = 230 if glare else 180
        for cx, cy in spots:
            cv2.circle(f, (cx, cy), 6, colour, -1)
        if lift:
            f = np.clip(f.astype(np.int16) + lift, 0, 255).astype(np.uint8)
        return f

    def test_glare_bigger_than_a_bulb_is_not_the_blink(self):
        # Dwell is 1 s (30 frames). A bulb is also on for both bookends
        # (about 1.2 s), still under the 3 s room cutoff. The rectangle is
        # brighter for the settle, the sweep, and the tail: 6+102+15 frames.
        frames = [self._frame() for _ in range(15)] + self._cue()
        frames += [self._frame(glare=True) for _ in range(6)]
        for i, bulb in enumerate(self.BULBS):
            frames += [self._frame([bulb], glare=True) for _ in range(30)]
            if i != len(self.BULBS) - 1:
                frames += [self._frame(glare=True) for _ in range(6)]
        frames += [self._frame(glare=True) for _ in range(15)]
        frames += self._cue() + [self._frame() for _ in range(15)]
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "glare.avi")
            self._write(path, frames)
            rejections: list[float] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 1000, rejections)
        self.assertEqual(len(cues), 2, cues)
        self._assert_on_bulbs(blinks)
        self.assertFalse(any(d >= 3 for d in rejections), rejections)
```

Replace the existing `_frame` method with this one. Call sites that omit `glare` stay on the value-180 rectangle.

- [ ] **Step 2: Run the test to verify it fails**

Run: `"$PY" test_reconstruct.py TestQuietRoom.test_glare_bigger_than_a_bulb_is_not_the_blink -v`

Expected: FAIL in `_assert_on_bulbs` (the centroid sits on the rectangle near `(130, 12)`, or the blink count is wrong).

- [ ] **Step 3: Add the room mask and the unmasked fallback**

Next to the other tunables:

```python
# A pixel lit longer than this many dwells is the room (window, leg, desk),
# not a one-second bulb. Three dwells is longer than one bulb plus the bookend.
ROOM_LIT_FACTOR = 3
```

Add:

```python
def _room_pixel_mask(
    video_path: str,
    meta: dict,
    baseline_gray: np.ndarray,
    dwell_ms: int,
) -> np.ndarray:
    """True where a pixel is lit for longer than ROOM_LIT_FACTOR dwells."""
    limit = ROOM_LIT_FACTOR * (max(dwell_ms, 1) / 1000.0) * meta["fps"]
    counts = np.zeros((meta["proc_h"], meta["proc_w"]), dtype=np.uint16)
    cap = cv2.VideoCapture(video_path)
    try:
        if not cap.isOpened():
            raise IOError(f"Cannot open video: {video_path!r}")
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            gray = cv2.cvtColor(_resize_small(frame, meta), cv2.COLOR_BGR2GRAY)
            excess = _positive_excess(gray, baseline_gray)
            counts += (excess > BLOB_THRESHOLD).astype(np.uint16)
    finally:
        cap.release()
    return counts > limit
```

`_measure_feed` gains `room_mask=None`. After computing the gray excess, zero room pixels before `_find_bright_blob`. Do not zero the colour excess; bookend colour still uses the unmasked picture.

```python
def _measure_feed(video_path, meta, baseline_gray, baseline_max, room_mask=None):
    ...
            excess = _positive_excess(gray, baseline_gray)
            if room_mask is not None:
                excess = excess.copy()
                excess[room_mask] = 0
            blob = _find_bright_blob(excess)
```

Thread `dwell_ms: int = 1000` and `mask_room: bool = True` from `analyse_feed` into `_scan_feed`, `_scan_with_room`, and `_scan_feed_legacy`.

In `_scan_with_room`, after the quiet baseline is known and before `_measure_feed`:

```python
    room = (
        _room_pixel_mask(video_path, meta, baseline_gray, dwell_ms)
        if mask_room else None
    )
    blobs, measured = _measure_feed(
        video_path, meta, baseline_gray, baseline_max, room
    )
```

There are two `_measure_feed` call sites in `_scan_with_room` (the first quiet side, and the rebuilt side). Pass `room` into both. Build `room` once from the first baseline; reuse that mask on the rebuild. The rebuild exists to fix the quiet boundary, not to relearn the room.

In `_scan_feed_legacy`, after `background` is built and before pass 2, when `mask_room` is true:

```python
    room = _room_pixel_mask(video_path, meta_from_legacy, background, dwell_ms)
```

Build that `meta` dict from the width, height, fps, and proc size pass 1 already computed (`frame_w`, `frame_h`, `fps`, `proc_w`, `proc_h`, `scale`). On pass 2, copy `diff` and set `diff[room] = 0` before `_find_bright_blob`. Colour stays on the unmasked diff. When `mask_room` is false, `room` is `None` and pass 2 is unchanged.

In `analyse_feed`, pass `dwell_ms`. If the masked scan yields no dwell-length blink and `problems` is empty, scan again unmasked and keep the first scan's cue frames (null those frames on the raw blobs too):

```python
    blobs, colours, fps, frame_w, frame_h, K = _scan_feed(
        video_path, problems, dwell_ms, True
    )
    cue_frames = _cue_frames(colours, fps)
    blinks = _blinks_from_blobs(
        _without_cues(blobs, cue_frames), fps, dwell_ms, None
    )
    if not blinks and not problems:
        raw, _colours, fps, frame_w, frame_h, K = _scan_feed(
            video_path, None, dwell_ms, False
        )
        blinks = _blinks_from_blobs(
            _without_cues(raw, cue_frames), fps, dwell_ms, rejections
        )
```

`_without_cues` is the existing loop that sets cue frames to `None`, extracted so both lists share it:

```python
def _without_cues(blobs, cue_frames):
    out = list(blobs)
    for first, last in cue_frames:
        for i in range(first, min(last + 1, len(out))):
            out[i] = None
    return out
```

The first `_blinks_from_blobs` call passes `rejections=None` so a masked miss does not record durations that the unmasked retry will record itself.

- [ ] **Step 4: Run the tests**

Run: `"$PY" test_reconstruct.py TestQuietRoom -v`

Expected: PASS. `test_bright_clips_keep_the_stays_bright_hint_beside_a_short_clip` still contains `stays bright`. `test_short_quiet_room_sets_the_reason_and_keeps_the_bookends` still returns no blinks and `QUIET_ROOM_REASON`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
fix: ignore pixels that stay lit for longer than three dwells

Exposure lifts a window or a leg above the quiet picture for the whole sweep, and that patch was winning over the bulb.
EOF
)"
```

---

### Task 3: Split a blink when the lit spot jumps

The sweep turns the next bulb on as the previous one turns off. There is no dark frame. After Task 2 the lit spot is the bulb, and it jumps. Today's state machine keeps one blink open until the spot disappears, so a whole sweep is rejected as too long.

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_blinks_from_blobs`, its call sites in `analyse_feed`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: `_blink_duration_s`, `_accept_blink_duration`, `_finalise_blink`, `ROOM_LIT_FACTOR` mask from Task 2, cue nulling from Task 1
- Produces:
  - `BLINK_JUMP_FRAC = 0.025`
  - `BLINK_JUMP_FRAMES = 2`
  - `_blinks_from_blobs(blobs, fps, dwell_ms, rejections, shorter_side: int)` 
  - Both call sites in `analyse_feed` pass `min(frame_w, frame_h)`.

- [ ] **Step 1: Write the failing tests**

Add to `TestQuietRoom`:

```python
    def test_no_dark_gap_between_bulbs_is_three_blinks(self):
        frames = (
            [self._frame() for _ in range(15)]
            + self._cue()
            + [self._frame() for _ in range(6)]
        )
        for bulb in self.BULBS:
            frames += [self._frame([bulb]) for _ in range(30)]
        frames += self._cue() + [self._frame() for _ in range(15)]
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "nogap.avi")
            self._write(path, frames)
            rejections: list[float] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 1000, rejections)
        self.assertEqual(len(cues), 2, cues)
        self._assert_on_bulbs(blinks)
        self.assertFalse(any(d >= 3 for d in rejections), rejections)

    def test_one_frame_centroid_spike_stays_one_blink(self):
        fps = 30
        blobs = [(40.0, 60.0)] * 10 + [(100.0, 100.0)] + [(40.0, 60.0)] * 10
        blinks = self.m._blinks_from_blobs(blobs, fps, 500, [], 120)
        self.assertEqual(len(blinks), 1, blinks)
        _t, cx, cy = blinks[0]
        self.assertLess(math.hypot(cx - 40, cy - 60), 5)

    def test_phone_shape_smear_glare_and_no_gap(self):
        frames = [self._frame() for _ in range(15)]
        for i, colour in enumerate((self.RED, self.BLUE, self.GREEN)):
            if i:
                smear = self._frame()
                cv2.circle(smear, self.BULBS[0], 2, (255, 255, 255), -1)
                frames += [smear for _ in range(6)]
            frames += [self._frame(self.BULBS, colour) for _ in range(6)]
        frames += [self._frame(glare=True) for _ in range(6)]
        for bulb in self.BULBS:
            frames += [self._frame([bulb], glare=True) for _ in range(30)]
        frames += [self._frame(glare=True) for _ in range(15)]
        frames += self._cue() + [self._frame() for _ in range(15)]
        with tempfile.TemporaryDirectory() as d:
            path = str(Path(d) / "phone.avi")
            self._write(path, frames)
            rejections: list[float] = []
            blinks, cues, *_ = self.m.analyse_feed(path, 1000, rejections)
        self.assertEqual(len(cues), 2, cues)
        self._assert_on_bulbs(blinks)
        self.assertFalse(any(d >= 3 for d in rejections), rejections)
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `"$PY" test_reconstruct.py TestQuietRoom.test_no_dark_gap_between_bulbs_is_three_blinks TestQuietRoom.test_one_frame_centroid_spike_stays_one_blink TestQuietRoom.test_phone_shape_smear_glare_and_no_gap -v`

Expected: the no-gap test FAILs (`len(blinks)` is 0 because 3.0 s is outside the 1000 ms dwell window). The spike test FAILs if a one-frame jump already splits, or errors if `shorter_side` is not accepted yet. The phone-shape test FAILs.

- [ ] **Step 3: Split on a jump that holds for two frames**

Add the constants next to `ROOM_LIT_FACTOR`:

```python
# Fraction of the shorter full-resolution side. A lit spot that stays this far
# from the blink so far, for BLINK_JUMP_FRAMES in a row, is the next bulb.
BLINK_JUMP_FRAC = 0.025
BLINK_JUMP_FRAMES = 2
```

Replace `_blinks_from_blobs` with:

```python
def _blinks_from_blobs(
    blobs: list[Optional[tuple[float, float]]],
    fps: float,
    dwell_ms: int,
    rejections: Optional[list[float]],
    shorter_side: int,
) -> list[tuple[float, float, float]]:
    """On/off state machine over per-frame blob centroids (see `detect_blinks`).

    A dark frame ends a blink. So does a centroid that stays more than
    ``BLINK_JUMP_FRAC`` of *shorter_side* away from the blink's mean, for
    ``BLINK_JUMP_FRAMES`` in a row. One far frame is a glitch and stays in
    the current blink.
    """
    jump_px = BLINK_JUMP_FRAC * max(shorter_side, 1)
    blinks: list[tuple[float, float, float]] = []
    blink_start_frame = 0
    in_blink = False
    accum: list[tuple[float, float]] = []
    away: list[tuple[int, tuple[float, float]]] = []

    def close(end_frame: int) -> None:
        nonlocal in_blink, accum, away
        if not accum or end_frame < blink_start_frame:
            accum = []
            away = []
            in_blink = False
            return
        duration_s = _blink_duration_s(blink_start_frame, end_frame, fps)
        if _accept_blink_duration(duration_s, dwell_ms):
            _finalise_blink(blinks, blink_start_frame / fps, accum)
        else:
            _log(
                f"    rejected blink at frame {blink_start_frame}: "
                f"duration {duration_s * 1000:.0f} ms outside "
                f"dwell window ({dwell_ms} ms)"
            )
            if rejections is not None:
                rejections.append(duration_s)
        accum = []
        away = []
        in_blink = False

    def open_from(pending: list[tuple[int, tuple[float, float]]]) -> None:
        nonlocal in_blink, blink_start_frame, accum, away
        started = pending[0][0]
        points = [p for _, p in pending]
        close(started - 1)
        in_blink = True
        blink_start_frame = started
        accum = points
        away = []

    for frame_idx, blob in enumerate(blobs):
        if blob is None:
            if in_blink:
                if len(away) >= BLINK_JUMP_FRAMES:
                    open_from(away)
                elif away:
                    accum.extend(p for _, p in away)
                    away = []
                close(frame_idx - 1)
            continue
        if not in_blink:
            in_blink = True
            blink_start_frame = frame_idx
            accum = [blob]
            away = []
            continue
        cx = sum(p[0] for p in accum) / len(accum)
        cy = sum(p[1] for p in accum) / len(accum)
        if math.hypot(blob[0] - cx, blob[1] - cy) > jump_px:
            away.append((frame_idx, blob))
            if len(away) >= BLINK_JUMP_FRAMES:
                open_from(away)
            continue
        if away:
            accum.extend(p for _, p in away)
            away = []
        accum.append(blob)

    if in_blink:
        if len(away) >= BLINK_JUMP_FRAMES:
            open_from(away)
            close(blink_start_frame + len(accum) - 1)
        else:
            accum.extend(p for _, p in away)
            close(blink_start_frame + len(accum) - 1)
    return blinks
```

Pass `min(frame_w, frame_h)` as `shorter_side` at both `analyse_feed` call sites added in Task 2. In `detect_blinks`'s docstring, replace the sentence that begins `Frame-rate independence:` with: `Blinks end on a dark frame or when the lit spot stays jumped for two frames, so cameras at different frame rates still line up, including a sweep with no dark frame between bulbs.` No other caller constructs blinks.

In `_scan_feed_legacy`, the `meta` dict passed to `_room_pixel_mask` is exactly:

```python
    meta = {
        "frame_w": frame_w,
        "frame_h": frame_h,
        "fps": float(fps),
        "scale": scale,
        "proc_w": proc_w,
        "proc_h": proc_h,
    }
```

- [ ] **Step 4: Run the tests**

Run: `"$PY" test_reconstruct.py -v`

Expected: PASS, including `TestQuietRoom` and the stays-bright tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
fix: start a new blink when the lit spot jumps

The sweep turns the next bulb on immediately, so a dark frame never arrives to split the previous one.
EOF
)"
```

---

### Task 4: `run.sh` picks up a newer scanner, then check the phone clips

`./scripts/run.sh` rebuilds the web UI and starts the Go server. It points `DLM_CV_RUNTIME_DIR` at `dist/cvruntime/linux_<arch>` and runs the `reconstruct.py` copied there when the bundle was built. It does not rebuild OpenCV. The running bundle is still the 29 September copy.

**Files:**
- Modify: `scripts/run.sh`
- Modify: `docs/engineering/build-and-run.md` (the `DLM_CV_RUNTIME_DIR` row only)

**Interfaces:**
- Consumes: the source file `backend/internal/cvruntime/src/reconstruct.py` from Tasks 1–3
- Produces: when `run.sh` selects `dist/cvruntime/linux_<arch>` and the source `reconstruct.py` is newer than the bundle copy, it copies the source over the bundle copy before `exec go run`.

- [ ] **Step 1: Copy the scanner when the source is newer**

In `scripts/run.sh`, inside the block that sets `DLM_CV_RUNTIME_DIR` because `${cv_dir}/python` and `${cv_dir}/reconstruct.py` exist, after the `export`:

```bash
  src_py="${ROOT}/backend/internal/cvruntime/src/reconstruct.py"
  if [[ -f "${src_py}" && "${src_py}" -nt "${cv_dir}/reconstruct.py" ]]; then
    cp "${src_py}" "${cv_dir}/reconstruct.py"
    echo "run.sh: updated ${cv_dir}/reconstruct.py from source"
  fi
```

- [ ] **Step 2: Document that row**

In `docs/engineering/build-and-run.md`, extend the `DLM_CV_RUNTIME_DIR` row so it also says: if unset, `run.sh` points at `dist/cvruntime/linux_<arch>/` when that directory exists, and copies `backend/internal/cvruntime/src/reconstruct.py` over the bundle copy when the source file is newer. It does not re-download Python or OpenCV; that remains `scripts/build-cvruntime.sh`.

- [ ] **Step 3: Confirm the three phone clips**

The server must be restarted with `./scripts/run.sh` before a UI upload uses the new copy. This step checks the files already on disk, without the UI.

The job copies are:

- `/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_180459.mp4`
- `/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_180846.mp4`
- `/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_181158.mp4`

Run, with the bundle interpreter:

```bash
cd /workspaces/dlm/backend/internal/cvruntime/src
/workspaces/dlm/dist/cvruntime/linux_amd64/python/bin/python - << 'PY'
import importlib.util
spec = importlib.util.spec_from_file_location("reconstruct", "reconstruct.py")
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
files = [
    "/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_180459.mp4",
    "/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_180846.mp4",
    "/workspaces/dlm/backend/data/runtime/capture/34149818-23f5-4065-8a0e-f2802d939b88/20261017_181158.mp4",
]
for path in files:
    rejections = []
    blinks, cues, *_ = m.analyse_feed(path, 1000, rejections)
    long = [round(d, 1) for d in rejections if d >= 3]
    print(path.split("/")[-1], "blinks", len(blinks), "cues", len(cues), "long_rejections", long)
PY
```

Expected: each line has `cues` ≥ 1, `blinks` ≥ 10, and `long_rejections` empty. One clip can take a few minutes. A line with `blinks 0` or a rejection near 160 seconds means the mask or the jump threshold is wrong; stop and fix that, do not retune by editing the phone files.

- [ ] **Step 4: Commit**

```bash
git add scripts/run.sh docs/engineering/build-and-run.md
git commit -m "$(cat <<'EOF'
fix: refresh the bundled scanner when its source is newer

run.sh was still launching the CV bundle's old reconstruct.py after the source had changed.
EOF
)"
```
