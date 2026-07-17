# Security notes (§9)

This page collects the baseline security expectations for dlm. It is a hobbyist app that
usually runs on a home network, so the goal here is *sensible defaults that bound risk on a small
Raspberry Pi* — not enterprise hardening.

Part of the [dlm architecture](architecture.md); see the [glossary](glossary.md) for unfamiliar terms.

## 9. Security notes (baseline)

**In plain terms:** dlm ships as one process on a home LAN with no login in the MVP (minimum
viable product). The main risks are (1) untrusted *input* — CSV files, video uploads, and huge
batch requests that could exhaust the Pi's memory — and (2) untrusted *code* — the Python routines
a user writes. The rules below keep those in check.

### Keep the attack surface small

- **Prefer same-origin in production.** When the UI and API share one origin, the browser sends
  requests without cross-origin checks, so you can avoid opening up CORS (Cross-Origin Resource
  Sharing, the browser rules that decide whether a page may call an API on another origin). CORS is
  mainly needed for local development, when the Next.js dev server runs on a different port.
- **Secrets live in environment variables only.** Never bake secrets into the client bundle beyond
  genuinely public constants — anything shipped to the browser is readable by anyone.
- **No container is required as a trust boundary.** The product does not assume it runs inside
  Docker (REQ-004); don't design security that depends on a container sandbox.

### Bound memory and denial-of-service (DoS) risk on the Pi

The Pi has little RAM, so large requests are both a memory risk and a DoS (denial-of-service —
overwhelming the server so it can't serve real users) risk. Reject oversized requests *early*,
before fully parsing them.

- **CSV upload (`POST /api/v1/models`):** enforce a maximum request body size (for example with
  Go's `http.MaxBytesReader`, or a server-level limit). Make it large enough for a 1000-row CSV but
  small enough to bound memory.
- **Scene batch updates (`PATCH /api/v1/scenes/{id}/lights/state/batch`):** the body can carry one
  JSON object per light — up to ~1000 per scene in the worst case. Enforce a reasonable maximum body
  size (a few MB, well below the Pi's spare RAM) and reject oversize requests with `413` or `400`
  before a full parse if needed.
- **Video uploads (REQ-048 / REQ-049, `POST /api/v1/models/capture`):** these are large multi-file
  uploads. Enforce a per-request *and* per-file maximum size (well below the Pi's RAM and disk
  headroom), restrict to an allowed container/codec list, and **stream** files to a work directory
  under `DLM_DATA_DIR` rather than buffering in memory. Reject oversize uploads early (`413` / `400`).
  Clean up work directories when a job reaches a terminal state, and sweep stale directories on
  startup. Treat uploaded videos as untrusted input to the computer-vision (CV) child process:
  apply timeouts and bounded concurrency.

### Treat user-authored code as untrusted

- **User Python routines (REQ-022):** production routines run in a supervised `python3` child
  process (§3.17). That child's `scene` shim reaches the local API over loopback (127.0.0.1, the
  machine's own address) — see §3.15 — or an in-process equivalent. Assume the user's code can reach
  the local API and the filesystem within that process. Harden with timeouts, resource limits, and
  the body/rate limits in this section. Never expose admin tokens or third-party API keys to routine
  source or bundles.
  - The optional browser Pyodide path (Python compiled to WebAssembly) is for editor lint/format
    only (§3.17); it must **not** run production routines (REQ-038).
- **Bundled CV runtime (REQ-048):** the shipped OpenCV runtime (§3.23.1) is *product-controlled*
  code, not operator-authored, so it is a lower risk than user Python. Even so, run it as a
  supervised child with a `ctx` (context) timeout/cancellation and captured stderr. It is **not** a
  path for executing user-supplied code (unlike §3.17).

### Data and destructive operations

- **SQLite file:** treat `DLM_DB_PATH` as persistent storage on the Pi (SD card or USB). Operators
  should back it up with ordinary file-backup practices.
- **Factory reset (`POST /api/v1/system/factory-reset`):** this is destructive and, in the MVP, is
  unauthenticated. Treat network exposure accordingly (see §3.14) — for example, keep the app off the
  public internet.
