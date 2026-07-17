# Consistency review fixes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix release packaging, REQ-041 changed-only SSE deltas, capture marker/scale-hint wiring, remove orphaned browser routine runners, and align requirements/design docs with the code.

**Architecture:** Fix-forward in place. Behavioral changes land behind failing tests first (`internal/lightstate`, `internal/store` scene bulk paths, `httpapi` capture). Handlers keep notifying from the returned state lists. Docs are updated last so they describe the final code. Release workflow YAML is a small independent fix with no Go tests.

**Tech Stack:** Go 1.25, GitHub Actions, Next.js/TypeScript (frontend type + delete unused files), Markdown design/requirements docs.

**Spec:** [`docs/superpowers/specs/2026-07-17-consistency-review-fixes-design.md`](../specs/2026-07-17-consistency-review-fixes-design.md)

## Global Constraints

- Do not renumber or reuse `REQ-*` codes or `§N.N` section numbers.
- Do not put `REQ-*` IDs in `README.md`.
- Go tests: `cd backend && go test ./...` (toolchain from `backend/go.mod`, Go ≥ 1.25).
- Frontend tests: `cd web && npm test` / `npm run lint` when touching `web/`.
- Default marker when `marker=true`: `Dictionary: "DICT_4X4_50"`, `EdgeLengthM: 0.1` (printable 100 mm asset).
- Python `scene` mutators stay positional `patch` dict in code; design docs are updated to match.
- No SIGTERM staging implementation; no WLED per-index `lastApplied` cache.
- Commit only when the user has asked for commits (repo rule). Plan steps still say “Commit” — skip those steps unless the user explicitly requests commits, then batch at task boundaries.

## File map

| Path | Responsibility |
|------|----------------|
| `.github/workflows/release.yml` | Arm64 cross-compile + Windows packaging |
| `backend/internal/lightstate/store.go` | `BatchPatch` / `ResetAll` return changed-only DTOs |
| `backend/internal/lightstate/store_test.go` | Unit tests for changed-only |
| `backend/internal/store/scenes.go` | Scene region/batch `States` = changed-only |
| `backend/internal/store/scenes_test.go`, `store_test.go` | Update / add store-level expectations |
| `backend/internal/httpapi/capture_models.go` | Parse `marker` / `scale_hint` into `CreateParams` |
| `backend/internal/httpapi/capture_models_test.go` | Capture CreateParams in fake; assert wiring |
| `backend/internal/cvruntime/contract.go` | Comment path for mechanism B |
| `web/components/PythonRoutineHost.tsx` | Delete |
| `web/components/ShapeAnimationRoutineHost.tsx` | Delete |
| `web/public/dlm-python-scene-worker.mjs` | Delete |
| `web/lib/models.ts` | Capture status union |
| `web/lib/pythonSceneApiCatalog.ts`, `pythonRoutineSamples.ts` | Comment sync targets |
| Design / requirements markdown under `docs/` | Align prose with code |

---

### Task 1: Fix release workflow (High)

**Files:**
- Modify: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: none
- Produces: correct `GOOS=linux GOARCH=arm64` build line; packaging step that does not self-`cp` the Windows exe

- [ ] **Step 1: Patch the arm64 build line**

In `.github/workflows/release.yml`, change the cross-compile block so the first binary is explicitly arm64:

```yaml
          GOOS=linux  GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_linux_arm64 ./cmd/server
          GOOS=linux  GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_linux_amd64 ./cmd/server
          GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/dlm_windows_amd64.exe ./cmd/server
```

- [ ] **Step 2: Fix Windows packaging**

In the “Package release archives” step, remove:

```yaml
          cp dist/dlm_windows_amd64.exe dist/dlm_windows_amd64.exe
```

Replace with a comment only, e.g.:

```yaml
          # windows/amd64 — bare binary already in dist/; CV runtime bundle support pending
```

Keep the upload `files:` list unchanged (`dist/dlm_windows_amd64.exe` still exists from the build step).

- [ ] **Step 3: Sanity-check the YAML locally**

Run:

```bash
# Confirm no self-cp remains and arm64 env is set
rg -n 'GOARCH=arm64|cp dist/dlm_windows' .github/workflows/release.yml
```

Expected: one `GOARCH=arm64` on the arm64 `go build` line; no `cp dist/dlm_windows_amd64.exe dist/dlm_windows_amd64.exe`.

- [ ] **Step 4: Commit** (only if user requested commits)

```bash
git add .github/workflows/release.yml
git commit -m "$(cat <<'EOF'
fix(ci): cross-compile Pi release as arm64 and drop Windows self-cp

EOF
)"
```

---

### Task 2: `lightstate.BatchPatch` returns changed-only (High, REQ-041)

**Files:**
- Modify: `backend/internal/lightstate/store.go` (`BatchPatch`)
- Modify: `backend/internal/lightstate/store_test.go`
- Modify: `backend/internal/store/store_test.go` (idempotent batch length expectation)

**Interfaces:**
- Consumes: existing `BatchPatch(modelID string, ids []int, patch Patch) ([]DTO, bool, error)`
- Produces: same signature; `[]DTO` contains **only** lights whose triple changed; on full noop returns empty slice + `allUnchanged=true`

- [ ] **Step 1: Write the failing unit test**

Add to `backend/internal/lightstate/store_test.go`:

```go
func TestBatchPatch_ReturnsOnlyChangedLights(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 3) // adjust to the real Ensure/Create API used in this package's tests
	on := true
	// light 0 already default off — patching On:false is a no-op for that light
	// light 1 will change when On:true
	ids := []int{0, 1}
	out, allUnchanged, err := s.BatchPatch("m1", ids, Patch{On: &on})
	if err != nil {
		t.Fatal(err)
	}
	if allUnchanged {
		t.Fatal("expected at least one write")
	}
	if len(out) != 1 || out[0].ID != 1 || !out[0].On {
		t.Fatalf("want only changed light 1, got %+v", out)
	}
}
```

Adapt `EnsureModel` / setup to match existing helpers in `store_test.go` (read the file; do not invent a different constructor).

Also add a full-noop case expecting `len(out)==0` and `allUnchanged==true`.

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/lightstate/ -run TestBatchPatch_ReturnsOnlyChangedLights -v
```

Expected: FAIL because current code appends every id.

- [ ] **Step 3: Minimal implementation**

In `BatchPatch`, only append to `out` when `!rowUnchanged`:

```go
		if !rowUnchanged {
			allUnchanged = false
			sl[lightID] = lightTriple{on: on, color: color, brightnessPct: br}
			out = append(out, DTO{ID: lightID, On: on, Color: color, BrightnessPct: br})
		}
```

Do **not** append on noop rows.

- [ ] **Step 4: Update existing store test that expects full lists on noop**

In `backend/internal/store/store_test.go` `TestStore_BatchPatchLightStates`, change the idempotent assertion from `len(st3) != 2` to `len(st3) != 0` (with `allNoop` still true).

- [ ] **Step 5: Run tests**

```bash
cd backend && go test ./internal/lightstate/ ./internal/store/ -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit** (only if user requested commits)

```bash
git add backend/internal/lightstate/store.go backend/internal/lightstate/store_test.go backend/internal/store/store_test.go
git commit -m "$(cat <<'EOF'
fix(lights): return only changed lights from BatchPatch

EOF
)"
```

---

### Task 3: `lightstate.ResetAll` returns changed-only (High, REQ-041)

**Files:**
- Modify: `backend/internal/lightstate/store.go` (`ResetAll`)
- Modify: `backend/internal/lightstate/store_test.go`
- Modify: `backend/internal/store/store_test.go` (`TestStore_ResetAllLightStates_noopWhenAlreadyDefault`)

**Interfaces:**
- Consumes: `ResetAll(modelID string) ([]DTO, bool, error)`
- Produces: `[]DTO` = lights that were not already at defaults (empty when `allDefault`)

- [ ] **Step 1: Write failing test**

```go
func TestResetAll_ReturnsOnlyChangedLights(t *testing.T) {
	// model with 2 lights; set light 0 away from default; leave light 1 at default
	// ResetAll should return only id 0
}
```

Follow existing setup patterns in `lightstate/store_test.go`.

- [ ] **Step 2: Run to verify fail**

```bash
cd backend && go test ./internal/lightstate/ -run TestResetAll_ReturnsOnlyChangedLights -v
```

- [ ] **Step 3: Implement**

Rewrite `ResetAll` so that when not `allDefault`, it walks lights, resets non-default ones, and appends only those that changed; when `allDefault`, return `nil`/empty + `true`.

- [ ] **Step 4: Fix `TestStore_ResetAllLightStates_noopWhenAlreadyDefault`**

Expect `len(states)==0` (or nil) with `unchangedAll==true`. Keep `TestStore_ResetAllLightStates` expecting `len==2` when both lights were non-default.

- [ ] **Step 5: Run tests**

```bash
cd backend && go test ./internal/lightstate/ ./internal/store/ -count=1
```

- [ ] **Step 6: Commit** (only if user requested commits)

```bash
git add backend/internal/lightstate/store.go backend/internal/lightstate/store_test.go backend/internal/store/store_test.go
git commit -m "$(cat <<'EOF'
fix(lights): return only changed lights from ResetAll

EOF
)"
```

---

### Task 4: Scene bulk patches return changed-only States (High, REQ-041)

**Files:**
- Modify: `backend/internal/store/scenes.go` (`patchSceneLightsByRegion` ~1185–1239, `PatchSceneLightsBatch` ~1325–1395)
- Modify: `backend/internal/store/scenes_test.go` (idempotent `len(res2.States) != 2` → `0`)

**Interfaces:**
- Consumes: existing `SceneBulkPatchResult{UpdatedCount, States, UnchangedAll}`
- Produces: `States` contains only lights whose triple changed; full noop → empty `States` + `UnchangedAll=true`
- Downstream: `routineengine` already notifies with `res.States` — no engine change once store is fixed

- [ ] **Step 1: Write failing test**

In `scenes_test.go`, add a test that:
1. Creates a 2-light model in a scene.
2. Sets light 0 on, leaves light 1 off/default.
3. Calls `PatchSceneLightsScene` (or cuboid covering both) with `On: true`.
4. Asserts `!UnchangedAll`, `UpdatedCount==1`, `len(States)==1`, `States[0]` is light 0.

Also keep/adjust the existing idempotent test to expect `len(States)==0`.

- [ ] **Step 2: Run to verify fail**

```bash
cd backend && go test ./internal/store/ -run 'TestPatchScene.*Changed|TestStore_SceneBulk' -v
```

(Use the actual new test name.)

- [ ] **Step 3: Implement**

In both loops, only `append` to `updated` when `!rowUnchanged` (same for batch path).

- [ ] **Step 4: Run full store + httpapi light tests**

```bash
cd backend && go test ./internal/store/ ./internal/httpapi/ ./internal/routineengine/ -count=1
```

Expected: PASS. Fix any test that assumed `States` lists every matched light on noop or mixed writes.

- [ ] **Step 5: Commit** (only if user requested commits)

```bash
git add backend/internal/store/scenes.go backend/internal/store/scenes_test.go
git commit -m "$(cat <<'EOF'
fix(scenes): emit changed-only States for bulk light patches

EOF
)"
```

---

### Task 5: Wire capture `marker` + `scale_hint` (Medium)

**Files:**
- Modify: `backend/internal/httpapi/capture_models.go`
- Modify: `backend/internal/httpapi/capture_models_test.go`

**Interfaces:**
- Consumes: multipart form fields `marker`, `scale_hint`; `reconstruct.CreateParams{Marker *cvruntime.Marker, ScaleHint *float64}`
- Produces: non-empty `CreateParams` when fields present; default marker `DICT_4X4_50` / `0.1`

- [ ] **Step 1: Extend the fake to record CreateParams**

In `capture_models_test.go`, change `fakeReconstructCtrl`:

```go
type fakeReconstructCtrl struct {
	jobs   map[string]*reconstruct.Job
	last   reconstruct.CreateParams
	next   int
}

func (f *fakeReconstructCtrl) Create(_ context.Context, files []io.Reader, _ []string, p reconstruct.CreateParams) (string, error) {
	f.last = p
	// ... existing body ...
}
```

- [ ] **Step 2: Write failing HTTP test**

Post multipart with ≥2 tiny dummy files + `marker=true` + `scale_hint=1.5`. Assert:

- `fake.last.Marker != nil`
- `fake.last.Marker.Dictionary == "DICT_4X4_50"`
- `fake.last.Marker.EdgeLengthM == 0.1`
- `fake.last.ScaleHint != nil && *fake.last.ScaleHint == 1.5`

Follow existing multipart helpers in the same test file.

- [ ] **Step 3: Run to verify fail**

```bash
cd backend && go test ./internal/httpapi/ -run Capture -v
```

Expected: FAIL (`CreateParams{}` still empty).

- [ ] **Step 4: Implement parsing in `postModelsCapture`**

After collecting files, before `Create`:

```go
	params := reconstruct.CreateParams{}
	if m := strings.TrimSpace(r.FormValue("marker")); m == "true" || m == "1" {
		params.Marker = &cvruntime.Marker{Dictionary: "DICT_4X4_50", EdgeLengthM: 0.1}
	}
	if sh := strings.TrimSpace(r.FormValue("scale_hint")); sh != "" {
		v, err := strconv.ParseFloat(sh, 64)
		if err != nil || !(v > 0) || math.IsNaN(v) || math.IsInf(v, 0) {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "scale_hint must be a positive finite number")
			return
		}
		params.ScaleHint = &v
	}
	jobID, err := a.reconstruct.Create(r.Context(), fileReaders, fileNames, params)
```

Add imports: `math`, `strconv` as needed. Use `r.FormValue` only after `ParseMultipartForm`.

- [ ] **Step 5: Run tests**

```bash
cd backend && go test ./internal/httpapi/ -count=1
```

- [ ] **Step 6: Commit** (only if user requested commits)

```bash
git add backend/internal/httpapi/capture_models.go backend/internal/httpapi/capture_models_test.go
git commit -m "$(cat <<'EOF'
fix(capture): forward marker and scale_hint into reconstruct jobs

EOF
)"
```

---

### Task 6: Remove orphaned browser routine hosts + capture status type (Low)

**Files:**
- Delete: `web/components/PythonRoutineHost.tsx`
- Delete: `web/components/ShapeAnimationRoutineHost.tsx`
- Delete: `web/public/dlm-python-scene-worker.mjs`
- Modify: `web/lib/pythonSceneApiCatalog.ts` (file header comment)
- Modify: `web/lib/pythonRoutineSamples.ts` (file header comment)
- Modify: `web/lib/models.ts` (`CaptureJobStatus`)
- Modify: `web/public/dlm-python-editor-worker.mjs` (drop “same CDN as scene-worker” wording if present)

**Interfaces:**
- Consumes: none (components are unused)
- Produces: no production imports of deleted files; status union matches backend

- [ ] **Step 1: Confirm nothing imports the hosts**

```bash
rg -n 'PythonRoutineHost|ShapeAnimationRoutineHost|dlm-python-scene-worker' web --glob '!out/**'
```

Expected: only the files about to be deleted / comments.

- [ ] **Step 2: Delete the three files**

```bash
rm web/components/PythonRoutineHost.tsx \
   web/components/ShapeAnimationRoutineHost.tsx \
   web/public/dlm-python-scene-worker.mjs
```

Do not hand-edit `web/out/` or `backend/internal/webdist/` (regenerated by `release:sync` / `run.sh`).

- [ ] **Step 3: Update comments + status type**

In `web/lib/models.ts`, set:

```ts
export type CaptureJobStatus = "pending" | "running" | "succeeded" | "failed";
```

Point catalog/sample comments at `backend/internal/routineengine/bootstrap.py` and this catalog — not the deleted worker.

- [ ] **Step 4: Run frontend tests**

```bash
cd web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 5: Commit** (only if user requested commits)

```bash
git add -A web/components web/public web/lib
git commit -m "$(cat <<'EOF'
chore(web): remove unused in-browser routine hosts; align capture statuses

EOF
)"
```

---

### Task 7: Align design + requirements docs (Medium/Low)

**Files:**
- Modify: `docs/design/backend-service.md` (§3.8 cone pose if present, §3.12 bounds + `scenes` table, §3.2 capture `marker=true` note if table lists the field)
- Modify: `docs/design/frontend.md` (§4.9 framing)
- Modify: `docs/design/backend-lights-and-automation.md` (§3.17 signatures/samples/stop; §3.19–§3.20 WLED elision; §3.23 `marker=true` default)
- Modify: `docs/design/deployment.md` (§6.9 mechanism B; §6.10 README → user guide)
- Modify: `docs/design/overview.md` (REQ-030 / scene-worker wording)
- Modify: `docs/requirements/requirements.md` (REQ-004 / §12 packaging prose + index “in a nutshell” if needed)
- Modify: `backend/internal/cvruntime/contract.go` (package comment paths)

**Interfaces:**
- Consumes: final code behavior from Tasks 1–6
- Produces: docs that do not contradict code; no `§` renumbering

- [ ] **Step 1: §3.12 / §4.9 bounds**

Replace the `+1 m` / mins-at-0 AABB text with: per-scene persisted `margin_m` (default `0.3`), applied on all six sides, min corner reduced then clamped ≥ 0 (cite §3.15 / `GetSceneDimensions`). Add `margin_m` to the logical `scenes` table in §3.12. Update §4.9 framing to tight light bounds + `margin_m` / framing margin (not `[0,0,0]…(Mmax+1)`).

- [ ] **Step 2: §3.17 Python API + stop**

Change the canonical method table and sample snippets from keyword args to positional `patch`, e.g.:

```text
scene.set_lights_in_sphere(center, radius, {"on": True, "color": colour, "brightness_pct": 100})
```

Forced-stop: cooperative cancel via context → OS kill (`CommandContext` / SIGKILL on Unix) within REQ-040’s 2 s bound; remove SIGTERM→SIGKILL staging as a required sequence.

- [ ] **Step 3: §3.19–§3.20 WLED**

State that push elision is commit/model-level (only call `PushModel` when logical model state changed). Remove or rewrite `lastApplied[idx]` as a non-implemented future optimization.

- [ ] **Step 4: §6.9 / §6.10 / packaging requirements**

- §6.9: mechanism B — Linux release archive contains binary + sibling `runtime/cv/`; no extraction under `DLM_DATA_DIR`.
- §6.10: README is short landing page; download/run/`systemd` documented in `docs/userguide/` (link those files).
- `requirements.md`: Linux may ship as `.tar.gz` (binary + CV runtime); Windows may be a bare `.exe`. Soften “single program file” wording without renumbering REQ codes. Update feature-index nutshell rows for REQ-004/043 if they claim “one single program file” only.

- [ ] **Step 5: Capture marker default + overview + contract.go**

Document `marker=true` → `DICT_4X4_50` / `0.1` m. Fix overview.md scene-worker runtime claims. Fix `contract.go` package comment to say sibling `runtime/cv/` (not `dist/cvruntime/...` as the *runtime* path; build output path may still be mentioned as build-time).

- [ ] **Step 6: Cone pose**

Where §3.8 describes the sample cone, state Y-up (base in xz at y=0, apex at `(0, h, 0)`) or “equivalent consistent pose as in `samples/cone.go`”.

- [ ] **Step 7: Grep for stale phrases**

```bash
rg -n 'Mmax\+1|lastApplied|SIGTERM|DLM_DATA_DIR/runtime/cv|dlm-python-scene-worker|on=True, color=' docs/design docs/requirements backend/internal/cvruntime/contract.go
```

Expected: no stale contradictions for the findings above (SIGTERM only if intentionally historical; prefer none in normative §3.17 stop text).

- [ ] **Step 8: Commit** (only if user requested commits)

```bash
git add docs/design docs/requirements backend/internal/cvruntime/contract.go
git commit -m "$(cat <<'EOF'
docs: align design and requirements with review-fix code

EOF
)"
```

---

### Task 8: Final verification

**Files:** none new

- [ ] **Step 1: Backend tests**

```bash
cd backend && go test ./... -count=1
```

Expected: all PASS.

- [ ] **Step 2: Frontend tests**

```bash
cd web && npm test && npm run lint
```

Expected: all PASS.

- [ ] **Step 3: Release workflow spot-check**

```bash
rg -n 'GOOS=linux\s+GOARCH=arm64|cp dist/dlm_windows_amd64.exe dist/dlm_windows' .github/workflows/release.yml
```

Expected: arm64 env present; self-`cp` absent.

- [ ] **Step 4: Orphan grep**

```bash
rg -n 'PythonRoutineHost|ShapeAnimationRoutineHost|dlm-python-scene-worker' web --glob '!out/**' || true
```

Expected: no hits (or only historical mentions outside `web/` that were updated in Task 7).

---

## Self-review (plan vs spec)

| Spec section | Task |
|--------------|------|
| §1 Release | Task 1 |
| §2 SSE changed-only | Tasks 2–4 |
| §3 Marker / scale_hint | Task 5 (+ doc note in Task 7) |
| §4 Doc alignment | Task 7 |
| §5 Code cleanup | Task 6 |
| Success criteria / verification | Task 8 |

No TBD placeholders. Types/names match existing packages (`reconstruct.CreateParams`, `cvruntime.Marker`, `SceneBulkPatchResult`). Commit steps gated by user rule.
