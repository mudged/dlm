# Coding standards

How code should be written for this repository. These conventions apply to every change,
whether made by a human or an agent. They complement the Superpowers workflow described in the
repository [`AGENTS.md`](../../AGENTS.md) (brainstorm, plan, TDD, review, verify).

## Repository shape

- **Monorepo.** The Go backend lives in [`backend/`](../../backend) (single Go module); the
  Next.js + Tailwind frontend lives in [`web/`](../../web). Repo root holds shared scripts and
  documentation.
- **Single binary, pure Go, no cgo.** The product build is one self-contained executable.
  SQLite is `modernc.org/sqlite` (pure Go). **Do not introduce cgo** into the main binary or add a
  second long-running daemon for normal operation.
- The frontend is compiled to a **static export** that is embedded into the Go binary
  (`backend/internal/webdist/`). See [`build-and-run.md`](build-and-run.md).

## Language and framework expectations

- **Backend:** Go (HTTP API/service). Follow the package boundaries and API surface defined in
  [`../design/architecture.md`](../design/architecture.md); do not invent a divergent architecture.
- **Frontend:** Next.js with Tailwind CSS. Build **responsive, mobile-first** UI that works on
  phone, tablet, and desktop. Do not rely on hover-only interactions for essential steps.
- **CV/Python:** Reconstruction runs in a bundled OpenCV runtime; user routines run in the host
  `python3`. See [`cv-runtime.md`](cv-runtime.md) and [`environment-and-api.md`](environment-and-api.md).

## HTTP conventions

- **Error envelope:** API errors use `{ "error": { "code", "message", "details"? } }`. Use the
  helpers in `backend/internal/httpapi/json.go` rather than hand-rolling error responses.
- Prefer **batch/bulk** routes over per-light requests for high-throughput light updates (see
  [`environment-and-api.md`](environment-and-api.md)).

## Testing

- **Test-driven development.** Write a failing test first, watch it fail, then write the minimal
  code to pass (RED-GREEN-REFACTOR). Tests should verify the behaviour described in
  [`../requirements/`](../requirements) (the feature tour and its "try this → you should see"
  checklist).
- **Commands:**
  - Backend: `cd backend && go test ./...`
  - Frontend: `cd web && npm test` and `cd web && npm run lint`
- **Add or adjust tests with every change.** Do not break existing tests.

## Specifications are authoritative

- Requirements and acceptance criteria live in [`../requirements/`](../requirements); the design
  lives in [`../design/architecture.md`](../design/architecture.md). Keep code aligned with them.
- If you hit a genuine conflict with the design, **stop and fix the specification first** (update
  the design/requirements), then implement — do not silently work around the architecture.

## Writing requirements

The docs in [`../requirements/`](../requirements) are deliberately **not** a formal, templated spec.
Keep them that way when you add or change requirements:

- **Audience: a curious non-technical reader** (aim for a ~15-year-old with little background). Use
  plain language, short sentences, and everyday analogies. Explain any unavoidable jargon in one
  friendly line the first time it appears, or avoid it. No field tables, no
  "User story / Scope / Business rules" template, no Gherkin.
- **`requirements.md` is a themed feature tour.** Group related behaviour into readable sections
  rather than one block per requirement. Add a Mermaid diagram only where it genuinely helps a
  beginner. Keep concrete, testable facts accurate (limits, hex colours, measurements, timings,
  technology names).
- **`acceptance-criteria.md` is a "try this → you should see" checklist**, grouped by the same
  themes.
- **Feature codes stay out of the prose.** Every requirement still has a stable `REQ-NNN` code
  (referenced from source-code comments and [`../design/architecture.md`](../design/architecture.md)),
  but the code does **not** appear inline in the readable text. Instead:
  - keep the **feature code index** table at the bottom of `requirements.md` up to date, and
  - when you add a new requirement, **append a new `REQ-NNN`** (never reuse or renumber existing
    codes) with a new row mapping it to the section that describes it.
- This keeps the many existing `REQ-NNN` references resolvable while the docs stay approachable.

## Writing design docs

The design docs in [`../design/`](../design) explain **how** the product is built. Keep them readable
for a junior developer as you edit or extend them:

- **Audience: a junior developer** who knows basic Go, JavaScript/TypeScript, and HTTP, but **not**
  this project's domain (home LEDs, WLED, three.js, computer-vision reconstruction, Raspberry Pi
  constraints). Explain domain jargon and acronyms in one short clause the first time they appear, or
  link to [`../design/glossary.md`](../design/glossary.md). Add new terms to the glossary.
- **Lead with plain language.** Start each major section with a one-to-three-sentence
  "**In plain terms:**" summary before the detailed bullets, tables, or schemas.
- **Bold sparingly.** Do not bold whole sentences or every noun; reserve bold for genuine emphasis.
- **Use Mermaid diagrams** where a picture helps a beginner (architecture, data flow, sequence of a
  request). Keep existing diagrams; improve rather than delete.
- **Keep concrete facts accurate and authoritative:** HTTP methods/paths, status and error codes,
  JSON field names, SQL schema, numeric limits, hex colours, timings, env var names, and
  `GOOS`/`GOARCH` values. Do not soften these into vague prose.
- **Stability contract (do not break):**
  - Never renumber, reuse, or delete a `§N.N` **section number** — they are cited from source-code
    comments and other docs. Add new sections with new numbers; keep moved content under the same
    number in whichever file now hosts it, and update the `§`→file index in
    [`../design/architecture.md`](../design/architecture.md).
  - Never renumber or reuse a `REQ-*` **feature code**; keep the traceability references intact.
- **Structure:** the design is split into topic files with [`architecture.md`](../design/architecture.md)
  as the entry-point hub (big-picture overview plus the `§`→file section index). Other files (and
  source-code comments) may link to `architecture.md` by name and cite `§` numbers, so keep that file
  as the stable landing page.

## Writing the user guide

The docs in [`../userguide/`](../userguide) tell an **end user** how to install and operate the app.
Keep them beginner-friendly whenever you add or change a page:

- **Audience: a curious 15-year-old with little technical background.** Assume no programming
  knowledge and no familiarity with this project. Use plain language, short sentences, and everyday
  analogies (fairy lights, taking photos of a statue, a game and its data folder). Be encouraging,
  not intimidating.
- **Explain jargon on first use, or avoid it.** Terms like *terminal*, *Raspberry Pi*, *systemd*,
  *CSV*, *ESP32/WLED*, *SSE* get one friendly sentence the first time they appear (a `>` blockquote
  "jargon check" works well). Prefer a plain word over a technical one when both fit.
- **Lead with the goal and the payoff**, then the steps. Number step-by-step instructions and keep
  each step to one clear action.
- **Use Mermaid diagrams** for anything with a flow or a "big picture" (download→run→open, the
  capture workflow, how the sections relate). Keep them simple — a handful of boxes, short labels.
- **Keep concrete facts accurate.** File names (`dlm_linux_arm64.tar.gz`), the address
  `http://127.0.0.1:8080/`, the default port `8080`, the CSV header `id,x,y,z` with 0-based
  sequential IDs, supported video formats (MP4/MOV/MKV/WebM), env var names (`DLM_DATA_DIR`,
  `DLM_DB_PATH`, `DLM_PYTHON3`, `HTTP_LISTEN`), and the `runtime/cv/` layout must stay correct even
  when the prose is casual.
- **Push technical depth elsewhere.** Developer/build topics belong in
  [`../engineering/`](../engineering); API and streaming details live in
  [`environment-and-api.md`](environment-and-api.md). Link out to them instead of explaining them
  here.
- **Don't put internal `REQ-*` codes in the user guide.** Those belong in
  [`../requirements/`](../requirements) and [`../design/`](../design).
