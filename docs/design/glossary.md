# Glossary

Plain-language definitions of the terms, acronyms, and product-specific words used across the dlm
design docs. If a word in another page is unfamiliar, look here first.

Part of the [dlm architecture](architecture.md).

## Product concepts

- **dlm / Domestic Light & Magic** — this project: a hobbyist app for designing and animating home
  LED light installations.
- **Model** — a saved 3D shape made of individual lights, each with a position (`x, y, z` in metres)
  and an id. Imported from a CSV file. Example built-in models: sphere, cube, cone.
- **Light** — one addressable LED in a model. It has a fixed position and a changeable **state**.
- **Light state / "state triple"** — the three changeable properties of a light: `on` (boolean),
  `color` (hex string like `#ff0000`), and `brightness_pct` (0–100). Stored in memory, not the
  database (see *LightStateStore*).
- **Scene** — one or more models placed together in a shared coordinate space so you can control and
  animate them as a group.
- **Routine** — an automation that changes lights over time. There are exactly two kinds:
  a **Python scene script** (user-written Python) and a **shape animation** (a configured, built-in
  animation). Routines run on the server, with no browser needed.
- **Device** — a physical light controller on the network, currently **WLED**. A device can be
  assigned to one model so that changing the model's lights also drives the real LEDs.
- **Capture (light sequence)** — a server-driven sweep that turns on one physical LED at a time
  (~1 second each), used to figure out which LED is where during camera reconstruction.
- **Reconstruction** — turning two or more uploaded videos of the blinking capture sweep into 3D
  positions for each light, so a model can be built "from video".

## Backend / Go terms

- **Single binary** — the whole product ships as one executable file with the web UI embedded
  inside it; there is no separate web server or database server to install.
- **embed / `//go:embed`** — a Go feature that bakes files (here, the built web UI) into the compiled
  binary so they ship as part of the executable.
- **cgo** — Go's mechanism for calling C code. dlm avoids it ("pure Go, no cgo") because it
  complicates cross-compiling for the Raspberry Pi.
- **cross-compile** — building an executable on one machine (e.g. an x86 laptop) that runs on a
  different CPU/OS (e.g. the Pi's ARM64).
- **GOOS / GOARCH** — Go environment variables that select the target operating system and CPU
  architecture for a build (e.g. `GOOS=linux GOARCH=arm64`).
- **ARM64 / arm64** — the 64-bit CPU architecture used by the Raspberry Pi 4.
- **SQLite** — a small database that lives in a single file; dlm uses the pure-Go driver
  `modernc.org/sqlite`.
- **LightStateStore** — the in-memory, authoritative store of every light's current state
  (on/color/brightness). It is *not* persisted to SQLite; the database only stores geometry
  (positions). See §3.9, §3.21.
- **Transaction** — a database operation that is all-or-nothing: either every change commits, or none
  do.
- **Migration** — code that updates the database schema (e.g. adding a column) when the app starts.
- **Seed** — inserting default data (the sample models and routines) into an empty database.
- **FK (foreign key)** — a database column that points at a row in another table (e.g. a light's
  `model_id` points at its model).
- **UUID** — a long, unique identifier string used as a primary key.
- **RFC3339** — a standard text format for timestamps, e.g. `2026-07-17T12:00:00Z`.
- **loopback / 127.0.0.1** — the machine's own network address; used when the server calls its own
  API (for example, a Python routine calling back into the API).
- **SIGTERM / SIGKILL** — Unix signals used to ask a child process to stop politely (SIGTERM) or to
  force-kill it (SIGKILL) if it ignores the polite request.
- **ticker / `time.Ticker`** — a Go timer that fires on a fixed interval; used to drive shape
  animation frames on the server.
- **slog** — Go's standard structured logging package.

## Web / frontend terms

- **Next.js** — the React framework used to build the UI. Here it is used only as a build tool.
- **Static export (`output: 'export'`)** — building the Next.js app into plain HTML/JS/CSS files with
  no Node.js server needed at runtime. Those files are embedded into the Go binary.
- **Client Component / `"use client"`** — a React component that runs in the browser. dlm uses these
  for interactivity because there is no Node server at runtime.
- **SSR / RSC (Server-Side Rendering / React Server Components)** — React features that render on a
  Node server per request. dlm does **not** use these at runtime (they'd require Node on the Pi).
- **Hydration** — the browser attaching React's interactivity to server-rendered/exported HTML after
  the page loads.
- **Tailwind CSS** — a utility-class CSS framework used for styling; `dark:` variants power dark mode.
- **three.js** — a JavaScript library for drawing 3D graphics with WebGL. dlm uses it to show the
  lights in 3D.
- **WebGL** — the browser API for GPU-accelerated 3D graphics that three.js sits on top of.
- **InstancedMesh** — a three.js object that draws many copies of the same shape (e.g. hundreds of
  light spheres) efficiently in one draw call.
- **LineSegments** — a three.js object used to draw the thin wires connecting consecutive lights.
- **Raycaster / raycast / picking** — shooting a ray from the cursor into the 3D scene to find which
  object (which light) is under the pointer.
- **OrbitControls** — a three.js helper that lets the user rotate/pan/zoom the camera with mouse or
  touch.
- **Emissive** — a material property that makes a surface look like it is glowing (used for "on"
  lights).
- **AABB (axis-aligned bounding box)** — the smallest box, aligned to the x/y/z axes, that contains a
  set of points; used to frame the view and draw the faint boundary cuboid.
- **CodeMirror** — the in-browser code editor component used for writing Python routines.
- **IA (information architecture)** — how features/pages are organized and navigated.

## HTTP / networking terms

- **SSE (Server-Sent Events)** — a one-way stream from server to browser over HTTP. dlm uses it to
  push live light changes so clients don't have to poll repeatedly. See §3.18.
- **`deltas[]`** — the list of just-changed lights included in each SSE event, so clients update only
  what changed instead of rebuilding everything.
- **elision** — skipping a write when the new value equals the current value (a no-op), to avoid
  needless work and network traffic to devices. See §3.19.
- **PATCH** — the HTTP method for a partial update (change some fields, leave others alone).
- **error envelope** — dlm's standard JSON error shape: `{ "error": { "code", "message", "details"? } }`.
- **CORS (Cross-Origin Resource Sharing)** — browser rules governing whether a page may call an API
  on a different origin. Mostly relevant to local dev.
- **SSRF (server-side request forgery)** — an attack where a server is tricked into making requests
  to unintended internal addresses. dlm guards device `base_url` values with an allowlist. See §3.20.
- **reverse proxy** — a server (e.g. nginx) placed in front of the app that forwards requests to it;
  optional for dlm.
- **SPA (single-page application)** — a web app that loads once and updates in the browser without
  full page reloads.

## Devices, hardware, and computer vision

- **WLED** — popular open-source firmware for ESP-based controllers that drive addressable LED
  strips; dlm talks to it over its HTTP JSON API.
- **mDNS** — "multicast DNS", a way to discover devices on a local network by name; an optional way
  to find WLED devices.
- **Raspberry Pi 4 Model B** — the small, low-power target computer dlm is designed to run on
  (ARM64, limited CPU/RAM).
- **systemd unit** — a Linux service definition that starts dlm on boot and restarts it if it
  crashes.
- **CV (computer vision)** — extracting information from images/video; here, finding blinking lights
  and computing their 3D positions.
- **OpenCV** — the computer-vision library used by the reconstruction runtime.
- **Triangulation** — computing a 3D point from where it appears in two or more camera views.
- **Fiducial marker** — a printed pattern (like an ArUco/AprilTag square) placed in the scene to give
  the cameras a known reference for scale, position, and orientation. Optional.

## Traceability

- **REQ-NNN** — a stable requirement code from `docs/requirements/`. These codes appear throughout
  the design as traceability anchors and are cited from source-code comments. Never renumber or reuse
  them. The full map from requirements to design sections is in
  [appendix-traceability.md](appendix-traceability.md).
- **§N.N** — a design section number (e.g. §3.18). These are stable anchors cited from code and other
  docs. The [architecture hub](architecture.md) has an index mapping every § to the file it lives in.
