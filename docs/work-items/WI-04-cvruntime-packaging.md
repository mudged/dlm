# WI-04 — Bundled OpenCV runtime + packaging/CI (REQ-048)

- **Model:** Medium (Sonnet) — treat as a **spike**; the packaging is the risky part.
- **Depends on:** none (but coordinate the JSON contract below with WI-05 and WI-06)
- **Area:** Infra + Go (`backend/internal/cvruntime`, build scripts, GitHub Actions)
- **Requirements:** REQ-048 (esp. business rule 5: no separate Python install), REQ-003 (Pi/ARM64), REQ-004 (single executable), REQ-043 (cross-platform)
- **Architecture:** `docs/architecture.md` §3.1, §3.23.1, §6.9

## Context

The Go app must run OpenCV-based reconstruction **without the operator installing Python or
OpenCV**. This is **deliberately separate** from REQ-045 user Python routines (those still use the
operator's system `python3` via `internal/routineengine`). Here, the product **ships its own**
self-contained OpenCV+Python runtime that the Go process invokes as a supervised child.

Constraints: the **main Go binary stays pure-Go / cgo-free** (do **not** add `gocv`/cgo — it
breaks the cross-compile posture in §3.4 and still needs a host OpenCV install). The CV runtime is
a **separate child process**, not linked into Go.

Existing relevant files:
- `backend/internal/config/config.go` — env/config (`DLM_DATA_DIR`, etc.).
- `.github/workflows/release.yml` — cross-compiles 3 targets with `CGO_ENABLED=0`.
- `.github/workflows/ci.yml` — PR build/test.

## Decision to make (and document)

Pick **one** packaging mechanism and document it in `docs/advanced-setup.md` (WI-09 expands it):

- **(A) Embed + first-run extract (single-file deliverable, matches REQ-004):** `//go:embed` a
  compressed, platform-matched runtime archive into `internal/cvruntime`; extract on first use to
  `DLM_DATA_DIR/runtime/cv/<version>/` (checksum-gated). Larger binary; build job must place the
  correct per-target archive before each `go build`.
- **(B) Sibling asset in the release archive (smaller binary):** ship the runtime as a directory
  next to the executable; resolve it relative to the binary. The single *download* is still
  self-contained (no operator Python), but it is a folder, not one file.

**Recommended for the MVP spike: (B)** to avoid ballooning all three cross-compiled binaries and
to keep CI simple, with a clear TODO that (A) may be adopted later for a strict single-file. Either
way **the operator installs no Python** (REQ-048 BR5). Keep the resolver pluggable so switching is cheap.

## Tasks

### 1. `internal/cvruntime` package

- **Runtime resolution** (`resolve.go`): locate the CV runtime dir in this order:
  1. `DLM_CV_RUNTIME_DIR` env override (absolute path) — used in dev/tests.
  2. (mechanism A) extracted embed under `DLM_DATA_DIR/runtime/cv/<version>/` (extract if absent, verify checksum).
  3. (mechanism B) `<dir-of-executable>/runtime/cv/` (or a documented relative path).
  Return a clear, actionable error if none found (so reconstruction fails with a helpful message, not a panic).
- **Invocation** (`run.go`): `func Run(ctx context.Context, spec JobSpec) (Result, error)` that launches the
  bundled interpreter running the `reconstruct` entrypoint (WI-05) as a **supervised child**:
  honor `ctx` cancellation/timeout, capture stderr, kill on timeout (`SIGTERM` then `SIGKILL`),
  parse the child's JSON `Result` from stdout or a result file.
- **Contract types** (`contract.go`): the JSON exchanged with the child — keep this stable; WI-05 and WI-06 import it.

```jsonc
// JobSpec (Go → child): written as JSON to a temp file or stdin
{
  "feeds": [{ "path": "/abs/work/feed0.mp4" }, { "path": "/abs/work/feed1.mp4" }],
  "marker": { "dictionary": "aruco_4x4_50", "edge_length_m": 0.1 },   // optional
  "scale_hint_m": null,                                               // optional
  "dwell_ms": 1000                                                    // from REQ-047 sweep, for blink windowing
}
// Result (child → Go):
{
  "status": "succeeded",            // or "failed"
  "light_count": 200,
  "lights": [ { "id": 0, "x": 0.0, "y": 0.0, "z": 0.0 }, ... ],  // SI metres, ids 0..n-1
  "missing": [ 17, 92 ],
  "low_confidence": [ 5 ],
  "error": null                     // human-readable string when status=failed
}
```

### 2. Build script(s)

- `scripts/build-cvruntime.sh` (or `.mjs`) that produces the runtime bundle for a given
  `GOOS/GOARCH`: a self-contained Python (e.g. **python-build-standalone**) + `opencv-python-headless`
  + `numpy` (and the WI-05 `reconstruct.py`). **Use `opencv-python-headless`** (no GUI/X deps — the
  Pi is headless). Output to a known path (`dist/cvruntime/<goos>_<goarch>/` for mechanism B, or a
  compressed archive consumed by `//go:embed` for mechanism A).
- For the spike, **get `linux/arm64` and `linux/amd64` working first**; Windows can follow. Pin
  versions; record them. Verify `aarch64` OpenCV wheels resolve (PyPI manylinux aarch64 / piwheels).
- Keep a tiny **placeholder** so `go build` / `go test` work before the bundle exists (mirror the
  `internal/webdist/placeholder.txt` pattern), and so `DLM_CV_RUNTIME_DIR` can point at a real bundle in tests.

### 3. CI / release integration

- `.github/workflows/release.yml`: add a step (per target) that runs `scripts/build-cvruntime.sh`
  before/around `go build` and **packages the runtime with each binary** (e.g. zip/tar per platform
  containing `dlm_<os>_<arch>` + `runtime/cv/`), updating the uploaded assets accordingly.
- `.github/workflows/ci.yml`: PR CI should still pass **without** building the heavy runtime (it is
  large/slow). Ensure Go builds/tests with the placeholder; gate any runtime-dependent integration
  test behind `DLM_CV_RUNTIME_DIR` being set (skip otherwise).

## Acceptance / tests

- `cd backend && go test ./...` passes with the placeholder (no real runtime needed for unit tests).
- A small `cvruntime` test verifies: resolver honors `DLM_CV_RUNTIME_DIR`; `Run` invokes a stub
  "interpreter" (a tiny script you point at via env in the test) and round-trips a `JobSpec`→`Result`
  JSON; timeout/cancel kills the child.
- Document (briefly, in code comments + a note for WI-09) the chosen mechanism, versions, and on-disk footprint.

## Out of scope

- The actual CV algorithm (WI-05) — here you only need a stub child that echoes a valid `Result` for tests.
- Job orchestration / HTTP (WI-06).

## Definition of done

A `cvruntime.Run(ctx, spec)` exists that launches a shipped (no-operator-Python) interpreter and
returns a parsed `Result`; a build script produces the bundle for at least the two Linux targets;
CI stays green with a placeholder; the packaging decision is documented.
