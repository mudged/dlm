# Bookend slot offset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep dwell blinks on a clip that has both colour flashes when the backward index sits a steady offset away from the forward index, and drop only a blink whose offset matches neither neighbour.

**Architecture:** One change inside `_number_from_bookends` in `reconstruct.py`. The backward walk ends on the forward count's last slot. A new helper keeps a forward slot when its offset equals the previous or next blink's offset. Opening-only and closing-only numbering, ambiguous-slot dropping, and `light_count` filtering stay as they are. The design sentence in §3.23.3 is updated to match. The dev server reads the copied bundle script, so that one file is copied after the tests pass.

**Tech Stack:** Python 3.12, `unittest`. No OpenCV work in the new tests. The Go server is unchanged.

## Global Constraints

- Do not add a `REQ-*` code. Do not edit `docs/requirements/` or `docs/userguide/`.
- Do not renumber or reuse `§3.23.3`.
- Only the `kind == "both"` branch changes. Opening-only and closing-only numbering stay as they are.
- The backward anchor is `forward[-1]`, not `round((times[-1] - times[0]) / period)`.
- Offset is `backward index − forward index`. A blink is kept at its forward index when that offset equals the previous blink's offset or the next blink's offset. Neighbours are the blinks next to it in time, compared before any blink is removed.
- A blink whose offset differs from every adjacent blink is discarded. One blink with no neighbour is kept.
- A change of offset between two stretches does not drop either boundary blink.
- After this filter, `_add_feed_slots` still drops a slot that one feed filled twice.
- `light_count` still discards indexes at or above `light_count` and does not renumber a clip that has the opening flash.
- The failure text `no blink events detected in any feed` stays.
- The three phone clips under `/workspaces/dlm/captures` are not test fixtures.
- Tests run from `backend/internal/cvruntime/src`:

```bash
cd /workspaces/dlm/backend/internal/cvruntime/src
PY=python3
python3 -c "import cv2" 2>/dev/null || PY=/workspaces/dlm/dist/cvruntime/linux_amd64/python/bin/python
```

Every later test command assumes that `PY` and that working directory.

## File map

- `docs/design/backend-lights-and-automation.md` — replace the both-anchors sentence in §3.23.3.
- `backend/internal/cvruntime/src/reconstruct.py` — add `_agreeing_forward_slots` and use it from the both-anchors branch.
- `backend/internal/cvruntime/src/test_reconstruct.py` — two `TestBookends` cases that call `number_from_bookends`.
- `dist/cvruntime/linux_amd64/reconstruct.py` — a copy of the source script so the running server picks it up. Do not run `./scripts/build-cvruntime.sh`.

---

### Task 1: Keep a steady forward/backward offset

**Files:**
- Modify: `docs/design/backend-lights-and-automation.md` (the both-anchors paragraph in §3.23.3)
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_number_from_bookends`, new `_agreeing_forward_slots`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py` (`TestBookends`)

**Interfaces:**
- Consumes: `_assign_slots(times: list[float], period: float) -> list[int]`, `_backward_slots(times: list[float], period: float, last_index: int) -> list[int]`, `_add_feed_slots`, `number_from_bookends(feeds, light_count) -> tuple[dict, list]`
- Produces: `_agreeing_forward_slots(forward: list[int], backward: list[int]) -> list[Optional[int]]`. Each item is the forward slot, or `None` when that blink is discarded. The both-anchors branch assigns that list to `slots_by_feed[fi]`.

- [ ] **Step 1: Update the §3.23.3 both-anchors paragraph**

In `docs/design/backend-lights-and-automation.md`, replace this text:

```markdown
When both anchors exist, number each blink forward from the opening signal (index 0) and backward from
the closing signal, where the closing signal marks the last slot of that same span. Discard a blink
that receives two different indexes. Leave the surrounding indexes as they are.
```

with:

```markdown
When both anchors exist, number each blink forward from the opening signal (index 0) and backward from the closing signal. The backward count ends on the same slot the forward count gave the last blink. The offset is the backward index minus the forward index. A blink whose offset equals the offset of the previous blink or the next blink is kept at the forward index. A blink whose offset differs from both of those, or from its only neighbour, is discarded. One blink alone is kept. Where the offset changes between two stretches, both blinks at that change stay, because each matches the rest of its own stretch. Leave the other indexes as they are.
```

Leave the following `light_count` sentences in that paragraph unchanged. Do not renumber the section.

- [ ] **Step 2: Write the failing tests**

Add these methods to `TestBookends` in `test_reconstruct.py`, immediately after `test_blinks_outside_both_cues_ignored_and_first_is_slot_zero`:

```python
    def test_steady_offset_keeps_the_forward_indexes(self):
        times = [0, 0.6, 1.2, 1.8, 2.8, 3.8, 4.8, 5.8, 6.8]
        feed = self._feed(
            "steady.mp4",
            [(t, i * 10) for i, t in enumerate(times)],
            [(-1.0, -0.5), (7.0, 7.5)],
        )
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(rejected, [])
        self.assertEqual(sorted(by_light), list(range(9)))

    def test_odd_blink_is_dropped_and_its_slot_stays(self):
        feed = self._feed(
            "extra.mp4",
            [(0.0, 0), (1.5, 15), (1.9, 19), (2.3, 23), (3.8, 38)],
            [(-1.0, -0.5), (4.5, 5.0)],
        )
        by_light, rejected = self.m.number_from_bookends([feed], None)

        self.assertEqual(rejected, [])
        self.assertEqual(sorted(by_light), [0, 2, 3, 5])
        self.assertEqual(by_light[2], [(0, 15.0, 0.0)])

    def test_offset_change_between_stretches_keeps_both_boundaries(self):
        # Offsets are 0, 0, 1, 1. Each boundary blink matches its own stretch.
        kept = self.m._agreeing_forward_slots([0, 1, 2, 3], [0, 1, 3, 4])
        self.assertEqual(kept, [0, 1, 2, 3])
```

- [ ] **Step 3: Run the tests to verify they fail**

Run:

```bash
"$PY" test_reconstruct.py \
  TestBookends.test_steady_offset_keeps_the_forward_indexes \
  TestBookends.test_odd_blink_is_dropped_and_its_slot_stays \
  TestBookends.test_offset_change_between_stretches_keeps_both_boundaries -v
```

Expected: all three FAIL. The steady test fails with `[] != [0, 1, 2, 3, 4, 5, 6, 7, 8]`. The extra-blink test fails with `[] != [0, 2, 3, 5]`. The stretch test fails with `AttributeError: module 'reconstruct' has no attribute '_agreeing_forward_slots'`.

- [ ] **Step 4: Implement the offset check**

In `reconstruct.py`, add this function immediately above `_number_from_bookends`:

```python
def _agreeing_forward_slots(
    forward: list[int], backward: list[int]
) -> list[Optional[int]]:
    """Forward slots to keep when both bookends numbered the blink.

    Offset is the backward index minus the forward index. A blink stays
    when that offset equals the previous or next blink's offset. One blink
    with no neighbour stays. Comparison happens before any blink is removed,
    so a change of offset between two stretches keeps both boundary blinks.
    """
    deltas = [b - f for f, b in zip(forward, backward)]
    kept: list[Optional[int]] = []
    n = len(deltas)
    for i, slot in enumerate(forward):
        neighbours = []
        if i:
            neighbours.append(deltas[i - 1])
        if i + 1 < n:
            neighbours.append(deltas[i + 1])
        if not neighbours or any(deltas[i] == d for d in neighbours):
            kept.append(slot)
        else:
            kept.append(None)
    return kept
```

In `_number_from_bookends`, replace the `elif kind == "both":` block with:

```python
        elif kind == "both":
            forward = _assign_slots(times, period)
            backward = _backward_slots(times, period, forward[-1])
            slots_by_feed[fi] = _agreeing_forward_slots(forward, backward)
```

Do not change the opening-only branch, the closing-only branch, or `_add_feed_slots`.

- [ ] **Step 5: Run the new tests and the existing bookend tests**

Run:

```bash
"$PY" test_reconstruct.py TestBookends -v
```

Expected: PASS, including `test_steady_offset_keeps_the_forward_indexes`, `test_odd_blink_is_dropped_and_its_slot_stays`, `test_offset_change_between_stretches_keeps_both_boundaries`, `test_blinks_outside_both_cues_ignored_and_first_is_slot_zero`, and the closing-only and opening-only tests.

- [ ] **Step 6: Copy the script into the bundle the running server uses**

```bash
cp /workspaces/dlm/backend/internal/cvruntime/src/reconstruct.py \
  /workspaces/dlm/dist/cvruntime/linux_amd64/reconstruct.py
```

Do not run `./scripts/build-cvruntime.sh`.

- [ ] **Step 7: Commit**

```bash
git add docs/design/backend-lights-and-automation.md \
  backend/internal/cvruntime/src/reconstruct.py \
  backend/internal/cvruntime/src/test_reconstruct.py \
  dist/cvruntime/linux_amd64/reconstruct.py
git commit -m "$(cat <<'EOF'
fix: keep blinks when bookend counts differ by a steady offset

The backward count was anchored on a single rounding of the whole sweep, so every bulb failed the equality check and the scan reported no blinks.

EOF
)"
```
