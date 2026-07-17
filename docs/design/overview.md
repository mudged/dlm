# Goals, constraints, and repository layout

This file introduces the "dlm" (Domestic Light & Magic) architecture at the highest level: the big
runtime decision (one Go binary, no Node.js at runtime), the requirement-by-requirement summary of
how the product is built (§1), and the folder structure of the repository (§2). Read it first to get
your bearings before diving into the deeper per-area documents.

Part of the [dlm architecture](architecture.md); see the [glossary](glossary.md) for unfamiliar terms.

This document defines the technical structure and deployment for the product described in
`docs/requirements/requirements.md` (REQ-001–REQ-049). For the full map of which requirement is
covered by which architecture section, see `appendix-traceability.md`.

## Architectural resolution: REQ-004 (single binary) vs Next.js

**In plain terms:** The web UI is authored with Next.js (a React framework), but the shipped product
does **not** run Node.js on the Raspberry Pi. Instead, one Go program serves both the JSON API and a
pre-built static copy of the UI. React still runs, but only in the user's browser.

**Decision:** Full Next.js SSR/Node at runtime is out of scope for the shipped product. (SSR =
server-side rendering, where a Node server builds HTML for each request.) The open item in REQ-004 is
resolved as follows:

- Next.js + App Router + Tailwind remain the **authoring** stack under `web/` (toolchain, components,
  styling).
- **Runtime** on the Pi is a **single Go process** that:
  - Serves the JSON API (`/health`, `/api/v1/...`).
  - Serves the UI as **static assets** produced by `next build` with `output: 'export'` (static HTML,
    JS, CSS), embedded in the binary via Go `embed`, or baked in at link time from a generated
    filesystem tree.
- **REQ-002 (reactive UI):** Achieved with client-side React — hydration (React attaching to
  server-rendered HTML in the browser) and `"use client"` components — plus browser `fetch` to
  same-origin `/api/v1/...` on the Go listener. There are **no** React Server Components (RSC, React
  components that render on a Node server per request) that require a Node runtime at request time.
  Any async server-only data patterns used during development MUST be replaced or duplicated with
  client fetches or build-time data before release.

This meets REQ-004 rule 1 (no separate Node.js runtime in the distribution) and aligns with Pi RAM
constraints (§6).

---

## 1. Goals and constraints

**In plain terms:** This is the master traceability table. Each row names a requirement code
(`REQ-NNN`) from the requirements doc and summarises how the architecture answers it. The `§N.N`
references point at the detailed section that implements that requirement; a section index in
`architecture.md` maps every `§` to the file it now lives in. Skim it as a lookup table — you rarely
read it top-to-bottom, but you return to it to find "where is REQ-022 handled?".

| Requirement | Architectural response |
|-------------|-------------------------|
| REQ-001 | Go binary serves an HTTP API plus an embedded static UI built from Next.js + Tailwind source in `web/`. |
| REQ-002 | Mobile-first Tailwind, Client Components for interactivity; `fetch('/api/v1/…')` from the browser to the same Go origin. |
| REQ-003 | Primary release target **linux/arm64** (Pi 4B, 64-bit OS); document CPU/RAM; one long-lived app process. |
| REQ-004 | One downloadable app per release target (Linux may be a `.tar.gz` with the binary + sibling `runtime/cv/`; Windows may be a bare `.exe`); UI assets embedded in the Go binary; Docker/compose is not the canonical install path. |
| REQ-005 | Domain types + CSV contract; ordered chain adjacency (id `i` ↔ `i±1` only); metadata (name, creation instant) stored with each model; coordinates as `float64` in Go/API JSON. |
| REQ-006 | REST JSON API for models + Next.js client pages (list, detail, upload, delete); multipart upload for CSV; `409` `model_in_scenes` when a delete is blocked (§3.13). |
| REQ-007 | Authoritative validation in Go when ingesting CSV; transactional create (all-or-nothing); `400` responses with a clear error envelope. |
| REQ-008 | `scripts/run.sh` (or documented equivalent) from repo root: `npm run release:sync` in `web/` then `go run ./cmd/server` in `backend/`; README documents the exact invocation (§3.7). |
| REQ-009 | `internal/samples` builds three ordered light paths on each solid's boundary (500 ≤ n ≤ 1000, 0.05–0.10 m consecutive chords, §3.8) with even placement on faces / surfaces per §3.8; `store.SeedDefaultSamples` when `models` is empty (§3.8). |
| REQ-010 | `three` used directly; client-only detail view: `InstancedMesh` (or equivalent §4.7) ø 0.02 m markers for all n lights (no LOD/decimation); `LineSegments` `i↔i+1` only (the chain); segment colour `#D0D0D0`, opacity 0.15 (85% transparent), thinner/subtler than the spheres; `Raycaster` + DOM tooltip (§4.7). |
| REQ-011 | REST `/api/v1/models/{id}/lights/...` and scene §3.15 routes for read/write of logical state; authoritative data lives in `LightStateStore` (§3.9, §3.21), not SQLite (REQ-039). See §3.18 (REQ-029) for aggregate paths. |
| REQ-012 | three.js materials: **on** = opaque base colour × brightness + emissive glow (REQ-028); **off** = `#D0D0D0`, opacity 0.15 (matches REQ-010 segments); the client merges API state; no indefinite staleness after writes (§4.7, §8.7). |
| REQ-013 | Model detail light table: client-side pagination over the `GET` detail payload; page-size presets 25 / 50 / 100 (default 50); "go to id" computes the target page + inline validation; checkbox multi-select with optional Shift+click range on the current page; selection retained across pages in React state (`Set<number>`) until cleared or navigated away; bulk apply calls `PATCH …/lights/state/batch` (§3.10); list and three.js updated from the response (§4.8, §8.8). |
| REQ-014 | Insert defaults `on=false`, `color="#ffffff"`, `brightness_pct=100` (§3.9); `POST …/lights/state/reset` (§3.11); model-detail Reset button (not hover-only) calls reset then merges `states` into the UI + §4.7 (§4.6, §8.9). |
| REQ-015 | `scenes` + `scene_models` tables (§3.12); `POST /scenes` computes all offsets from the ordered `model_id` list; `GET /models/{id}` is unchanged (canonical coords); `GET /scenes/{id}` returns `x,y,z` + `sx,sy,sz` (§3.13); composite three.js §4.9; `409` `model_in_scenes` (§3.2, §8.4). |
| REQ-016 | §4.7 / §4.9: a "Reset camera" control re-applies `applyDefaultFraming` (the same pure function as initial load, recomputed from current bounds); sets `OrbitControls` target + camera position + `update()`; does not call the API. §4.10 IA is an optional link only — primary placement is adjacent to the viewport. |
| REQ-017 | `POST /api/v1/system/factory-reset` (§3.2, §3.14); `store.FactoryReset` transaction deletes `routine_runs` + `routines` (including `shape_animation` rows), all `devices` (REQ-035/REQ-036), then scenes/models/lights in FK-safe order, then `SeedDefaultSamples` + `SeedDefaultPythonRoutines` (§3.8.1); confirmation dialog §4.10 must mention devices among the data removed; on success, navigate to `/models` + show a success message (the architectural default for post-reset navigation). |
| REQ-018 | §4.11: Tailwind `dark:` variant on `html` (add/remove `class="dark"` or `data-theme="dark"` per Tailwind v4/v3 config); before any stored user choice, derive the initial light/dark value from `window.matchMedia('(prefers-color-scheme: dark)')` (or equivalent) when available; fall back to light if the platform exposes no scheme; after the user toggles theme, persist `light` or `dark` in `localStorage` key `dlm-theme` and re-apply on load before first paint (an inline blocking script in `layout` is recommended to avoid a flash) so the persisted value overrides `prefers-color-scheme` until cleared or changed; shell tokens: light `bg-white` + `text-gray-900`; dark `bg-gray-900` + `text-white` (a dark grey background, not mandating `#000`); `AppShell`: header (burger, `faLightbulb` regular logo, exact title `Domestic Light & Magic`, theme toggle with icons) + collapsible left `<aside>` nav + `main`; Font Awesome Free packages `@fortawesome/fontawesome-svg-core`, `@fortawesome/react-fontawesome`, `@fortawesome/free-regular-svg-icons`, `@fortawesome/free-solid-svg-icons`; buttons and button-styled controls include a visible Font Awesome icon + an accessible name. |
| REQ-019 | §4.7, §4.9: `ModelLightsCanvas` / `SceneLightsCanvas` set a fixed dark-grey clear/scene background and a matching letterbox wrapper (see the §4.7 viewport subsection) independent of the `html` `dark` class; this does not replace the REQ-018 shell tokens for the `main` chrome. |
| REQ-020 | §3.2, §3.12, §3.13, §3.15, §8.15: scene-space API for dimensions + full/filtered light retrieval (cuboid/sphere) + transactional bulk updates (cuboid/sphere + `PATCH …/lights/state/scene` + `PATCH …/lights/state/batch`), all computed from derived scene coordinates (`sx/sy/sz`) and validated with explicit geometry rules. |
| REQ-021 | §3.2, §3.14, §3.15, §3.16, §3.17, §3.17.2, §4.9, §4.11, §4.12, §4.14, §8.16/§8.17/§8.21/§8.22: two kinds of routine; `internal/routineengine` supervises runs after `POST …/start` (REQ-038 headless); the browser only starts/stops/observes (REQ-041); stop ≤ 2 s (REQ-040); effects flow only through §3.15 semantics (not `/models/.../lights` for routine automation); one active run per scene; `409` while running on destructive definition mutations. |
| REQ-022 | §3.2, §3.15, §3.16, §3.17 (a supervised `python3` child + REQ-030 in CPython), §4.9, §4.11, §4.12, §4.13 (CodeMirror 6 editor; optional Pyodide for lint/format only), §6.2, §8.17: `python_source` stored in SQLite; the production loop runs in `python3` with a `scene` shim (`bootstrap.py`) calling loopback §3.15; cooperative cancel then `CommandContext` / SIGKILL within REQ-040. |
| REQ-023 | §4.12 (the create flow chooses Python or shape animation — exactly two kinds, no third engine), §4.13, §4.14, §4.9, §4.11; `POST /api/v1/routines` `type` is `python_scene_script` or `shape_animation` (§3.16). |
| REQ-024 | §4.13 (`<section id="python-scene-api-catalog">` immediately below the editor — the complete `scene` API manifest from §3.17 including REQ-030 `scene.random_hex_colour`, `scene.max_x`/`max_y`/`max_z`; per-entry commented snippets for API items only; REQ-032 default routines are not required as whole-script catalog rows (REQ-024 rule 7); picker + insert at caret or end-of-file). |
| REQ-025 | §4.13 (`PYTHON_ROUTINE_DEFAULT_SOURCE` in `web/` — applied when creating a `python_scene_script` with empty `python_source` and an optional "Reset template"); uses `scene.set_lights_in_sphere` with a positional `patch` dict (e.g. `{"on": True, "color": colour, "brightness_pct": 100}` — §3.17); SHOULD use `scene.random_hex_colour()` for the demo colour (REQ-030) instead of `import random` for that single pattern. |
| REQ-026 | §3.15 (axis mapping for `size.width`, `size.height`, `size.depth` and `max.x/y/z`), §3.17 (Python `scene.width`, `scene.height`, `scene.depth`, `scene.max_x`, `scene.max_y`, `scene.max_z` from `GET …/dimensions`), §4.13 (REQ-024 API reference manifest). |
| REQ-027 | §4.13 and §4.14 (unified panel per kind: one scene `<select>` for run + `SceneLightsCanvas`; SSE `…/lights/events` per REQ-041; Start/Stop; "Reset scene lights" → `POST …/routines/runs/…/stop` for that scene when a run is active (or equivalent server ordering) then `PATCH …/lights/state/scene` with REQ-014 defaults; "Reset camera" → `applyDefaultFraming`), §8.18, §8.22; the viewport works with or without an active run after stop. |
| REQ-028 | §4.7 (emissive mapping + scene lights), §4.9 (composite reuse), §4.13, §4.14 (unified live viewports), §6.5 (browser GPU on the Pi and other clients), §8.5, §8.7 (material updates after PATCH); brightness for the glow tracks the authoritative in-memory light state (REQ-039) — not durable DB columns. |
| REQ-029 | §3.2 (batch/bulk routes listed below), §3.10, §3.15, §3.18 (write path, connections, observer strategy — REQ-041 tightens the shipped UI to MUST-use push + deltas), §4.3, §4.7, §4.9, §6, §7, §8.19, §9 (payload/rate limits). |
| REQ-030 | §3.17 (`scene.random_hex_colour()` — a local `random.randrange(0x1000000)` + `"#%06x"` in the `python3` child running user scripts, same distribution as documented in REQ-030), §4.13 (REQ-024 manifest + completions), `web/lib/pythonSceneApiCatalog.ts`. Production `scene` surface is `backend/internal/routineengine/bootstrap.py` only (no browser scene-worker runtime). |
| REQ-031 | §3.9, §3.10, §3.15, §3.16, §3.17, §3.17.2, §3.18, §3.19, §3.20, §4.7, §4.9, §4.13, §4.14, §8.7, §8.8, §8.20 (equivalence check before the memory update, device push, and three.js redraw). |
| REQ-032 | §3.8.1 (seed three `python_scene_script` rows on empty `routines` + on factory reset), §3.17.1 (geometry + timing for the growing sphere + sweeping cuboid + normative behaviour for the random-colour cycle over all lights), `web/lib/pythonRoutineSamples.ts` (exports `PYTHON_SAMPLE_*` for all three bodies — single source of truth); an optional toolbar "Load sample" MAY reuse those strings, but the REQ-024 catalog must not be the only delivery of full default scripts; the REQ-025 default new-routine template is unchanged unless the product replaces it later. |
| REQ-033 | §3.16 (schema + `POST`/`PATCH`), §3.17.2 (Go `time.Ticker` simulation + internal `BatchPatchSceneLights` / equivalent — no browser `fetch` for production ticks), §4.12, §4.14, §6.2, §8.21, §8.22 (the unified viewport observes SSE/GET; optional TS ghost preview only). |
| REQ-034 | §3.12–§3.13, §4.7, §4.9, §4.14: persisted `margin_m` per scene; a faint AABB (axis-aligned bounding box) wire on model and scene canvases (tight light bounds + padding) using the same `#D0D0D0` / opacity 0.15 line recipe as the REQ-010 chain segments (§4.7 shared material constants). |
| REQ-035 | §3.20: device registry, type `wled` (WLED = the open-source LED-controller firmware that physically drives the lights), mapping a model light `idx` → WLED LED indices (§3.20); `base_url` must be `http`/`https` + a non-empty host on the allowlist (REQ-035 BR 6), enforced in `backend/internal/store/devices.go` and mapped to `400` `invalid_base_url` in `backend/internal/httpapi/devices.go`. |
| REQ-036 | §3.20: at most one `model_id` per device and vice versa (a nullable FK (foreign key) + UNIQUE constraints, or equivalent). |
| REQ-037 | §4.11 nav + §4.15 (Devices routes/pages) — Must: list, add via manual `base_url`/host/port (REQ-035 MVP), assign/unassign, edit name, delete/forget (cascades the assignment per §3.20); discovery (`POST …/devices/discover`) is optional until implemented. |
| REQ-038 | §3.16 normative placement paragraph + §3.17–§3.17.2: `internal/routineengine` drives Python and shape ticks on the server; each commit updates `LightStateStore` then §3.20 when assigned; the browser observes via REQ-041 only; headless operation is normative. |
| REQ-039 | §3.3, §3.9, §3.21: `internal/lightstate` (in-memory per model: a map of light `idx` → state triple); normative startup-sync and unassign policies (§3.21); API reads/writes hit memory first. |
| REQ-040 | §3.17, §3.17.2: after an accepted `POST …/stop`, no further routine-originated state mutations within ≤ 2 SI seconds (cancel supervisor context → `exec.CommandContext` kills `python3` / SIGKILL on Unix; the shape `time.Ticker` is cancelled before the next tick). |
| REQ-041 | §3.18 (SSE (Server-Sent Events) event schema with `deltas[]`), §4.3, §4.7, §4.9, §4.13, §4.14: the shipped web three.js uses `EventSource` (or WebSocket if documented) as the primary observer after an initial `GET`; merge deltas into existing meshes without a full graph rebuild on each message when only a subset changed (§3.19). REQ-041 BR 6 fan-out failure logging lives in §3.18 (`RevisionHub.Notify*` emits `slog.Warn` with `model_id` + wrapped error when `store.ListSceneIDsForModel` fails). |
| REQ-042 | §4.16, §4.9, §4.13, §4.14, §3.2 (runs list + start `409` payload), §8.23: the UI reflects server truth for whether a scene has a `running` routine (`GET …/routines/runs`); on mount / scene selection / return navigation, re-fetch scene + runs and re-subscribe `EventSource` `…/lights/events` (REQ-041); in dev `next`, `web/lib/sseUrl.ts` uses the direct origin + `CORS_ALLOWED_ORIGINS`; on `409` `scene_routine_conflict`, parse `error.details.run_id` / `routine_id` to restore Stop (`web/lib/routines.ts` — `SceneRoutineConflictError`). |
| REQ-043 | §3.4, §6.6: three canonical release targets (`GOOS`/`GOARCH`) — windows/amd64, linux/amd64, linux/arm64 (Pi 4 class); Linux ships as `.tar.gz` (binary + sibling `runtime/cv/`), Windows may be a bare `.exe`; static UI embedded; documented artifact names. |
| REQ-044 | §6.7, §6.8, §6.11, §7 (CI diagram): GitHub Actions CI (build + lint/test) on PRs; branch protection requires green checks before merge; a release workflow tags and uploads assets to GitHub Releases. |
| REQ-045 | §6.9: the deployment host needs the product binary (Linux: plus sibling `runtime/cv/` from the release archive) for the baseline (API + embedded UI + SQLite + shape animation + capture CV); `python3` on `PATH` (≥ a documented minimum — the implementor picks, e.g. 3.11+) is required only when running user Python scene routines (§3.17). |
| REQ-046 | §6.10, §3.7: `README.md` stays a short landing page (no `REQ-*` IDs); download/run/`systemd` live in `docs/userguide/` (`getting-started.md`, `running-as-a-service.md`); the developer `./scripts/run.sh` remains secondary (REQ-008) per `docs/engineering/` cross-links as needed. |
| REQ-047 | §3.22, §3.2 (`/devices/{id}/capture/*`), §3.3 (`devices.light_count`), §4.15, §8.24: `internal/capture` runs a server-side sweep that drives one WLED LED `on` for ≈ 1 s in `idx` order `0 … n−1` using the device's configured `light_count` (the device need not be assigned — REQ-036); all-off on stop/completion within REQ-040's 2 s bound; at most one active sweep per device; a deterministic ordinal→index mapping for REQ-048. |
| REQ-048 | §3.23 (`internal/reconstruct` pipeline), §3.23.1 (a bundled OpenCV runtime, no separate Python install — distinct from REQ-045), §6.9, §8.25: ≥ 2 video feeds → per-feed 2D blink detection keyed to the sweep ordinal (REQ-047) → multi-view triangulation → per-light `x,y,z` (SI m, REQ-005); optional fiducial markers improve pose/scale/alignment (their absence does not gate the flow); undetected lights are reported, never fabricated; an async server-side job (Pi-feasible, REQ-003), with no browser needed to finish. |
| REQ-049 | §3.23.2, §3.2 (`/models/capture*`), §4.17, §8.25: the Models "create from video" path uploads ≥ 2 files, runs the REQ-048 job, shows a review (detected count + missing/low-confidence), then on explicit confirm persists a normal model via REQ-005/REQ-007 validation (§3.3 transaction); cancel discards; an optional printable fiducial-marker artifact (not required to create a model). |

**Assumed Pi context:** Raspberry Pi 4 Model B, 64-bit OS, ARM64 userspace. **2–8 GB RAM** — with no
Node at runtime, 4 GB is practical for modest traffic; off-device `next export` builds are recommended.

**Canonical production listener:** A single Go process on one TCP address (e.g. `:8080`, or behind an
optional reverse proxy on `:80`/`443` that forwards all paths to that one upstream). There is **no**
second application process in the product distribution.

**Dev vs prod:** Local development may still run `next dev` alongside `go run` for fast iteration; that
two-process pattern is not part of the shipped artifact (REQ-004).

---

## 2. Repository layout

**In plain terms:** dlm is a monorepo — one Git root holding both the Go backend and the Next.js
frontend. The Go code is a single module under `backend/`; the web source lives under `web/`; the
build pipeline copies the built UI into the backend so it can be embedded.

Monorepo (single Git root), one Go module under `backend/`, no `go.work` unless later expanded.

```
dlm/
  scripts/
    run.sh                    # REQ-008: release:sync + go run server (repo root relative paths)
  backend/
    cmd/
      server/                 # main() — single binary entry; calls seed-if-empty (REQ-009)
    internal/
      config/
      httpapi/                # API mux + middleware + JSON handlers (incl. models HTTP)
      wiremodel/              # domain types, CSV parse + validation (REQ-005/007)
      samples/                # deterministic boundary sampling + id order: sphere, cube, cone (REQ-009 §3.8)
      store/                  # persistence for models, scenes, routines (SQLite, REQ-006); not per-light output state (REQ-039)
      lightstate/             # in-memory authoritative per-light state + load/save hooks (REQ-039, REQ-011)
      devices/                # device registry, WLED client, discovery adapter (REQ-035–REQ-038)
      capture/                # REQ-047: server-side device light-sweep controller (drives WLED by idx; independent of model assignment)
      reconstruct/            # REQ-048/049: camera-capture job orchestration + supervised bundled OpenCV child; produces candidate light coordinates
      cvruntime/              # REQ-048: resolves sibling runtime/cv/ (mechanism B) — self-contained OpenCV+Python; no operator Python install (§3.23.1)
      routineengine/          # REQ-021/038: supervised python3 + Go time.Ticker shape sim; internal §3.15/lightstate calls
      webdist/                # holds embedded payload (see §3.5); populated by build, not hand-edited
        placeholder.txt       # optional tiny file so empty embed works in dev before first UI build
  web/                        # Next.js + Tailwind (source only for runtime; sibling of backend/)
    app/
    components/               # incl. AppShell, theme toggle, nav (REQ-018 §4.11); ModelLightsCanvas / SceneLightsCanvas (REQ-019 §4.7–§4.9); start/stop + SSE observe only (REQ-021/038)
    lib/
  docs/
```

**Boundaries:**

- `backend/` MUST NOT import TypeScript sources from `web/` — only the **built static files** under
  `internal/webdist/` (or equivalent) prepared by the build pipeline.
- `web/` MUST NOT import Go.
- **Public HTTP contract:** `GET /health`, the `/api/v1/*` JSON API, and static paths for the UI
  (`/`, `/_next/...`, assets) served from the embedded filesystem.

---
