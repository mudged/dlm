# Advanced setup and developer notes

This document is for people who want more detail than the main [README](../README.md): performance tuning, development workflows, cross-compilation, and API-oriented tips.

## One-command build and run (repository standard)

From the **repository root** (same as README), the supported path is:

```bash
./scripts/run.sh
```

This installs web dependencies when needed (`npm ci` if `web/node_modules` is missing), runs `npm run release:sync` in `web/` to build the Next.js static export into the Go embed tree, then starts the Go server with `go run ./cmd/server` in `backend/` (the script **replaces** the shell via `exec`).

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/).

### Environment variables for `run.sh`

| Variable | Effect |
|----------|--------|
| `DLM_SKIP_NPM_CI=1` | Skip `npm ci`; only runs `release:sync` (fastest when `node_modules` is already correct). |
| `DLM_FORCE_NPM_CI=1` | Always run `npm ci` before `release:sync` (clean install, CI, or after `package-lock.json` changes). |

If `web/node_modules` already exists and neither skip nor force is set, the script skips `npm ci` and only runs `release:sync`.

### Why `run.sh` can feel slow

Typical bottlenecks:

1. **`npm ci`** — reinstalls `web/node_modules` from the lockfile (network + disk). Mitigations: rely on skip-when-present behavior, `DLM_SKIP_NPM_CI=1`, or a dev environment with dependencies pre-installed.
2. **`next build`** (inside `release:sync`) — production compile and static export; often on the order of tens of seconds to a few minutes depending on CPU. Next caches under `web/.next/`; the second build is usually faster.
3. **`go run ./cmd/server`** — compiles Go; subsequent runs reuse the build cache.

To measure: `time ./scripts/run.sh` or run `time npm ci`, `time npm run release:sync`, and `time go run ./cmd/server` separately under `web/` and `backend/`.

## Prerequisites (versions)

- **Go:** The module in `backend/go.mod` declares the required language version (currently **1.25.x**). Install a matching toolchain from [go.dev/dl](https://go.dev/dl/). If your system ships an older Go, install a newer release and ensure `go version` matches.
- **Node.js:** Use an **Active LTS** release for building `web/` (see [nodejs.org](https://nodejs.org/)). Node is for building the UI; a small board can run only the compiled Go binary if you build elsewhere and copy artifacts.
- **Python 3** (optional): For **Python scene routines** on the host, install Python 3 and ensure `python3` is on `PATH`. Other features do not require it.

## Manual two-step run (same result as the script)

```bash
cd web && npm run release:sync
cd ../backend && go run ./cmd/server
```

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/) — HTML and `/api/v1/...` are served by the same process.

### Cross-compile for Raspberry Pi (arm64)

From `backend/`:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server
```

## Development: two processes (not what you ship)

1. **Go server** (API + embedded static UI after a sync, or rebuild after UI changes):

   ```bash
   cd backend
   go run ./cmd/server
   ```

2. **Next.js dev server** (hot reload; proxies API to the Go port):

   ```bash
   cd web
   npm run dev
   ```

Open [http://localhost:3000](http://localhost:3000) (or `http://127.0.0.1:3000`). The dev app rewrites `/api/v1` to the Go backend (default `http://127.0.0.1:8080`).

**Server-Sent Events:** `next dev` can buffer `text/event-stream`, so the UI may open `EventSource` directly against the Go origin (`http://127.0.0.1:8080`). The Go server defaults `CORS_ALLOWED_ORIGINS` to common Next dev origins (`:3000` and `:8000`, localhost and 127.0.0.1). If the API is not on `http://127.0.0.1:8080`, set `DLM_BACKEND_ORIGIN` and `NEXT_PUBLIC_DLM_API_ORIGIN` in `web/.env.local` (see `web/.env.local.example`). Override with `CORS_ALLOWED_ORIGINS` for other page origins; `CORS_ALLOWED_ORIGINS=-` disables CORS headers entirely.

### If the 3D view does not update during a routine (dev)

1. In DevTools → **Network**, find the **`events`** request whose URL ends with `/api/v1/scenes/<scene-id>/lights/events`. In **EventStream**, **data** lines should look like `{"seq":1,"deltas":[...]}`. If you only see unrelated payloads, select the request whose URL contains `/lights/events`.
2. **Go on a non-default port:** set `NEXT_PUBLIC_DLM_API_ORIGIN` and `DLM_BACKEND_ORIGIN` to that base URL.
3. **Embedded UI only:** after UI changes, run `cd web && npm run release:sync` and restart Go.

## Command reference

| Area | Command |
|------|---------|
| One-shot build + run | `./scripts/run.sh` from repo root (`DLM_FORCE_NPM_CI=1` or `DLM_SKIP_NPM_CI=1` as above) |
| Sync export → embed tree | `cd web && npm run release:sync` |
| Go tests | `cd backend && go test ./...` |
| Web lint | `cd web && npm run lint` |
| Web tests | `cd web && npm test` |
| Web export only | `cd web && npm run build` (output in `web/out/`) |
| Download Go deps | `cd backend && go mod download` |
| Install frontend deps | `cd web && npm ci` |

## Continuous integration (GitHub Actions)

Workflows live under `.github/workflows/`.

| Workflow file | When it runs | Purpose |
|---------------|----------------|---------|
| `ci.yml` | Pull requests and pushes to `main` | **Web lint and test** (`npm ci`, `npm run lint`, `npm run test` in `web/`), **Backend tests** (`go test ./...` in `backend/`), and **Build Go binary with embedded UI** (`npm ci` + `npm run release:sync` in `web/`, then `go build` in `backend/`). |

### Branch protection (maintainers)

Configure the repository’s branch protection rule for **`main`** to **require status checks** before merge. Add these **exact** job names so all three gates must pass:

1. **Web lint and test**
2. **Backend tests**
3. **Build Go binary with embedded UI**

Exact labels appear in GitHub under the PR checks UI; they match the `name:` field on each job in `ci.yml`.

### Release binaries (maintainers)

The **`release.yml`** workflow runs when you push a tag matching **`v*`** (for example **`v1.2.3`**). It:

1. Builds the Next.js static export and copies it into **`backend/internal/webdist/`** (`npm ci` + `npm run release:sync` in `web/`).
2. Builds CV runtime bundles for **linux/arm64** and **linux/amd64** (`scripts/build-cvruntime.sh`).
3. Cross-compiles **`CGO_ENABLED=0`** binaries into **`dist/`** and packages Linux releases as **`.tar.gz` archives** (binary + `runtime/cv/` sibling):
   - **`dlm_linux_arm64.tar.gz`** — Raspberry Pi / Linux ARM64
   - **`dlm_linux_amd64.tar.gz`** — Linux x86_64
   - **`dlm_windows_amd64.exe`** — Windows x86_64 (bare binary; CV runtime bundle pending)

4. Creates or updates the **GitHub Release** for that tag and attaches those three assets (`softprops/action-gh-release`).

Cut a release locally:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Then confirm the workflow run succeeded and assets appear on the Releases page.

## CV runtime bundle (3D reconstruction)

The Go server runs light-position reconstruction from uploaded videos using a **self-contained OpenCV+Python runtime**. No separate Python or OpenCV install is required on the operator host; the bundle is completely independent of any Python used for scene routines (those still use the host's `python3`).

### Packaging mechanism

**Mechanism B** (chosen for the MVP): the runtime ships as a directory **alongside** the Go binary inside the Linux release archive, rather than being embedded in the binary. This keeps individual binary sizes manageable while still providing a self-contained single download.

Release archive layout:

```
dlm_linux_arm64.tar.gz
├── dlm_linux_arm64          # the Go server binary
└── runtime/
    └── cv/
        ├── python/          # python-build-standalone CPython 3.12.x
        └── reconstruct.py   # CV entrypoint
```

Extract the archive and run `./dlm_linux_arm64` from the extracted directory; the binary discovers `runtime/cv/` as a sibling automatically.

Mechanism A (embedding everything inside one binary via `//go:embed` and extracting on first use to `DLM_DATA_DIR/runtime/cv/<version>/`) is a possible future upgrade. The resolver in `backend/internal/cvruntime/resolve.go` is designed to be switched cheaply.

### Runtime resolution order

When the Go server starts a reconstruction job it locates the CV runtime directory as follows:

1. **`DLM_CV_RUNTIME_DIR`** environment variable — an absolute path to an extracted bundle. Use this in development and CI to point at a local build.
2. **`<dir-of-executable>/runtime/cv/`** — the sibling directory from the release archive (mechanism B).

If neither is found, reconstruction fails with a descriptive error message.

### Environment variables (camera capture)

| Variable | Default | Effect |
|----------|---------|--------|
| **`DLM_CV_RUNTIME_DIR`** | *(unset)* | Absolute path to a CV runtime bundle directory (contains `python/` and `reconstruct.py`). Overrides sibling resolution. Required for local reconstruction testing when no release layout is present. |
| **`DLM_CAPTURE_DWELL_MS`** | `1000` | Milliseconds each light stays on during a device **capture sweep** (Devices screen). Must be a positive integer. |
| **`DLM_DATA_DIR`** | `data` | Root data directory. Reconstruction jobs write temporary uploads under **`DLM_DATA_DIR/runtime/capture/<job_id>/`**; these directories are cleaned up when a job finishes or is discarded. |

### Building a bundle locally

```bash
# linux/amd64 (native on a typical dev machine)
GOOS=linux GOARCH=amd64 scripts/build-cvruntime.sh

# linux/arm64 (cross-compiled on an amd64 host)
GOOS=linux GOARCH=arm64 scripts/build-cvruntime.sh
```

Output lands in `dist/cvruntime/<goos>_<goarch>/`. Set `DLM_CV_RUNTIME_DIR` to that path for local testing:

```bash
export DLM_CV_RUNTIME_DIR="${PWD}/dist/cvruntime/linux_amd64"
cd backend && go run ./cmd/server
```

To iterate on the CV script without rebuilding the full bundle, you can also point `DLM_CV_RUNTIME_DIR` at any directory that contains a working `python/` tree and `reconstruct.py` (for example a local venv with `opencv-python-headless` installed, laid out to match the bundle shape).

### Pinned versions

| Component | Version |
|-----------|---------|
| python-build-standalone release | 20241002 |
| CPython | 3.12.7 |
| opencv-python-headless | 4.10.0.84 |
| numpy | 2.1.3 |

Update all four together and re-run the build script to verify the bundle works end-to-end before pushing a release tag.

### On-disk footprint

| Platform | Approximate size |
|----------|-----------------|
| linux/amd64 | ~200 MB |
| linux/arm64 | ~220 MB |

The bulk is the OpenCV wheel (~100 MB) and NumPy (~60 MB). The Python interpreter itself is ~25 MB.

### CI and tests

Pull-request **CI does not build the CV runtime bundle** (the download and pip install are large and slow). Backend unit tests use a **fake CV runner** and stub runtime directories, so `go test ./...` passes without a real bundle.

Any integration test that invokes the real OpenCV child process should be gated behind **`DLM_CV_RUNTIME_DIR`** being set (skip otherwise). Set it to the output of `scripts/build-cvruntime.sh` when running those tests locally.

### Windows

CV runtime bundling for Windows is not yet implemented (amd64 wheel compatibility is tracked in the work-item backlog). The `dlm_windows_amd64.exe` release asset is a bare binary; reconstruction features are not available on Windows until the bundle is added.

---

## High-throughput light updates (API-oriented)

- Prefer **batch** routes when updating many lights (`PATCH /api/v1/models/{id}/lights/state/batch` and scene bulk routes in `docs/architecture.md`) instead of one HTTP request per light.
- Reuse HTTP keep-alive. For many parallel clients, terminating TLS + HTTP/2 in a reverse proxy while proxying HTTP/1.1 to Go is a common pattern (see architecture doc).
- After commits, the server emits **SSE** on `GET /api/v1/models/{id}/lights/events` and `GET /api/v1/scenes/{id}/lights/events` (`text/event-stream`). Subscribers typically refetch authoritative state after each event.
- If SSE drops behind a proxy, increase **`HTTP_WRITE_TIMEOUT_SEC`** or set `0` for no write deadline when debugging; the handler extends deadlines periodically.

## Embedded database and CSV upload

- **SQLite** is embedded (pure Go, no separate database server). The database file is created automatically at **`data/dlm.db`** by default (see server configuration if you override the data directory).
- **CSV upload** for models: `POST /api/v1/models` expects the file field name **`file`** (not `csv`). Header must be exactly `id,x,y,z`. Light IDs must be **0-based** sequential integers (`0`, `1`, `2`, …).

For system design and diagrams, see **[architecture.md](architecture.md)**.
