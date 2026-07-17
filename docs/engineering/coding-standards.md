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
