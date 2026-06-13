# WI-25 — Robust cross-feed blink correspondence (fix ordinal drift)

- **Model:** Hard (Opus) — algorithmic correctness, careful reasoning required
- **Depends on:** none (coordinates with WI-26 on what counts as "valid")
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`)
- **Source:** Code review findings (High). Repo root `/workspaces/dlm`.

## Context

Lights are identified by **blink order**: each light blinks once in sequence, and detection assigns the
k-th detected blink in a feed to light index `k`. `align_detections` (`reconstruct.py` ~241–257) maps
per-feed blink ordinal directly to light index, and `detect_blinks` (~168–202) is a brightness
edge-trigger.

**Bug:** if one feed **misses** a blink (occlusion, noise, threshold) or **adds** a spurious one, that
feed's ordinals shift relative to the others. Ordinal `k` then refers to **different physical lights**
in different feeds. The pipeline still triangulates those mismatched 2D points and emits coordinates —
potentially for the **wrong** light ID, or for a light that should be reported missing. `low_confidence`
does **not** exclude a point from `lights`, so this can violate REQ-048 BR 2/BR 7 ("never fabricate")
in practice (`triangulate_all` ~568–588). There is no cross-feed blink-count consistency check or
resynchronization.

Key files:
- `backend/internal/cvruntime/src/reconstruct.py` — `detect_blinks` (~132–204), `align_detections` (~241–257), `triangulate_all` (~538–590), `by_light` construction, the `missing`/`low_confidence` bookkeeping (~724–728), and the JSON `Result` assembly (~632–648).
- Go contract: `backend/internal/cvruntime/contract.go`, and `wiremodel.ValidateLights` (`backend/internal/wiremodel/csv.go`) which requires sequential 0-based IDs — see WI-27 for the contiguity interaction.
- `docs/requirements.md` REQ-048 (BR 2, BR 7).

## Tasks

1. **Detect cross-feed inconsistency:** compare the number of detected blinks per feed. If feeds
   disagree on blink count, do **not** blindly map ordinal→index. Add a consistency check and decide a
   principled policy (below). Log per-feed counts to stderr for diagnostics.
2. **Make correspondence robust** rather than positional. Options (pick the most reliable that fits the
   capture model; document the choice):
   - Use the **expected light count** (the sweep drives a known number of lights, available from the
     job spec / device `light_count`) to anchor indexing, and use per-blink **timestamps** to align
     blinks across feeds (blinks of the same light occur at ~the same time across synchronized feeds),
     instead of raw ordinal position.
   - When a feed's blink for index `k` is absent or ambiguous, mark light `k` **missing** in that feed
     (so it simply has fewer views) rather than letting later blinks slide into slot `k`.
3. **Never emit mismatched correspondences:** a light should only get coordinates from 2D points that
   correspond to the **same** physical blink across feeds. If correspondence for an index can't be
   established consistently, put the index in `missing`, not `lights`.
4. Keep `missing` / `low_confidence` semantics correct: missing = no reliable ≥2-view correspondence;
   low_confidence = triangulated but high reprojection error (still in `lights`).

## Acceptance / tests

- New test: a multi-feed fixture where one feed drops a single blink → the affected light is reported
  `missing` (or correctly re-aligned), and **no other** light receives wrong coordinates. Assert that
  later lights keep correct IDs/positions.
- New test: a feed with one spurious extra blink does not shift downstream IDs.
- Existing happy-path and occlusion tests still pass.
- Run the cvruntime Python tests per `backend/internal/cvruntime/src/README.md`.

## Out of scope

- Geometry sanity (cheirality/NaN) — WI-26. Non-contiguous ID handling on the Go side — WI-27.

## Definition of done

A missed or spurious blink in one feed never mislabels other lights; indices that can't be reliably
corresponded across feeds are reported missing, not fabricated. Proven by new tests.
