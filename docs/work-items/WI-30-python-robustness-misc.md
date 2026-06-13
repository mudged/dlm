# WI-30 — Python robustness misc (dwell, exit codes, capture release, scale validation)

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** CV / Python (`backend/internal/cvruntime/src/reconstruct.py`) + one Go default
- **Source:** Code review findings (Low). Independent sub-tasks; do as many as cleanly apply.

## Context & tasks

1. **`dwell_ms` accepted but unused** — `reconstruct.py` ~112–115, 686–697. `detect_blinks` takes
   `dwell_ms` but detection is purely brightness edge-triggered with no min/max dwell validation, so
   stray flashes / compression flicker / auto-exposure pumping can create extra blinks. Either **use**
   `dwell_ms` to reject blinks shorter/longer than expected (debounce around the dwell), or remove the
   unused parameter and document that detection is dwell-agnostic. Note Go currently passes
   `dwell_ms: 0` (see task 4).

2. **Inconsistent exit codes** — `reconstruct.py` ~706–708 vs ~681–684, 737–741. "No blinks detected"
   prints `status:"failed"` JSON but exits `0`, while other failures `sys.exit(1)`. Go reads the JSON
   `status` so it copes, but make logical failures consistent: prefer always emitting a `status:"failed"`
   JSON result **and** exiting non-zero on failure (confirm `run.go` ~70–83 still treats a `failed`
   JSON correctly when exit code is non-zero — adjust whichever side is needed so success/JSON/exit are
   coherent). Keep stdout = JSON only, stderr = diagnostics.

3. **`VideoCapture` leak on exception** — `reconstruct.py` ~132–204, 425–448. `detect_blinks` and
   `_aruco_pose_single_feed` call `cap.release()` only on normal paths; an exception mid-loop leaks the
   fd until GC. Wrap captures in `try/finally` (or a context-manager helper) so `release()` always runs.

4. **Validate scale inputs > 0** — `reconstruct.py` ~378, 587, 686–688. Zero/negative `scale_hint_m`
   or `edge_length_m` silently produce zero or sign-flipped coordinates with `status:"succeeded"`.
   Validate these are finite and `> 0` up front; on bad input emit `status:"failed"` with a clear error.

5. **Go default `dwell_ms: 0`** — `backend/internal/httpapi/capture_models.go` ~72 uses empty
   `CreateParams{}`, so the JobSpec gets `"dwell_ms": 0`. If task 1 keeps `dwell_ms` meaningful, set a
   sensible default here (and/or in the spec construction) so the contract isn't broken; if `dwell_ms`
   is removed, drop it from the Go spec too. Keep Python and Go in sync.

## Acceptance / tests

- Python tests: bad scale (`0`, negative) → clean `status:"failed"`; capture is released even when
  detection raises (can assert via a fake/mock or by structure review); exit code/JSON coherent on the
  no-blinks path.
- If `dwell_ms` is kept: a fixture with a too-short flash is rejected as a blink.
- `cd backend && go test ./...` and the cvruntime Python tests (see src `README.md`) pass.

## Out of scope

- Detection-algorithm robustness (WI-29) and correspondence (WI-25).

## Definition of done

Invalid scale inputs fail cleanly, captures are always released, failure exit-codes/JSON are coherent,
and `dwell_ms` is either honored or removed consistently across Python and Go. Tests green.
