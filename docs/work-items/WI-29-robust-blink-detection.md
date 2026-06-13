# WI-29 — Robust background model & reliable frame rewind for blink detection

- **Model:** Medium (Sonnet)
- **Depends on:** none (complements WI-25)
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`)
- **Source:** Code review findings (Medium). Repo root `/workspaces/dlm`.

## Context

Blink detection builds a per-pixel background from the first ~10 frames, then rewinds the capture to
scan for blinks. Two reliability issues:

1. **Unreliable rewind** — `reconstruct.py` ~166–167 uses `cap.set(cv2.CAP_PROP_POS_FRAMES, 0)`. On
   some MP4/MKV backends (common for phone uploads) frame seeking is unreliable, so the blink pass may
   not align with the frames the background was built from → missed/spurious blinks.
2. **Background contaminated by always-on light** — `reconstruct.py` ~149–163 builds the background as
   the per-pixel **minimum** over the first `BG_FRAMES` (~10) frames. If a bulb is lit throughout that
   window, those pixels never appear dark in the background, so later blinks at those locations may not
   register → ordinal gaps / missed lights (which then feed the drift in WI-25).

Key files:
- `backend/internal/cvruntime/src/reconstruct.py` — `detect_blinks` and background construction (~132–204, esp. ~149–167), `BG_FRAMES`, `MAX_PROC_DIM` downscale.

## Tasks

1. **Reliable rewind:** instead of relying on `CAP_PROP_POS_FRAMES = 0`, re-open the `VideoCapture`
   (open a fresh `cv2.VideoCapture(path)` for the blink pass) so reading restarts from frame 0
   deterministically across backends. Release captures properly (coordinate with WI-30's release fix).
   Alternatively, read all needed frames in a single forward pass (build background from the first N
   while buffering only what's necessary), avoiding a seek entirely — prefer this if memory allows on a
   Pi.
2. **Better background:** make the background robust to always-on pixels. Options (document the choice):
   - Use a temporal **median** (or a low percentile) over a larger sampling of frames spread across the
     video rather than the min of the first 10 consecutive frames, so a light that is on during the
     first window doesn't poison its pixels; or
   - Detect blinks via **per-pixel temporal change** (frame-to-frame brightness delta / peak detection)
     rather than absolute difference from a static background, which is inherently robust to always-on
     regions.
3. Keep it Pi-friendly: continue processing frames one at a time with the existing downscale; don't load
   the whole video into memory.

## Acceptance / tests

- New test/fixture: a video where one light is on during the initial background window still has its
  blink detected (no missed light from background contamination).
- New test: detection works on a fixture/back-end where naive `POS_FRAMES` seek would misalign (or a
  unit test of the re-open path).
- Existing detection tests still pass; run cvruntime Python tests per the src `README.md`.

## Out of scope

- Cross-feed correspondence/missing-light policy (WI-25) — this item reduces *how often* blinks are
  missed; WI-25 makes the system correct *when* they are.

## Definition of done

Blink detection rewinds reliably across container backends and tolerates always-on lights during the
background window, reducing missed/spurious blinks. Covered by new tests.
