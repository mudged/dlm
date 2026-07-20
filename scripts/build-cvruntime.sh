#!/usr/bin/env bash
# build-cvruntime.sh — Build a self-contained OpenCV+Python CV runtime bundle.
#
# Packaging mechanism: B — sibling asset alongside the release binary.
# See docs/design/architecture.md §3.23.1 and docs/engineering/cv-runtime.md.
#
# Usage (from repository root):
#   GOOS=linux GOARCH=arm64  scripts/build-cvruntime.sh
#   GOOS=linux GOARCH=amd64  scripts/build-cvruntime.sh
#
# Output:
#   dist/cvruntime/<goos>_<goarch>/
#     python/          — self-contained CPython (python-build-standalone)
#     reconstruct.py   — CV entrypoint (copied from backend/python/ or placeholder)
#
# On-disk footprint (approximate, after pip install):
#   linux/amd64:  ~200 MB  (CPython ~25 MB + OpenCV ~100 MB + NumPy ~60 MB + misc)
#   linux/arm64:  ~220 MB  (arm64 OpenCV wheels are slightly larger)
#
# Pinned component versions — update all three together and re-test:
#   python-build-standalone release:  20241002
#   CPython:                          3.12.7
#   opencv-python-headless:           4.10.0.84
#   numpy:                            2.1.3
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"

PBS_RELEASE="20241002"
PY_VERSION="3.12.7"
OPENCV_VERSION="4.10.0.84"
NUMPY_VERSION="2.1.3"

# Map GOOS/GOARCH → python-build-standalone triple and manylinux platform tag.
# Use the short manylinux2014_* tag (not the dual manylinux_2_17_*.manylinux2014_*
# form): pip's --platform matcher rejects the dual tag and then reports
# "No matching distribution" for opencv-python-headless / numpy.
case "${GOOS}_${GOARCH}" in
  linux_arm64)
    PBS_TRIPLE="aarch64-unknown-linux-gnu"
    PBS_MACHINE="aarch64"
    WHEEL_PLATFORM="manylinux2014_aarch64"
    ;;
  linux_amd64)
    PBS_TRIPLE="x86_64-unknown-linux-gnu"
    PBS_MACHINE="x86_64"
    WHEEL_PLATFORM="manylinux2014_x86_64"
    ;;
  *)
    echo "ERROR: unsupported target ${GOOS}/${GOARCH}" >&2
    echo "       Supported: linux/arm64, linux/amd64 (Windows support planned post-spike)" >&2
    exit 1
    ;;
esac

PBS_URL="https://github.com/astral-sh/python-build-standalone/releases/download/${PBS_RELEASE}/cpython-${PY_VERSION}+${PBS_RELEASE}-${PBS_TRIPLE}-install_only.tar.gz"
OUTDIR="${REPO_ROOT}/dist/cvruntime/${GOOS}_${GOARCH}"

echo "==> CV runtime bundle: target=${GOOS}/${GOARCH}  output=${OUTDIR}"

# ---- Download python-build-standalone --------------------------------------

TMPDIR_BUILD="$(mktemp -d)"
trap 'rm -rf "${TMPDIR_BUILD}"' EXIT

echo "==> Downloading python-build-standalone ${PY_VERSION} (${PBS_TRIPLE})..."
curl -fsSL --retry 3 --retry-delay 5 \
  "${PBS_URL}" -o "${TMPDIR_BUILD}/python.tar.gz"

echo "==> Extracting into ${OUTDIR}/python/ ..."
rm -rf "${OUTDIR}"
mkdir -p "${OUTDIR}"
tar -xzf "${TMPDIR_BUILD}/python.tar.gz" -C "${OUTDIR}/"

PYTHON="${OUTDIR}/python/bin/python3"

echo "==> Upgrading pip inside the bundle..."
NATIVE_MACHINE="$(uname -m)"

# pip must run natively; only invoke the bundle's Python when it is the native arch.
if [[ "${PBS_MACHINE}" == "${NATIVE_MACHINE}" ]]; then
  "${PYTHON}" -m pip install --quiet --upgrade pip
else
  echo "    Skipping pip upgrade (cross-compile: host=${NATIVE_MACHINE} target=${PBS_MACHINE})"
fi

# ---- Install OpenCV + NumPy ------------------------------------------------

echo "==> Installing opencv-python-headless==${OPENCV_VERSION} numpy==${NUMPY_VERSION}..."

if [[ "${PBS_MACHINE}" == "${NATIVE_MACHINE}" ]]; then
  # Native: run the bundled Python directly.
  "${PYTHON}" -m pip install \
    --quiet \
    --no-cache-dir \
    "opencv-python-headless==${OPENCV_VERSION}" \
    "numpy==${NUMPY_VERSION}"
else
  # Cross-compile: use the host Python to download pre-built target-platform wheels
  # and install them into the bundle's site-packages.  No emulation required.
  echo "    Cross-platform wheel install (host=${NATIVE_MACHINE} → target=${PBS_MACHINE})"
  PY_SITE="${OUTDIR}/python/lib/python3.12/site-packages"
  # --abi cp312 is required so pip accepts opencv's cp37-abi3 wheels and
  # numpy's cp312 wheels for the target; without it resolution returns none.
  python3 -m pip install \
    --quiet \
    --no-cache-dir \
    --target "${PY_SITE}" \
    --platform "${WHEEL_PLATFORM}" \
    --implementation cp \
    --python-version "312" \
    --abi cp312 \
    --only-binary=:all: \
    "opencv-python-headless==${OPENCV_VERSION}" \
    "numpy==${NUMPY_VERSION}"
fi

# ---- Copy the CV entrypoint ------------------------------------------------

RECONSTRUCT_SRC="${REPO_ROOT}/backend/python/reconstruct.py"
if [[ -f "${RECONSTRUCT_SRC}" ]]; then
  cp "${RECONSTRUCT_SRC}" "${OUTDIR}/reconstruct.py"
  echo "==> Copied entrypoint from ${RECONSTRUCT_SRC}"
else
  echo "==> WARNING: ${RECONSTRUCT_SRC} not found — writing WI-05 placeholder stub."
  cat > "${OUTDIR}/reconstruct.py" << 'PYEOF'
"""reconstruct.py — WI-05 placeholder stub.

Replace this file with the real implementation from WI-05.
Until then, every Run() call will return status=failed with an explanatory message.
"""
import json
import sys

print(json.dumps({
    "status": "failed",
    "light_count": 0,
    "lights": [],
    "missing": [],
    "low_confidence": [],
    "error": "reconstruct.py is a WI-05 placeholder; real implementation pending",
}))
PYEOF
fi

# ---- Summary ---------------------------------------------------------------

SIZE="$(du -sh "${OUTDIR}" | cut -f1)"
echo "==> Bundle ready: ${OUTDIR}/ (${SIZE})"
echo "    Interpreter: ${OUTDIR}/python/bin/python3"
echo "    Entrypoint:  ${OUTDIR}/reconstruct.py"
echo "    Set DLM_CV_RUNTIME_DIR=${OUTDIR} to use this bundle in dev."
