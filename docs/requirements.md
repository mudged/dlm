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

- In scope: User-facing flows (and backing behavior per architecture) to **list** all stored models, **view** details of a selected model (including its lights and metadata as appropriate), **delete** an existing model, and **create** a new model by **uploading** a CSV in the format defined under REQ-005. **Deletion** MUST respect **scene** membership per **REQ-015**: a model that **is assigned** to **one or more** scenes MUST **not** be deletable until that **reference** is **resolved** (see business rules).
- Out of scope: Bulk export, editing lights in-place after import, versioning, and sharing links unless added later.

**Business rules**

1. The application MUST expose a way to see **all** models the user may access (subject to any future auth).
2. The application MUST allow **viewing** a single model’s content (lights and metadata) after selection.
3. The application MUST allow **deleting** a model such that it no longer appears in listings and is removed per persistence rules in architecture, **except** where rule **5** applies.
4. The application MUST allow **creating** a model by **uploading** a CSV file; successful creation MUST persist the model and its metadata (name and creation date supplied or derived per architecture—e.g. user-provided name at upload time).
5. When **REQ-015** is in effect, the system MUST **refuse** to **delete** a model that is **assigned** to **one or more** scenes. The user MUST be **clearly informed** **why** deletion is blocked: the model **is in use** by **one or more** scenes, **that** scenes **reference** models, and **what** the user can do next (e.g. **remove** the model from **each** listed scene or **delete** those scenes)—exact presentation (list of scene names, links, etc.) is deferred to architecture.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: List and detail usable without horizontal scrolling for primary content; upload control reachable and labeled clearly.
- Tablet: List and detail use space efficiently; upload remains obvious in the create flow.
- Desktop: Efficient browsing of lists and details; keyboard-accessible actions where controls exist.

**Dependencies**

- REQ-001, REQ-002, REQ-005, REQ-015

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

As a user, when I **view** a model, I want **every** light drawn as a **2 cm white sphere**, each light linked to its **previous** and **next** along the wire (**REQ-005**) by **barely visible** **`#D0D0D0`** **light-grey** segments where those neighbors exist, and a way to see each light’s **index and coordinates**—so that I can interpret spatial layout and **wire order** without missing lights.

**Scope**

- In scope: On the **model view** / **detail** experience (the same flow as **viewing** a single model in **REQ-006**), the product MUST render the model’s light positions in **3D** using **three.js** (the JavaScript library from https://threejs.org/ , typically consumed as the `three` package). **Every** light returned for the model MUST be **drawn** (no deliberate omission, decimation, or level-of-detail that hides lights). **Each** light MUST appear as a **sphere** of **2 cm** diameter (**0.02 m**; **SI meters** per **REQ-005**). **Default** sphere appearance (no per-light state or **REQ-012** default) MUST be **white** with **solid** (opaque) **fill**. **Straight line segments** MUST connect **only** **consecutive** lights in **ascending id order**, consistent with **REQ-005** adjacency (one segment per pair **(i, i+1)** for **i = 0 … n − 2** when **n > 1**). Segments MUST use **`#D0D0D0`** (**canonical `#RRGGBB` light grey**, softer than mid-grey so the wire reads as **ambient** rather than **emphasis**) and **85% transparent**—i.e. **15% opaque** (e.g. alpha **0.15** in an **0–1** opacity scale)—and **thin** enough that they read as **subtle** guidance, **barely visible**, and **less prominent** than the light **spheres** (including **on** lights at **100%** brightness). **Pointer hover** over a light’s sphere MUST reveal that light’s **id** and **x**, **y**, **z**. **Touch** and **tablet** users MUST have an **equivalent** way to discover id and coordinates (**REQ-002**).
- Out of scope: Non-three.js renderers; **export** of screenshots or video; **editing** positions in the 3D view; segments between **non-consecutive** ids (e.g. skipping **i+1**); **animated** or **pulsing** hover chrome beyond showing the required data.

**Business rules**

1. Whenever the user is **viewing** a single model’s content (lights and metadata context per **REQ-006**), the UI MUST include a **three.js**-based **3D view** that reflects the model’s **x**, **y**, **z** positions for **each** light in the payload (**all lights drawn**).
2. The **three.js** dependency MUST be a **direct** front-end dependency (declared in the Next.js app’s package manifest), not loaded only indirectly through unrelated packages, so the visualization stack is explicit and auditable.
3. The 3D view MUST remain **usable** on **mobile**, **tablet**, and **desktop** viewports per **REQ-002** (e.g. visible canvas or viewport, no reliance on desktop-only interaction for basic inspection unless alternatives are documented in architecture).
4. Each light MUST be represented by a **sphere** with **diameter exactly 0.02 m** (**2 cm**). **Colour and on/off appearance** MUST follow **REQ-012** when per-light state is available; until then or where state is not defined, the sphere MUST appear **white** with **solid fill** (exact material or shading is deferred to architecture).
5. For each **i** from **0** to **n − 2** (where **n** is the light count), a **single straight Euclidean segment** MUST be drawn between the positions of light **i** and light **i + 1** (this matches **REQ-005** chain adjacency). Segments MUST use **`#D0D0D0`** and **85% transparency** (**15% opacity**) as in **Scope**, remain **thin**, and MUST **not** appear **more visually prominent** than the light spheres. If **n ≤ 1**, no segments are required.
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

As a user viewing a model, I want the **3D visualization** to **match** each light’s **stored state**, so that **colour**, **brightness**, and **on/off** are **immediately obvious** from the **spheres** (e.g. **blue** when set to blue, subtle **#D0D0D0** when **off**).

**Scope**

- In scope: When viewing a model (**REQ-006**, **REQ-010**), each light’s **sphere** (**0.02 m** diameter) MUST reflect **REQ-011** state: **on** = **filled** **opaque** appearance using the light’s **hex colour** modulated by **brightness**; **off** = appearance using **`#D0D0D0`** and **85% transparency** (**15% opacity**) as **REQ-010** wire segments, applied to the **sphere** (filled or thin shell—exact geometry deferred to architecture) so **off** lights remain **discernible** but **less prominent** than **on** lights. **Any** change to persisted light state via the API (**REQ-011**) MUST be **reflected** in the visualization **without** requiring a full page reload if the client is already viewing that model (e.g. **poll**, **push**, or **same-session** refetch—architecture defines mechanism; the requirement is **timely** consistency from the user’s perspective).
- Out of scope: Changing **wire segment** styling based on **on/off** state unless added later; **export** of rendered images.

**Business rules**

1. For a light that is **on**, the **sphere** MUST appear **filled** (solid or equivalent **opaque** **surface** fill) and MUST use the **current** **hex colour** and **brightness** from **REQ-011** (architecture defines how **percentage** maps to rendered colour or intensity).
2. For a light that is **off**, the **sphere** MUST use **`#D0D0D0`** and **85% transparency** (**15% opacity**), **consistent** with **REQ-010** segment styling, so it reads as **not lit** yet remains **locatable**; it MUST **not** appear **more visually prominent** than **on** lights or than the **wire segments**.
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

### REQ-015 — Scenes: composite space, placement, CRUD, and three.js view

| Field | Value |
|-------|-------|
| **ID** | REQ-015 |
| **Title** | Scenes: composite space, placement, CRUD, and three.js view |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, I want **scenes** that group **one or more** light **models** in a shared **3D** space with a **name**, **automatic** **integer** **placement** **computed** **on** **create** **so** **each** **model** **fits** **fully** **inside** the **scene** **(including** **boundary** **rules)**, **without** **changing** **stored** **model** **coordinates**, so that **the** **composite** **view** **shows** **lights** **in** **scene** **coordinates** **while** **the** **model** **detail** **view** **still** **shows** **original** **coordinates**; I can **list**, **delete**, and **open** a scene, **visualize** with **three.js**, **add** further models (default **to** **the** **right**), **adjust** placements, and **remove** models—including **confirming** when **removing** the **last** one would **delete** the **whole** scene.

**Scope**

- In scope: A **scene** entity with **metadata** including a **name**; **persistence** and user flows to **create** (see business rules), **list**, **delete**, and **select** a scene. **Scene creation** MUST include **one or more** **models** (per **REQ-005** / **REQ-006**) chosen at **create** **time**; **initial** **placements** MUST be **computed** **automatically** **per** **business** **rule** **2** (**not** **user-entered** **offsets** **as** **the** **default** **path**). **Coordinate** **separation** (**rule** **4**): **canonical** **model** **light** **x**, **y**, **z** **in** **persistence** MUST **not** **be** **updated** **because** **of** **scenes**; **scene** **API** **and** **3D** **composite** **view** **use** **derived** **scene-space** **positions** (**canonical** **plus** **offsets**). A scene **always** has **at least one** model **while** it **exists**; there are **no** persisted **empty** scenes. **Placement** **offsets** are **integers** **≥ 0** (**rule** **5**). **Negative** **scene-space** **coordinates** **for** **any** **light** **are** **forbidden**; the **valid** **region** is the **non-negative** **octant** **from** **(0, 0, 0)**. The **scene**’s **display** **volume** **adjusts** with **at least 1 SI meter** **margin** **beyond** **tight** **bounds** **(rule** **7**). When the user **views** a **scene**, **three.js** (**REQ-010**, **REQ-012**) **draws** **every** **light** **at** **derived** **scene** **positions**, **segments** **only** **within** **each** **model’s** **id** **chain**. **Add**/**remove** **models** **and** **optional** **placement** **edits** **after** **create** **per** **rules** **6**, **10**, **11**.
- Out of scope: **Editing** individual light **positions** inside the scene view (only membership and **model-level** placement unless a later requirement adds it); **import**/**export** of scene definitions as files; **rotation** or **scale** of models within a scene unless added later.

**Business rules**

1. The application MUST allow the user to **create** a scene, **list** all scenes, **delete** a scene (explicit **whole-scene** delete), and **open**/**select** a scene for viewing.
2. **Creating** a scene MUST require a **name** and **one or more** **models** to **attach** in **that** **same** **flow** (multi-step wizard is acceptable if **creation** **cannot** **complete** without **at least** **one** model **saved** with the scene). For **each** **model** **chosen** **at** **create** **time**, the system MUST **automatically** **compute** **integer** **offsets** **ox**, **oy**, **oz** **≥ 0** such that **every** **light** **of** **that** **model** lies **fully** **within** the **non-negative** **scene** **region** **after** **composition** **with** **that** **model’s** **stored** **REQ-005** **coordinates**, **and** **such** **that** **the** **order** **of** **models** **in** **the** **create** **flow** **yields** **a** **valid** **combined** **layout** (**each** **successive** **model** **positioned** **relative** **to** **those** **already** **placed**, **consistent** **with** **rule** **6** **for** **horizontal** **stacking** **where** **applicable**). **Automatic** **calculation** MUST **keep** **each** **model’s** **entire** **footprint** **inside** the **allowed** **non-negative** **scene** **space** **and** **satisfy** **the** **display**/**framing** **boundary** **rules** (**rules** **5**, **7**, **and** **architecture**). The **create** **flow** MUST **not** **require** **the** **user** **to** **type** **or** **pick** **numeric** **offsets** **for** **initial** **placement** (**architecture** **MAY** **offer** **optional** **advanced** **override** **only** **if** **requirements** **are** **still** **met** **without** **it**).
3. A scene MUST have a **name** (uniqueness policy deferred to open questions / architecture). A persisted scene MUST **always** have **at least one** model **until** it is **deleted** as a **whole** or via rule **10**.
4. **Persisted** **model** **data** (**REQ-005** **light** **x**, **y**, **z** **in** **storage**) MUST **remain** **unchanged** **when** **a** **model** **is** **assigned** **to**, **moved** **within**, or **removed** **from** **a** **scene**; **scene** **membership** **only** **adds** **or** **updates** **placement** **offsets** **and** **does** **not** **rewrite** **the** **model’s** **canonical** **coordinates**. **Scene-space** **positions** **used** **for** **rendering** **and** **disclosed** **in** **the** **scene** **context** **are** **derived** **from** **canonical** **coordinates** **plus** **offsets**.
5. Each **persisted** **placement** (**offsets** **ox**, **oy**, **oz**) MUST be **integers** **≥ 0**. For **every** light, the **composed** **scene-space** coordinates MUST satisfy **x, y, z ≥ 0**; **no** light **position** may lie **outside** that **non-negative** **region** (**fully** or **partially**). The system MUST **reject** or **block** **invalid** placements (**including** **user** **edits** **after** **create**) with **clear** feedback.
6. When **adding** a model to a scene that **already** contains **at least one** model, the **default** **initial** placement MUST position the **new** model **to the** **right** of the **existing** **layout**—i.e. **offset** along the **scene** axis that **architecture** defines as **“right”** in the **default** **camera**/**viewer** **convention**—so it sits **beside** the **current** **combined** **axis-aligned** **footprint** **without** **overlap** at **default** (exact **gap** or **touching** rule per architecture). The **scene** **volume** MUST **grow** (or **otherwise** **adjust**) **automatically** to **satisfy** **rules** **5** and **7**.
7. The **scene**’s **display** volume MUST **automatically** **size** so that it **encloses** **all** placed model geometry with a **margin** of **at least 1 meter** beyond the **axis-aligned bounds** of **all** light positions in **scene** space (how **1 m** is **allocated** across **faces** is architecture).
8. When viewing a scene, the product MUST use **three.js** as a **direct** front-end dependency (**REQ-010**) and MUST render **all** lights of **all** assigned models **without** omission when **n ≤ 1000** per model (**REQ-005**), with **segments** only between **consecutive** **ids** **within** the **same** model (**REQ-005**), **not** between models, **using** **derived** **scene-space** **positions** (**rule** **4**).
9. **Per-light** **state** (**REQ-011**, **REQ-012**) MUST apply in the scene view **per model** as in the single-model view (colours, brightness, on/off, and **timely** updates).
10. **Removing** a model when **more than one** model remains MUST **persist** **without** deleting the scene. **Removing** the **last** **remaining** model MUST **not** **silently** delete the scene: the user MUST **first** **confirm** a **clear** **explanation** that **removing** this **last** model will **delete** the **entire** scene (and **that** the scene will **disappear** from **lists**). On **confirm**, the system MUST **delete** the scene (and **remove** the model assignment). On **cancel**, the scene MUST **remain** **unchanged**. This **confirm** flow MUST be **usable** without **hover-only** **essential** steps (**REQ-002**).
11. The scene view MUST provide **affordances** to **add** a **model** (from those available per **REQ-006**) and to **remove** a model, subject to **rules** **6** and **10**; the user MUST be able to **adjust** placements for models **already** in the scene (**subject** to **rule** **5**); **successful** changes MUST **persist** per architecture **without** **altering** **stored** **model** **light** **coordinates** (**rule** **4**).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Scene list, scene detail/view, **create** **without** **manual** **offset** **entry**, add/remove, placement editing, and **last-model** **confirmation** are reachable and readable; **three.js** viewport usable like the model view (**REQ-010**); **touch** equivalent for light **id** and **scene**/**model** **coordinate** **disclosure** where applicable.
- Tablet: Same as mobile; orientation changes handled where applicable.
- Desktop: Efficient navigation between scene list and scene view; **hover** on lights where applicable.

**Dependencies**

- REQ-002, REQ-005, REQ-006, REQ-010, REQ-011, REQ-012

**Open questions**

- Whether **scene names** must be **globally unique**.
- Exact mapping of **(0,0,0)** and **which** **axis** is **“right”** vs **“up”** in the **default** **three.js** **scene** setup.
- Whether **placement** integers are **meters** rounded, **decimeters**, or **unitless grid** cells (must be **consistent** with **REQ-005** **meter** coordinates in architecture).

---

### REQ-016 — Camera reset for model and scene 3D views

| Field | Value |
|-------|-------|
| **ID** | REQ-016 |
| **Title** | Camera reset for model and scene 3D views |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user viewing a **model** or a **scene** in **3D**, I want a **camera reset** control that returns the view to its **default** position and orientation, so that I can recover a predictable framing after panning, rotating, or zooming.

**Scope**

- In scope: On **single-model** **three.js** visualization (**REQ-010**, **REQ-006**) and **scene** **three.js** visualization (**REQ-015**), a **dedicated** user affordance (e.g. **“Reset camera”** or equivalent clear label) that restores the **default** camera (and any **architecture-defined** default **controls** state, e.g. orbit target) to the **same** baseline as when the user **first** opens that view in the session—or the product’s **documented** default if architecture fixes a single global default per view type.
- Out of scope: **Persisting** camera pose across sessions; **per-user** saved views; changing **light** state (**REQ-014**); **screenshots** or **export**.

**Business rules**

1. The **model** **detail**/**view** experience MUST expose a **camera reset** affordance **adjacent** to or **within** the **primary** **3D** viewport region (exact placement deferred to architecture) so the user can **find** it **without** leaving the view.
2. The **scene** **view** experience MUST expose the **same** class of **camera reset** affordance for the **scene** **three.js** viewport.
3. Activating **camera reset** MUST **only** affect **client-side** **navigation** (camera and control state); it MUST **not** alter **persisted** **models**, **scenes**, **placements**, or **per-light** **state** (**REQ-005**–**REQ-015**).
4. After **camera reset**, the **visible** **framing** MUST match the **default** **defined** in **docs/architecture.md** for that **view** **type** (model vs scene), so **repeat** activations yield the **same** baseline (modulo viewport size).
5. The control MUST be **reachable** and **operable** on **mobile**, **tablet**, and **desktop** **without** **hover-only** **essential** steps (**REQ-002**).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Reset control **tappable**, label **clear**; **no** reliance on **hover** for the reset action itself.
- Tablet: Same as mobile; **orientation** changes **do not** remove the control from the primary view chrome.
- Desktop: Reset available **near** the **3D** view; **keyboard** **focus** **order** **includes** the control where applicable.

**Dependencies**

- REQ-002, REQ-010, REQ-015

**Open questions**

- Whether **reset** also **clears** **transient** **UI** (e.g. **selected** light for **tap**/**hover** **details**) or **only** the **camera**.

---

### REQ-017 — Options: factory reset with confirmation

| Field | Value |
|-------|-------|
| **ID** | REQ-017 |
| **Title** | Options: factory reset with confirmation |
| **Priority** | Must |
| **Actor(s)** | End user; operator |

**User story**

As a user, I want an **Options** area that includes **factory reset**, so that I can **wipe** **all** **application** **data** and return to a **clean** state **with** only the **default** **sample** **models** (**REQ-009**) **restored**—but **only** after I **confirm** **explicitly** **because** the **operation** is **dangerous** and **irreversible**.

**Scope**

- In scope: A **distinct** **Options** **section** (or **screen**/**panel** labeled **Options** or equivalent **clear** **navigation** **target**) in the **UI** that includes an action labeled **Factory reset** (or **equivalent** **unambiguous** **wording**). **Factory reset** MUST **remove** **all** **persisted** **user-relevant** **data** the product stores for **models**, **scenes** (**REQ-015**), **per-light** **state** (**REQ-011**), **scene routines** (**REQ-021** and **REQ-022** definitions and any **persisted** **run** **state**), and **any** **other** **application** **content** **tied** to those **entities** (exact **store** **shape** per **architecture**), then **re-seed** the **system** so the **user** sees the **same** **default** **sample** **models** as on a **fresh** **install** per **REQ-009** (three **samples**; **no** **user-uploaded** **models** or **user-created** **scenes** **remain**). **Before** **any** **deletion** or **re-seed** **runs**, the user MUST be **prompted** with a **confirmation** **step** that **warns** of **data** **loss** and **irreversibility**; **Cancel** MUST **leave** **data** **unchanged**; **Confirm** MUST **complete** the **reset** and **surface** **success** **feedback** (exact **copy** deferred to **architecture**).
- Out of scope: **Partial** **wipe** (e.g. **only** **scenes**); **scheduled** **reset**; **remote** **admin** **API** **for** **factory** **reset** unless added later; **export** **before** **wipe** (user may **export** **elsewhere** if **features** **exist**—not **required** here).

**Business rules**

1. The **product** MUST expose an **Options** **section** (or **dedicated** **Options** **view**) **discoverable** from **primary** **navigation** or **settings** **pattern** **documented** in **architecture**; it MUST **list** **Factory reset** as **one** of its **actions**.
2. **Factory reset** MUST **not** **run** on a **single** **mis-click**: the user MUST **first** **see** a **blocking** **prompt** or **dialog** **before** **irreversible** **effects** **begin**, **explaining** that **all** **models**, **scenes**, **routines** (**REQ-021**, **REQ-022**), and **related** **data** will be **permanently** **removed** and **only** **default** **samples** will **remain** (wording **may** **name** **consequences** **explicitly** per **UX** **review**).
3. **Until** the user **explicitly** **confirms** (e.g. **Confirm** on the **dialog**), **no** **factory** **reset** **side** **effects** **may** **occur**; **dismissal** or **Cancel** MUST **preserve** **current** **data**.
4. After **confirmed** **factory reset**, **no** **user-created** **models**, **no** **scenes**, **no** **routine** **definitions** (**REQ-021**, **REQ-022**), **no** **persisted** **routine** **run** **state**, and **no** **leftover** **state** **from** **prior** **entities** **may** **remain** **visible** in **listings** (or **equivalent** **discovery** **surfaces**); the **model** **list** MUST **match** **REQ-009** **expectations** for a **fresh** **seed** (**exactly** **three** **sample** **models** **identifiable** as **sphere**, **cube**, **cone** per **REQ-009** **naming** **rules**).
5. **Per-light** **defaults** after **re-seed** MUST **align** with **REQ-014** / **REQ-011** for **newly** **present** **models**.
6. The **entire** **flow** MUST **satisfy** **REQ-002**: **no** **hover-only** **requirement** for **opening** **Options**, **starting** **factory** **reset**, **or** **confirming** / **canceling**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Options** and **factory** **reset** **reachable** **without** **horizontal** **scroll** **for** **primary** **labels**; **confirmation** **readable** **without** **truncating** **the** **warning**.
- Tablet: Same as mobile.
- Desktop: **Options** **easy** **to** **find**; **dialog** **keyboard** **accessible** where **modals** **are** **used**.

**Dependencies**

- REQ-002, REQ-006, REQ-009, REQ-011, REQ-014, REQ-015, REQ-021, REQ-022

**Open questions**

- Whether **factory** **reset** **requires** **typing** a **phrase** (e.g. **“RESET”**) **in** **addition** **to** **Confirm** for **extra** **safety**.
- **Post-reset** **navigation** (**stay** **on** **page** vs **redirect** **to** **model** **list**).

---

### REQ-018 — Application shell: themes, collapsible navigation, branding, and Font Awesome controls

| Field | Value |
|-------|-------|
| **ID** | REQ-018 |
| **Title** | Application shell: themes, collapsible navigation, branding, and Font Awesome controls |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, I want a **consistent application shell** with **light and dark themes**, a **collapsible left navigation** opened from a **burger** control, **clear branding** (**title** and **logo**), and **buttons that use Font Awesome icons**, so that the product is **readable**, **navigable on small screens**, and **visually coherent** across device classes.

**Scope**

- In scope: **Two** UI **themes**—**light** and **dark**—that the user can rely on for **page-level** chrome and primary content areas (exact component scope per architecture, but **global** theme semantics below apply). **Light** theme: **white** (or **white-equivalent**) **background** for the main shell and **dark** **foreground** text for primary reading. **Dark** theme: **dark grey** **background** for the main shell and **white** **foreground** text for primary reading. **Primary navigation** lives in a **left** **column** **panel** that **collapses** and **expands**; a **burger** (**hamburger**) **button** **toggles** that panel. The **application title** shown in the shell MUST be the exact string **`Domestic Light & Magic`**. The **application logo** MUST be the **Font Awesome** **classic** **regular** **lightbulb** icon as catalogued at [Font Awesome: lightbulb (classic, regular)](https://fontawesome.com/icons/lightbulb?f=classic&s=regular) (e.g. **`fa-regular fa-lightbulb`** in Font Awesome’s conventional class naming, or **equivalent** for the **same** **style** and **glyph** in the **version** the project adopts). **Buttons** in the UI (see business rules) MUST **display** a **Font Awesome** icon as part of their **visible** affordance (icons MAY pair with text labels where clarity requires it). **Licensing** and **delivery** of Font Awesome assets (Kit, npm package, or **embedded** SVG per architecture) MUST comply with **Font Awesome** **license** **terms** for the chosen **tier**.
- Out of scope: **Per-control** custom colours that **break** theme **contrast** **intentionally**; **non-button** **decorative** **illustrations** **outside** the **logo** and **button** **icon** rules unless a later requirement adds them; **animation** of the **burger** icon beyond **toggle** **state** **feedback** unless architecture specifies it.

**Business rules**

1. **Default** **vs** **override:** On **first** **load** **or** **until** **the** **user** **sets** **an** **explicit** **theme** **preference**, the **application** **MUST** **follow** **the** **user**’s **system** **(or** **user** **agent**)** **default** **for** **light** **vs** **dark** **appearance**—e.g. **`prefers-color-scheme`** **where** **the** **platform** **exposes** **it**. **After** **the** **user** **manually** **chooses** **light** **or** **dark**, **that** **choice** **MUST** **persist** **across** **sessions** **per** **architecture** **and** **override** **the** **system** **signal** **until** **the** **user** **changes** **it** **again** **(or** **a** **documented** **reset** **path** **exists**)**. **Where** **no** **system** **scheme** **is** **available**, **architecture** **defines** **a** **fallback** **(e.g.** **light**)**. The **product** **MUST** **still** **expose** **both** **themes** **and** **a** **discoverable** **way** **to** **switch** **between** **them** (**REQ-002**); **manual** **switching** **MUST** **not** **depend** **on** **hover-only** **affordances**.
2. In **light** theme, the **main** **application** **background** (shell / page canvas behind primary content) MUST be **white** or **substantially** **white**, and **primary** **body**/**chrome** **text** MUST be **dark** so **normal** reading meets **legible** contrast against that background (exact tokens in architecture).
3. In **dark** theme, the **main** **application** **background** MUST be **dark** **grey** (**not** **pure** **black** as the **only** mandated colour—architecture picks a specific **grey**), and **primary** **body**/**chrome** **text** MUST be **white** or **near-white** for **legible** contrast.
4. **Primary** **navigation** MUST be presented in a **left** **side** **menu** **region** that **collapses** to **hidden** or **icon-only**/**minimal** form and **expands** to show **navigation** **labels** **and** **destinations** per architecture; a **burger** **button** MUST **toggle** **collapse**/**expand** and MUST remain **usable** on **touch** **devices** (**REQ-002**).
5. The **visible** **application** **title** in the shell (e.g. header or adjacent branding) MUST read **`Domestic Light & Magic`** exactly (including **`&`** and spacing).
6. The **application** **logo** MUST use the **Font Awesome** **classic** **regular** **lightbulb** icon identified at [fontawesome.com/icons/lightbulb?f=classic&s=regular](https://fontawesome.com/icons/lightbulb?f=classic&s=regular); it MUST **read** as the **product** **mark** **alongside** the **title** per architecture (size and placement responsive per **REQ-002**).
7. **Buttons** (**`<button>`** **elements** and **button-styled** **controls** used for **actions**—e.g. **submit**, **cancel**, **delete**, **nav** **affordances** **styled** as **buttons**, **toolbar** **actions**) MUST include a **Font Awesome** icon in their **visible** UI (same **family**/**kit** as the **logo** unless architecture documents **why** a **split** is **impossible**). **Pure** **text** **links** **without** **button** **styling**, **text** **inputs**, and **native** **file** **inputs** are **exempt**; **icon-only** **buttons** are **allowed** if **accessible** **names** are **provided** (**REQ-002**).
8. Theme **colours**, **collapsible** **nav**, **branding**, and **button** **icons** MUST remain **usable** on **mobile**, **tablet**, and **desktop** (**REQ-002**): **no** **hover-only** **requirement** to **open** the **menu** or **switch** **theme**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Burger** **control** **reachable** **without** **horizontal** **scroll**; **expanded** **menu** **does** **not** **trap** **focus** **without** **a** **close** **path**; **theme** **toggle** **reachable**.
- Tablet: **Same** as **mobile**; **orientation** **changes** **keep** **branding** **and** **burger** **discoverable**.
- Desktop: **Collapsible** **nav** **may** **default** **expanded** **at** **wide** **widths** per **architecture** but **burger** **still** **collapses**/**expands**; **title** **and** **logo** **remain** **visible** in **header** **or** **equivalent**.

**Dependencies**

- REQ-001, REQ-002

**Open questions**

- **Exact** **hex** or **design-token** values for **dark** **grey** **background** and **dark** **text** in **light** **mode**.
- Whether **every** **destructive** or **primary** **modal** **action** **must** **reuse** the **same** **icon** **rules** when **implemented** as **links** **styled** **as** **buttons** vs **native** **buttons** (clarify **“button”** **scope** during **architecture**).

---

### REQ-019 — Three.js model and scene views: fixed dark-grey viewport

| Field | Value |
|-------|-------|
| **ID** | REQ-019 |
| **Title** | Three.js model and scene views: fixed dark-grey viewport |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user **viewing** a **model** or a **scene** in **three.js**, I want the **3D** **viewport** **always** **shown** **on** **a** **dark** **grey** **backdrop**, **whether** **the** **rest** **of** **the** **app** **is** **in** **light** **or** **dark** **mode**, so that **lights**, **wire** **segments**, and **labels** **stay** **easy** **to** **see**.

**Scope**

- In scope: The **WebGL** **rendering** **surface** **(scene** **background** **and** **any** **architectural** **match** **for** **margins** **/** **letterboxing** **inside** **the** **same** **visual** **frame** **as** **the** **canvas**)** **for** **(a)** **single-model** **detail** **(**REQ-010**)** **and** **(b)** **scene** **composite** **detail** **(**REQ-015**)** **MUST** **use** **a** **dark** **grey** **(not** **white** **or** **near-white**)** **regardless** **of** **whether** **REQ-018** **shell** **theme** **is** **light** **or** **dark**. **The** **same** **dark-grey** **policy** **applies** **in** **both** **UI** **themes** **so** **contrast** **with** **REQ-010**/**012** **sphere** **and** **segment** **colours** **remains** **stable**. **Exact** **colour** **or** **design** **token** **is** **deferred** **to** **architecture** **(must** **read** **clearly** **as** **dark** **grey** **to** **users**)**.
- Out of scope: **Changing** **REQ-010**/**012** **segment** **or** **default** **sphere** **colours** **unless** **architecture** **shows** **a** **contrast** **defect** **on** **the** **chosen** **grey**; **non-three.js** **UI** **panels** **(tables,** **forms**)** **which** **follow** **REQ-018** **shell** **tokens**; **export** **of** **screenshots**.

**Business rules**

1. **Model** **detail** **three.js** **(**REQ-010**)** **and** **scene** **detail** **three.js** **(**REQ-015**)** **MUST** **clear** **/** **fill** **the** **3D** **view** **with** **a** **dark** **grey** **background** **in** **light** **shell** **mode** **and** **in** **dark** **shell** **mode** **(**REQ-018**)**.
2. **Grid** **helpers**, **axes**, **or** **other** **non-data** **scene** **chrome** **in** **those** **views** **MUST** **remain** **subordinate** **to** **the** **lights** **and** **wires** **(**REQ-010**/** **REQ-012**)**; **architecture** **tunes** **helper** **contrast** **against** **the** **fixed** **dark-grey** **viewport**.
3. **REQ-002** **still** **requires** **the** **3D** **viewport** **and** **its** **controls** **(e.g.** **orbit,** **reset** **camera**)** **to** **be** **usable** **on** **mobile,** **tablet,** **and** **desktop** **without** **hover-only** **essential** **steps**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Canvas** **area** **reads** **as** **dark** **grey**; **touch** **orbit** **and** **pick** **behaviour** **unchanged** **from** **REQ-010**/**015**.
- Tablet: **Same** **as** **mobile**; **orientation** **changes** **do** **not** **switch** **the** **viewport** **to** **a** **light** **background**.
- Desktop: **Same** **policy**; **shell** **chrome** **may** **still** **toggle** **light**/**dark** **per** **REQ-018** **without** **changing** **the** **three.js** **backdrop** **policy**.

**Dependencies**

- REQ-002, REQ-010, REQ-015, REQ-018

**Open questions**

- **Whether** **the** **fixed** **three.js** **grey** **should** **match** **a** **single** **token** **shared** **with** **REQ-018** **dark** **shell** **or** **a** **dedicated** **“viz** **grey”** **token**.

---

### REQ-020 — Scene spatial API for dimensions, filtered retrieval, and bulk updates

| Field | Value |
|-------|-------|
| **ID** | REQ-020 |
| **Title** | Scene spatial API for dimensions, filtered retrieval, and bulk updates |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As a user or integrator, I want a scene-level API that returns scene dimensions, returns scene-composed light coordinates, filters lights by cuboid or sphere, and bulk-updates lights inside those regions, so that I can query and control lights in scene space without manually translating model coordinates.

**Scope**

- In scope: Scene API operations to retrieve a scene's dimensions, retrieve all lights currently in the scene, retrieve only lights inside a cuboid filter, retrieve only lights inside a sphere filter, bulk update lights inside a cuboid, and bulk update lights inside a sphere. For all scene-light retrieval and region filtering, coordinates are scene-composed positions (model coordinates plus scene placement), not original model coordinates.
- Out of scope: Defining a non-REST transport; changing model canonical coordinates in storage; rotation/scale transforms beyond scene placement offsets; shape types beyond cuboid and sphere.

**Business rules**

1. The scene API MUST expose a read operation that returns the scene dimensions used for scene-space queries (exact resource path and units formatting deferred to architecture, but numeric meaning MUST be unambiguous).
2. The scene API MUST expose a read operation that returns all lights in the scene with coordinates expressed in scene space, not in original model-local coordinates.
3. The scene API MUST expose a read operation that returns only lights within a caller-provided cuboid defined by a position and dimensions in scene space.
4. The scene API MUST expose a read operation that returns only lights within a caller-provided sphere defined in scene space.
5. The scene API MUST expose a bulk update operation that applies a requested light-state update to all lights within a caller-provided cuboid in scene space.
6. The scene API MUST expose a bulk update operation that applies a requested light-state update to all lights within a caller-provided sphere in scene space.
7. Bulk update payload semantics for light state MUST be consistent with REQ-011 (on/off, canonical hex colour, brightness percentage, and validation behavior).
8. Scene-region query and bulk-update operations MUST compute inclusion against scene-space coordinates derived from model coordinates plus scene placement (REQ-015), and MUST NOT rewrite canonical stored model coordinates (REQ-005).
9. For region-based operations, invalid geometry input (for example non-finite numbers or non-positive dimensions/radius) MUST be rejected with clear, actionable errors and MUST NOT partially apply updates.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A for API-only requirement; any UI consuming these APIs MUST remain usable per REQ-002.
- Tablet: N/A
- Desktop: N/A

**Dependencies**

- REQ-005, REQ-011, REQ-015

**Open questions**

- Inclusion boundary policy for region filters (inclusive vs exclusive for points exactly on cuboid faces or sphere surface).
- Whether scene dimensions are axis-aligned extents only or include explicit origin metadata in API responses.

---

### REQ-021 — Scenes: routines (definitions, run/stop via scene API, volumetric rules, first random-colour cycle type)

| Field | Value |
|-------|-------|
| **ID** | REQ-021 |
| **Title** | Scenes: routines (definitions, run/stop via scene API, volumetric rules, first random-colour cycle type) |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As a user or integrator, I want **routines** I can **create**, **list**, and **delete**, and **run against a scene** so they **change light state** according to **type-specific rules** using the **scene API**, so that I can automate effects—often limited to lights inside a **volumetric region** at a **position in scene space**—without hand-editing each light.

**Scope**

- In scope: A **routine definition** persisted with **name**, **description**, and **type** (type identifies behavior; additional **parameters** for a type are deferred to architecture unless specified below). User or API flows to **create**, **list**, and **delete** definitions. **Start** a routine **against a chosen scene** and **stop** a running instance **without** deleting the definition. While **running**, the implementation MUST apply state changes **only** through **scene-level** operations consistent with **REQ-020** (and **REQ-011** field semantics), using **scene-space** geometry for any **region-scoped** updates; canonical stored model coordinates MUST NOT be rewritten (**REQ-015**). **Routine rules** in general MAY restrict which lights are affected using **cuboid** or **sphere** volumes in **scene space** (**REQ-020** shapes and composition rules). The **first** concrete routine type (**see business rules**) affects **all lights in the scene** and **does not** require a volumetric sub-region parameter. **Stopping** ends automation; **per-light state** remains whatever was last successfully written (**REQ-011**).
- Out of scope: Editing routine definitions after create (unless implemented as delete-and-recreate); physical hardware protocols; new volumetric shapes beyond **REQ-020**; routines that **move** or **reorder** lights in **REQ-005** space; authentication policy.

**Business rules**

1. The product MUST support **creating** a routine definition with **name**, **description**, and **type**; **listing** all routine definitions; and **deleting** a definition. **Name** and **type** are **required** at create time; **description** MAY be empty if architecture allows, but the field MUST exist.
2. The product MUST support **starting** a routine **run** **scoped to exactly one scene** that exists at start time, and **stopping** an active run for that routine–scene pair (or **run identifier**—exact API deferred to architecture). **Start** MUST fail with **clear, actionable** errors if the scene does not exist or is not usable.
3. While a routine **run** is **active**, any change to **on/off**, **hex colour**, or **brightness** for lights in that scene MUST be performed **via the scene API surface** defined under **REQ-020** (for example bulk updates by region or whole-scene operations as architecture maps them), not by APIs that **only** target a model in isolation **when** the operation is **scene-contextual** automation. (Direct per-light model APIs **REQ-011** remain valid for **manual** user or integrator actions outside this requirement.)
4. **Volumetric targeting (general):** For routine types that **limit** effects to a **subset** of lights, inclusion MUST be evaluated against **scene-space** positions (**REQ-015**/**REQ-020**), using **cuboid** and/or **sphere** parameters supplied when **starting** the run or fixed by the type per architecture. **Invalid** region geometry MUST be rejected per **REQ-020** error expectations.
5. **First routine type** (type identifier **deferred to architecture**, but behavior is normative): **“Random colour cycle — all scene lights.”** When this type is **started** on a scene: (**a**) **every** light in that scene MUST be set **on** (`on` **true**), **brightness** **100%**, with **hex colour** valid per **REQ-011**; (**b**) thereafter, at **most once per elapsed SI second** while the run remains **active**, the system MUST assign **each** light in the scene a **new** **hex colour** chosen **independently** and **uniformly at random** from the set of **REQ-011**-valid colours (same canonical form and rejection rules as **REQ-011**). **Successive** seconds’ colours need not be distinct from prior values for a given light (true randomness allows repeats). The **approximate** **one-second** cadence MUST be **documented** in **docs/architecture.md** (server-driven timer vs client coordination, drift bounds).
6. **Stopping** this first type MUST **cease** further automated updates **promptly**; lights **retain** the **last** successfully **persisted** state (**REQ-011**). **Stopping** MUST **not** by itself **reset** lights to **REQ-014** defaults unless the user explicitly invokes **reset** elsewhere.
7. **Concurrency and deletion:** If **multiple** runs could conflict, architecture MUST define **allowed** concurrency (e.g. **one** active run per scene) and **deterministic** error or **queue** behavior (**open questions**). Deleting a routine definition **while** a run is active MUST either be **disallowed** with clear feedback or MUST **implicitly stop** the run first—architecture picks **one** policy and documents it.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: If the UI exposes routines, **list**, **create**, **delete**, **start**, and **stop** MUST be usable without **hover-only** essential steps (**REQ-002**); labels identify **scene** and **routine** clearly.
- Tablet: Same as mobile; orientation changes do not hide primary actions.
- Desktop: Efficient access to routine list and run controls alongside scene workflows.

**Dependencies**

- REQ-002 (when UI surfaces routines), REQ-011, REQ-015, REQ-020

**Open questions**

- Stable **machine-readable** **type** string for the first routine vs display name.
- Whether **more than one** routine run may be **active** on the **same** scene **simultaneously**.
- Whether **deleting** a definition **while running** **cancels** the run or is **blocked**.

---

### REQ-022 — Python scene routines: in-browser editor, scene library, documentation, and run loop with forced stop

| Field | Value |
|-------|-------|
| **ID** | REQ-022 |
| **Title** | Python scene routines: in-browser editor, scene library, documentation, and run loop with forced stop |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user who is **new to Python**, I want to **author**, **save**, **load**, **duplicate**, and **delete** **Python routines** for a **scene** using an **in-browser** editor with **syntax highlighting**, **checking**, **completion**, and **formatting**, with a **documented** **easy** **scene** **API** on the **same** **page**, so that I can **automate** **lights** in **scene** **space** **without** **mastering** **raw** **HTTP**; when I **run** a routine **against** a **scene**, it should **loop** **continuously** **until** I **stop** it, and **stopping** MUST be able to **forcibly** **terminate** the routine if **needed**.

**Scope**

- In scope: A **user-facing** experience (same **page** or **clear** **single** **workflow**) that includes: (**1**) an **in-browser** **Python** **code** **editor** **implemented** **with** **CodeMirror** **6** (**major** **version** **6** **of** **CodeMirror** **and** **the** **`@codemirror/*`** **packages**) **with** **full** **syntax** **highlighting**; (**2**) **syntax** **and** **static** **checking** that surfaces **issues** in the **editor** (e.g. **diagnostics** **before** or **during** **edit**, **exact** **mechanism** per **architecture**); (**3**) **code** **completion** (**autocomplete** **appropriate** **for** **Python** **and** **the** **provided** **scene** **API**); (**4**) **automatic** **code** **formatting** **on** **user** **request** or **on** **save** (**exact** **trigger** per **architecture**, but **formatting** MUST **be** **available** **and** **enabled** **by** **default** **where** **the** **product** **supports** **it**); (**5**) **persistence** and **management** of **Python** **routine** **definitions**: **save**, **load** **(open** **existing)**, **duplicate**, and **delete**; (**6**) a **Python** **library** **or** **module** **surface** **supplied** **by** **the** **application** that **wraps** **scene-oriented** **capabilities** (**REQ-020** **semantics**, **REQ-015** **scene** **composition**) **with** **simple** **methods** **and** **attributes**—**illustrative** **examples** **include** **`scene.height`** (or **equivalent** **documented** **name** **for** **scene** **vertical** **extent**) **and** **`scene.getLightsWithinSphere`** (**or** **equivalent** **documented** **name** **for** **sphere** **filtered** **retrieval** **in** **scene** **space**); (**7**) **on** **the** **same** **page** **as** **the** **editor**, **reference** **documentation** **for** **that** **library** (**signatures**, **parameters**, **return** **shapes** **at** **a** **novice-friendly** **level**, **short** **examples**, **and** **plain-language** **explanations**) **because** **the** **primary** **audience** **is** **Python** **novices**; (**8**) **running** a **saved** **Python** **routine** **against** a **chosen** **scene** **such** **that** **while** **the** **run** **is** **active**, **the** **implementation** **repeatedly** **executes** **the** **user’s** **script** **in** **a** **loop** (**repeated** **execution** **of** **the** **routine** **body** **or** **documented** **equivalent** **loop** **semantics** **in** **`docs/architecture.md`**); (**9**) **stopping** an **active** **Python** **routine** **run** **MUST** **end** **the** **loop** **promptly** **in** **the** **common** **case** **and** **MUST** **support** **forcible** **termination** (**e.g.** **timeout**, **interrupt**, **sandbox** **kill**, **or** **worker** **cancellation**—**exact** **means** **per** **architecture**) **when** **the** **routine** **does** **not** **yield** **to** **a** **cooperative** **stop** (**infinite** **loop**, **hang**, **or** **overlong** **iteration**).
- Out of scope: **Replacing** **REQ-021** **built-in** **non-Python** **routine** **types** (**coexistence** **is** **in** **scope**); **arbitrary** **pip** **packages** **unless** **explicitly** **added** **later**; **editing** **model** **CSV** **or** **scene** **geometry** **from** **Python**; **authentication** **policy**; **remote** **IDE** **integration**.

**Business rules**

1. The **application** MUST provide **save**, **load** **(select** **and** **open** **an** **existing** **definition)**, **duplicate**, and **delete** **for** **Python** **routine** **definitions** **with** **clear** **labels** **or** **icons** **per** **REQ-018** **where** **they** **are** **buttons**.
2. The **application** MUST **use** **CodeMirror** **6** **for** **Python** **editing** **in** **the** **browser** **on** **the** **REQ-022** **authoring** **surface** **(the** **editable** **code** **buffer** **and** **its** **extensions** **such** **as** **linting** **or** **completion** **MUST** **be** **built** **on** **the** **`@codemirror/*`** **version** **6** **ecosystem**)**. The **editor** MUST **provide** **Python** **syntax** **highlighting** **across** **that** **buffer**.
3. The **editor** MUST provide **checking** that **surfaces** **syntax** **or** **static** **issues** **to** **the** **user** **(e.g.** **underline** **or** **problem** **panel)** **without** **requiring** **a** **separate** **desktop** **tool**.
4. **Code** **completion** **and** **auto-formatting** (**format** **document** **or** **format** **on** **save**) MUST **be** **enabled** **for** **the** **Python** **editor** **(user** **MAY** **disable** **per** **architecture** **if** **a** **setting** **exists**, **but** **defaults** **MUST** **favor** **novices**: **completion** **and** **formatting** **on** **by** **default** **where** **technically** **feasible**).
5. The **product** MUST **expose** **a** **documented** **Python** **API** **object** **(e.g.** **`scene`**) **bound** **to** **the** **currently** **selected** **scene** **during** **a** **run** **that** **maps** **to** **scene** **capabilities** **consistent** **with** **REQ-020** **(dimensions,** **queries,** **bulk** **updates** **in** **scene** **space)** **and** **REQ-011** **field** **semantics** **for** **light** **state**; **exact** **method** **names** **MAY** **differ** **from** **the** **examples** **in** **the** **user** **story** **if** **`docs/architecture.md`** **and** **the** **on-page** **reference** **list** **the** **canonical** **names**.
6. **Reference** **documentation** **for** **the** **Python** **scene** **library** MUST **appear** **on** **the** **same** **page** **as** **the** **editor** **(e.g.** **collapsible** **panel,** **tabs,** **or** **split** **layout)** **and** MUST **target** **novices**: **plain** **language**, **parameter** **descriptions**, **at** **least** **one** **short** **example** **per** **major** **operation**, **and** **cross-links** **or** **anchors** **between** **editor** **and** **docs** **where** **helpful**.
7. **Starting** a **Python** **routine** **against** a **scene** MUST **execute** **the** **script** **in** **a** **continuous** **loop** **while** **the** **run** **remains** **active** (**architecture** **documents** **iteration** **timing**, **whether** **sleeps** **are** **implicit**, **and** **fairness** **with** **other** **runs**).
8. **Stopping** a **Python** **routine** **run** MUST **cease** **further** **loop** **iterations** **promptly** **under** **normal** **conditions**; **the** **implementation** MUST **also** **support** **forcible** **termination** **when** **the** **routine** **does** **not** **respond** **to** **cooperative** **stop** **within** **architecture-defined** **bounds** (**documented** **in** **`docs/architecture.md`**).
9. **Python** **routine** **automation** **MUST** **affect** **lights** **only** **through** **the** **documented** **scene** **API** **surface** **(wrapping** **REQ-020**/**REQ-011** **semantics)**; **it** MUST **not** **rewrite** **canonical** **stored** **model** **coordinates** (**REQ-005**, **REQ-015**).
10. **REQ-002** **applies**: **editor**, **documentation**, **and** **run**/**stop** **controls** MUST **remain** **usable** **on** **mobile**, **tablet**, **and** **desktop** **without** **hover-only** **essential** **steps** (**layout** **MAY** **stack** **editor** **and** **docs** **on** **narrow** **viewports**).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Editor** **and** **documentation** **reachable** **without** **losing** **context** **(e.g.** **tabbed** **docs** **or** **expandable** **sections)**; **save**/**load**/**duplicate**/**delete**/**run**/**stop** **touch-friendly**; **diagnostics** **readable**.
- Tablet: **Same** **as** **mobile**; **optional** **side-by-side** **editor** **and** **docs** **when** **width** **allows**.
- Desktop: **Split** **or** **multi-pane** **layout** **encouraged** **so** **reference** **docs** **stay** **visible** **alongside** **code**; **keyboard** **shortcuts** **for** **format**/**save** **where** **architecture** **provides** **them**.

**Dependencies**

- REQ-002, REQ-011, REQ-015, REQ-020, REQ-021 (coexistence with other routine mechanisms), REQ-018 (buttons/icons where applicable)

**Open questions**

- Whether **Python** **routines** **are** **stored** **as** **a** **new** **REQ-021** **routine** **type** **or** **a** **parallel** **entity** **with** **its** **own** **run** **lifecycle** **(architecture** **must** **choose** **one** **and** **document** **API** **URLs** **or** **UI** **entry** **points**).
- **Execution** **placement** (**in-browser** **e.g.** **WebAssembly** **vs** **server-side** **sandbox**) **and** **resource** **limits** (**CPU**, **memory**, **wall-clock** **per** **iteration**).
- **Interaction** **with** **concurrent** **REQ-021** **runs** **on** **the** **same** **scene** **(allowed**, **serialized**, **or** **rejected** **with** **clear** **errors**).

---

### REQ-023 — Create routine: single type dropdown (includes Python)

| Field | Value |
|-------|-------|
| **ID** | REQ-023 |
| **Title** | Create routine: single type dropdown (includes Python) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user **creating** a **new** **routine**, I want to **choose** the **kind** of **routine** (**built-in** **type** **or** **Python**) **from** **one** **dropdown**, so that **there** **is** **a** **single**, **predictable** **create** **path** **and** **I** **do** **not** **rely** **on** **a** **separate** **button** **only** **for** **Python**.

**Scope**

- In scope: The **user-facing** **flow** **to** **start** **creating** **a** **new** **routine** **definition** (**REQ-021**, **REQ-022**) **MUST** **include** **a** **labeled** **routine** **type** **selector** **implemented** **as** **a** **dropdown** (**native** **`<select>`** **or** **an** **accessible** **custom** **control** **with** **the** **same** **semantics** **per** **architecture**). **Every** **routine** **category** **the** **user** **may** **newly** **create** **MUST** **appear** **as** **an** **option** **in** **that** **dropdown**, **including** **each** **built-in** **routine** **type** **from** **REQ-021** **and** **the** **Python** **routine** **path** **from** **REQ-022** (**one** **option** **per** **distinct** **create** **kind**, **with** **clear** **human-readable** **labels**). **Choosing** **Python** **MUST** **continue** **into** **the** **Python** **authoring** **experience** **(**REQ-022**)**; **choosing** **a** **built-in** **type** **MUST** **follow** **REQ-021** **create** **semantics** **for** **that** **type**. The **product** **MUST** **not** **require** **a** **separate** **primary** **“create** **Python** **routine”** **(or** **equivalent)** **button** **or** **entry** **that** **is** **the** **only** **way** **to** **start** **a** **new** **Python** **routine**—**Python** **must** **be** **reachable** **via** **the** **same** **type** **dropdown** **as** **other** **kinds**.
- Out of scope: **Changing** **REQ-021**/**022** **persistence** **or** **API** **contracts** **except** **where** **the** **UI** **must** **bind** **to** **them**; **editing** **existing** **definitions**; **ordering** **or** **grouping** **of** **options** **beyond** **clarity** **(**architecture** **may** **use** **optgroups** **or** **separators**)**.

**Business rules**

1. **New** **routine** **creation** **UI** **MUST** **expose** **a** **dropdown** **(or** **semantically** **equivalent** **single-choice** **list)** **for** **routine** **type** **before** **the** **user** **can** **complete** **creation** **(**exact** **placement** **in** **the** **wizard**/**form** **per** **architecture**)**.
2. **The** **dropdown** **options** **MUST** **collectively** **cover** **all** **creatable** **routine** **kinds**: **built-in** **types** **defined** **under** **REQ-021** **and** **the** **Python** **routine** **kind** **under** **REQ-022** **(**no** **creatable** **kind** **omitted** **from** **the** **list**)**.
3. **The** **UI** **MUST** **not** **present** **a** **standalone** **primary** **action** **whose** **sole** **purpose** **is** **to** **start** **creating** **only** **a** **Python** **routine** **(**e.g.** **a** **second** **“Python”** **button** **next** **to** **“New** **routine”**)** **when** **Python** **is** **already** **selectable** **in** **the** **type** **dropdown** **(**secondary** **shortcuts** **or** **deep** **links** **are** **out** **of** **scope** **unless** **they** **do** **not** **replace** **the** **dropdown** **path**)**.
4. **REQ-002** **applies**: **the** **type** **control** **MUST** **be** **operable** **on** **mobile**, **tablet**, **and** **desktop** **without** **hover-only** **essential** **steps** **(**native** **select** **typically** **satisfies** **touch** **and** **keyboard** **when** **labeled** **and** **focusable**)**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Type** **dropdown** **full** **width** **or** **touch-friendly** **hit** **target**; **label** **visible** **(**e.g.** **“Routine** **type”** **or** **equivalent**)**.
- Tablet: **Same** **as** **mobile**; **orientation** **changes** **do** **not** **hide** **the** **control**.
- Desktop: **Dropdown** **integrated** **with** **other** **create** **fields**; **keyboard** **navigation** **to** **open** **and** **change** **selection** **where** **the** **platform** **allows**.

**Dependencies**

- REQ-002, REQ-021, REQ-022

**Open questions**

- **Exact** **display** **labels** **and** **optional** **grouping** **(**optgroups**)** **for** **options**.
- **Whether** **a** **non-primary** **shortcut** **(**e.g.** **nav** **link** **to** **Python** **editor**)** **remains** **for** **discoverability** **without** **violating** **rule** **3**.

---

### REQ-024 — Python routine view: bottom API reference with descriptions and snippets

| Field | Value |
|-------|-------|
| **ID** | REQ-024 |
| **Title** | Python routine view: bottom API reference with descriptions and snippets |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user authoring **Python** **routines**, I want a **single reference list** of **every** **callable** **operation** on the **documented** **Python** **scene** **API**, each with a **short** **description** and a **sample** **code** **snippet**, placed **at** **the** **bottom** **of** **the** **Python** **routine** **page**, so that I can **look** **up** **usage** **without** **leaving** **the** **authoring** **context**.

**Scope**

- In scope: On the **Python** **routine** **authoring** **surface** (**REQ-022**), a **dedicated** **reference** **section** **ordered** **after** **primary** **content** **above** **the** **fold** **in** **document** **flow**—i.e. **at** **the** **bottom** **of** **the** **page** **(after** **the** **editor**, **run**/**stop**, **scene** **selection** **for** **runs**, **and** **other** **primary** **controls** **as** **architecture** **lays** **them** **out**)**—that **lists** **all** **public** **methods**, **functions**, **and** **documented** **attributes** **exposed** **on** **the** **Python** **scene** **binding** **(the** **`scene`** **object** **or** **equivalent)**. **Each** **entry** **MUST** **include** **a** **novice-friendly** **description** **and** **at** **least** **one** **copy-paste-friendly** **sample** **code** **snippet** **showing** **typical** **use** **(may** **assume** **`scene`** **is** **already** **bound** **during** **a** **run** **unless** **architecture** **documents** **otherwise)**.
- Out of scope: **Replacing** **REQ-022**’s **requirement** **for** **on-page** **reference** **elsewhere** **on** **the** **page** **(this** **REQ** **adds** **placement** **and** **completeness** **rules** **for** **the** **bottom** **section)**; **full** **OpenAPI** **HTTP** **catalog** **as** **the** **only** **form** **of** **docs** **unless** **architecture** **collapses** **Python** **to** **thin** **HTTP** **wrappers** **only**.

**Business rules**

1. The **Python** **routine** **view** **MUST** **include** **a** **reference** **section** **at** **the** **bottom** **of** **the** **page** **in** **vertical** **reading** **order** **(after** **primary** **authoring** **and** **run** **workflow** **regions)**.
2. The **list** **MUST** **enumerate** **every** **Python**-**exposed** **scene** **API** **surface** **element** **that** **the** **product** **supports** **for** **routines** **(no** **deliberate** **omission** **of** **a** **public** **operation** **from** **this** **catalog)**.
3. **Each** **catalog** **item** **MUST** **include** **a** **plain-language** **description** **and** **at** **least** **one** **sample** **code** **snippet** **illustrating** **its** **use**.
4. **REQ-002** **applies**: **the** **bottom** **section** **MUST** **remain** **readable** **and** **scrollable** **on** **mobile**, **tablet**, **and** **desktop** **without** **horizontal** **scrolling** **for** **primary** **snippet** **content** **where** **avoidable**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Snippets** **use** **horizontal** **scroll** **only** **if** **necessary**; **touch** **targets** **for** **any** **expand/collapse** **per** **entry** **meet** **REQ-002**.
- Tablet: **Same** **as** **mobile**; **bottom** **section** **remains** **discoverable** **after** **editor** **(e.g.** **heading** **or** **anchor)**.
- Desktop: **Snippets** **may** **use** **syntax** **styling** **consistent** **with** **the** **editor** **where** **feasible**; **anchor** **links** **from** **editor** **to** **API** **entries** **are** **optional** **but** **encouraged** **where** **helpful**.

**Dependencies**

- REQ-002, REQ-022

**Open questions**

- **Whether** **long** **snippets** **are** **collapsed** **by** **default** **on** **small** **viewports**.

---

### REQ-025 — Python routine default code: sphere interior colour changes

| Field | Value |
|-------|-------|
| **ID** | REQ-025 |
| **Title** | Python routine default code: sphere interior colour changes |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **novice** **author**, I want the **default** **or** **starter** **code** **shown** **when** **I** **begin** **a** **new** **Python** **routine** **to** **demonstrate** **changing** **the** **colours** **of** **lights** **that** **lie** **inside** **a** **sphere** **in** **scene** **space**, so that **I** **immediately** **see** **a** **realistic** **pattern** **for** **region-scoped** **updates**.

**Scope**

- In scope: **Initial** **editor** **buffer** **content** **for** **a** **newly** **created** **Python** **routine** **definition** **(and** **any** **product-provided** **“reset** **to** **default** **template”** **action** **if** **present)** **MUST** **be** **working** **example** **code** **that** **updates** **light** **state** **(**on**/**off**, **canonical** **hex** **colour**, **brightness** **per** **REQ-011**)** **for** **lights** **whose** **positions** **fall** **inside** **a** **caller-defined** **or** **documented** **sphere** **in** **scene** **space**, **using** **only** **the** **documented** **Python** **scene** **API** **(**REQ-022**, **REQ-020** **semantics**)**.
- Out of scope: **Overwriting** **user** **code** **on** **every** **load** **of** **saved** **definitions**; **mandating** **a** **specific** **sphere** **center**/**radius** **beyond** **“valid** **per** **REQ-020**”**.

**Business rules**

1. **New** **Python** **routine** **definitions** **MUST** **open** **with** **default** **template** **code** **whose** **primary** **illustrated** **behavior** **is** **changing** **colours** **(**and** **other** **REQ-011** **fields** **as** **needed** **for** **a** **coherent** **demo**)** **for** **lights** **inside** **a** **sphere** **region** **in** **scene** **space**.
2. The **template** **MUST** **use** **sphere** **filtering**/**targeting** **consistent** **with** **REQ-020** **(scene-space** **geometry**, **no** **canonical** **model** **coordinate** **rewrites**)**.
3. **REQ-022** **defaults** **for** **editor** **features** **(completion,** **formatting**)** **remain** **in** **force**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Default** **template** **is** **plain** **text** **in** **the** **editor**; **no** **extra** **UX** **beyond** **REQ-022**.
- Tablet: **Same** **as** **mobile**.
- Desktop: **Same** **as** **mobile**.

**Dependencies**

- REQ-011, REQ-020, REQ-022

**Open questions**

- **Whether** **the** **template** **should** **include** **explicit** **`sleep`**/**delay** **calls** **or** **rely** **on** **the** **runtime** **loop** **(**REQ-022**)** **only**.

---

### REQ-026 — Python scene binding: width, depth, and height

| Field | Value |
|-------|-------|
| **ID** | REQ-026 |
| **Title** | Python scene binding: width, depth, and height |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As an **author** **of** **Python** **routines**, I want the **`scene`** **object** **to** **expose** **width**, **depth**, **and** **height** **(or** **documented** **equivalents** **for** **all** **three** **axis-aligned** **extents**)**, so that **I** **can** **position** **effects** **and** **regions** **relative** **to** **the** **scene’s** **size**.

**Scope**

- In scope: The **Python** **binding** **for** **the** **active** **scene** **during** **a** **routine** **(**REQ-022**)** **MUST** **expose** **three** **numeric** **attributes** **corresponding** **to** **the** **scene’s** **axis-aligned** **spatial** **extents**: **vertical** **extent** **(**already** **illustrated** **as** **height** **in** **REQ-022**)** **and** **the** **two** **remaining** **orthogonal** **horizontal** **extents** **as** **`width`** **and** **`depth`** **(or** **architecture-chosen** **names** **documented** **as** **the** **canonical** **mapping** **to** **those** **extents**)**. **Values** **MUST** **align** **with** **REQ-020** **dimension** **semantics** **(SI** **meters** **unless** **architecture** **documents** **a** **single** **consistent** **unit** **for** **scene** **APIs**)**.
- Out of scope: **Adding** **new** **HTTP** **fields** **beyond** **what** **REQ-020** **already** **requires** **if** **the** **backend** **already** **carries** **sufficient** **data** **to** **derive** **all** **three** **extents**; **rotation** **or** **non-axis-aligned** **bounding** **primitives**.

**Business rules**

1. The **Python** **`scene`** **object** **MUST** **expose** **`height`** **and** **MUST** **also** **expose** **`width`** **and** **`depth`** **(or** **equivalent** **documented** **names** **for** **all** **three** **extents**)**.
2. **Reading** **these** **attributes** **MUST** **reflect** **the** **same** **numeric** **meaning** **as** **the** **scene** **dimension** **data** **defined** **under** **REQ-020** **(no** **contradictory** **values** **between** **REST** **and** **Python** **for** **the** **same** **scene** **snapshot**)**.
3. **`docs/architecture.md`**, **the** **REQ-022** **on-page** **reference**, **and** **the** **REQ-024** **bottom** **catalog** **MUST** **list** **all** **three** **attributes** **with** **definitions** **of** **which** **world** **axis** **each** **maps** **to**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **N/A** **for** **attribute** **surface** **alone**; **docs** **showing** **them** **follow** **REQ-002**/**REQ-024**.
- Tablet: **N/A**
- Desktop: **N/A**

**Dependencies**

- REQ-020, REQ-022, REQ-024

**Open questions**

- **Whether** **dimensions** **are** **evaluated** **from** **dynamic** **light** **bounds** **or** **fixed** **scene** **volume** **metadata** **(**REQ-015**/** **REQ-020**)**.

---

### REQ-027 — Python routine visual debug: selected scene, reset lights, reset camera

| Field | Value |
|-------|-------|
| **ID** | REQ-027 |
| **Title** | Python routine visual debug: selected scene, reset lights, reset camera |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **user** **debugging** **a** **Python** **routine**, I want **to** **demonstrate** **the** **routine** **against** **a** **scene** **I** **select** **with** **a** **live** **three.js** **visualization** **(visual** **debug**)**, **and** **I** **want** **buttons** **to** **reset** **the** **scene’s** **lights** **to** **a** **clean** **baseline** **and** **to** **reset** **the** **camera**, **so** **that** **I** **can** **iterate** **without** **manual** **cleanup** **or** **losing** **my** **bearing** **in** **3D**.

**Scope**

- In scope: On the **Python** **routine** **authoring** **experience** (**REQ-022**), **visual** **debug** **capabilities** **that** **let** **the** **user** **pick** **which** **persisted** **scene** (**REQ-015**)** **drives** **a** **dedicated** **or** **shared** **three.js** **viewport** **showing** **that** **scene** **while** **the** **routine** **runs** **or** **when** **previewing** **(**exact** **coupling** **to** **run** **lifecycle** **per** **architecture**)**. **The** **viewport** **MUST** **follow** **REQ-010**/**REQ-012**/**REQ-015**/**REQ-019** **visual** **rules** **for** **spheres**, **segments**, **and** **dark-grey** **background** **as** **applicable** **to** **scene** **composite** **views**. **Two** **controls** **(each** **a** **button** **or** **equivalent** **accessible** **action** **with** **REQ-018** **icon** **rules** **where** **they** **are** **buttons**)**:
  - **Reset** **scene** **lights**: **sets** **every** **light** **in** **the** **selected** **debug** **scene** **to** **the** **default** **per-light** **state** **in** **REQ-014** **(**off**, **100%** **brightness**, **`#FFFFFF`** **hex**)** **and** **persists** **per** **REQ-011**; **does** **not** **change** **scene** **membership**, **placements**, **or** **canonical** **model** **coordinates** (**REQ-005**, **REQ-015**)**.
  - **Reset** **camera**: **restores** **the** **default** **camera** **and** **control** **baseline** **for** **that** **viewport** **only**, **per** **REQ-016** **semantics** **(client-side** **navigation** **only**)**.
- Out of scope: **Persisting** **camera** **pose** **across** **sessions**; **factory** **reset** **(**REQ-017**)**; **deleting** **or** **creating** **scenes** **from** **this** **panel**.

**Business rules**

1. The **user** **MUST** **be** **able** **to** **select** **a** **scene** **(from** **existing** **scenes**)** **as** **the** **target** **for** **visual** **debug** **of** **the** **Python** **routine**.
2. While **visual** **debug** **is** **active** **with** **a** **chosen** **scene**, **light** **state** **changes** **from** **the** **routine** **MUST** **become** **visible** **in** **the** **viewport** **within** **the** **same** **class** **of** **timeliness** **as** **REQ-012** **(no** **indefinite** **staleness** **after** **successful** **writes**)**.
3. **Reset** **scene** **lights** **MUST** **apply** **REQ-014** **defaults** **to** **all** **lights** **in** **the** **selected** **scene** **and** **MUST** **persist** **them** **so** **other** **clients** **and** **reloads** **see** **the** **same** **result**.
4. **Reset** **camera** **MUST** **not** **alter** **persisted** **models**, **scenes**, **placements**, **or** **per-light** **state**; **it** **MUST** **only** **reset** **client** **navigation** **state** **for** **the** **debug** **viewport**.
5. **REQ-002** **and** **REQ-018** **apply** **to** **scene** **selection**, **debug** **viewport** **interactions**, **and** **the** **two** **reset** **actions**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Scene** **selector**, **debug** **viewport**, **and** **reset** **buttons** **remain** **reachable** **without** **hover-only** **steps**; **viewport** **supports** **touch** **navigation** **consistent** **with** **REQ-010**/**REQ-015**.
- Tablet: **Same** **as** **mobile**; **orientation** **changes** **keep** **reset** **actions** **visible** **in** **the** **primary** **debug** **region** **or** **overflow** **menu** **only** **if** **equally** **discoverable**.
- Desktop: **Reset** **controls** **sit** **adjacent** **to** **or** **within** **the** **debug** **viewport** **chrome** **per** **architecture**.

**Dependencies**

- REQ-002, REQ-010, REQ-011, REQ-012, REQ-014, REQ-015, REQ-016, REQ-018, REQ-019, REQ-022

**Open questions**

- **Whether** **visual** **debug** **requires** **an** **active** **run** **or** **may** **show** **a** **static** **post-run** **state**.
- **Whether** **reset** **scene** **lights** **also** **stops** **an** **active** **run** **or** **leaves** **run** **state** **unchanged** **(**architecture** **must** **define** **to** **avoid** **surprise** **re-application** **of** **routine** **effects**)**.

---

### REQ-028 — Three.js light spheres: brightness-scaled emissive glow

| Field | Value |
|-------|-------|
| **ID** | REQ-028 |
| **Title** | Three.js light spheres: brightness-scaled emissive glow |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **user** **viewing** **a** **scene** **(**or** **a** **single** **model**)** **in** **three.js**, I want **each** **on** **light’s** **sphere** **to** **glow** **in** **proportion** **to** **its** **brightness** **and** **to** **read** **as** **emitting** **light**, **so** **that** **100%** **brightness** **looks** **clearly** **bright** **and** **dimmer** **settings** **look** **appropriately** **subdued**.

**Scope**

- In scope: **three.js** **rendering** **of** **light** **spheres** **where** **REQ-012** **already** **governs** **colour** **and** **on**/**off** **appearance** **(**single-model** **detail** **per** **REQ-010**, **scene** **composite** **detail** **per** **REQ-015**, **and** **Python** **routine** **visual** **debug** **per** **REQ-027** **where** **those** **spheres** **are** **shown**)**. **For** **lights** **in** **the** **on** **state**, **the** **material** **MUST** **use** **an** **emissive** **(**self-illuminated**)** **treatment** **so** **the** **sphere** **appears** **to** **emit** **light**, **not** **only** **a** **flat** **tinted** **surface**. **The** **strength** **of** **that** **glow** **(**or** **architecture-defined** **equivalent** **visual** **metric**)** **MUST** **scale** **monotonically** **with** **the** **persisted** **brightness** **percentage** **from** **REQ-011** **(**0** **through** **100**)** **so** **that** **100%** **reads** **as** **strong** **glow** **and** **lower** **values** **read** **weaker** **while** **remaining** **ordered** **(**no** **inversion** **where** **a** **lower** **percent** **looks** **brighter** **than** **a** **higher** **one** **for** **the** **same** **hex** **colour**)**. **REQ-019** **dark-grey** **viewport** **remains** **so** **glow** **is** **visible** **against** **the** **background**.
- Out of scope: **Mandating** **a** **specific** **three.js** **API** **(**e.g.** **`MeshStandardMaterial.emissive`** **vs** **post-processing**)**; **physical** **global** **illumination** **or** **accurate** **luminaire** **photometry**; **changing** **REQ-010** **2** **cm** **sphere** **diameter** **or** **wire** **segment** **styling** **except** **where** **architecture** **must** **reconcile** **glow** **with** **existing** **rules**.

**Business rules**

1. **On** **lights** **(**`on` **true** **per** **REQ-011**/**REQ-012**)** **MUST** **use** **a** **material** **that** **includes** **a** **clear** **emissive** **(**light-emitting**)** **component** **tied** **to** **the** **current** **hex** **colour** **so** **the** **sphere** **reads** **as** **a** **light** **source** **rather** **than** **a** **purely** **diffuse** **ball**.
2. **Emissive** **strength** **(**or** **documented** **equivalent**)** **MUST** **scale** **with** **brightness** **percentage** **from** **REQ-011**: **at** **100%** **brightness** **the** **glow** **MUST** **be** **visibly** **strong**; **at** **lower** **percents** **the** **glow** **MUST** **be** **weaker** **in** **a** **way** **users** **can** **perceive** **as** **dimmer** **light** **(**exact** **curve** **per** **architecture**, **subject** **to** **monotonicity** **in** **rule** **3**)**.
3. **For** **two** **on** **lights** **with** **the** **same** **hex** **colour** **and** **different** **brightness** **values**, **the** **higher** **brightness** **MUST** **not** **appear** **less** **glowing** **than** **the** **lower** **(**monotonic** **ordering** **of** **glow** **vs** **percent**)**.
4. **Off** **lights** **MUST** **keep** **REQ-012** **dim** **grey** **transparent** **appearance**; **their** **emissive** **contribution** **MUST** **remain** **negligible** **so** **they** **do** **not** **appear** **to** **glow** **like** **on** **lights** **or** **outshine** **REQ-010** **wire** **segments**.
5. **When** **per-light** **state** **updates** **(**REQ-011**, **REQ-012**)** **or** **the** **user** **navigates** **between** **views**, **glow** **MUST** **stay** **consistent** **with** **persisted** **state** **(**no** **indefinite** **staleness** **after** **successful** **writes**)** **for** **the** **same** **semantics** **as** **REQ-012** **colour**/**opacity** **updates**.
6. **`docs/architecture.md`** **MUST** **describe** **the** **chosen** **three.js** **material**/**rendering** **approach** **for** **REQ-028** **(**including** **how** **brightness** **maps** **to** **emissive** **intensity** **or** **equivalent**)** **and** **how** **it** **fits** **Pi**/**WebGL** **constraints** **from** **REQ-003** **where** **relevant**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Glow** **and** **sphere** **legibility** **remain** **acceptable** **on** **small** **WebGL** **viewports**; **touch** **hover-equivalent** **(**REQ-010**)** **unchanged**.
- Tablet: **Same** **as** **mobile**; **no** **hover-only** **dependency** **for** **perceiving** **brightness** **differences** **(**REQ-002**)**.
- Desktop: **User** **can** **compare** **brightness** **levels** **when** **multiple** **on** **lights** **are** **visible** **without** **excessive** **bloom** **that** **obscures** **ids** **(**architecture** **may** **cap** **or** **tone-map** **as** **needed**)**.

**Dependencies**

- REQ-002, REQ-003, REQ-010, REQ-011, REQ-012, REQ-015, REQ-019, REQ-027

**Open questions**

- **Whether** **a** **single** **global** **cap** **on** **emissive** **intensity** **is** **needed** **when** **many** **100%** **lights** **fill** **the** **viewport** **(**architecture** **trade-off** **vs** **REQ-010** **“do** **not** **merge** **lights”**)**.

---

### REQ-029 — High-throughput light state updates (batching, connections, push)

| Field | Value |
|-------|-------|
| **ID** | REQ-029 |
| **Title** | High-throughput light state updates (batching, connections, push) |
| **Priority** | Must |
| **Actor(s)** | Integrator; operator; end user (indirectly via UI) |

**User story**

As an integrator or operator driving dynamic lighting, I want the system to sustain frequent updates to many lights without being dominated by per-request overhead, so that scenes with hundreds of lights can change multiple times per second while viewers stay sufficiently up to date.

**Scope**

- In scope: Non-functional expectations for **persisted** per-light state (same fields and validation semantics as **REQ-011**) when **updating large sets** in **model** and **scene** contexts (**REQ-015**, **REQ-020**), and for **viewers** to remain sufficiently fresh relative to **REQ-012** when change rates are high. **Scale assumption** (design target): on the order of **hundreds** of lights (consistent with **REQ-005**’s upper bound) with **multiple** aggregate **update cycles per second** across writes and/or viewer refresh. **Illustrative** mechanisms the architecture **must consider and document** (not all mandatory in isolation): **HTTP/2** or other **connection reuse**, **client connection pooling** / keep-alive, **batch or bulk write** APIs (including **REQ-020** where scene-scoped), **Server-Sent Events**, **WebSocket**, or **similar server-push** for distributing state changes to connected clients. Solutions **must remain plausible** on **Raspberry Pi 4** constraints (**REQ-003**).
- Out of scope: Hard numeric SLOs unless raised in **open questions**; mandating one specific protocol when another meets the same goals; physical lighting protocols (e.g. DMX).

**Business rules**

1. **`docs/architecture.md` MUST** describe how the product meets high-throughput light updates at the stated scale: the **write path** (batching, use of **REQ-020** where applicable, persistence and transaction boundaries as relevant) and the **observer path** (how UIs and integrators obtain timely state—**server-push vs polling** with rationale).
2. Integrators **MUST NOT** be limited to **one HTTP request per light** as the **only** supported path for high-frequency multi-light changes: the product **MUST** expose **documented** **aggregate** update paths (for example **REQ-020** region bulk updates, and/or **model-scoped** batch operations if architecture defines them). **REQ-011** per-light read and write operations **remain** required for granular control.
3. **Connection reuse:** **`docs/architecture.md` SHOULD** document which **HTTP** features are enabled or assumed (for example **HTTP/2**, **HTTP/1.1 keep-alive**) and **client pooling** or equivalent guidance for integrators and for the shipped web UI.
4. For **low-latency** refresh or **many concurrent viewers**, **`docs/architecture.md` SHOULD** specify **server-push** (**SSE**, **WebSocket**, or **equivalent**) **or** justify **bounded polling** that still satisfies **REQ-012**-class timeliness under the stated load.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A at API level; any UI that displays live light state **MUST** remain usable per **REQ-002** at high update rates where the product targets them.
- Tablet: N/A; same as mobile where applicable.
- Desktop: N/A; same as mobile where applicable.

**Dependencies**

- REQ-003, REQ-011, REQ-012, REQ-015, REQ-020

**Open questions**

- Target **p99** latency or **updates per second** caps (if any) for the Pi deployment.
- Whether **model-scoped** batch writes are **required** in addition to **REQ-020** scene bulk for integrator workflows.
- How **multi-tab** or **multi-client** viewers stay consistent when push channels are used.

---
