# CV runtime bundle (3D reconstruction)

The Go server runs light-position reconstruction from uploaded videos using a **self-contained
OpenCV + Python runtime**. No separate Python or OpenCV install is required on the operator host; the
bundle is completely independent of any Python used for scene routines (those still use the host's
`python3` — see [`environment-and-api.md`](environment-and-api.md)).

## Packaging mechanism

**Mechanism B** (chosen for the MVP): the runtime ships as a directory **alongside** the Go binary
inside the Linux release archive, rather than being embedded in the binary. This keeps individual
binary sizes manageable while still providing a self-contained single download.

Release archive layout:

```
dlm_linux_arm64.tar.gz
├── dlm_linux_arm64          # the Go server binary
└── runtime/
    └── cv/
        ├── python/          # python-build-standalone CPython 3.12.x
        └── reconstruct.py   # CV entrypoint
```

Extract the archive and run `./dlm_linux_arm64` from the extracted directory; the binary discovers
`runtime/cv/` as a sibling automatically.

Mechanism A (embedding everything inside one binary via `//go:embed` and extracting on first use to
`DLM_DATA_DIR/runtime/cv/<version>/`) is a possible future upgrade. The resolver in
`backend/internal/cvruntime/resolve.go` is designed to be switched cheaply.

## Runtime resolution order

When the Go server starts a reconstruction job it locates the CV runtime directory as follows:

1. **`DLM_CV_RUNTIME_DIR`** environment variable — an absolute path to an extracted bundle. Use this
   in development and CI to point at a local build.
2. **`<dir-of-executable>/runtime/cv/`** — the sibling directory from the release archive
   (mechanism B).

If neither is found, reconstruction fails with a descriptive error message.

## Environment variables (camera capture)

| Variable | Default | Effect |
|----------|---------|--------|
| **`DLM_CV_RUNTIME_DIR`** | *(unset)* | Absolute path to a CV runtime bundle directory (contains `python/` and `reconstruct.py`). Overrides sibling resolution. Required for local reconstruction testing when no release layout is present. |
| **`DLM_CAPTURE_DWELL_MS`** | `1000` | Milliseconds each light stays on during a device **capture sweep** (Devices screen). Must be a positive integer. |
| **`DLM_DATA_DIR`** | `data` | Root data directory. Reconstruction jobs write temporary uploads under **`DLM_DATA_DIR/runtime/capture/<job_id>/`**; these directories are cleaned up when a job finishes or is discarded. |

## Building a bundle locally

```bash
# linux/amd64 (native on a typical dev machine)
GOOS=linux GOARCH=amd64 scripts/build-cvruntime.sh

# linux/arm64 (cross-compiled on an amd64 host)
GOOS=linux GOARCH=arm64 scripts/build-cvruntime.sh
```

Output lands in `dist/cvruntime/<goos>_<goarch>/`. Set `DLM_CV_RUNTIME_DIR` to that path for local
testing:

```bash
export DLM_CV_RUNTIME_DIR="${PWD}/dist/cvruntime/linux_amd64"
cd backend && go run ./cmd/server
```

To iterate on the CV script without rebuilding the full bundle, you can also point
`DLM_CV_RUNTIME_DIR` at any directory that contains a working `python/` tree and `reconstruct.py`
(for example a local venv with `opencv-python-headless` installed, laid out to match the bundle
shape).

## Pinned versions

| Component | Version |
|-----------|---------|
| python-build-standalone release | 20241002 |
| CPython | 3.12.7 |
| opencv-python-headless | 4.10.0.84 |
| numpy | 2.1.3 |

Update all four together and re-run the build script to verify the bundle works end-to-end before
pushing a release tag.

## On-disk footprint

| Platform | Approximate size |
|----------|-----------------|
| linux/amd64 | ~200 MB |
| linux/arm64 | ~220 MB |

The bulk is the OpenCV wheel (~100 MB) and NumPy (~60 MB). The Python interpreter itself is ~25 MB.

## CI and tests

Pull-request **CI does not build the CV runtime bundle** (the download and pip install are large and
slow). Backend unit tests use a **fake CV runner** and stub runtime directories, so `go test ./...`
passes without a real bundle. Any integration test that invokes the real OpenCV child process should
be gated behind **`DLM_CV_RUNTIME_DIR`** being set (skip otherwise). Set it to the output of
`scripts/build-cvruntime.sh` when running those tests locally.

## Windows

CV runtime bundling for Windows is not yet implemented (amd64 wheel compatibility is still open). The
`dlm_windows_amd64.exe` release asset is a bare binary; reconstruction features are not available on
Windows until the bundle is added.
