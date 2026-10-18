# Wider hole fill — design

**Date:** 2026-10-18
**Status:** Pending spec review
**Scope:** Amend REQ-053. Between two lights the scan saw, fill a run of up to five missing lights from their coordinates, including when the scan was not sure about those lights. Leave a seen light where it was scanned. Keep the start and end rules at a run of three from lights the scan is sure about. Do not add a requirement code. Do not change blink detection, bookends, camera pose, triangulation, or the Confirm error.

## Context

`apply_wire_spacing` runs after unseen ids are marked missing and before the result JSON is written. Today an anchor is a triangulated light that is not low-confidence. A run of one, two, or three ids between those anchors is filled. A low-confidence light in that run is moved onto the line and listed in `adjusted`. A separate fallback slides a low-confidence light toward one sure light when no sure neighbours have a positive step. Confirm still refuses the save while `missing` is not empty.

On a real chair scan, 105 of 111 detected lights were low-confidence, so only two lights counted as anchors. They were 21 ids apart. Short holes between lights the scan had seen were left missing, and Confirm refused the save.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Where | The same `apply_wire_spacing` call. No new pipeline stage |
| Seen anchor | Every light that already has a coordinate when this step starts, including a low-confidence light. Its coordinate is never written over. A guessed id is not an anchor |
| Interior hole of 1 to 5 | Evenly along the straight line between the two seen anchors. Every placed id was missing and is listed in `guessed` |
| Interior run of 6 or more | Left whole. No partial fill |
| Sure anchor | A triangulated light that is not low-confidence. Used for the ends and for the longest sure step |
| End hole of 1 to 3 | Only before the first seen light or after the last. Only when that end light and the neighbouring id are both sure, and that neighbour step is longer than 0. Each hop is the longest sure step. A low-confidence light at that end blocks it |
| End run of 4 or more | Left whole |
| Low-confidence light | Stays in `low_confidence` at its scanned coordinate. It is not moved onto a line and it is not slid toward a sure light |
| `adjusted` | Always `[]`. The key stays on the result |
| Typed gap | The Maximum gap box, Go validation, and `max_gap_m` on the job spec stay. `apply_wire_spacing` still accepts `max_gap_m` and does not read it |
| Confirm | A non-empty `missing` list still refuses the save. `guessed` does not |
| Persistence | The saved model stores `id, x, y, z` only |
| Requirements | Amend **REQ-053** in place. Do not append a new code |

## Placement

Inputs stay the triangulated points, `missing`, `low_confidence`, `span`, and `max_gap_m`. `span` is the value the missing-id loop already used, so every id in `0 … span−1` is already a point or an entry in `missing`. This step places only missing ids in that span. It returns updated points, `missing`, `low_confidence`, `guessed`, and `adjusted`. The caller's point dict is not mutated.

Run the interior first, then the ends. The end pass places only an id that is still missing. A seen light is never passed to the placement.

**1. Interior.** Walk seen anchors in id order. For each successive pair, let `k` be the number of ids strictly between them. Those ids are missing. When `k` is 1, 2, 3, 4, or 5, place them evenly on the segment. Id `prev + j` goes to `prev_point + (j / (k + 1)) * (next_point - prev_point)` for `j = 1 … k`. List each placed id in `guessed` and remove it from `missing`. When `k` is 6 or more, leave every id in that run missing. The longest sure step is not used.

**2. Ends.** Let the first seen light be `s` and the last be `t`.

- The prefix is ids `0 … s−1`. When `s` is sure, id `s+1` is sure, `k = s` is 1, 2, or 3, and `V = point(s+1) - point(s)` has length greater than 0, place id `s−j` at `point(s) - j * (max_step / |V|) * V` for `j = 1 … k`.
- The suffix is ids `t+1 … span−1`. When `t` is sure, id `t−1` is sure, `k = span−1−t` is 1, 2, or 3, and `V = point(t) - point(t−1)` has length greater than 0, place id `t+j` at `point(t) + j * (max_step / |V|) * V` for `j = 1 … k`.
- Otherwise leave that end missing. A low-confidence light at the start or end does not grow the string. A run of 4 or more at an end is left whole.

`max_step` is the longest Euclidean distance among sure lights whose ids differ by 1 and whose distance is greater than 0. A low-confidence light is not part of any pair that sets `max_step`. The neighbour's own length is not the hop length when a longer sure step exists. `max_gap_m` is not read. When the end rule runs, a sure neighbour step of length greater than 0 already exists, so `max_step` is defined.

There is no third phase. Delete the fallback that slides a low-confidence light toward one sure light.

`guessed` is sorted ascending. `adjusted` is empty. `low_confidence` is unchanged. Each id is in at most one of `missing`, `low_confidence`, `guessed`, and `adjusted`. A guessed id is present in `lights`. `light_count` stays the span. It does not become the number of filled positions.

Worked interior of five. Seen lights at id 0 `(0, 0, 0)` and id 6 `(0.60, 0, 0)`. Id 6 may be low-confidence. Ids 1 through 5 are guessed at `(0.10, 0, 0)`, `(0.20, 0, 0)`, `(0.30, 0, 0)`, `(0.40, 0, 0)`, and `(0.50, 0, 0)`. Id 6 stays at its scanned coordinate and stays in `low_confidence` when it started there.

Worked interior of six. Seen lights at id 0 and id 7. Ids 1 through 6 stay in `missing`. `guessed` has none of them.

Worked split. Sure light at id 0 `(0, 0, 0)`, low-confidence light at id 2 `(1, 1, 1)`, sure light at id 5 `(0.5, 0, 0)`. Id 1 is guessed at `(0.5, 0.5, 0.5)`. Id 3 is guessed at `(5/6, 2/3, 2/3)`. Id 4 is guessed at `(2/3, 1/3, 1/3)`. Id 2 stays at `(1, 1, 1)` and stays in `low_confidence`.

Worked end of three. Sure lights at id 3 `(0, 0, 0)` and id 4 `(0.10, 0, 0)`. Ids 0, 1, and 2 are missing. The longest sure step is 10 cm. Id 2 is guessed at `(-0.10, 0, 0)`, id 1 at `(-0.20, 0, 0)`, and id 0 at `(-0.30, 0, 0)`.

Worked end of four. Sure lights at id 4 and id 5, with ids 0 through 3 missing. All four stay in `missing`.

Worked low-confidence end. Id 1 is low-confidence at `(0, 0, 0)`. Ids 2 and 3 are sure. Id 0 is missing. Id 0 stays missing. Id 1 is not moved.

## Review, requirements, and result

`guessed` and `adjusted` stay on the capture result as arrays, on success and on failure. `adjusted` is always empty, never JSON `null`. The failed-result helper already emits both as empty arrays. The TypeScript `CaptureJobResult` already has both required arrays. No new JSON field. Go validation of `max_gap_cm` is unchanged.

The review page keeps the guessed lines: "Guessed IDs: …" and "These positions were filled in along the string. They were not seen in the videos." The Adjusted block stays in the page and renders only when `adjusted` is non-empty, so it no longer appears. Low-confidence ids stay listed as low-confidence. The preview still draws every entry in `lights`, with no separate colour for a guessed light. The green line "All lights detected with high confidence." still requires `missing`, `low_confidence`, `guessed`, and `adjusted` to be empty. When `missing` is non-empty, the page still says Confirm will not save while any light is missing. Confirm still copies `lights` only.

The only page-copy change is the Maximum gap helper. Replace "This number is used only when the scan cannot measure a step from lights it is sure about." with "Filling a missing light does not use this number." The field stays, still prefilled with 10.

In `docs/requirements/requirements.md` §11, replace the hole sentences with:

> A hole of up to five lights between lights the scan saw is filled in along the string and named before you save. A longer hole is still reported, and the model is not saved while any light is missing. At the very start or end, only a hole of up to three is filled, and only from lights the scan is sure about.

Drop the sentence that says an unsure light is moved onto the usual spacing. Leave the REQ-053 index row and its section pointer. Keep the code out of the prose.

In `docs/requirements/acceptance-criteria.md`, under "Building a model from video":

- Keep the check that three missed lights in a row between lights the scan saw are named as guessed and Confirm can save.
- Add a check that five missed lights in a row between lights the scan saw are named as guessed and Confirm can save.
- Replace the four-in-a-row check with a six-in-a-row check: those ids are named as missing, the page says Confirm will not save while a light is missing, and the save is refused.

In `docs/userguide/build-model-from-video.md`, drop the adjusted bullet. After the review list, add:

> A hole of up to five lights between lights the scan saw is filled in along the string. A longer hole is left missing, and the model is not saved while any light is missing. At the very start or end, only a hole of up to three is filled, and only from lights the scan is sure about.

Replace the Maximum gap paragraph with: the box is prefilled with 10, and filling a missing light does not use that number. No `REQ-*` codes.

In `docs/design/backend-lights-and-automation.md`, replace the reporting sentence that begins `A hole of one, two, or three lights` and the gap-fill node in the pipeline diagram. In §3.23.4, replace the "In plain terms:" opening and the bullets so they match this document. Keep the `max_gap_cm` validation fact, and state that the fill step does not read `max_gap_m`. Delete the fallback bullet. Do not renumber §3.23 or §3.23.4.

In `docs/design/frontend.md` §4.17, update the Maximum gap helper and stop describing `adjusted` ids as lights that were moved. `adjusted` remains in the listed result fields and is always empty.

In `docs/design/overview.md`, update the REQ-048, REQ-049, and REQ-053 rows to the five-between-seen-lights rule, the three-at-the-ends rule, and an always-empty `adjusted` list. In `docs/design/appendix-traceability.md`, update the REQ-053 bullet the same way. `docs/design/backend-service.md` stays, because the JSON keys do not change.

## Tests

Python tests in `backend/internal/cvruntime/src/test_reconstruct.py` call `apply_wire_spacing` with positions. They do not decode video. Compare coordinates within `1e-6` m.

These existing cases stay:

- Ids 0 and 1 missing, sure lights at id 2 `(0, 0, 0)`, id 3 `(0.08, 0, 0)`, and id 4 `(0.18, 0, 0)`. Ids 0 and 1 are guessed at `(-0.20, 0, 0)` and `(-0.10, 0, 0)`. The hop is the longest sure step, 10 cm, not the 8 cm neighbour step.
- Sure lights at id 0 `(0, 0, 0)`, id 1 `(0.10, 0, 0)`, and id 2 `(0.18, 0, 0)`, with id 3 missing. Id 3 is guessed at `(0.28, 0, 0)`. The hop is the 10 cm sure step, not the 8 cm neighbour step.
- One sure light and no sure neighbour leaves the end missing.
- Two sure lights on the same point do not grow that end.
- Three missing ids between two seen lights are placed at 1/4, 2/4, and 3/4 of the segment.
- An interior hole is filled on the segment even when that segment is longer than the longest sure step.
- A low-confidence light on the same point as a sure light stays in `low_confidence`.
- With no seen light, `guessed` and `adjusted` are empty.
- The caller's point dict is not mutated.

These existing cases change:

- Four missing ids between id 0 `(0, 0, 0)` and id 5 `(0.50, 0, 0)` are guessed at `(0.10, 0, 0)`, `(0.20, 0, 0)`, `(0.30, 0, 0)`, and `(0.40, 0, 0)`.
- A low-confidence light at id 4 `(9, 9, 9)` between sure lights at id 3 `(0.08, 0, 0)` and id 5 `(0.24, 0, 0)` stays at `(9, 9, 9)` and stays in `low_confidence`. `adjusted` is empty.
- The worked split above. The low-confidence light keeps `(1, 1, 1)`. The holes on either side are guessed.
- Sure light at id 2 `(0, 0, 0)`, low-confidence light at id 4 `(0, 0.50, 0)`, missing id 3. Id 3 is guessed at `(0, 0.25, 0)`. Id 4 stays at `(0, 0.50, 0)` and stays in `low_confidence`.

New cases:

- The worked interior of five, including when id 6 is low-confidence.
- The worked interior of six.
- The worked end of three and the worked end of four. The same bounds apply to a suffix.
- The worked low-confidence end, and the mirror at the last seen light.
- Ids 0 and 1 missing, sure lights at id 2 `(0, 0, 0)` and id 3 `(0.08, 0, 0)`, and a low-confidence light at id 4 `(10, 0, 0)`. Ids 0 and 1 are guessed at `(-0.16, 0, 0)` and `(-0.08, 0, 0)`. Id 4 stays at `(10, 0, 0)`.
- The four id lists are disjoint and sorted.

Go tests already save a succeeded job whose `missing` list is empty and reject one whose `missing` list is not. Do not add Go tests. Do not change the web test for `max_gap_cm`. The helper sentence is the only page change; check it in the browser. This repo has no component test harness.

## Out of scope

- Dropping a disagreeing camera view, or otherwise raising how many lights count as sure.
- Removing the Maximum gap box, its validation, or `max_gap_m` from the job spec.
- Filling a run of six or more, including filling the first five of a longer run.
- Extrapolating an end from a low-confidence light.
- Moving a low-confidence light.
- Storing guessed or adjusted on the saved model, or colouring guessed lights in the preview.
- The wording of the missing-ids Confirm error.
- Blink detection, bookends, pose estimation, and triangulation.
