# WI-09 — README + advanced-setup docs for camera capture (REQ-047–049)

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** WI-04 (packaging mechanism decided), WI-06 (API + flow exist). Ideally also after WI-03/WI-08 so the UX is final.
- **Area:** Docs (`README.md`, `docs/advanced-setup.md`)
- **Requirements:** REQ-046 (README operator focus), REQ-045/REQ-048 (runtime prerequisites), REQ-047/REQ-049 (usage)
- **Architecture:** `docs/architecture.md` §6.9, §6.10, §3.23.1

## Context

`README.md` is the **hobbyist/operator-facing** doc; `docs/advanced-setup.md` holds contributor
detail. Per `AGENTS.md`, **README must NOT mention internal `REQ-*` identifiers**, traceability, or
spec-only terms — describe behaviour in plain language. Capability identifiers belong in `docs/`.

This item documents the new camera-capture feature for operators and the bundled OpenCV runtime for
contributors. Confirm the exact behaviour from the shipped code (don't restate this prompt verbatim).

## Tasks

1. **`README.md`** (plain language, no `REQ-*`):
   - A short **"Build a model from video"** section: point a camera at your lights from two or more
     angles, start the capture light sequence from the **Devices** screen (it lights each bulb in turn
     for ~1 second), record from each angle, then on the **Models** screen choose **create from video**,
     upload the recordings, review the detected lights, and save. Mention the **optional printable
     marker** and that more angles improve accuracy.
   - A note in the prerequisites/run section that **camera capture needs no extra install** — the
     OpenCV runtime ships with the app (contrast with the existing note that *custom Python routines*
     require system `python3`). Keep both statements consistent and clearly separated.
   - If the release packaging changed (per WI-04 mechanism choice — e.g. the download is now an archive
     containing the binary + a `runtime/` folder), update the **download/run** and **update** steps and
     the systemd notes accordingly (working directory must let the app find its runtime).
2. **`docs/advanced-setup.md`** (contributor detail):
   - The chosen CV runtime packaging mechanism (embed-and-extract vs sibling asset), pinned versions
     (Python, `opencv-python-headless`, numpy), how to build it (`scripts/build-cvruntime.sh`), and the
     approximate **on-disk footprint** and extraction path (`DLM_DATA_DIR/runtime/cv/<version>/`).
   - The `DLM_CV_RUNTIME_DIR` / `DLM_CAPTURE_DWELL_MS` env vars (and any others introduced), and how to
     run reconstruction locally for development (system venv with `opencv-python-headless`).
   - A pointer that PR CI uses a placeholder runtime and that runtime-dependent integration tests are
     gated behind `DLM_CV_RUNTIME_DIR`.

## Acceptance / tests

- No `REQ-*` identifiers in `README.md` (grep to confirm).
- README and `docs/advanced-setup.md` accurately match the shipped commands, env vars, and packaging
  (verify against the code from WI-04/WI-06, not just this prompt).
- If a docs/markdown linter or link-checker runs in CI, it passes.

## Out of scope

- Editing `docs/requirements.md` / `docs/architecture.md` (settled). Code changes.

## Definition of done

Operators can follow `README.md` to capture a model from video and understand that no extra install
is needed for capture; contributors can build/run the CV runtime from `docs/advanced-setup.md`; no
internal identifiers leak into the README.
