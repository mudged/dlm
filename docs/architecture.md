# Architecture

This document defines technical structure and deployment for the product described in `docs/requirements.md` (**REQ-001–REQ-030**). It satisfies **REQ-001** (Go + Next.js + Tailwind), **REQ-002** (responsive, client-interactive UI), **REQ-003** (Raspberry Pi 4 Model B, **ARM64**, resource awareness), **REQ-004** (**one runnable executable** per release target; **no** mandatory Docker/OCI/compose packaging at this stage), **REQ-005** (wire light **chain** model: **CSV** interchange, **metadata**, **at most two** logical neighbors per light by **consecutive id**—**endpoints** have **one**), **REQ-006** (list / view / delete / create via CSV upload; **delete blocked** with **409** when the model is **referenced** by **one or more** **scenes**, with an **actionable** **error** payload), **REQ-007** (server-side CSV validation and actionable errors), **REQ-008** (single command to build UI and run the Go server locally), **REQ-009** (default **sphere**, **cube**, **cone** samples: lights on **exterior** nominal surfaces with **even** coverage of **face planes** (**cube**) and **surface area** (**sphere** / **cone**), **not** **edge-only** or **single-curve-only** layouts; consecutive spacing **0.05–0.10 m**; **500–1000** lights each; **≤ 0.03 m** surface deviation; **~2 m** characteristic size), **REQ-010** (**three.js** **3D** view on **model detail**: **every** light as **2 cm** sphere; **all** lights drawn for **n ≤ 1000**; **wire** segments **only** between **consecutive** **ids**; segments **`#D0D0D0`** at **85% transparency** (**15% opacity**), **subtler** than spheres; **hover** / **touch** disclosure of **id** and **coordinates**), **REQ-011** (**REST** **read/write** of per-light **on/off**, **hex** **#RRGGBB** colour, **brightness** **0–100%**, persisted in **SQLite**), **REQ-012** (**3D** **spheres** reflect state: **on** = **opaque** **filled** **colour** × **brightness**; **off** = **`#D0D0D0`** at **85% transparency** matching segments; **timely** UI sync after **writes**), **REQ-013** (**model detail**: **paginated** light **list** with **page-size** control and **go to id**; **multi-select** and **bulk apply** via **batch** **PATCH**), **REQ-014** (**default** all lights **off**, **`#FFFFFF`**, **100%** brightness on create/seed; **reset** control + **authoritative** API to restore that state for **all** lights in **one** action), **REQ-015** (**scenes**: **named** **composite** **3D** **space**; **create** with **≥ 1** **model** **and** **server-computed** **integer** **offsets** **(no** **client-supplied** **offsets** **on** **create)**; **canonical** **`lights.x/y/z`** **unchanged** **by** **scenes**; **derived** **`sx/sy/sz`** **=** **canonical** **+** **offset** **for** **API**/**UI** **in** **scene** **context**; **non-negative** **containment** **+** **≥ 1 m** **margin** **beyond** **max** **extent** **per** **axis**; **three.js** **composite** **view** **reusing** **REQ-010**/**012** **per** **model**; **add**/**remove**/**optional** **placement** **edit** **after** **create**; **last** **model** **→** **confirm** **then** **delete** **scene**; **+X** **default** **for** **models** **added** **after** **create**), **REQ-016** (**camera reset** on **model** and **scene** **three.js** views: **client-only** restore of **default** **framing**; **no** **persisted** **data** **changes**), **REQ-017** (**Options** **UI** with **factory reset**: **confirmed** **wipe** of **all** **SQLite** **application** **data** **then** **re-seed** **§3.8** **samples** **to** **match** **fresh** **install**), **REQ-018** (**application** **shell**: **light**/**dark** **themes** (**white**/dark **grey** **surfaces**, **dark**/ **white** **primary** **text**); **collapsible** **left** **nav** **with** **burger** **toggle**; **branding** **title** **`Domestic Light & Magic`** **+** **Font Awesome** **classic** **regular** **lightbulb** **logo**; **Font Awesome** **icons** **on** **action** **buttons** **throughout** **the** **UI** **per** **§4.11**; **default** **theme** **follows** **`prefers-color-scheme`** **until** **the** **user** **overrides** **with** **a** **persisted** **choice**), **REQ-019** (**model** **and** **scene** **three.js** **views**; **fixed** **dark-grey** **WebGL** **backdrop** **in** **both** **shell** **themes**; **§4.7** **and** **§4.9**), **REQ-020** (**scene spatial API** for **dimensions**, **all-lights retrieval**, **cuboid/sphere filtering**, **cuboid/sphere bulk updates**, **whole-scene** **`…/state/scene`**, **and** **per-light** **`…/state/batch`** **in** **§3.15**, **all** in **scene coordinates** without rewriting canonical model coordinates; **§3.15**, **§8.15**), and **REQ-021** (**scene** **routines**: **persisted** **definitions**; **start**/**stop** **runs** **per** **scene**; **in-process** **scheduler** **uses** **only** **§3.15** **scene** **state** **endpoints** **(shared** **`store`** **with** **HTTP** **handlers)** **—** **not** **`/models/.../lights`** **for** **routine** **effects**; **one** **active** **run** **per** **scene**; **first** **type** **`random_colour_cycle_all`**; **§3.14**, **§3.16**, **§4.9**, **§4.12**, **§8.16**), and **REQ-022** (**Python** **scene** **routines**: **persisted** **`python_scene_script`** **definitions** **with** **source** **text**; **in-browser** **editor** **+** **novice** **instructional** **copy** **(**~12-year-old** **beginner** **Python** **tone** **on** **this** **page**)** **+** **API** **docs** **placement**/**interaction** **per** **REQ-024** **in** **§4.13**; **unified** **run**/**live** **viewport** **per** **REQ-027**; **run** **loop** **and** **forced** **stop** **in** **the** **browser** **(Pyodide** **worker)** **with** **`scene`** **shim** **calling** **§3.15** **via** **`fetch`** — **§3.17**, **§8.17**), **and** **REQ-023** (**new** **routine** **creation** **UI**: **one** **labeled** **`type`** **`<select>`** **(or** **accessible** **equivalent)** **listing** **every** **creatable** **kind** **including** **`python_scene_script`**; **no** **standalone** **primary** **Python-only** **create** **button** **—** **§4.12**, **§4.13**), **REQ-024** (**Python** **routine** **page** **—** **complete** **`scene`** **API** **reference** **directly** **below** **the** **code** **editor:** **selectable** **entry**, **detail** **+** **sample** **with** **brief** **`#`** **comments**, **insert-example** **at** **caret** **or** **end** **of** **buffer** **—** **§4.13**), **REQ-025** (**default** **`python_source`** **for** **new** **`python_scene_script`** **—** **sphere-region** **colour** **change** **demo:** **§4.13**), **REQ-026** (**`scene.width`**, **`scene.depth`**, **`scene.height`** **from** **`GET …/dimensions`→`size`:** **§3.15**, **§3.17**, **§4.13**), **REQ-027** (**Python** **routine** **page** **—** **one** **unified** **region** **for** **scene** **target**, **Start**/**Stop**, **live** **`SceneLightsCanvas`**, **reset** **scene** **lights**, **reset** **camera** **—** **no** **duplicate** **run** **vs** **debug** **sections:** **§4.13**, **§8.18**), **and** **REQ-028** (**three.js** **on** **lights:** **emissive** **glow** **scaled** **monotonically** **with** **`brightness_pct`** **0–100** **with** **strong** **read** **at** **100%**; **§4.7**, **§4.9**, **§4.13** **unified** **live** **viewport**; **§6.5** **client** **GPU** **notes**), **and** **REQ-029** (**high-throughput** **light** **state** **updates** **at** **hundreds** **of** **lights** **and** **multi-Hz** **aggregate** **change** **rates:** **aggregate** **write** **APIs** **§3.10**/**§3.15**, **HTTP** **keep-alive** **and** **optional** **reverse-proxy** **HTTP/2** **§3.18**, **observer** **path** **via** **bounded** **polling** **§4.7**/**§4.9** **with** **documented** **optional** **SSE** **§3.18**/**§8.19**, **Pi**/**SQLite** **limits** **§6**/**§9**), **and** **REQ-030** (**Python** **`scene.random_hex_colour()`** **—** **synchronous** **helper** **with** **no** **HTTP** **that** **returns** **a** **REQ-011**-valid **`#`+6** **lowercase** **hex** **digits** **using** **the** **same** **distribution** **as** **`"#%06x"** **%** **`random.randrange(0x1000000)`** **in** **Pyodide**; **§3.17**, **§4.13** **catalog** **+** **completions**, **`web/lib/pythonSceneApiCatalog.ts`** **/** **`dlm-python-scene-worker.mjs`** **stay** **in** **sync**).

## Architectural resolution: REQ-004 (single binary) vs Next.js

**Decision:** **Full Next.js SSR/Node at runtime is out of scope** for the shipped product. The open item in REQ-004 is resolved as follows:

- **Next.js + App Router + Tailwind** remain the **authoring** stack under `web/` (toolchain, components, styling).
- **Runtime** on the Pi is a **single Go process** that:
  - Serves the **JSON API** (`/health`, `/api/v1/...`).
  - Serves the UI as **static assets** produced by **`next build` with `output: 'export'`** (static HTML, JS, CSS), **embedded** in the binary via Go **`embed`**, or **baked in at link time** from a generated filesystem tree.
- **REQ-002 (reactive UI):** Achieved with **client-side React** (hydration and **`"use client"`** components) and browser **`fetch`** to **same-origin** `/api/v1/...` on the Go listener. **No** React Server Components that require a Node runtime at request time; any **async server-only** data patterns used during development MUST be replaced or duplicated with **client** fetches or **build-time** data before release.

This meets REQ-004 rule 1 (**no separate Node.js runtime** in the distribution) and aligns with Pi RAM constraints (**§6**).

---

## 1. Goals and constraints

| Requirement | Architectural response |
|-------------|-------------------------|
| REQ-001 | **Go** binary serves **HTTP API** + **embedded static UI** built from **Next.js + Tailwind** source in `web/`. |
| REQ-002 | **Mobile-first** Tailwind, **Client Components** for interactivity; **`fetch('/api/v1/…')`** from the browser to the same Go origin. |
| REQ-003 | Primary release target **linux/arm64** (Pi 4B, 64-bit OS); document **CPU/RAM**; **one** long-lived app process. |
| REQ-004 | **One executable file** per release target; assets **embedded** (or generated inside the process); **Docker/compose not** the canonical install path. |
| REQ-005 | **Domain types** + **CSV** contract; **ordered chain** adjacency (**id** **i** ↔ **i±1** only); **metadata** (**name**, **creation instant**) stored with each model; coordinates as **float64** in Go/API JSON. |
| REQ-006 | **REST JSON API** for models + **Next.js** client pages (list, detail, upload, delete); **multipart** upload for CSV; **409** **`model_in_scenes`** when **delete** **blocked** (**§3.13**). |
| REQ-007 | **Authoritative validation** in Go when ingesting CSV; **transactional** create (all-or-nothing); **400** responses with clear **error** envelope. |
| REQ-008 | **`scripts/run.sh`** (or documented equivalent) from repo root: **`npm run release:sync`** in `web/` then **`go run ./cmd/server`** in `backend/`; **README** documents exact invocation (**§3.7**). |
| REQ-009 | **`internal/samples`** builds three **ordered** light paths on each solid’s boundary (**500 ≤ n ≤ 1000**, **0.05–0.10 m** consecutive chords, **§3.8**) with **even** placement on **faces** / **surfaces** per **§3.8**; **`store.SeedDefaultSamples`** when **`models`** empty (**§3.8**). |
| REQ-010 | **`three`** direct; **client-only** detail: **InstancedMesh** (or equivalent **§4.7**) **ø 0.02 m** markers for **all n** lights (**no** LOD/decimation); **`LineSegments`** **i↔i+1** only (**chain**); segment colour **`#D0D0D0`**, **opacity 0.15** (**85%** transparent), thinner/subtler than spheres; **`Raycaster`** + **DOM** tooltip (**§4.7**). |
| REQ-011 | **REST** **`/api/v1/models/{id}/lights/...`** for **bulk** + **per-id** **state** **read** and **`PATCH`** **per** **light**; **validation** and **defaults** in **Go** (**§3.2**, **§3.3**, **§3.9**). **Integrators** **with** **high** **multi-light** **update** **rates** **also** **use** **§3.18** **(**REQ-029**)**. |
| REQ-012 | **three.js** **materials**: **on** = **opaque** **base** **colour** × **brightness** **+** **emissive** **glow** **(**REQ-028**)**; **off** = **`#D0D0D0`**, **opacity 0.15** (match **REQ-010** segments); **client** merges **API** state; **no** indefinite staleness after **writes** (**§4.7**, **§8.7**). |
| REQ-013 | **Model detail** **light table**: **client-side** **pagination** over **`GET` detail** payload; **page size** presets **25 / 50 / 100** (default **50**); **go to id** computes target page + **inline validation**; **checkbox** multi-select with **optional** **Shift+click** range on **current page**; **selection** **retained** across pages in **React state** (**`Set<number>`**) until **clear** or **navigate away**; **bulk apply** calls **`PATCH …/lights/state/batch`** (**§3.10**); **list** and **three.js** updated from **response** (**§4.8**, **§8.8**). |
| REQ-014 | **Insert defaults** **on=false**, **`color="#ffffff"`**, **`brightness_pct=100`** (**§3.9**); **`POST …/lights/state/reset`** (**§3.11**); **model detail** **Reset** button (non–hover-only) calls reset then merges **`states`** into **UI** + **§4.7** (**§4.6**, **§8.9**). |
| REQ-015 | **`scenes`** + **`scene_models`** (**§3.12**); **`POST /scenes`** **computes** **all** **offsets** **from** **ordered** **`model_id`** **list**; **`GET /models/{id}`** **unchanged** **(canonical** **coords)**; **`GET /scenes/{id}`** **returns** **`x,y,z`** **+** **`sx,sy,sz`** (**§3.13**); **composite** **three.js** **§4.9**; **409** **`model_in_scenes`** (**§3.2**, **§8.4**). |
| REQ-016 | **§4.7** / **§4.9**: **`Reset camera`** **control** **re-applies** **`applyDefaultFraming`** **(same** **pure** **function** **as** **initial** **load** **from** **current** **bounds**)**; **`OrbitControls`** **target** **+** **camera** **position** **+** **`update()`**; **does** **not** **call** **API**. **§4.10** **IA** **optional** **link** **only**—**primary** **placement** **adjacent** **to** **viewport**. |
| REQ-017 | **`POST /api/v1/system/factory-reset`** (**§3.2**, **§3.14**); **`store.FactoryReset`** **transaction** **deletes** **`routine_runs`** **+** **`routines`** **before** **scenes**/**models** **then** **`SeedDefaultSamples`**; **Options** **route** **§4.10** **+** **modal** **before** **POST**; **post-success** **navigate** **`/models`** **+** **success** **message** (**architectural** **default** **for** **REQ-017** **open** **question**). |
| REQ-018 | **§4.11**: **Tailwind** **`dark:`** **variant** on **`html`** (**add**/**remove** **`class="dark"`** **or** **`data-theme="dark"`** **per** **Tailwind** **v4**/**v3** **config**); **before** **any** **stored** **user** **choice**, **derive** **initial** **light**/**dark** **from** **`window.matchMedia('(prefers-color-scheme: dark)')`** **(or** **equivalent**)** **when** **available**; **fallback** **to** **light** **if** **the** **platform** **does** **not** **expose** **a** **scheme**; **after** **the** **user** **toggles** **theme**, **persist** **`light`** **or** **`dark`** **in** **`localStorage`** **key** **`dlm-theme`** **and** **re-apply** **on** **load** **before** **first** **paint** **(inline** **blocking** **script** **in** **`layout`** **recommended** **to** **avoid** **flash**)** **so** **the** **persisted** **value** **overrides** **`prefers-color-scheme`** **until** **cleared** **or** **changed**; **shell** **tokens**: **light** **`bg-white`** **+** **`text-gray-900`**; **dark** **`bg-gray-900`** **+** **`text-white`** (**dark** **grey** **background**, **not** **mandating** **`#000`**); **`AppShell`**: **header** **(burger,** **`faLightbulb`** **regular** **logo,** **exact** **title** **`Domestic Light & Magic`**, **theme** **toggle** **with** **icons)** **+** **collapsible** **left** **`<aside>`** **nav** **+** **`main`**; **Font Awesome** **Free** **`@fortawesome/fontawesome-svg-core`**, **`@fortawesome/react-fontawesome`**, **`@fortawesome/free-regular-svg-icons`**, **`@fortawesome/free-solid-svg-icons`**; **buttons** **and** **button-styled** **controls** **include** **a** **visible** **Font Awesome** **icon** **+** **accessible** **name**. |
| REQ-019 | **§4.7**, **§4.9**: **`ModelLightsCanvas`** **/** **`SceneLightsCanvas`** **set** **a** **fixed** **dark-grey** **clear** **/** **scene** **background** **and** **matching** **letterbox** **wrapper** **(see** **§4.7** **viewport** **subsection**)** **independent** **of** **`html`** **`dark`** **class**; **does** **not** **replace** **REQ-018** **shell** **tokens** **for** **`main`** **chrome**. |
| REQ-020 | **§3.2**, **§3.12**, **§3.13**, **§3.15**, **§8.15**: scene-space API for dimensions + full/filtered light retrieval (cuboid/sphere) + transactional bulk updates (cuboid/sphere + **`PATCH …/lights/state/scene`** + **`PATCH …/lights/state/batch`**), all computed from derived scene coordinates (**`sx/sy/sz`**) and validated with explicit geometry rules. |
| REQ-021 | **§3.2**, **§3.14**, **§3.15**, **§3.16**, **§4.9**, **§4.11**, **§4.12**, **§8.16**: routine definitions CRUD + scene-scoped start/stop + in-process **1** **s** scheduler **for** **`random_colour_cycle_all`** **only** (**§3.16** **skips** **`python_scene_script`** **runs**); automation uses **only** **§3.15** **`…/lights/state/*`** **handlers** **or** **the** **same** **`store`** **functions** **they** **call** (**not** **`/models/.../lights`** **for** **routine** **effects**); **one** active run per scene (**any** **type**); **409** on delete while active. |
| REQ-022 | **§3.2**, **§3.15**, **§3.16** (**schema** **+** **start/stop** **shared** **with** **REQ-021**), **§3.17** **(**incl.** **REQ-030** **local** **helpers**)**, **§4.9** (**link** **to** **editor**), **§4.11** (**nav**), **§4.12** (**list** **entry**), **§4.13** (**novice** **copy** **+** **REQ-024**/**REQ-027** **layout**), **§6.2** (**bundle** **size** **note**), **§7**, **§8.17**: **`python_scene_script`** **+** **`python_source`**; **PATCH** **routine**; **CodeMirror** **6** **(mandatory)** **—** **`EditorView`** **+** **`codemirror`** **`basicSetup`** **+** **`@codemirror/*`** **(see** **§4.13)**; **Pyodide** **worker** **`scene`** **→** **`fetch`** **§3.15** **(**network** **methods** **only**)**; **cooperative** **stop** **+** **`worker.terminate()`** **after** **bounded** **wait**. |
| REQ-023 | **§4.12** (**primary** **create** **path** **+** **`type`** **dropdown** **+** **Mermaid** **flow**), **§4.13** (**editor** **routes** **as** **post-create** **/ deep** **link** **only**), **§4.9** (**optional** **“Edit** **Python**”** **shortcut**), **§4.11** (**no** **nav** **entry** **Python-only** **create**); **HTTP** **unchanged** (**`POST /api/v1/routines`** **still** **carries** **`type`** **in** **JSON**). |
| REQ-024 | **§4.13** (**`<section id="python-scene-api-catalog">`** **immediately** **below** **editor** **—** **complete** **manifest** **from** **§3.17** **including** **REQ-030** **`scene.random_hex_colour`**; **picker** **per** **entry**; **detail** **+** **commented** **sample**; **insert** **at** **caret** **or** **EOF**). |
| REQ-025 | **§4.13** (**`PYTHON_ROUTINE_DEFAULT_SOURCE`** **in** **`web/`** **—** **applied** **when** **creating** **`python_scene_script`** **with** **empty** **`python_source`** **and** **optional** **“Reset** **template”**); **uses** **`scene.set_lights_in_sphere`** **with** **`on`**, **`color`**, **`brightness_pct`**; **SHOULD** **use** **`scene.random_hex_colour()`** **for** **the** **demo** **colour** **(**REQ-030**)** **instead** **of** **`import** **random`** **for** **that** **single** **pattern**. |
| REQ-026 | **§3.15** (**axis** **mapping** **for** **`size.width`**, **`size.height`**, **`size.depth`**), **§3.17** (**Python** **`scene.width`**, **`scene.height`**, **`scene.depth`** **from** **`GET …/dimensions`**), **§4.13** (**REQ-024** **API** **reference** **manifest**). |
| REQ-027 | **§4.13** (**unified** **panel:** **one** **scene** **`<select>`** **for** **run** **+** **`SceneLightsCanvas`**; **Start**/**Stop**; **Reset** **scene** **lights** **→** **`PATCH …/lights/state/scene`**, **Reset** **camera** **→** **`applyDefaultFraming`**), **§8.18**; **resolved:** **reset** **lights** **does** **not** **auto-stop** **run**; **viewport** **works** **with** **or** **without** **active** **run**. |
| REQ-028 | **§4.7** **(**emissive** **mapping** **+** **scene** **lights**)**, **§4.9** **(**composite** **reuse**)**, **§4.13** **(**Python** **unified** **live** **viewport**)**, **§6.5** **(**browser** **GPU** **on** **Pi** **and** **other** **clients**)**, **§8.5**, **§8.7** **(**material** **updates** **after** **PATCH**)**. |
| REQ-029 | **§3.2** **(**batch**/**bulk** **routes** **listed** **below**)**, **§3.10**, **§3.15**, **§3.18** **(**write** **path**, **connections**, **observer** **strategy**, **optional** **SSE**)**, **§4.3**, **§4.7**, **§4.9**, **§6**, **§7**, **§8.19**, **§9** **(**payload**/**rate** **limits**)**. |
| REQ-030 | **§3.17** (**`scene.random_hex_colour()`** **—** **local** **`random.randrange(0x1000000)`** **+** **`"#%06x"`** **formatting** **in** **Pyodide**), **§4.13** (**REQ-024** **manifest** **row** **+** **novice** **sample**; **`pythonRoutineCodemirror`** **completion**), **`web/lib/pythonSceneApiCatalog.ts`**, **`public/dlm-python-scene-worker.mjs`**. |

**Assumed Pi context:** Raspberry Pi 4 Model B, **64-bit OS**, **ARM64** userspace. **2–8 GB RAM** — with **no Node** at runtime, **4 GB** is practical for modest traffic; **off-device** `next export` builds recommended.

**Canonical production listener:** Single Go process on **one TCP address** (e.g. **`:8080`** or behind an **optional** reverse proxy on **`:80`/`443`** that forwards **all** paths to that one upstream). **No** second application process from the product distribution.

**Dev vs prod:** Local development **may** still run **`next dev`** alongside **`go run`** for fast iteration; that **two-process** pattern is **not** part of the **shipped** artifact (REQ-004).

---

## 2. Repository layout

Monorepo (single Git root), **one Go module** under `backend/`, **no `go.work`** unless later expanded.

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
      store/                  # persistence for models (SQLite, REQ-006); SeedDefaultSamples (REQ-009)
      webdist/                # holds embedded payload (see §3.5); populated by build, not hand-edited
        placeholder.txt       # optional tiny file so empty embed works in dev before first UI build
  web/                        # Next.js + Tailwind (source only for runtime; sibling of backend/)
    app/
    components/               # incl. AppShell, theme toggle, nav (REQ-018 §4.11); ModelLightsCanvas / SceneLightsCanvas (REQ-019 viz backdrop §4.7–§4.9)
    lib/
  docs/
```

**Boundaries:**

- **`backend/` MUST NOT import** TypeScript sources from `web/` — only **built static files** under `internal/webdist/` (or equivalent) prepared by the **build pipeline**.
- **`web/` MUST NOT** import Go.
- **Public HTTP contract:** **`GET /health`**, **`/api/v1/*`** JSON API; **static paths** for UI (`/`, `/_next/...`, assets) from embedded FS.

---

## 3. Golang service (runtime)

### 3.1 Module and packages

- **Module path:** `example.com/dlm/backend` (or successor); unchanged conceptually.
- **`cmd/server`:** Load config; construct **`http.Server`**; mount **API sub-router** and **static file server** from **`embed`**; graceful **SIGINT/SIGTERM** shutdown.
- **`internal/config`:** **Env-based** listen address, timeouts, optional **CORS** (primarily for **dev** when UI dev server uses another origin); production **same-origin** reduces CORS; **SQLite** path via **`DLM_DB_PATH`** and/or **`DLM_DATA_DIR`** (**§3.3**).
- **`internal/httpapi`:** Middleware (**request ID**, **slog**, **recover**, optional **CORS**); **JSON** handlers and error envelope `{ "error": { "code", "message", "details"? } }`; **models**, **light-state**, **scenes**, **scene-region** routes (**§3.15**), and **routines**/**routine-runs** (**§3.16**, **§3.17**) delegate to **`internal/store`** and **`internal/wiremodel`**; **routine** **scheduler** (**§3.16**) runs **inside** **`cmd/server`** (**tick** **handlers** **for** **`random_colour_cycle_all`** **only**; **`python_scene_script`** **runs** **are** **client-executed** **per** **§3.17**). **High-throughput** **light** **update** **and** **connection-reuse** **guidance** **for** **integrators** **and** **reverse-proxy** **deployment** **is** **§3.18** (**REQ-029**).
- **`internal/wiremodel`:** Parses and validates uploaded **CSV** per **§3.6**; returns structured errors for HTTP **400** responses (**REQ-007**).
- **`internal/store`:** **SQLite** repository (see **§3.3**, **§3.12**, **§3.16**); opened at process start; migrations or `CREATE IF NOT EXISTS` for schema (**REQ-006**); **idempotent default seed** when no models exist (**§3.8**, **REQ-009**); **scene** CRUD and **reference** checks for **model** **delete** (**REQ-015**, **REQ-006** rule **5**); **routines** (**incl.** **`python_source`** **for** **`python_scene_script`**, **§3.17**) and **`routine_runs`** persistence (**§3.16**).
- **`internal/samples`:** Pure functions that return **`[]wiremodel.Light`** (sequential **id**s from **0**) for the three canonical shapes: **on-surface** positions, **even** coverage per **§3.8**, consecutive **dᵢ** in band; no I/O (**REQ-009**).

### 3.2 HTTP surface

| Kind | Path | Purpose |
|------|------|--------|
| API | `GET /health` | Liveness; **no auth**. |
| API | `GET /api/v1/models` | List models (**metadata** + **light_count**). **REQ-006** |
| API | `POST /api/v1/models` | Create model: **`multipart/form-data`** with **`name`** (string) + **`file`** (CSV). **REQ-006/007** |
| API | `GET /api/v1/models/{id}` | Model **metadata** + ordered **lights** `{ id, x, y, z, on, color, brightness_pct }[]` (**§3.9**). **REQ-006**, **REQ-011** |
| API | `DELETE /api/v1/models/{id}` | Delete model (**204** / **404**). **409** if the model is **assigned** to **any** scene (**REQ-006** rule **5**, **§3.13**). **REQ-006** |
| API | `GET /api/v1/scenes` | List scenes (**id**, **name**, **created_at**, **model_count**). **REQ-015** |
| API | `POST /api/v1/scenes` | Create scene: **JSON** **`{ "name", "models": [ { "model_id": "<uuid>" }, … ] }`** — **≥ 1** **element**; **array** **order** **is** **placement** **order** (**REQ-015** **rule** **2**). **Server** **computes** **all** **`offset_x/y/z`** **(§3.12** **create-time** **algorithm**)**; **client** **MUST** **not** **send** **offsets** **on** **create** (**reject** **or** **ignore** **unknown** **offset** **fields** **per** **§3.13**). **201** + scene summary. **REQ-015** |
| API | `GET /api/v1/scenes/{id}` | Scene **metadata** + **placements** + **full** **lights** (positions + state) **per** **model** for **one** **round-trip** (**§3.13**). **REQ-015** |
| API | `DELETE /api/v1/scenes/{id}` | Delete scene and **all** **placements** (**204** / **404**). Used after **user** **confirms** **deleting** **the** **whole** **scene** when **removing** **the** **last** **model** (**REQ-015**). |
| API | `POST /api/v1/scenes/{id}/models` | **Add** a **model** to an **existing** scene: **`{ "model_id", "offset_x"?, "offset_y"?, "offset_z"? }`** — if **offsets** **omitted**, server **computes** **default** **“to** **the** **right”** **placement** (**§3.13**). **201** / **400** / **404** / **409** (duplicate model in scene). |
| API | `PATCH /api/v1/scenes/{id}/models/{modelId}` | **Update** **integer** **offsets** **`{ "offset_x", "offset_y", "offset_z" }`** (all required or **PATCH** **semantics** per implementor—**must** **re-validate** **containment**). **200** with **updated** **placement** + **optional** **derived** **bounds** **metadata**. |
| API | `DELETE /api/v1/scenes/{id}/models/{modelId}` | **Remove** **model** **from** **scene**. If **> 1** models: **204**. If **this** **is** **the** **last** **model**: **409** **`scene_last_model`** — **no** **mutation**; **message** **states** **that** **confirming** **will** **delete** **the** **entire** **scene**; client **shows** **modal** then calls **`DELETE /api/v1/scenes/{id}`** (**§3.13**). |
| API | `GET /api/v1/scenes/{id}/dimensions` | Return scene-space dimensions used for region queries: **`origin`**, **`size`**, **`max`**, and **`margin_m`** (see **§3.15**). **REQ-020** |
| API | `GET /api/v1/scenes/{id}/lights/events` | **SSE** (**`text/event-stream`**): JSON **`data:`** lines **`{ "seq": <uint64> }`** after scene-relevant light state commits (**REQ-029**, **§3.18**). |
| API | `GET /api/v1/scenes/{id}/lights` | Return all scene lights as a flattened list in scene coordinates (**`sx/sy/sz`**) plus model/light identity. **REQ-020** |
| API | `POST /api/v1/scenes/{id}/lights/query/cuboid` | Return only lights inside a caller-supplied cuboid in scene space. **REQ-020** |
| API | `POST /api/v1/scenes/{id}/lights/query/sphere` | Return only lights inside a caller-supplied sphere in scene space. **REQ-020** |
| API | `PATCH /api/v1/scenes/{id}/lights/state/cuboid` | Transactional bulk state update for all lights inside a cuboid in scene space; state semantics match **REQ-011**. **REQ-020** |
| API | `PATCH /api/v1/scenes/{id}/lights/state/sphere` | Transactional bulk state update for all lights inside a sphere in scene space; state semantics match **REQ-011**. **REQ-020** |
| API | `PATCH /api/v1/scenes/{id}/lights/state/scene` | Transactional bulk state update for **every** light currently in the scene (all derived **`sx/sy/sz`**); body is **only** **`on`**, **`color`**, **`brightness_pct`** (at least one required); **no** geometry object. **REQ-020** (**uniform** **whole-scene** **patch** **for** **routines** **§3.16**) |
| API | `PATCH /api/v1/scenes/{id}/lights/state/batch` | Transactional **per-light** **updates** **within** **one** **scene**: body **`{ "updates": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, … ] }`** — **each** **object** **identifies** **one** **light** **by** **`model_id` + `light_id` (idx)** **and** **must** **be** **a** **member** **of** **the** **scene** **at** **request** **time** **else** **`400`**; **at** **least** **one** **state** **field** **per** **element**; **partial** **merge** **per** **row** **like** **§3.9**; **all** **or** **nothing** **in** **one** **transaction**. **REQ-020** (**scene-level** **batch** **for** **routines** **and** **integrators** **§3.15**) |
| API | `GET /api/v1/routines` | List routine definitions (**id**, **name**, **description**, **type**, **`python_source`?**, **created_at**). **`python_source`** **present** **only** **when** **`type`** **is** **`python_scene_script`** (**REQ-022**); **omitted** **or** **`null`** **for** **other** **types**. **REQ-021**, **REQ-022** |
| API | `GET /api/v1/routines/{id}` | **200** **one** **definition** **(same** **fields** **as** **list** **row,** **including** **`python_source`** **for** **`python_scene_script`)**. **404** **if** **unknown**. **REQ-022** **(editor** **load** **without** **scanning** **the** **full** **list**)**. |
| API | `POST /api/v1/routines` | Create definition: **`{ "name", "description", "type", "python_source"? }`** — **`name`** and **`type`** **non-empty** **trimmed**; **`description`** **string** **(may** **be** **`""`)**. **When** **`type`** **is** **`python_scene_script`**, **`python_source`** **MUST** **be** **present** **(string,** **UTF-8** **text;** **empty** **allowed** **but** **run** **may** **fail** **with** **clear** **client**/**server** **error)**. **201** + full row. Unknown **`type`** → **400**. **REQ-021**, **REQ-022** **(UI** **must** **populate** **`type`** **from** **§4.12** **dropdown** **—** **REQ-023**)** |
| API | `PATCH /api/v1/routines/{id}` | **REQ-022** **only** **for** **`python_scene_script`:** **JSON** **`{ "name"?, "description"?, "python_source"? }`** **with** **≥ 1** **field** **after** **trim** **semantics** **per** **field**. **200** **updated** **row**. **404** **if** **missing**. **409** **`routine_not_editable`** **if** **`type`** **≠** **`python_scene_script`**. **409** **`routine_run_active`** **if** **a** **`running`** **run** **exists** **for** **this** **`routine_id`** **(same** **policy** **as** **delete** **—** **stop** **first)**. |
| API | `DELETE /api/v1/routines/{id}` | Delete definition. **204** / **404**. **409** **`routine_run_active`** if a **persisted** **run** **row** **exists** **with** **`status=running`** **for** **this** **`routine_id`** (**§3.16**). **REQ-021** |
| API | `POST /api/v1/scenes/{id}/routines/{routineId}/start` | Start (or **idempotently** **confirm**) an **automated** **run** **of** **`routineId`** **on** **scene** **`id`**. **201** **`{ "run_id", "scene_id", "routine_id", "status" }`** **or** **200** **if** **already** **running** **same** **pair** **(see** **§3.16**)**. **404** missing scene or routine; **409** **`scene_routine_conflict`** if **another** **routine** **already** **has** **`status=running`** **on** **this** **scene**. **REQ-021** |
| API | `POST /api/v1/scenes/{id}/routines/runs/{runId}/stop` | Stop **run** **`runId`** **if** **it** **belongs** **to** **scene** **`id`**. **200** **`{ "run_id", "status": "stopped" }`** **or** **204**. **404** if **mismatch** **or** **unknown**. **REQ-021** |
| API | `GET /api/v1/scenes/{id}/routines/runs` | **`200`**: **`{ "runs": [ { "id", "routine_id", "routine_name", "status" } ] }`** — **only** **`status=running`** **rows** **for** **this** **scene** (**empty** **array** **or** **one** **element** **per** **§3.16** **concurrency** **rule**). **REQ-021** |
| API | `GET /api/v1/models/{id}/lights/state` | **All** lights’ **state** for the model (**ordered** by **`id`**). **REQ-011** |
| API | `GET /api/v1/models/{id}/lights/events` | **SSE** (**`text/event-stream`**): **`data:`** lines **`{ "seq": <uint64> }`** after successful light-state writes for this model (**REQ-029**, **§3.18**). |
| API | `GET /api/v1/models/{id}/lights/{lightId}/state` | **One** light’s **state** (**404** if model or **lightId** missing). **REQ-011** |
| API | `PATCH /api/v1/models/{id}/lights/{lightId}/state` | **Partial** update of **`on`**, **`color`**, **`brightness_pct`** (JSON body; omitted fields unchanged). **200** returns the **full** **updated** **state** object. **REQ-011** |
| API | `PATCH /api/v1/models/{id}/lights/state/batch` | **Atomic** partial update of **many** lights: JSON body **`{ "ids": [<int>, …], "on"?, "color"?, "brightness_pct"? }`** with **at least one** of **`on`**, **`color`**, **`brightness_pct`** present; **omitted** fields **unchanged** per row. **200** returns **`{ "states": [ { "id", "on", "color", "brightness_pct" }, … ] }`** sorted by **`id`**. **400** if **`ids`** empty, duplicate ids, any id **∉ [0, n−1]**, or merged values invalid. **REQ-013** (**§3.10**). |
| API | `POST /api/v1/models/{id}/lights/state/reset` | **REQ-014:** **no** body (or empty **JSON** object). **One** **transaction** sets **every** light in the model to **`on=false`**, **`color="#ffffff"`**, **`brightness_pct=100`**. **200** returns **`{ "states": [ … ] }`** for **all** **n** lights, **`id`** ascending. **404** if model missing. |
| API | `POST /api/v1/system/factory-reset` | **REQ-017:** **no** body (or **`{}`**). **One** **transaction** **deletes** **all** **`scene_models`**, **`scenes`**, **`lights`**, **`models`** **rows** **in** **FK-safe** **order**, **then** **inserts** **the** **three** **default** **samples** **via** **the** **same** **logic** **as** **`SeedDefaultSamples`** (**§3.8**, **§3.14**). **200** **`{ "ok": true }`**. **500** **only** **on** **unexpected** **store** **failure** **(no** **partial** **wipe** **visible** **after** **commit**)**. **Register** **before** **any** **`{id}`** **catch-alls** **so** **`system`** **is** **not** **parsed** **as** **an** **id**. |
| API | `/api/v1/*` | Other versioned **JSON** endpoints as the product grows. |
| Static | `/`, `/*.html`, **`/_next/**`**, other export assets | **Next static export** tree from embed; **SPA / HTML5** fallback policy: serve **`index.html`** for unmatched **non-API** GET if needed (implementor defines exact fallback rules). |

**Ordering:** Register **API** routes **before** static/`NotFound` handling so `/api/v1` is never swallowed by UI fallback. Register **literal** **path** **segments** **`routines`**, **`start`**, **`stop`**, **`factory-reset`**, **`reset`**, **`state`**, **`events`**, **`batch`**, **`system`** **before** **any** **`{id}`** **pattern** **that** **could** **swallow** **them** (**`lights/events`** **before** **`lights/{lightId}`**). Register **`PATCH /api/v1/routines/{id}`** **so** **`routines`** **is** **not** **matched** **as** **an** **id** **elsewhere**.

### 3.3 Persistence (**REQ-006**)

- **Engine:** **SQLite** accessed from the **same Go process** (prefer a **pure Go** driver such as **`modernc.org/sqlite`** to avoid **cgo** cross-compile friction on **linux/arm64**).
- **Location:** Configurable path via environment (e.g. **`DLM_DB_PATH`**), defaulting to a file under an application **data directory** (e.g. **`DLM_DATA_DIR`** + `dlm.db`). The DB file is **runtime state**, not embedded in the binary; creating/opening it at startup satisfies **REQ-004** (no extra shipped daemon).
- **Schema (logical):**
  - **`models`:** `id` (TEXT UUID primary key), `name` (TEXT **NOT NULL**, **UNIQUE**), `created_at` (TEXT **RFC3339 UTC**).
  - **`lights`:** `model_id` (TEXT FK → `models.id`), `idx` (INTEGER light index), `x`, `y`, `z` (REAL), **`on`** (INTEGER **NOT NULL**, **0** = off / **1** = on), **`color`** (TEXT **NOT NULL**, canonical **`#RRGGBB`**), **`brightness_pct`** (REAL **NOT NULL**, **0..100** inclusive), primary key **`(model_id, idx)`**.
- **Transactions:** **`POST /api/v1/models`** MUST parse+validate CSV, then insert **`models`** + all **`lights`** in **one transaction**; each new **`lights`** row MUST receive **default** **`on`**, **`color`**, **`brightness_pct`** per **§3.9** (**REQ-014**); rollback on any validation failure so **no partial model** is stored (**REQ-007**).
- **Migrations:** Implementor MAY use a minimal migration hook or `CREATE TABLE IF NOT EXISTS` at startup; document schema version if evolved.

**Architectural resolutions (requirements open questions):**

- **Model name uniqueness:** **UNIQUE** on `models.name` so the list UI stays unambiguous; duplicate names return **409** with a clear message (product may relax later if requirements change).
- **Creation date:** Set by the **server** at successful create time, stored and exposed as **UTC RFC3339** (no client clock trust).
- **Upload naming:** **`name`** is a **required** multipart field (non-empty after trim); **not** inferred from filename alone.
- **Empty model:** **Allowed**: CSV may contain **only** the header row (**0** lights); still creates a model row with **light_count = 0**.

### 3.4 Build and cross-compile (release artifact)

- **Release binary (Pi):** After the **web static bundle** is copied into **`internal/webdist/`** (or embed path the implementor chooses):

  `GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server`

- **Single file:** The **only** shipped application binary is **`dlm-arm64`** (name per project convention). **systemd** may wrap it with **`ExecStart=/usr/local/bin/dlm`** etc.; that is allowed per REQ-004.

### 3.5 Embedding static UI

- **Mechanism:** `//go:embed` on a package (e.g. **`internal/webdist`**), embedding the **full** `out/` tree from **`next export`** (path names preserved: **`_next/static/...`**).
- **Build contract:** A **documented** step (**Make**, `task`, or **CI** script) runs **`npm ci && npm run build`** in `web/` (with **`output: 'export'`**), then **syncs** `web/out/` → `backend/internal/webdist/` (or `backend/web/build/`).
- **`embed` empty-dir issue:** Repository may keep a **small placeholder** file under `webdist/` so **`go test ./...`** works **before** the first UI bake; release builds **replace** the directory contents with the real export.

### 3.6 CSV interchange and validation (**REQ-005**, **REQ-007**)

- **Encoding:** **UTF-8** text; parse with Go **`encoding/csv`**.
- **Delimiter:** Comma (`,`). **RFC 4180** quoting rules SHOULD be supported so numeric fields may be quoted.
- **Header row:** **Required** first record; must be **exactly** the four fields **`id`**, **`x`**, **`y`**, **`z`** in that order (case-sensitive, no extra columns). Wrong or missing header → **400** with **format** error (matches acceptance scenario for `idx,x,y,z`).
- **Data rows:** Each row provides one light; **`id`** must parse as an integer; **`x`**, **`y`**, **`z`** must parse as **finite** floating-point numbers (reject **NaN** / **Inf**).
- **Row count:** After the header, **0 ≤ n ≤ 1000** data rows; **n > 1000** → **400** referencing the cap.
- **Id sequence:** Let **n** be the number of data rows. Sorting rows by **`id`**, the multiset of ids MUST equal **`{0,1,…,n−1}`** exactly once each (contiguous from **0**, no gaps, no duplicates). Implementor MAY require rows to appear sorted by **`id`** ascending for simpler validation, or accept any order and validate the set—behavior MUST match tests derived from **`docs/acceptance_criteria.md`**.
- **Authoritative checks:** All rules above are enforced **in Go** on upload; the UI MAY add client-side hints but MUST NOT trust them for security or integrity.

**Chain topology (REQ-005 rule 6):** The **CSV** does **not** encode edges explicitly; **logical** adjacency is **only** **(i, i+1)** for **i = 0 … n−2**. **Visualization** and **any** future features MUST **not** introduce **extra** adjacencies between **non-consecutive** **ids**.

**JSON API shapes (illustrative):**

- **List item:** `{ "id": "<uuid>", "name": "…", "created_at": "<RFC3339>", "light_count": <int> }`
- **Detail:** same metadata plus `"lights": [ { "id": 0, "x": 0, "y": 0, "z": 0, "on": true, "color": "#RRGGBB", "brightness_pct": 100 }, … ]` ordered by **`id`** ascending (**§3.9**).

**HTTP errors:** Use existing **`{ "error": { "code", "message" } }`** envelope; optional **`details`** field for row/column hints (**REQ-007**). **409 Conflict** for duplicate **`name`**.

### 3.7 Local build-and-run script (**REQ-008**)

- **Canonical path:** **`scripts/run.sh`** at the **repository root**, executable (`chmod +x`), **`#!/usr/bin/env bash`** with **`set -euo pipefail`** (or equivalent strictness).
- **Behavior (single invocation):**
  1. Resolve repo root (e.g. `ROOT="$(cd "$(dirname "$0")/.." && pwd)"`).
  2. Under **`cwd = $ROOT/web`**: run **`npm ci`** when **`node_modules`** is missing or **`DLM_FORCE_NPM_CI=1`**; skip **`npm ci`** when **`DLM_SKIP_NPM_CI=1`** or when **`node_modules`** already exists (see **README** for trade-offs). Then **`npm run release:sync`** so **`web/out/`** is copied to **`backend/internal/webdist/dist/`**.
  3. Run **`go run ./cmd/server`** with **`cwd = $ROOT/backend`** (or **`go build -o … && exec ./dlm`** if preferred—document in **README**). Use **`exec`** for the final Go process so **SIGINT** reaches the server.
- **Prerequisites:** **Node.js** + **npm** on **PATH**, **Go** on **PATH**; network for **`npm ci`** on cold clone or when forcing a clean install.
- **README.md** MUST show the exact line, e.g. **`./scripts/run.sh`** or **`bash scripts/run.sh`**, and note **WSL / Git Bash** on Windows if applicable.
- **AGENTS.md** already references **REQ-008**; keep **README** and this section aligned when the script changes.
- **Not** a substitute for **Pi** production install (still **one binary** + optional **systemd**); operators may build off-device and copy **`dlm-arm64`** per **§3.4**.

### 3.8 Default sample models (**REQ-009**)

**Units:** **SI meters**. Each sample returns **`[]wiremodel.Light`** with sequential **`id` 0 … n−1** (this order is the **polyline** order for **REQ-010** segment drawing).

**Counts:** Each of the three samples MUST have **500 ≤ n ≤ 1000** vertices (**REQ-009** / **REQ-005**). Choose a **deterministic** **n** per shape (fixed constant or derived from a reproducible rule) so tests and seeds are stable.

**Consecutive spacing:** For every **i ∈ {0,…,n−2}**, let **dᵢ** be the Euclidean distance between light **i** and **i+1**. Require **0.05 ≤ dᵢ ≤ 0.10** (5–10 cm). Tests SHOULD use a small tolerance (e.g. **`1e-3`** m) on the bounds to absorb floating-point error; document the chosen **ε**.

**Surface placement (nominal solids, centered/aligned as today):**

- **Nominal boundary:** Analytic surface of the intended **sphere** (**radius R = 1.0** m, diameter **2 m**, center **origin**), **axis-aligned cube** (**edge 2** m, center **origin**, faces at **±1** m), or **right circular cone** (**height 2** m, base radius **1** m, base in **z = 0** plane, apex at **(0,0,2)** — or equivalent consistent pose documented in code).
- **Exterior only:** Each vertex MUST lie **on** the nominal boundary **or** in the **thin outer shell** allowed by requirements: **closest** distance from the point to the **nominal surface** MUST be **≤ 0.03** m, and the point MUST **not** lie in the **interior** of the solid (e.g. sphere: **‖p‖ ≥ R − δ** with **δ** tiny for float; cube: not strictly inside the **2×2×2** volume; cone: on or outside the **lateral surface + base disk** region as defined by implementor’s half-space test).

**Coverage intent (REQ-009 “not edge-only”):**

- **Cube:** Lights MUST be placed on **all six** exterior **square face planes**. A **wireframe** layout (vertices and **edges only**) is **not** sufficient. The **majority** of lights MUST lie in the **interior** of each face’s square (parameterized patch), not only on **edges** or **corners**—architecture sets a **quota**: at least **⌊0.85 n⌋** lights MUST lie **strictly** in the **open** face patches (each coordinate on a face is **inside** the **(−1,1)²** parameter rectangle for that face, i.e. not on the **rim** of the square in face-local **(u,v)**). The remaining lights MAY be used for **transitions** between faces or to satisfy **dᵢ** if needed. **Evenness (cube):** Partition **n** across the six faces with counts differing by **at most one** (e.g. base **q = ⌊n/6⌋**, **r = n mod 6** faces get **q+1** lights). On each face, place that many points on a **deterministic quasi-uniform 2D pattern** in **(u,v)** (e.g. **regular staggered grid**, **2D Halton** pairs, or **boustrophedon** rows) so **area** coverage is **even** within the face; document the chosen pattern in code comments.
- **Sphere:** Lights MUST lie on **‖p‖ = R** and MUST **not** read as **concentrated on a single narrow strip** or **one 1D curve** only. **Evenness (sphere):** Use a **point set** known for **approximate equal area** per point on the sphere (e.g. **Fibonacci / golden-angle** lattice, **HEALPix-style** construction, or **subdivided icosahedron** vertices). After placement, **order** into the final polyline (below) so **REQ-010** still shows a connected path; the **underlying** set MUST pass tests that **no** hemisphere (or other fixed **cap** of bounded area) contains more than a **documented fraction** of lights (e.g. **≤ 55%** of **n** in any closed hemisphere through the center)—implementor picks caps and thresholds to match acceptance tests.
- **Cone:** Lights MUST cover **both** the **lateral** (curved) surface **and** the **flat base disk** (**z = 0**, **ρ ≤ r**). **Evenness (cone):** Split **n** between **lateral** and **base** in proportion to **surface areas** (lateral **π r ℓ**, base **π r²** with slant height **ℓ**), rounding with a **deterministic rule** so the two counts sum to **n**. On the lateral patch, use a **(height, azimuth)** quasi-uniform grid or spiral; on the base, use a **polar** or **2D** quasi-uniform pattern in the **disk**. Same **≥ 85%** rule as the cube applies **within each part**: at least **⌊0.85 n_lat⌋** lateral lights MUST lie in the **interior** of the lateral patch (not only on the **rim** or **apex**), and at least **⌊0.85 n_base⌋** base lights in the **interior** of the **disk** (**ρ** strictly between **0** and **r**), unless **n** for that part is too small—in which case document the **degenerate** exception in tests.

**Ordering vs spacing (two-step design):**

Requirements demand **both** **even** **2D/area** placement **and** **consecutive** **dᵢ** in **[0.05, 0.10]**. Architecturally, treat this as:

1. **Target positions:** Produce an **unordered** (or weakly ordered) multiset of **on-surface** points meeting **evenness** and **per-face/per-part quotas** above.
2. **Path construction:** Build a **single open polyline** **P₀,…,P_{n−1}** through **exactly** those points (or through **refined** points after **subdivision**) such that every **dᵢ** lies in the band. **Allowed techniques:** (a) generate points **along** a **space-filling** or **serpentine** path on the unfolded or parameterized surface so consecutive samples are naturally **~0.075 m** apart; (b) **sort** / **chain** with **greedy nearest-neighbor** from a **fixed seed** and then **split long edges** / **merge short jumps** by inserting or removing intermediates **on the same surface**; (c) **walk** face patches in a **fixed order** (cube) with **boustrophedon** rows at **~0.075 m** step, using **short** **surface** segments across **edges** only where needed. **Subdivide** any segment longer than **0.10 m** with collinear **on-surface** inserts; **avoid** interior shortcuts that leave the boundary.

**Spacing algorithm (implementor):** After the polyline exists, **verify** all **dᵢ**; adjust with **subdivision** or **local reordering** while **preserving** **evenness** tests. If **n** falls outside **[500,1000]**, **retune** step size or **pattern density** and regenerate—do **not** relax **dᵢ** or **surface** rules.

**Architectural resolutions (requirements open questions):**

- **Open vs closed path:** **Open polyline** is the default; **REQ-010** draws segments **(i, i+1)** only, so **no** automatic **(n−1,0)** unless requirements change.
- **User deletes samples:** **Seed only when `SELECT COUNT(*) FROM models` is 0** immediately after migrations on **process startup**. Deleting a subset does **not** re-seed during that run; if the user deletes **all** models, **next** process start seeds again.
- **Fixed English names:** **`Sample sphere`**, **`Sample cube`**, **`Sample cone`**.

**Per-shape summary (`internal/samples`):**

| Shape | Characteristic size | Coverage + ordering |
|-------|---------------------|---------------------|
| **Sphere** | **R = 1** m | **Area-even** point set on **‖p‖ = R**; **ordered** into a polyline with **dᵢ ∈ [0.05,0.10]**; **n ∈ [500,1000]**; **hemisphere / cap** tests to block **single-strip** dominance. |
| **Cube** | **Edge 2** m | **Six faces**, **~n/6** lights per face (±1); **interior** **(u,v)** patterns; **open polyline** visiting faces with **dᵢ** in band—**not** an **edge-only** walk. |
| **Cone** | **h = 2** m, **r = 1** m | **Area-split** between **lateral** and **base**; **quasi-uniform** on each; **open polyline** with **dᵢ** in band covering **both** parts. |

**Integration:**

- After **`store.Open`** + **`migrate`**, if **model count is 0**, **`store.SeedDefaultSamples(ctx)`** runs **one transaction** inserting three models and lights via **`samples.SphereLights()`**, **`samples.CubeLights()`**, **`samples.ConeLights()`**.
- Samples use the **same** persistence path as user CSV so **REQ-005** invariants hold; inserted **`lights`** rows MUST receive **default** **`on`**, **`color`**, **`brightness_pct`** per **§3.9**.

**Tests:** **`internal/samples`**: assert **500 ≤ n ≤ 1000**; **0.05 − ε ≤ dᵢ ≤ 0.10 + ε** for all consecutive pairs; **surface** predicates (distance-to-nominal-surface **≤ 0.03**, not interior) per shape; characteristic **~2 m** extent within documented tolerance; **cube**: each of the **six** faces has the expected **light count** (±**1** of **⌊n/6⌋** / **⌈n/6⌉**) and **≥ ⌊0.85 n⌋** lights in **open** face interiors; **sphere**: **cap / hemisphere** bound per **§3.8**; **cone**: **both** lateral and base receive lights per **area** split, with **interior** quotas analogous to the cube where **n** per part allows. Integration: fresh DB list shows **three** expected names.

### 3.9 Per-light state API and persistence (**REQ-011**, **REQ-012**, **REQ-031**)

**Canonical JSON field names (API):**

| Field | Type | Rules |
|-------|------|--------|
| **`on`** | boolean | **`true`** = lit appearance (**REQ-012** **opaque** **filled**); **`false`** = **off** (**REQ-012** **`#D0D0D0`** at **85%** transparency in **three.js**, **§4.7**). |
| **`color`** | string | **Exactly** **`#` + 6** **hex** digits **`[0-9A-Fa-f]`** (normalize to **lowercase** in responses if desired; accept any case on input). |
| **`brightness_pct`** | number | **IEEE-754** number **in** **[0, 100]**; semantics for **rendering** in **§4.7**; when **`on`** is **`false`**, stored value is **still** persisted but **does not** drive a **lit** appearance. |

**Default state (REQ-011 rule 6 + REQ-014):** On **insert** of each **`lights`** row (**CSV** create, **sample** seed, or **migration** backfill), set **`on = false`**, **`color = "#ffffff"`** (canonical lowercase in DB/API responses acceptable), **`brightness_pct = 100`**. This matches **REQ-014** (**all off**, **full white** stored, **100%** brightness for when the light is **turned on**).

**Legacy databases** that used **`on = true`** on insert SHOULD be **migrated** (or **reset** via **UI/API**) so **product** **behavior** matches **REQ-014**; **implementor** documents the **chosen** **path**.

**Endpoints (summary — see §3.2 table):**

- **`GET /api/v1/models/{id}/lights/state`** → **`{ "states": [ { "id": <int>, "on": <bool>, "color": "<hex>", "brightness_pct": <num> }, … ] }`** sorted by **`id`** ascending. **`id`** is the light index (**REQ-005**).
- **`GET /api/v1/models/{id}/lights/{lightId}/state`** → **`{ "id", "on", "color", "brightness_pct" }`** (**404** if unknown model or **`lightId`** not in **`[0, n−1]`**).
- **`PATCH /api/v1/models/{id}/lights/{lightId}/state`** with **`Content-Type: application/json`** body containing **any subset** of **`on`**, **`color`**, **`brightness_pct`**. Unspecified fields **remain unchanged**. **Validate** merged result: **`color`** format, **`brightness_pct`** in **[0, 100]**; reject with **400** and **`error.message`** on violation. **200** response body is the **full** **updated** state object (same shape as the single-light **GET**).

**Implementation notes:**

- **`internal/store`:** Implement **read**/**patch** against **`lights`** columns; **`PATCH`** SHOULD run in a **short** transaction (**BEGIN IMMEDIATE** optional) so concurrent updates stay consistent. **REQ-031:** **Compare** **merged** **vs** **stored** **effective** **triple** **before** **`UPDATE`** **(**§3.19**)** **;** **optional** **`unchanged`** **flag** **on** **`200`** **responses**.
- **`GET /api/v1/models/{id}`** MUST embed **`on`**, **`color`**, **`brightness_pct`** on **each** element of **`lights`** so the **detail** page needs **one** round-trip for positions + state (**REQ-006** + **REQ-011**).
- **Bulk** **`PUT`** of **all** states in **one** request is **not** required by requirements; **optional** future endpoint if product needs **show-mode** uploads.

**Upload limits:** **`PATCH`** bodies are **tiny**; existing **`MaxBytesReader`** on **`POST /api/v1/models`** is unchanged; add a **small** limit (e.g. **8 KiB**) on **single-light** **`PATCH`** **JSON** if a **global** body cap is not already applied. **`PATCH …/state/batch`** MAY allow a **larger** cap (e.g. **64 KiB**) so **≤ 1000** integer **ids** fit comfortably (**REQ-013**).

### 3.10 Batch light state update (**REQ-013**, extends **REQ-011**)

**Rationale:** Applying the **same** **`on`**, **`color`**, and **`brightness_pct`** to **many** selected lights via **N** sequential **`PATCH …/lights/{lightId}/state`** calls risks **slow** UX, **partial** failure mid-batch, and **many** round trips on **Pi-class** networks. **REQ-013** is satisfied with **one** **authoritative** **transaction** on the server.

**Endpoint:** **`PATCH /api/v1/models/{id}/lights/state/batch`**

**Request body (JSON):**

| Field | Type | Rules |
|-------|------|--------|
| **`ids`** | array of int | **Required**; **non-empty**; **no** duplicates; each **id** MUST satisfy **0 ≤ id ≤ n−1** for the model’s light count **n** (**REQ-005**). |
| **`on`** | boolean | **Optional**; if present, applied to **every** listed light. |
| **`color`** | string | **Optional**; if present, canonical **`#RRGGBB`** for **all** listed lights (**same** validation as **§3.9**). |
| **`brightness_pct`** | number | **Optional**; if present, **in** **[0, 100]** for **all** listed lights. |

**Rule:** At **least one** of **`on`**, **`color`**, **`brightness_pct`** MUST appear in the body (otherwise **400** — empty patch object is rejected to catch client bugs).

**REQ-031** **note:** **When** **the** **body** **is** **valid** **but** **every** **listed** **light** **already** **matches** **the** **merged** **state** **(**§3.19** **equivalence**)** **,** **the** **handler** **SHOULD** **return** **`200`** **with** **`states[]`** **and** **perform** **no** **`UPDATE`s** **(**optional** **response** **field** **`unchanged_all`: true** **)** **—** **see** **§3.19**.

**Behavior:**

1. **Validate** model exists (**404** if not).
2. **Validate** **`ids`** (non-empty, unique, in range); on failure **400** with **`error.message`** naming **out-of-range** or **duplicate** ids where practical.
3. **For each** id, **merge** the provided fields onto the **current** row (**same** semantics as **single** **`PATCH`**).
4. **Reject** with **400** if **merged** **`color`** or **`brightness_pct`** would be invalid (same rules as **§3.9**).
5. Run **all** updates in **one** **SQLite** **transaction** (**BEGIN** … **COMMIT**); on success **200** with **`{ "states": [ … ] }`** containing the **full** **post-update** state for **each** id in **ascending** **id** order.

**Store:** Implement in **`internal/store`** as e.g. **`BatchPatchLightStates(ctx, modelID, ids, patch)`**; **`internal/httpapi`** maps **JSON** ↔ **store** and **error** envelope.

**Client:** After **200**, **merge** **`states`** into the **in-memory** **`lights`** array used by **§4.7** and the **paginated** table (**§4.8**) in **one** **React** **setState** (or equivalent) so **REQ-012** **timeliness** holds for **bulk** applies (**no** **indefinite** **staleness**).

**Compatibility:** Single-light **`PATCH …/lights/{lightId}/state`** remains **available** for integrators and simple UI paths (**REQ-011**).

### 3.11 Reset all lights to default state (**REQ-014**)

**Endpoint:** **`POST /api/v1/models/{id}/lights/state/reset`**

**Request:** **No** JSON body required (**`Content-Length: 0`** or **`{}`** accepted and ignored).

**Behavior:**

1. **404** if the model does **not** exist.
2. **BEGIN** transaction; **UPDATE** **every** **`lights`** row for **`model_id`** to **`on = 0`**, **`color = '#ffffff'`** (or stored canonical case), **`brightness_pct = 100`**.
3. **COMMIT**; **200** with **`{ "states": [ { "id", "on", "color", "brightness_pct" }, … ] }`** covering **all** **n** lights, sorted by **`id`**.

**Store:** e.g. **`ResetAllLightStates(ctx, modelID)`** in **`internal/store`**; register **`POST …/lights/state/reset`** on the **API** mux with the **other** **`…/lights/…`** **routes** so the **literal** **`reset`** **segment** is **not** mistaken for a **`lightId`** in any **future** **catch-all** patterns (**ordering** per **§3.2**).

**Client:** **Model detail** exposes a **Reset** (or **Reset lights**) **button** (**§4.6**). On **200**, **merge** **`states`** into **`lights`** and refresh **§4.7** + **§4.8** in **one** update (**same** **timeliness** expectation as **§8.7**).

**Integrators:** **Idempotent** aside from **touching** **`updated`** semantics if added later; **safe** to call repeatedly.

### 3.12 Scene persistence and containment (**REQ-015**)

**Tables (logical, SQLite):**

| Table | Columns | Notes |
|-------|---------|--------|
| **`scenes`** | `id` (TEXT **UUID** PK), `name` (TEXT **NOT NULL**, **UNIQUE**), `created_at` (TEXT **RFC3339** UTC) | One row per scene. **No** row without **≥ 1** **`scene_models`** row **after** successful **create** (**transaction**). |
| **`scene_models`** | `scene_id` (TEXT **FK** → **`scenes.id`** **ON DELETE CASCADE**), `model_id` (TEXT **FK** → **`models.id`** **ON DELETE RESTRICT** or **check** before **model** **delete**), `offset_x`, `offset_y`, `offset_z` (INTEGER **NOT NULL**) | **PK** **`(scene_id, model_id)`** — each **model** **at most once** per **scene**. **Offsets** are **signed** **integers** **interpreted** **as** **whole** **SI** **meters** added to **each** **light’s** **`x`**, **`y`**, **`z`** (**float64**) from **`lights`**. **API** **accepts** **only** **non-negative** **offsets**; **invalid** **combinations** **fail** **containment** **checks** below. |

**Referential integrity:** **Deleting** a **model** **MUST** **fail** with **409** if **any** **`scene_models`** row references **`model_id`** (**REQ-006** **rule** **5**). Use **`ON DELETE RESTRICT`** on **`scene_models.model_id`** and/or an **explicit** **pre-check** in the **handler** so the **JSON** **error** **matches** **`model_in_scenes`** (**§3.13**).

**Scene-space position:** For **light** **`L`** **with** **model-space** **`(x, y, z)`** and **placement** **`(ox, oy, oz)`**:

**`sx = x + ox`**, **`sy = y + oy`**, **`sz = z + oz`** (all **float64** **arithmetic**).

**Canonical vs derived (REQ-015 rule 4):** **`lights`** **rows** **and** **`GET /api/v1/models/{id}`** **always** **expose** **the** **stored** **`x`**, **`y`**, **`z`** **(REQ-005)** — **never** **rewritten** **when** **a** **model** **joins** **or** **leaves** **a** **scene** **or** **when** **`scene_models`** **offsets** **change**. **Only** **`scene_models`** **stores** **`offset_*`**. **`GET /api/v1/scenes/{id}`** **and** **the** **scene** **three.js** **view** **use** **derived** **`sx`**, **`sy`**, **`sz`** **(plus** **canonical** **`x`**, **`y`**, **`z`** **on** **each** **light** **object** **for** **labels** **/** **comparison**).**

**Containment (authoritative in Go):** For **every** **light** **in** **every** **model** **in** **the** **scene**, **`sx`, `sy`, `sz`** **MUST** be **≥ 0** (treat **tiny** **negative** **float** **noise** **< 1e‑9** as **0** if needed—**document** **ε**). **Violations** → **400** on **create**/**add**/**patch** with **`error.message`** **naming** **model** **and** **axis** **where** **possible**.

**Automatic bounds / margin for rendering:** Let **`Mmax_x`**, **`Mmax_y`**, **`Mmax_z`** be the **maxima** **of** **`sx`**, **`sy`**, **`sz`** **respectively** **over** **all** **lights** **in** **the** **scene**. The **visual**/**framing** **AABB** **uses** **origin** **`(0,0,0)`** and **upper** **corner** **`(Mmax_x + 1, Mmax_y + 1, Mmax_z + 1)`** **meters** (**≥ 1 m** **padding** **beyond** **tight** **max** **per** **axis**; **mins** **at** **0** **per** **REQ-015**). **Camera** **framing** (**§4.9**) **uses** **this** **box** **(+** **sphere** **radius** **for** **2** **cm** **markers**).

**Canonical “right” axis (+X):** **three.js** **default** **camera** **looks** **toward** **−Z** **with** **+Y** **up**; **“to** **the** **right”** **of** **the** **existing** **layout** **means** **increasing** **scene** **`+X`**. **Document** **in** **UI** **copy** **if** **needed** **for** **user** **mental** **model**.

**Default placement for `POST …/scenes/{id}/models` (omit offsets):**

1. **Compute** **`M_x`** = **maximum** **`sx`** over **all** **lights** **already** **in** **the** **scene** (if **no** **lights**, **`M_x = 0`**).
2. **Gap:** **`gap_m = 0`** **meters** **abutting** **(architecture** **default**; **implementor** **MAY** **use** **a** **small** **positive** **constant** **e.g.** **`0.1`** **if** **tests** **prefer** **visual** **separation**—**document**).
3. For the **incoming** **model**, **load** **all** **lights**; **let** **`m_min_x`** = **min(x)**, **`m_min_y`** = **min(y)**, **`m_min_z`** = **min(z)** **over** **its** **lights**.
4. **`offset_x`** = **smallest** **integer** **≥** **`M_x + gap_m - m_min_x`** **(ceil** **to** **int**). **`offset_y`** = **smallest** **integer** **≥** **`−m_min_y`** **so** **that** **`y + offset_y ≥ 0`** **for** **all** **lights** **(typically** **`max(0, ⌈−m_min_y⌉)`** **if** **all** **`y`** **share** **offset**). **`offset_z`** **analogous** **for** **`z`**. **Re-run** **full** **containment** **check** **after** **rounding**.

**Create-time placement for `POST /api/v1/scenes` (REQ-015 rule 2):** **Input** **is** **an** **ordered** **list** **`[model_id₁, …, model_idₙ]`** **(JSON** **array** **order**)**.** **Store** **implementation** **`CreateScene(ctx, name, modelIDs []string)`** **(or** **equivalent**)**:**

1. **Validate** **unique** **ids**, **existence**, **`len ≥ 1`**, **scene** **name**.
2. **`placed := []scenePlacementLights{}`** **(empty** **accumulator**)**.**
3. **For** **each** **`mid`** **in** **`modelIDs`** **order**:
   - **Load** **`Detail(mid)`** **lights**.
   - **If** **`len(placed) == 0`**: **first** **model** — **compute** **minimal** **non-negative** **integers** **`ox`, `oy`, `oz`** **so** **every** **light** **has** **`x+ox ≥ 0`**, **`y+oy ≥ 0`**, **`z+oz ≥ 0`** **(e.g.** **`ox = max(0, ⌈−min_x(lights) − ε⌉)`**, **same** **for** **`y`**, **`z`**)**. **This** **anchors** **the** **first** **footprint** **in** **the** **non-negative** **octant** **without** **user** **input**.
   - **Else**: **`ox`, `oy`, `oz` ← `DefaultOffsetsForNewModel(placed, lights)`** **(same** **routine** **as** **`POST …/scenes/{id}/models`** **with** **omitted** **offsets**)**.**
   - **Append** **`{model_id: mid, offsetX/Y/Z: ox/oy/oz, lights}`** **to** **`placed`**; **run** **`validateScenePlacements(placed)`**; **on** **failure** **abort** **whole** **transaction** **with** **400**.
4. **BEGIN**; **insert** **`scenes`** **+** **all** **`scene_models`** **rows**; **COMMIT**.

**Optional** **REQ-015** **advanced** **override** **(not** **MVP**)**:** **If** **product** **later** **allows** **client-supplied** **offsets** **on** **create**, **document** **a** **separate** **flag** **or** **endpoint**; **default** **path** **remains** **fully** **automatic**.

**Scene name uniqueness:** **Reuse** **same** **policy** **as** **models**: **`UNIQUE`** **on** **`scenes.name`**; **duplicate** → **409**.

### 3.13 Scenes HTTP API behavior (**REQ-015**)

**`POST /api/v1/scenes`** — **Body** **`{ "name": string, "models": [ { "model_id": "<uuid>" }, … ] }`**. **Each** **object** **SHOULD** **contain** **only** **`model_id`**; **if** **`offset_x`**, **`offset_y`**, **`offset_z`** **appear**, **return** **`400`** **`validation_failed`** **(do** **not** **accept** **client** **offsets** **on** **create** **—** **REQ-015** **rule** **2**)** **or** **strict** **JSON** **schema** **rejection**. **Validate:** **`name`** **non-empty** **trimmed**; **`models.length ≥ 1`**; **no** **duplicate** **`model_id`**; **every** **`model_id`** **exists**. **Server** **runs** **§3.12** **create-time** **placement** **algorithm** **(no** **offsets** **from** **client**)**; **then** **transaction:** **insert** **`scenes`** + **`scene_models`**. **Response** **`201`:** **`{ "id", "name", "created_at", "model_count" }`**. **Canonical** **`lights`** **rows** **are** **never** **updated** **by** **this** **handler**.

**`GET /api/v1/scenes/{id}`** — **`200`:** **`{ "id", "name", "created_at", "items": [ { "model_id", "name", "offset_x", "offset_y", "offset_z", "lights": [ { "id", "x", "y", "z", "sx", "sy", "sz", "on", "color", "brightness_pct" }, … ] } ] }`**. **`x`, `y`, `z`** **MUST** **match** **persisted** **`lights`** **(canonical** **REQ-005**)**; **`sx`, `sy`, `sz`** **MUST** **equal** **canonical** **+** **offsets** **(derived)**. **Scene** **tooltip**/**labels** **SHOULD** **show** **both** **scene** **and** **model-local** **coordinates** **where** **space** **allows** (**§4.9**). **`PATCH /api/v1/scenes/{id}/models/{modelId}`** **updates** **only** **`scene_models`** **offsets** **—** **never** **`lights`**.**

**`POST /api/v1/scenes/{id}/models`** — **Add** **one** **model**. **Optional** **offsets** → **default** **§3.12**. **409** if **`model_id`** **already** **in** **scene**.

**`PATCH /api/v1/scenes/{id}/models/{modelId}`** — **Update** **offsets**; **re-validate** **full** **scene** **containment**.

**`DELETE /api/v1/scenes/{id}/models/{modelId}`** — If **remaining** **model** **count** **after** **delete** **≥ 1**: **delete** **row**, **204**. If **this** **is** **the** **only** **model**: **409** **`{ "error": { "code": "scene_last_model", "message": "…", "details": { "scene_id": "…" } } }`** — **no** **row** **deleted**. **Client** **shows** **confirm** **dialog** **then** **`DELETE /api/v1/scenes/{id}`**.

**`DELETE /api/v1/models/{id}`** — **Before** **delete**, **`SELECT`** **from** **`scene_models`**. If **any** **row**: **`409`** **`{ "error": { "code": "model_in_scenes", "message": "…", "details": { "scenes": [ { "id", "name" }, … ] } } }`** (**REQ-006** **rule** **5**).

**Light state in scene view:** **Same** **persistence** **as** **model** **detail**; **composite** **view** **uses** **`GET /api/v1/scenes/{id}`** **payload** **or** **parallel** **`PATCH`** **to** **`/models/{id}/lights/…`** **after** **pick** **model**+**light** **(implementor** **chooses** **simplest** **consistent** **with** **REQ-012** **timeliness**)**.** **Recommended:** **scene** **detail** **response** **includes** **current** **state**; **per-light** **PATCH** **from** **scene** **page** **reuses** **existing** **endpoints** **by** **`model_id`** **+** **`lightId`** **then** **merge** **into** **local** **state** **for** **that** **model’s** **group** **in** **three.js**.

### 3.14 Factory reset (**REQ-017**)

**Purpose:** **Atomically** return the **SQLite** **database** to the **same** **logical** **state** **as** **first** **startup** **with** **an** **empty** **`models`** **table** **after** **migrations** (**REQ-009** **three** **samples** **only**; **REQ-014** **defaults** **on** **inserted** **lights**).

**Store API (conceptual):** **`FactoryReset(ctx) error`** **in** **`internal/store`:**

1. **`BEGIN IMMEDIATE`**.
2. **`DELETE FROM routine_runs`** **(or** **equivalent** **child** **table** **name** **—** **must** **clear** **before** **`routines`** **if** **FK** **from** **runs** **→** **routines**)**.**
3. **`DELETE FROM routines`**.
4. **`DELETE FROM scene_models`** **(order** **deletes** **to** **satisfy** **foreign** **keys** **—** **child** **tables** **first** **per** **actual** **`ON DELETE`** **clauses**)**.**
5. **`DELETE FROM scenes`**.
6. **`DELETE FROM lights`**.
7. **`DELETE FROM models`**.
8. **Run** **the** **same** **insert** **logic** **as** **`SeedDefaultSamples`** **(**§3.8**)** **inside** **this** **transaction** **—** **if** **today’s** **`SeedDefaultSamples`** **opens** **its** **own** **transaction**, **refactor** **so** **the** **insert** **path** **can** **run** **on** **a** **supplied** **`*sql.Tx`** **(or** **equivalent**)** **without** **an** **intermediate** **`COMMIT`**.
9. **`COMMIT`**.

**Invariants:** **No** **commit** **unless** **steps** **2–8** **all** **succeed**; **on** **error**, **`ROLLBACK`** **and** **return** **500** **with** **a** **generic** **message** **(client** **may** **offer** **retry** **copy**)**.** **After** **200**, **`GET /api/v1/models`** **lists** **only** **the** **three** **samples**; **`GET /api/v1/scenes`** **is** **empty**; **`GET /api/v1/routines`** **is** **empty**.

**Security / abuse:** **Endpoint** **is** **unauthenticated** **(same** **as** **the** **rest** **of** **the** **MVP** **API**)**.** **Operator** **documentation** **should** **state** **that** **any** **client** **that** **can** **reach** **the** **server** **can** **invoke** **factory** **reset**; **future** **auth** **may** **protect** **this** **route** **first**.

**HTTP:** **`POST /api/v1/system/factory-reset`** **—** **empty** **body** **or** **`{}`** **accepted**; **200** **`{ "ok": true }`**. **Idempotent** **in** **effect:** **repeated** **calls** **still** **yield** **three** **deterministic** **samples** **per** **§3.8**.

### 3.15 Scene spatial API (dimensions, region queries, and region bulk updates) (**REQ-020**)

This section extends **REQ-015** scene composition with explicit API contracts for querying and updating lights by **scene-space geometry**. All region checks are performed against **derived** coordinates (**`sx/sy/sz`**), never canonical model coordinates (**`x/y/z`**).

**Canonical invariants reused from §3.12 / §3.13:**

1. **Derived coordinate rule:** **`sx = x + offset_x`**, **`sy = y + offset_y`**, **`sz = z + offset_z`**.
2. **No canonical rewrite:** Region reads and writes MUST NOT mutate stored canonical coordinates.
3. **State semantics:** Bulk update fields follow **REQ-011** exactly (**`on`**, canonical **`#RRGGBB`** **`color`**, **`brightness_pct`** in **0..100**).

**Geometry payload contracts (JSON):**

- **Cuboid query/update body**
  - **`position`**: `{ "x": number, "y": number, "z": number }` (minimum corner in scene space).
  - **`dimensions`**: `{ "width": number, "height": number, "depth": number }` (all strictly `> 0`).
- **Sphere query/update body**
  - **`center`**: `{ "x": number, "y": number, "z": number }`.
  - **`radius`**: number strictly `> 0`.
- All numeric inputs MUST be finite (`NaN` / `Inf` rejected).
- **Boundary policy resolution (REQ-020 open question):** inclusion is **inclusive**:
  - Cuboid: `sx ∈ [x, x+width]`, `sy ∈ [y, y+height]`, `sz ∈ [z, z+depth]`.
  - Sphere: Euclidean distance `dist((sx,sy,sz), center) <= radius`.

**Dimensions endpoint (resolves REQ-020 dimensions shape):**

- **`GET /api/v1/scenes/{id}/dimensions`** returns:
  - `origin`: always `{ "x": 0, "y": 0, "z": 0 }` (scene non-negative octant anchor).
  - `size`: `{ "width", "height", "depth" }` derived from §3.12 visual/query AABB (`max + 1m margin`).
  - `max`: `{ "x", "y", "z" }` same upper corner as §3.12 (`Mmax + 1` per axis).
  - `margin_m`: numeric margin value (`1` in current architecture baseline).
- This shape keeps axis-aligned dimensions unambiguous while explicitly disclosing origin metadata.
**Axis mapping for `size` (REQ-026, REQ-015 “right” convention):** **`width`** **=** **extent** **along** **scene** **+X** **(the** **default** **“right”** **axis** **when** **adding** **models** **after** **create)**; **`height`** **=** **extent** **along** **+Y** **(vertical** **up** **in** **the** **default** **three.js** **view** **—** **must** **match** **`SceneLightsCanvas`** **/** **`ModelLightsCanvas`** **orientation)**; **`depth`** **=** **extent** **along** **+Z** **(depth** **into** **the** **screen** **in** **the** **same** **default** **camera** **framing)**. **Python** **`scene.width`**, **`scene.height`**, **`scene.depth`** **MUST** **read** **these** **same** **three** **numbers** **from** **one** **`GET …/dimensions`** **response** **(cached** **in** **the** **worker** **for** **the** **iteration** **if** **needed)**.

**Read endpoints (scene coordinates only):**

- **`GET /api/v1/scenes/{id}/lights`** → flattened list:
  - Each item includes at minimum `{ "scene_id", "model_id", "light_id", "x", "y", "z", "sx", "sy", "sz", "on", "color", "brightness_pct" }`.
  - `x/y/z` are canonical for traceability; region matching and client spatial behavior MUST use `sx/sy/sz`.
- **`POST /api/v1/scenes/{id}/lights/query/cuboid`** and **`.../query/sphere`**:
  - Validate geometry payload then return only matching lights in the same flattened shape.

**Bulk update endpoints (all-or-nothing):**

- **`PATCH /api/v1/scenes/{id}/lights/state/cuboid`**
- **`PATCH /api/v1/scenes/{id}/lights/state/sphere`**
- **`PATCH /api/v1/scenes/{id}/lights/state/scene`** — **whole-scene** **uniform** update (**REQ-020** **+** **REQ-021**):
  - **Body:** **`{ "on"?, "color"?, "brightness_pct"? }`** with **at least one** field present (**same** **partial-merge** **semantics** **as** **per-light** **PATCH** **§3.9** **—** **omitted** **fields** **leave** **existing** **DB** **values** **unchanged**).
  - **Matching set:** **every** **light** **row** **reachable** **through** **current** **`scene_models`** **for** **`scene_id`** **(zero** **lights** **is** **valid** **—** **`updated_count: 0`**).
  - **Transaction:** **one** **`UPDATE`** **wave** **(or** **batched** **updates)** **per** **matched** **`(model_id, idx)`**; **rollback** **on** **any** **validation** **failure** **(e.g.** **invalid** **hex** **after** **merge**)**.
  - **Response:** **same** **shape** **as** **region** **bulk** **updates** **(recommended** **`updated_count` + `states[]`)** **—** **MAY** **omit** **`sx/sy/sz`** **in** **`states`** **if** **payload** **size** **is** **a** **concern**.

- **`PATCH /api/v1/scenes/{id}/lights/state/batch`** — **per-light** **updates** **scoped** **to** **one** **scene** (**REQ-020** **extension** **for** **REQ-021** **and** **integrators**):
  - **Body:** **`{ "updates": [ { "model_id": "<uuid>", "light_id": <int>, "on"?, "color"?, "brightness_pct"? }, … ] }`**.
  - **Rules:** **`updates`** **non-empty**; **no** **duplicate** **`(model_id, light_id)`** **in** **the** **array**; **each** **`(model_id, light_id)`** **must** **exist** **in** **`lights`** **and** **`scene_models`** **for** **`scene_id`**; **each** **element** **has** **≥ 1** **of** **`on`**, **`color`**, **`brightness_pct`**; **validate** **merged** **state** **per** **row** **like** **§3.9**/**§3.10**.
  - **Transaction:** **all** **rows** **updated** **in** **one** **`BEGIN…COMMIT`**; **on** **any** **error**, **rollback** **(no** **partial** **apply**)**.
  - **Response:** **`{ "updated_count": <int>, "states": [ { "model_id", "id", "on", "color", "brightness_pct" }, … ] }`** **in** **request** **order** **or** **sorted** **by** **`model_id`** **then** **`id`** **(document** **chosen** **order**)**.

- Request body for **cuboid**/**sphere** = geometry payload + at least one state field (`on`, `color`, `brightness_pct`).
- Execution model:
  1. Validate scene id exists and geometry/state payload is valid.
  2. Resolve matching lights by derived coordinates (`sx/sy/sz`) from current scene composition.
  3. In one DB transaction, apply validated state fields to all matched `(model_id, idx)` rows.
  4. On any validation/store failure, rollback entire operation (no partial updates).
- Recommended response shape:
  - `{ "updated_count": <int>, "states": [ { "model_id", "id", "on", "color", "brightness_pct", "sx", "sy", "sz" }, ... ] }`
  - `updated_count` MAY be `0` (valid no-op if region matches no lights).

**Error model (REQ-020 rule 9):**

- Geometry validation failure → **`400`** with `error.code = "validation_failed"` and actionable `details`.
- Scene missing → **`404`**.
- Any transactional failure after validation → **`500`** with generic message (no partial writes committed).

### 3.16 Scene routines — persistence, HTTP, scheduler (**REQ-021**, **shared** **with** **REQ-022**)

**Resolved requirements open questions (architecture):**

- **Machine-readable type** for the **first** routine: **`random_colour_cycle_all`** (display name **“Random colour cycle (all lights)”** **or** **equivalent** **in** **UI** **copy**).
- **REQ-022 storage (resolved):** **`python_scene_script`** **routine** **definitions** **live** **in** **the** **same** **`routines`** **table** **with** **`python_source`** **TEXT** **(see** **§3.17**)**; **same** **`routine_runs`** **start/stop** **HTTP** **as** **built-in** **types**.
- **Concurrency:** **At most one** **`routine_runs`** row with **`status = 'running'`** **per** **`scene_id`** (**any** **routine** **type**). **Starting** a **second** **routine** **on** **the** **same** **scene** **while** **one** **is** **running** → **`409`** **`scene_routine_conflict`** **with** **actionable** **`details`** **(existing** **`run_id`**, **`routine_id`)**.
- **Delete while running:** **`DELETE /api/v1/routines/{id}`** **is** **blocked** **`409`** **`routine_run_active`** **if** **any** **`routine_runs`** row **for** **that** **`routine_id`** **has** **`status = 'running'`**. **Stop** **the** **run** **first** **(or** **delete** **the** **scene** **after** **stopping**)**. **`PATCH /api/v1/routines/{id}`** **for** **`python_scene_script`** **uses** **the** **same** **`routine_run_active`** **rule** **(§3.2** **table**)**.

**Tables (logical, SQLite):**

| Table | Columns | Notes |
|-------|---------|--------|
| **`routines`** | `id` (TEXT **UUID** PK), `name` (TEXT **NOT NULL**), `description` (TEXT **NOT NULL**, **allow** **`""`**), `type` (TEXT **NOT NULL** — **`random_colour_cycle_all`**, **`python_scene_script`**, **or** **future** **identifiers**), `python_source` (TEXT **NOT NULL**, **`""`** **when** **`type`** **≠** **`python_scene_script`**), `created_at` (TEXT **RFC3339** UTC) | **Migration:** **add** **`python_source`** **with** **default** **`""`** **for** **existing** **rows**. |
| **`routine_runs`** | `id` (TEXT **UUID** PK), `routine_id` (TEXT **FK** → **`routines.id`** **ON DELETE RESTRICT**), `scene_id` (TEXT **FK** → **`scenes.id`** **ON DELETE CASCADE**), `status` (TEXT **NOT NULL**: **`running`** \| **`stopped`**), `started_at`, `stopped_at` (TEXT **RFC3339** **nullable** **for** **`stopped_at`**) | **UNIQUE** **partial** **index** **recommended:** **at** **most** **one** **`running`** **row** **per** **`scene_id`** **(enforce** **in** **store** **transaction** **+** **index** **or** **`UNIQUE(scene_id)`** **where** **`status='running'`** **if** **SQLite** **version** **supports** **partial** **UNIQUE**)**. **When** **scene** **deleted**, **runs** **for** **that** **scene** **cascade** **away** **(scheduler** **drops** **in-memory** **handle** **on** **next** **tick** **or** **via** **explicit** **cancel** **hook**)**. |

**Routine types (extensibility):**

- **`python_scene_script`:** **REQ-022** **—** **no** **server** **scheduler** **ticks** **(§3.17)**. **`POST …/start`** **inserts** **`routine_runs`** **`running`** **and** **returns** **`201`** **like** **other** **types** **but** **does** **not** **invoke** **Go** **colour** **logic**; **light** **changes** **happen** **only** **when** **a** **browser** **client** **executes** **the** **saved** **`python_source`** **in** **a** **Pyodide** **worker** **(§3.17)**. **If** **no** **client** **is** **running** **the** **worker**, **the** **row** **stays** **`running`** **without** **effects** **until** **stop** **or** **until** **a** **client** **loads** **the** **scene**/**editor** **UI** **and** **starts** **the** **loop** **(document** **this** **limitation** **in** **the** **Python** **routines** **help** **text** **§4.13**)**.
- **`random_colour_cycle_all`:** **REQ-021** **step** **(a)** **then** **(b)** **map** **to** **HTTP** **as** **follows** **(same** **semantics** **if** **the** **scheduler** **calls** **shared** **`store`** **code** **without** **loopback** **HTTP** **—** **see** **below**)**:**
  1. **Turn** **every** **light** **on** **at** **100%** **brightness** **with** **a** **valid** **hex** **colour:** **`PATCH /api/v1/scenes/{id}/lights/state/scene`** **with** **`{ "on": true, "brightness_pct": 100, "color": "#ffffff" }`** **(or** **any** **valid** **fixed** **hex**)**.
  2. **Assign** **each** **light** **an** **independent** **uniform** **random** **`#rrggbb`:** **build** **`updates[]`** **with** **one** **object** **per** **light** **`(model_id, light_id)`** **in** **the** **scene**, **each** **with** **`on: true`**, **`brightness_pct: 100`**, **`color`** **drawn** **uniformly** **from** **24-bit** **RGB** **(e.g.** **`crypto/rand`** **or** **`math/rand`** **seeded** **once** **at** **process** **start** **—** **document** **choice**)** **and** **validated** **with** **the** **same** **function** **as** **REQ-011** **writes**. **Apply** **`PATCH /api/v1/scenes/{id}/lights/state/batch`** **with** **that** **`updates`** **array** **(§3.15)**.
  3. **On** **each** **subsequent** **scheduler** **tick** **while** **`running`:** **repeat** **step** **2** **with** **fresh** **random** **`color`** **values** **(still** **`on: true`**, **`brightness_pct: 100`** **on** **every** **row** **so** **external** **off** **commands** **do** **not** **strand** **lights** **off** **while** **the** **routine** **runs**)**.

**REQ-021 rule 3 (scene API only):** **Routine** **automation** **MUST** **not** **use** **`PATCH /api/v1/models/{id}/lights/…`** **for** **effects** **that** **are** **part** **of** **the** **routine**. **Use** **only** **`PATCH /api/v1/scenes/{id}/lights/state/*`** **endpoints** **(including** **`…/scene`** **and** **`…/batch`** **in** **§3.15**)**. **Preferred** **implementation** **in** **`cmd/server`:** **a** **single** **shared** **`store`** **function** **invoked** **by** **both** **the** **HTTP** **handlers** **and** **the** **routine** **scheduler** **(no** **`http.Client`** **to** **localhost**)** **so** **validation** **and** **SQL** **stay** **identical**.

**Timer cadence (REQ-021 rule 5):** **`cmd/server`** **starts** **one** **`time.Ticker`** **with** **period** **`1s`** **after** **the** **HTTP** **server** **is** **ready**. **Each** **tick:** **for** **each** **`routine_runs`** **row** **with** **`status='running'`**, **load** **`routines.type`** **for** **`routine_id`**; **if** **`type`** **is** **`python_scene_script`**, **skip** **(no** **server** **work** **—** **§3.17**)**; **if** **`type`** **is** **`random_colour_cycle_all`**, **run** **the** **handler** **below**. **Ticks** **are** **not** **aligned** **to** **UTC** **second** **boundaries**; **at** **most** **one** **colour** **advance** **per** **~1** **s** **of** **scheduler** **time** **per** **eligible** **run**. **Under** **CPU** **load** **on** **Pi**, **a** **tick** **may** **slip** **(document** **best-effort** **timing** **in** **README**)**.

**Start/stop semantics:**

- **`POST …/start`:** **Validate** **scene** **+** **routine** **exist** **and** **`type`** **is** **implemented**. **If** **a** **`running`** **row** **already** **exists** **for** **`(scene_id, routine_id)`:** **`200`** **idempotent** **`{ "run_id", "status": "running" }`** **(no** **new** **row**)**. **Else** **if** **another** **`running`** **row** **exists** **for** **`scene_id`:** **`409`**. **Else** **insert** **`routine_runs`** **`running`**. **If** **`type`** **is** **`random_colour_cycle_all`**, **before** **`201`:** **run** **steps** **1** **and** **2** **above** **synchronously** **(shared** **`store`** **or** **internal** **call** **chain)** **so** **the** **client** **sees** **all** **lights** **on** **with** **per-light** **random** **colours** **immediately**; **on** **failure**, **rollback** **the** **insert** **and** **return** **`4xx`/`5xx`** **without** **leaving** **a** **`running`** **row**. **If** **`type`** **is** **`python_scene_script`**, **return** **`201`** **after** **commit** **(no** **mandatory** **server** **light** **mutation**)** — **client** **begins** **the** **Pyodide** **loop** **per** **§3.17**.
- **`POST …/stop`:** **Set** **`status=stopped`**, **`stopped_at=now`**. **Idempotent** **if** **already** **stopped**. **Scheduler** **skips** **stopped** **rows**.

**Scene deletion:** **`ON DELETE CASCADE`** **on** **`routine_runs.scene_id`** **cleans** **DB** **rows**; **server** **MUST** **remove** **any** **in-memory** **scheduler** **subscription** **for** **that** **`scene_id`** **(e.g.** **on** **next** **tick** **diff** **or** **`DeleteScene`** **callback**)**.

**UI support:** **§4.12** (**list** **+** **create** **all** **types** **through** **one** **`type`** **dropdown** **per** **REQ-023**)**; **§4.13** (**Python** **authoring** **+** **docs**)**; **§4.9** (**start/stop** **any** **type** **on** **scene**).

**Future routine types (REQ-021 rule 4, volumetric):** Types that **affect** **only** **a** **subset** **of** **lights** **MUST** **use** **`PATCH …/lights/state/cuboid`**, **`…/sphere`**, **and/or** **`…/batch`** **with** **membership** **derived** **from** **scene-space** **`sx/sy/sz`** **(§3.15)**. **`POST …/start`** **MAY** **accept** **an** **optional** **JSON** **body** **`{ "region"?: { "kind": "cuboid" \| "sphere", … } }`** **once** **such** **types** **exist**; **until** **then**, **unknown** **body** **fields** **on** **start** **SHOULD** **return** **`400`** **or** **be** **ignored** **per** **product** **JSON** **strictness** **(document** **chosen** **behavior**)**.

### 3.17 Python scene routines — editor, Pyodide worker, `scene` shim, loop, and forced stop (**REQ-022**, **REQ-030**)

**Execution placement (REQ-022 open question resolved):** **Python** **runs** **in** **the** **browser** **using** **[Pyodide](https://pyodide.org/)** **(WebAssembly)** **inside** **a** **dedicated** **`Worker`**. **Rationale:** **Preserves** **REQ-004** **—** **no** **`python3`** **interpreter** **shipped** **or** **required** **on** **the** **Pi** **for** **routines**; **light** **writes** **still** **go** **through** **the** **same** **Go** **§3.15** **validation** **as** **other** **clients**. **Trade-off:** **large** **static** **assets** **(~** **tens** **of** **MB** **compressed** **—** **see** **§6.2**)** **and** **CPU/RAM** **use** **on** **the** **client** **device**; **acceptable** **for** **REQ-022** **novice** **authoring** **on** **desktop** **or** **tablet** **and** **documented** **as** **heavy** **on** **low-end** **phones**.

**`scene` API (Python, bound per run):** **Inject** **a** **`scene`** **object** **into** **the** **user’s** **globals** **before** **each** **iteration** **(or** **once** **per** **worker** **lifetime** **with** **captured** **`sceneId`)**. **Most** **members** **are** **a** **thin** **layer** **over** **`pyodide.ffi.to_js`** **and** **`js.fetch`** **to** **same-origin** **`/api/v1/scenes/{sceneId}/…`**. **REQ-030** **resolved:** **`scene.random_hex_colour()`** **is** **synchronous**, **invokes** **no** **`fetch`**, **and** **is** **implemented** **entirely** **inside** **the** **worker** **using** **the** **embedded** **Python** **`random`** **module:** **each** **call** **returns** **`"#%06x"** **%** **`random.randrange(0x1000000)`** **(**lowercase** **hex** **digits** **—** **same** **string** **shape** **and** **uniform** **24-bit** **distribution** **as** **that** **expression** **in** **CPython**/**Pyodide**)**. **Canonical** **mapping** **(names** **in** **§4.13** **REQ-024** **API** **reference** **manifest** **MUST** **match** **these** **or** **document** **aliases**)**:**

| Python surface (canonical names) | HTTP (§3.15) |
|----------------------------------|--------------|
| **`scene.random_hex_colour()`** **(returns** **`str`,** **REQ-030)** | **None** **—** **local** **`random.randrange(0x1000000)`** **→** **`"#%06x"`** |
| **`scene.width`** **(float,** **meters)** | **`GET …/dimensions`** **→** **`size.width`** **(REQ-026** **—** **+X** **extent** **per** **§3.15** **mapping)** |
| **`scene.height`** **(float,** **meters)** | **`GET …/dimensions`** **→** **`size.height`** **(+Y** **extent)** |
| **`scene.depth`** **(float,** **meters)** | **`GET …/dimensions`** **→** **`size.depth`** **(+Z** **extent)** |
| **`scene.get_lights_within_sphere(center, radius)`** | **`POST …/lights/query/sphere`** **with** **`center`**, **`radius`** |
| **`scene.get_lights_within_cuboid(position, dimensions)`** | **`POST …/lights/query/cuboid`** |
| **`scene.get_all_lights()`** | **`GET …/lights`** |
| **`scene.set_lights_in_sphere(center, radius, on=…, color=…, brightness_pct=…)`** | **`PATCH …/lights/state/sphere`** |
| **`scene.set_lights_in_cuboid(…)`** | **`PATCH …/lights/state/cuboid`** |
| **`scene.set_all_lights(…)`** | **`PATCH …/lights/state/scene`** |
| **`scene.update_lights_batch(updates)`** | **`PATCH …/lights/state/batch`** **with** **`{ "updates": … }`** |

**Return** **types** **should** **mirror** **JSON** **(e.g.** **lists** **of** **dicts** **with** **`model_id`**, **`light_id`**, **`sx`**, **`sy`**, **`sz`**, **state** **fields)** **—** **exact** **shapes** **documented** **in** **§4.13** **for** **novices**.

**Loop:** **After** **`POST …/start`** **returns** **`201`**, **the** **UI** **that** **initiated** **start** **(or** **the** **scene** **detail** **page** **when** **it** **detects** **an** **active** **`python_scene_script`** **run** **via** **`GET …/routines/runs`)** **spawns** **(or** **reuses)** **the** **worker**, **loads** **`python_source`**, **and** **repeatedly** **`exec`** **or** **`runPythonAsync`** **the** **script** **body** **until** **stopped**. **Default** **gap** **between** **iterations:** **`asyncio.sleep(0.05)`** **(50** **ms)** **—** **tunable** **later** **via** **optional** **UI** **setting**; **document** **in** **README**.

**Cooperative stop:** **Between** **iterations**, **the** **worker** **checks** **whether** **`runId`** **is** **still** **`running`** **(lightweight** **`GET …/scenes/{id}/routines/runs`** **or** **host** **`postMessage`** **after** **host** **received** **`201`** **from** **`POST …/stop`)**. **On** **stop**, **exit** **loop** **and** **`postMessage({ done: true })`**.

**Forced stop (REQ-022 rule 8):** **If** **the** **worker** **does** **not** **finish** **within** **`T_force`** **after** **the** **user** **invokes** **stop** **(architecture** **default** **`T_force` = 5000` ms)** **—** **e.g.** **tight** **infinite** **loop** **with** **no** **`await`** **—** **the** **host** **calls** **`worker.terminate()`**. **Recommended** **ordering:** **`POST …/stop`** **(idempotent)** **then** **wait** **`T_force`** **then** **`terminate()`** **if** **still** **alive** **so** **DB** **state** **matches** **user** **intent** **even** **when** **the** **worker** **is** **stuck**.

**Editor** **(§4.13,** **REQ-022** **rules** **2–4):** **CodeMirror** **6** **only** **—** **the** **`@codemirror/*`** **major** **version** **6** **ecosystem** **with** **Python** **grammar** **(e.g.** **`@codemirror/lang-python`)**; **canonical** **mount** **in** **this** **product:** **`EditorView`** **from** **`@codemirror/view`** **with** **`EditorState.create`** **and** **`basicSetup`** **from** **the** **npm** **`codemirror`** **(v6)** **package**, **plus** **lint**/**completion**/**theme** **extensions** **(e.g.** **`@codemirror/lint`**, **`@codemirror/autocomplete`**, **`@codemirror/theme-one-dark`** **`oneDark`)**; **Monaco** **or** **any** **other** **non-CodeMirror** **primary** **editing** **widget** **for** **this** **buffer** **is** **not** **permitted**. **Diagnostics** **via** **Pyodide** **`ast.parse`** **on** **debounced** **edit**/**save**/**run** **or** **optional** **Pyright** **worker** **if** **bundle** **budget** **allows**; **completion** **for** **`scene.`** **methods** **from** **a** **static** **manifest** **derived** **from** **the** **table** **above**; **format** **on** **save** **via** **`black`** **in** **Pyodide** **when** **available**, **else** **indent** **normalization** **only** **—** **defaults** **on** **per** **REQ-022**.

**REQ-021** **rule** **3** **(scene** **API** **only):** **Python** **must** **not** **call** **`/models/.../lights`** **for** **routine** **effects**; **the** **`scene`** **shim** **only** **targets** **§3.15** **scene** **routes**.

### 3.18 High-throughput light state updates (**REQ-029**)

**Design target:** Workloads with **on the order of hundreds** of lights (**REQ-005** upper bound) and **multiple aggregate update cycles per second** across **writes** and **viewer refresh**, while staying **credible** on **Raspberry Pi 4** (**REQ-003**). Prefer **fewer, larger HTTP transactions** and **reused TCP connections** over **storms** of **per-light** **`PATCH`** calls.

**Write path (aggregate APIs — satisfies REQ-029 business rule 2 alongside REQ-011 per-light endpoints):**

- **Model scope:** **`PATCH /api/v1/models/{id}/lights/state/batch`** (**§3.10**) — one **SQLite** transaction, **many** **`ids`**, **same** **partial** **state** **fields**; **`POST …/lights/state/reset`** (**§3.11**) — **all** lights in **one** round-trip.
- **Scene scope:** **`PATCH /api/v1/scenes/{id}/lights/state/{cuboid|sphere|scene|batch}`** (**§3.15**) — **region**, **whole-scene**, or **explicit** **per-light** **rows** in **one** **request** **body**.
- **Granular control:** **`PATCH …/lights/{lightId}/state`** remains **required** (**REQ-011**); **integrators** **driving** **high-frequency** **multi-light** **effects** **SHOULD** **prefer** **batch**/**bulk** **routes** **above**.

**Connection reuse (REQ-029 business rule 3):**

- Go’s **`net/http.Server`** uses **HTTP/1.1 keep-alive** by **default** for **eligible** clients. **Document** in **README** that **integrators** **SHOULD** **reuse** connections (**Go** **`http.Client`** **with** **shared** **`Transport`**; **curl** **keep-alive**; **browser** **`fetch`** **—** **same-origin** **tabs** **typically** **pool** **per** **origin**).
- **Optional reverse proxy** (**§7**, **canonical** **production** **listener** **note** **in** **§1**): Terminate **TLS** and expose **HTTP/2** to **browsers** (**multiplexing** **many** **requests** **over** **one** **connection**) while **proxying** **HTTP/1.1** to **`127.0.0.1:8080`** **(or** **configured** **upstream**)** — **recommended** when **many** **parallel** **API** **calls** are **expected**.

**Observer path — freshness without polling the full state every tick (REQ-029 business rule 4):**

- **Shipped architectural default:** **Bounded polling** meets **REQ-012** **timeliness** for **typical** **UI** **use:** **§4.7** **optional** **`GET …/models/{id}/lights/state`** **every** **≤ 5 s** on **model** **detail**; **§4.9** **`GET …/scenes/{id}`** **every** **≤ 2 s** while **a** **routine** **run** **is** **active**. **Rationale:** **Simple**, **stateless** **on** **the** **server**, **low** **connection** **count** **on** **Pi**, **adequate** when **external** **write** **rates** **are** **modest** **or** **only** **the** **acting** **tab** **needs** **instant** **feedback** (**local** **`fetch`** **merge** **after** **PATCH** **—** **§4.7** **State** **sync**).
- **When** **sustained** **multi-Hz** **writes** **from** **external** **clients** **must** **propagate** **to** **open** **browser** **tabs** **without** **short-interval** **full** **state** **polling:** **implementor** **SHOULD** **add** **Server-Sent** **Events** (**SSE**): e.g. **`GET /api/v1/models/{id}/lights/events`** **or** **`GET /api/v1/scenes/{id}/lights/events`** with **`Content-Type: text/event-stream`**, **emitting** **`data:`** **JSON** **lines** **after** **successful** **`COMMIT`** (**minimal** **payload** **e.g.** **`{ "seq": <uint64> }`** **or** **`{ "revision": … }`**) **so** **clients** **using** **`EventSource`** **trigger** **one** **`GET`** **of** **authoritative** **state** **(or** **apply** **documented** **deltas** **if** **added** **later**)**. **WebSocket** **is** **acceptable** **if** **bidirectional** **integration** **is** **required**; **SSE** **is** **preferred** **for** **one-way** **server→client** **fan-out** **(simpler** **upgrade** **path** **behind** **some** **proxies**)**. **Multi-tab** **behavior** **(REQ-029** **open** **question**)**:** **each** **tab** **may** **open** **its** **own** **SSE** **connection**, **or** **the** **product** **may** **document** **a** **single-tab** **subscriber** **pattern** **with** **`BroadcastChannel`** **(advanced**)**.
- **SQLite** **limits:** **Writes** **serialize** **through** **one** **writer** **per** **process**; **batch** **endpoints** **reduce** **HTTP** **and** **transaction** **framing** **overhead** **but** **not** **underlying** **storage** **latency** **—** **see** **§6** **for** **Pi** **RAM/CPU** **and** **§9** **for** **body** **size** **and** **abuse** **bounds**.
- **REQ-031** **(redundant** **work** **elision):** **Equivalence** **checks**, **persistence** **short-circuits**, **client** **redraw** **skips**, **and** **future** **physical** **sync** **hooks** **are** **specified** **in** **§3.19**. **Observer** **paths** **(**polling**, **SSE**)** **SHOULD** **apply** **the** **same** **comparison** **when** **merging** **fetched** **state** **into** **three.js** **so** **repeated** **identical** **payloads** **do** **not** **rebuild** **meshes** **(**§4.7**, **§4.9**)**.

### 3.19 Redundant light-state elision and future physical sync (**REQ-031**)

**Goals:** (1) **Reduce** **three.js** **redraw** **lag** **by** **not** **rebuilding** **or** **re-uploading** **GPU** **state** **when** **effective** **per-light** **appearance** **would** **not** **change**. (2) **Reduce** **redundant** **SQLite** **`UPDATE`s** **and** **transaction** **work** **when** **the** **merged** **logical** **state** **equals** **the** **already** **stored** **row**. (3) **Reserve** **a** **clear** **extension** **point** **so** **a** **future** **WLED** **(**or** **equivalent**)** **bridge** **receives** **only** **meaningful** **changes**, **not** **no-op** **repeats** **(**device** **drivers** **remain** **out** **of** **scope** **until** **a** **later** **requirement**)**.

#### Canonical equivalence (logical state)

**Effective** **state** **for** **comparison** **is** **the** **triple** **`(on: bool, color: string, brightness_pct: float64)`** **after** **merge** **(**partial** **`PATCH`** **applied** **to** **current** **row** **per** **§3.9**)**.

**Normalization** **(shared** **by** **Go** **and** **TypeScript** **—** **implement** **one** **pure** **function** **per** **language**, **unit-tested**)**:

1. **`on`:** **boolean** **as** **stored** **/** **JSON** **decoded**.
2. **`color`:** **lowercase** **the** **hex** **body** **(**`#RRGGBB`** **→** **`#rrggbb`**)**; **reject** **invalid** **shapes** **before** **comparison** **(**validation** **unchanged** **§3.9**)**.
3. **`brightness_pct`:** **treat** **as** **numeric** **in** **[0,** **100]**; **for** **equality**, **compare** **at** **fixed** **tolerance** **e.g.** **`|a−b| ≤ 1e-9`** **(**or** **round** **to** **a** **documented** **decimal** **precision** **if** **JSON** **float** **noise** **appears**)**.

**Two** **triples** **are** **equivalent** **iff** **all** **three** **normalized** **fields** **match**.

**Not** **compared** **here:** **light** **`x,y,z`**, **scene** **offsets**, **model** **metadata** **—** **REQ-031** **applies** **only** **to** **per-light** **state** **fields** **above**.

#### Server (`internal/store` + handlers)

**Single-light** **`PATCH /api/v1/models/{id}/lights/{lightId}/state`:**

1. **Load** **current** **row**; **compute** **merged** **triple** **from** **body** **+** **row**.
2. **If** **merged** **≡** **current** **(**equivalence** **above**)**:** **skip** **`UPDATE`** **;** **return** **`200`** **with** **JSON** **body** **identical** **to** **today’s** **shape** **`{ "id", "on", "color", "brightness_pct" }`** **(**authoritative** **values** **=** **merged** **normalized** **)**. **MAY** **include** **an** **optional** **boolean** **`"unchanged": true`** **so** **integrators** **and** **tests** **can** **assert** **no-op** **behavior** **(**clients** **that** **ignore** **unknown** **fields** **are** **unchanged**)**.
3. **Else:** **`UPDATE`** **as** **today**; **`200`** **without** **`unchanged`** **or** **with** **`"unchanged": false`**.

**Batch** **`PATCH …/lights/state/batch` (**§3.10**):** **For** **each** **id**, **merge** **patch** **onto** **current** **row**; **collect** **only** **ids** **where** **merged** **≢** **current** **for** **`UPDATE`s**. **If** **the** **set** **is** **empty** **(**all** **requested** **lights** **already** **match**)**:** **`200`** **with** **`states[]`** **listing** **every** **requested** **id** **in** **ascending** **order** **with** **full** **objects** **(**no** **`UPDATE`** **)** **and** **optional** **`"unchanged_all": true`**. **If** **some** **change:** **run** **one** **transaction** **updating** **only** **the** **changed** **ids**; **response** **`states[]`** **covers** **all** **requested** **ids** **(**unchanged** **rows** **read** **without** **write**)**. **Optional** **`changed_count`** **/** **`unchanged_count`** **in** **JSON** **for** **telemetry**.

**Scene** **bulk** **routes** **(**§3.15**:** **`cuboid`**, **`sphere`**, **`scene`**, **`batch`**)**:** **Same** **pattern** **—** **resolve** **matching** **lights**, **merge** **patch** **per** **row**, **issue** **`UPDATE`s** **only** **for** **rows** **where** **merged** **≢** **stored**; **`updated_count`** **in** **the** **response** **counts** **rows** **actually** **written** **(**may** **be** **0** **when** **every** **match** **is** **already** **equivalent**)**. **Region** **match** **with** **zero** **lights** **remains** **distinct** **(**valid** **`updated_count: 0`** **with** **no** **state** **payload** **per** **existing** **§3.15** **wording**)**.

**`POST …/lights/state/reset` (**§3.11**):** **If** **every** **light** **is** **already** **at** **REQ-014** **defaults** **(**equivalence** **to** **default** **triple**)**:** **skip** **`UPDATE`** **;** **`200`** **with** **`states[]`** **as** **today** **(**optional** **top-level** **`unchanged_all`: true** **)**. **Else** **full** **`UPDATE`** **as** **today**.

**Routine** **scheduler** **(**§3.16**)** **and** **internal** **`store`** **calls:** **Reuse** **the** **same** **“compare** **before** **`UPDATE`”** **helper** **so** **random-colour** **ticks** **that** **by** **chance** **repeat** **a** **colour** **(**unlikely**)** **or** **any** **future** **routine** **that** **re-sends** **the** **same** **state** **does** **not** **touch** **SQLite** **unnecessarily**.

**Concurrency:** **Equivalence** **check** **+** **conditional** **`UPDATE`** **MUST** **run** **inside** **the** **same** **transaction** **that** **reads** **the** **row** **(**`BEGIN IMMEDIATE`** **or** **repeatable** **read** **pattern** **as** **today**)** **so** **two** **writers** **cannot** **both** **skip** **based** **on** **stale** **reads**.

#### Client (Next.js + three.js)

**Last-rendered** **cache:** **Keep** **a** **`Map`** **or** **parallel** **arrays** **of** **normalized** **triples** **(**or** **hashes**)** **aligned** **to** **`lights[i]`** **for** **the** **mounted** **model** **or** **scene** **view**. **On** **every** **authoritative** **merge** **(**successful** **`PATCH`**, **`GET`** **poll**, **SSE-triggered** **refetch**)**:

1. **For** **each** **light**, **if** **new** **triple** **≡** **last-rendered** **triple**, **skip** **any** **per-light** **material** **/** **`InstancedMesh`** **attribute** **work** **for** **that** **index** **(**including** **emissive** **recalc** **§4.7**)**.
2. **If** **all** **indices** **are** **unchanged** **(**whole** **payload** **equivalent**)**:** **skip** **full** **scene** **rebuild** **(**no** **new** **`BufferGeometry`**, **no** **re** **`setMatrixAt`** **storm** **for** **unchanged** **instances**)** **—** **still** **update** **React** **state** **if** **needed** **for** **reference** **equality** **(**prefer** **structural** **sharing** **/** **stable** **references** **when** **the** **server** **sent** **`unchanged`** **)**.
3. **On** **navigation** **away** **from** **model**/**scene** **detail**, **or** **after** **factory** **reset** **success**, **drop** **the** **cache**.

**Optimistic** **UI** **(**if** **any**)**:** **If** **the** **client** **pre-merges** **a** **user** **edit**, **do** **not** **send** **`PATCH`** **when** **the** **control** **value** **matches** **the** **already** **known** **authoritative** **triple** **(**reduces** **HTTP** **as** **well** **as** **DB** **work**)** — **still** **send** **when** **the** **user** **explicitly** **chooses** **“Apply”** **on** **bulk** **panels** **if** **the** **product** **defines** **that** **as** **always** **confirming** **(**optional** **UX** **choice** **—** **prefer** **skip** **for** **performance** **per** **REQ-031**)**.

**Python** **worker** **(**§3.17**):** **Before** **`fetch`** **on** **`scene.set_*`**, **the** **shim** **MAY** **compare** **intended** **patch** **to** **last** **known** **state** **for** **that** **light** **(**from** **last** **`get_*`** **or** **cached** **run** **state**)** **and** **short-circuit** **(**no** **HTTP**)** **when** **equivalent** **—** **document** **in** **README** **if** **enabled** **(**trade-off:** **stale** **cache** **vs** **fewer** **writes** **—** **must** **invalidate** **on** **external** **tab** **updates** **or** **rely** **on** **server** **no-op** **anyway**)**.

#### Observer path alignment (**REQ-029** + **REQ-031**)

**When** **`GET …/lights/state`** **or** **`GET …/scenes/{id}`** **returns** **data** **unchanged** **from** **the** **client’s** **last** **merge** **(**byte-for-byte** **or** **per-light** **equivalence**)**:** **do** **not** **trigger** **a** **full** **three.js** **rebuild** **—** **update** **timestamps**/**refs** **only** **if** **required** **for** **React**. **This** **complements** **bounded** **polling** **without** **adding** **jank** **when** **nothing** **moved**.

#### Future physical sync (WLED or equivalent) — design hook only

**No** **WLED** **code** **ships** **under** **this** **requirement.** **Architecture** **for** **a** **later** **feature:**

- **Subscribe** **inside** **the** **Go** **process** **(**or** **a** **sidecar** **documented** **later**)** **to** **“logical** **light** **state** **committed** **with** **`unchanged`** **false”** **events** **—** **i.e.** **only** **after** **a** **successful** **transaction** **that** **actually** **wrote** **at** **least** **one** **row** **whose** **effective** **triple** **changed**.
- **Maintain** **per** **(**device**, **segment**, **or** **mapped** **fixture** **id**)** **a** **last-applied** **triple** **(**or** **WLED-specific** **projection** **of** **it**)** **on** **the** **bridge**; **suppress** **network** **I/O** **to** **the** **device** **when** **the** **new** **logical** **state** **maps** **to** **the** **same** **device** **command** **as** **the** **last** **successful** **apply**.
- **Use** **the** **same** **normalization** **rules** **as** **above** **where** **meaningful**; **device** **quantization** **(**e.g.** **8-bit** **per** **channel**)** **MAY** **require** **a** **documented** **secondary** **rounding** **step** **specific** **to** **the** **hardware** **layer**.

```mermaid
flowchart LR
  subgraph go["Go binary"]
    H[HTTP handlers]
    ST[store: merge + compare]
    DB[(SQLite)]
    HOOK["Future: PhysicalSync hook\n(only if row changed)"]
    H --> ST
    ST --> DB
    ST -.->|"after real COMMIT"| HOOK
  end
  subgraph ext["Future external"]
    WLED["WLED or equivalent\n(not implemented now)"]
  end
  HOOK -.->|"only non-equivalent\ncommands"| WLED
```

---

## 4. Next.js + Tailwind (build-time / authoring)

### 4.1 Stack

- **Next.js App Router** under `web/app/` for structure and layouts.
- **Tailwind** + **TypeScript** as today.
- **`next.config`:** **`output: 'export'`** (static HTML export). **`images`:** use **`unoptimized: true`** or avoid `next/image` features incompatible with export.

### 4.2 Runtime-incompatible patterns (must not ship)

- **`next start`**, **standalone** Node output, **SSR** dependencies, **Route Handlers** or **middleware** that must execute on Node for **navigations**, **rewrites** to a separate upstream for **HTML**.
- **Server Components** that **fetch at request time** without a static **generateStaticParams** strategy — for MVP, prefer **client `fetch`** to **`/api/v1/...`**.

### 4.3 Data access patterns (shipped UI)

| Pattern | Use |
|--------|-----|
| **Client `fetch('/api/v1/...')`** | Primary pattern for reactive UI (**REQ-002**) against **same Go origin**. **Reuse** **connections** **per** **origin** **(browser** **default)** **and** **prefer** **batch**/**bulk** **endpoints** **for** **many** **light** **updates** **(**§3.18**, **REQ-029**)**. |
| **Optional `EventSource` (future)** | **If** **SSE** **endpoints** **from** **§3.18** **ship**, **subscribe** **for** **push** **notifications** **then** **`GET`** **authoritative** **state** **—** **keeps** **tabs** **fresh** **under** **high** **external** **write** **rates** **without** **sub-second** **polling**. |
| **Build-time static** | `generateStaticParams` / SSG **only** if values are fixed at build; API must still be available at runtime for live data where needed. |

### 4.4 Environment

- **Runtime** UI has **no** `process.env.NEXT_PUBLIC_*` server injection from Node; **build-time** env may set **`NEXT_PUBLIC_API_BASE`** to **`''`** (same origin) so client code calls relative URLs.

### 4.5 Responsive behavior (**REQ-002**)

Unchanged intent: **Tailwind breakpoints**, **touch targets**, **`"use client"`** for interactivity.

### 4.6 Models UI (**REQ-002**, **REQ-006**, **REQ-010**, **REQ-011**, **REQ-012**, **REQ-014**)

- **Routes (App Router):** e.g. **`/models`** (list), **`/models/new`** (upload form: **name** text input + **file** input), **`/models/[id]`** (detail: metadata; **§4.8** **paginated** light **table** + **§4.7** **3D** view; per-light or bulk controls per **REQ-011** / **REQ-013**; **Reset lights** **button** calling **`POST …/lights/state/reset`** per **§3.11** (**REQ-014**), **reachable** on **mobile** / **tablet** / **desktop** without **hover-only** use). **Model** **delete** **control:** on **`409`** **`model_in_scenes`**, **show** **`error.message`** **and** **`details.scenes`** (names/links to **`/scenes/{id}`**) per **§3.13**.
- **Client data:** **`"use client"`** pages/components call **`fetch`** with **`GET`**, **`POST`** (**`FormData`** for multipart), **`DELETE`**, and **`PATCH`** (**JSON**) against **`/api/v1/models…`** on the **same origin** (**§4.3**).
- **Feedback:** Inline / banner display of **400** / **409** **`message`** from API; loading states on list, detail, upload, and delete (**REQ-002**).
- **Navigation:** **Primary** **IA** **is** **the** **collapsible** **left** **nav** **in** **§4.11** (**Models**, **Scenes**, **Routines**, **Options**, **home** **`/`** **if** **distinct**); **no** **hover-only** **paths** **to** **these** **destinations** (**REQ-002**, **REQ-018**).

### 4.7 Three.js visualization on model detail (**REQ-010**, **REQ-012**, **REQ-016**, **REQ-019**, **REQ-028**, **REQ-031**)

**Dependency:** Declare **`three`** in **`web/package.json`** as a **direct** dependency (satisfies REQ-010 business rule 2). Pin a **stable semver** range in lockfile; bump intentionally when upgrading. Using **`@react-three/fiber`** / **`@react-three/drei`** is **optional**—if used, **`three`** MUST still appear **directly** in **`dependencies`** (not only as a transitive peer).

**Where:** The **model detail** route (**`/models/[id]`** or equivalent, e.g. **`/models/detail?id=`** for static export) MUST mount a **client-only** visualization after **`GET /api/v1/models/{id}`** returns **`lights`** including **`on`**, **`color`**, **`brightness_pct`** (**§3.6**, **§3.9**). All geometry uses **world-space meters** (**REQ-005** / **REQ-009**).

**SSR / static export:** **WebGL** is **browser-only**. The Three.js entry MUST run only on the client: e.g. **`"use client"`** + **`WebGLRenderer`** after mount, or **`next/dynamic`** with **`ssr: false`**. **Do not** assume **`window`**, **`document`**, or **GPU** during **Node** prerender of that subtree.

#### Geometry and materials (REQ-010 rules 4–5, 7; REQ-012; REQ-028)

- **Per-light marker:** Each light is a **sphere** with **diameter 0.02 m** (**2 cm**) → **`SphereGeometry`** with **radius `0.01`**.
- **Canonical visualization grey:** **`#D0D0D0`** — parse once to **`THREE.Color`** for **wire segments** and **off** spheres (**REQ-010** / **REQ-012**).
- **On vs off (REQ-012):**
  - **`on === true`:** **Filled** **opaque** (or effectively **opaque**) **surface** **that** **also** **satisfies** **REQ-028** **(**emissive** **glow** **—** **see** **subsection** **below**)**. **Canonical** **material:** **`MeshStandardMaterial`** **(or** **`MeshPhysicalMaterial`**) **`side: FrontSide`**, **`transparent: false`** (or **opacity 1**), **`metalness: 0`**, **`roughness`** **in** **~0.25–0.45** **(pick** **constants** **and** **document** **in** **code**)**. **Base** **`color`:** parse API **`color`** **hex** **to** **`THREE.Color`**, **convert** **to** **linear** **working** **space** **as** **appropriate** **for** **the** **renderer**, **then** **multiply** **RGB** **by** **`brightness_pct / 100`** **for** **the** **diffuse** **albedo** **(REQ-012)**, **clamping** **per** **channel** **to** **[0,1]**. **`MeshBasicMaterial`** **without** **an** **additional** **documented** **glow** **technique** **does** **not** **meet** **REQ-028**. At **`brightness_pct === 0`**, the **on** sphere MAY appear **nearly** **black** **with** **minimal** **emissive**; **product** **“off”** **in** **the** **persistence** **sense** **remains** **`on: false`**.
  - **`on === false`:** **Filled** **sphere** (same **geometry** as **on**) with **`MeshBasicMaterial`** (**or** equivalent): **`color`** = **`#D0D0D0`**, **`transparent: true`**, **`opacity: 0.15`** (**85%** transparency per requirements), **`depthWrite: false`** if needed to reduce **z-fighting** with **segments** and **neighbors**. **`emissive`** **MUST** **be** **black** **and** **`emissiveIntensity`** **MUST** **be** **0** **if** **the** **material** **exposes** **those** **fields** **(REQ-028** **rule** **4)**. **Do not** present **off** lights as **more** visually **prominent** than **on** lights or than **wire** segments (**REQ-012**).
- **Emissive glow (REQ-028, on lights only):**
  - **`emissive`:** **Set** **from** **the** **same** **hex** **as** **the** **light** **(typically** **the** **linear** **RGB** **of** **the** **parsed** **`color`** **before** **or** **after** **brightness** **scaling** **—** **choose** **one** **rule** **and** **apply** **consistently** **so** **hue** **matches** **the** **sphere)**.
  - **`emissiveIntensity`:** **Map** **`brightness_pct`** **(0–100)** **monotonically** **to** **a** **non-negative** **intensity** **e.g.** **`k * (brightness_pct / 100)`** **with** **documented** **`k`** **(tune** **~0.6–1.2** **so** **100%** **reads** **clearly** **bright** **on** **`VIZ_VIEWPORT_BG`** **and** **low** **percents** **read** **weaker)** **or** **`k * max(ε, brightness_pct/100)^γ`** **with** **small** **ε** **and** **γ** **≥** **1** **for** **perceptual** **spacing**. **Higher** **`brightness_pct`** **MUST** **never** **produce** **a** **weaker** **glow** **than** **a** **lower** **`brightness_pct`** **for** **the** **same** **`color`** **(REQ-028** **rule** **3)**.
  - **Scene** **lights** **in** **three.js:** **Use** **modest** **`AmbientLight`** **and** **optionally** **a** **very** **low-intensity** **`DirectionalLight`** **so** **the** **non-emissive** **shading** **does** **not** **flatten** **or** **wash** **out** **the** **emissive** **read**; **avoid** **bright** **key** **lights** **that** **make** **all** **spheres** **look** **evenly** **lit** **plastic**.
  - **Clipping** **and** **many** **100%** **lights** **(REQ-028** **open** **question):** **Apply** **a** **per-instance** **cap** **on** **`emissiveIntensity`** **after** **the** **brightness** **curve** **if** **needed**; **use** **renderer** **`outputColorSpace = THREE.SRGBColorSpace`** **(or** **current** **equivalent)** **and** **sane** **`toneMapping`** **(**e.g.** **`ACESFilmic`** **or** **`Reinhard`**) **with** **`toneMappingExposure`** **tuned** **so** **the** **frame** **does** **not** **blow** **out** **to** **flat** **white**. **Prefer** **this** **over** **mandatory** **full-screen** **bloom** **passes** **on** **Pi-class** **or** **integrated** **GPUs** **unless** **profiling** **shows** **headroom** **(**§6.5**)**.
  - **Instancing:** **When** **using** **`InstancedMesh`** **for** **on** **lights**, **either** **one** **shared** **`MeshStandardMaterial`** **with** **per-instance** **`setColorAt`** **/** **`instanceColor`** **for** **base** **and** **a** **custom** **`onBeforeCompile`** **or** **`ShaderMaterial`** **variant** **for** **per-instance** **`emissiveIntensity`**, **or** **document** **an** **equivalent** **approach** **(e.g.** **rebuild** **instance** **attributes** **on** **state** **change** **—** **O(n)** **acceptable** **for** **n** **≤** **1000**)**.
- **Draw all lights (no omission):** For **n** lights (**n ≤ 1000**), the scene MUST contain **exactly n** **visible** markers at the **correct** **positions**—**no** decimation. **Rendering strategy (implementor picks one):**
  - **A.** **Two** **`InstancedMesh`** **layers**: (**1**) **on** lights — **`InstancedMesh`** + **`instanceColor`** (and **brightness** factored per instance or via **custom** attribute); (**2**) **off** lights — second **`InstancedMesh`** with **shared** **`#D0D0D0`**, **`opacity 0.15`**, **non-wireframe** **material**. When **`on`** toggles, **move** instances between **layers** or **rebuild** **both** from **authoritative** **state** (**O(n)** OK for **n ≤ 1000**).
  - **B.** **Up to n** **individual** **`Mesh`** **nodes** — acceptable if **performance** on **Pi-class** **clients** remains acceptable.
- **Wire polyline (REQ-005 chain, REQ-010):** **`LineSegments`** (or **`LineBasicMaterial`** **lines**) only between **(i, i+1)** for **i = 0 … n−2**. **Colour** **`#D0D0D0`**, **`transparent: true`**, **`opacity: 0.15`**; **linewidth** where supported is **thin** (note: **WebGL** **line** **width** is often **1** px); segments MUST read **subtler** than **spheres**. **Style** does **not** vary with **on/off** (**REQ-012** **out** **of** **scope** for segment state).
- **Framing (baseline for REQ-016):** **Implement** a **pure** **helper** **`applyDefaultFraming(bounds: THREE.Box3, camera, controls, viewportWidth, viewportHeight)`** **(name** **local** **to** **codebase**)** **that:**
  - Sets **`controls.target`** **to** **the** **center** **of** **`bounds`** **(axis-aligned** **bounding** **box** **of** **all** **light** **positions** **±** **0.01** **m** **sphere** **radius**)**.
  - Positions **`camera`** **outside** **the** **bounds** **on** **a** **stable** **direction** **(e.g.** **normalized** **`(1, 1, 1)`** **from** **center** **scaled** **so** **the** **entire** **bounds** **fits** **in** **the** **vertical** **and** **horizontal** **FOV** **with** **a** **small** **padding** **factor** **≥** **1.1**)**—** **document** **chosen** **vector** **and** **padding** **in** **code** **comments** **so** **reset** **and** **first** **load** **match**.
  - Calls **`camera.updateProjectionMatrix()`**, **`controls.update()`** **as** **needed** **after** **`ResizeObserver`** **driven** **size** **changes**.
- **Initial load:** **After** **`lights`** **arrive**, **compute** **`bounds`**, **then** **`applyDefaultFraming`** **once**.
- **Reset camera (REQ-016):** **Secondary** **button** **(e.g.** **label** **“Reset camera”)** **next** **to** **the** **canvas** **toolbar** **or** **overlay** **corner** **calls** **`applyDefaultFraming`** **again** **using** **the** **current** **`lights`** **bounds** **(recomputed** **if** **data** **changed**)**—** **no** **`fetch`**. **Repeat** **clicks** **yield** **the** **same** **baseline** **for** **the** **same** **`lights`** **and** **viewport** **size** **(REQ-016** **rule** **4**)**. **Architectural** **resolution** **for** **REQ-016** **open** **question:** **reset** **affects** **only** **camera** **and** **`OrbitControls`** **state**; **implementor** **MAY** **clear** **the** **hover**/**tap** **label** **for** **less** **confusion** **but** **is** **not** **required** **to** **clear** **list** **selection** **or** **pagination**.

#### State sync (REQ-012 rule 3; REQ-029 observer path; REQ-031 elision)

- **After a successful `PATCH …/lights/{lightId}/state`**, **`PATCH …/lights/state/batch`** (**§3.10**, **REQ-013**), or **`POST …/lights/state/reset`** (**§3.11**, **REQ-014**) initiated from **this** **browser** **session**, the **client** MUST **merge** the **JSON** **response** (or **refetch** **`GET …/lights/state`** / **model** **detail**) and **update** **three.js** **meshes** **before** the **next** **`requestAnimationFrame`** **paint** **following** **the** **`fetch`** **resolution** (i.e. **no** **indefinite** **staleness** **after** **confirmed** **write**). **REQ-031:** **When** **the** **merged** **per-light** **state** **is** **equivalent** **to** **what** **was** **already** **rendered** **(**§3.19** **normalization**)** **,** **skip** **redundant** **mesh**/**material**/**instance** **rebuild** **for** **those** **lights** **(**§3.19** **client** **subsection**)**.
- **Concurrent** **sessions** **(another** **tab** **or** **REST** **client):** **Optional** **`setInterval`** **poll** of **`GET /api/v1/models/{id}/lights/state`** every **≤ 5 s** while the **detail** **route** **is** **mounted**; if **absent**, **manual** **browser** **refresh** **still** **shows** **truth** — **document** **in** **README** **that** **live** **multi-user** **sync** **may** **lag** **up** **to** **one** **poll** **period**. **Under** **sustained** **multi-Hz** **writes** **from** **other** **clients**, **this** **polling** **alone** **may** **not** **meet** **every** **integrator** **expectation** **—** **see** **§3.18** **(optional** **SSE**)** **and** **§8.19**. **REQ-031:** **Poll** **responses** **that** **match** **the** **last** **merged** **state** **should** **not** **force** **a** **full** **three.js** **rebuild** **(**§3.19** **observer** **alignment**)**.

#### Picking, hover, and touch (REQ-010 rule 6; REQ-012 rule 4)

- **Raycasting:** Use **`THREE.Raycaster`** against the **same** **meshes** **used** **for** **filled**/**wire** **markers** (**union** **of** **both** **`InstancedMesh`** **layers** **or** **all** **`Mesh`** **targets**). Map **`instanceId`** / **object** **back** **to** **`lights[i]`**.
- **Desktop hover:** Show **`id`**, **`x`**, **`y`**, **`z`**; **MAY** **also** **show** **`on`**, **`color`**, **`brightness_pct`** (**REQ-012** **open** **question** **resolved** **here** **as** **optional** **but** **recommended** **for** **debuggability**).
- **Touch / tablet equivalent:** Same **tap** **threshold** **behavior** **as** **before**; **pinned** **label** **may** **include** **state** **fields**.

**Interaction (orbit, REQ-002, REQ-016):** Retain **`OrbitControls`** for **rotate / zoom / pan**; **touch** gestures as today. **REQ-016** **reset** **does** **not** **change** **polar**/**azimuth** **limits** **unless** **already** **set**; **if** **limits** **exist**, **ensure** **default** **camera** **pose** **remains** **within** **them** **or** **temporarily** **relax** **limits** **during** **reset** **(implementor** **picks** **simplest** **consistent** **behavior**)**.**

**Layout:** **WebGL canvas** in a **responsive** container (**full width**, **bounded height** via **`min-h-[…]`** / **`max` viewport height**). **`ResizeObserver`** updates **camera.aspect** and **`renderer.setSize`**.

**Viewport background (REQ-019):** **Independently** **of** **REQ-018** **shell** **light**/**dark**, **set** **`scene.background`** **and** **`WebGLRenderer.setClearColor`** **(same** **RGB)** **to** **one** **fixed** **dark** **grey** **that** **reads** **clearly** **as** **grey** **(not** **white** **or** **near-white**)**—** **architecture** **default** **`#262626`** **(≈** **Tailwind** **`neutral-800`**)**, **centralized** **as** **a** **named** **constant** **e.g.** **`VIZ_VIEWPORT_BG`** **in** **`web/lib/`** **so** **model** **and** **scene** **canvases** **stay** **aligned**. **If** **the** **canvas** **is** **letterboxed** **inside** **a** **wrapper**, **style** **that** **wrapper** **(and** **any** **padding** **region** **inside** **the** **same** **visual** **frame** **as** **the** **WebGL** **element**)** **with** **the** **same** **hex** **so** **light** **shell** **mode** **does** **not** **show** **white** **margins** **around** **the** **3D** **view**. **Do** **not** **tie** **this** **colour** **to** **`html`** **`dark`** **or** **to** **shell** **`bg-white`**/**`bg-gray-900`** **tokens**.

**Edge cases:** **`n === 0`:** Initialize renderer + empty scene + overlay copy; **no** spheres or segments. **`n === 1`:** One sphere; **no** segments. **WebGL unavailable:** Inline / console error acceptable per prior REQ-010 note.

**Testing note:** Unit-test **pure** helpers (**hex** + **brightness** → **THREE.Color**, **segment** **vertex** **pairs**, **instance** **matrices**, **pick** **index**) in **`web/lib/`**; **manual** verify **on/off** **appearance**, **PATCH** **refresh**, **hover** and **tap** on **mobile** + **desktop**.

### 4.8 Model detail: paginated light list, go-to-id, multi-select, bulk apply (**REQ-013**)

**Data source:** The **detail** page already loads **all** **`lights`** (positions + state) via **`GET /api/v1/models/{id}`** (**§3.6**, **§3.9**). **Pagination** is **purely client-side**: slice **`lights`** by **`page`** and **`pageSize`** so **REQ-005**’s **n ≤ 1000** stays a **single** **HTTP** **fetch** while the **table** shows one page at a time.

**Page size:** Expose **exactly three** choices: **25**, **50**, **100** lights per page (**REQ-013** rule 2). **Default** **`pageSize = 50`**. Changing **page size** resets **`page`** to **1** or **clamps** so the **first** light on the **prior** page remains **visible** if possible (implementor picks **simplest**: reset to **1** is acceptable).

**Trivial case:** For **n = 1** light, **REQ-013** allows **omitting** pagination chrome or showing a **single** row **without** **page** controls.

**Navigation controls:** **Previous** / **Next** (disabled at bounds); **“Page X of Y”** (derived from **n** and **`pageSize`**); optional **first/last** if space allows (**REQ-002** **touch** targets).

**Go to light id:** Text field (or number input) + **“Go”** / submit: parse **integer**; if **id ∉ [0, n−1]**, show **inline** **error** (do **not** change **page**); else set **`page = floor(id / pageSize) + 1`** (0-based indexing: **`page`** for **id** is **`⌊id / pageSize⌋ + 1`** in **1-based** UI terms). **Scroll** or **focus** the **row** for that **id** if the table is scrollable within the page.

**Multi-select:**

- **Per row:** **checkbox** in the **first** column (**touch-friendly**, **REQ-002**).
- **Header:** **“Select page”** toggles **all** **ids** on the **current** **page** only.
- **Shift+click** (desktop): **optional** **range** select between **last** **anchor** and **clicked** row **on the same page** only (reduces accidental cross-page ambiguity).
- **Cross-page selection (architectural resolution for REQ-013 open question):** Maintain **`selectedIds: Set<number>`** in **component** **state**. **Changing pages** does **not** clear **selection**. **Clear selection** button and **unmounting** the **detail** route **do** clear. **Bulk apply** sends **`selectedIds`** (as **array**) to **`PATCH …/state/batch`**.

**Bulk apply panel:** When **`selectedIds.size ≥ 1`**, show **controls** mirroring **REQ-011** fields: **on/off** toggle, **colour** (**`#RRGGBB`** input or picker), **brightness** (**0–100**). **Apply** calls **`PATCH /api/v1/models/{id}/lights/state/batch`** (**§3.10**). **Disable** **Apply** while **request** **in** **flight**; on **success**, **merge** **`states`** into **`lights`** and refresh **table** + **§4.7** scene (**REQ-012**). On **400**, show **API** **`message`**.

**Accessibility / responsive:** Table **MAY** **stack** as **cards** on **narrow** **viewports**; **checkboxes** and **bulk** **panel** remain **reachable** **without** **hover-only** **affordances** (**REQ-013** rule 7).

### 4.9 Scenes UI and composite three.js (**REQ-015**, **REQ-010**, **REQ-012**, **REQ-016**, **REQ-019**, **REQ-021**, **REQ-022**, **REQ-023**, **REQ-027**, **REQ-028**, **REQ-029**, **REQ-031**)

- **Routes:** **`/scenes`** (list), **`/scenes/new`** (**create** **flow**: **scene** **name** + **ordered** **multi-select** **of** **≥ 1** **model** **—** **no** **per-row** **offset** **inputs**; **submit** **`POST /api/v1/scenes`** **with** **`models`** **in** **that** **order** **and** **let** **the** **server** **compute** **offsets** **per** **§3.12**), **`/scenes/[id]`** (**detail** **composite** **view** **+** **optional** **offset** **editing** **via** **`PATCH …/models/{modelId}`** **after** **create**).
- **Routines on scene detail (REQ-021, REQ-022, REQ-023):** **Panel** **or** **toolbar** **region** **with** **`GET /api/v1/routines`** **(populate** **a** **`<select>`** **or** **list** **of** **definitions** **—** **include** **`python_scene_script`** **rows** **with** **a** **clear** **label** **e.g.** **“Python”**)**; **`POST …/scenes/{id}/routines/{routineId}/start`** **and** **`POST …/scenes/{id}/routines/runs/{runId}/stop`** **with** **Font Awesome** **icons** **on** **buttons**. **After** **`GET /api/v1/scenes/{id}/routines/runs`**, **show** **active** **run** **(routine** **name** **+** **Stop**)** **or** **“No** **routine** **running”**. **On** **`409`** **`scene_routine_conflict`**, **surface** **`error.message`** **and** **suggest** **stopping** **the** **other** **routine** **first**. **Link** **or** **button** **“Edit** **Python** **routine”** **→** **`/routines/python/[routineId]`** **(§4.13)** **when** **the** **selected** **definition** **is** **`python_scene_script`** **—** **this** **is** **an** **optional** **non-primary** **shortcut** **(REQ-023** **open** **question**)** **and** **MUST** **not** **replace** **creating** **new** **Python** **definitions** **via** **the** **`/routines`** **create** **flow** **with** **`type`** **dropdown** **(§4.12)**.
- **Data:** **`GET /api/v1/scenes`**, **`POST /api/v1/scenes`**, **`GET /api/v1/scenes/{id}`**, **`POST …/models`**, **`PATCH …/models/{modelId}`**, **`DELETE …/models/{modelId}`**, **`DELETE /api/v1/scenes/{id}`** per **§3.13**. **On** **`409`** **`scene_last_model`**, **show** **modal** **copy** **that** **removing** **the** **last** **model** **deletes** **the** **entire** **scene**; **on** **confirm**, **`DELETE /api/v1/scenes/{id}`** **then** **redirect** **to** **`/scenes`**.
- **Live updates while a routine runs (REQ-012, REQ-029, REQ-031):** **While** **`/scenes/[id]`** **is** **mounted** **and** **`GET …/routines/runs`** **shows** **`running`**, **poll** **`GET /api/v1/scenes/{id}`** **every** **`≤ 2 s`** **(architectural** **default** **between** **§4.7** **5** **s** **and** **built-in** **routine** **1** **s** **cadence**)** **and** **merge** **`items[].lights`** **into** **`SceneLightsCanvas`** **state** **so** **REQ-012** **is** **not** **indefinitely** **stale** **during** **automation**. **Apply** **§3.19** **per-light** **equivalence** **so** **unchanged** **poll** **payloads** **do** **not** **rebuild** **the** **composite** **canvas**. **For** **`python_scene_script`**, **the** **Pyodide** **worker** **(§3.17)** **also** **drives** **updates** **via** **`fetch`** **—** **polling** **remains** **the** **source** **of** **truth** **for** **other** **tabs** **and** **for** **merging** **server** **state** **into** **the** **canvas**. **Stop** **polling** **when** **no** **run** **is** **active** **or** **on** **unmount**. **Integrators** **updating** **the** **same** **scene** **at** **high** **frequency** **via** **§3.15** **SHOULD** **use** **bulk**/**batch** **routes** **(**§3.18**)** **rather** **than** **per-light** **model** **`PATCH`** **storms**; **optional** **scene-scoped** **SSE** **(**§3.18**/**§8.19**)** **reduces** **need** **for** **faster** **polling** **when** **implemented**.
- **Composite** **three.js:** **Refactor** **or** **duplicate** **§4.7** **patterns** **into** **a** **`SceneLightsCanvas`** **(or** **extend** **`ModelLightsCanvas`**) **that** **accepts** **`items[]`**: **for** **each** **model**, **build** **the** **same** **2** **cm** **spheres** **and** **`#D0D0D0`** **`opacity`** **0.15** **segments** **between** **consecutive** **`id`** **only** **within** **that** **model**, **using** **`sx`, `sy`, `sz`** **from** **API** **(or** **client-composed** **positions** **identical** **to** **server** **rules**). **No** **segments** **between** **models**. **Per-light** **state** **materials** **match** **§4.7** (**REQ-012**, **REQ-028** **emissive** **rules**). **Apply** **the** **same** **REQ-019** **fixed** **dark-grey** **viewport** **treatment** **as** **§4.7** **(scene** **background** **+** **renderer** **clear** **+** **letterbox** **wrapper**)**. **Picking** **must** **identify** **which** **model** **and** **which** **light** **id** **for** **hover**/**tap** **(show** **scene** **coordinates** **and** **model** **id** **+** **light** **`id`** **as** **needed**).
- **Framing (REQ-016):** **Fit** **camera** **to** **§3.12** **AABB** **`[0,0,0]`** **–** **`(Mmax+1)`** **per** **axis** **plus** **marker** **radius** **margin** **using** **the** **same** **`applyDefaultFraming`** **pattern** **as** **§4.7** **but** **with** **bounds** **derived** **from** **scene-space** **positions** **`(sx,sy,sz)`** **for** **all** **lights**. **A** **“Reset camera”** **control** **on** **the** **scene** **canvas** **re-invokes** **that** **same** **fit** **(no** **API** **call**)**.**
- **Add** **model** **control:** **calls** **`POST …/scenes/{id}/models`** **without** **offsets** **to** **get** **default** **+X** **placement** **or** **with** **explicit** **integers** **after** **user** **edit**. **Placement** **inputs** **validate** **≥ 0** **client-side** **for** **fast** **feedback**; **authoritative** **errors** **from** **API** **400**.
- **REQ-002:** **Same** **touch**/**pointer** **expectations** **as** **model** **detail**; **no** **hover-only** **blocking** **flows** **for** **add**/**remove**/**confirm**.

### 4.10 Options and factory reset (**REQ-017**)

**Route:** **`/options`** **(App Router)** **or** **equivalent** **single** **page** **with** **`<h1>Options</h1>`** **(or** **product** **title** **+** **“Options”**)** **and** **a** **short** **explanation** **that** **this** **area** **holds** **destructive** **maintenance** **actions**.

**Information architecture:** **“Options”** **lives** **in** **the** **same** **collapsible** **left** **navigation** **as** **Models**, **Scenes**, **and** **Routines** (**§4.11**); **when** **the** **aside** **is** **collapsed** **on** **mobile**, **open** **it** **with** **the** **burger** **to** **reach** **primary** **destinations** **(REQ-002**, **REQ-018**).

**Factory reset row:** **Primary** **button** **or** **destructive-styled** **control** **labeled** **Factory reset** **(or** **Reset all data**)** **opens** **a** **native** **`<dialog>`** **or** **modal** **with:**
- **Title** **e.g.** **“Erase all data?”**
- **Body** **text** **stating** **explicitly** **that** **all** **models** **(including** **uploads**)** **,** **all** **scenes**, **and** **all** **saved** **routines** **(built-in** **and** **Python,** **including** **any** **routine** **run** **records**)** **will** **be** **permanently** **deleted**, **that** **the** **action** **cannot** **be** **undone**, **and** **that** **after** **completion** **only** **the** **three** **default** **sample** **models** **will** **remain**.
- **Cancel** **(default** **focus** **on** **desktop** **where** **appropriate**)** **closes** **with** **no** **request**.
- **Confirm** **(e.g.** **button** **label** **Erase everything**)** **submits** **`POST /api/v1/system/factory-reset`** **(disable** **while** **in-flight** **to** **prevent** **double** **submit**)**.**

**Client behavior on success:** **On** **`200`**, **clear** **any** **client** **cached** **model**/**scene** **detail** **state** **(React** **query** **cache** **or** **router** **refresh**)**; **navigate** **to** **`/models`** **via** **`router.push`** **or** **`router.replace`** **and** **show** **a** **non-blocking** **success** **banner** **e.g.** **“All data was reset. Sample models were restored.”** **This** **resolves** **the** **REQ-017** **post-reset** **navigation** **open** **question** **as** **the** **architectural** **default**.

**Client behavior on failure:** **Show** **API** **`error.message`** **or** **generic** **failure** **text**; **leave** **user** **on** **Options** **with** **dialog** **closed** **or** **open** **per** **UX** **consistency**.

**REQ-017** **typed** **phrase** **(e.g.** **type** **RESET**)**:** **Not** **required** **for** **MVP** **per** **this** **architecture** **(optional** **hardening** **later**)**.**

**Accessibility:** **Modal** **must** **trap** **focus** **or** **use** **native** **`dialog`** **with** **visible** **Cancel**/**Confirm**; **Escape** **cancels** **if** **the** **implementor** **enables** **it** **(recommended).**

### 4.11 Application shell, themes, navigation, and Font Awesome (**REQ-018**)

**Goal:** One **client-side** **shell** wraps **all** **App Router** **pages** (**`web/app/layout.tsx`** **mounts** **`AppShell`** **as** **`"use client"`** **wrapper** **or** **layout** **composition** **that** **does** **not** **break** **static** **export**): **branding**, **theme**, **burger**/**aside**, **and** **icon** **conventions** **stay** **consistent** **across** **models**, **scenes**, **and** **options**.

#### Layout structure

- **Top** **header** **(full** **width,** **sticky** **optional):**
  - **Burger** **`<button>`** **(left):** **toggles** **`navOpen`** **state**; **icon** **`faBars`** **(solid)** **or** **equivalent** **from** **the** **same** **Font Awesome** **Free** **kit**.
  - **Branding** **(after** **burger):** **`FontAwesomeIcon`** **with** **`faLightbulb`** **from** **`@fortawesome/free-regular-svg-icons`** — **same** **glyph** **as** [Font Awesome lightbulb classic regular](https://fontawesome.com/icons/lightbulb?f=classic&s=regular). **Adjacent** **visible** **title** **text** **must** **be** **exactly** **`Domestic Light & Magic`** **(REQ-018** **rule** **5**)**.**
  - **Theme** **toggle** **`<button>`** **(right** **or** **next** **to** **branding** **per** **space):** **icons** **`faSun`**/**`faMoon`** **(solid)** **or** **one** **icon** **that** **flips** **with** **`aria-pressed`**; **toggle** **between** **light** **and** **dark** **modes**.
- **Left** **`<aside>`** **(primary** **navigation):**
  - **When** **expanded:** **vertical** **list** **of** **`next/link`** **entries** **(or** **buttons** **that** **`router.push`**) **for** **`/`** **(home),** **`/models`**, **`/scenes`**, **`/routines`**, **`/options`** — **each** **row** **is** **a** **button-styled** **control** **with** **a** **Font Awesome** **icon** **+** **label** **(REQ-018** **rule** **7**)**. **Do** **not** **add** **a** **separate** **nav** **destination** **that** **only** **creates** **Python** **routines** **(REQ-023** **—** **Python** **is** **chosen** **on** **`/routines/new`** **via** **`type`** **`<select>`** **§4.12**)**.**
  - **When** **collapsed** **(narrow** **viewport):** **aside** **is** **off-screen** **or** **`hidden`** **except** **as** **an** **overlay** **drawer** **(full-height** **`fixed`** **panel** **`z-index`** **above** **`main`**) **opened** **by** **burger**; **tapping** **a** **backdrop** **or** **a** **close** **control** **dismisses** **the** **drawer** **(REQ-018** **responsive** **notes** — **no** **focus** **trap** **without** **escape**)**.**
  - **When** **collapsed** **(wide** **desktop):** **optional** **pattern** **—** **narrow** **rail** **(icons** **only)** **vs** **full** **labels**; **burger** **still** **toggles** **between** **those** **two** **widths** **so** **REQ-018** **“collapsible** **left** **menu”** **is** **satisfied** **even** **if** **default** **is** **expanded** **at** **`lg+`**.**
- **`<main>`** **fills** **remaining** **width** **with** **page** **content**; **apply** **shell** **background**/**text** **tokens** **here** **per** **REQ-018**. **three.js** **viewports** **(REQ-019)** **do** **not** **inherit** **shell** **background** **for** **the** **WebGL** **clear** **/** **scene** **fill**:** **they** **use** **the** **fixed** **dark-grey** **policy** **in** **§4.7** **regardless** **of** **light**/**dark** **shell**.

#### Theming (Tailwind)

- **Mechanism:** **Enable** **Tailwind** **`darkMode: 'class'`** **(v3)** **or** **the** **v4** **equivalent** **so** **light** **=** **absence** **of** **`dark`** **on** **`<html>`**, **dark** **=** **`class="dark"`** **on** **`<html>`** **(set** **from** **a** **small** **client** **effect** **on** **mount** **+** **on** **toggle**)**.**
- **Initial** **vs** **persisted** **(REQ-018** **rule** **1**)**:** **On** **load**, **if** **`localStorage`** **`dlm-theme`** **is** **`light`** **or** **`dark`**, **apply** **that** **value** **and** **ignore** **`prefers-color-scheme`** **for** **the** **shell**. **If** **the** **key** **is** **absent** **or** **invalid**, **read** **`window.matchMedia('(prefers-color-scheme: dark)')`** **(or** **equivalent**)** **and** **apply** **`dark`** **when** **it** **matches**, **else** **`light`** **when** **the** **API** **exists** **but** **does** **not** **match** **dark**; **if** **no** **media-query** **signal** **is** **available**, **default** **shell** **to** **`light`**. **When** **the** **user** **presses** **the** **theme** **toggle**, **write** **`dlm-theme`** **to** **`light`** **or** **`dark`** **and** **update** **`<html>`** **immediately**. **Optional** **product** **enhancement** **(not** **required** **by** **REQ-018**)**:** **a** **third** **“Use** **system**” **control** **that** **removes** **`dlm-theme`** **and** **re-subscribes** **to** **`prefers-color-scheme`** **changes**.
- **Tokens** **(implement** **via** **Tailwind** **`bg-*`**, **`text-*`**, **`border-*`** **with** **`dark:`** **variants):**
  - **Light:** **`bg-white`** **(or** **`bg-neutral-50`**) **for** **shell** **+** **`main`**; **`text-gray-900`** **(or** **`text-neutral-900`**) **for** **primary** **reading** **text** **and** **nav** **labels**.
  - **Dark:** **`bg-gray-900`** **for** **shell** **+** **`main`** **—** **dark** **grey** **background** **per** **REQ-018** **(not** **`#000`** **as** **the** **only** **choice**)**; **`text-white`** **for** **primary** **text**. **Cards**/**panels** **inside** **`main`** **MAY** **use** **`bg-gray-800`** **for** **elevation** **contrast** **against** **`bg-gray-900`** **(architecture** **allows** **one** **step** **lighter** **grey** **for** **nested** **surfaces**)**.**
- **Contrast:** **Keep** **WCAG-minded** **pairings** **for** **shell** **body** **text** **vs** **surface** **(REQ-018**)**. **three.js** **backdrop** **is** **governed** **by** **REQ-019** **(§4.7**)** **and** **is** **not** **required** **to** **match** **shell** **grey**; **pick** **helper**/**grid** **colours** **that** **stay** **visible** **on** **`VIZ_VIEWPORT_BG`** **without** **competing** **with** **lights** **and** **segments**.

#### Font Awesome (delivery and licensing)

- **Packages** **(npm,** **bundled** **into** **static** **export**)**:** **`@fortawesome/fontawesome-svg-core`**, **`@fortawesome/react-fontawesome`**, **`@fortawesome/free-regular-svg-icons`** **(contains** **`faLightbulb`**), **`@fortawesome/free-solid-svg-icons`** **(e.g.** **`faBars`**, **`faSun`**, **`faMoon`**, **`faTrash`**, **`faUpload`**, **`faPlus`)**. **Pin** **versions** **in** **`package-lock.json`**; **respect** **[Font Awesome Free license](https://fontawesome.com/license/free)** **(attribution** **in** **`README`** **or** **About** **if** **required** **by** **license** **at** **ship** **time**)**.**
- **Alternative** **not** **preferred** **for** **MVP:** **Kit** **script** **tag** **from** **CDN** **—** **adds** **runtime** **network** **dependency** **and** **complicates** **offline**/**Pi** **use**; **SVG-in-JS** **tree-shaking** **above** **is** **the** **default** **architecture**.
- **Button** **rule** **(REQ-018** **rule** **7**)**:** **Every** **`<button>`** **used** **for** **actions** **(submit,** **delete,** **cancel,** **reset,** **camera** **reset,** **pagination,** **bulk** **apply,** **modal** **confirm/cancel,** **factory** **reset,** **go** **to** **id,** **etc.**)** **MUST** **render** **a** **`<FontAwesomeIcon`** **`icon={…}`** **`/>`** **before** **or** **after** **the** **visible** **label** **(consistent** **placement** **per** **component** **family**)**. **Exempt:** **native** **form** **fields** **(`<input>`**, **`<textarea>`**, **`<select>`**, **`type="file"`**) **and** **plain** **text** **links** **without** **button** **styling**. **Button-styled** **links** **`className`** **matching** **primary**/**secondary** **button** **treatment** **MUST** **also** **include** **an** **icon**. **Icon-only** **buttons** **MUST** **have** **`aria-label`** **or** **visually** **hidden** **text**.

#### Interaction with existing sections

- **§4.6–§4.9, §4.12–§4.13:** **Replace** **ad-hoc** **page** **headers** **with** **reliance** **on** **`AppShell`** **title** **area** **where** **redundant**; **page** **`<h1>`** **MAY** **remain** **for** **route** **topic** **(e.g.** **“Models”)** **below** **the** **global** **brand** **strip** **or** **omit** **if** **the** **nav** **already** **disambiguates** **(implementor** **chooses** **minimal** **duplication**)**.**
- **§4.7** **/§4.9** **“Reset** **camera”** **and** **other** **toolbar** **buttons:** **each** **gets** **a** **Font Awesome** **icon** **per** **above**.

#### Static export note

- **Font Awesome** **SVG** **icons** **are** **pure** **React** **+** **tree-shaken** **JS** — **compatible** **with** **`output: 'export'`**; **no** **server** **runtime** **required**.

```mermaid
sequenceDiagram
  actor User as User device
  participant Inline as Inline layout script
  participant LS as localStorage
  participant MQ as prefers-color-scheme media query
  participant Root as html element
  participant Shell as AppShell (React client)

  User->>Inline: Parse HTML first paint
  Inline->>LS: getItem dlm-theme
  alt dlm-theme is light or dark
    LS-->>Inline: stored value
    Inline->>Root: apply dark class or omit for light per stored value
  else missing or invalid
    Inline->>MQ: prefers-color-scheme dark query
    MQ-->>Inline: matches or not
    Inline->>Root: dark class if dark matches else light shell
  end
  User->>Shell: React hydration
  Shell-->>User: Shell matches html class from inline script

  User->>Shell: Toggle theme
  Shell->>Root: classList toggle dark
  Shell->>LS: setItem dlm-theme light or dark
  Shell-->>User: Shell and main surfaces update via Tailwind dark variants

  User->>Shell: Tap burger
  Shell->>Shell: setNavOpen true or false
  Shell-->>User: Aside drawer or rail expands or collapses
```

### 4.12 Routines management UI (**REQ-021**, **REQ-023**)

- **Routes:** **`/routines`** **(App Router)** — **list** **`GET /api/v1/routines`** **in** **a** **table** **or** **card** **list** **(name**, **type**, **description** **truncated**, **created** **date**)**. **`/routines/new`** **(or** **equivalent** **single** **“New** **routine”** **flow** **reachable** **only** **from** **the** **list** **page** **—** **no** **second** **primary** **entry** **point** **for** **Python-only** **creation)** **hosts** **the** **create** **form** **below** **(REQ-023** **—** **one** **path** **for** **all** **new** **definitions**)**.
- **Create (REQ-021 + REQ-023):** **Before** **submit**, **the** **user** **MUST** **choose** **routine** **kind** **from** **one** **labeled** **`type`** **`<select>`** **(native** **or** **accessible** **custom** **listbox** **with** **same** **semantics)** **that** **lists** **every** **creatable** **`type`** **the** **server** **accepts** **on** **`POST /api/v1/routines`**, **including** **`random_colour_cycle_all`** **(label** **e.g.** **“Random** **colour** **cycle** **(all** **lights)”)** **and** **`python_scene_script`** **(label** **e.g.** **“Python** **(scene** **API)”)**. **Additional** **built-in** **types** **MUST** **appear** **here** **when** **added** **to** **the** **API**. **Submit** **`POST /api/v1/routines`** **with** **`name`** **(required)**, **`description`** **(optional** **empty** **string)**, **`type`**, **and** **`python_source`** **when** **`type`** **is** **`python_scene_script`** **(may** **be** **`""`** **per** **§3.2**)**. **When** **`python_scene_script`** **is** **selected**, **after** **successful** **`201`**, **`router.push(`/routines/python/${id}`)`** **(or** **equivalent)** **to** **the** **§4.13** **editor** **is** **the** **preferred** **pattern** **—** **avoid** **a** **standalone** **primary** **“New** **Python** **routine”** **`<button>`** **in** **the** **shell** **or** **list** **header** **that** **bypasses** **this** **`type`** **dropdown** **(REQ-023** **rule** **3**)**. **Reject** **unknown** **`type`** **values** **client-side** **to** **match** **server** **400**.
- **Delete:** **Per-row** **delete** **button** **`DELETE /api/v1/routines/{id}`**; **on** **`409`** **`routine_run_active`**, **show** **message** **to** **stop** **the** **run** **on** **the** **scene** **first** **(user** **navigates** **via** **scene** **detail** **§4.9**)**.
- **Python** **rows:** **Per-row** **“Edit”** **or** **row** **click** **navigates** **to** **`/routines/python/[id]`** **(§4.13)**. **Duplicate** **action** **`POST /api/v1/routines`** **with** **copied** **`name`** **(suffix** **“** **(copy)”)**, **`description`**, **`type`**, **`python_source`**.
- **Run**/**stop** **are** **not** **required** **on** **this** **page** **(optional** **“Open** **scenes”** **link**)**; **primary** **run** **UX** **is** **§4.9** **scene** **detail**.
- **REQ-002**/**REQ-018:** **All** **actions** **use** **buttons** **with** **Font Awesome** **icons**; **no** **hover-only** **essential** **steps**.

**REQ-023 — create flow (browser → API):**

```mermaid
flowchart LR
  subgraph UI["Next.js /routines/new"]
    A["New routine"] --> B["Labeled type select"]
    B --> C{"type value"}
    C -->|random_colour_cycle_all| D["POST /api/v1/routines"]
    C -->|python_scene_script| E["POST with python_source"]
    D --> F["List or success"]
    E --> G["Navigate /routines/python/id"]
  end
  subgraph API["Go httpapi"]
    D --> H[(store routines)]
    E --> H
  end
```

### 4.13 Python routine editor, API reference below editor, unified run and live viewport (**REQ-022**–**REQ-031**, **REQ-023**)

- **Routes (App Router):** **`/routines/python/[id]`** **(edit** **existing** **—** **primary** **destination** **after** **creating** **a** **`python_scene_script`** **row** **via** **§4.12**)**. **`/routines/python/new`** **MAY** **remain** **as** **a** **redirect** **to** **`/routines/new`** **with** **`type`** **pre-selected** **or** **as** **a** **bookmarkable** **helper** **that** **immediately** **`POST`s** **a** **placeholder** **definition** **then** **redirects** **to** **`/routines/python/[id]`** **—** **it** **MUST** **not** **be** **the** **only** **documented** **primary** **way** **to** **start** **a** **new** **Python** **routine** **(REQ-023** **—** **canonical** **create** **UX** **is** **§4.12** **`type`** **dropdown**)**.
- **Vertical** **page** **order** **(REQ-002,** **REQ-022,** **REQ-024,** **REQ-027):** **(1)** **toolbar** **+** **persistence** **actions**; **(2)** **CodeMirror** **editor** **(full-width** **primary** **code** **area)**; **(3)** **`<section id="python-scene-api-catalog">`** **or** **equivalent** **landmark** **—** **the** **REQ-024** **API** **reference** **—** **placed** **immediately** **below** **the** **editor** **in** **DOM** **order** **with** **no** **other** **primary** **workflow** **block** **between** **editor** **and** **reference** **except** **a** **single** **heading**/**divider** **if** **needed**; **(4)** **one** **unified** **“run** **and** **watch”** **region** **(REQ-027)** **after** **the** **reference** **section** **containing** **together:** **a** **single** **`GET /api/v1/scenes`** **`<select>`** **(or** **accessible** **combobox)** **that** **is** **the** **only** **scene** **target** **for** **both** **routine** **start**/**stop** **and** **`SceneLightsCanvas`**, **Start**/**Stop** **buttons**, **`SceneLightsCanvas`** **(§4.9** **reuse)**, **Reset** **scene** **lights**, **Reset** **camera**. **Forbidden:** **two** **parallel** **top-level** **sections** **that** **separate** **“run** **in** **scene”** **from** **“visual** **debug”** **with** **duplicate** **scene** **pickers** **or** **viewports**. **Mobile** **(REQ-002):** **stack** **(1)–(4)** **vertically**; **optional** **accordion** **only** **if** **every** **panel** **remains** **discoverable** **and** **does** **not** **place** **the** **API** **reference** **above** **the** **editor** **or** **split** **the** **unified** **run**/**viewport** **into** **two** **primary** **accordions** **for** **the** **same** **workflow**. **Desktop** **MAY** **widen** **the** **editor** **and** **reference** **but** **MUST** **not** **move** **REQ-024** **above** **the** **editor** **or** **interleave** **other** **primary** **blocks** **between** **editor** **and** **reference**.
- **Editor** **stack** **(mandatory** **per** **REQ-022):** **CodeMirror** **6** **only** — **Primary** **pattern** **in** **this** **repo:** **client** **component** **(e.g.** **`web/components/PythonCodeMirrorEditor.tsx`)** **creates** **`EditorView`** **in** **`React.useEffect`**, **`EditorState.create({ doc, extensions })`** **with** **`basicSetup`** **from** **`codemirror`**, **`python()`** **from** **`@codemirror/lang-python`**, **`lintGutter`**/**`linter`**/**`autocompletion`** **from** **`@codemirror/lint`** **/** **`@codemirror/autocomplete`**, **`oneDark`** **(or** **equivalent)** **from** **`@codemirror/theme-one-dark`**, **`keymap.of([indentWithTab])`**, **and** **`EditorView.updateListener`** **for** **`onChange`**; **expose** **`EditorView`** **ref** **or** **callback** **to** **parent** **so** **REQ-024** **insert** **can** **`dispatch`** **transactions**. **Sync** **external** **`value`** **when** **it** **differs** **from** **the** **document** **(e.g.** **after** **`GET /api/v1/routines/{id}`)**. **Alternatives:** **a** **thin** **React** **wrapper** **such** **as** **`@uiw/react-codemirror`** **MAY** **replace** **hand-rolled** **`EditorView`** **lifecycle** **if** **the** **buffer** **remains** **pure** **CodeMirror** **6**. **Do** **not** **use** **Monaco** **or** **other** **non-CodeMirror** **editors**. **Wire** **`scene.`** **completion** **from** **the** **same** **manifest** **as** **§3.17** **/** **REQ-024** **(parameter** **snippets** **for** **novices)**. **Format** **button** **and/or** **format-on-save** **per** **§3.17**.
- **Instructional** **copy** **(REQ-022** **rule** **10):** **Section** **headings**, **form** **labels**, **primary** **tooltips**, **empty** **states**, **and** **short** **inline** **help** **on** **this** **page** **MUST** **use** **simple** **language** **appropriate** **for** **a** **twelve-year-old** **who** **has** **just** **started** **Python** **—** **short** **sentences**, **common** **words**, **brief** **explanations** **when** **a** **technical** **term** **is** **required**, **and** **no** **long** **expository** **paragraphs** **in** **page** **chrome**.
- **Default** **template** **(REQ-025):** **Define** **`PYTHON_ROUTINE_DEFAULT_SOURCE`** **(or** **equivalent)** **in** **`web/`** **as** **the** **initial** **`doc`** **when** **the** **client** **creates** **`python_scene_script`** **with** **`python_source: ""`** **(or** **omit** **and** **let** **client** **substitute** **before** **`PATCH`)** **—** **content** **MUST** **demonstrate** **`await scene.set_lights_in_sphere(...)`** **(or** **sync** **wrapper** **if** **architecture** **uses** **sync** **shim)** **with** **reasonable** **`center`**/**`radius`**, **setting** **`on`**, **`color`** **(canonical** **`#rrggbb`)**, **and** **`brightness_pct`**. **For** **the** **random** **demo** **colour,** **SHOULD** **use** **`colour = scene.random_hex_colour()`** **(**REQ-030**)** **rather** **than** **`import random`** **and** **`"#%06x"** **%** **`random.randrange(0x1000000)`** **alone** **—** **so** **beginners** **see** **the** **documented** **helper** **first**. **Every** **template** **line** **or** **logical** **block** **MUST** **include** **brief** **Python** **`#`** **comments** **matching** **the** **brevity** **standard** **for** **REQ-024** **samples**. **Optional** **toolbar** **“Reset** **to** **template”** **MAY** **replace** **the** **buffer** **after** **confirm**.
- **API** **reference** **(REQ-024):** **Rendered** **from** **the** **same** **ordered** **manifest** **as** **§3.17** **(every** **`scene`** **property** **and** **method,** **including** **`scene.random_hex_colour`** **per** **REQ-030**)**. **UI** **requirements:** **(a)** **picker** **—** **`<select>`**, **searchable** **list**, **or** **equivalent** **—** **so** **exactly** **one** **catalog** **entry** **is** **in** **focus** **at** **a** **time**; **(b)** **detail** **area** **showing** **plain-language** **description**, **parameters**/**returns** **at** **novice** **level**, **and** **at** **least** **one** **sample** **usage** **string** **that** **includes** **Python** **`#`** **comments** **with** **short**, **non-verbose** **explanations**; **(c)** **Insert** **example** **button** **(REQ-018** **icon** **+** **label** **in** **simple** **wording)** **that** **inserts** **the** **currently** **shown** **sample** **via** **`EditorView.dispatch`** **of** **a** **CodeMirror** **6** **`Transaction`** **(**e.g.** **`insert`**, **`replaceSelection`**, **or** **equivalent** **change** **spec**)** **at** **the** **main** **selection** **anchor** **when** **the** **editor** **view** **has** **focus** **and** **a** **defined** **anchor** **(implementor** **MAY** **replace** **the** **current** **selection** **or** **insert** **at** **caret** **only** **—** **document** **the** **choice** **when** **implemented)**; **if** **no** **caret** **is** **available**, **append** **at** **the** **end** **of** **the** **document**. **Sample** **rendering** **MAY** **use** **read-only** **CodeMirror** **or** **styled** **`<pre>`**. **Heading** **e.g.** **“Scene** **API** **(**pick** **one** **to** **read** **more**)**”** **with** **anchor** **`#python-scene-api-catalog`**. **Implementor** **MUST** **keep** **manifest** **and** **§3.17** **in** **sync** **(single** **TS** **module** **exporting** **rows** **recommended)**.
- **Unified** **live** **viewport** **(REQ-027,** **REQ-028,** **REQ-031):** **The** **single** **scene** **`<select>`** **in** **the** **unified** **region** **determines** **`sceneId`** **for** **`SceneLightsCanvas`** **fed** **by** **`GET /api/v1/scenes/{id}`** **(per-light** **state** **+** **`sx/sy/sz`)**. **After** **each** **successful** **`PATCH`** **from** **the** **Pyodide** **worker** **for** **that** **`sceneId`**, **host** **`postMessage`** **/** **`iterationComplete`** **and** **refetch** **`GET …/scenes/{id}`** **or** **`GET …/lights`** **to** **merge** **into** **the** **canvas** **(REQ-012**/**REQ-028** **timeliness** **class** **—** **no** **indefinite** **staleness** **after** **writes)** **;** **apply** **§3.19** **so** **unchanged** **refetches** **do** **not** **rebuild** **the** **canvas**. **Reset** **scene** **lights** **→** **`PATCH /api/v1/scenes/{sceneId}/lights/state/scene`** **with** **`{ "on": false, "color": "#ffffff", "brightness_pct": 100 }`** **then** **merge**; **does** **not** **auto** **`POST …/stop`** **—** **user** **stops** **the** **run** **separately** **or** **accepts** **next** **iteration** **may** **overwrite**. **Reset** **camera** **→** **`applyDefaultFraming`** **(REQ-016,** **client-only)**. **Viewport** **usable** **with** **no** **active** **run** **(static** **inspection)** **and** **with** **an** **active** **run** **(live** **updates)**.
- **Persistence** **actions** **(toolbar):** **Load** **`GET /api/v1/routines/{id}`** **on** **mount**; **Save** **`PATCH /api/v1/routines/{id}`** **(or** **initial** **`POST`** **on** **new)**; **Duplicate** **—** **`POST /api/v1/routines`** **with** **copied** **fields**; **Delete** **`DELETE …`** **(same** **409** **rules)**; **Load** **list** **is** **the** **`/routines`** **page** **—** **this** **page** **MAY** **include** **a** **`<select>`** **of** **other** **Python** **routines** **for** **quick** **switch** **(optional)**.
- **Start**/**Stop** **(same** **`sceneId`** **as** **canvas):** **Optional** **`?scene=id`** **from** **`/scenes/[id]`** **(§4.9)** **pre-selects** **the** **unified** **scene** **`<select>`**. **Start** **`POST …/scenes/{sceneId}/routines/{routineId}/start`** **then** **attach** **§3.17** **worker** **with** **`sceneId`**, **`runId`**, **`routineId`**. **Stop** **`POST …/stop`** **then** **cooperative** **shutdown** **then** **`worker.terminate()`** **after** **`T_force`** **if** **needed**.
- **REQ-018:** **Toolbar** **and** **reference** **buttons** **(Save,** **Format,** **Run/Stop,** **Duplicate,** **Delete,** **Insert** **example,** **Reset** **scene** **lights,** **Reset** **camera** **where** **present)** **each** **include** **a** **visible** **Font** **Awesome** **icon**.

**Vertical** **layout** **(summary** **—** **REQ-024** **+** **REQ-027):**

```mermaid
flowchart TB
  T[Toolbar and persistence]
  E[CodeMirror Python editor]
  R[REQ-024 API reference: picker, detail, commented sample, Insert example]
  U[REQ-027 unified region: one scene select, Start or Stop, SceneLightsCanvas, resets]
  T --> E
  E --> R
  R --> U
```

---

## 5. UI ↔ API coordination (single process)

**Production:**

1. User opens **`https://host/`** (or `http://host:8080/`).
2. **Go** serves **`index.html`** and **JS/CSS** from embed.
3. React hydrates; components call **`fetch('/api/v1/…')`** (e.g. **models** endpoints, **`/api/v1/status`** if present) → **same Go server**.
4. Optional **Caddy/nginx** terminates TLS and proxies **everything** to **one** Go port (**REQ-029** **§3.18:** **HTTP/2** **to** **clients** **here** **is** **recommended** **when** **many** **API** **requests** **are** **in** **flight**).

**No** path-based split between Node and Go inside the product.

**Development (non-shipping):** Engineers may use **`next dev`** with **proxy** to Go API or **dual ports** + CORS; **document** in `README` as dev-only.

---

## 6. Raspberry Pi 4 Model B deployment

### 6.1 ARM64 and resources

- **Target:** **linux/arm64** executable.
- **RAM:** **Single Go** process + embedded static assets — significantly lower than **Go + Node**; **2–4 GB** can suffice for light use; **4 GB+** recommended headroom.
- **CPU:** **No SSR** on device; Go serves files and JSON — suitable for Pi **4**.

### 6.2 Process model

- **One** `systemd` service — the **application binary** only.
- **Routine scheduler (REQ-021):** The **~1** **s** **`time.Ticker`** **and** **routine** **tick** **logic** **for** **`random_colour_cycle_all`** **run** **in** **the** **same** **Go** **process** **as** **HTTP** **(§3.16)** — **no** **separate** **daemon** **or** **cron** **for** **MVP**. **Python** **routines** **do** **not** **use** **this** **ticker** **(§3.17)**.
- **Python** **(REQ-022)** **client** **bundle:** **Pyodide** **assets** **ship** **as** **part** **of** **the** **static** **export** **(or** **lazy-loaded** **from** **same** **origin** **chunks**)**; **first** **editor** **visit** **may** **download** **tens** **of** **MB** **—** **document** **in** **README** **for** **Pi** **users** **(prefer** **desktop** **browser** **for** **authoring** **if** **bandwidth** **is** **limited**)**. **Does** **not** **change** **the** **single** **Go** **binary** **artifact** **(REQ-004)**.
- **Optional** separate **`caddy.service`** or **`nginx`** is **OS/infrastructure**, not part of REQ-004’s single binary (the **product** remains one file).

### 6.3 Distribution (**REQ-004** / anti-Docker)

- **Canonical install:** copy **one binary** + optional **unit file**; **not** Docker-first.
- **Docs** (`README`, this file) MUST describe **binary + systemd** path; **do not** require **Dockerfile** or **compose** for production.

### 6.4 Networking

- Go binds **`:8080`** (or configured); reverse proxy maps **80/443** → that socket.
- **REQ-029:** **Enabling** **HTTP/2** **(and** **TLS**)** **on** **the** **proxy** **toward** **browsers** **reduces** **connection** **churn** **for** **many** **parallel** **`fetch`** **calls** **—** **see** **§3.18**. **The** **Go** **listener** **may** **remain** **HTTP/1.1** **behind** **the** **proxy**.

### 6.5 Browser WebGL / three.js on constrained clients (**REQ-003**, **REQ-028**)

- **Rendering** **runs** **in** **the** **user’s** **browser**, **not** **on** **the** **Pi** **CPU** **unless** **the** **user** **opens** **the** **UI** **on** **the** **Pi** **itself** **(Chromium** **on** **Raspberry** **Pi** **OS)**. **REQ-028** **emissive** **spheres** **use** **standard** **three.js** **material** **paths** **(**`MeshStandardMaterial`** **+** **`emissive`**/**`emissiveIntensity`**) **that** **are** **broadly** **supported** **on** **WebGL2**; **avoid** **depending** **on** **optional** **post-processing** **bloom** **for** **baseline** **compliance**.
- **Integrated** **GPUs** **(Pi** **browser,** **older** **laptops):** **Keep** **fragment** **work** **bounded** **—** **≤** **1000** **instanced** **spheres** **+** **tone** **mapping** **as** **in** **§4.7** **is** **the** **expected** **ceiling**; **profile** **if** **adding** **extra** **passes**.

---

## 7. System boundaries (flowchart)

```mermaid
flowchart TB
  subgraph client["Client devices"]
    B[Browser]
  end

  subgraph pi["Raspberry Pi 4 (linux/arm64)"]
    subgraph edge["Optional reverse proxy (not part of product binary)"]
      P[Caddy or nginx]
    end
    subgraph app["Single process: Go binary"]
      G["HTTP: /health, /api/v1/* + static export from embed"]
    end
  end

  B -->|"HTTPS or HTTP"| P
  P -->|"all paths -> one upstream"| G
  B -.->|"Dev: direct to Go :8080"| G
  G --> DB[(SQLite models + lights + scenes + scene_models + routines + routine_runs + python_source column)]
```

---

## 8. Request flows (sequence diagrams)

### 8.1 Initial page load (static + hydration)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary (static + API)

  User->>B: Open /
  B->>P: GET /
  P->>G: GET /
  G-->>P: index.html (from embed)
  P-->>B: HTML
  B->>P: GET /_next/static/...
  P->>G: GET /_next/static/...
  G-->>P: JS/CSS assets
  P-->>B: Assets
  B-->>User: Hydrated responsive UI
```

**REQ-018** **(shell** **theme**)**:** **Before** **first** **paint**, **an** **inline** **script** **in** **`web/app/layout.tsx`** **(or** **equivalent**)** **MUST** **read** **`localStorage`** **`dlm-theme`** **and** **fall** **back** **to** **`prefers-color-scheme`** **when** **the** **key** **is** **absent**/**invalid**, **then** **set** **`html`** **`class`** **per** **§4.11** **so** **the** **document** **does** **not** **flash** **the** **wrong** **shell** **theme**.

### 8.2 Client calls JSON API (same origin)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary

  User->>B: Action triggers fetch
  B->>P: GET /api/v1/status
  P->>G: GET /api/v1/status
  G-->>P: 200 JSON
  P-->>B: 200 JSON
  B-->>User: Updated UI (client state)
```

### 8.3 Create model (multipart CSV upload)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Submit name + CSV file
  B->>P: POST /api/v1/models (multipart)
  P->>G: POST /api/v1/models (multipart)
  G->>G: Parse CSV + validate (wiremodel)
  alt validation failure
    G-->>P: 400 JSON error
    P-->>B: 400 JSON error
    B-->>User: Show actionable message
  else duplicate name
    G-->>P: 409 JSON error
    P-->>B: 409 JSON error
    B-->>User: Show conflict message
  else success
    G->>S: BEGIN; insert model + lights; COMMIT
    S-->>G: OK
    G-->>P: 201 JSON (id, metadata, light_count)
    P-->>B: 201 JSON
    B-->>User: Navigate or refresh list
  end
```

### 8.4 List, view, and delete models

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Open models list
  B->>P: GET /api/v1/models
  P->>G: GET /api/v1/models
  G->>S: Query model summaries
  S-->>G: Rows
  G-->>P: 200 JSON array
  P-->>B: 200 JSON array
  B-->>User: Render responsive list

  User->>B: Select model
  B->>P: GET /api/v1/models/{id}
  P->>G: GET /api/v1/models/{id}
  G->>S: Load model + lights (positions + on color brightness_pct)
  S-->>G: Rowset
  G-->>P: 200 JSON detail
  P-->>B: 200 JSON detail
  B-->>User: Render detail (metadata + three.js 3D of lights + state, client-side)

  User->>B: Confirm delete
  B->>P: DELETE /api/v1/models/{id}
  P->>G: DELETE /api/v1/models/{id}
  alt model referenced by scenes
    G->>S: Check scene_models for model_id
    S-->>G: One or more rows
    G-->>P: 409 JSON model_in_scenes with scene list
    P-->>B: 409 JSON
    B-->>User: Explain model in use; link to scenes
  else success
    G->>S: Delete model (cascade lights)
    S-->>G: OK
    G-->>P: 204 No Content
    P-->>B: 204 No Content
    B-->>User: Update list / redirect
  end
```

### 8.5 Model detail: JSON from Go, WebGL in the browser (**REQ-010**, **REQ-012**, **REQ-019**, **REQ-028**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant R as Client React + three.js

  User->>B: Open model detail
  B->>R: Mount detail route (client)
  R->>P: GET /api/v1/models/{id}
  P->>G: GET /api/v1/models/{id}
  G-->>P: 200 JSON (metadata + lights with on color brightness_pct)
  P-->>R: 200 JSON (metadata + lights with on color brightness_pct)
  R->>R: Build scene, fixed dark-grey scene background and renderer clear per REQ-019 independent of shell theme
  R->>R: On-spheres MeshStandardMaterial base color times brightness plus emissive and emissiveIntensity from brightness_pct off-spheres D0D0D0 15 percent opacity zero emissive segments D0D0D0 15 percent opacity between consecutive ids OrbitControls
  R->>R: Raycast hover or tap shows id and x y z overlay
  R-->>User: WebGL canvas + metadata UI (same origin, no Node SSR)
```

**Boundary:** **Go** returns **positions** and **state** **fields**; **all** **WebGL** allocation and **draw** calls run in the **browser** on the **user device**.

### 8.6 Picking: raycast → id and coordinates (**REQ-010** rule 6)

```mermaid
sequenceDiagram
  actor User as User device
  participant C as Canvas + three.js scene
  participant R as THREE.Raycaster
  participant U as DOM overlay (React)

  User->>C: pointermove desktop or tap touch
  C->>R: setFromCamera normalized device coords
  R->>C: intersect InstancedMesh spheres
  alt hit instanceId i
    C->>U: show lights i id x y z
  else no hit
    C->>U: clear or hide label
  end
```

**Boundary:** **No** round-trip to **Go** for hover; labels use **already-fetched** **`lights`** from **§8.5**.

### 8.7 Update light state (**PATCH**) and refresh 3D (**REQ-011**, **REQ-012**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store
  participant R as Client React + three.js

  User->>B: Change light state (UI or REST client)
  B->>P: PATCH /api/v1/models/{id}/lights/{lightId}/state
  P->>G: PATCH (JSON body subset)
  alt validation failure
    G-->>P: 400 JSON error
    P-->>B: 400 JSON error
    B-->>User: Show error message
  else success
    G->>S: UPDATE lights SET on color brightness_pct
    S-->>G: OK
    G-->>P: 200 JSON full state
    P-->>R: 200 JSON full state
    R->>R: Merge state into lights array and update materials or instances including emissiveIntensity per REQ-028
    R-->>User: Canvas reflects new colour brightness and glow within same frame tick after fetch
  end
```

**Boundary:** **Persistence** and **validation** are **authoritative** in **Go**; the **browser** **must** **reconcile** **three.js** **with** **the** **`200`** **response** **(or** **immediate** **refetch)** **so** **the** **view** **does** **not** **stay** **stale** **after** **a** **successful** **write** **(REQ-012**, **including** **REQ-028** **emissive** **strength** **when** **`brightness_pct`** **or** **`on`** **changes**)**.

### 8.8 Bulk update light state (**REQ-013**, **§3.10**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store
  participant R as Client React + three.js + light table

  User->>B: Select multiple lights and apply on color brightness
  B->>R: User confirms bulk apply
  R->>P: PATCH /api/v1/models/{id}/lights/state/batch (ids + patch fields)
  P->>G: PATCH JSON body
  alt validation failure
    G-->>P: 400 JSON error
    P-->>R: 400 JSON error
    R-->>User: Show actionable message
  else success
    G->>S: BEGIN; UPDATE lights for each id; COMMIT
    S-->>G: OK
    G-->>P: 200 JSON states array
    P-->>R: 200 JSON states array
    R->>R: Merge states into lights; refresh table and three.js meshes
    R-->>User: List and canvas match persisted state
  end
```

**Boundary:** **One** **transaction** in **SQLite** for **all** **ids** in the **batch**; the **UI** **must** **treat** **`200`** **`states`** as **authoritative** for **both** **§4.8** and **§4.7**.

### 8.9 Reset all lights to defaults (**REQ-014**, **§3.11**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store
  participant R as Client React + three.js + light table

  User->>B: Click Reset lights
  B->>R: Invoke reset handler
  R->>P: POST /api/v1/models/{id}/lights/state/reset (no body)
  P->>G: POST
  alt model missing
    G-->>P: 404 JSON error
    P-->>R: 404
    R-->>User: Show error message
  else success
    G->>S: BEGIN; UPDATE all lights to off, white, 100 percent brightness; COMMIT
    S-->>G: OK
    G-->>P: 200 JSON states array for all lights
    P-->>R: 200 JSON states array
    R->>R: Merge states into lights; refresh table and three.js meshes
    R-->>User: Model matches default visual and list state
  end
```

**Boundary:** **Same** **timeliness** expectation as **§8.7** — **no** **indefinite** **staleness** after **200**.

### 8.10 Create scene with models (**REQ-015**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Submit scene name and ordered model list (no offsets)
  B->>P: POST /api/v1/scenes (JSON: model_id list only)
  P->>G: POST /api/v1/scenes
  G->>G: Compute offsets per §3.12 create-time algorithm; validate containment
  alt validation failure
    G-->>P: 400 JSON error
    P-->>B: 400
    B-->>User: Show actionable message
  else success
    G->>S: BEGIN; insert scenes + scene_models; COMMIT
    S-->>G: OK
    G-->>P: 201 JSON scene id
    P-->>B: 201
    B-->>User: Navigate to scene detail
  end
```

### 8.11 Load scene detail and render composite WebGL (**REQ-015**, **REQ-019**, **REQ-012**, **REQ-028**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store
  participant R as Client React + three.js

  User->>B: Open scene detail
  R->>P: GET /api/v1/scenes/{id}
  P->>G: GET /api/v1/scenes/{id}
  G->>S: Load scene, placements, models, lights, state
  S-->>G: Rowsets
  G-->>P: 200 JSON items with scene-space positions
  P-->>R: 200 JSON
  R->>R: Build composite scene: per-model spheres with §4.7 REQ-028 emissive on-spheres and chain segments only within each model
  R->>R: Apply same REQ-019 fixed dark-grey viewport as model detail §4.7
  R-->>User: WebGL canvas + add remove placement controls
```

### 8.12 Remove last model from scene (**REQ-015**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Remove model from scene
  B->>P: DELETE /api/v1/scenes/{sid}/models/{mid}
  P->>G: DELETE
  alt last model in scene
    G-->>P: 409 scene_last_model
    P-->>B: 409 JSON
    B-->>User: Modal explains whole scene will be deleted
    User->>B: Confirm
    B->>P: DELETE /api/v1/scenes/{sid}
    P->>G: DELETE scene
    G->>S: CASCADE delete scene_models and scene row
    S-->>G: OK
    G-->>P: 204
    P-->>B: 204
    B-->>User: Redirect to scene list
  else more than one model
    G->>S: DELETE scene_models row
    S-->>G: OK
    G-->>P: 204
    P-->>B: 204
    B-->>User: Refresh scene detail
  end
```

### 8.13 Factory reset all data (**REQ-017**, **§3.14**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Open Options then Factory reset
  B-->>User: Show blocking warning dialog
  User->>B: Confirm erase
  B->>P: POST /api/v1/system/factory-reset
  P->>G: POST
  alt success
    G->>S: BEGIN; DELETE routine_runs, routines, scene_models, scenes, lights, models; SeedDefaultSamples; COMMIT
    S-->>G: OK
    G-->>P: 200 { ok: true }
    P-->>B: 200
    B->>B: Clear client caches; navigate to /models; show success banner
    B-->>User: Model list shows three samples only
  else failure
    G-->>P: 500 JSON error
    P-->>B: 500
    B-->>User: Show error; data unchanged if transaction rolled back
  end
```

### 8.14 Reset camera (client only) (**REQ-016**)

```mermaid
sequenceDiagram
  actor User as User device
  participant R as Client React + three.js + OrbitControls

  User->>R: Click Reset camera on model or scene view
  R->>R: applyDefaultFraming from current bounds; controls.update()
  R-->>User: View returns to default framing; no HTTP request
```

### 8.15 Scene region query and bulk update in scene coordinates (**REQ-020**)

```mermaid
sequenceDiagram
  actor User as Integrator or UI user
  participant B as Browser or API client
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant S as SQLite store

  User->>B: Query cuboid or sphere region in a scene
  B->>P: POST /api/v1/scenes/{id}/lights/query/{shape}
  P->>G: POST
  G->>G: Validate geometry payload (finite numbers, positive dimensions/radius)
  alt invalid geometry
    G-->>P: 400 validation_failed
    P-->>B: 400 JSON error with actionable details
  else valid geometry
    G->>S: Load scene placements + light rows
    S-->>G: Rows
    G->>G: Compute sx,sy,sz and filter inclusion (inclusive boundaries)
    G-->>P: 200 matched lights in scene coordinates
    P-->>B: 200 JSON
  end

  User->>B: Bulk update state for same region
  B->>P: PATCH /api/v1/scenes/{id}/lights/state/{shape}
  P->>G: PATCH
  G->>G: Validate geometry + REQ-011 state fields
  G->>S: BEGIN; resolve matches by sx/sy/sz; update lights; COMMIT
  alt store or validation failure during transaction
    G->>S: ROLLBACK
    G-->>P: 4xx or 5xx JSON error
    P-->>B: Error response; no partial writes
  else success
    S-->>G: OK
    G-->>P: 200 updated_count + updated states
    P-->>B: 200 JSON
  end
```

### 8.16 Scene routine start, server tick, and stop (**REQ-021**, **§3.16**)

```mermaid
sequenceDiagram
  actor User as User device
  participant B as Browser
  participant P as Reverse proxy (optional)
  participant G as Go binary
  participant T as Routine scheduler (in-process ticker)
  participant S as SQLite store

  User->>B: Create routine definition
  B->>P: POST /api/v1/routines
  P->>G: POST
  G->>S: INSERT routines row
  S-->>G: OK
  G-->>P: 201 routine json
  P-->>B: 201

  User->>B: Start routine on scene
  B->>P: POST /api/v1/scenes/{sid}/routines/{rid}/start
  P->>G: POST
  G->>S: BEGIN; INSERT routine_runs running; COMMIT
  G->>S: Shared store: PATCH scene lights/state/scene then state/batch (initial colours)
  S-->>G: OK
  G-->>P: 201 run json
  P-->>B: 201

  loop Each 1s tick while running
    T->>S: SELECT running routine_runs
    S-->>T: rows
    T->>S: Shared store: scene lights/state/batch per random_colour_cycle_all
    S-->>T: OK
  end

  User->>B: Stop routine
  B->>P: POST /api/v1/scenes/{sid}/routines/runs/{runId}/stop
  P->>G: POST
  G->>S: UPDATE routine_runs set stopped
  S-->>G: OK
  G-->>P: 200
  P-->>B: 200
  T->>S: Next tick finds no running row for scene; no further updates
```

### 8.17 Python routine: start, Pyodide worker loop, scene API fetch, cooperative and forced stop (**REQ-022**, **REQ-030**, **§3.17**)

**REQ-030** **note:** **`scene.random_hex_colour()`** **runs** **only** **inside** **the** **worker** **(Python** **`random`)** **and** **does** **not** **add** **HTTP** **traffic** **to** **§3.15**; **it** **does** **not** **appear** **as** **a** **`fetch`** **leg** **in** **the** **diagram** **below**.

```mermaid
sequenceDiagram
  actor User as User device
  participant Page as Next.js page (§4.13)
  participant W as Pyodide Worker
  participant B as Browser fetch
  participant P as Reverse proxy (optional)
  participant G as Go binary (§3.15)
  participant S as SQLite store

  User->>Page: Save Python routine
  Page->>P: PATCH /api/v1/routines/{id} (python_source, …)
  P->>G: PATCH
  G->>S: UPDATE routines
  S-->>G: OK
  G-->>Page: 200

  User->>Page: Start on scene
  Page->>P: POST /api/v1/scenes/{sid}/routines/{rid}/start
  P->>G: POST
  G->>S: INSERT routine_runs running (§3.16)
  S-->>G: OK
  G-->>Page: 201 run_id

  Page->>W: new Worker; init Pyodide; postMessage(sceneId, runId, source)
  loop Each iteration until stopped
    W->>W: exec user script (scene bound)
    W->>B: fetch PATCH/GET §3.15 (via js.fetch)
    B->>P: same-origin API
    P->>G: HTTP
    G->>S: transactional light state updates
    S-->>G: OK
    G-->>W: JSON
    W->>W: asyncio.sleep(iteration_gap)
    W->>B: optional GET …/routines/runs to see if stopped
  end

  User->>Page: Stop
  Page->>P: POST …/routines/runs/{runId}/stop
  P->>G: POST
  G->>S: UPDATE routine_runs stopped
  G-->>Page: 200
  Page->>W: postMessage cancel or rely on poll
  alt worker exits cooperatively within T_force
    W-->>Page: done
  else still running after T_force
    Page->>W: worker.terminate()
  end
```

---

### 8.18 Python routine page: unified scene target, live viewport sync, and resets (**REQ-027**, **REQ-028**, **§4.13**)

```mermaid
sequenceDiagram
  actor User as User device
  participant Page as Next.js §4.13 page
  participant Canvas as SceneLightsCanvas
  participant W as Pyodide Worker
  participant G as Go binary §3.15

  User->>Page: Select target scene (run + viewport)
  Page->>G: GET /api/v1/scenes/{id}
  G-->>Page: items + lights + state
  Page->>Canvas: mount / update props

  User->>Page: Start routine on same scene
  Page->>G: POST …/routines/…/start
  G-->>Page: 201 run_id
  Page->>W: worker loop (scene shim PATCH §3.15)

  loop Each successful worker PATCH for that scene
    W->>G: PATCH …/lights/state/…
    G-->>W: 200 + states
    W-->>Page: postMessage applied
    Page->>G: GET /api/v1/scenes/{id}/lights (or GET scene detail)
    G-->>Page: fresh lights + state
    Page->>Canvas: merge state (REQ-012 and REQ-028 timeliness)
  end

  User->>Page: Reset scene lights
  Page->>G: PATCH …/lights/state/scene (off, #ffffff, 100%)
  G-->>Page: 200
  Page->>Canvas: merge state

  User->>Page: Reset camera
  Page->>Canvas: applyDefaultFraming (REQ-016, no API)
```

### 8.19 Optional push path for high-throughput observers (**REQ-029**, **§3.18**)

**Note:** **This** **diagram** **describes** **an** **optional** **future** **or** **parallel** **implementation** **path** **when** **bounded** **polling** **(**§4.7**, **§4.9**)** **is** **insufficient** **for** **multi-tab** **freshness** **under** **sustained** **external** **writes**.

```mermaid
sequenceDiagram
  participant Ext as External client
  participant G as Go binary
  participant S as SQLite store
  participant B as Browser observer tab

  Ext->>G: PATCH bulk model or scene state
  G->>S: COMMIT light rows
  S-->>G: OK
  G-->>Ext: 200 JSON
  Note over G,B: Optional SSE after successful commit per §3.18
  G-->>B: text/event-stream data line
  B->>G: GET authoritative state
  G->>S: read
  S-->>G: rows
  G-->>B: 200 JSON
  B->>B: merge into three.js and tables REQ-012
```

### 8.20 Light state write with no-op elision (**REQ-031**, **§3.19**)

```mermaid
sequenceDiagram
  actor User as User device
  participant R as Client React + three.js
  participant G as Go binary
  participant S as SQLite store

  User->>R: Edit light to same values as shown
  R->>R: Compare merged triple to last rendered (§3.19)
  alt client skip HTTP
    R-->>User: No fetch; canvas unchanged
  else client sends PATCH
    R->>G: PATCH .../lights/{id}/state
    G->>S: BEGIN; read row; merge; compare
    alt merged equivalent to stored
      S-->>G: no UPDATE
      G-->>R: 200 full state optional unchanged true
      R->>R: Skip three.js rebuild if triple matches cache
    else state changed
      G->>S: UPDATE lights COMMIT
      S-->>G: OK
      G-->>R: 200 full state
      R->>R: Update materials instances REQ-012 REQ-028
    end
    R-->>User: Timely correct appearance
  end
```

**Note:** **Integrators** **that** **bypass** **the** **shipped** **UI** **still** **benefit** **from** **server-side** **§3.19** **skips** **;** **the** **client** **branch** **in** **the** **diagram** **is** **optional** **but** **recommended** **for** **fewer** **round-trips**.


## 9. Security notes (baseline)

- Prefer **same-origin** in production to minimize **CORS** surface.
- **Secrets** via env only; **no** secrets baked into client bundles beyond **public** constants.
- **No** mandatory **container** trust boundary from the product’s perspective (REQ-004).
- **Upload limits:** Enforce a **maximum request body size** on **`POST /api/v1/models`** (e.g. **`http.MaxBytesReader`** or server limit) large enough for **1000** CSV rows but small enough to bound **memory** and **DoS** risk on the Pi.
- **Scene batch updates:** **`PATCH /api/v1/scenes/{id}/lights/state/batch`** **can** **carry** **one** **JSON** **object** **per** **light** **(up** **to** **~1000** **per** **scene** **in** **worst** **case**)**; **enforce** **a** **reasonable** **max** **body** **size** **(e.g.** **several** **MB** **below** **Pi** **RAM** **headroom**)** **and** **reject** **oversize** **with** **`413`** **or** **`400`** **before** **full** **parse** **if** **needed**.
- **SQLite file:** Treat **`DLM_DB_PATH`** as **persistent** storage on the Pi (e.g. SD card or USB); operators should **back up** the DB file with normal file backup practices.
- **Factory reset:** **`POST /api/v1/system/factory-reset`** **is** **destructive** **and** **unauthenticated** **in** **MVP**; **treat** **network** **exposure** **accordingly** **(see** **§3.14**)**.**
- **User-authored Python (REQ-022):** **Scripts** **run** **only** **inside** **the** **browser** **worker** **and** **may** **only** **invoke** **same-origin** **`fetch`** **to** **the** **documented** **§3.15** **routes** **via** **the** **`scene`** **shim** **—** **do** **not** **expose** **admin** **tokens** **or** **third-party** **API** **keys** **into** **the** **worker**. **Malicious** **scripts** **can** **still** **spam** **the** **Go** **API** **(rate** **limits** **and** **body** **size** **limits** **§9** **apply**)**.

---

## 10. Traceability to requirements

| REQ | Addressed in sections |
|-----|------------------------|
| REQ-001 | §1–§5, §7–§8 |
| REQ-002 | §1, §4, §5, §8 |
| REQ-003 | §1, §3.4, §6, §7 |
| REQ-004 | §1 (resolution), §3.3–§3.5, §4.1–§4.2, §5–§6 |
| REQ-005 | §1, §3.1, §3.6 |
| REQ-006 | §1, §3.1–§3.3, §3.2, §3.12–§3.13, §4.3, §4.6, §7–§8 (incl. **§8.4** **409** **path**) |
| REQ-007 | §1, §3.3, §3.6, §8.3 |
| REQ-008 | §1, §2, §3.1, §3.7, §3.5 (release:sync contract) |
| REQ-009 | §1, §2, §3.1, §3.3, §3.8 |
| REQ-010 | §1, §4.6, §4.7, §8.4–§8.6 |
| REQ-011 | §1, §3.2, §3.3, §3.9, §3.10 (validation semantics shared with batch), §4.6, §8.7 |
| REQ-012 | §1, §3.9, §4.6, §4.7 (with REQ-028 emissive on-spheres), §8.5, §8.7 |
| REQ-013 | §1, §3.2, §3.10, §4.6, §4.8, §4.7 (state sync), §8.8 |
| REQ-014 | §1, §3.2, §3.3, §3.8, §3.9 (defaults), §3.11, §4.6, §4.7 (state sync), §8.9 |
| REQ-015 | §1, §3.1, §3.2, §3.12–§3.13, §4.6, §4.9 (composite three.js + REQ-028 via §4.7), §7 (DB), §8.10–§8.12 |
| REQ-016 | §1, §4.7, §4.9, §8.14 |
| REQ-017 | §1, §3.2, §3.14, §4.6, §4.10, §8.13, §9 |
| REQ-018 | §1, §2, §4.5, §4.6, §4.7, §4.9, §4.10, §4.11, §4.12, §4.13, §8.1 (inline theme + hydration), §10 |
| REQ-019 | §1, §4.7, §4.9, §4.11 (contrast note), §8.5, §8.11, §10 |
| REQ-020 | §1, §3.2, §3.12, §3.13, §3.15, §8.15 |
| REQ-021 | §1, §3.2, §3.14, §3.15, §3.16, §4.9, §4.11, §4.12, §7, §8.16 |
| REQ-022 | §1, §3.2, §3.15, §3.16, §3.17, §4.9, §4.11, §4.12, §4.13, §6.2, §7, §8.17 |
| REQ-023 | §1, §3.2 (`POST /routines` note), §3.16 (UI support), §4.9, §4.11 (nav: no Python-only item), §4.12, §4.13 |
| REQ-024 | §4.13 (`#python-scene-api-catalog` below editor; picker + commented samples + insert; manifest sync with §3.17 incl. REQ-030) |
| REQ-025 | §4.13 (`PYTHON_ROUTINE_DEFAULT_SOURCE`, sphere colour demo; SHOULD use `scene.random_hex_colour()` per REQ-030) |
| REQ-026 | §3.15 (axis mapping for `size`), §3.17 (`scene.width` / `height` / `depth`), §4.13 |
| REQ-027 | §3.15 (`PATCH …/lights/state/scene` for reset), §4.7, §4.9, §4.13, §8.18 (with REQ-028 materials via shared canvas) |
| REQ-028 | §1, §4.7, §4.9, §4.13, §6.5, §8.5, §8.7 |
| REQ-029 | §1, §3.2, §3.10, §3.15, §3.18, §3.19 (observer alignment), §4.3, §4.7, §4.9, §6, §7, §8.19, §9 |
| REQ-030 | §1, §3.17, §4.13, §8.17 (local helper note; no HTTP in sequence) |
| REQ-031 | §3.9, §3.10 (batch note), §3.15 (bulk semantics), §3.16 (scheduler/store), §3.18, §3.19, §4.7, §4.9, §4.13, §8.7, §8.8, §8.20 |

---

**Next step:** After **material** **code** **or** **requirement** **changes**, invoke **`@implementor`** **then** **`@verifier`** **to** **keep** **tests**, **`docs/traceability_matrix.md`**, **and** **this** **document** **aligned**. **REQ-022** **editor** **surface** **is** **implemented** **with** **`EditorView`**, **`codemirror`** **`basicSetup`**, **and** **`@codemirror/*`** **per** **§3.17** **and** **§4.13** **above**, **with** **novice** **instructional** **copy** **on** **that** **page**. **REQ-023** **is** **satisfied** **by** **the** **§4.12** **create** **flow** **(single** **`type`** **dropdown** **including** **Python**)** **without** **a** **redundant** **primary** **Python-only** **create** **button**. **REQ-024** **adds** **the** **interactive** **API** **reference** **below** **the** **editor** **(**picker**, **commented** **samples**, **insert** **at** **caret**/**EOF**)** **in** **§4.13**. **REQ-027** **unifies** **run** **controls** **and** **`SceneLightsCanvas`** **under** **one** **scene** **selector** **(§4.13**, **§8.18**)**. **REQ-028** **is** **specified** **in** **§4.7**, **§4.9**, **§4.13**, **§6.5**, **and** **§8.5**/**§8.7** **above**. **REQ-029** **is** **specified** **in** **§3.18**, **with** **cross-references** **in** **§4.3**, **§4.7**, **§4.9**, **and** **§8.19**. **REQ-030** **adds** **`scene.random_hex_colour()`** **(**§3.17**, **§4.13** **manifest** **+** **template** **preference**, **worker** **implementation**)**. **REQ-031** **adds** **equivalence-based** **elision** **for** **persistence** **and** **three.js** **(**§3.19**)** **plus** **a** **documented** **hook** **for** **future** **WLED-class** **physical** **sync** **(**no** **device** **code** **in** **this** **requirement**)**.
