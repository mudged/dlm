# WI-27 — Handle non-contiguous light IDs end-to-end (middle gaps)

- **Model:** Medium (Sonnet)
- **Depends on:** none (interacts with WI-25/WI-26 which produce `missing`)
- **Area:** Backend Go (`backend/internal/reconstruct`, `backend/internal/wiremodel`) + CV contract
- **Source:** Code review finding (High, contract gap).

## Context

The Python side correctly omits undetected lights from `lights` and lists them in `missing`
(`reconstruct.py` ~632–648, 724–728). But the Go confirm path runs `wiremodel.ValidateLights` on
**only** the triangulated subset, which requires contiguous 0-based IDs `0 … len(lights)-1`
(`backend/internal/reconstruct/reconstruct.go` ~215–221; `backend/internal/wiremodel/csv.go` ~86–128).

So if light 2 is missing but 0, 1, 3, 4 were triangulated, the IDs are non-contiguous relative to
`len(lights)` and **confirm fails**. Missing lights at the **end** happen to work; a gap in the
**middle** blocks model creation entirely. This is an end-to-end contract mismatch between Python
output shape and Go validation.

Key files:
- `backend/internal/reconstruct/reconstruct.go` — `Confirm` and where it builds `[]wiremodel.Light` and calls validation (~215–221).
- `backend/internal/wiremodel/csv.go` — `ValidateLights` / `ParseLightsCSV` and the sequential-ID rule (~86–128).
- `docs/requirements.md` REQ-005 / REQ-007 (light ID rules) — **read these first**: confirm whether the product rule is "IDs must be contiguous 0-based" (in which case the fix is to define behavior for gaps) or whether gaps are permissible. Do not silently violate REQ-005/007; if there's a genuine conflict, flag it.

## Tasks

Pick the approach that matches REQ-005/REQ-007 (read them; decide explicitly and document):

- **Option A — gaps are allowed:** relax validation so a model can be created with non-contiguous IDs,
  treating `missing` lights as absent rather than renumbered. Persist real IDs; ensure downstream
  (lightstate, rendering, SSE) tolerates gaps. This preserves the physical light↔ID mapping (important
  if IDs map to WLED indices).
- **Option B — gaps not allowed / must be contiguous:** at confirm time, **surface missing lights to
  the user** (the review-then-confirm UI already exists) and require resolution — either the user
  re-captures, or confirm is blocked with a clear `validation_failed` explaining which IDs are missing
  (not a generic contiguity error). Do **not** silently renumber, which would mislabel physical lights.

Whichever is chosen:
1. Make the error path **clear and specific** ("lights 2, 5 were not detected") instead of an opaque
   sequential-ID failure.
2. Ensure the chosen behavior is consistent between the CSV upload path and the capture-confirm path
   (don't diverge the two validators unexpectedly).

## Acceptance / tests

- Reconstruct/confirm test (fake `CVRunner`): a candidate result with a middle gap (`missing: [2]`)
  produces the chosen behavior — either a successfully created model with correct IDs (Option A) or a
  specific, user-actionable error naming the missing IDs (Option B). Not a panic, not an opaque error.
- `wiremodel` tests updated to match the decided rule.
- `cd backend && go test ./...` passes.

## Out of scope

- The CV detection improvements that reduce missing lights (WI-25/WI-26).

## Definition of done

A reconstruction with a mid-sequence missing light no longer dead-ends with an opaque validation
failure; behavior matches REQ-005/REQ-007 and is consistent across CSV and capture paths. Covered by
tests.
