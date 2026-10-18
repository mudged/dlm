# Bookend slot offset — design

**Date:** 2026-10-18
**Status:** Approved for planning
**Scope:** When a clip has both the opening and the closing colour flash, keep blinks whose forward and backward indexes differ by a steady offset. Drop only a blink whose offset matches neither neighbour. This corrects numbering inside `reconstruct.py`. It does not add a requirement, change the Go server, or change the failure text.

## Context

A capture with both red–blue–green flashes numbers each blink twice. The forward count starts at index 0 on the first blink after the opening flash and adds `round(gap / period)` for each later gap. The backward count walks from the last blink. A blink is kept only when the two indexes are equal.

The backward count currently starts at `round(sweep length / period)`. The forward count ends on the sum of the per-gap roundings. On the three clips in `/workspaces/dlm/captures` (`20261017_180459.mp4`, `20261017_180846.mp4`, `20261017_181158.mp4`) those totals differ by 4 to 6 slots, so every blink is discarded. The job then reports `no blink events detected in any feed`.

Those clips do contain both flashes and a sweep of dwell-length blinks. The failure is the equality test, not missing footage.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Which clips | Only a clip that already has both an opening and a closing flash (`kind == "both"`). Opening-only and closing-only numbering stay as they are |
| Backward anchor | The last blink's backward index is the slot the forward count assigned to that same blink (`forward[-1]`). It is not `round(sweep length / period)` |
| Offset | `backward index − forward index` for that blink |
| Keep | A blink whose offset equals the offset of the previous blink or the next blink, kept at the **forward** index. The opening flash stays index 0 |
| Drop | A blink whose offset differs from every adjacent blink. One blink with no neighbour is kept |
| Two stretches | A change of offset between two stretches does not drop the blinks at the boundary. Each of those blinks matches the other blinks in its own stretch |
| Same forward slot twice | Unchanged. After this filter, a slot that still has two blinks from one clip is ambiguous and both are dropped |
| `light_count` | Unchanged. Indexes at or above `light_count` are discarded after this filter. `light_count` does not renumber a clip that has the opening flash |
| Failure text | Unchanged. `no blink events detected in any feed` remains the message when no feed contributes a light |
| Requirements | No new `REQ-*`. This is a correction of the both-anchors rule already in §3.23.3 |

## Numbering

In `_number_from_bookends`, for a both-anchors span:

1. `forward = _assign_slots(times, period)`.
2. `backward = _backward_slots(times, period, forward[-1])`. When `period <= 0`, `_assign_slots` is already the positional fallback and `forward[-1]` is `len(times) - 1`.
3. Compare each blink's offset with the previous and next blink in time order, before any blink is removed.
4. Pass the surviving forward indexes to the existing `_add_feed_slots` path, which still drops a slot that one feed filled twice.

Worked examples. `period` is the median gap, as `_estimate_period` already computes it. Cues sit outside the blinks, so the span is the blink list.

**Steady sweep the clock rounds short.** Blinks at 0, 0.6, 1.2, 1.8, 2.8, 3.8, 4.8, 5.8, 6.8 seconds. Median gap is 1 s. The forward count ends at slot 8. `round(6.8 / 1)` is 7, so today's equality test keeps nothing. After this change the offsets agree and the clip contributes indexes 0 through 8.

**One extra blink on top of a real one.** Blinks at 0, 1.5, 1.9, 2.3, 3.8 seconds. Median gap is 0.95 s. Forward slots are 0, 2, 2, 3, 5. The blink at 1.9 s is the only one whose offset differs from both neighbours, so it is discarded. Slot 2 is left with one blink. The clip contributes indexes 0, 2, 3, and 5.

## Design doc

In `docs/design/backend-lights-and-automation.md` §3.23.3, replace the sentence that begins `When both anchors exist` through `Leave the surrounding indexes as they are.` with:

> When both anchors exist, number each blink forward from the opening signal (index 0) and backward from the closing signal. The backward count ends on the same slot the forward count gave the last blink. The offset is the backward index minus the forward index. A blink whose offset equals the offset of the previous blink or the next blink is kept at the forward index. A blink whose offset differs from both of those, or from its only neighbour, is discarded. One blink alone is kept. Where the offset changes between two stretches, both blinks at that change stay, because each matches the rest of its own stretch. Leave the other indexes as they are.

Leave the rest of that paragraph, including the `light_count` sentences, unchanged. Do not renumber §3.23.3. Do not edit `docs/requirements/` or `docs/userguide/`.

## Tests

Add tests next to the existing bookend numbering tests in `backend/internal/cvruntime/src/test_reconstruct.py`. They call `number_from_bookends` with one feed, no `light_count`, and two cues that bracket the blinks.

- The 0 … 6.8 s sweep above contributes indexes 0 through 8 and rejects nothing.
- The 0, 1.5, 1.9, 2.3, 3.8 s sweep contributes indexes 0, 2, 3, and 5 and rejects nothing.
- `test_blinks_outside_both_cues_ignored_and_first_is_slot_zero` and the closing-only / opening-only tests stay as they are.

The three phone clips are not fixtures. They are slow to decode and stay out of the unit tests.

## Out of scope

- Blink detection, the room mask, cue matching, and dwell limits.
- The wording of `no blink events detected in any feed`.
- Pose estimation and triangulation.
