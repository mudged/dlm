# Consistency review fixes — design

**Date:** 2026-07-17  
**Status:** Approved for planning (brainstorming complete)  
**Scope:** Address all High, Medium, and Low findings from the requirements / design / code consistency review (scope option C).

## Context

A cross-cutting review found that product behavior is largely aligned, but:

1. The GitHub Release workflow cannot publish working Pi binaries.
2. Bulk SSE light deltas do not match REQ-041’s “changed-only” contract.
3. Optional video-reconstruction marker / scale-hint form fields are dropped.
4. Several design/requirements passages contradict each other or the code.

Approach chosen: **fix-forward in place** — behavioral code + TDD first, then align docs to the final code.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Scope | Everything (High + Medium + Low) |
| README vs §6.10 | Keep short README; update design §6.10 to match AGENTS.md / user guide |
| Capture `marker=true` | Map to default ArUco `DICT_4X4_50` + `edge_length_m: 0.1` (printable 100 mm asset); forward `scale_hint` |
| Python `scene` mutators | Design follows code: positional `patch` dict |
| Stop signals | Document `CommandContext` → SIGKILL; no SIGTERM staging required for REQ-040 |
| WLED `lastApplied` | Document whole-model elision only; no per-index cache in this pass |
| Packaging prose | Requirements acknowledge Linux `.tar.gz` (binary + `runtime/cv/`) |

## §1 Release pipeline (High)

**File:** `.github/workflows/release.yml`

1. Prefix the Pi build with `GOOS=linux GOARCH=arm64` so `dist/dlm_linux_arm64` is actually arm64.
2. Remove the Windows self-`cp` (`cp dist/dlm_windows_amd64.exe dist/dlm_windows_amd64.exe`), which exits non-zero under `bash -eo pipefail` and aborts packaging before upload. Leave the bare `.exe` in `dist/` for `softprops/action-gh-release`.
3. No new Go tests required. Optional workflow sanity check (e.g. inspect binary arch) is nice-to-have only.

## §2 SSE deltas — changed-only (High, REQ-041)

**Goal:** Bulk / reset / scene-region / routine commits emit SSE `deltas[]` only for lights whose on/color/brightness triple actually changed.

**Implementation:**

1. TDD first: failing tests for mixed batch patches (some lights already at target, some not) asserting returned / SSE delta IDs are the changed subset only.
2. Change `internal/lightstate` (`BatchPatch`, `ResetAll`) so returned DTO slices are **changed-only**; preserve `unchangedAll` / all-default short-circuit semantics.
3. Align scene-region / batch helpers in `internal/store` the same way if they currently append every matched light.
4. Handlers keep “if `!unchangedAll` → notify with returned list.”
5. Routine / shape-animation ticks use the same contract (only changed lights in the notify payload).
6. Client `lightStateSignature` elision remains as a safety net.

**Out of scope:** WLED per-index HTTP suppression cache.

## §3 Marker / scale-hint wiring (Medium, REQ-048 / REQ-049)

**File:** `backend/internal/httpapi/capture_models.go` (+ tests)

After parsing video `files`:

1. If form field `marker` is truthy (`"true"` / `"1"`), set  
   `CreateParams.Marker = &cvruntime.Marker{Dictionary: "DICT_4X4_50", EdgeLengthM: 0.1}`.
2. If `scale_hint` is a valid finite positive float, set `CreateParams.ScaleHint`.
3. Pass that struct into `reconstruct.Create` (stop using empty `CreateParams{}`).
4. API test covers multipart with `marker=true` and optional `scale_hint`.
5. Design §3.2 / §3.23 note that `marker=true` means “default printable marker.”
6. Confirm-path: add finite XYZ re-validation if inexpensive; otherwise document that Python `allow_nan=False` + JSON decode already guarantee it.

## §4 Documentation alignment

Update docs so they describe the code after §§1–3:

| Area | Change |
|------|--------|
| `docs/design/backend-service.md` §3.12 | Replace `+1 m` / mins-at-0 AABB with persisted `margin_m` (default 0.3 m, all six sides, min corner clamped ≥ 0). Include `margin_m` in the logical `scenes` table. |
| `docs/design/frontend.md` §4.9 | Framing uses tight light bounds + scene `margin_m` / framing margin (not `[0,0,0]…(Mmax+1)`). |
| `docs/design/backend-lights-and-automation.md` §3.17 | Mutating `scene.*` methods and sample snippets use positional `patch` dict. Forced-stop text matches `CommandContext` → SIGKILL within REQ-040’s 2 s bound. |
| `docs/design/backend-lights-and-automation.md` §3.19–§3.20 | WLED push elision = whole-model / commit-level only (no unimplemented `lastApplied[idx]`). |
| `docs/design/deployment.md` §6.9 | Mechanism B: sibling `runtime/cv/` next to the binary in the Linux release archive; no `DLM_DATA_DIR` extraction claim. |
| `docs/design/deployment.md` §6.10 | README is the short hobbyist landing page; download / run / `systemd` live in `docs/userguide/` (no REQ-\* in README). |
| `docs/requirements/requirements.md` REQ-004 / REQ-043 | Acknowledge Linux ships as `.tar.gz` (binary + CV runtime); Windows may still be a bare `.exe`. |
| `docs/design/overview.md` / comments | Drop references to in-browser production routine hosts / scene worker as runtime. |
| Sample cone (§3.8) | Document Y-up pose (or “equivalent pose in code”) matching `samples/cone.go`. |
| `backend/internal/cvruntime/contract.go` | Comment paths match mechanism B / `runtime/cv/`. |

## §5 Code cleanup (Low)

1. **Delete orphaned client routine runners:**  
   `web/components/PythonRoutineHost.tsx`,  
   `web/components/ShapeAnimationRoutineHost.tsx`,  
   `web/public/dlm-python-scene-worker.mjs`  
   (and any sync’d copies under `web/out` / `backend/internal/webdist` if regenerated later).  
   **Keep** `dlm-python-editor-worker.mjs` for REQ-022 editor lint/format.  
   Update comments in `pythonSceneApiCatalog.ts` / `pythonRoutineSamples.ts` to point at `bootstrap.py` + catalog, not the deleted worker.
2. **Frontend capture status type:** `"pending" | "running" | "succeeded" | "failed"` (drop `"processing"`).
3. No SIGTERM staging implementation; docs only (§4).
4. No WLED per-index cache; docs only (§4).

## Testing strategy

| Change | Test |
|--------|------|
| Release workflow | Manual / review of YAML; optional arch check |
| Changed-only deltas | New Go unit/API tests (TDD): mixed batch, reset, preferably one scene bulk path |
| Marker / scale_hint | New or extended `capture_models` API test |
| Orphan deletion | `web` lint/tests still pass; grep confirms no imports |
| Doc edits | Review-only (no automated doc tests) |

## Non-goals

- Implementing keyword-arg Python `scene` API.
- Expanding README with a full systemd unit.
- Embedding/extracting the CV runtime into the Go binary (mechanism A).
- Per-index WLED `lastApplied` cache.
- Changing the printable marker asset or dictionary.

## Success criteria

- Tag release workflow would produce a real `linux/arm64` binary and complete packaging/upload without the self-`cp` failure.
- SSE / notify payloads for bulk paths list only changed lights.
- `marker=true` and `scale_hint` reach the CV job spec.
- Design, requirements, and code no longer contradict each other on the findings above.
- Orphaned browser routine hosts and scene worker are gone; editor worker remains.
