# Appendix: requirements traceability

This appendix maps requirement codes (`REQ-NNN`, defined in [`../requirements/`](../requirements))
to the design sections that realize them. It is a reference for verifying coverage, not a tutorial —
if you are learning the system, start with the [architecture hub](architecture.md) and the topic
pages it links to.

Part of the [dlm architecture](architecture.md); see the [glossary](glossary.md) for unfamiliar terms.

The per-requirement "architectural response" table lives in §1 (see
[overview.md](overview.md)). The blocks below are the dense cross-reference notes that used to sit at
the top of the architecture document; they list, for each group of requirements, exactly which
sections implement them.

## REQ-001–REQ-033 (summary)

Go plus an embedded Next.js static export. Models, scenes, and routines are stored in SQLite;
per-light output state lives in `LightStateStore` (in memory, not SQLite — REQ-039). The three.js
views (§4.7, §4.9) use a server-push observer path per REQ-041. The scene spatial API is §3.15.

Routine automation (Python and shape animation) runs inside `internal/routineengine`:

- Supervised `python3` with loopback HTTP calls to §3.15 (see §3.17).
- Go `time.Ticker` shape simulation (§3.17.2).

Runs progress headless, without any browser (REQ-038). Next.js may use Pyodide or similar **only**
for REQ-022 editor lint/format — never for applying routine effects. Stop latency is bounded by
REQ-040 (§3.17 / §3.17.2).

## REQ-034–REQ-042

- **REQ-034** — faint boundary cuboid plus `margin_m`, using the same line style as the inter-light
  wire (§4.7). Covered in §3.12–§3.13, §4.7, §4.9.
- **REQ-035–REQ-039** — add WLED devices, 1:1 device↔model assignment, the Devices UI (§4.15),
  routine automation on the Go service (`internal/routineengine`, §3.16–§3.17.2), in-memory light
  state (§3.3, §3.9, §3.21), and WLED push (§3.19–§3.20).
- **REQ-040** — routine stop within 2 seconds; see §3.17 and §3.17.2.
- **REQ-041** — SSE / WebSocket-class push plus delta apply for the shipped three.js viewports;
  see §3.18, §4.3, §4.7, §4.9, §4.13, §4.14.
- **REQ-042** — detect an active routine per scene, and resync the run controls and three.js after
  navigating back to a page; see §4.16, §4.9 / §4.13 / §4.14 (mount effects), §3.2
  (`GET …/routines/runs`, `POST …/start` 409 details), and §8.23.
- **REQ-043–REQ-046** — multi-platform release binaries, GitHub Actions CI and release, deployment
  runtime prerequisites, and README operator instructions; see §3.4 and §6.6–§6.12.

## REQ-047–REQ-050 (camera capture)

- **REQ-047** — a built-in device capture light sequence: a sequential, one-light-on-for-~1-second
  sweep, started/stopped from the Devices screen, run server-side, turning all lights off on
  stop/completion within the REQ-040 two-second bound. See §3.22, §4.15, §8.24.
- **REQ-048** — camera-based 3D reconstruction from two or more uploaded video feeds, using a bundled
  OpenCV runtime that needs no separate Python install (distinct from the user-routine Python of
  REQ-045). It is invoked as a supervised child by `internal/reconstruct`. See §3.23, §3.23.1, §6.9,
  §8.25.
- **REQ-049** — the create-model-from-video flow: upload → async reconstruct job → review detected
  lights → confirm to persist a normal model (per REQ-005 / REQ-007), plus an optional printable
  fiducial marker. See §3.23.2, §4.17, §8.25.
- **REQ-050** — red, blue, then green flashes of every light at the start and end of the capture
  sweep, used to number bulbs in each uploaded clip (including a clip that only caught the closing
  flash when the light count is provided). Clips with no signal are named and dropped. See §3.22.1,
  §3.23.3, §4.15, §4.17, §8.24, §8.25.
- **REQ-051** — dim-room capture. A lamp or window already visible before the opening flash is ignored; if that opening flash was missed, a light still visible after the closing flash is ignored instead. Each bulb must still be the brightest new thing. A clip with a bookend but no quiet moment of the room is named and dropped. A clip with no bookend still uses the darkest-sample background. See §3.23.
- **REQ-052** — three printable capture markers. Clips join only when they share one, or a clip shows two and links them. A clip that cannot be joined is named. Too little joined footage fails the job. See §3.23.2, §4.17.
