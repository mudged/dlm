# Requirements

This document captures the initial product requirements. IDs are immutable.

---

### REQ-001 — Go service and Next.js + Tailwind web UI

| Field | Value |
|-------|-------|
| **ID** | REQ-001 |
| **Title** | Go service and Next.js + Tailwind web UI |
| **Priority** | Must |
| **Actor(s)** | End user; operator / maintainer |

**User story**

As a user, I want a web application backed by a **Golang** service with a **Next.js** and **Tailwind CSS** front end, so that I have a maintainable full-stack product with a modern UI toolchain.

**Scope**

- In scope: HTTP-capable Go backend (API and/or static coordination as later architecture defines); Next.js app with Tailwind for styling; clear separation between backend and UI code in the repository layout (exact structure deferred to architecture).
- Out of scope: Choice of specific Go web framework, database, auth provider, and hosting topology beyond the Pi target in REQ-003.

**Business rules**

1. The product MUST include deployable Go server code and a Next.js application using Tailwind.
2. The UI MUST consume or coordinate with the backend in a way defined in `docs/architecture.md` (deferred until architect pass).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Not applicable at requirements-only level for this REQ; see REQ-002.
- Tablet: Not applicable at requirements-only level for this REQ; see REQ-002.
- Desktop: Not applicable at requirements-only level for this REQ; see REQ-002.

**Dependencies**

- None

**Open questions**

- None

---

### REQ-002 — Responsive, reactive UI across device classes

| Field | Value |
|-------|-------|
| **ID** | REQ-002 |
| **Title** | Responsive, reactive UI across device classes |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, I want the interface to **respond to my actions** and **adapt to mobile, tablet, and desktop** viewports, so that the product is usable on phones, tablets, and PCs.

**Scope**

- In scope: Responsive layout and typography; interactive client behavior (loading, navigation, forms, or other patterns as designed); touch-friendly patterns on small screens where controls exist.
- Out of scope: Native mobile apps; offline-first PWA requirements unless added in a later requirement.

**Business rules**

1. Primary user-facing flows MUST remain **usable** at typical mobile, tablet, and desktop widths without relying on desktop-only interaction patterns where alternatives exist.
2. The UI MUST be **reactive** (client-side interactivity appropriate for Next.js), not a static HTML-only brochure unless a later requirement explicitly narrows scope.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Layout stacks appropriately; readable text; interactive controls reachable; minimize horizontal scrolling for primary content.
- Tablet: Balanced use of space; orientation changes handled gracefully where applicable.
- Desktop: Efficient use of width; keyboard and pointer interactions supported.

**Dependencies**

- REQ-001

**Open questions**

- Minimum supported browsers and OS versions (if any).

---

### REQ-003 — Target deployment on Raspberry Pi 4 Model B

| Field | Value |
|-------|-------|
| **ID** | REQ-003 |
| **Title** | Target deployment on Raspberry Pi 4 Model B |
| **Priority** | Must |
| **Actor(s)** | Operator / maintainer |

**User story**

As an operator, I want the system **designed and documented for deployment** on a **Raspberry Pi 4 Model B**, so that I can run the stack on that hardware reliably within its constraints.

**Scope**

- In scope: Architecture and deployment documentation addressing Pi 4 Model B (including **ARM64** / resource awareness), process model for Go and Next.js artifacts on the device, and operational assumptions needed to install and run the stack.
- Out of scope: Providing production Pi images or CI hardware; exact OS distribution unless captured in architecture.

**Business rules**

1. `docs/architecture.md` MUST describe how the Go service and Next.js build fit the Pi 4 B deployment context after the architect pass.
2. Documentation MUST acknowledge **ARM64** and typical Pi **CPU/RAM** limits when stating runtime and build choices.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-001

**Open questions**

- Preferred OS on the Pi (e.g. Raspberry Pi OS 64-bit) and whether a reverse proxy is mandatory.
- REQ-004 constrains packaging to a **single executable**; `docs/architecture.md` MUST be updated to reconcile Pi deployment (this REQ) with that constraint.

---

### REQ-004 — Single executable packaging (no Docker at this stage)

| Field | Value |
|-------|-------|
| **ID** | REQ-004 |
| **Title** | Single executable packaging (no Docker at this stage) |
| **Priority** | Must |
| **Actor(s)** | Operator / maintainer |

**User story**

As an operator, I want the product **packaged as a single executable file** per target platform, so that deployment on the Pi (and elsewhere) means copying and running **one binary** without orchestrating multiple product processes or container runtimes from the distribution.

**Scope**

- In scope: Defining that the **primary deliverable** for a given OS/CPU (e.g. **linux/arm64** for Pi) is **one executable file** that fulfills the product’s runtime obligations (HTTP API and serving the UI per REQ-001), subject to how `docs/architecture.md` realizes that shape **within a single process/binary**.
- In scope: Explicitly deferring **Docker / OCI images**, **Dockerfile**-mandated workflows, and **docker-compose** (or equivalent) as **required** distribution or runtime packaging **at this stage**.
- Out of scope: Operators voluntarily using containers locally; third-party tools not shipped as part of the product.
- Out of scope: **Microsoft Windows** and **macOS** executable variants unless added by a later requirement (initial focus remains Pi / **linux/arm64** unless extended).

**Business rules**

1. For each supported **release target** documented in `docs/architecture.md`, the product MUST ship (or document a reproducible build of) **exactly one runnable executable file** as the application binary—**not** a bundle that **requires** a separate Node.js runtime binary, **nor** a **second** application-specific daemon shipped alongside the main program for routine operation.
2. Supporting assets (if any) MUST either be **embedded in that executable** or **generated at runtime** by it; **shell scripts** or loose file trees that must accompany the binary for **normal** startup are **not** acceptable as the primary packaging story **except** OS-level service wrappers (e.g. **systemd** unit files) that only invoke the single binary.
3. **Docker** images, **container** as the **canonical** install path, and **compose** files as **mandatory** deployment are **out of scope** at this stage; documentation MUST NOT present containers as the **only** or **required** way to run the product.
4. `docs/architecture.md` MUST be updated after this requirement is accepted so that the **single-process / single-binary** model, Pi constraints (REQ-003), and Next.js + Tailwind behavior (REQ-001/002) are **consistent**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A at requirements level (REQ-002 still governs rendered UI).
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-001, REQ-003

**Open questions**

- Whether **full Next.js SSR** must be preserved inside one binary or whether a **static export + client hydration** pattern controlled by architecture is acceptable if UX acceptance (REQ-002) is still met.

---

### REQ-005 — Wire light model (data shape, CSV, metadata)

| Field | Value |
|-------|-------|
| **ID** | REQ-005 |
| **Title** | Wire light model (data shape, CSV, metadata) |
| **Priority** | Must |
| **Actor(s)** | End user; operator |

**User story**

As a user, I want a **model** to represent a configuration of lights on a wire with clear structure and a standard file format, so that I can author, exchange, and reason about layouts consistently.

**Scope**

- In scope: A model contains **up to 1000** lights; each light has a **sequential index** starting at **0** and a position in **3D space** (**x**, **y**, **z** coordinates). Lights form a **single ordered chain** along **ascending id**: each light is logically connected **only** to its **predecessor** and **successor** on that chain (see business rules). A model is representable as a **CSV** file with columns **id**, **x**, **y**, **z** (where **id** is the light index). Each model has **metadata** including a **name** and **creation date**.
- Out of scope: **Branching** or **mesh** topologies (extra adjacencies beyond the chain); physical hardware; animation; colour or brightness in the CSV interchange; import formats other than the defined CSV unless added later.

**Business rules**

1. A model MUST NOT describe more than **1000** lights.
2. Light indices MUST be **non-negative integers** forming a **contiguous sequence** starting at **0** (i.e. for *n* lights, indices **0** through **n − 1** with **no gaps** and **no duplicates**).
3. Coordinates **x**, **y**, and **z** MUST be **real numbers** (finite; architecture may fix representation and precision).
4. The interchange CSV MUST use the field names **id**, **x**, **y**, **z** (column order and delimiter as specified in acceptance criteria).
5. Model metadata MUST include **name** and **creation date** (storage format and timezone policy deferred to architecture).
6. **Adjacency along the wire** is defined **only** by consecutive **id** values: for **n** lights (**n ≥ 0**), light **i** (**0 ≤ i < n**) has **at most two** logical neighbors—**i − 1** when **i > 0** and **i + 1** when **i < n − 1**. When **n > 1**, light **0** has **exactly one** neighbor (**1**); light **n − 1** has **exactly one** neighbor (**n − 2**); every light with **0 < i < n − 1** has **exactly two** neighbors. There MUST be **no** defined adjacency between **non-consecutive** ids (e.g. **i** and **i + 2**). When **n ≤ 1**, there are **no** pairwise adjacencies.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A for this data-definition requirement; see REQ-006.
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-001

**Open questions**

- Whether model **names** must be **unique** in the system.
- Preferred **timezone** and precision for **creation date**.

---

### REQ-006 — List, view, delete, and create models via CSV upload

| Field | Value |
|-------|-------|
| **ID** | REQ-006 |
| **Title** | List, view, delete, and create models via CSV upload |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, I want to **list** all models, **open** one to inspect it, **delete** models I no longer need, and **add** new models by **uploading** a CSV file, so that I can manage my light configurations end-to-end from the application.

**Scope**

- In scope: User-facing flows (and backing behavior per architecture) to **list** all stored models, **view** details of a selected model (including its lights and metadata as appropriate), **delete** an existing model, and **create** a new model by **uploading** a CSV in the format defined under REQ-005.
- Out of scope: Bulk export, editing lights in-place after import, versioning, and sharing links unless added later.

**Business rules**

1. The application MUST expose a way to see **all** models the user may access (subject to any future auth).
2. The application MUST allow **viewing** a single model’s content (lights and metadata) after selection.
3. The application MUST allow **deleting** a model such that it no longer appears in listings and is removed per persistence rules in architecture.
4. The application MUST allow **creating** a model by **uploading** a CSV file; successful creation MUST persist the model and its metadata (name and creation date supplied or derived per architecture—e.g. user-provided name at upload time).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: List and detail usable without horizontal scrolling for primary content; upload control reachable and labeled clearly.
- Tablet: List and detail use space efficiently; upload remains obvious in the create flow.
- Desktop: Efficient browsing of lists and details; keyboard-accessible actions where controls exist.

**Dependencies**

- REQ-001, REQ-002, REQ-005

**Open questions**

- Whether **name** is entered in the UI at upload time or derived from the **filename**.

---

### REQ-007 — Validate CSV on model upload

| Field | Value |
|-------|-------|
| **ID** | REQ-007 |
| **Title** | Validate CSV on model upload |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, when I upload a CSV to create a model, I want the system to **reject invalid files** with clear feedback, so that only **well-formed** data with **correct** ids and values enters the system.

**Scope**

- In scope: **Validation** of uploaded CSV for **correct structure** (expected columns, parseable rows) and **correct values**: **sequential** **id** values per REQ-005, valid numeric **x**, **y**, **z**, and respect of the **1000**-light maximum.
- Out of scope: Linting coordinate ranges (e.g. unit cubes); duplicate-row detection beyond id rules unless specified in scenarios.

**Business rules**

1. On upload, the system MUST verify the file is **CSV** with the required columns **id**, **x**, **y**, **z** (and acceptable header/row rules in acceptance criteria).
2. The system MUST reject uploads where **id** values are not **integers** forming a **contiguous sequence** from **0** through **n − 1** for the row count **n**, or where **n** exceeds **1000**.
3. The system MUST reject uploads where **x**, **y**, or **z** cannot be interpreted as **finite** real numbers.
4. On rejection, the user MUST receive **actionable** feedback (e.g. error indicating row/column or rule violated); no partial model MUST be persisted for that failed upload.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Error messages readable without truncation of the primary explanation.
- Tablet: Same as mobile.
- Desktop: Errors visible near the upload action or in a consistent notification region.

**Dependencies**

- REQ-005, REQ-006

**Open questions**

- Whether **empty** models (**n = 0**) are allowed.

---

### REQ-008 — Single command to build and run locally

| Field | Value |
|-------|-------|
| **ID** | REQ-008 |
| **Title** | Single command to build and run locally |
| **Priority** | Must |
| **Actor(s)** | Developer; operator |

**User story**

As a developer, I want **one command** from the repository (or a **single script** invoked from the repo root) that **builds** the UI for embedding and **starts** the application, so that I can run the full stack without remembering multiple steps.

**Scope**

- In scope: A **documented** entry point (e.g. shell script under `scripts/`, or a **Makefile** target, or equivalent agreed in architecture) runnable from a **normal clone** that performs: produce the **static export** needed for embed, then **launch** the **Go** server so the app is reachable (default listen per configuration, e.g. `:8080`). The command MUST be suitable for **local** use on Linux/macOS/WSL; Windows MAY use Git Bash/WSL or a documented alternative.
- Out of scope: **Production** Pi install scripts as the only path; **Docker** as the canonical runner (REQ-004 still applies); CI matrix beyond “this command is documented and works in the devcontainer or primary dev OS.”

**Business rules**

1. **README.md** MUST document the **exact** command(s) to invoke (including script path or target name).
2. The workflow MUST **not** require a **second manual step** after the single command for a **standard** “build UI + run server” session (one invocation completes both).
3. **AGENTS.md** (or this requirements set referenced there) MUST acknowledge that **REQ-008** exists so agents keep **README** and the script in sync when the workflow changes.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A (developer tooling).
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-001, REQ-004

**Open questions**

- Whether the same command also runs **`go test`** or **lint** (optional flags vs separate targets).

---

### REQ-009 — Default sample models (sphere, cube, cone)

| Field | Value |
|-------|-------|
| **ID** | REQ-009 |
| **Title** | Default sample models (sphere, cube, cone) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, I want **three sample light models** available **by default** that illustrate **geometric shapes** with lights **spread across each shape’s faces**—not just **tracing edges**—so that I can see **even coverage** of **planes** and **surfaces** before uploading my own CSV.

**Scope**

- In scope: On a **fresh** application data set (e.g. empty database or first startup), the system MUST expose **exactly three** predefined models, distinct in name and geometry, representing: a **sphere**, a **cube**, and a **cone**. Coordinates use **SI meters**. For **each** sample, **all** lights MUST lie on the **outside surface** of the nominal solid (the boundary of the intended **sphere**, **cube**, or **cone**), MUST NOT lie **inside** that solid, and MUST NOT deviate from that surface by more than **0.03 m** in the sense of **closest distance** from the light’s position to the **nominal surface** (lights sit **on** the surface or within a **3 cm** tolerance band **outside** it—not floating arbitrarily far above, and not interior). **Consecutive** lights along the model’s ordered path (**id** order **0, 1, …**) MUST have pairwise Euclidean spacing **at least 0.05 m** (**5 cm**) and **at most 0.10 m** (**10 cm**). Each sample MUST contain **at least 500** and **at most 1000** lights (**REQ-005** still caps any model at **1000**). Each shape MUST have an overall characteristic size of **about 2 m** (sphere **diameter** ~2 m; cube **edge length** ~2 m; **right circular cone** with **height** ~2 m and base diameter consistent with the geometry—exact proportions deferred to architecture within these bounds).
- **Face and surface coverage (not edge-only):** The **cube** sample MUST place lights on the **six** nominal **square face planes** (the flat **exterior** faces), with lights **distributed across the interior of those faces**—the design MUST **not** confine lights to **edges** or **vertices** only (a **wireframe** or **edge-tracing** layout is **not** sufficient). Lights on the **cube** MUST be **evenly distributed** over the **combined** area of those faces (architecture defines how **evenness** is achieved and verified, e.g. per-face quotas, low clustering, or area-proportional density). The **sphere** has **no** flat **planes**; lights MUST lie on the **spherical** surface and MUST be **evenly distributed** over **surface area** as a whole, **not** concentrated on a **single** narrow **curve** or **strip** as if **tracing** a one-dimensional path only. The **cone** sample MUST cover the **lateral** (curved) surface and the **flat** **base** **plane**; lights MUST be **evenly distributed** over the **lateral** area and **evenly** over the **base disc**—**not** only along the **base rim** or **apex**–**rim** **creases** as the sole locus of lights.
- Out of scope: User editing of sample models; optional **toggle to hide** samples; lighting beyond positions; user-uploaded CSV models (rules above apply to **default samples** only unless a later requirement extends them).

**Business rules**

1. The three samples MUST appear in the **same** **list** and **detail** flows as user-created models (**REQ-006**).
2. Each sample MUST have a light count **n** with **500 ≤ n ≤ 1000** and MUST satisfy **REQ-005** (indices **0 … n − 1**, **≤ 1000**).
3. For every sample, consecutive lights **(id i, id i+1)** MUST satisfy **0.05 m ≤** Euclidean distance **≤ 0.10 m** for **i = 0 … n − 2**.
4. Every sample light position MUST be **on or outside** the nominal solid’s surface (not in the interior) and MUST have **closest distance** to the **nominal surface** **≤ 0.03 m** (architecture defines the nominal surfaces and tests).
5. Sample model **names** MUST make the shape identifiable (e.g. containing “sphere”, “cube”, “cone” or equivalent localized labels—English product default acceptable).
6. Samples MUST be **regenerated** or **re-seeded** when appropriate on empty store per persistence rules in architecture (e.g. migration seed, idempotent startup hook).
7. The **cube** sample MUST satisfy the **face-plane** and **even distribution** rules in **Scope** (lights on **all six** **faces**, **not** **edge-only**; **even** spread over **face** **areas**).
8. The **sphere** and **cone** samples MUST satisfy the **surface** and **even distribution** rules in **Scope** (**area**-wide **even** coverage; **not** **edge-only** or **single-curve-only** layouts that fail those intent tests).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Samples appear in the list like other models; readable names.
- Tablet: Same as mobile.
- Desktop: Same.

**Dependencies**

- REQ-005, REQ-006

**Open questions**

- Whether samples are **removed** if the user deletes them manually.
- Whether the sampling path for each shape MUST be a **single** open polyline, a **closed** loop on the surface, or multiple patches (must still meet spacing, **evenness**, and **face**/**surface** rules).
- Quantitative definition of **“evenly distributed”** (e.g. minimum spacing between **any** pair, **Voronoi** / **blue-noise**-style metrics, or maximum **local** density **ratio**).

---

### REQ-010 — Three.js visualization when viewing a model

| Field | Value |
|-------|-------|
| **ID** | REQ-010 |
| **Title** | Three.js visualization when viewing a model |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, when I **view** a model, I want **every** light drawn as a **2 cm white sphere**, each light linked to its **previous** and **next** along the wire (**REQ-005**) by **barely visible** **`#B5B5B5`** **light-grey** segments where those neighbors exist, and a way to see each light’s **index and coordinates**—so that I can interpret spatial layout and **wire order** without missing lights.

**Scope**

- In scope: On the **model view** / **detail** experience (the same flow as **viewing** a single model in **REQ-006**), the product MUST render the model’s light positions in **3D** using **three.js** (the JavaScript library from https://threejs.org/ , typically consumed as the `three` package). **Every** light returned for the model MUST be **drawn** (no deliberate omission, decimation, or level-of-detail that hides lights). **Each** light MUST appear as a **sphere** of **2 cm** diameter (**0.02 m**; **SI meters** per **REQ-005**). **Default** sphere appearance (no per-light state or **REQ-012** default) MUST be **white** with **solid** (opaque) **fill**. **Straight line segments** MUST connect **only** **consecutive** lights in **ascending id order**, consistent with **REQ-005** adjacency (one segment per pair **(i, i+1)** for **i = 0 … n − 2** when **n > 1**). Segments MUST use **`#B5B5B5`** (**canonical `#RRGGBB` light grey**) and **75% transparent**—i.e. **25% opaque** (e.g. alpha **0.25** in an **0–1** opacity scale)—and **thin** enough that they read as **subtle** guidance, **barely visible**, and **less prominent** than the light **spheres** (including **on** lights at **100%** brightness). **Pointer hover** over a light’s sphere MUST reveal that light’s **id** and **x**, **y**, **z**. **Touch** and **tablet** users MUST have an **equivalent** way to discover id and coordinates (**REQ-002**).
- Out of scope: Non-three.js renderers; **export** of screenshots or video; **editing** positions in the 3D view; segments between **non-consecutive** ids (e.g. skipping **i+1**); **animated** or **pulsing** hover chrome beyond showing the required data.

**Business rules**

1. Whenever the user is **viewing** a single model’s content (lights and metadata context per **REQ-006**), the UI MUST include a **three.js**-based **3D view** that reflects the model’s **x**, **y**, **z** positions for **each** light in the payload (**all lights drawn**).
2. The **three.js** dependency MUST be a **direct** front-end dependency (declared in the Next.js app’s package manifest), not loaded only indirectly through unrelated packages, so the visualization stack is explicit and auditable.
3. The 3D view MUST remain **usable** on **mobile**, **tablet**, and **desktop** viewports per **REQ-002** (e.g. visible canvas or viewport, no reliance on desktop-only interaction for basic inspection unless alternatives are documented in architecture).
4. Each light MUST be represented by a **sphere** with **diameter exactly 0.02 m** (**2 cm**). **Colour and on/off appearance** MUST follow **REQ-012** when per-light state is available; until then or where state is not defined, the sphere MUST appear **white** with **solid fill** (exact material or shading is deferred to architecture).
5. For each **i** from **0** to **n − 2** (where **n** is the light count), a **single straight Euclidean segment** MUST be drawn between the positions of light **i** and light **i + 1** (this matches **REQ-005** chain adjacency). Segments MUST use **`#B5B5B5`** and **75% transparency** (**25% opacity**) as in **Scope**, remain **thin**, and MUST **not** appear **more visually prominent** than the light spheres. If **n ≤ 1**, no segments are required.
6. When the user moves the **pointer** so that it **hovers** a light’s **sphere** (primary **hit target** for that light), the UI MUST show that light’s **id** and **x**, **y**, **z** (same numeric meaning as the API / **REQ-005**). On **touch-first** devices, the product MUST provide a **documented** equivalent (e.g. **tap** to select the nearest light or the picked light) that surfaces the same **id** and **coordinates** without requiring hover.
7. The renderer MUST NOT skip or merge lights for performance when **n ≤ 1000** (per **REQ-005**); **all n** spheres and the **n − 1** (or **0**) segments described above MUST be present.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Pan/rotate/zoom as today; **tap** (or equivalent) to show **id** + **x,y,z** for a light; tooltip or panel MUST remain readable at small widths.
- Tablet: Same as mobile; orientation changes handled where applicable.
- Desktop: **Hover** on a light sphere shows **id** and **coordinates**; pointer still works with orbit-style navigation without constant accidental hovers (architecture may tune hit radius or affordances).

**Dependencies**

- REQ-001, REQ-002, REQ-006, REQ-005

**Open questions**

- Minimum three.js **major** version or LTS alignment (if any).
- Whether **orbit controls** or another interaction model is mandatory vs architecture choice.
- Preferred **decimal precision** or **formatting** for displayed coordinates (e.g. fixed decimals vs full float).

---

### REQ-011 — REST API for per-light state (query and control)

| Field | Value |
|-------|-------|
| **ID** | REQ-011 |
| **Title** | REST API for per-light state (query and control) |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As a user or integrator, I want a **REST API** on the model that lets me **read** and **individually update** each light’s **state**, so that external tools or the UI can drive and inspect lights programmatically.

**Scope**

- In scope: HTTP **REST** resources or operations scoped to a **model** and **individual lights** (by light **id** per **REQ-005**). **Read** operations that return the **current** state for one or all lights in the model. **Write** operations that set state **per light** (not batch-only). State fields: **on** or **off** (boolean), **colour** as a **hex** RGB string, and **brightness** as a **percentage** (0–100 inclusive, semantics below).
- Out of scope: Non-REST protocols; authentication and authorization policy (unless already required elsewhere); animation timelines; physical hardware protocols.

**Business rules**

1. The API MUST allow **querying** the **current** state of lights for a model: **all lights** in **one** response **and** the state of **one** light by **id** (exact resource layout deferred to **docs/architecture.md**).
2. The API MUST allow **updating** the state of **each** light **individually** (by **id**), including **on/off**, **hex colour**, and **brightness** **percentage**; partial updates that change only some fields MUST be supported where REST semantics allow (e.g. **PATCH**-style behavior).
3. **Hex colour** MUST use a **canonical** form agreed in architecture (e.g. **`#RRGGBB`** with six **hexadecimal** digits); invalid values MUST be rejected with a **clear** error.
4. **Brightness** MUST be a **number** interpreted as **percent** with **0** = minimum and **100** = maximum for the **on** appearance; behavior when the light is **off** (whether brightness is stored or ignored visually) is defined in **REQ-012**.
5. Successful **writes** MUST be **persisted** with the model (same durability as model data per architecture) so reloads and other clients see the same state.
6. **Default** state for lights in a newly created or legacy model without prior state MUST match **REQ-014** (all **off**, **100%** brightness, white **`#FFFFFF`** in canonical hex) and MUST be **consistent** across API and UI; **docs/architecture.md** MAY elaborate representation only.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A for API-only requirement; any UI that calls the API MUST remain usable per **REQ-002**.
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-001, REQ-005, REQ-006

**Open questions**

- Whether **bulk** read/write endpoints are **required** in addition to per-light operations.
- Exact **HTTP methods** and **URL** patterns (architecture).

---

### REQ-012 — Model visualization reflects per-light state

| Field | Value |
|-------|-------|
| **ID** | REQ-012 |
| **Title** | Model visualization reflects per-light state |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user viewing a model, I want the **3D visualization** to **match** each light’s **stored state**, so that **colour**, **brightness**, and **on/off** are **immediately obvious** from the **spheres** (e.g. **blue** when set to blue, subtle **#B5B5B5** when **off**).

**Scope**

- In scope: When viewing a model (**REQ-006**, **REQ-010**), each light’s **sphere** (**0.02 m** diameter) MUST reflect **REQ-011** state: **on** = **filled** **opaque** appearance using the light’s **hex colour** modulated by **brightness**; **off** = appearance using **`#B5B5B5`** and **75% transparency** (**25% opacity**) as **REQ-010** wire segments, applied to the **sphere** (filled or thin shell—exact geometry deferred to architecture) so **off** lights remain **discernible** but **less prominent** than **on** lights. **Any** change to persisted light state via the API (**REQ-011**) MUST be **reflected** in the visualization **without** requiring a full page reload if the client is already viewing that model (e.g. **poll**, **push**, or **same-session** refetch—architecture defines mechanism; the requirement is **timely** consistency from the user’s perspective).
- Out of scope: Changing **wire segment** styling based on **on/off** state unless added later; **export** of rendered images.

**Business rules**

1. For a light that is **on**, the **sphere** MUST appear **filled** (solid or equivalent **opaque** **surface** fill) and MUST use the **current** **hex colour** and **brightness** from **REQ-011** (architecture defines how **percentage** maps to rendered colour or intensity).
2. For a light that is **off**, the **sphere** MUST use **`#B5B5B5`** and **75% transparency** (**25% opacity**), **consistent** with **REQ-010** segment styling, so it reads as **not lit** yet remains **locatable**; it MUST **not** appear **more visually prominent** than **on** lights or than the **wire segments**.
3. The visualization MUST **update** when light state changes from **REQ-011** while the user is viewing the affected model, so the **sphere** **appearance** matches the **latest** **persisted** state within a **reasonable** delay (architecture may set bounds; **stale** display after a successful write is **not** acceptable indefinitely).
4. **REQ-010** **segments** and **hover**/**touch** **id** and **coordinates** behavior remain in force; state fields (**on/off**, **colour**, **brightness**) MAY be shown on hover/tap in addition if architecture chooses.
5. Lights **without** stored state yet MUST follow the **default** defined for **REQ-011** and still render per rules **1** and **2** above.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: On/off and colour cues MUST remain **discernible** at small viewports (contrast, transparency level tuned per **REQ-002**).
- Tablet: Same as mobile.
- Desktop: Same; **hover** may also show state fields if implemented.

**Dependencies**

- REQ-002, REQ-010, REQ-011

**Open questions**

- Maximum acceptable **latency** after an API update before the 3D view refreshes.
- Whether **hover**/**tap** must display **brightness** and **colour** in addition to **id**/**coordinates**.

---

### REQ-013 — Model view: multi-select light settings and paged light list

| Field | Value |
|-------|-------|
| **ID** | REQ-013 |
| **Title** | Model view: multi-select light settings and paged light list |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user **viewing** a model, I want to **select multiple lights** and **apply the same settings** to all of them at once, and I want the **list of lights** to be **paginated** with control over **how many lights appear per page** and a way to **jump to a specific light by its id**, so that I can manage large models without scrolling through thousands of rows and can batch-update state efficiently.

**Scope**

- In scope: On the **model view** (**REQ-006**), a **list** (or equivalent tabular presentation) of the model’s lights that is **paged** (only one **page** of lights shown at a time). Controls to **change the page size** (number of lights per page) from **presets** or an agreed control pattern. A **“go to light”** (or equivalent) control that accepts a light **id** (**REQ-005**) and navigates the list to the **page** that contains that id, with **clear feedback** if the id is **invalid** or **out of range** for the model. **Multi-select** of lights **on the current page** (and, if architecture supports it, **across pages** via retained selection—see open questions) plus a **bulk apply** action that sets **on/off**, **hex colour**, and **brightness** (**REQ-011**) **to the same values** for **every selected** light; successful updates MUST **persist** per **REQ-011**.
- Out of scope: New state fields beyond **REQ-011**; reordering lights; deleting lights from the model; export of selection.

**Business rules**

1. The model view MUST include a **paginated** light list when the model has **more than one** light; for a **single** light, pagination MAY be omitted or trivial (architecture).
2. The user MUST be able to **change how many lights** are shown **per page** using at least **three** distinct positive page-size choices within **1** and **1000** (inclusive), documented or labeled in the UI (exact values deferred to architecture).
3. The user MUST be able to **jump** to the page containing a given light **id** by entering or choosing that **id**; if the id is **not** an integer in **0 … n − 1** for the model’s light count **n**, the system MUST **not** change page silently and MUST show **actionable** feedback.
4. The user MUST be able to **select multiple lights** (multi-select) from the list using **pointer** and **keyboard**-friendly patterns on desktop and an **equivalent** on **touch** (**REQ-002**), e.g. checkboxes or a documented multi-select gesture.
5. When **one or more** lights are selected, the user MUST be able to **apply** settings (**on/off**, **hex colour**, **brightness** per **REQ-011**) so that **each selected** light receives the **same** applied values; validation rules for colour and brightness MUST match **REQ-011**.
6. After a **successful** bulk apply, **REQ-012** applies: the **3D view** and list MUST reflect updated state within the same **timeliness** expectations as single-light updates (**no indefinite staleness** after success).
7. Pagination and multi-select MUST remain **usable** on **mobile**, **tablet**, and **desktop** (**REQ-002**): primary controls reachable, readable labels, and no reliance on **hover-only** affordances for essential actions.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Page size and “go to id” controls usable without horizontal scrolling for primary content; multi-select via touch-friendly controls; bulk apply clearly labeled.
- Tablet: Same as mobile; orientation changes handled where applicable.
- Desktop: Keyboard navigation where applicable; multi-select with checkboxes or shift-range if architecture provides it.

**Dependencies**

- REQ-002, REQ-006, REQ-011, REQ-012

**Open questions**

- Whether **selection** is **per-page only** or MUST **persist** when the user changes pages (cross-page bulk apply).
- Whether **REQ-011** needs an explicit **batch** HTTP operation for performance; until then, repeated per-light updates MAY satisfy rule **5** if UX meets timeliness (**rule 6**).
- Exact **preset** page sizes and default page size.

---

### REQ-014 — Default light state (off, 100% white) and model reset

| Field | Value |
|-------|-------|
| **ID** | REQ-014 |
| **Title** | Default light state (off, 100% white) and model reset |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user working with a model, I want **every** light to **start** in a **known** state—**off** but configured as **full white** at **100%** brightness—and I want a **reset** control that returns **all** lights to that state after I have changed them, so that behaviour is predictable and I can quickly restore a clean baseline.

**Scope**

- In scope: **Initial** per-light state for **each** light when a model first exists in the system (including **new** models after **CSV** upload or **seeded** samples, and **legacy** rows **backfilled** per architecture): **on** = **false** (**off**), **brightness** = **100** (percent, per **REQ-011**), **hex colour** = **`#FFFFFF`** (six-digit canonical form per **REQ-011**). On the **model view** (**REQ-006**), a **reset** affordance (e.g. a **Reset** button or equivalent clearly labeled control) that sets **every** light in the **current** model to that **same** triple (**off**, **100%**, **`#FFFFFF`**) in **one** deliberate user action, with **persistence** per **REQ-011**.
- Out of scope: **Undo**/**redo** stacks; resetting **only** selected lights; resetting **3D** camera or navigation; **bulk** HTTP beyond what architecture needs to implement the single user gesture.

**Business rules**

1. For **every** light in a model, the **initial** stored state (including after model **creation** and any **migration** of older data) MUST be **off** (**on** = **false**), **brightness** = **100**, and **hex colour** = **`#FFFFFF`** (rejecting invalid hex remains per **REQ-011**).
2. The **model view** MUST include a **reset** control, **reachable** and **operable** without **hover-only** interaction (**REQ-002**), that restores **all** lights in that model to the state in rule **1** and **persists** it so **API** clients and **reloads** see the same result.
3. After a **successful** reset, the **3D** visualization (**REQ-012**) and any **light list** showing state (**REQ-013** where applicable) MUST **reflect** the new state within the same **timeliness** expectations as other **REQ-011** writes (**no indefinite staleness**).
4. **REQ-011** default semantics for lights **without** prior stored state MUST align with rule **1** (not **on** at default unless the user or API has turned them **on**).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Reset control visible and tappable; label makes intent clear (e.g. “Reset lights” or “Reset all lights”).
- Tablet: Same as mobile; orientation changes do not hide the control in the primary model view.
- Desktop: Reset available alongside other model-view actions; keyboard focus order includes the control where applicable.

**Dependencies**

- REQ-002, REQ-006, REQ-011, REQ-012, REQ-013

**Open questions**

- Whether the reset action **requires** a **confirmation** step for large models (architecture / UX choice).

---
