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
  code to pass (RED-GREEN-REFACTOR). Acceptance tests trace to the Gherkin scenarios in
  [`../requirements/acceptance-criteria.md`](../requirements/acceptance-criteria.md).
- **Commands:**
  - Backend: `cd backend && go test ./...`
  - Frontend: `cd web && npm test` and `cd web && npm run lint`
- **Add or adjust tests with every change.** Do not break existing tests. The acceptance suite in
  `backend/internal/acceptance/` reads the specs and source by path; if you move a documented file,
  update those references too.

## Specifications are authoritative

- Requirements and acceptance criteria live in [`../requirements/`](../requirements); the design
  lives in [`../design/architecture.md`](../design/architecture.md). Keep code aligned with them.
- If you hit a genuine conflict with the design, **stop and fix the specification first** (update
  the design/requirements), then implement — do not silently work around the architecture.
