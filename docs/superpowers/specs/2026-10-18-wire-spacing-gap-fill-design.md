# Wire spacing gap fill — design

**Date:** 2026-10-18
**Status:** Approved for planning
**Scope:** After triangulation, fill a short hole in a scanned light string and move a low-confidence light onto that string. Show those lights on the review screen. A run of four or more still blocks Confirm. This adds REQ-053. It does not change blink detection, bookends, camera pose, or triangulation.

## Context

A capture sweep numbers every bulb. Triangulation gives a 3D point only when a bulb is seen from two or more clips. Any id in the span with no point is listed in `missing`. Confirm refuses the save while `missing` is non-empty, because a model must be the contiguous run `0, 1, 2, …`.

The straight-line distance between two neighbouring bulbs is at most the wire between them. Slack or a bend makes that line shorter. The longest line between two confident neighbours is the closest observation to a stretched wire.

Coordinates are metres when the job used the printed markers or a scale hint. With neither, the scan treats the distance between the first two cameras as 1 m. The typed gap is used only in the fallback below, so a scan that already has a confident pair does not depend on that scale.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Where | A new step in the CV child, after unseen ids are marked missing and before the JSON result is written |
| Anchors | Confident triangulated lights only. Their coordinates stay as scanned. A low-confidence light is not an anchor. Guesses do not become anchors |
| Step length | The longest Euclidean distance among confident lights whose ids differ by 1 and whose distance is greater than 0. One such pair is enough |
| End direction | The step from the end anchor to its immediate confident neighbour. Length is the longest step, not that neighbour's own length |
| Interior hole of 1 to 3 | Evenly along the straight line between the two anchors, even when that spacing is shorter or longer than the longest step |
| End hole of 1 to 3 | Repeat the end direction, each hop the longest step long, when that end anchor's immediate neighbour is confident and that neighbour step has length greater than 0 |
| Run of 4 or more | Left whole. No partial fill |
| Low-confidence inside a filled hole | Its scanned coordinate is replaced. It is listed in `adjusted` and removed from `low_confidence` |
| Unseen light inside a filled hole | Listed in `guessed` and removed from `missing` |
| Fallback | Only when no confident adjacent pair has length greater than 0. A low-confidence light within 3 ids of a confident light moves along the scanned direction to `index gap × typed gap`. Missing lights are not invented from one anchor |
| Typed gap | Form field **Maximum gap (cm)**, prefilled with 10. Blank or omitted means 10 cm. Stored on the job spec as `max_gap_m` (default `0.10`) |
| Invalid gap | **400** `maximum gap must be a positive number of centimetres up to 100`. This includes 0, negative, above 100, and any non-finite value |
| Confirm | A non-empty `missing` list still refuses the save. `guessed` and `adjusted` do not |
| Persistence | The saved model stores `id, x, y, z` only. Guessed and adjusted are review labels |
| Requirements | New **REQ-053**. Do not renumber or reuse an existing code |

## Placement

Add `apply_wire_spacing` in `backend/internal/cvruntime/src/reconstruct.py`. Call it after the span loop that marks unseen ids missing, and before `_make_result`. Inputs are the triangulated points, `missing`, `low_confidence`, `span`, and `max_gap_m`. `span` is the same value that loop already used, so every id in `0 … span−1` is already a point or an entry in `missing`. This step places only those ids. It returns updated points, `missing`, `low_confidence`, `guessed`, and `adjusted`.

An anchor is a triangulated id that is not in `low_confidence`. When there are no anchors, skip the interior and the ends. The fallback has no anchor to measure from, so it changes nothing: `guessed` and `adjusted` stay empty.

A usable pair is two anchors whose ids differ by 1 and whose Euclidean distance is greater than 0. `max_step` is the maximum of those distances. When there is no usable pair, `max_step` is `max_gap_m`.

Run the three phases in this order. Later phases do not move an id that an earlier phase already placed.

**1. Interior.** Walk anchors in id order. For each successive pair of anchors, let `k` be the number of ids strictly between them. When `k` is 1, 2, or 3, place those ids evenly on the segment. Id `prev + j` goes to `prev_point + (j / (k + 1)) * (next_point - prev_point)` for `j = 1 … k`. An id that was low-confidence is `adjusted`. An id that was missing is `guessed`. When `k` is 4 or more, leave every id in that run as it was.

**2. Ends.** Let the first anchor be `s` and the last be `t`.

- The prefix is ids `0 … s−1`. When its length `k` is 1, 2, or 3, and id `s+1` is an anchor, and the vector `V = point(s+1) - point(s)` has length greater than 0, place id `s−j` at `point(s) - j * (max_step / |V|) * V` for `j = 1 … k`.
- The suffix is ids `t+1 … span−1`. When its length `k` is 1, 2, or 3, and id `t−1` is an anchor, and `V = point(t) - point(t−1)` has length greater than 0, place id `t+j` at `point(t) + j * (max_step / |V|) * V` for `j = 1 … k`.
- Otherwise leave that end as it was. A run of 4 or more at an end is left whole.

The same `adjusted` / `guessed` split as the interior applies.

**3. Fallback.** Run this phase only when there is no usable pair. For each id still in `low_confidence`, choose the anchor with the smallest id distance. When that distance is a tie, choose the lower id. When the distance is 1, 2, or 3, and the scanned point's Euclidean distance from the anchor is greater than 0, move the light from the anchor toward the scanned point so the distance is `id distance × max_gap_m`. List it in `adjusted` and remove it from `low_confidence`. Leave every other low-confidence light, and leave every missing light, unchanged.

`guessed` and `adjusted` are sorted ascending. Each id is in at most one of `missing`, `low_confidence`, `guessed`, and `adjusted`. A guessed or adjusted id is present in `lights`. `light_count` stays the span already computed. It does not become the number of filled positions.

Worked end. Anchors at id 2 `(0, 0, 0)`, id 3 `(0.08, 0, 0)`, and id 4 `(0.18, 0, 0)`. Ids 0 and 1 are missing. The neighbour step at the start is 8 cm along +X. The longest step is 10 cm. Id 1 is guessed at `(-0.10, 0, 0)`. Id 0 is guessed at `(-0.20, 0, 0)`.

Worked interior. Anchors at id 3 `(0.08, 0, 0)` and id 5 `(0.24, 0, 0)`. Id 4 was low-confidence. It is adjusted to `(0.16, 0, 0)` and removed from `low_confidence`. The longest step is not used, because both ends of the hole are anchors.

## Job spec and result

The upload form field is `max_gap_cm`. The page prefills `10`. A blank box omits the field. Go trims the value. An absent or blank value becomes `max_gap_m = 0.10` on the CV job spec. A present value must be a finite number `> 0` and `≤ 100`. Go divides a valid value by 100 to get metres. `100` becomes `1.0`. The child receives `max_gap_m` already set. When the key is absent, the child uses `0.10`. When the key is present and not a positive finite number, the child fails the job with `max_gap_m must be a positive finite number`. The 100 cm cap is enforced on the HTTP boundary only.

`cvruntime.Result` and the JSON result gain:

- `guessed`: array of ids, sorted, `[]` when empty
- `adjusted`: array of ids, sorted, `[]` when empty

Both keys are always present on success and on failure. They are never JSON `null`. The failed-result helper includes them as empty arrays. The TypeScript `CaptureJobResult` in `web/lib/models.ts` gains the same two required arrays.

## Review and confirm

On the create-from-video form, the field label is **Maximum gap (cm)**. The helper text is: "This number is used only when the scan cannot measure a step from lights it is sure about."

The review screen adds two lines under the existing missing and low-confidence lines:

- `Guessed IDs: …` and "These positions were filled in along the string. They were not seen in the videos."
- `Adjusted IDs: …` and "These were seen, then moved onto the usual step."

When `missing` is non-empty, the review also says: "Confirm will not save a model while any light is missing." The Confirm button stays available. The server still returns the existing missing-ids error.

The green line "All lights detected with high confidence." appears only when `missing`, `low_confidence`, `guessed`, and `adjusted` are all empty.

The read-only preview draws the coordinates in `lights`, including guessed and adjusted lights, with the same spheres as today. No separate colour for a guessed or adjusted light.

Confirm copies `lights` into the model. It does not store `guessed` or `adjusted`.

## Documents

In `docs/requirements/requirements.md` §11, replace "Lights it couldn't place are honestly reported, never made up." with:

> A hole of up to three lights is filled in along the string and named before you save. A longer hole is still reported, and the model is not saved while any light is missing. A light the scan was unsure about can be moved onto the usual spacing and is named as adjusted.

Append index row **REQ-053**, "Short holes in a scanned model are filled and named", pointing at §11. Keep the code out of the prose.

In `docs/requirements/acceptance-criteria.md`, under "Building a model from video", add two checks:

- A scan that missed three lights in a row between lights it saw shows those ids as guessed, and Confirm can save the model.
- A scan that missed four lights in a row shows those ids as missing, says Confirm will not save while a light is missing, and the save is refused.

In `docs/userguide/build-model-from-video.md`, add the maximum-gap box and the guessed and adjusted review lines. No `REQ-*` codes. The box is prefilled with 10 and the guide says it is used only when the scan cannot measure a step from lights it is sure about. A guessed light was filled in along the string. An adjusted light was seen, then moved onto the usual step.

In `docs/design/backend-lights-and-automation.md` §3.23, replace the sentence that begins `The pipeline never fabricates coordinates` with:

> A hole of one, two, or three lights is filled by §3.23.4, and a low-confidence light inside that hole is moved onto the filled line. A longer hole stays in `missing`. The review step (§4.17) shows guessed, adjusted, missing, and low-confidence ids so the user can re-record before confirming.

Add a spacing node to the existing pipeline diagram between triangulation and the result.

Add **§3.23.4 Wire spacing** with an "In plain terms:" opening and the placement rules in this document. Do not renumber §3.23 or §3.23.1–§3.23.3. Extend the §3.15–§3.23.3 row in `docs/design/architecture.md` to §3.23.4.

In `docs/design/backend-service.md`, add optional `max_gap_cm` to `POST /api/v1/models/capture`, and add `guessed` and `adjusted` to the capture-job result.

In `docs/design/frontend.md` §4.17, add the field, the two review lines, the confirm-refusal sentence, and the rule for the green line.

In `docs/design/overview.md`, add a REQ-053 row citing §3.23.4 and §4.17. In the REQ-048 row, replace `undetected lights are reported, never fabricated` with `a hole of up to three lights is filled by REQ-053, and a longer hole is reported and blocks confirm`. In the REQ-049 row, include guessed and adjusted in the review. Add a REQ-053 bullet to `docs/design/appendix-traceability.md`.

## Tests

Python tests in `backend/internal/cvruntime/src/test_reconstruct.py` call `apply_wire_spacing` with positions. They do not decode video.

- The worked end above guesses ids 0 and 1 at `(-0.10, 0, 0)` and `(-0.20, 0, 0)`.
- The worked interior above adjusts id 4 to `(0.16, 0, 0)` and leaves `low_confidence` without id 4.
- Three missing ids between two anchors are placed at 1/4, 2/4, and 3/4 of the segment and listed in `guessed`.
- Four missing ids between two anchors stay in `missing`. `guessed` has none of them.
- A low-confidence light inside a run of four keeps its scanned coordinate and stays in `low_confidence`.
- An end anchor whose next id is not confident leaves that end missing.
- Two confident neighbours on the same point do not grow that end.
- With one confident light and no usable pair, a low-confidence light two ids away moves to `2 × max_gap_m` along the scanned direction. A missing id beside that light stays missing.
- A low-confidence light on the same point as the confident light stays in `low_confidence`.
- With no confident light, `guessed` and `adjusted` are empty.
- The four id lists are disjoint and sorted. Compare coordinates within `1e-6` m.

Go tests:

- A blank or omitted `max_gap_cm` is stored as `0.10`. `100` is stored as `1.0`. `0`, a negative number, `101`, and a non-number return **400** with `maximum gap must be a positive number of centimetres up to 100`.
- Confirm saves a succeeded job whose holes were guessed and whose `missing` list is empty.
- Confirm still returns the missing-ids error when `missing` is non-empty, including when other ids were guessed.

Web tests in `web/lib/models.test.ts`: `createCaptureJob` sends `max_gap_cm` when the caller passes it, and omits the key when the caller does not. The review sentences are checked in the browser when the page is built. This repo has no component test harness.

## Out of scope

- Fitting a curve through the lights.
- Accepting or rejecting each guessed light on its own.
- Storing guessed or adjusted on the saved model, or colouring those lights in the preview.
- Filling a run of four or more, including filling the first three of a longer run.
- Using the typed gap to rescale a scan that already has a confident pair.
- Blink detection, bookends, pose estimation, triangulation, and the wording of the missing-ids confirm error.
