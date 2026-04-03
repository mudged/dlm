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

- In scope: A model contains **up to 1000** lights; each light has a **sequential index** starting at **0** and a position in **3D space** (**x**, **y**, **z** coordinates). A model is representable as a **CSV** file with columns **id**, **x**, **y**, **z** (where **id** is the light index). Each model has **metadata** including a **name** and **creation date**.
- Out of scope: Physical wire topology beyond “lights on a wire” as a conceptual grouping; animation, color, or brightness per light; import formats other than the defined CSV unless added later.

**Business rules**

1. A model MUST NOT describe more than **1000** lights.
2. Light indices MUST be **non-negative integers** forming a **contiguous sequence** starting at **0** (i.e. for *n* lights, indices **0** through **n − 1** with **no gaps** and **no duplicates**).
3. Coordinates **x**, **y**, and **z** MUST be **real numbers** (finite; architecture may fix representation and precision).
4. The interchange CSV MUST use the field names **id**, **x**, **y**, **z** (column order and delimiter as specified in acceptance criteria).
5. Model metadata MUST include **name** and **creation date** (storage format and timezone policy deferred to architecture).

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

As a user, I want **three sample light models** available **by default** that illustrate **geometric shapes**, so that I can explore the product before uploading my own CSV.

**Scope**

- In scope: On a **fresh** application data set (e.g. empty database or first startup), the system MUST expose **exactly three** predefined models, distinct in name and geometry, representing: a **sphere**, a **cube**, and a **cone**. Coordinates use **SI meters**; **consecutive** lights (by **id** order **0, 1, …**) MUST be separated by a Euclidean distance of **exactly 0.1 m** (**10 cm**). Each shape MUST have an overall vertical extent (**height**) of **about 2 m** (interpreted as: sphere **diameter** ~2 m; cube **edge length** ~2 m; **right circular cone** with **height** ~2 m and base diameter chosen consistently with the geometry—exact proportions deferred to architecture within these bounds).
- Out of scope: User editing of sample models; optional **toggle to hide** samples; lighting beyond positions.

**Business rules**

1. The three samples MUST appear in the **same** **list** and **detail** flows as user-created models (**REQ-006**).
2. Each sample MUST stay within **REQ-005** (**≤ 1000** lights per model). If a literal **0.1 m** spacing along the chosen path would exceed **1000** lights, architecture MUST adjust the **path length** or **sampling** while keeping rules testable (document any cap in architecture).
3. Sample model **names** MUST make the shape identifiable (e.g. containing “sphere”, “cube”, “cone” or equivalent localized labels—English product default acceptable).
4. Samples MUST be **regenerated** or **re-seeded** when appropriate on empty store per persistence rules in architecture (e.g. migration seed, idempotent startup hook).

**Responsive / UX notes** *(when UI is involved)*

- Mobile: Samples appear in the list like other models; readable names.
- Tablet: Same as mobile.
- Desktop: Same.

**Dependencies**

- REQ-005, REQ-006

**Open questions**

- Whether samples are **removed** if the user deletes them manually.
- Whether **multiple** coordinate rings on the sphere are allowed vs a **single** space curve.

---
