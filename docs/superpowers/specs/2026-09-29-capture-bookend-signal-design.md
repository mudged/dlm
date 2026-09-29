# Capture bookend signal — design

**Date:** 2026-09-29  
**Status:** Approved 2026-09-29  
**Scope:** A red–blue–green flash at both ends of the device capture sweep, and video analysis that uses those flashes to number lights. Normative product text lives in the requirements, design, and user guide (updated with this spec).

## Context

Reconstruction numbers each bulb from the order of one-second blinks in the uploaded clips. Each clip treats its first accepted blink as light 0, then uses the one-second rhythm to skip gaps. That stays correct when every clip contains the sweep from the first bulb and nothing else in the shot lasts about a second.

Two ordinary recordings break it:

- A phone starts late, so its first blink is a later bulb. Another phone that caught the start still calls that moment an earlier index. The two views are paired to the wrong bulbs.
- A one-second bright event before the sweep becomes light 0, and every later index shifts with it.

Cameras and the server do not share a clock. The signal has to be inside the light show, visible in each clip on its own.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Signal | All lights flash red, then blue, then green, before light 0 and again after the last light |
| Pulse | 200 ms on, 200 ms dark between colours. Colours are full-brightness WLED RGB: red `(255,0,0)`, blue `(0,0,255)`, green `(0,255,0)` |
| Settle | 500 ms all-off after the opening green, and 500 ms all-off after the last bulb, before the closing signal |
| Sweep body | Unchanged: one white bulb at a time for `DLM_CAPTURE_DWELL_MS` (default 1000 ms), index order `0 … n−1` |
| Short dwell | Start is rejected with **422** `capture_dwell_too_short` when the configured dwell is under 500 ms |
| Status | Running status gains `phase`: `preamble`, `sweep`, or `postamble`. `current_index` is present only during `sweep` |
| Late clip | An ending signal alone numbers backward. Absolute indexes need the device light count, sent as optional `light_count` on the upload |
| No signal | That clip is dropped and named in the result. Fewer than two usable clips fails the job |
| Exposure | Pulses stay at 200 ms so a long all-lights flash does not blow out the following bulbs |

## Light show

For a device with `n` lights (`n ≥ 1`) and dwell `D` ms (`D ≥ 500`):

1. **Preamble.** For red, then blue, then green: paint every LED that colour for 200 ms, then all off for 200 ms. After green, stay all off for 500 ms (this 500 ms replaces an extra trailing gap; the 200 ms gaps are only between the three colours).
2. **Sweep.** For `k = 0 … n−1`: only LED `k` white `(255,255,255)` at full brightness for `D` ms, every other LED off. One WLED frame per bulb.
3. **Settle.** All off for 500 ms.
4. **Postamble.** Red, then blue, then green: 200 ms on, 200 ms all-off between those colours. After green, all off. The 500 ms settle is the darkness before this flash.
5. **Done.** All off, status returns to idle.

Stop at any point, including during a colour flash, writes all off and sends no further sweep frames. The strip is dark within 2 seconds (the existing stop bound).

`internal/capture` plays this sequence. A new driver method paints every LED one RGB colour in a single WLED state frame (the same shape as today's single-LED and all-off frames). Marker timings are fixed constants. Only the bulb dwell stays on `DLM_CAPTURE_DWELL_MS`.

## Status the device page polls

`GET /api/v1/devices/{id}/capture`, and the start response, use:

| Moment | Body |
|--------|------|
| Opening flashes | `{ "state":"running", "light_count", "phase":"preamble" }` |
| Bulb `k` | `{ "state":"running", "light_count", "phase":"sweep", "current_index": k }` |
| Closing flashes | `{ "state":"running", "light_count", "phase":"postamble" }` |
| Idle | `{ "state":"idle", "light_count" }` |

`current_index` is omitted outside `sweep`. The device page copy is:

- `preamble` — Starting — red, blue, green flash
- `sweep` — Lighting `current_index + 1` / `light_count`
- `postamble` — Finishing — red, blue, green flash. Keep recording until this ends.

The page still tells the operator to start recording before pressing Start capture.

Start also returns **422** `capture_dwell_too_short` when dwell is under 500 ms, alongside the existing `capture_no_lights` and `capture_conflict` responses.

## How a clip is numbered

Cue search runs before blink numbering. A cue is a red pulse, then a blue pulse, then a green pulse, in that order. Each pulse lasts 100–350 ms. The dark gap between pulses lasts 80–450 ms. A pulse's colour is the channel among the bright pixels that is at least 1.5× each of the other two. Time ranges that belong to a cue are excluded from blink detection, so a 200 ms flash cannot also count as a bulb.

| Cues in the clip | Numbering |
|------------------|-----------|
| Opening and closing | Blinks before the opening green turns off, and blinks after the closing red turns on, are ignored. The first blink after the opening signal is light 0. Cadence still skips a gap when a bulb is missed. |
| Closing only | The last one-second blink before the closing red is the last bulb. Counting backward needs `light_count` (below). |
| Opening only | Number forward from light 0 after the opening signal. Bulbs after the recording stopped are absent. |
| None | Drop the clip. |
| Three or more, or one cue with dwell-length blinks on both sides | Drop the clip as ambiguous. |

When both anchors exist, each blink is numbered forward from the opening signal (light 0) and backward from the closing signal, and the closing signal marks the last slot of that same span. A blink that receives two different indexes is discarded. The bulbs around it keep their indexes. A provided `light_count` does not renumber a clip that already contains the opening signal. Indexes at or above `light_count` are discarded, and unseen ids below `light_count` are listed in `missing`.

### Light count on upload

`POST /api/v1/models/capture` accepts an optional `light_count` field (integer `1 … 1000`). A missing field is allowed. A present but invalid value is **400**.

- With `light_count`, a closing-only clip uses `light_count − 1` as the index of the last blink before the closing signal. Indexes outside `0 … light_count−1` are discarded. Ids in that range with no point are returned in `missing`.
- Without `light_count`, a clip that has the opening signal still numbers from 0. A closing-only clip copies the last index from any other clip in the same job that has an opening signal. If no clip in the job has an opening signal, the closing-only clip is dropped with a reason that the opening flash is missing and the light count is needed to count backward.

The create-from-video page shows an optional light-count field with that explanation. It is left blank unless the operator fills it.

### Result

The CV result gains:

```json
"rejected_feeds": [{ "file": "clip.mp4", "reason": "no start or end signal found" }]
```

`file` is the upload's base name. The review screen lists each rejected clip and its reason next to the existing missing and low-confidence lists.

If fewer than two clips remain usable, the job status is `failed` and `error` is a plain sentence that the recordings missed the start and end flashes. The job does not offer a model to confirm.

## Tests

- Go: the frame order is preamble colours, bulbs `0 … n−1` white, postamble colours, then all off. Stop during the preamble writes all off and never lights bulb 0. Dwell under 500 ms fails start. Status `phase` matches the segment, and `current_index` appears only during the sweep.
- Vision: synthetic clips for both signals, a late start with `light_count`, a one-second flash before the opening signal that is ignored, a clip with no signal that is rejected, and a job that fails when fewer than two clips remain.

## Docs updated with this spec

- `docs/requirements/requirements.md` §11 and the feature-code index (`REQ-050`)
- `docs/requirements/acceptance-criteria.md`
- `docs/design/backend-lights-and-automation.md` §3.22, §3.22.1, §3.23, §3.23.3
- `docs/design/backend-service.md` capture and reconstruct API rows
- `docs/design/frontend.md` §4.15 and §4.17
- `docs/design/request-flows.md` §8.24 and §8.25
- `docs/design/overview.md`, `architecture.md`, `appendix-traceability.md`, `glossary.md`
- `docs/engineering/cv-runtime.md` dwell minimum
- `docs/userguide/build-model-from-video.md`
