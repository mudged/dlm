# Capture bookend signal — remaining tasks

Tasks 1–3 are complete on main (`51966e7..a1e7cdd`). This file restores Tasks 4–6 after the original plan file was removed from disk. The approved spec is `docs/superpowers/specs/2026-09-29-capture-bookend-signal-design.md`.

Controller decision: do not copy the `align_detections` loop. Extract one shared helper that `align_detections` and the bookend path both call. `align_detections`'s signature stays the same so `TestCrossFeedCorrespondence` keeps passing.

### Task 4: Number clips from the bookend

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py`
- Modify: `backend/internal/cvruntime/src/test_reconstruct.py`
- Modify: `backend/internal/cvruntime/src/gen_fixtures.py`

These three files already contain unrelated uncommitted edits (clearer "no blinks" errors, and similar). Do not revert them. Build the bookend on top.

**Requirements (verbatim from the approved spec):**

Cue search runs before blink numbering. A cue is a red pulse, then a blue pulse, then a green pulse, in that order. Each pulse lasts 100–350 ms. The dark gap between pulses lasts 80–450 ms. A pulse's colour is the channel among the bright pixels that is at least 1.5× each of the other two. Time ranges that belong to a cue are excluded from blink detection.

| Cues in the clip | Numbering |
|------------------|-----------|
| Opening and closing | Blinks before the opening green turns off, and blinks after the closing red turns on, are ignored. The first blink after the opening signal is light 0. Cadence still skips a gap when a bulb is missed. |
| Closing only | The last blink before the closing red is the last bulb. With `light_count`, that blink is index `light_count − 1`. Without `light_count`, copy the last index from another clip in the job that has an opening signal. If no clip has an opening signal, drop this clip. Reason must say the opening flash is missing and the light count is needed to count backward. |
| Opening only | Number forward from light 0 after the opening signal. |
| None | Drop the clip. Reason: `no start or end signal found`. |
| Three or more, or one cue with dwell-length blinks on both sides | Drop the clip as ambiguous. |

When both anchors exist, number forward from the opening signal and backward from the closing signal, where the closing signal marks the last slot of that same span. Discard a blink that receives two different indexes. `light_count` does not renumber a clip that already has the opening signal. Indexes `>= light_count` are discarded. Unseen ids in `0 … light_count−1` go in `missing`.

`light_count` in the job spec, when present, is an integer `1 … 1000`. An invalid value fails the job with `light_count must be a whole number from 1 to 1000`. Absent means unknown.

Fewer than two usable clips fails the job. `error` is a plain sentence that the recordings missed the start and end flashes, and it names each dropped file and reason. The result of a successful job includes `rejected_feeds: [{ "file", "reason" }]` using the upload base name. `_make_error` includes `"rejected_feeds": []`.

`gen_fixtures.py` paints both bookends (200 ms red, 200 ms dark, 200 ms blue, 200 ms dark, 200 ms green) after the leading gap and again after a 500 ms dark settle following the last bulb, so existing pipeline tests still number from the opening flash.

**Tests to add in `test_reconstruct.py` (`TestBookends`):**
- `find_cues` finds one red-blue-green cue in a synthetic 30 fps AVI (6 frames per colour, 6 black frames between).
- A dwell-length blink before the opening cue is ignored; blinks between the two cues are kept and start at slot 0.
- A clip with only the closing cue and `light_count=5` and three blinks starts at slot 2.
- A clip with no cue is rejected with reason containing `no start or end signal`.
- A one-second white flash before the opening cue is ignored when `dwell_ms=1000`.

**Commit only:**
- `backend/internal/cvruntime/src/reconstruct.py`
- `backend/internal/cvruntime/src/test_reconstruct.py`
- `backend/internal/cvruntime/src/gen_fixtures.py`

Message:

```
Number capture clips from the red-blue-green bookend.

A late recording can count backward from the closing flash, and a clip with no signal is dropped instead of shifting every light.
```

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py -v`

### Task 5: Pass light count and rejected clips through the API

**Files:**
- Modify: `backend/internal/cvruntime/contract.go`
- Modify: `backend/internal/reconstruct/reconstruct.go`
- Modify: `backend/internal/httpapi/capture_models.go`
- Modify: `backend/internal/httpapi/capture_models_test.go`

`capture_models.go` already has unrelated uncommitted edits. Do not revert them.

**Produces:**
- `cvruntime.RejectedFeed` with `file` and `reason`
- `Result.RejectedFeeds []RejectedFeed \`json:"rejected_feeds"\`` (no omitempty)
- `JobSpec.LightCount *int \`json:"light_count,omitempty"\``
- `reconstruct.CreateParams.LightCount *int` copied onto the job spec in `Create`

`POST /api/v1/models/capture` reads optional form field `light_count`. Empty is allowed. `0`, `nope`, and `1001` return **400** `bad_request` with message `light_count must be a whole number from 1 to 1000`, and Create is not called. `12` is forwarded.

`fakeReconstructCtrl.last` is a `reconstruct.CreateParams`. `multipartBody` in `capture_models_test.go` already accepts extra form fields.

**Commit only those four Go files.** Message:

```
Accept a capture light count and return clips the bookend rejected.

A closing-only recording can be numbered from the device light count, and the review can name clips that missed the signal.
```

Run: `cd backend && go test ./internal/httpapi/ ./internal/reconstruct/ ./internal/cvruntime/ -count=1 -timeout 180s`

### Task 6: Create-from-video light count and rejected clips

**Files:**
- Modify: `web/lib/models.ts`
- Modify: `web/lib/models.test.ts`
- Modify: `web/app/models/new/NewModelClient.tsx`

`web/lib/models.ts` already has unrelated uncommitted edits. Do not revert them.

`createCaptureJob` accepts optional `light_count`. When it is an integer from 1 to 1000, set form field `light_count`. When omitted, the form has no `light_count` key.

`CaptureJobResult` includes `rejected_feeds?: { file: string; reason: string }[]`. `getCaptureJob` preserves them.

On Create from video, add input `id="capture-light-count"` labelled `Light count (optional)` with helper text: `Leave this empty when every clip includes the opening red, blue, and green flash. If a clip only caught the closing flash, enter the same light count you set on the device.`

Empty submits. `0` and `abc` set `uploadError` to `Light count must be a whole number from 1 to 1000.` and do not call `createCaptureJob`.

On review, when `rejected_feeds` is non-empty, show heading `Clips set aside` and a list `{file}: {reason}`. A failed job still shows `error.message` and no Confirm button.

**Commit only those three web files.** Message:

```
Ask for a light count when a capture clip missed the opening flash.

The review names any clip the bookend signal could not use, so those lights are not guessed.
```

Run: `cd web && npm test -- lib/models.test.ts && npm run lint`
