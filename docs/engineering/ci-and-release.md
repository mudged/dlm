# Continuous integration and releases

Workflows live under [`.github/workflows/`](../../.github/workflows).

## Continuous integration (`ci.yml`)

| Workflow file | When it runs | Purpose |
|---------------|--------------|---------|
| `ci.yml` | Pull requests and pushes to `main` | **Web lint and test** (`npm ci`, `npm run lint`, `npm run test` in `web/`), **Backend tests** (`go test ./...` in `backend/`), and **Build Go binary with embedded UI** (`npm ci` + `npm run release:sync` in `web/`, then `go build` in `backend/`). |

### Branch protection (maintainers)

Configure the repository’s branch protection rule for **`main`** to **require status checks** before
merge. Add these **exact** job names so all three gates must pass:

1. **Web lint and test**
2. **Backend tests**
3. **Build Go binary with embedded UI**

Exact labels appear in GitHub under the PR checks UI; they match the `name:` field on each job in
`ci.yml`.

## Release binaries (`release.yml`, maintainers)

The **`release.yml`** workflow runs when you push a tag matching **`v*`** (for example **`v1.2.3`**).
It:

1. Builds the Next.js static export and copies it into **`backend/internal/webdist/`** (`npm ci` +
   `npm run release:sync` in `web/`).
2. Builds CV runtime bundles for **linux/arm64** and **linux/amd64** (`scripts/build-cvruntime.sh` —
   see [`cv-runtime.md`](cv-runtime.md)).
3. Cross-compiles **`CGO_ENABLED=0`** binaries into **`dist/`** and packages Linux releases as
   **`.tar.gz` archives** (binary + `runtime/cv/` sibling):
   - **`dlm_linux_arm64.tar.gz`** — Raspberry Pi / Linux ARM64
   - **`dlm_linux_amd64.tar.gz`** — Linux x86_64
   - **`dlm_windows_amd64.exe`** — Windows x86_64 (bare binary; CV runtime bundle pending)
4. Creates or updates the **GitHub Release** for that tag and attaches those three assets
   (`softprops/action-gh-release`).

Cut a release locally:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Then confirm the workflow run succeeded and the assets appear on the Releases page.

### CV runtime and CI

Pull-request **CI does not build the CV runtime bundle** (the download and pip install are large and
slow). Backend unit tests use a **fake CV runner** and stub runtime directories, so `go test ./...`
passes without a real bundle. Integration tests that invoke the real OpenCV child process should be
gated behind **`DLM_CV_RUNTIME_DIR`** being set (skip otherwise). See [`cv-runtime.md`](cv-runtime.md).
