# Requirements

This document captures the initial product requirements. IDs are immutable.

---

## REQ-001 — Go service and Next.js + Tailwind web UI

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

## REQ-002 — Responsive, reactive UI across device classes

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

## REQ-003 — Target deployment on Raspberry Pi 4 Model B

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

## REQ-004 — Single executable packaging (no Docker at this stage)

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
- Out of scope: **macOS** executable variants unless added by a later requirement (Windows and multiple Linux architectures are covered by REQ-043).

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

## REQ-005 — Wire light model (data shape, CSV, metadata)

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

## REQ-006 — List, view, delete, and create models via CSV upload

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

## REQ-007 — Validate CSV on model upload

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

## REQ-008 — Single command to build and run locally

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

## REQ-009 — Default sample models (sphere, cube, cone)

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

## REQ-010 — Three.js visualization when viewing a model

| Field | Value |
|-------|-------|
| **ID** | REQ-010 |
| **Title** | Three.js visualization when viewing a model |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user, when I **view** a model, I want **every** light drawn as a **2 cm white sphere**, each light linked to its **previous** and **next** along the wire (**REQ-005**) by **barely visible** **`#D0D0D0`** **light-grey** segments where those neighbors exist, and a way to see each light’s **index and coordinates**—so that I can interpret spatial layout and **wire order** without missing lights.

**Scope**

- In scope: On the **model view** / **detail** experience (the same flow as **viewing** a single model in **REQ-006**), the product MUST render the model’s light positions in **3D** using **three.js** (the JavaScript library from https://threejs.org/ , typically consumed as the `three` package). **Every** light returned for the model MUST be **drawn** (no deliberate omission, decimation, or level-of-detail that hides lights). **Each** light MUST appear as a **sphere** of **2 cm** diameter (**0.02 m**; **SI meters** per **REQ-005**). **Default** sphere appearance (no per-light state or **REQ-012** default) MUST be **white** with **solid** (opaque) **fill**. **Straight line segments** MUST connect **only** **consecutive** lights in **ascending id order**, consistent with **REQ-005** adjacency (one segment per pair **(i, i+1)** for **i = 0 … n − 2** when **n > 1**). Segments MUST use **`#D0D0D0`** (**canonical `#RRGGBB` light grey**, softer than mid-grey so the wire reads as **ambient** rather than **emphasis**) and **85% transparent**—i.e. **15% opaque** (e.g. alpha **0.15** in an **0–1** opacity scale)—and **thin** enough that they read as **subtle** guidance, **barely visible**, and **less prominent** than the light **spheres** (including **on** lights at **100%** brightness). **Pointer hover** over a light’s sphere MUST reveal that light’s **id** and **x**, **y**, **z**. **Touch** and **tablet** users MUST have an **equivalent** way to discover id and coordinates (**REQ-002**). The **model** **three.js** **view** MUST **also** **show** **the** **faint** **axis-aligned** **boundary** **cuboid** **per** **REQ-034** **(**tight** **AABB** **on** **light** **`x,y,z`** **plus** **fixed** **0.3** **m** **padding**)**.
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

- REQ-001, REQ-002, REQ-006, REQ-005, REQ-034

**Open questions**

- Minimum three.js **major** version or LTS alignment (if any).
- Whether **orbit controls** or another interaction model is mandatory vs architecture choice.
- Preferred **decimal precision** or **formatting** for displayed coordinates (e.g. fixed decimals vs full float).

---

## REQ-011 — REST API for per-light state (query and control)

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
5. Successful writes MUST update **authoritative** per-light state held **in memory** in the running service (**REQ-039**): API readers, routine automation, and connected visualization clients MUST observe the same logical state without requiring a process restart. The application MUST **not** store per-light operational state (**on**/**off**, colour, brightness) in **durable** application storage (e.g. SQLite) for later reload (**REQ-039**). When a model has an assigned device (**REQ-036**), successful writes MUST also drive the physical output per **REQ-038** subject to connectivity and device capability.
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

## REQ-012 — Model visualization reflects per-light state

| Field | Value |
|-------|-------|
| **ID** | REQ-012 |
| **Title** | Model visualization reflects per-light state |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user viewing a model, I want the **3D visualization** to **match** each light’s **stored state**, so that **colour**, **brightness**, and **on/off** are **immediately obvious** from the **spheres** (e.g. **blue** when set to blue, subtle **#D0D0D0** when **off**).

**Scope**

- In scope: When viewing a model (**REQ-006**, **REQ-010**), each light’s **sphere** (**0.02 m** diameter) MUST reflect **REQ-011** state: **on** = **filled** **opaque** appearance using the light’s **hex colour** modulated by **brightness**; **off** = appearance using **`#D0D0D0`** and **85% transparency** (**15% opacity**) as **REQ-010** wire segments, applied to the **sphere** (filled or thin shell—exact geometry deferred to architecture) so **off** lights remain **discernible** but **less prominent** than **on** lights. **Any** change to authoritative in-memory light state via the API (**REQ-011**, **REQ-039**) MUST be **reflected** in the visualization **without** requiring a full page reload if the client is already viewing that model (e.g. **poll**, **push**, or **same-session** refetch—architecture defines mechanism; the requirement is **timely** consistency from the user’s perspective).
- Out of scope: Changing **wire segment** styling based on **on/off** state unless added later; **export** of rendered images.

**Business rules**

1. For a light that is **on**, the **sphere** MUST appear **filled** (solid or equivalent **opaque** **surface** fill) and MUST use the **current** **hex colour** and **brightness** from **REQ-011** (architecture defines how **percentage** maps to rendered colour or intensity).
2. For a light that is **off**, the **sphere** MUST use **`#D0D0D0`** and **85% transparency** (**15% opacity**), **consistent** with **REQ-010** segment styling, so it reads as **not lit** yet remains **locatable**; it MUST **not** appear **more visually prominent** than **on** lights or than the **wire segments**.
3. The visualization MUST **update** when light state changes from **REQ-011** while the user is viewing the affected model, so the **sphere** **appearance** matches the **latest** **authoritative** state (**REQ-039**) within a **reasonable** delay (architecture may set bounds; **stale** display after a successful write is **not** acceptable indefinitely).
4. **REQ-010** **segments** and **hover**/**touch** **id** and **coordinates** behavior remain in force; state fields (**on/off**, **colour**, **brightness**) MAY be shown on hover/tap in addition if architecture chooses.
5. Lights **without** established authoritative state yet in the running service MUST follow the **default** defined for **REQ-011** (**REQ-014**, **REQ-039**) and still render per rules **1** and **2** above.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: On/off and colour cues MUST remain **discernible** at small viewports (contrast, transparency level tuned per **REQ-002**).
- Tablet: Same as mobile.
- Desktop: Same; **hover** may also show state fields if implemented.

**Dependencies**

- REQ-002, REQ-010, REQ-011, REQ-041

**Open questions**

- Maximum acceptable **latency** after an API update before the 3D view refreshes.
- Whether **hover**/**tap** must display **brightness** and **colour** in addition to **id**/**coordinates**.

---

## REQ-013 — Model view: multi-select light settings and paged light list

| Field | Value |
|-------|-------|
| **ID** | REQ-013 |
| **Title** | Model view: multi-select light settings and paged light list |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user **viewing** a model, I want to **select multiple lights** and **apply the same settings** to all of them at once, and I want the **list of lights** to be **paginated** with control over **how many lights appear per page** and a way to **jump to a specific light by its id**, so that I can manage large models without scrolling through thousands of rows and can batch-update state efficiently.

**Scope**

- In scope: On the **model view** (**REQ-006**), a **list** (or equivalent tabular presentation) of the model’s lights that is **paged** (only one **page** of lights shown at a time). Controls to **change the page size** (number of lights per page) from **presets** or an agreed control pattern. A **“go to light”** (or equivalent) control that accepts a light **id** (**REQ-005**) and navigates the list to the **page** that contains that id, with **clear feedback** if the id is **invalid** or **out of range** for the model. **Multi-select** of lights **on the current page** (and, if architecture supports it, **across pages** via retained selection—see open questions) plus a **bulk apply** action that sets **on/off**, **hex colour**, and **brightness** (**REQ-011**) **to the same values** for **every selected** light; successful updates MUST update **authoritative** state per **REQ-011** and **REQ-039**.
- Out of scope: New state fields beyond **REQ-011**; reordering lights; deleting lights from the model; export of selection.

**Business rules**

1. The model view MUST include a **paginated** light list when the model has **more than one** light; for a **single** light, pagination MAY be omitted or trivial (architecture).
2. The user MUST be able to **change how many lights** are shown **per page** using at least **three** distinct positive page-size choices within **1** and **1000** (inclusive), documented or labeled in the UI (exact values deferred to architecture).
3. The user MUST be able to **jump** to the page containing a given light **id** by entering or choosing that **id**; if the id is **not** an integer in **0 … n − 1** for the model’s light count **n**, the system MUST **not** change page silently and MUST show **actionable** feedback.
4. The user MUST be able to **select multiple lights** (multi-select) from the list using **pointer** and **keyboard**-friendly patterns on desktop and an **equivalent** on **touch** (**REQ-002**), e.g. checkboxes or a documented multi-select gesture.
5. When **one or more** lights are selected, the user MUST be able to **apply** settings (**on/off**, **hex colour**, **brightness** per **REQ-011**) so that **each selected** light receives the **same** applied values; validation rules for colour and brightness MUST match **REQ-011**.
6. When **two** **or** **more** **lights** **are** **selected** **and** **the** **user** **applies** **settings** **in** **one** **deliberate** **bulk** **action**, **the** **client** **MUST** **use** **the** **documented** **model-scoped** **batch** **light-state** **HTTP** **API** **from** **`docs/architecture.md`** **(**a** **single** **request** **listing** **all** **target** **light** **ids** **and** **the** **shared** **state** **fields**)** **rather** **than** **relying** **on** **a** **sequence** **of** **per-light** **requests** **as** **the** **only** **mechanism** **. **A** **single** **selected** **light** **MAY** **still** **use** **the** **per-light** **REQ-011** **surface** **if** **architecture** **allows** **.
7. After a **successful** bulk apply, **REQ-012** applies: the **3D view** and list MUST reflect updated state within the same **timeliness** expectations as single-light updates (**no indefinite staleness** after success).
8. Pagination and multi-select MUST remain **usable** on **mobile**, **tablet**, and **desktop** (**REQ-002**): primary controls reachable, readable labels, and no reliance on **hover-only** affordances for essential actions.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Page size and “go to id” controls usable without horizontal scrolling for primary content; multi-select via touch-friendly controls; bulk apply clearly labeled.
- Tablet: Same as mobile; orientation changes handled where applicable.
- Desktop: Keyboard navigation where applicable; multi-select with checkboxes or shift-range if architecture provides it.

**Dependencies**

- REQ-002, REQ-006, REQ-011, REQ-012

**Open questions**

- Whether **selection** is **per-page only** or MUST **persist** when the user changes pages (cross-page bulk apply).
- Exact **preset** page sizes and default page size.

---

## REQ-014 — Default light state (off, 100% white) and model reset

| Field | Value |
|-------|-------|
| **ID** | REQ-014 |
| **Title** | Default light state (off, 100% white) and model reset |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user working with a model, I want **every** light to **start** in a **known** state—**off** but configured as **full white** at **100%** brightness—and I want a **reset** control that returns **all** lights to that state after I have changed them, so that behaviour is predictable and I can quickly restore a clean baseline.

**Scope**

- In scope: **Initial** per-light state for **each** light when a model first exists in the system (including **new** models after **CSV** upload or **seeded** samples, and **legacy** rows **backfilled** per architecture): **on** = **false** (**off**), **brightness** = **100** (percent, per **REQ-011**), **hex colour** = **`#FFFFFF`** (six-digit canonical form per **REQ-011**). On the **model view** (**REQ-006**), a **reset** affordance (e.g. a **Reset** button or equivalent clearly labeled control) that sets **every** light in the **current** model to that **same** triple (**off**, **100%**, **`#FFFFFF`**) in **one** deliberate user action, updating **authoritative** state per **REQ-011** and **REQ-039** (including propagation to any assigned device per **REQ-038**).
- Out of scope: **Undo**/**redo** stacks; resetting **only** selected lights; resetting **3D** camera or navigation.

**Business rules**

1. For **every** light in a model, the **initial** stored state (including after model **creation** and any **migration** of older data) MUST be **off** (**on** = **false**), **brightness** = **100**, and **hex colour** = **`#FFFFFF`** (rejecting invalid hex remains per **REQ-011**).
2. The **model view** MUST include a **reset** control, **reachable** and **operable** without **hover-only** interaction (**REQ-002**), that restores **all** lights in that model to the state in rule **1** and updates **authoritative** in-memory state so **API** clients and **other** **connected** **clients** see the same result **without** requiring a full page reload (**REQ-039**); after a **process** **restart**, initial state follows **REQ-039** sync rules rather than a database snapshot.
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

## REQ-015 — Scenes: composite space, placement, CRUD, and three.js view

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
8. When viewing a scene, the product MUST use **three.js** as a **direct** front-end dependency (**REQ-010**) and MUST render **all** lights of **all** assigned models **without** omission when **n ≤ 1000** per model (**REQ-005**), with **segments** only between **consecutive** **ids** **within** the **same** model (**REQ-005**), **not** between models, **using** **derived** **scene-space** **positions** (**rule** **4**). The **scene** **three.js** **view** MUST **also** **show** **the** **faint** **axis-aligned** **boundary** **cuboid** **per** **REQ-034** **(**tight** **AABB** **on** **all** **lights’** **`sx,sy,sz`** **plus** **the** **scene’s** **persisted** **boundary** **margin** **`m`**, **default** **0.3** **m** **—** **rules** **12**–**13**)**.
9. **Per-light** **state** (**REQ-011**, **REQ-012**) MUST apply in the scene view **per model** as in the single-model view (colours, brightness, on/off, and **timely** updates).
10. **Removing** a model when **more than one** model remains MUST **persist** **without** deleting the scene. **Removing** the **last** **remaining** model MUST **not** **silently** delete the scene: the user MUST **first** **confirm** a **clear** **explanation** that **removing** this **last** model will **delete** the **entire** scene (and **that** the scene will **disappear** from **lists**). On **confirm**, the system MUST **delete** the scene (and **remove** the model assignment). On **cancel**, the scene MUST **remain** **unchanged**. This **confirm** flow MUST be **usable** without **hover-only** **essential** steps (**REQ-002**).
11. The scene view MUST provide **affordances** to **add** a **model** (from those available per **REQ-006**) and to **remove** a model, subject to **rules** **6** and **10**; the user MUST be able to **adjust** placements for models **already** in the scene (**subject** to **rule** **5**); **successful** changes MUST **persist** per architecture **without** **altering** **stored** **model** **light** **coordinates** (**rule** **4**).
12. **Boundary** **margin** **for** **visualization** **(**REQ-034**)** **:** **Each** **persisted** **scene** **MUST** **store** **one** **non-negative** **finite** **number** **`m`** **(**SI** **meters**)** **—** **the** **symmetric** **padding** **applied** **on** **all** **six** **faces** **when** **drawing** **the** **faint** **boundary** **cuboid** **around** **the** **tight** **light** **AABB**. **On** **create**, **`m`** **MUST** **default** **to** **0.3** **(**30** **cm**)**. **Existing** **scenes** **without** **`m`** **MUST** **be** **treated** **as** **`m` = 0.3** **after** **migration** **(**architecture** **implements** **default** **column** **/** **read** **fallback**)**.
13. **Editing** **`m`** **:** **The** **scene** **detail** **/** **primary** **scene** **management** **UI** **(**where** **the** **user** **views** **the** **composite** **three.js** **viewport** **)** **MUST** **expose** **a** **control** **(**e.g.** **numeric** **input** **with** **unit** **label** **cm** **or** **m** **)** **to** **read** **and** **change** **`m`**, **subject** **to** **validation** **(**rule** **12** **and** **architecture** **max** **if** **any**)**. **Saving** **`m`** **MUST** **persist** **it** **and** **the** **next** **render** **of** **the** **boundary** **MUST** **reflect** **the** **new** **value** **(**REQ-034** **rule** **4**)**. **The** **control** **MUST** **not** **rely** **on** **hover-only** **steps** **to** **apply** **(**REQ-002**)**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Scene list, scene detail/view, **create** **without** **manual** **offset** **entry**, add/remove, placement editing, and **last-model** **confirmation** are reachable and readable; **three.js** viewport usable like the model view (**REQ-010**); **touch** equivalent for light **id** and **scene**/**model** **coordinate** **disclosure** where applicable.
- Tablet: Same as mobile; orientation changes handled where applicable.
- Desktop: Efficient navigation between scene list and scene view; **hover** on lights where applicable.

**Dependencies**

- REQ-002, REQ-005, REQ-006, REQ-010, REQ-011, REQ-012, REQ-034, REQ-041

**Open questions**

- Whether **scene names** must be **globally unique**.
- Exact mapping of **(0,0,0)** and **which** **axis** is **“right”** vs **“up”** in the **default** **three.js** **scene** setup.
- Whether **placement** integers are **meters** rounded, **decimeters**, or **unitless grid** cells (must be **consistent** with **REQ-005** **meter** coordinates in architecture).

---

## REQ-016 — Camera reset for model and scene 3D views

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

## REQ-017 — Options: factory reset with confirmation

| Field | Value |
|-------|-------|
| **ID** | REQ-017 |
| **Title** | Options: factory reset with confirmation |
| **Priority** | Must |
| **Actor(s)** | End user; operator |

**User story**

As a user, I want an **Options** area that includes **factory reset**, so that I can **wipe** **all** **application** **data** and return to a **clean** state **with** the **default** **sample** **models** (**REQ-009**) **and** the **default** **Python** **sample** **routines** (**REQ-032**) **restored**—but **only** after I **confirm** **explicitly** **because** the **operation** is **dangerous** and **irreversible**.

**Scope**

- In scope: A **distinct** **Options** **section** (or **screen**/**panel** labeled **Options** or equivalent **clear** **navigation** **target**) in the **UI** that includes an action labeled **Factory reset** (or **equivalent** **unambiguous** **wording**). **Factory reset** MUST **remove** **all** **persisted** **user-relevant** **data** the product stores for **models**, **scenes** (**REQ-015**), **the** **entire** **device** **registry** (**REQ-035**) **and** **all** **device–model** **assignments** (**REQ-036**), **scene routines** (**REQ-021**, **REQ-022**, **REQ-033** definitions and any **persisted** **run** **state**), and **any** **other** **application** **content** **tied** to those **entities** (exact **store** **shape** per **architecture**); **per-light** **operational** **state** **is** **not** **durable** **(**REQ-039**)** **but** **the** **running** **process** **must** **reinitialize** **or** **clear** **in-memory** **state** **as** **architecture** **defines** **on** **reset**, then **re-seed** the **system** so the **user** sees the **same** **default** **sample** **models** as on a **fresh** **install** per **REQ-009** (**three** **samples**; **no** **user-uploaded** **models** or **user-created** **scenes** **remain**) **and** the **same** **three** **default** **Python** **routine** **definitions** as on a **fresh** **install** per **REQ-032** (**no** **other** **routine** **definitions** **remain**, **including** **no** **user** **shape** **animation** **definitions**). **Before** **any** **deletion** or **re-seed** **runs**, the user MUST be **prompted** with a **confirmation** **step** that **warns** of **data** **loss** and **irreversibility**; **Cancel** MUST **leave** **data** **unchanged**; **Confirm** MUST **complete** the **reset** and **surface** **success** **feedback** (exact **copy** deferred to **architecture**).
- Out of scope: **Partial** **wipe** (e.g. **only** **scenes**); **scheduled** **reset**; **remote** **admin** **API** **for** **factory** **reset** unless added later; **export** **before** **wipe** (user may **export** **elsewhere** if **features** **exist**—not **required** here).

**Business rules**

1. The **product** MUST expose an **Options** **section** (or **dedicated** **Options** **view**) **discoverable** from **primary** **navigation** or **settings** **pattern** **documented** in **architecture**; it MUST **list** **Factory reset** as **one** of its **actions**.
2. **Factory reset** MUST **not** **run** on a **single** **mis-click**: the user MUST **first** **see** a **blocking** **prompt** or **dialog** **before** **irreversible** **effects** **begin**, **explaining** that **all** **models**, **scenes**, **registered** **devices** (**REQ-035**, **REQ-036**), **routines** (**REQ-021**, **REQ-022**, **REQ-033**), and **related** **data** will be **permanently** **removed** and **only** **default** **sample** **models** and **default** **sample** **Python** **routines** will **remain** (wording **may** **name** **consequences** **explicitly** per **UX** **review**).
3. **Until** the user **explicitly** **confirms** (e.g. **Confirm** on the **dialog**), **no** **factory** **reset** **side** **effects** **may** **occur**; **dismissal** or **Cancel** MUST **preserve** **current** **data**.
4. After **confirmed** **factory reset**, **no** **user-created** **models**, **no** **scenes**, **no** **registered** **devices**, **no** **persisted** **routine** **run** **state**, and **no** **leftover** **state** **from** **prior** **entities** **may** **remain** **visible** in **listings** (or **equivalent** **discovery** **surfaces**) **except** **the** **re-seeded** **defaults**; the **model** **list** MUST **match** **REQ-009** **expectations** for a **fresh** **seed** (**exactly** **three** **sample** **models** **identifiable** as **sphere**, **cube**, **cone** per **REQ-009** **naming** **rules**), **and** the **routine** **definition** **list** MUST **match** **REQ-032** **expectations** for a **fresh** **seed** (**exactly** **three** **default** **Python** **sample** **routines** **with** **the** **behaviors** **named** **there**; **no** **other** **routine** **definitions**).
5. **Per-light** **defaults** after **re-seed** MUST **align** with **REQ-014** / **REQ-011** for **newly** **present** **models**.
6. The **entire** **flow** MUST **satisfy** **REQ-002**: **no** **hover-only** **requirement** for **opening** **Options**, **starting** **factory** **reset**, **or** **confirming** / **canceling**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Options** and **factory** **reset** **reachable** **without** **horizontal** **scroll** **for** **primary** **labels**; **confirmation** **readable** **without** **truncating** **the** **warning**.
- Tablet: Same as mobile.
- Desktop: **Options** **easy** **to** **find**; **dialog** **keyboard** **accessible** where **modals** **are** **used**.

**Dependencies**

- REQ-002, REQ-006, REQ-009, REQ-011, REQ-014, REQ-015, REQ-021, REQ-022, REQ-032, REQ-033, REQ-035, REQ-036

**Open questions**

- Whether **factory** **reset** **requires** **typing** a **phrase** (e.g. **“RESET”**) **in** **addition** **to** **Confirm** for **extra** **safety**.
- **Post-reset** **navigation** (**stay** **on** **page** vs **redirect** **to** **model** **list**).

---

## REQ-018 — Application shell: themes, collapsible navigation, branding, and Font Awesome controls

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

## REQ-019 — Three.js model and scene views: fixed dark-grey viewport

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

## REQ-020 — Scene spatial API for dimensions, filtered retrieval, and bulk updates

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

## REQ-021 — Scenes: routines (Python and shape animation, run/stop via scene API, volumetric rules)

| Field | Value |
|-------|-------|
| **ID** | REQ-021 |
| **Title** | Scenes: routines (Python and shape animation, run/stop via scene API, volumetric rules) |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As a user or integrator, I want **scene routines** I can **create**, **list**, and **delete**, and **run against a scene** so they **change light state** using the **scene API**, so that I can automate effects—often limited to lights inside a **volumetric region** in **scene space**—without hand-editing each light.

**Scope**

- In scope: **Routine** **definitions** **support** **two** **kinds**: **(**a**)** **Python** **routines** **(**REQ-022**)** **persisted** **with** **name**, **description**, **and** **routine** **kind** **Python**; **(**b**)** **declarative** **shape** **animation** **routines** **(**REQ-033**)** **with** **name**, **description**, **and** **kind** **shape** **animation**. User or API flows to **create**, **list**, and **delete** definitions (exact **create** **path** **per** **REQ-023**). **Start** a routine **against a chosen scene** and **stop** a running instance **without** deleting the definition. While **running**, **Python** **execution** (**REQ-022**) MUST apply state changes **only** through **scene-level** operations consistent with **REQ-020** (and **REQ-011** field semantics), using **scene-space** geometry for any **region-scoped** updates; **shape** **animation** **(**REQ-033**)** MUST apply **the** **same** **logical** **light-state** **updates** **through** **REQ-020**-class **scene** **API** **operations** **(**or** **documented** **equivalent** **server** **paths** **that** **preserve** **REQ-011** **semantics**)** **without** **executing** **user** **Python** **for** **that** **kind**. Canonical stored model coordinates MUST NOT be rewritten (**REQ-015**). **Effects** **authored** **in** **Python** MAY restrict which lights are affected using **cuboid** or **sphere** volumes in **scene space** (**REQ-020** shapes and composition rules). **Stopping** ends automation; **per-light** **authoritative** **state** remains whatever was last successfully applied in memory (**REQ-011**, **REQ-039**), including on physical devices when assigned (**REQ-038**). **Default** **sample** **behaviors** **(**including** **the** **former** **built-in** **“random** **colour** **cycle** **—** **all** **scene** **lights”** **effect**)** **are** **delivered** **as** **seeded** **Python** **definitions** **per** **REQ-032**, **not** **as** **separate** **non-Python** **routine** **types** **beyond** **the** **two** **kinds** **above**. **Physical** output for models with assigned devices follows **REQ-035**–**REQ-038**; **REQ-021** does not duplicate those specifications but **depends** on them when devices exist.
- Out of scope: **Routine** **kinds** **other** **than** **Python** **and** **shape** **animation** **(**REQ-033**)**; **editing** **routine** **definitions** **after** **create** **(**Python** **via** **REQ-022**; **shape** **animation** **via** **REQ-033** **authoring** **surface**)** **except** **as** **those** **requirements** **define**; **new** **volumetric** **shapes** **for** **REQ-020** **beyond** **cuboid** **and** **sphere** **(**shape** **animation** **uses** **those** **for** **light** **assignment**)**; **routines** **that** **move** **or** **reorder** **lights** **in** **REQ-005** **space**; **authentication** **policy**.

**Business rules**

1. The product MUST support **creating** a routine definition with **name**, **description**, and **routine** **kind** **either** **Python** **(**REQ-022**)** **or** **shape** **animation** **(**REQ-033**)**; **listing** all routine definitions; and **deleting** a definition. **Name** is **required** at create time; **description** MAY be empty if architecture allows, but the field MUST exist.
2. The product MUST support **starting** a routine **run** **scoped to exactly one scene** that exists at start time, and **stopping** an active run for that routine–scene pair (or **run identifier**—exact API deferred to architecture). **Start** MUST fail with **clear, actionable** errors if the scene does not exist or is not usable.
3. While a routine **run** is **active**, any change to **on/off**, **hex colour**, or **brightness** for lights in that scene performed **by** **the** **routine** MUST be effected **through** **the** **scene** **spatial** **and** **bulk** **update** **semantics** **from** **REQ-020** **with** **REQ-011** **field** **semantics**, **implemented** **entirely** **inside** **the** **Go** **service** **process**. **`docs/architecture.md`** **MUST** **document** **the** **concrete** **process** **model** **(**supervised** **Python** **child**, **native** **shape** **simulation**, **HTTP** **vs** **in-process** **calls** **to** **the** **same** **validation** **surface** **,** **packages**, **and** **stop** **semantics** **)** **without** **contradicting** **this** **requirement**. **Per-light** **model** **APIs** **(**REQ-011**)** **remain** **valid** **for** **manual** **user** **or** **integrator** **actions** **outside** **routine** **automation**. **The** **browser** **MUST** **NOT** **execute** **production** **routine** **automation** **(**no** **in-browser** **interpreter** **or** **timer** **that** **applies** **routine** **light** **mutations** **)** **;** **the** **web** **UI** **observes** **authoritative** **server** **state** **(**REQ-038** **,** **REQ-041**)** **.
4. **Volumetric targeting (when** **used** **by** **script**)**:** Inclusion MUST be evaluated against **scene-space** positions (**REQ-015**/**REQ-020**), using **cuboid** and/or **sphere** parameters as **the** **script** **and** **API** **define**. **Invalid** region geometry MUST be rejected per **REQ-020** error expectations.
5. **At** **most** **one** **active** **routine** **run** **per** **scene** **at** **a** **time** **:** **starting** **a** **run** **when** **that** **scene** **already** **has** **an** **active** **run** **MUST** **fail** **with** **a** **clear**, **actionable** **error** **(**e.g.** **HTTP** **409** **with** **a** **stable** **error** **code** **per** **architecture**)** **.
6. **Deleting** **or** **other** **destructive** **mutations** **of** **a** **routine** **definition** **while** **a** **run** **of** **that** **definition** **is** **active** **MUST** **be** **rejected** **with** **a** **clear**, **actionable** **error** **(**the** **user** **must** **stop** **the** **run** **first**)** **.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: If the UI exposes routines, **list**, **create**, **delete**, **start**, and **stop** MUST be usable without **hover-only** essential steps (**REQ-002**); labels identify **scene** and **routine** clearly.
- Tablet: Same as mobile; orientation changes do not hide primary actions.
- Desktop: Efficient access to routine list and run controls alongside scene workflows.

**Dependencies**

- REQ-002 (when UI surfaces routines), REQ-011, REQ-015, REQ-020, REQ-022, REQ-033, REQ-035, REQ-036, REQ-038, REQ-039, REQ-040

**Open questions**

- None

---

## REQ-022 — Python scene routines: in-browser editor, scene library, documentation, and run loop with forced stop

| Field | Value |
|-------|-------|
| **ID** | REQ-022 |
| **Title** | Python scene routines: in-browser editor, scene library, documentation, and run loop with forced stop |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user who is **new to Python** **(about** **the** **level** **of** **a** **twelve-year-old** **just** **starting)**, I want to **author**, **save**, **load**, **duplicate**, and **delete** **Python routines** for a **scene** using **clear**, **friendly** **wording** **in** **the** **UI**, an **in-browser** editor with **syntax highlighting**, **checking**, **completion**, and **formatting**, with **API** **documentation** **right** **under** **the** **editor** **that** **I** **can** **browse** **by** **function** **or** **attribute** **and** **paste** **examples** **into** **my** **code**, so that I can **automate** **lights** in **scene** **space** **without** **mastering** **raw** **HTTP**; when I **run** a routine **against** a **scene**, it should **loop** **continuously** **until** I **stop** it, and **stopping** MUST be able to **forcibly** **terminate** the routine if **needed**.

**Scope**

- In scope: A **user-facing** experience (same **page** or **clear** **single** **workflow**) that includes: (**1**) an **in-browser** **Python** **code** **editor** **implemented** **with** **CodeMirror** **6** (**major** **version** **6** **of** **CodeMirror** **and** **the** **`@codemirror/*`** **packages**) **with** **full** **syntax** **highlighting**; (**2**) **syntax** **and** **static** **checking** that surfaces **issues** in the **editor** (e.g. **diagnostics** **before** or **during** **edit**, **exact** **mechanism** per **architecture** — **MAY** **use** **Pyodide** **or** **similar** **in** **the** **browser** **for** **authoring** **only** **,** **not** **for** **production** **routine** **runs** **per** **REQ-021**/**REQ-038**)** **;** (**3**) **code** **completion** (**autocomplete** **appropriate** **for** **Python** **and** **the** **provided** **scene** **API**); (**4**) **automatic** **code** **formatting** **on** **user** **request** or **on** **save** (**exact** **trigger** per **architecture**, but **formatting** MUST **be** **available** **and** **enabled** **by** **default** **where** **the** **product** **supports** **it**); (**5**) **persistence** and **management** of **Python** **routine** **definitions**: **save**, **load** **(open** **existing)**, **duplicate**, and **delete**; (**6**) a **Python** **library** **or** **module** **surface** **supplied** **by** **the** **application** **that** **wraps** **scene-oriented** **capabilities** (**REQ-020** **semantics**, **REQ-015** **scene** **composition**) **with** **simple** **methods** **and** **attributes**—**illustrative** **examples** **include** **`scene.height`** (or **equivalent** **documented** **name** **for** **scene** **vertical** **extent**) **and** **`scene.getLightsWithinSphere`** (**or** **equivalent** **documented** **name** **for** **sphere** **filtered** **retrieval** **in** **scene** **space**); (**7**) **on** **the** **same** **page** **as** **the** **editor**, **reference** **documentation** **for** **that** **library** **with** **placement**, **interaction**, **and** **sample** **style** **per** **REQ-024** **(**below** **the** **editor**, **selectable** **entries**, **brief** **comments** **in** **samples**, **insert** **at** **caret**/**end**)** **and** **novice** **wording** **per** **business** **rule** **10**; (**8**) **starting** **a** **saved** **Python** **routine** **against** **a** **chosen** **scene** **in** **the** **unified** **region** **per** **REQ-027** **MUST** **cause** **the** **server** **to** **run** **the** **script** **in** **a** **continuous** **loop** **while** **the** **run** **remains** **active** **(**process** **model** **and** **fairness** **per** **`docs/architecture.md`** **and** **REQ-021**)** **;** **the** **browser** **only** **calls** **start**/**stop** **HTTP** **APIs** **and** **observes** **state** **(**REQ-038** **,** **REQ-041**)** **;** (**9**) **stopping** **an** **active** **Python** **routine** **run** **MUST** **meet** **REQ-040** **and** **architecture** **MUST** **document** **cooperative** **and** **forcible** **termination** **when** **the** **routine** **does** **not** **yield** **to** **a** **normal** **stop** **request** **.
- Out of scope: **Routine** **engines** **other** **than** **the** **Python** **interpreter** **(**REQ-022**)** **and** **the** **native** **shape** **animation** **engine** **(**REQ-033**)**; **arbitrary** **pip** **packages** **unless** **explicitly** **added** **later**; **editing** **model** **CSV** **or** **scene** **geometry** **from** **Python**; **authentication** **policy**; **remote** **IDE** **integration**.

**Business rules**

1. The **application** MUST provide **save**, **load** **(select** **and** **open** **an** **existing** **definition)**, **duplicate**, and **delete** **for** **Python** **routine** **definitions** **with** **clear** **labels** **or** **icons** **per** **REQ-018** **where** **they** **are** **buttons**.
2. The **application** MUST **use** **CodeMirror** **6** **for** **Python** **editing** **in** **the** **browser** **on** **the** **REQ-022** **authoring** **surface** **(the** **editable** **code** **buffer** **and** **its** **extensions** **such** **as** **linting** **or** **completion** **MUST** **be** **built** **on** **the** **`@codemirror/*`** **version** **6** **ecosystem**)**. The **editor** MUST **provide** **Python** **syntax** **highlighting** **across** **that** **buffer**.
3. The **editor** MUST provide **checking** that **surfaces** **syntax** **or** **static** **issues** **to** **the** **user** **(e.g.** **underline** **or** **problem** **panel)** **without** **requiring** **a** **separate** **desktop** **tool**.
4. **Code** **completion** **and** **auto-formatting** (**format** **document** **or** **format** **on** **save**) MUST **be** **enabled** **for** **the** **Python** **editor** **(user** **MAY** **disable** **per** **architecture** **if** **a** **setting** **exists**, **but** **defaults** **MUST** **favor** **novices**: **completion** **and** **formatting** **on** **by** **default** **where** **technically** **feasible**).
5. The **product** MUST **expose** **a** **documented** **Python** **API** **object** **(e.g.** **`scene`**) **bound** **to** **the** **currently** **selected** **scene** **during** **a** **run** **that** **maps** **to** **scene** **capabilities** **consistent** **with** **REQ-020** **(dimensions,** **queries,** **bulk** **updates** **in** **scene** **space)** **and** **REQ-011** **field** **semantics** **for** **light** **state**; **exact** **method** **names** **MAY** **differ** **from** **the** **examples** **in** **the** **user** **story** **if** **`docs/architecture.md`** **and** **the** **on-page** **reference** **list** **the** **canonical** **names**.
6. **Reference** **documentation** **for** **the** **Python** **scene** **library** MUST **appear** **on** **the** **same** **page** **as** **the** **editor** **and** MUST **satisfy** **REQ-024** **for** **layout** **(**directly** **below** **the** **code** **editor**), **per-entry** **selection** **and** **detail**, **commented** **sample** **snippets**, **and** **inserting** **the** **shown** **example** **into** **the** **editor** **at** **the** **caret** **or** **end** **of** **the** **buffer** **as** **defined** **there**; **in** **all** **cases** **the** **copy** **MUST** **remain** **novice-oriented** **(**plain** **language**, **parameter** **descriptions**)** **with** **cross-links** **or** **anchors** **between** **editor** **and** **docs** **where** **helpful**.
7. **Starting** a **Python** **routine** **against** a **scene** MUST **cause** **the** **server** **to** **execute** **the** **script** **in** **a** **continuous** **loop** **while** **the** **run** **remains** **active** (**architecture** **documents** **iteration** **timing**, **whether** **sleeps** **are** **implicit**, **and** **fairness** **with** **other** **runs** **—** **REQ-038**)**.
8. **Stopping** a **Python** **routine** **run** MUST **cease** **further** **server-side** **loop** **iterations** **promptly** **under** **normal** **conditions**; **the** **implementation** MUST **also** **support** **forcible** **termination** **when** **the** **routine** **does** **not** **respond** **to** **cooperative** **stop** **within** **architecture-defined** **bounds** (**documented** **in** **`docs/architecture.md`**, **REQ-040**)**.
9. **Python** **routine** **automation** **MUST** **affect** **lights** **only** **through** **the** **documented** **scene** **API** **surface** **(wrapping** **REQ-020**/**REQ-011** **semantics)**; **it** MUST **not** **rewrite** **canonical** **stored** **model** **coordinates** (**REQ-005**, **REQ-015**).
10. **All** **user-visible** **instructional** **text** **on** **the** **Python** **routine** **authoring** **surface** **(**section** **headings**, **labels**, **primary** **tooltips**, **empty** **states**, **and** **short** **inline** **help** **where** **the** **product** **offers** **it**)** **MUST** **use** **wording** **understandable** **to** **a** **twelve-year-old** **who** **has** **just** **started** **learning** **Python** **(**short** **sentences**, **everyday** **words**, **and** **brief** **explanations** **when** **a** **specialist** **term** **cannot** **be** **avoided**)** **without** **long** **paragraphs** **of** **exposition** **in** **the** **chrome** **around** **the** **editor**.
11. **REQ-002** **applies**: **editor**, **documentation**, **and** **run**/**stop** **controls** MUST **remain** **usable** **on** **mobile**, **tablet**, **and** **desktop** **without** **hover-only** **essential** **steps** (**layout** **MAY** **stack** **editor** **and** **docs** **on** **narrow** **viewports**).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Editor** **and** **documentation** **reachable** **without** **losing** **context** **(e.g.** **tabbed** **docs** **or** **expandable** **sections)**; **save**/**load**/**duplicate**/**delete**/**run**/**stop** **touch-friendly**; **diagnostics** **readable**.
- Tablet: **Same** **as** **mobile**; **optional** **side-by-side** **editor** **and** **docs** **when** **width** **allows**.
- Desktop: **API** **reference** **follows** **REQ-024** **(**below** **the** **editor**)**; **keyboard** **shortcuts** **for** **format**/**save** **where** **architecture** **provides** **them**; **optional** **secondary** **panes** **are** **allowed** **only** **if** **they** **do** **not** **contradict** **REQ-024** **placement** **or** **REQ-027** **unification**.

**Dependencies**

- REQ-002, REQ-011, REQ-015, REQ-020, REQ-021, REQ-018 (buttons/icons where applicable), REQ-024 (API reference placement and interaction), REQ-027 (unified run and live viewport), REQ-032 (default seeded sample routines), REQ-040

**Open questions**

- **Resource** **limits** **for** **supervised** **Python** **routine** **children** **(**CPU**, **memory**, **wall-clock** **per** **iteration**)** **—** **exact** **caps** **per** **architecture** **.

---

## REQ-023 — Create routine: choose Python or shape animation (no third executable kinds)

| Field | Value |
|-------|-------|
| **ID** | REQ-023 |
| **Title** | Create routine: choose Python or shape animation (no third executable kinds) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user **creating** a **new** **routine**, I want **a** **clear** **choice** **between** **Python** **and** **shape** **animation** **(**REQ-033**)** **so** **that** **I** **land** **in** **the** **right** **authoring** **experience** **without** **obsolete** **third-party** **engines** **or** **redundant** **duplicate** **entry** **points**.

**Scope**

- In scope: **REQ-021** **defines** **Python** **and** **shape** **animation** **routine** **kinds**. The **user-facing** **flow** **to** **start** **creating** a **new** **routine** **MUST** **let** **the** **user** **choose** **either** **(**a**)** **Python** **(**then** **proceed** **to** **REQ-022** **authoring**)** **or** **(**b**)** **shape** **animation** **(**then** **proceed** **to** **REQ-033** **authoring**)**. **The** **product** **MUST** **not** **offer** **any** **third** **creatable** **routine** **kind** **(**e.g.** **no** **legacy** **“test** **routine”** **engine** **separate** **from** **these** **two**)**. **Architecture** **MAY** **use** **a** **single** **“new** **routine”** **action** **followed** **by** **a** **kind** **choice**, **or** **two** **parallel** **primary** **actions** **(**e.g.** **“New** **Python** **routine”** **and** **“New** **shape** **animation”**)** **if** **each** **is** **distinct** **and** **not** **redundant** **(**rule** **3**)**. **The** **product** **MUST** **not** **require** **two** **separate** **primary** **buttons** **that** **both** **start** **the** **same** **Python-only** **flow** **(**one** **path** **per** **kind** **suffices**)**.
- Out of scope: **Changing** **REQ-021**/**022**/**033** **persistence** **or** **API** **contracts** **except** **where** **the** **UI** **must** **bind** **to** **them**; **editing** **existing** **definitions**.

**Business rules**

1. **New** **routine** **creation** **MUST** **present** **exactly** **the** **two** **creatable** **kinds** **defined** **in** **REQ-021**: **Python** **(**REQ-022**)** **and** **shape** **animation** **(**REQ-033**)** **—** **no** **additional** **executable** **routine** **engines**.
2. **If** **the** **UI** **shows** **a** **type** **or** **kind** **control**, **its** **options** **MUST** **map** **only** **to** **those** **two** **kinds** **(**not** **multiple** **distinct** **engines** **beyond** **Python** **vs** **declarative** **shape** **animation**)**.
3. **The** **UI** **MUST** **not** **present** **redundant** **standalone** **primary** **actions** **that** **start** **the** **same** **authoring** **flow** **(**e.g.** **two** **buttons** **both** **creating** **only** **Python** **when** **one** **“new** **Python** **routine”** **or** **a** **single** **new** **routine** **plus** **kind** **picker** **already** **covers** **it**)**.
4. **REQ-002** **applies**: **any** **create** **control** **MUST** **be** **operable** **on** **mobile**, **tablet**, **and** **desktop** **without** **hover-only** **essential** **steps**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Create** **and** **kind** **choice** **touch-friendly**; **labels** **distinguish** **Python** **from** **shape** **animation** **clearly**.
- Tablet: **Same** **as** **mobile**.
- Desktop: **Keyboard** **and** **pointer** **access** **to** **create** **flow**.

**Dependencies**

- REQ-002, REQ-021, REQ-022, REQ-033

**Open questions**

- **Preferred** **pattern** **:** **one** **“new** **routine”** **then** **modal** **kind** **choice** **vs** **two** **visible** **primary** **actions** **on** **the** **routine** **list**.

---

## REQ-024 — Python routine view: API reference below editor, selectable detail, commented samples, insert snippet

| Field | Value |
|-------|-------|
| **ID** | REQ-024 |
| **Title** | Python routine view: API reference below editor, selectable detail, commented samples, insert snippet |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user authoring **Python** **routines**, I want the **scene** **API** **documentation** **directly** **under** **the** **code** **editor**, where I can **pick** **a** **function** **or** **attribute** **to** **read** **more** **and** **see** **sample** **usage**, **tap** **a** **button** **to** **drop** **that** **example** **into** **my** **script** **at** **my** **cursor** **or** **the** **end** **of** **the** **file**, and **see** **short** **comments** **inside** **every** **sample**, so that I can **learn** **the** **API** **without** **leaving** **the** **page** **or** **retyping** **examples**.

**Scope**

- In scope: On the **Python** **routine** **authoring** **surface** (**REQ-022**), a **dedicated** **API** **reference** **region** **that** **appears** **immediately** **below** **the** **code** **editor** **in** **vertical** **document** **order** **(no** **other** **primary** **workflow** **block** **between** **the** **editor** **and** **this** **reference** **except** **minimal** **separators** **or** **headings** **per** **architecture**)** **and** **that** **catalogs** **all** **public** **methods**, **functions**, **and** **documented** **attributes** **on** **the** **Python** **scene** **binding** **(the** **`scene`** **object** **or** **equivalent)**. **The** **user** **MUST** **be** **able** **to** **select** **which** **catalog** **entry** **is** **expanded** **or** **detailed** **(**list**, **dropdown**, **search**, **or** **equivalent** **single-entry** **focus**)** **and** **see** **for** **that** **entry** **additional** **information** **including** **at** **least** **one** **sample** **showing** **typical** **usage** **(may** **assume** **`scene`** **is** **bound** **during** **a** **run** **unless** **architecture** **documents** **otherwise)**. **Each** **such** **sample** **MUST** **use** **Python** **`#`** **comments** **with** **brief**, **easy** **explanations** **(**not** **long** **paragraphs**)** **so** **a** **beginning** **reader** **can** **follow** **line** **by** **line**. **The** **product** **MUST** **provide** **an** **affordance** **(**e.g.** **a** **button** **with** **REQ-018** **icon** **rules** **where** **it** **is** **a** **button**)** **that** **inserts** **the** **currently** **displayed** **example** **snippet** **into** **the** **CodeMirror** **buffer** **at** **the** **active** **caret** **when** **the** **editor** **has** **focus** **and** **a** **caret** **position** **is** **defined**; **when** **no** **caret** **is** **available** **(**e.g.** **editor** **not** **focused**)** **the** **insertion** **MUST** **append** **at** **the** **end** **of** **the** **file** **(**exact** **edge** **cases** **such** **as** **read-only** **state** **per** **architecture**)**.
- Out of scope: **Full** **OpenAPI** **HTTP** **catalog** **as** **the** **only** **form** **of** **docs** **unless** **architecture** **collapses** **Python** **to** **thin** **HTTP** **wrappers** **only**; **replacing** **REQ-027**’s **unified** **run**/**viewport** **region** **(**which** **follows** **the** **reference** **in** **page** **order** **per** **architecture**)**; **using** **this** **reference** **as** **the** **primary** **place** **where** **the** **three** **default** **seeded** **Python** **sample** **routines** **(**REQ-032**)** **must** **live** **—** **those** **are** **persisted** **routine** **definitions** **(**opened** **from** **the** **routine** **list** **like** **any** **other** **Python** **routine**)** **and** **are** **not** **required** **to** **appear** **as** **dedicated** **REQ-024** **catalog** **entries** **for** **whole** **scripts**.

**Business rules**

1. The **API** **reference** **section** **MUST** **be** **placed** **directly** **below** **the** **code** **editor** **on** **the** **Python** **routine** **view** **so** **scrolling** **from** **the** **editor** **naturally** **reaches** **the** **documentation** **next**.
2. The **catalog** **MUST** **enumerate** **every** **Python**-**exposed** **scene** **API** **surface** **element** **that** **the** **product** **supports** **for** **routines** **(no** **deliberate** **omission** **of** **a** **public** **operation**)**.
3. The **user** **MUST** **be** **able** **to** **choose** **one** **function**, **method**, **or** **documented** **attribute** **at** **a** **time** **(**or** **clear** **equivalent** **focus** **model**)** **to** **view** **expanded** **detail** **and** **sample** **usage** **for** **that** **item**.
4. **Every** **sample** **code** **snippet** **shown** **in** **this** **reference** **MUST** **include** **Python** **comments** **that** **briefly** **describe** **what** **the** **code** **does** **without** **being** **verbose**.
5. The **product** **MUST** **expose** **a** **control** **that** **inserts** **the** **currently** **shown** **example** **into** **the** **editor** **at** **the** **caret** **when** **the** **caret** **is** **active** **in** **the** **editor**, **otherwise** **at** **the** **end** **of** **the** **buffer**.
6. **REQ-002** **and** **REQ-022** **apply**: **the** **reference** **region**, **selector**, **and** **insert** **control** **MUST** **remain** **usable** **on** **mobile**, **tablet**, **and** **desktop** **without** **hover-only** **essential** **steps**, **and** **labels** **MUST** **follow** **REQ-022** **plain** **wording** **expectations**.
7. **Per-entry** **samples** **in** **this** **catalog** **document** **individual** **API** **items** **per** **rules** **2**–**5**; **they** **do** **not** **replace** **the** **requirement** **to** **ship** **the** **three** **full** **default** **Python** **routines** **as** **first-class** **saved** **definitions** **per** **REQ-032**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Snippets** **use** **horizontal** **scroll** **only** **if** **necessary**; **touch** **targets** **for** **catalog** **selection** **and** **insert** **meet** **REQ-002**.
- Tablet: **Same** **as** **mobile**; **reference** **stays** **below** **the** **editor** **in** **vertical** **order**.
- Desktop: **Snippets** **may** **use** **syntax** **styling** **consistent** **with** **the** **editor** **where** **feasible**; **keyboard** **focus** **order** **should** **reach** **selector**, **detail**, **and** **insert** **without** **traps**.

**Dependencies**

- REQ-002, REQ-018, REQ-022, REQ-027 (page order relative to unified run viewport per architecture), REQ-032 (default routines are not required as whole-script catalog rows)

**Open questions**

- **Whether** **long** **snippets** **or** **detail** **panels** **are** **collapsed** **by** **default** **on** **small** **viewports**.
- **Whether** **insert** **should** **also** **replace** **a** **selection** **when** **text** **is** **highlighted** **(**architecture** **may** **define** **as** **caret** **only** **or** **replace** **selection**)**.

---

## REQ-025 — Python routine default code: sphere interior colour changes

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
3. The **template** **MUST** **include** **Python** **`#`** **comments** **that** **briefly** **explain** **each** **main** **step** **(**same** **brevity** **standard** **as** **REQ-024** **samples**)**.
4. **REQ-022** **defaults** **for** **editor** **features** **(completion,** **formatting**)** **remain** **in** **force**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Default** **template** **is** **plain** **text** **in** **the** **editor**; **no** **extra** **UX** **beyond** **REQ-022**.
- Tablet: **Same** **as** **mobile**.
- Desktop: **Same** **as** **mobile**.

**Dependencies**

- REQ-011, REQ-020, REQ-022, REQ-024

**Open questions**

- **Whether** **the** **template** **should** **include** **explicit** **`sleep`**/**delay** **calls** **or** **rely** **on** **the** **runtime** **loop** **(**REQ-022**)** **only**.

---

## REQ-026 — Python scene binding: width, depth, and height

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
3. **`docs/architecture.md`**, **the** **REQ-022** **on-page** **reference**, **and** **the** **REQ-024** **API** **reference** **MUST** **list** **all** **three** **attributes** **with** **definitions** **of** **which** **world** **axis** **each** **maps** **to**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **N/A** **for** **attribute** **surface** **alone**; **docs** **showing** **them** **follow** **REQ-002**/**REQ-024**.
- Tablet: **N/A**
- Desktop: **N/A**

**Dependencies**

- REQ-020, REQ-022, REQ-024

**Open questions**

- **Whether** **dimensions** **are** **evaluated** **from** **dynamic** **light** **bounds** **or** **fixed** **scene** **volume** **metadata** **(**REQ-015**/** **REQ-020**)**.

---

## REQ-027 — Routine authoring: unified run-in-scene and live viewport (Python and shape)

| Field | Value |
|-------|-------|
| **ID** | REQ-027 |
| **Title** | Routine authoring: unified run-in-scene and live viewport (Python and shape) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **user** **running** **or** **watching** **a** **Python** **or** **shape** **animation** **routine**, I want **one** **clear** **area** **where** **I** **pick** **the** **scene**, **start** **and** **stop** **the** **routine** **on** **that** **scene**, **and** **see** **the** **same** **scene** **live** **in** **three.js**, **with** **reset** **lights** **and** **reset** **camera**, **so** **that** **I** **do** **not** **juggle** **two** **different** **sections** **that** **do** **the** **same** **job** **(**like** **separate** **“run** **in** **scene”** **and** **“visual** **debug”** **blocks**)**.

**Scope**

- In scope: On **each** **routine** **authoring** **surface** **—** **Python** **(**REQ-022**)** **and** **shape** **animation** **(**REQ-033**)** **—** **a** **single** **unified** **primary** **region** **(**one** **section** **or** **card** **per** **architecture**)** **that** **combines** **(**a**)** **selection** **of** **which** **persisted** **scene** (**REQ-015**)** **the** **routine** **targets** **for** **execution**, **(**b**)** **run**/**stop** **(**or** **equivalent** **lifecycle** **controls** **)** **bound** **to** **that** **same** **scene**, **and** **(**c**)** **the** **live** **three.js** **viewport** **that** **shows** **that** **scene** **and** **updates** **when** **the** **routine** **changes** **light** **state** **(**the** **former** **“visual** **debug”** **behavior**)**. **The** **one** **scene** **choice** **MUST** **feed** **both** **execution** **and** **the** **viewport** **(**no** **second** **scene** **picker** **whose** **only** **purpose** **is** **to** **mirror** **the** **first**)**. **The** **viewport** **MUST** **follow** **REQ-010**/**REQ-012**/**REQ-015**/**REQ-019** **visual** **rules** **for** **spheres**, **segments**, **and** **dark-grey** **background** **as** **applicable** **to** **scene** **composite** **views**. **Two** **controls** **(each** **a** **button** **or** **equivalent** **accessible** **action** **with** **REQ-018** **icon** **rules** **where** **they** **are** **buttons**)**:
  - **Reset** **scene** **lights**: **sets** **every** **light** **in** **the** **selected** **scene** **to** **the** **default** **per-light** **state** **in** **REQ-014** **(**off**, **100%** **brightness**, **`#FFFFFF`** **hex**)** **by** **updating** **authoritative** **in-memory** **state** **per** **REQ-011** **and** **REQ-039** **(**not** **durable** **SQLite** **storage**)** **;** **does** **not** **change** **scene** **membership**, **placements**, **or** **canonical** **model** **coordinates** (**REQ-005**, **REQ-015**)** **. **Activating** **reset** **scene** **lights** **MUST** **stop** **any** **active** **routine** **run** **for** **that** **scene** **first** **(**same** **semantics** **as** **a** **successful** **user** **stop** **per** **REQ-021** **and** **REQ-040**)** **so** **the** **routine** **does** **not** **immediately** **overwrite** **the** **reset** **baseline** **.
  - **Reset** **camera**: **restores** **the** **default** **camera** **and** **control** **baseline** **for** **that** **viewport** **only**, **per** **REQ-016** **semantics** **(client-side** **navigation** **only**)**.
- Out of scope: **Persisting** **camera** **pose** **across** **sessions**; **factory** **reset** **(**REQ-017**)**; **deleting** **or** **creating** **scenes** **from** **this** **panel**; **a** **second** **top-level** **viewport** **for** **the** **same** **routine**/**scene** **pair** **under** **a** **different** **product** **name** **(**e.g.** **duplicated** **“debug”** **vs** **“run”** **panels**)**.

**Business rules**

1. The **product** **MUST** **present** **exactly** **one** **unified** **run**/**scene**/**viewport** **region** **on** **the** **Python** **routine** **authoring** **surface** **(**REQ-022**)** **and** **exactly** **one** **such** **region** **on** **the** **shape** **animation** **authoring** **surface** **(**REQ-033**)** **,** **each** **serving** **both** **starting**/**stopping** **the** **routine** **against** **a** **chosen** **scene** **and** **showing** **that** **scene** **live** **in** **three.js** **(**no** **parallel** **pair** **of** **sections** **that** **split** **“run** **in** **scene”** **from** **“visual** **debug”** **with** **separate** **viewports** **or** **scene** **selectors** **for** **the** **same** **workflow** **on** **that** **page**)**.
2. The **user** **MUST** **select** **one** **existing** **scene** **that** **is** **the** **target** **for** **routine** **execution** **and** **the** **source** **of** **truth** **for** **the** **live** **viewport**.
3. While **a** **scene** **is** **selected** **and** **the** **routine** **is** **running** **(**or** **when** **architecture** **allows** **static** **observation** **of** **last** **state** **after** **stop**)** **light** **state** **changes** **from** **the** **routine** **MUST** **become** **visible** **in** **the** **viewport** **within** **the** **same** **class** **of** **timeliness** **as** **REQ-012** **(no** **indefinite** **staleness** **after** **successful** **writes** **the** **client** **knows** **about**)**.
4. **Reset** **scene** **lights** **MUST** **apply** **REQ-014** **defaults** **to** **all** **lights** **in** **the** **selected** **scene** **in** **authoritative** **in-memory** **state** **(**REQ-011** **,** **REQ-039**)** **so** **other** **clients** **observe** **the** **same** **result** **via** **API** **or** **push** **;** **it** **MUST** **first** **stop** **any** **active** **routine** **run** **on** **that** **scene** **(**REQ-021** **,** **REQ-040**)** **.
5. **Reset** **camera** **MUST** **not** **alter** **persisted** **models**, **scenes**, **placements**, **or** **authoritative** **per-light** **state**; **it** **MUST** **only** **reset** **client** **navigation** **state** **for** **that** **viewport**.
6. **REQ-002** **and** **REQ-018** **apply** **to** **scene** **selection**, **run**/**stop**, **viewport** **interactions**, **and** **the** **two** **reset** **actions** **on** **both** **surfaces** **;** **novice** **wording** **(**REQ-022** **rule** **10**)** **applies** **to** **chrome** **on** **the** **Python** **page** **;** **shape** **authoring** **follows** **REQ-033** **responsive** **notes**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Scene** **selector**, **run**/**stop**, **viewport**, **and** **reset** **buttons** **remain** **reachable** **without** **hover-only** **steps**; **viewport** **supports** **touch** **navigation** **consistent** **with** **REQ-010**/**REQ-015**.
- Tablet: **Same** **as** **mobile**; **orientation** **changes** **keep** **primary** **actions** **discoverable** **in** **the** **unified** **region** **or** **an** **equally** **clear** **overflow** **pattern**.
- Desktop: **Reset** **controls** **sit** **adjacent** **to** **or** **within** **the** **viewport** **chrome** **per** **architecture**.

**Dependencies**

- REQ-002, REQ-010, REQ-011, REQ-012, REQ-014, REQ-015, REQ-016, REQ-018, REQ-019, REQ-021, REQ-022, REQ-033, REQ-041

**Open questions**

- None

---

## REQ-028 — Three.js light spheres: brightness-scaled emissive glow

| Field | Value |
|-------|-------|
| **ID** | REQ-028 |
| **Title** | Three.js light spheres: brightness-scaled emissive glow |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **user** **viewing** **a** **scene** **(**or** **a** **single** **model**)** **in** **three.js**, I want **each** **on** **light’s** **sphere** **to** **glow** **in** **proportion** **to** **its** **brightness** **and** **to** **read** **as** **emitting** **light**, **so** **that** **100%** **brightness** **looks** **clearly** **bright** **and** **dimmer** **settings** **look** **appropriately** **subdued**.

**Scope**

- In scope: **Per-light** **operational** **state** **(**on**/**off**, **hex** **colour**, **brightness** **per** **REQ-011**)** **is** **authoritative** **only** **in** **the** **running** **service’s** **in-memory** **representation** **(**REQ-039**)** **—** **not** **durable** **in** **SQLite** **or** **other** **on-disk** **storage** **for** **later** **reload** **as** **truth** **.** **This** **requirement** **governs** **only** **how** **the** **client** **renders** **spheres** **from** **the** **logical** **triple** **it** **currently** **holds** **or** **receives** **via** **API** **and** **push** **(**REQ-012**, **REQ-041** **where** **applicable**)** **;** **it** **does** **not** **imply** **that** **brightness** **is** **“persisted”** **in** **a** **database** **sense** **. **three.js** **rendering** **of** **light** **spheres** **where** **REQ-012** **already** **governs** **colour** **and** **on**/**off** **appearance** **(**single-model** **detail** **per** **REQ-010**, **scene** **composite** **detail** **per** **REQ-015**, **and** **unified** **routine** **authoring** **viewports** **per** **REQ-027** **where** **those** **spheres** **are** **shown**)**. **For** **lights** **in** **the** **on** **state**, **the** **material** **MUST** **use** **an** **emissive** **(**self-illuminated**)** **treatment** **so** **the** **sphere** **appears** **to** **emit** **light**, **not** **only** **a** **flat** **tinted** **surface**. **The** **strength** **of** **that** **glow** **(**or** **architecture-defined** **equivalent** **visual** **metric**)** **MUST** **scale** **monotonically** **with** **the** **current** **brightness** **percentage** **from** **authoritative** **in-memory** **light** **state** **(**REQ-011**, **REQ-039**)** **over** **0** **through** **100** **so** **that** **100%** **reads** **as** **strong** **glow** **and** **lower** **values** **read** **weaker** **while** **remaining** **ordered** **(**no** **inversion** **where** **a** **lower** **percent** **looks** **brighter** **than** **a** **higher** **one** **for** **the** **same** **hex** **colour**)**. **REQ-019** **dark-grey** **viewport** **remains** **so** **glow** **is** **visible** **against** **the** **background**.
- Out of scope: **Mandating** **a** **specific** **three.js** **API** **(**e.g.** **`MeshStandardMaterial.emissive`** **vs** **post-processing**)**; **physical** **global** **illumination** **or** **accurate** **luminaire** **photometry**; **changing** **REQ-010** **2** **cm** **sphere** **diameter** **or** **wire** **segment** **styling** **except** **where** **architecture** **must** **reconcile** **glow** **with** **existing** **rules**.

**Business rules**

1. **On** **lights** **(**`on` **true** **per** **REQ-011**/**REQ-012**)** **MUST** **use** **a** **material** **that** **includes** **a** **clear** **emissive** **(**light-emitting**)** **component** **tied** **to** **the** **current** **hex** **colour** **so** **the** **sphere** **reads** **as** **a** **light** **source** **rather** **than** **a** **purely** **diffuse** **ball**.
2. **Emissive** **strength** **(**or** **documented** **equivalent**)** **MUST** **scale** **with** **brightness** **percentage** **from** **REQ-011**: **at** **100%** **brightness** **the** **glow** **MUST** **be** **visibly** **strong**; **at** **lower** **percents** **the** **glow** **MUST** **be** **weaker** **in** **a** **way** **users** **can** **perceive** **as** **dimmer** **light** **(**exact** **curve** **per** **architecture**, **subject** **to** **monotonicity** **in** **rule** **3**)**.
3. **For** **two** **on** **lights** **with** **the** **same** **hex** **colour** **and** **different** **brightness** **values**, **the** **higher** **brightness** **MUST** **not** **appear** **less** **glowing** **than** **the** **lower** **(**monotonic** **ordering** **of** **glow** **vs** **percent**)**.
4. **Off** **lights** **MUST** **keep** **REQ-012** **dim** **grey** **transparent** **appearance**; **their** **emissive** **contribution** **MUST** **remain** **negligible** **so** **they** **do** **not** **appear** **to** **glow** **like** **on** **lights** **or** **outshine** **REQ-010** **wire** **segments**.
5. **When** **per-light** **state** **updates** **(**REQ-011**, **REQ-012**)** **or** **the** **user** **navigates** **between** **views**, **glow** **MUST** **stay** **consistent** **with** **the** **authoritative** **in-memory** **light** **state** **the** **client** **has** **observed** **(**REQ-039**)** **—** **no** **indefinite** **staleness** **after** **successful** **writes** **or** **equivalent** **push** **updates** **(**same** **timeliness** **class** **as** **REQ-012** **colour**/**opacity** **updates**)** **.
6. **`docs/architecture.md`** **MUST** **describe** **the** **chosen** **three.js** **material**/**rendering** **approach** **for** **REQ-028** **(**including** **how** **brightness** **maps** **to** **emissive** **intensity** **or** **equivalent**)** **and** **how** **it** **fits** **Pi**/**WebGL** **constraints** **from** **REQ-003** **where** **relevant**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Glow** **and** **sphere** **legibility** **remain** **acceptable** **on** **small** **WebGL** **viewports**; **touch** **hover-equivalent** **(**REQ-010**)** **unchanged**.
- Tablet: **Same** **as** **mobile**; **no** **hover-only** **dependency** **for** **perceiving** **brightness** **differences** **(**REQ-002**)**.
- Desktop: **User** **can** **compare** **brightness** **levels** **when** **multiple** **on** **lights** **are** **visible** **without** **excessive** **bloom** **that** **obscures** **ids** **(**architecture** **may** **cap** **or** **tone-map** **as** **needed**)**.

**Dependencies**

- REQ-002, REQ-003, REQ-010, REQ-011, REQ-012, REQ-015, REQ-019, REQ-027, REQ-039

**Open questions**

- **Whether** **a** **single** **global** **cap** **on** **emissive** **intensity** **is** **needed** **when** **many** **100%** **lights** **fill** **the** **viewport** **(**architecture** **trade-off** **vs** **REQ-010** **“do** **not** **merge** **lights”**)**.

---

## REQ-029 — High-throughput light state updates (batching, connections, push)

| Field | Value |
|-------|-------|
| **ID** | REQ-029 |
| **Title** | High-throughput light state updates (batching, connections, push) |
| **Priority** | Must |
| **Actor(s)** | Integrator; operator; end user (indirectly via UI) |

**User story**

As an integrator or operator driving dynamic lighting, I want the system to sustain frequent updates to many lights without being dominated by per-request overhead, so that scenes with hundreds of lights can change multiple times per second while viewers stay sufficiently up to date.

**Scope**

- In scope: Non-functional expectations for **authoritative** **in-memory** per-light state (**REQ-039**; same fields and validation semantics as **REQ-011**) when **updating large sets** in **model** and **scene** contexts (**REQ-015**, **REQ-020**), and for **viewers** to remain sufficiently fresh relative to **REQ-012** when change rates are high. **Scale assumption** (design target): on the order of **hundreds** of lights (consistent with **REQ-005**’s upper bound) with **multiple** aggregate **update cycles per second** across writes and/or viewer refresh. **Illustrative** mechanisms the architecture **must consider and document** (not all mandatory in isolation): **HTTP/2** or other **connection reuse**, **client connection pooling** / keep-alive, **batch or bulk write** APIs (including **REQ-020** where scene-scoped), **Server-Sent Events**, **WebSocket**, or **similar server-push** for distributing state changes to connected clients. Solutions **must remain plausible** on **Raspberry Pi 4** constraints (**REQ-003**).
- Out of scope: Hard numeric SLOs unless raised in **open questions**; mandating one specific protocol when another meets the same goals; low-level **WLED** **framing** (**REQ-035**–**REQ-038**); legacy **DMX** unless added later.

**Business rules**

1. **`docs/architecture.md` MUST** describe how the product meets high-throughput light updates at the stated scale: the **write path** (batching, use of **REQ-020** where applicable, in-memory authoritative updates and device push per **REQ-039**/**REQ-038** as relevant) and the **observer path** (how UIs and integrators obtain timely state—**server-push vs polling** with rationale).
2. Integrators **MUST NOT** be limited to **one HTTP request per light** as the **only** supported path for high-frequency multi-light changes: the product **MUST** expose **documented** **aggregate** update paths (for example **REQ-020** region bulk updates, and/or **model-scoped** batch operations if architecture defines them). **REQ-011** per-light read and write operations **remain** required for granular control.
3. **Connection reuse:** **`docs/architecture.md` SHOULD** document which **HTTP** features are enabled or assumed (for example **HTTP/2**, **HTTP/1.1 keep-alive**) and **client pooling** or equivalent guidance for integrators and for the shipped web UI.
4. For **low-latency** refresh or **many concurrent viewers**, **`docs/architecture.md` SHOULD** specify **server-push** (**SSE**, **WebSocket**, or **equivalent**) **or** justify **bounded polling** that still satisfies **REQ-012**-class timeliness under the stated load.
5. **Shipped** **web** **three.js** **viewports** **:** **REQ-041** **adds** **MUST-level** **push** **and** **incremental** **delta** **apply** **for** **those** **surfaces** **;** **rule** **4** **remains** **SHOULD** **for** **other** **observer** **clients** **unless** **they** **reuse** **the** **same** **mechanism**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: N/A at API level; any UI that displays live light state **MUST** remain usable per **REQ-002** at high update rates where the product targets them.
- Tablet: N/A; same as mobile where applicable.
- Desktop: N/A; same as mobile where applicable.

**Dependencies**

- REQ-003, REQ-011, REQ-012, REQ-015, REQ-020, REQ-041

**Open questions**

- Target **p99** latency or **updates per second** caps (if any) for the Pi deployment.
- Whether **model-scoped** batch writes are **required** in addition to **REQ-020** scene bulk for integrator workflows.
- How **multi-tab** or **multi-client** viewers stay consistent when push channels are used.

---

## REQ-030 — Python scene binding: random hex colour helper

| Field | Value |
|-------|-------|
| **ID** | REQ-030 |
| **Title** | Python scene binding: random hex colour helper |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a **user** **writing** **Python** **routines**, I want **the** **`scene`** **binding** **(**or** **documented** **equivalent** **per** **architecture**)** **to** **offer** **a** **simple** **way** **to** **get** **a** **random** **colour** **string** **I** **can** **pass** **into** **light** **state** **updates**, **with** **the** **same** **kind** **of** **value** **as** **`"#%06x"** **%** **random.randrange(0x1000000)`**, **so** **that** **I** **do** **not** **have** **to** **remember** **that** **formatting** **pattern** **or** **import** **`random`** **just** **for** **this** **one** **job**.

**Scope**

- In scope: A **single** **documented** **callable** **on** **the** **Python** **routine** **scene** **surface** **(**REQ-022**)** **that** **returns** **a** **string** **valid** **as** **`color`** **in** **REQ-011** **(**`#` **plus** **six** **hex** **digits**)** **for** **each** **call**, **using** **the** **statistical** **behavior** **defined** **in** **business** **rule** **2** **below**. **On-page** **reference** **(**REQ-024**)** **and** **`docs/architecture.md`** **MUST** **name** **the** **callable** **and** **show** **a** **short** **commented** **example** **consistent** **with** **REQ-024** **sample** **style**.
- Out of scope: **Palettes** **biased** **toward** **“bright”** **or** **pastel** **colours** **(**this** **requirement** **fixes** **the** **uniform** **24-bit** **pattern** **only**)**; **changing** **REQ-025** **default** **template** **wording** **unless** **architecture** **chooses** **to** **switch** **the** **template** **to** **the** **new** **helper**.

**Business rules**

1. The **product** **MUST** **expose** **exactly** **one** **primary** **documented** **entry** **point** **(**synchronous** **function**, **`async`** **function**, **or** **`scene`** **method** **per** **architecture**)** **for** **“random** **hex** **colour** **for** **light** **state”** **on** **the** **Python** **scene** **binding** **used** **during** **routines** **(**REQ-022**)**.
2. **Each** **call** **MUST** **return** **a** **string** **equivalent** **to** **formatting** **one** **integer** **sample** **uniformly** **from** **`0`** **inclusive** **through** **`0xFFFFFF`** **inclusive** **as** **`"#%06x"** **%** **that_integer`** **(**same** **distribution** **and** **string** **shape** **as** **`"#%06x"** **%** **random.randrange(0x1000000)`** **in** **standard** **Python** **3**)**. **Hex** **digits** **MUST** **be** **lowercase** **unless** **architecture** **documents** **a** **single** **consistent** **uppercase** **policy** **that** **still** **satisfies** **REQ-011**.
3. **The** **REQ-024** **API** **catalog** **MUST** **include** **this** **callable** **with** **a** **brief** **description** **and** **a** **sample** **that** **uses** **Python** **`#`** **line** **comments** **(**REQ-024** **sample** **rules**)**.
4. **CodeMirror** **completions** **(**REQ-022**)** **and** **any** **browser** **worker** **or** **authoring** **helper** **used** **for** **editor** **tooling** **MUST** **stay** **aligned** **with** **the** **chosen** **Python** **callable** **name** **and** **sync**/**`async`** **semantics** **documented** **in** **`docs/architecture.md`** **(**Python** **routine** **runtime** **and** **scene** **binding** **)** **.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **N/A** **(**API** **surface** **only**)**; **any** **UI** **labels** **for** **this** **feature** **follow** **REQ-002**/**REQ-018**.
- Tablet: **N/A**
- Desktop: **N/A**

**Dependencies**

- REQ-011, REQ-022, REQ-024

**Open questions**

- **Exact** **identifier** **(**e.g.** **`scene.random_hex_colour`** **vs** **module-level** **helper**)** **and** **whether** **the** **callable** **is** **`async`**.
- **Whether** **REQ-025** **default** **template** **should** **replace** **inline** **`random.randrange`** **with** **the** **new** **helper** **for** **pedagogical** **consistency**.

---

## REQ-031 — Skip redundant light-state work (visualization, in-memory authority, device traffic)

| Field | Value |
|-------|-------|
| **ID** | REQ-031 |
| **Title** | Skip redundant light-state work (visualization, in-memory authority, device traffic) |
| **Priority** | Should |
| **Actor(s)** | End user; web client (three.js views); service layer (authoritative in-memory state path) |

**User story**

As an **end** **user** **working** **with** **models** **or** **scenes**, I want **the** **product** **to** **avoid** **unnecessary** **three.js** **redraw** **work**, **unnecessary** **updates** **to** **authoritative** **in-memory** **per-light** **state** **(**REQ-039**)** , **and** **unnecessary** **traffic** **to** **assigned** **lighting** **devices** **(**REQ-035**–**REQ-038**)** , **when** **the** **effective** **state** **would** **not** **change**, **so** **that** **the** **UI** **stays** **responsive**, **the** **server** **does** **less** **redundant** **work**, **and** **external** **devices** **are** **not** **spammed** **with** **no-op** **updates** **during** **frequent** **or** **repeated** **operations** **(**including** **high-throughput** **paths** **per** **REQ-029**)**.

**Scope**

- In scope: **(**1**)** **Client-side** **detection** **that** **incoming** **or** **locally** **proposed** **per-light** **state** **(**on**/**off**, **hex** **colour**, **brightness** **per** **REQ-011**/**REQ-012**)** **is** **equivalent** **to** **what** **the** **visualization** **already** **shows** **for** **that** **light** **in** **the** **current** **view** **context**; **skipping** **full** **scene**/**mesh** **refresh** **or** **other** **expensive** **redraw** **work** **when** **the** **effective** **rendering** **state** **would** **not** **change**. **(**2**)** **Reducing** **redundant** **authoritative** **in-memory** **updates** **:** **where** **the** **implementation** **can** **determine** **that** **a** **proposed** **write** **(**REQ-011**,** **REQ-020**, **or** **equivalent** **paths**)** **would** **not** **change** **authoritative** **in-memory** **per-light** **state** **(**REQ-039**)** **after** **the** **same** **normalization** **as** **rule** **1**, **the** **product** **SHOULD** **avoid** **performing** **that** **update** **(**including** **skipping** **redundant** **device** **traffic** **per** **rule** **4** **when** **applicable**)** **without** **breaking** **REQ-012**-class **correctness** **for** **real** **changes** **(**per-light** **state** **is** **not** **durable** **application** **storage** **—** **REQ-039**)** **. **(**3**)** **An** **in-memory** **record** **of** **last-applied** **/** **last-known** **effective** **state** **(**cache** **or** **equivalent**)** **at** **client** **and** **/** **or** **server** **as** **architecture** **defines**. **(**4**)** **Physical** **sync** **:** **with** **REQ-035**–**REQ-038**, **design** **and** **documentation** **MUST** **ensure** **no-op** **logical** **state** **updates** **do** **not** **imply** **unnecessary** **outbound** **traffic** **to** **assigned** **devices** **(**WLED** **or** **equivalent**)** **when** **the** **effective** **output** **would** **not** **change** **(**same** **equivalence** **rules** **as** **elsewhere** **in** **this** **requirement**)**. **Applies** **to** **single-model** **three.js** **view** **(**REQ-010**/**REQ-012**)** **,** **scene** **composite** **view** **(**REQ-015**/**REQ-012**)** **,** **and** **the** **unified** **routine** **live** **viewports** **(**REQ-027**)** **for** **(**1**)** **;** **(**2**)** **covers** **any** **server** **path** **that** **commits** **authoritative** **per-light** **state** **including** **routines** **(**REQ-021**/**REQ-022**/**REQ-033**)** **where** **applicable**.
- Out of scope: **Defining** **WLED** **wire** **protocols** **or** **device** **registry** **(**those** **are** **REQ-035**–**REQ-038**)** **—** **this** **requirement** **only** **constrains** **equivalence** **and** **skip** **behavior** **for** **visualization**, **server** **authoritative** **state**, **and** **device** **traffic**; **hard** **frame-rate** **or** **latency** **SLOs** **(**Pi** **constraints** **remain** **per** **REQ-003**/**REQ-029**)**; **changing** **REQ-011** **resource** **shapes** **or** **error** **codes** **in** **a** **breaking** **way** **(**idempotent** **repeats** **may** **still** **succeed** **if** **submitted** **by** **an** **integrator** **that** **bypasses** **product** **short-circuits**)**.

**Business rules**

1. **Before** **applying** **incoming** **light** **state** **to** **the** **three.js** **visualization** **(**including** **after** **poll**, **push**, **or** **successful** **local** **mutation** **that** **would** **normally** **trigger** **a** **refresh**)** **,** **the** **client** **MUST** **determine** **whether** **the** **effective** **rendering** **relevant** **fields** **(**on**/**off**, **colour**, **brightness** **per** **REQ-012**)** **for** **each** **affected** **light** **differ** **from** **the** **last** **state** **already** **applied** **in** **that** **view** **context**. **When** **they** **are** **equivalent** **(**after** **any** **canonical** **normalization** **documented** **in** **`docs/architecture.md`** **for** **colour** **strings** **or** **numeric** **forms**)** **,** **the** **client** **MUST** **not** **perform** **a** **full** **visualization** **rebuild** **or** **other** **redraw** **work** **that** **exists** **solely** **to** **reflect** **that** **same** **state** **again**.
2. **The** **product** **SHOULD** **avoid** **applying** **redundant** **per-light** **state** **updates** **when** **the** **proposed** **new** **state** **is** **equivalent** **to** **the** **already** **authoritative** **in-memory** **state** **for** **that** **light** **(**REQ-039**)** **(**same** **field** **set** **and** **normalization** **as** **rule** **1**)** **,** **including** **for** **paths** **invoked** **by** **automation** **(**REQ-021**/**REQ-022**)** **and** **bulk** **/** **region** **updates** **(**REQ-020**)** **where** **architecture** **places** **the** **comparison**. **Observable** **API** **behavior** **for** **integrators** **(**success** **vs** **no-op** **response** **shape**)** **MUST** **be** **documented** **in** **`docs/architecture.md`** **so** **clients** **and** **device** **sync** **can** **rely** **on** **consistent** **semantics**.
3. **The** **product** **SHOULD** **maintain** **in-memory** **(**or** **architecture-defined** **equivalent**)** **records** **of** **last-applied** **/** **last-known** **effective** **per-light** **state** **where** **they** **reduce** **redundant** **work** **:** **e.g.** **per** **active** **model** **or** **scene** **view** **on** **the** **client** **(**updated** **when** **state** **is** **applied** **to** **the** **renderer** **or** **confirmed** **from** **the** **server**)** **,** **and** **/** **or** **on** **the** **server** **for** **the** **authoritative** **state** **short-circuit** **in** **rule** **2**. **Records** **MUST** **be** **cleared** **or** **resynchronized** **on** **navigation**, **model**/**scene** **change**, **or** **architecture-defined** **invalidation** **(**another** **tab**, **external** **integrator**, **or** **administrative** **action** **if** **documented**)**.
4. **`docs/architecture.md`** **MUST** **describe** **how** **assigned** **devices** **(**REQ-035**–**REQ-038**)** **receive** **only** **state** **changes** **that** **are** **not** **equivalent** **to** **what** **was** **last** **successfully** **applied** **on** **the** **device**/**channel** **(**or** **document** **an** **explicit** **alternative** **that** **still** **meets** **the** **“no** **unnecessary** **traffic”** **goal**)** **,** **reusing** **or** **aligning** **with** **the** **same** **equivalence** **rules** **as** **rules** **1**–**2** **where** **possible**.
5. **When** **per-light** **state** **does** **change**, **timeliness** **and** **correctness** **MUST** **remain** **consistent** **with** **REQ-012** **(**no** **indefinite** **staleness** **after** **successful** **writes** **the** **client** **knows** **about**)** **and** **REQ-010**/**REQ-015**/**REQ-027** **drawing** **rules** **unchanged**; **authoritative** **in-memory** **state** **and** **API** **read** **results** **MUST** **reflect** **real** **updates** **without** **skipping** **them** **due** **to** **stale** **caches**.
6. **`docs/architecture.md`** **MUST** **describe** **where** **equivalence** **is** **evaluated** **(**client**, **server**, **both**)** **,** **what** **is** **cached**, **invalidation** **rules**, **and** **documented** **no-op** **update** **behavior** **(**including** **device** **push** **skips** **where** **applicable**)** **,** **in** **line** **with** **REQ-029** **observer**/**refresh** **strategy** **where** **relevant**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Same** **responsiveness** **goal** **as** **desktop**; **no** **extra** **jank** **from** **redundant** **refreshes** **on** **touch** **devices**.
- Tablet: **Same** **as** **mobile**.
- Desktop: **Pointer** **and** **keyboard** **flows** **unchanged**; **benefit** **is** **smoother** **visual** **updates** **and** **fewer** **redundant** **round-trips** **under** **load**.

**Dependencies**

- REQ-010, REQ-011, REQ-012, REQ-015, REQ-020, REQ-021, REQ-022, REQ-027, REQ-029, REQ-035, REQ-036, REQ-038, REQ-039

**Open questions**

- **Whether** **cache** **is** **per** **browser** **tab** **only** **or** **coordinated** **across** **tabs** **(**if** **ever** **required**)**.
- **Granularity** **for** **bulk** **or** **region** **updates** **(**REQ-020**)** **:** **per-light** **short-circuit** **vs** **batch-level** **heuristics**.
- **Exact** **HTTP** **/** **API** **signaling** **when** **a** **write** **is** **skipped** **as** **a** **no-op** **(**e.g.** **204** **vs** **200** **with** **unchanged** **body**)** **for** **integrators**.

---

## REQ-032 — Default seeded Python sample routines (three: growing sphere, sweeping cuboid, random colour cycle all)

| Field | Value |
|-------|-------|
| **ID** | REQ-032 |
| **Title** | Default seeded Python sample routines (three: growing sphere, sweeping cuboid, random colour cycle all) |
| **Priority** | Must |
| **Actor(s)** | End user learning Python routines; maintainer (product defaults) |

**User story**

As a **novice** **Python** **user**, I want **three** **ready-made** **Python** **routines** **already** **in** **my** **routine** **list** **on** **first** **use**, **using** **only** **the** **standard** **documented** **Python** **scene** **API** **and** **heavily** **commented** **in** **plain** **language**, **so** **that** **I** **can** **run**, **read**, **edit**, **and** **duplicate** **them** **like** **any** **other** **routine** **without** **finding** **full** **scripts** **only** **under** **API** **help**.

**Scope**

- In scope: **Exactly** **three** **distinct** **persisted** **Python** **routine** **definitions** **(**REQ-021**, **REQ-022**)** **that** **the** **product** **creates** **automatically** **when** **the** **application** **first** **starts** **with** **no** **pre-existing** **routine** **store** **(**fresh** **install**)** **and** **again** **after** **a** **confirmed** **factory** **reset** **(**REQ-017**)** **—** **same** **triple** **as** **fresh** **install**. **Each** **MUST** **implement** **the** **normative** **behaviors** **in** **business** **rules** **3**–**5** **below**. **They** **MUST** **rely** **only** **on** **the** **public** **Python** **scene** **binding** **and** **documented** **helpers** **(**e.g.** **REQ-030** **random** **colour** **when** **appropriate**)** **as** **named** **in** **`docs/architecture.md`** **and** **the** **REQ-024** **catalog**. **Comments** **MUST** **be** **frequent**, **short**, **and** **task-oriented** **(**consistent** **with** **REQ-024** **sample** **comment** **style** **and** **REQ-022** **novice** **wording**)**. **These** **definitions** **MUST** **be** **editable**, **duplicable**, **and** **otherwise** **managed** **through** **the** **same** **surfaces** **as** **user-created** **Python** **routines** **(**REQ-022**)**. **They** **MUST** **not** **be** **delivered** **primarily** **as** **REQ-024** **API** **reference** **catalog** **entries** **for** **whole** **scripts** **(**REQ-024** **rule** **7**)** **—** **users** **open** **them** **from** **the** **routine** **list** **like** **any** **saved** **definition**.
- Out of scope: **Changing** **REQ-025** **default** **template** **for** **brand-new** **blank** **routines** **unless** **architecture** **aligns** **it** **with** **one** **of** **these** **behaviors**; **mandating** **a** **specific** **number** **of** **HTTP** **requests** **per** **frame** **(**REQ-029** **still** **applies**)** **beyond** **“reasonable”** **smooth** **updates** **documented** **in** **architecture**; **re-creating** **the** **three** **defaults** **automatically** **after** **the** **user** **deletes** **them** **(**deletion** **is** **allowed** **per** **REQ-022**; **only** **fresh** **install** **and** **factory** **reset** **re-seed** **the** **triple**)**.

**Business rules**

1. **On** **first** **start** **with** **an** **empty** **routine** **store** **and** **after** **confirmed** **factory** **reset**, **exactly** **three** **Python** **routine** **definitions** **MUST** **exist** **(**no** **fewer**, **no** **more** **at** **seed** **time**)** **with** **human-readable** **names** **and** **descriptions** **that** **let** **a** **user** **recognize** **(**a**)** **growing** **sphere**, **(**b**)** **sweeping** **cuboid**, **and** **(**c**)** **random** **colour** **cycle** **over** **all** **scene** **lights** **(**architecture** **may** **adjust** **exact** **strings** **but** **MUST** **preserve** **recognizability**)**.
2. **Each** **seeded** **routine** **MUST** **be** **valid** **Python** **code** **runnable** **against** **any** **scene** **that** **satisfies** **REQ-015**/**REQ-020** **(**subject** **to** **normal** **REQ-020** **validation**)**.
3. **Growing-sphere** **routine** **(**looping** **cycle**)** **:** **(**a**)** **Choose** **a** **new** **independent** **uniform** **random** **REQ-011**-valid **hex** **colour** **each** **cycle** **(**REQ-030** **helper** **SHOULD** **be** **used** **unless** **architecture** **documents** **a** **clear** **reason** **not** **to**)**. **(**b**)** **Place** **a** **sphere** **in** **scene** **space** **centered** **at** **the** **geometric** **center** **of** **the** **scene** **axis-aligned** **extent** **(**REQ-026**/**REQ-020**)**. **(**c**)** **The** **sphere** **starts** **with** **a** **small** **positive** **radius** **and** **over** **10** **SI** **seconds** **its** **radius** **increases** **monotonically** **until** **every** **light** **position** **in** **the** **scene** **lies** **inside** **or** **on** **the** **closed** **sphere** **(**REQ-020** **sphere** **semantics**)**. **(**d**)** **While** **the** **sphere** **grows**, **every** **light** **inside** **the** **current** **closed** **sphere** **MUST** **be** **on** **with** **brightness** **100** **percent** **and** **the** **cycle** **colour**. **(**e**)** **The** **routine** **MUST** **not** **turn** **off** **or** **recolour** **lights** **solely** **because** **they** **fall** **outside** **the** **sphere** **during** **growth** **(**lights** **not** **yet** **enclosed** **keep** **prior** **state**)**. **(**f**)** **After** **growth** **completes**, **the** **next** **cycle** **begins** **immediately** **with** **a** **new** **small** **sphere** **and** **new** **random** **colour** **while** **the** **run** **remains** **active** **(**REQ-022** **loop**)**.
4. **Sweeping-cuboid** **routine** **(**looping** **cycle**)** **:** **(**a**)** **Each** **cycle** **uses** **a** **new** **independent** **uniform** **random** **REQ-011**-valid **hex** **colour** **(**REQ-030** **helper** **SHOULD** **be** **used**)**. **(**b**)** **Cuboid** **width** **and** **depth** **equal** **scene** **width** **and** **depth** **(**REQ-026**/**REQ-020**)** **and** **height** **exactly** **0.2** **SI** **meters**. **(**c**)** **Start** **each** **cycle** **at** **the** **bottom** **of** **the** **scene** **volume** **(**minimum** **on** **the** **vertical** **axis** **documented** **for** **height**)** **spanning** **full** **width** **and** **depth**. **(**d**)** **Over** **10** **SI** **seconds** **translate** **monotonically** **to** **the** **top** **without** **leaving** **scene** **bounds**. **(**e**)** **At** **each** **update**, **lights** **inside** **or** **on** **the** **closed** **cuboid** **MUST** **be** **on** **at** **100** **percent** **with** **the** **cycle** **colour**. **(**f**)** **Lights** **that** **were** **inside** **on** **a** **prior** **update** **in** **the** **same** **cycle** **but** **are** **no** **longer** **inside** **MUST** **be** **set** **off** **(**REQ-011**)**. **(**g**)** **After** **reaching** **the** **top**, **start** **the** **next** **cycle** **at** **the** **bottom** **with** **a** **new** **random** **colour**.
5. **Random** **colour** **cycle** **—** **all** **scene** **lights** **routine** **(**successor** **to** **the** **former** **non-Python** **test** **routine** **behavior** **such** **as** **random_colour_cycle_all**)** **:** **When** **started** **on** **a** **scene**, **(**a**)** **every** **light** **in** **that** **scene** **MUST** **be** **set** **on** **(**`on`** **true**)** **with** **brightness** **100** **percent** **and** **REQ-011**-valid **hex** **colour**; **(**b**)** **thereafter**, **at** **most** **once** **per** **elapsed** **SI** **second** **while** **the** **run** **remains** **active**, **each** **light** **MUST** **receive** **a** **new** **hex** **colour** **chosen** **independently** **and** **uniformly** **at** **random** **from** **REQ-011**-valid **colours** **(**repeats** **allowed**)**. **The** **approximate** **one-second** **cadence** **MUST** **be** **documented** **in** **`docs/architecture.md`**. **Stopping** **MUST** **cease** **further** **automated** **updates** **promptly**; **lights** **retain** **last** **authoritative** **in-memory** **state** **(**REQ-011**, **REQ-039**)** **and** **stopping** **MUST** **not** **by** **itself** **reset** **to** **REQ-014** **defaults**.
6. **All** **three** **routines** **MUST** **NOT** **read** **or** **write** **canonical** **model** **coordinates** **directly**; **they** **MUST** **use** **scene-space** **operations** **only** **(**REQ-022**/**REQ-020**)**.
7. **`docs/architecture.md`** **MUST** **name** **where** **the** **initial** **seed** **content** **is** **defined** **(**e.g.** **embedded** **strings** **or** **shipped** **files**)** **and** **how** **those** **definitions** **relate** **to** **REQ-025** **(**new** **blank** **routine** **template**)** **if** **applicable**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Opening**, **running**, **editing**, **duplicating**, **or** **deleting** **these** **definitions** **MUST** **follow** **REQ-002** **and** **REQ-022** **expectations**.
- Tablet: **Same** **as** **mobile**.
- Desktop: **Same** **non-hover-only** **expectations** **for** **essential** **actions**.

**Dependencies**

- REQ-011, REQ-017, REQ-020, REQ-021, REQ-022, REQ-024, REQ-026, REQ-030

**Open questions**

- **Preferred** **temporal** **resolution** **(**number** **of** **updates** **across** **each** **10** **s** **cycle**)** **for** **the** **two** **geometric** **samples** **vs** **Pi**/**REQ-029** **load** **(**architecture** **trade-off**)**.

---

## REQ-033 — Shape animation routines (declarative scene-space shapes, light painting, run in scene)

| Field | Value |
|-------|-------|
| **ID** | REQ-033 |
| **Title** | Shape animation routines (declarative scene-space shapes, light painting, run in scene) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As an end user, I want to **define** **and** **edit** **shape** **animation** **routines** **with** **a** **name** **and** **description** **like** **Python** **routines**, **configure** **animated** **sphere** **or** **cuboid** **shapes** **with** **per-shape** **brightness** **and** **speed** **(**fixed** **or** **random** **within** **bounds**)** , **preview** **them** **with** **run** **on** **scene** **in** **the** **same** **unified** **viewport** **as** **Python** **authoring** **(**REQ-027**)** , **and** **have** **every** **light** **inside** **each** **shape** **show** **that** **shape’s** **colour** **and** **brightness** **while** **other** **lights** **show** **the** **background** **or** **turn** **off**, **so** **that** **I** **can** **create** **moving** **colour** **regions** **without** **writing** **code**.

**Scope**

- In scope: A **second** **routine** **kind**, **shape** **animation**, **alongside** **Python** **(**REQ-021**)**. **Persisted** **definition** **includes** **name** **(**required**)** , **description** **(**optional** **if** **architecture** **allows**)** , **and** **structured** **parameters** **below**. **Create**, **list**, **duplicate**, **edit**, **delete** **(**same** **lifecycle** **expectations** **as** **REQ-022** **where** **applicable** **except** **no** **Python** **buffer** **or** **REQ-024** **catalog**)**. **Start** **and** **stop** **a** **run** **against** **one** **scene** **(**REQ-021** **run** **lifecycle** **semantics**)**. **Add**/**edit** **surface** **MUST** **include** **the** **unified** **region** **specified** **in** **REQ-027** **:** **scene** **selection**, **run**/**stop**, **live** **three.js** **viewport** **for** **that** **scene**, **reset** **scene** **lights**, **reset** **camera** **—** **so** **the** **user** **can** **author** **and** **control** **runs** **while** **viewing** **the** **scene** **(**no** **duplicate** **scene** **picker**/**viewport** **for** **this** **workflow**)** **;** **the** **viewport** **observes** **server** **state** **(**REQ-041**)** **and** **MAY** **show** **non-authoritative** **ghost** **previews** **during** **editing** **only** **per** **architecture** **. **While** **a** **shape** **animation** **run** **is** **active**, **simulation** **and** **per-tick** **light** **assignment** **run** **entirely** **in** **the** **Go** **service** **(**REQ-021** **,** **REQ-038**)** **using** **the** **same** **scene** **spatial** **and** **bulk** **update** **semantics** **as** **documented** **in** **`docs/architecture.md`** **;** **the** **browser** **MUST** **NOT** **drive** **production** **ticks** **. **Canonical** **model** **coordinates** **MUST** **NOT** **be** **rewritten** **(**REQ-015**)**. **Stopping** **ends** **automation**; **lights** **keep** **last** **authoritative** **in-memory** **state** **(**REQ-021**, **REQ-039**)**. **Factory** **reset** **removes** **shape** **animation** **definitions** **(**REQ-017**)**.
- Out of scope: **Authoring** **shape** **animations** **in** **Python** **(**use** **REQ-022**)**; **new** **geometric** **primitives** **beyond** **sphere** **and** **axis-aligned** **cuboid** **for** **this** **kind**; **physics** **simulation** **beyond** **the** **stated** **edge** **behaviors**; **per-light** **assignment** **without** **going** **through** **scene-space** **volumes**.

**Business rules**

1. **Background:** The definition MUST allow **either** **(**a**)** **background** **appearance** **with** **REQ-011**-valid **hex** **colour** **and** **brightness** **percent** **(**0**–**100**)** **applied** **to** **every** **light** **not** **inside** **any** **active** **animated** **shape** **on** **that** **update** **tick** **(**lights** **on** **with** **that** **colour** **and** **brightness**)** , **or** **(**b**)** **a** **“no** **background** **lights”** **mode** **(**user-visible** **wording** **deferred** **to** **architecture**)** **where** **every** **light** **not** **inside** **any** **active** **shape** **is** **set** **off** **(**REQ-011** **off** **semantics**)**.
2. **Shapes** **list:** **Between** **1** **and** **20** **inclusive** **shapes** **per** **definition** **;** **each** **shape** **has** **type** **sphere** **or** **cuboid** **(**axis-aligned** **box** **in** **scene** **space**)**.
3. **Size:** **User** **supplies** **positive** **SI** **extent** **(**sphere** **diameter** **or** **radius** **—** **architecture** **picks** **one** **and** **documents** **it** **;** **cuboid** **width** **depth** **height** **along** **scene** **x** **y** **z** **per** **REQ-026**/**REQ-020**)** **either** **as** **fixed** **values** **or** **as** **independent** **uniform** **random** **values** **per** **shape** **instance** **between** **user-specified** **lower** **and** **upper** **bounds** **(**inclusive** **or** **half-open** **—** **architecture** **documents** **)** **on** **each** **run** **start** **and** **each** **loop** **cycle** **per** **rule** **10**.
4. **Shape** **colour** **and** **brightness:** **Colour** **is** **either** **fixed** **REQ-011**-valid **hex** **or** **“random”** **so** **that** **each** **time** **the** **shape** **colour** **is** **re-sampled** **(**e.g.** **new** **loop** **cycle** **—** **architecture** **ties** **to** **rule** **10**)** **it** **is** **a** **uniform** **random** **REQ-011**-valid **colour** **(**same** **distribution** **intent** **as** **REQ-030**)**. **Per** **shape**, **the** **user** **MUST** **configure** **brightness** **as** **a** **percent** **from** **0** **through** **100** **inclusive** **per** **REQ-011** **(**one** **fixed** **value** **per** **shape** **in** **the** **definition**)** **applied** **to** **every** **light** **painted** **by** **that** **shape**. **Lights** **inside** **the** **closed** **shape** **volume** **(**REQ-020** **inclusion** **semantics** **for** **sphere**/**cuboid**)** **MUST** **be** **on** **with** **that** **shape’s** **current** **hex** **colour** **and** **that** **brightness**.
5. **Initial** **position:** **Per** **shape**, **either** **(**a**)** **explicit** **scene-space** **coordinates** **(**center** **for** **sphere**;** **minimum** **corner** **or** **center** **—** **architecture** **documents** **one** **—** **for** **cuboid**)** **within** **the** **scene** **axis-aligned** **extent**, **or** **(**b**)** **random** **placement** **with** **the** **shape** **touching** **or** **contained** **against** **one** **user-chosen** **face** **of** **the** **scene** **bounding** **box** **:** **top**, **bottom**, **left**, **right**, **back**, **or** **front** **(**mapped** **to** **REQ-026** **width**/**depth**/**height** **faces** **in** **architecture**)** **such** **that** **the** **whole** **shape** **remains** **inside** **the** **closed** **scene** **volume** **when** **placement** **completes** **(**if** **impossible** **given** **size**, **validation** **MUST** **fail** **with** **clear** **errors** **at** **save** **or** **run** **per** **architecture**)**.
6. **Motion** **(**direction** **and** **scalar** **speed**)** **:** **Each** **shape** **moves** **in** **a** **straight** **line** **in** **scene** **space** **at** **constant** **scalar** **speed**. **The** **user** **configures** **(**a**)** **a** **direction** **via** **signed** **components** **dx**, **dy**, **dz** **(**not** **all** **zero**)** **that** **the** **implementation** **normalizes** **to** **a** **unit** **vector** **(**scene** **x**, **y**, **z**)** **and** **(**b**)** **a** **positive** **scalar** **speed** **stored** **and** **simulated** **in** **SI** **meters** **per** **second** **(**m/s**)** **—** **the** **UI** **MAY** **accept** **or** **display** **centimeters** **per** **second** **provided** **values** **convert** **consistently** **(**e.g.** **5** **cm/s** **=** **0.05** **m/s**)**. **Speed** **is** **either** **(**i**)** **a** **single** **fixed** **positive** **value** **or** **(**ii**)** **a** **uniform** **random** **value** **between** **user** **lower** **and** **upper** **bounds** **(**both** **positive** **;** **inclusive** **or** **half-open** **per** **architecture**)** **drawn** **on** **each** **run** **start** **and** **each** **loop** **cycle** **per** **rule** **10**. **Velocity** **is** **that** **unit** **direction** **multiplied** **by** **the** **current** **scalar** **speed**. **Architecture** **documents** **simulation** **timestep** **and** **relation** **to** **REQ-029**.
7. **Scene** **edges:** **The** **closed** **scene** **axis-aligned** **volume** **(**from** **REQ-020**/**REQ-026**)** **is** **the** **boundary**. **Per** **shape**, **when** **any** **part** **of** **the** **shape** **would** **cross** **outside** **that** **volume**, **one** **of** **these** **behaviors** **applies** **(**user** **selects** **per** **shape**)** :
   - **Pac-Man** **(**wrap**)** **:** **the** **shape** **re-enters** **from** **the** **opposite** **face** **so** **the** **solid** **volume** **stays** **inside** **the** **scene** **(**continuous** **motion** **—** **architecture** **defines** **exact** **wrap** **rule** **for** **sphere** **vs** **cuboid**)**.
   - **Stop** **and** **disappear** **:** **the** **shape** **is** **removed** **from** **the** **active** **set** **for** **the** **remainder** **of** **the** **run** **(**no** **longer** **paints** **lights**)** ; **it** **does** **not** **contribute** **to** **overlap** **resolution** **(**rule** **8**)**.
   - **Deflect** **random** **angle** **:** **velocity** **direction** **changes** **to** **a** **new** **random** **unit** **vector** **(**or** **equivalent** **uniform** **direction** **on** **the** **sphere** **of** **directions**)** **chosen** **independently** **at** **each** **deflection** **event** **;** **speed** **(**magnitude**)** **remains** **unchanged** **unless** **architecture** **documents** **otherwise**.
   - **Deflect** **inflection** **angle** **:** **specular** **reflection** **against** **the** **struck** **axis-aligned** **scene** **face** **(**angle** **of** **incidence** **equals** **angle** **of** **reflection** **in** **the** **plane** **normal** **to** **that** **face**)** **;** **architecture** **documents** **corner** **and** **simultaneous** **multi-face** **contact** **policy**.
8. **Overlap:** **When** **a** **light** **lies** **inside** **more** **than** **one** **active** **shape** **volume**, **the** **product** **MUST** **apply** **one** **deterministic** **precedence** **rule** **(**e.g.** **lowest** **shape** **index** **wins** **or** **highest** **z** **center** **—** **documented** **in** **`docs/architecture.md`** **)** **so** **colour** **assignment** **is** **stable** **and** **testable**.
9. **Light** **assignment** **each** **tick:** **After** **positions** **are** **updated**, **for** **every** **light** **in** **the** **scene** **:** **if** **it** **lies** **inside** **at** **least** **one** **active** **shape** **per** **rule** **8**, **set** **its** **state** **per** **rule** **4** **;** **else** **set** **per** **rule** **1** **(**background** **or** **off**)**.
10. **Loop** **vs** **stop:** **While** **the** **run** **is** **active**, **the** **animation** **MUST** **repeat** **in** **a** **loop** **(**reinitialize** **shapes** **per** **rules** **3**–**7** **and** **user** **config** **—** **architecture** **defines** **whether** **random** **choices** **are** **re-drawn** **each** **cycle**)** **until** **either** **(**a**)** **the** **user** **stops** **the** **run**, **or** **(**b**)** **every** **shape** **has** **reached** **a** **terminal** **state** **(**all** **“stop** **and** **disappear”** **with** **no** **active** **shapes** **remaining**)**. **If** **at** **least** **one** **shape** **uses** **Pac-Man**, **random** **deflect**, **or** **inflection** **deflect**, **or** **any** **behavior** **that** **keeps** **shapes** **active**, **the** **loop** **continues** **indefinitely** **until** **stop** **(**rule** **10b** **only** **applies** **when** **no** **active** **shapes** **remain**)**.
11. **Concurrency** **and** **validation:** **Invalid** **numeric** **ranges** **(**e.g.** **min** **>** **max** **for** **size** **or** **speed**)** , **non-positive** **speed** **or** **speed** **bounds**, **direction** **with** **all** **components** **zero**, **brightness** **outside** **0**–**100**, **empty** **shape** **list**, **more** **than** **20** **shapes**, **or** **non-finite** **parameters** **MUST** **be** **rejected** **at** **save** **or** **run** **with** **clear** **errors**. **Interaction** **with** **other** **active** **runs** **on** **the** **same** **scene** **follows** **REQ-021** **business** **rule** **5** **(**at** **most** **one** **active** **run** **per** **scene**)** **.
12. **`docs/architecture.md`** **MUST** **describe** **persistence** **shape** **for** **definitions**, **simulation** **tick**, **speed** **(**SI** **m/s** **and** **any** **cm/s** **UI** **mapping**)** , **wrap**/**deflect** **details**, **overlap** **precedence**, **and** **how** **REQ-027** **applies** **to** **this** **kind** **(**no** **Python** **editor** **or** **REQ-024** **on** **this** **surface** **unless** **optional** **cross-links** **only**)**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: **Parameter** **forms** **and** **REQ-027** **controls** **remain** **usable** **per** **REQ-002** **;** **adding**/**reordering** **up** **to** **20** **shapes** **must** **not** **rely** **on** **hover-only** **gestures**.
- Tablet: **Same** **as** **mobile** **;** **optional** **wider** **layout** **for** **shape** **list** **and** **viewport**.
- Desktop: **Efficient** **editing** **of** **many** **shapes** **;** **keyboard** **where** **architecture** **provides** **it**.

**Dependencies**

- REQ-002, REQ-011, REQ-015, REQ-017, REQ-020, REQ-021, REQ-026, REQ-027, REQ-029, REQ-030 (distribution intent for random colours), REQ-023, REQ-040

**Open questions**

- **Exact** **policy** **when** **two** **faces** **are** **hit** **in** **one** **tick** **(**inflection** **and** **random** **deflect**)**.

---

## REQ-034 — Three.js scene and model views: faint axis-aligned boundary cuboid (scene margin configurable, default 30 cm)

| Field | Value |
|-------|-------|
| **ID** | REQ-034 |
| **Title** | Three.js scene and model views: faint axis-aligned boundary cuboid (scene margin configurable, default 30 cm) |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user viewing a **model** or a **scene** in **three.js**, I want a **faint** outline of the **axis-aligned bounding cuboid** that **encloses** the relevant lights **plus** a **symmetric** **margin** on **every** **side**, so that I see **where** the **lit** **volume** **ends** **without** **cluttering** the view (**similar** **visual** **weight** **to** the **inter-light** **wire** **segments** **in** **REQ-010**). For **scenes**, I want that **margin** **size** **to** **be** **adjustable** **(**with** **a** **sensible** **default** **)** **so** **I** **can** **match** **how** **much** **“air”** **surrounds** **the** **models** **in** **the** **preview**.

**Scope**

- In scope: On **every** **three.js** **viewport** **that** **visualizes** **lights** **for** **a** **single** **model** **(**REQ-010**)** **or** **for** **a** **composite** **scene** **(**REQ-015**)** — **including** **embedded** **scene** **previews** **where** **architecture** **reuses** **the** **same** **canvas** **(**e.g.** **REQ-027** **unified** **run**/**watch** **region**, **REQ-033** **shape** **authoring** **viewport**)** — **the** **product** **MUST** **draw** **one** **closed** **cuboid** **(**12** **edges** **or** **equivalent** **wireframe** **box**)** **in** **the** **same** **coordinate** **space** **as** **the** **lights** **(**model** **local** **`x,y,z`** **for** **the** **model** **view**;** **derived** **scene-space** **`sx,sy,sz`** **for** **the** **scene** **view**)**. **Appearance** **MUST** **be** **subtle** **(**faint**)** **and** **MUST** **read** **as** **guidance** **only** **—** **not** **more** **prominent** **than** **the** **REQ-010** **inter-light** **segments** **(**architecture** **may** **reuse** **the** **same** **colour**/**opacity** **recipe** **or** **document** **a** **closely** **matched** **alternative**)**. **Geometry** **(**tight** **AABB** **from** **light** **extremes** **only**)** **is** **independent** **of** **whether** **underlying** **models** **are** **“regular”** **cuboids**. **Per** **REQ-015**, **each** **scene** **stores** **a** **user-adjustable** **boundary** **margin** **(**SI** **meters** **—** **same** **value** **applied** **on** **all** **six** **faces** **)** **with** **default** **0.3** **m** **(**30** **cm**)** **;** **the** **model** **view** **uses** **a** **fixed** **0.3** **m** **margin** **(**no** **per-model** **setting** **in** **this** **requirement** **unless** **added** **later**)**.
- Out of scope: **Per-face** **different** **margins** **(**one** **scalar** **per** **scene** **only**)**; **rotated** **(**non-axis-aligned**)** **bounding** **prisms**; **export** **or** **measurement** **tools**; **changing** **REQ-020** **API** **semantics** **unless** **architecture** **chooses** **to** **align** **server** **dimensions** **with** **this** **visual** **margin**.

**Business rules**

1. **Tight** **axis-aligned** **box** **from** **lights** **only** **:** **Compute** **per-axis** **minimum** **and** **maximum** **over** **all** **light** **positions** **used** **in** **that** **viewport** **(**for** **a** **model** **view** **:** **canonical** **`x,y,z`** **of** **every** **light** **in** **the** **model** **;** **for** **a** **scene** **view** **:** **`sx,sy,sz`** **of** **every** **light** **in** **the** **scene** **)**. **Ignore** **non-finite** **values** **(**treat** **as** **errors** **per** **architecture** **if** **they** **cannot** **occur** **)**. **If** **there** **are** **no** **lights**, **no** **boundary** **cuboid** **is** **required** **(**or** **architecture** **defines** **a** **degenerate** **case**)**.
2. **Padding** **(**model** **view**)** **:** **Expand** **the** **tight** **box** **by** **exactly** **0.3** **SI** **meters** **on** **each** **axis** **in** **both** **directions** **(**subtract** **0.3** **from** **each** **minimum**, **add** **0.3** **to** **each** **maximum** **after** **the** **tight** **AABB** **is** **known**)**.
3. **Padding** **(**scene** **view**)** **:** **Let** **`m`** **be** **the** **scene’s** **persisted** **boundary** **margin** **(**SI** **meters** **—** **REQ-015** **)**. **`m`** **MUST** **default** **to** **0.3** **when** **a** **scene** **is** **created** **or** **when** **legacy** **data** **has** **no** **stored** **value** **(**migration** **default** **0.3**)**. **`m`** **MUST** **be** **finite** **and** **≥** **0** **;** **upper** **cap** **(**if** **any**)** **is** **architecture** **(**e.g.** **to** **protect** **performance** **or** **framing**)**. **Expand** **the** **tight** **scene** **AABB** **by** **`m`** **on** **each** **axis** **in** **both** **directions** **(**same** **formula** **as** **rule** **2** **with** **`m`** **in** **place** **of** **0.3**)**. **Embedded** **previews** **that** **render** **a** **specific** **scene** **MUST** **use** **that** **scene’s** **current** **`m`**.
4. **Drawing** **:** **The** **boundary** **MUST** **appear** **whenever** **the** **viewport** **shows** **the** **corresponding** **model** **or** **scene** **(**including** **after** **camera** **reset** **per** **REQ-016**)** **and** **MUST** **update** **when** **light** **positions**, **scene** **composition**, **or** **(**for** **scenes**)** **the** **stored** **margin** **`m`** **changes** **without** **requiring** **a** **full** **page** **reload** **(**same** **timeliness** **class** **as** **REQ-012** **for** **state-driven** **rebuilds**)**.
5. **Prominence** **:** **The** **boundary** **MUST** **remain** **visually** **subtle** **(**faint**)** **—** **user** **can** **discern** **it** **against** **the** **REQ-019** **fixed** **dark-grey** **viewport** **but** **it** **MUST** **not** **dominate** **light** **spheres** **or** **the** **inter-light** **wire**.
6. **Same** **recipe** **as** **inter-light** **wire** **:** **The** **boundary** **cuboid** **edges** **(**12** **edges** **or** **equivalent**)** **MUST** **use** **the** **same** **colour**, **opacity**/**transparency**, **and** **line** **thickness** **policy** **as** **REQ-010** **straight** **segments** **between** **consecutive** **lights** **(**canonical** **:** **hex** **`#D0D0D0`**, **85%** **transparency** **meaning** **15%** **opacity**, **thin** **barely** **visible** **lines**)** **so** **the** **box** **reads** **as** **the** **same** **visual** **family** **as** **the** **wire** **(**architecture** **MAY** **document** **one** **shared** **three.js** **material** **or** **line** **style** **used** **for** **both**)**.
7. **Coordinate** **consistency** **:** **Scene** **boundary** **MUST** **use** **the** **same** **scene-space** **interpretation** **as** **REQ-015** **/** **REQ-020** **(**derived** **`sx`**, **`sy`**, **`sz`** **)** **so** **it** **aligns** **with** **where** **lights** **are** **drawn**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Boundary **visible** **on** **small** **canvases** **;** **does** **not** **require** **extra** **gestures** **to** **see**. **Scene** **margin** **control** **(**REQ-015**)** **MUST** **be** **usable** **without** **hover-only** **essential** **steps** **(**REQ-002**)**.
- Tablet: Same as mobile.
- Desktop: Same; **optional** **future** **toggle** **to** **hide** **boundary** **is** **out** **of** **scope** **unless** **added** **later**.

**Dependencies**

- REQ-002, REQ-010, REQ-012, REQ-015, REQ-019, REQ-016 (camera reset must not permanently remove boundary), REQ-027 (scene preview reuse), REQ-033 (shape routine scene viewport)

**Open questions**

- Whether **REQ-020** **`GET …/dimensions`** **should** **expose** **or** **derive** **from** **the** **same** **`m`** **(**architect** **decision** **—** **simulation** **vs** **pure** **visual** **)**.
- **Dashed** **vs** **solid** **edges** **for** **the** **cuboid** **(**if** **not** **using** **the** **same** **line** **primitive** **as** **REQ-010** **segments**)** **—** **architecture** **(**rule** **6** **fixes** **colour**/**opacity**/**thickness** **policy** **)**.

---

## REQ-035 — Physical lighting devices (ESP32 / WLED first; extensible types)

| Field | Value |
|-------|-------|
| **ID** | REQ-035 |
| **Title** | Physical lighting devices (ESP32 / WLED first; extensible types) |
| **Priority** | Must |
| **Actor(s)** | End user; operator; integrator |

**User story**

As an operator, I want the product to treat a **device** as a **physical** controller that drives real lights arranged according to a **model’s** light layout, starting with **ESP32**-based **WLED** controllers that support **individually addressable** strings, so that what I design in software can match what is installed—and I want the design to allow **other** device kinds later without rewriting the core concepts.

**Scope**

- In scope: Definition of **device** as a first-class concept: a network-reachable (or discoverable) unit that maps logical light indices **0 … n−1** for an assigned model to hardware outputs. **Initial** implementation target: **WLED** on **ESP32** (or architecture-documented equivalent) with per-segment or per-LED control consistent with REQ-005 ordering. **Device type** (e.g. WLED) is explicit so additional types can be added. Persisted **device registry** data (identity, type, connection parameters, user metadata such as name) as architecture defines—**excluding** per-light colour state (**REQ-039**).
- Out of scope: Mandating a specific WLED build, LED chipset, or wiring topology; certification of hardware; cloud-only bridges unless architecture chooses to support them.

**Business rules**

1. The product MUST represent devices with a **device type**; the first supported type MUST be **WLED** (ESP32-class) suitable for individually addressable strips or strings as architecture documents.
2. The architecture MUST describe how logical light **id** order (**REQ-005**) maps to WLED segments/LED indices and any grouping assumptions.
3. Additional device types MAY be introduced later; shared UX and APIs SHOULD treat **type** as an extension point (new type-specific configuration, same assignment and sync rules **REQ-036**–**REQ-039** where applicable).
4. Device **configuration** and **metadata** needed to reach the device (address, credentials if any, name) MAY be persisted; **light output state** MUST NOT be persisted as application durable state (**REQ-039**).
5. **MVP registration:** The product MUST support **adding** a device using **operator-supplied** **connection** **parameters** (e.g. **base** **URL**, **host**, **port**, **path** prefix—as **architecture** documents for **WLED**). **Automated** **network** **discovery** (**mDNS**, **broadcast**, **scan**) **MAY** ship in the **same** or a **later** release; **REQ-037** **still** **requires** a **manual** **add** **path** **so** **implementation** **is** **not** **blocked** **on** **discovery** **protocol** **choices**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Device list and detail remain usable per **REQ-002**; discovery results must be scannable without horizontal clutter.
- Tablet: Same; optional two-column layout for metadata vs actions.
- Desktop: Efficient editing of names and viewing connection details.

**Dependencies**

- REQ-001, REQ-005, REQ-006

**Open questions**

- None (**factory** **reset** **clears** **devices** **per** **REQ-017**; **MVP** **manual** **add** **per** **business** **rule** **5**; **optional** **discovery** **per** **architecture**).

---

## REQ-036 — Exclusive one-to-one device–model assignment

| Field | Value |
|-------|-------|
| **ID** | REQ-036 |
| **Title** | Exclusive one-to-one device–model assignment |
| **Priority** | Must |
| **Actor(s)** | End user; operator |

**User story**

As an operator, I want **at most one** device driving **one** model and **at most one** model per device at any time, while still allowing **models** and **devices** to exist **without** each other until I link them, so that I never get ambiguous ownership of lights or double-mapping.

**Scope**

- In scope: Assignment relation between a **device** and a **model** with **cardinality** **1:1** when linked: a device assigned to a model is the **sole** device for that model; that model is the **sole** model for that device. **Unassigned** models and **unassigned** devices are both valid. Operations to **assign** and **unassign** with validation when a conflicting assignment exists.
- Out of scope: Many-to-many or one-to-many fan-out from one model to multiple devices in this requirement (future extension only).

**Business rules**

1. At any time, a given **device** MUST be assigned to **at most one** model, or to **none**.
2. At any time, a given **model** MUST have **at most one** assigned device, or **none**.
3. Assigning device **D** to model **M** when **D** or **M** is already linked MUST either be **rejected** with a clear error or resolved through an **explicit** user flow (e.g. break existing link first)—**silent** reassignment of a third party’s link is **not** allowed.
4. **Scenes** (**REQ-015**) compose models; when a routine runs on a scene, each model instance in that scene uses the **model’s** assignment (**REQ-038**): if the model has a device, physical output for that model’s lights MUST follow the same logical state as the visualization for those lights.
5. **Assignment** **UX:** **Assign** and **unassign** MUST be available through the **Devices** **management** **flows** (**REQ-037**). A **model** **detail** **view** **SHOULD** **show** **whether** **a** **device** **is** **linked** **and** **MAY** **offer** **the** **same** **operations** **or** **a** **clear** **link** **to** **the** **device** **—** **architecture** **chooses** **surface(s)** **as** **long** **as** **REQ-037** **minimum** **is** **met** **and** **the** **API** **remains** **consistent**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Current assignment visible on model detail and device detail; unassign must be confirmable without hover-only steps (**REQ-002**).
- Tablet / Desktop: Same clarity for which model a device drives.

**Dependencies**

- REQ-006, REQ-015, REQ-035

**Open questions**

- None (**REQ-037** **+** **business** **rule** **5** **define** **surfaces**).

---

## REQ-037 — Device management UI: manual add, assign, optional discovery, metadata

| Field | Value |
|-------|-------|
| **ID** | REQ-037 |
| **Title** | Device management UI: manual add, assign, optional discovery, metadata |
| **Priority** | Must |
| **Actor(s)** | End user; operator |

**User story**

As an operator, I want a **dedicated** area in the app to **add** devices **with** **manual** **connection** **details** **(**required** **for** **MVP** **per** **REQ-035**)** **, **optionally** **discover** them on the network **when** **architecture** **ships** **that** **feature** **, **connect** them to a model or **disconnect**, and **view**/**edit** metadata such as a **friendly name**, so that I can manage hardware without treating devices as a hidden integrator-only feature.

**Scope**

- In scope: A **new** primary **section** in the shell navigation (or equivalent discoverable entry per **REQ-018** patterns) for **Devices**. Flows: **list** known devices; **add** device (**manual** **connection** **parameters** **MUST** **be** **supported** **per** **REQ-035** **business** **rule** **5**; **automated** **discovery** **and** **“refresh** **discovery”** **MAY** **be** **added** **when** **architecture** **implements** **them**); **remove** or **forget** device from registry; **assign**/**unassign** model (**REQ-036**); **view** and **edit** persisted metadata (**name** minimum). When **discovery** exists, it SHOULD surface candidates on the local network compatible with **REQ-035**.
- Out of scope: Pixel-perfect parity with vendor-native WLED apps; full OTA firmware management.

**Business rules**

1. The UI MUST expose **Devices** as a **first-class** destination alongside existing major areas (models, scenes, routines, options).
2. The user MUST be able to **add** a device using **manual** **connection** **parameters** (**REQ-035**). When **architecture** provides **discovery**, the user MUST be able to **initiate** or **refresh** it and **add** a chosen candidate; until then, **manual** **add** **alone** **satisfies** **this** **rule**.
3. The user MUST be able to **assign** the device to a model and **unassign** it, obeying **REQ-036**.
4. The user MUST be able to **view** and **edit** at least **device name** (and architecture may add type-specific fields).
5. Essential actions MUST follow **REQ-002** (no hover-only critical steps). **Priority:** device UI polish is **secondary** to correct **routine**/**device** sync (**REQ-039**); MVP may be minimal if documented.
6. **Removing** a **device** from the **registry** (**delete**/**forget**) MUST **clear** **any** **model** **assignment** **for** **that** **device** **in** **the** **same** **logical** **operation** (**no** **orphaned** **assignment** **rows**), **then** **remove** **the** **device** **record** **per** **architecture** **transaction** **rules**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Stacked list, full-width forms, clear assign/unassign affordances.
- Tablet: List + detail split where space allows.
- Desktop: Same information hierarchy; keyboard-friendly naming fields.

**Dependencies**

- REQ-002, REQ-018, REQ-035, REQ-036

**Open questions**

- None (**delete** **cascades** **per** **business** **rule** **6**; **type-specific** **secrets** **per** **architecture**).

---

## REQ-038 — Routine runs synchronize visualization and physical lights; server-side headless execution

| Field | Value |
|-------|-------|
| **ID** | REQ-038 |
| **Title** | Routine runs synchronize visualization and physical lights; server-side headless execution |
| **Priority** | Must |
| **Actor(s)** | End user; operator; integrator |

**User story**

As an operator, I want a **routine** running on a **scene** to update **both** the on-screen lights and any **physical** lights for models that have devices, in **lockstep** with the same logical state, and I want **routine** **automation** **to** **keep** **running** **without** **any** **browser** **open** **(**headless** **on** **the** **Pi** **or** **via** **integrator** **`curl`**) **so** **that** **installations** **do** **not** **depend** **on** **a** **workstation** **tab** **. **The** **browser** **must** **only** **be** **a** **view** **onto** **the** **same** **authoritative** **state** **the** **server** **already** **holds** **(**REQ-039** **,** **REQ-041**)** **—** **it** **must** **not** **be** **the** **execution** **host** **for** **routine** **loops** **(**REQ-021**)** **.

**Scope**

- In scope: When a **routine run** (**REQ-021**, **REQ-033**) is **active** on a **scene**, every update to logical per-light state for lights belonging to a **model** that has an **assigned device** MUST be **applied** to that **device** as architecture defines (WLED API or equivalent), in addition to updating **authoritative** state consumed by the UI (**REQ-039**). **Routine** **execution** **MUST** **run** **inside** **the** **single** **Go** **service** **process** **;** **`docs/architecture.md`** **MUST** **document** **packages**, **scheduling**, **and** **how** **Python** **and** **shape** **kinds** **meet** **REQ-021** **without** **browser-hosted** **production** **loops** **. **Browsers** **MAY** **subscribe** **(**REQ-041**)** **or** **poll** **for** **observability** **but** **MUST** **NOT** **be** **required** **for** **the** **run** **to** **progress** **or** **for** **light** **mutations** **to** **continue** **. **Per-model** mapping: scene placement does not change device mapping (**REQ-015**); device maps **model local** light indices **0…n−1** to hardware.
- Out of scope: Guaranteeing sub-millisecond sync between LED output and GPU frame; offline device behavior beyond documented error/retry policy.

**Business rules**

1. **Server-side execution:** Starting a routine (**REQ-021**) MUST cause **ongoing** **server-side** **automation** **on** **the** **backend** **until** **stop** **or** **failure** **,** **independent** **of** **any** **web** **client** **session** **(**process** **model** **per** **`docs/architecture.md`** **and** **REQ-021**)** **.
2. **Physical sync:** For each **model** in the **scene** with an **assigned** **device**, state changes from the routine MUST be **reflected** on the **device** for all affected logical lights, subject to **REQ-039** in-memory authority and **REQ-031**-style **no-op** suppression where architecture applies.
3. **Visualization sync:** Connected clients showing the **same** scene or model MUST observe state consistent with **REQ-012** / **REQ-015** for logical lights (timeliness class as architecture documents under **REQ-029**).
4. **Models without devices:** Routine behavior is **unchanged** for logical state; no physical traffic is required.
5. **Concurrency:** **REQ-021** **business** **rule** **5** **(**at** **most** **one** **active** **run** **per** **scene**)** **is** **normative** **;** **architecture** **MUST** **document** **how** **manual** **API** **writes** **interact** **with** **an** **active** **run** **(**e.g.** **last-writer** **,** **rejected** **writes** **)** **so** **behavior** **is** **deterministic** **and** **not** **silent** **corruption** **.
6. **Headless** **control:** **Routine** **start** **and** **stop** **MUST** **be** **available** **through** **the** **same** **documented** **`/api/v1`** **HTTP** **surface** **the** **web** **UI** **uses**, **unless** **architecture** **adds** **an** **explicit** **documented** **alias** **for** **integrators** **with** **the** **same** **semantics** **.
7. **Browser** **role:** **The** **product** **MUST** **NOT** **use** **the** **browser** **to** **execute** **production** **routine** **bodies** **(**including** **Pyodide** **workers** **or** **client** **timers** **that** **apply** **routine** **effects** **)** **;** **authoring-time** **lint** **/** **format** **in** **the** **browser** **is** **allowed** **only** **where** **architecture** **explicitly** **scopes** **it** **as** **non-authoritative** **(**REQ-022**)** **.

**Responsive / UX notes** *(when UI is involved)*

- Mobile / Tablet / Desktop: Start/stop and status MUST remain usable per **REQ-002** when exposed; **headless** operation MUST not depend on these surfaces.

**Dependencies**

- REQ-003, REQ-015, REQ-020, REQ-021, REQ-022, REQ-029, REQ-031, REQ-035, REQ-036, REQ-039, REQ-040, REQ-041

**Open questions**

- None (**headless** **API** **per** **business** **rule** **6**; **WLED** **rate**/**batch** **limits** **documented** **under** **REQ-029**/**REQ-031**/**architecture**).

---

## REQ-039 — Authoritative in-memory light state; sync at startup and on assignment

| Field | Value |
|-------|-------|
| **ID** | REQ-039 |
| **Title** | Authoritative in-memory light state; sync at startup and on assignment |
| **Priority** | Must |
| **Actor(s)** | End user; operator; integrator |

**User story**

As an integrator, I want the app to treat **light output state** as **runtime** data held **in memory** for **performance**, **not** stored in the database, with clear rules for **initial** state at **startup** and when I **link** a device, so that the system stays fast and the **truth** for pixels lives in **RAM** (and on the **wire**), not in SQLite.

**Scope**

- In scope: **Per-light** operational state (**REQ-011** fields) is **authoritative** in **server memory** for the running process. It MUST **not** be written to **durable** application storage for reload after restart. The service SHOULD maintain an **in-memory** representation (per model or global structure per architecture) to serve reads and drive pushes to devices **without** re-querying hardware for full state on every operation. **Synchronization** events: **(a)** **process start**: for each model **with** an assigned device, perform a **sync** with that device as architecture defines (direction: align logical state to device, or push defaults then device follows—**documented**); **(b)** **first association** of a device to a model: sync **when the link is created**; **(c)** models **without** a device: initial logical state is **all lights off** with **REQ-014** default colour/brightness semantics for the **off** appearance. **Product priority**: **correctness** and **throughput** of **routine-driven** updates and **device** output are **primary**; **Devices** UI (**REQ-037**) and **authoring** UIs are **secondary**—the product is intended to run **headless** with the UI as an **occasional** **read**/**configure** **surface** **only** **;** **routine** **execution** **never** **depends** **on** **the** **browser** **(**REQ-038** **,** **REQ-021**)** **.
- Out of scope: Long-term **analytics** or **history** of light states in the app database.

**Business rules**

1. **No durable light state:** The application MUST **not** persist per-light **on**/**off**, **colour**, or **brightness** in **SQLite** or equivalent **application** store for **reload** after **restart** (**REQ-011** rule **5** as updated).
2. **In-memory authority:** All **REQ-011** read APIs MUST reflect the **current** **in-memory** authoritative state for the running server (plus documented **consistency** if scaled-out is ever introduced—default single instance).
3. **Startup:** On service start, for each model **without** an assigned device, authoritative state MUST be initialized to **all lights off** per **REQ-014** defaults. For each model **with** an assigned device, a **sync** MUST run per architecture (read from device, or push initial pattern—**documented**).
4. **On assignment:** When a device is **newly** assigned to a model, a **sync** MUST run **immediately** after the link succeeds (**same** architectural meaning as rule **3**).
5. **Performance:** Architecture MUST describe how in-memory structures support **REQ-029** and **REQ-031** (last-known, equivalence, push coalescing).
6. **Priority:** Implementation and verification MUST favor **routine**/**device** **correctness** and **performance** over **Devices** UI depth and over **visual** **polish** of **occasional** configuration flows, without violating **REQ-002** **baseline** usability where UI exists.
7. **Startup** **vs** **device-reported** **state:** If **startup** **sync** **observes** **per-light** **state** **on** **the** **device** **that** **differs** **from** **REQ-014** **defaults**, **architecture** **MUST** **document** **one** **consistent** **policy** (**examples:** **treat** **device** **as** **initial** **authority** **for** **logical** **state**; **or** **initialize** **to** **REQ-014** **defaults** **in** **memory** **then** **push** **to** **the** **device**). **Implementations** **MUST** **follow** **that** **documented** **policy**.
8. **Unassign** **behavior:** **Architecture** **MUST** **document** **what** **happens** **to** **physical** **outputs** **when** **a** **device** **is** **unassigned** **from** **a** **model** (**e.g.** **leave** **hardware** **unchanged**, **push** **all** **off**, **or** **other** **explicit** **behavior**) **so** **operators** **can** **predict** **installed** **effects**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile / Tablet / Desktop: **REQ-037** and routine authoring remain **basically** usable (**REQ-002**); **polish** and **feature** **breadth** for those surfaces are **explicitly** lower priority than **REQ-038**/**REQ-039** server and device behavior.

**Dependencies**

- REQ-011, REQ-014, REQ-029, REQ-031, REQ-035, REQ-036

**Open questions**

- None (**policies** **for** **rules** **7**–**8** **are** **architecture** **documentation** **obligations**, **not** **open** **product** **questions**).

---

## REQ-040 — Routine stop completes within two seconds

| Field | Value |
|-------|-------|
| **ID** | REQ-040 |
| **Title** | Routine stop completes within two seconds |
| **Priority** | Must |
| **Actor(s)** | End user; operator; integrator |

**User story**

As an operator, when I **stop** an **active** **routine** **run** (**Python** **or** **shape** **animation**), I want **automation** **to** **end** **quickly** **(**within** **about** **two** **seconds**)** **so** **that** **lights** **do** **not** **keep** **changing** **unexpectedly** **after** **I** **have** **requested** **stop**.

**Scope**

- In scope: **Wall-clock** **bound** **from** **a** **successful** **stop** **request** **(**UI** **or** **documented** **HTTP** **API** **equivalent** **to** **`POST …/routines/runs/…/stop`** **per** **REQ-021**/**architecture**)** **until** **the** **run** **is** **fully** **stopped** **:** **no** **further** **routine-driven** **light-state** **updates** **may** **be** **applied** **for** **that** **run** **,** **and** **persisted** **run** **status** **reflects** **completion** **(**within** **normal** **process** **health**)**. **Applies** **to** **both** **routine** **kinds** **(**REQ-022**,** **REQ-033**)** **and** **must** **hold** **whether** **or** **not** **a** **browser** **is** **open** **(**REQ-038**)**.
- Out of scope: **Guarantees** **when** **the** **OS** **kills** **the** **process**, **disk** **full**, **or** **other** **catastrophic** **failures** **;** **sub-second** **latency** **SLOs** **stricter** **than** **2** **s** **unless** **added** **later** **;** **undo** **of** **already-applied** **light** **state** **at** **stop** **(**REQ-032** **preserves** **last** **state**)**.

**Business rules**

1. **Two** **second** **maximum** **(**normal** **conditions**)** **:** **After** **the** **server** **accepts** **a** **stop** **request** **for** **an** **active** **run**, **the** **implementation** **MUST** **reach** **a** **state** **where** **that** **run** **performs** **no** **further** **iterations**, **ticks**, **or** **Python** **loop** **bodies** **and** **emits** **no** **further** **routine** **originated** **updates** **to** **authoritative** **light** **state** **(**REQ-039**)** **within** **at** **most** **2** **SI** **seconds** **(**inclusive** **of** **cooperative** **shutdown** **and** **any** **forcible** **termination** **path** **per** **REQ-022** **business** **rule** **8**)**.
2. **`docs/architecture.md` MUST** **describe** **how** **the** **supervised** **Python** **routine** **runtime** **and** **the** **shape** **animation** **engine** **meet** **rule** **1**, **including** **cooperative** **stop**, **forcible** **termination** **when** **needed**, **and** **cancelling** **in-flight** **work** **before** **further** **routine-originated** **light-state** **commits** **;** **aggregate** **timeouts** **must** **not** **exceed** **the** **2** **s** **budget** **unless** **architecture** **documents** **a** **narrow** **exception** **with** **tests** **.**
3. **API** **and** **UI** **:** **The** **same** **stop** **semantics** **apply** **whether** **stop** **is** **invoked** **from** **the** **web** **UI** **or** **from** **the** **HTTP** **API** **(**REQ-038** **business** **rule** **6**)**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile / Tablet / Desktop: **Stop** **remains** **reachable** **per** **REQ-002** **;** **REQ-040** **does** **not** **add** **new** **UI** **surfaces**, **only** **timing** **behavior** **after** **stop** **is** **activated**.

**Dependencies**

- REQ-021, REQ-022, REQ-033, REQ-038, REQ-039

**Open questions**

- None.

---

## REQ-041 — Server-push visualization updates with incremental client apply

| Field | Value |
|-------|-------|
| **ID** | REQ-041 |
| **Title** | Server-push visualization updates with incremental client apply |
| **Priority** | Must |
| **Actor(s)** | End user; integrator |

**User story**

As a user watching **live** **light** **state** **in** **a** **three.js** **viewport**, I want **updates** **to** **arrive** **via** **a** **push** **channel** **(**not** **by** **the** **browser** **repeatedly** **polling** **the** **server** **in** **a** **tight** **loop**)** **and** **I** **want** **only** **what** **changed** **to** **be** **applied** **in** **the** **scene** **graph** **so** **the** **view** **stays** **smooth** **and** **the** **Pi** **is** **not** **overloaded** **with** **redundant** **work**.

**Scope**

- In scope: **All** **shipped** **web** **three.js** **experiences** **that** **display** **authoritative** **per-light** **state** **while** **that** **state** **may** **change** **at** **runtime** **(**including** **at** **minimum** **the** **single-model** **view** **REQ-010**, **the** **scene** **composite** **view** **REQ-015**, **and** **the** **unified** **routine** **live** **viewport** **REQ-027** **when** **it** **shows** **live** **state**)** **must** **use** **a** **server-initiated** **streaming** **or** **push** **mechanism** **(**Server-Sent** **Events**, **WebSocket**, **or** **architecturally** **equivalent** **HTTP** **streaming** **/** **long-lived** **subscription** **)** **as** **the** **primary** **means** **of** **learning** **about** **state** **changes** **after** **the** **initial** **snapshot** **load**, **instead** **of** **relying** **on** **high-frequency** **repeated** **polling** **(**e.g.** **many** **requests** **per** **second** **per** **open** **viewport**)**. **Payloads** **MUST** **be** **expressible** **as** **deltas** **(**only** **lights** **whose** **effective** **on**/**off**, **colour**, **or** **brightness** **changed** **)** **when** **the** **change** **set** **is** **partial** **;** **the** **client** **MUST** **apply** **those** **deltas** **by** **updating** **existing** **three.js** **objects** **(**materials**/**meshes** **per** **light** **or** **documented** **equivalent**)** **without** **discarding** **and** **recreating** **the** **entire** **scene** **graph** **on** **every** **message** **when** **only** **a** **subset** **of** **lights** **changed**. **Initial** **load** **and** **reconnect** **after** **connection** **loss** **MAY** **use** **a** **full** **state** **fetch** **per** **architecture** **;** **REQ-031** **equivalence** **and** **skip** **rules** **remain** **in** **force** **on** **the** **client**.
- Out of scope: **Mandating** **one** **named** **protocol** **when** **another** **meets** **push** **+** **delta** **semantics** **;** **binary** **vs** **JSON** **encoding** **;** **third-party** **integrator** **clients** **outside** **the** **shipped** **web** **UI** **(**REQ-029** **still** **encourages** **documented** **observer** **paths** **for** **them**)**.

**Business rules**

1. **No** **polling** **hammer** **:** **For** **each** **open** **live** **viewport** **described** **in** **scope**, **the** **client** **MUST** **not** **use** **a** **timer-driven** **or** **tight** **loop** **that** **issues** **HTTP** **GET** **(**or** **equivalent** **read** **requests**)** **at** **a** **rate** **whose** **primary** **purpose** **is** **to** **poll** **full** **light** **state** **(**architecture** **documents** **an** **upper** **bound** **or** **defines** **“high-frequency”** **for** **verification** **—** **e.g.** **no** **more** **than** **one** **full** **state** **poll** **per** **few** **seconds** **while** **push** **is** **healthy**)** **;** **push** **is** **the** **primary** **path** **for** **ongoing** **changes**.
2. **Delta** **payloads** **:** **When** **a** **single** **logical** **change** **event** **affects** **only** **a** **proper** **subset** **of** **lights** **in** **the** **viewport**, **the** **server** **MUST** **send** **a** **representation** **that** **lists** **only** **affected** **lights** **(**ids** **+** **new** **state** **fields** **or** **documented** **equivalent**)** **rather** **than** **requiring** **the** **client** **to** **replace** **full** **state** **for** **all** **lights** **to** **interpret** **that** **event**.
3. **Incremental** **apply** **:** **The** **client** **MUST** **merge** **incoming** **deltas** **into** **the** **existing** **three.js** **scene** **for** **affected** **lights** **only** **(**update** **materials**/**colours**/**visibility** **per** **REQ-010**/**REQ-012**/**REQ-028**)** **and** **MUST** **not** **rebuild** **the** **full** **mesh**/**line** **graph** **from** **scratch** **on** **every** **delta** **message** **unless** **architecture** **documents** **that** **the** **event** **semantically** **requires** **a** **full** **rebuild** **(**e.g.** **model** **or** **scene** **identity** **changed**)**.
4. **`docs/architecture.md` MUST** **name** **the** **chosen** **push** **transport** **(**SSE** **vs** **WebSocket** **vs** **other**)** **,** **the** **URL** **shape** **or** **subscription** **pattern**, **delta** **message** **schema**, **reconnect** **and** **full** **resync** **behavior**, **and** **how** **this** **interacts** **with** **REQ-029** **and** **REQ-031**.
5. **Pi** **:** **The** **design** **MUST** **remain** **plausible** **on** **Raspberry** **Pi** **4** **constraints** **(**REQ-003**)** **(**connection** **count**, **message** **rate**, **JSON** **size**)**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile / Tablet / Desktop: **Live** **viewports** **remain** **usable** **per** **REQ-002** **;** **push** **failures** **should** **degrade** **to** **documented** **slower** **polling** **or** **user-visible** **stale** **state** **indicators** **per** **architecture**, **not** **silent** **infinite** **hammer**.

**Dependencies**

- REQ-002, REQ-003, REQ-010, REQ-012, REQ-015, REQ-027, REQ-029, REQ-031, REQ-039

**Open questions**

- **Whether** **multiple** **tabs** **share** **one** **push** **connection** **or** **one** **per** **tab** **(**architecture** **/** **REQ-029** **alignment**)**.

---

## REQ-042 — Routine run visibility and visualization resync when returning to a scene

| Field | Value |
|-------|-------|
| **ID** | REQ-042 |
| **Title** | Routine run visibility and visualization resync when returning to a scene |
| **Priority** | Must |
| **Actor(s)** | End user |

**User story**

As a user managing **routines** on a **scene**, I want the **web UI** to **show** whether **that** **scene** **already** **has** **an** **active** **routine** **run**, and I want the **three.js** **visualization** **on** **the** **scene** **detail** **page** **and** **on** **routine** **authoring** **pages** **(Python** **and** **shape**)** **to** **reflect** **the** **current** **authoritative** **light** **state**. If I **navigate** **away** **and** **later** **return** **while** **a** **run** **is** **still** **active** **on** **the** **server**, I want **the** **UI** **and** **viewport** **to** **resynchronize** **with** **backend** **state** **so** **I** **do** **not** **see** **a** **stale** **“not** **running”** **or** **frozen** **lights** **view**.

**Scope**

- In scope: **Client** **behavior** **for** **(**a**)** **detecting** **and** **displaying** **whether** **a** **given** **scene** **has** **a** **running** **routine** **(**per** **documented** **HTTP** **API** **for** **listing** **runs** **on** **a** **scene**, **REQ-021**)** **on** **the** **scene** **detail** **page** **and** **on** **the** **unified** **routine** **authoring** **surfaces** **(REQ-027**)** **;** **(**b**)** **keeping** **the** **live** **scene** **three.js** **view** **aligned** **with** **authoritative** **per-light** **state** **for** **that** **scene** **while** **the** **page** **is** **open** **(REQ-041** **push** **+** **deltas**, **REQ-039**)** **;** **(**c**)** **after** **a** **full** **navigation** **away** **and** **back** **(**or** **equivalent** **route** **remount**)** **while** **a** **routine** **run** **remains** **active**, **re-loading** **run** **status** **from** **the** **server** **and** **re-establishing** **live** **visualization** **so** **the** **displayed** **run** **state** **and** **light** **colours**/**brightness**/**on-off** **match** **the** **server** **(**initial** **scene** **fetch**, **re-subscription** **to** **push**, **or** **architecture-documented** **equivalent** **—** **including** **degraded** **polling** **when** **push** **is** **unhealthy** **per** **REQ-041** **responsive** **notes**)**.
- Out of scope: **New** **routine** **types** **;** **persisting** **UI-only** **flags** **that** **override** **server** **truth** **;** **cross-tab** **synchronization** **beyond** **what** **the** **server** **and** **documented** **APIs** **already** **provide**.

**Business rules**

1. **Run** **detection** **:** **The** **product** **MUST** **use** **server-reported** **routine** **run** **state** **(**e.g.** **GET** **`/api/v1/scenes/{id}/routines/runs`** **or** **documented** **successor** **per** **REQ-021**/**architecture**)** **to** **determine** **whether** **a** **scene** **has** **an** **active** **run** **and** **which** **routine** **it** **is**, **and** **MUST** **reflect** **that** **in** **the** **UI** **on** **the** **scene** **detail** **page** **and** **on** **Python** **and** **shape** **routine** **pages** **when** **a** **scene** **is** **selected** **(**so** **the** **user** **can** **start**, **stop**, **or** **see** **that** **a** **run** **is** **already** **in** **progress**)**.
2. **Live** **visualization** **:** **On** **those** **same** **pages**, **the** **three.js** **scene** **viewport** **MUST** **track** **current** **authoritative** **light** **state** **for** **the** **selected** **scene** **(**REQ-012**,** **REQ-015**,** **REQ-039**)** **using** **the** **push** **and** **delta** **rules** **in** **REQ-041** **while** **the** **connection** **is** **healthy** **;** **when** **push** **is** **not** **available**, **the** **product** **MUST** **fall** **back** **to** **documented** **slower** **refresh** **or** **full** **resync** **so** **the** **view** **does** **not** **remain** **indefinitely** **stale** **(**REQ-041** **responsive** **notes**)**.
3. **Return** **navigation** **:** **When** **the** **user** **opens** **again** **the** **scene** **detail** **page** **or** **a** **routine** **authoring** **page** **for** **a** **scene** **that** **still** **has** **an** **active** **run** **on** **the** **server**, **the** **UI** **MUST** **re-fetch** **(**or** **equivalent**)** **run** **status** **and** **scene**/**light** **state** **so** **that** **(**a**)** **running** **vs** **stopped** **controls** **and** **labels** **match** **the** **server**, **and** **(**b**)** **the** **viewport** **shows** **the** **current** **lights** **(**not** **only** **the** **state** **from** **the** **previous** **visit**)**.
4. **Conflict** **on** **start** **:** **If** **the** **user** **attempts** **to** **start** **a** **second** **run** **while** **one** **is** **already** **active** **(**HTTP** **409** **per** **REQ-021**)** **,** **the** **UI** **MUST** **recover** **to** **a** **consistent** **state** **(**e.g.** **surface** **the** **existing** **run** **id** **from** **the** **error** **payload** **or** **refresh** **runs** **)** **so** **the** **user** **can** **stop** **before** **retrying** **—** **documented** **in** **architecture** **if** **multiple** **patterns** **are** **allowed**.

**Responsive / UX notes** *(when UI is involved)*

- Mobile / Tablet / Desktop: **Run** **status** **(**running** **/** **not** **)** **and** **stop** **remain** **discoverable** **per** **REQ-002** **;** **no** **hover-only** **dependency** **for** **knowing** **a** **scene** **is** **busy**.

**Dependencies**

- REQ-002, REQ-015, REQ-021, REQ-027, REQ-038, REQ-039, REQ-041

**Open questions**

- None

---

## REQ-043 — Cross-platform single Go executables (Windows, Linux, Linux ARM)

| Field | Value |
|-------|-------|
| **ID** | REQ-043 |
| **Title** | Cross-platform single Go executables (Windows, Linux, Linux ARM) |
| **Priority** | Must |
| **Actor(s)** | Operator / maintainer; integrator |

**User story**

As an operator, I want the product built and published as **one Go executable per major deployment platform** (Windows amd64, Linux amd64, and Linux ARM—for example arm64 for Raspberry Pi), so that I can download the right binary for my machine without building from source.

**Scope**

- In scope: At least three release targets: **windows/amd64**, **linux/amd64**, and **linux/arm64** (GOOS/GOARCH names or equivalent in documentation). Each target is one application binary that satisfies REQ-004 for that platform. Raspberry Pi 4 aligns with **linux/arm64** (REQ-003). **linux/arm** (v7) MAY be added in architecture if needed; the MUST set is the three above.
- In scope: `docs/architecture.md` MUST list the canonical GOOS/GOARCH pairs, artifact naming (file name per target), and how the static UI is embedded in the same binary.
- Out of scope: macOS binaries unless a later requirement adds them; OS-specific code signing or notarization policies unless captured later.

**Business rules**

1. For each canonical release target, distribution MUST be exactly one product executable file per logical version without requiring a separate Node.js runtime or second application daemon for routine operation, consistent with REQ-004.
2. Built artifacts for each listed target MUST be produced by documented automated means (REQ-044) so operators need not invoke `go` cross-compilation manually for standard downloads.
3. `docs/architecture.md` MUST cross-reference REQ-003 (Pi) and REQ-004 so deployment notes name which published artifact maps to Raspberry Pi (linux/arm64) and which to desktop Windows or Linux.

**Responsive / UX notes** *(when UI is involved)*

- N/A at requirements level; README coverage is REQ-046.

**Dependencies**

- REQ-003, REQ-004

**Open questions**

- Whether **linux/arm** (v7) must be first-class alongside arm64 for older Pi boards.

---

## REQ-044 — GitHub Actions: CI (build + test gates) and release binaries

| Field | Value |
|-------|-------|
| **ID** | REQ-044 |
| **Title** | GitHub Actions: CI (build + test gates) and release binaries |
| **Priority** | Must |
| **Actor(s)** | Maintainer; operator |

**User story**

As a maintainer, I want **continuous integration in GitHub Actions** that **builds and tests** the repository on **pull requests**, with those checks **required before merge**, and I want a **release workflow** that **tags** the repository and **publishes downloadable binaries** for each supported platform.

**Scope**

- In scope: GitHub Actions workflows that (a) run an automated **build and test** pipeline (Go tests, frontend lint/tests as defined for the repo) on pull requests (and typically on pushes to main); (b) align with repository policy so **merge to the protected branch is blocked** until those checks pass (branch protection / required status checks); (c) provide a **release process** (manual or tag-triggered per architecture) that creates a Git tag, associates release metadata as appropriate, and uploads per-platform binaries from REQ-043 as GitHub Release assets (or documented equivalent download surface).
- Out of scope: Exact GitHub billing or org policy beyond documenting that required checks must pass before merge; third-party release hosts other than GitHub Releases unless architecture adds mirrors.

**Business rules**

1. **PR gate:** Submitting or merging must not proceed per repo policy when the documented CI build-and-test workflow fails (for example `go test ./...`, `npm run lint`, `npm test`, or equivalents aligned with project scripts).
2. **Release:** A documented procedure (in `docs/advanced-setup.md` or `README.md` as appropriate for audience) MUST describe how to cut a version (tag) such that publish artifacts for windows/amd64, linux/amd64, and linux/arm64 attach to that release.
3. `docs/architecture.md` MUST summarize CI and release (workflow names, trigger events, artifact outputs) after the architect pass.

**Responsive / UX notes** *(when UI is involved)*

- N/A

**Dependencies**

- REQ-043

**Open questions**

- Semantic versioning vs calendar tags for release names.

---

## REQ-045 — Deployment runtime prerequisites (Python for custom routines only)

| Field | Value |
|-------|-------|
| **ID** | REQ-045 |
| **Title** | Deployment runtime prerequisites (Python for custom routines only) |
| **Priority** | Must |
| **Actor(s)** | Operator / maintainer |

**User story**

As an operator, I want to run the published application on the deployment machine **without installing extra prerequisites** beyond what the OS already provides, **except** that **Python** must be available when I use **custom user-authored Python routines** (REQ-021, REQ-022).

**Scope**

- In scope: Baseline operation (HTTP API, embedded web UI, SQLite or documented storage, shape animation routines that do not execute user Python) must not require installing Node.js, a separate Go toolchain, or other product-specific runtimes on the deployment host beyond the single binary (REQ-004, REQ-043).
- In scope: **Python** (interpreter meeting a minimum version documented in `docs/architecture.md`) is the only explicit additional runtime dependency for features that execute **user Python** routine code (REQ-021, REQ-022). Operators who do not use Python routines may omit Python only if architecture documents lazy behavior (attempt to run Python when missing must fail clearly).
- Out of scope: Bundling a full Python distribution inside the Go binary unless architecture chooses to do so; Windows Store vs python.org install paths.

**Business rules**

1. README and `docs/architecture.md` MUST state clearly that Python routine execution requires Python on the PATH (or documented absolute path policy) and that no other product runtime besides the binary is needed for non-Python features.
2. Operators following README download instructions (REQ-046) must be able to start the server and open the UI without running `npm install` or `go build` on the Pi or deployment host.

**Responsive / UX notes** *(when UI is involved)*

- N/A

**Dependencies**

- REQ-004, REQ-021, REQ-022, REQ-043

**Open questions**

- Minimum supported Python 3.x minor (for example 3.11 vs 3.12).

---

## REQ-046 — README: binary download, Raspberry Pi focus, systemd, updates

| Field | Value |
|-------|-------|
| **ID** | REQ-046 |
| **Title** | README: binary download, Raspberry Pi focus, systemd, updates |
| **Priority** | Must |
| **Actor(s)** | Operator / maintainer |

**User story**

As an operator deploying on a Raspberry Pi, I want **README.md** to tell me how to **download and run** the released binary, how to **install it as a service** that starts on boot, and how to **update** the application, without relying only on contributor-only documentation.

**Scope**

- In scope: README.md (hobbyist-facing) must include sections or clear subsections that cover (a) where to download official release binaries (GitHub Releases or documented URL) and which artifact to pick for Raspberry Pi (linux/arm64) vs other platforms (REQ-043); (b) minimal commands to copy the binary, make it executable (Unix), or run it on Windows/Linux on a fresh machine (REQ-045); (c) Raspberry Pi–focused instructions for running as a **systemd** service (`.service` unit example or equivalent) that starts after reboot; (d) how to update (replace binary, restart service, preserve `data/` or documented DB path per architecture) when a new release is published.
- Out of scope: Duplicating every contributor command from `docs/advanced-setup.md`; REQ-046 demands operator-facing coverage in README.md while advanced detail may stay in `docs/advanced-setup.md`.

**Business rules**

1. README.md MUST lead with or prominently feature **download release binary and run** for operators, including Pi (linux/arm64) and at least a pointer to Windows and desktop Linux artifacts.
2. README.md MUST include a worked **systemd** example (unit file snippet or copy-paste block) for Raspberry Pi OS or Debian-based Pi OS that runs the single binary with documented working directory and environment as needed.
3. README.md MUST explain updating in plain steps: fetch new release, stop service, replace binary, start service, and notes on database or file compatibility per architecture.
4. README.md MUST NOT introduce internal spec identifiers (`REQ-*`) per repository contributor policy—describe behavior in plain language (those remain in `docs/`).

**Responsive / UX notes** *(when UI is involved)*

- N/A

**Dependencies**

- REQ-003, REQ-043, REQ-044, REQ-045

**Open questions**

- Whether README also shows `./scripts/run.sh` for developers as a secondary path; recommended: keep both developer and binary-operator flows discoverable.

---

