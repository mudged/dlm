# Agent workflow (multi-agent development)

This repo uses a **spec-driven, multi-agent** process. Agent definitions live under [`.github/agents/`](.github/agents/). Each definition is a Markdown file with YAML frontmatter (`name`, `description`) and instructions the AI should follow when acting in that role.

## Documentation folders

| Path | Audience |
|------|-----------|
| **[`docs/user/`](docs/user/)** | People installing or running the app (start with **[README.md](README.md)**). |
| **`docs/agentic-development/`** | Specifications, acceptance criteria, architecture, traceability matrix, and templates used by agents and contributors. |

See **[`docs/README.md`](docs/README.md)** for an overview.

## Product context

- **Backend:** Go (HTTP API/service).
- **Frontend:** Next.js with Tailwind CSS; UI must be **responsive** (mobile, tablet, desktop).
- **Deployment target:** Raspberry Pi 4 Model B (plan for ARM64, constrained CPU/RAM, and a clear run model for the Go process and the Next.js build).

## Agent pipeline (order)

| Order | Agent file | Cursor invoke | Responsibility |
|------:|------------|---------------|----------------|
| 1 | [`requirements.agent.md`](.github/agents/requirements.agent.md) | `@requirements` | Turn intent into `docs/agentic-development/requirements.md` and `docs/agentic-development/acceptance_criteria.md` (templates under `docs/agentic-development/templates/`). |
| 2 | [`architect.agent.md`](.github/agents/architect.agent.md) | `@architect` | Produce `docs/agentic-development/architecture.md` with Go + Next.js structure, Pi deployment notes, and **Mermaid** diagrams for boundaries and flows. |
| 3 | [`implementor.agent.md`](.github/agents/implementor.agent.md) | `@implementor` | Implement code and tests from architecture + Gherkin; TDD aligned with acceptance criteria. |
| 4 | [`verifier.agent.md`](.github/agents/verifier.agent.md) | `@verifier` | Audit vs architecture, run tests, report gaps or update `docs/agentic-development/traceability_matrix.md` on success. |

**Handoffs:** Each agent’s instructions end with who to invoke next; follow that chain unless you intentionally revisit an earlier stage.

## Local build and run (REQ-008)

**REQ-008** (see `docs/agentic-development/requirements.md`) requires a **single documented command** (script or Makefile target) that builds the Next static export for embed and starts the Go server. **`README.md` MUST stay aligned** with that command: the hobbyist-facing README documents **`./scripts/run.sh`** from the repo root as the primary way to install and run (see **[README.md](README.md)**). **`README.md` MUST NOT** mention internal requirement IDs (`REQ-*`), traceability, or other identifiers that only exist under **`docs/agentic-development/`**—those belong in specifications and **[docs/user/advanced-setup.md](docs/user/advanced-setup.md)** for contributors. When changing how the app is built or launched locally, update **requirements/architecture** if behavior changes, then **README** (keep it approachable), **docs/user/advanced-setup.md** (technical detail), then implementation.

## Supporting documents (expected paths)

Create and maintain these as the process runs (templates define shape):

- `docs/agentic-development/requirements.md`
- `docs/agentic-development/acceptance_criteria.md`
- `docs/agentic-development/architecture.md`
- `docs/agentic-development/traceability_matrix.md` (after successful verification)
- `docs/agentic-development/templates/requirement-template.md`
- `docs/agentic-development/templates/acceptance-criteria-template.md`
- `docs/agentic-development/templates/traceability-matrix-template.md`

If a template file is missing, add it before strict compliance with “MUST use template” steps is possible.

## Boundaries (by role)

- **Requirements:** No implementation; no Go/TS source or deployment manifests.
- **Architect:** No application source; specifications and diagrams only.
- **Implementor:** Code + tests; escalate spec conflicts to **@architect** rather than inventing divergent architecture.
- **Verifier:** No application code changes; audit, test execution, reports, and traceability updates only.

## Validation notes (agent files)

The agent files were aligned with this stack and deployment target: **Maven / Vert.x / Java** references in the architect role were replaced with **Go module layout, Next.js + Tailwind structure, and Raspberry Pi 4 deployment** planning. The requirements role no longer refers only to “Java.”

When editing `.github/agents/*.agent.md`, keep frontmatter valid, preserve the handoff at the end of each workflow, and update this file if the pipeline or document paths change.

## Cursor Cloud specific instructions

### System dependencies

- **Go ≥ 1.25** is required (`go.mod` declares `go 1.25.0`). The VM update script installs Go 1.25.x automatically. Verify with `go version`.
- **Node.js LTS** (22.x) and **npm** are pre-installed and sufficient.

### Key commands (see `README.md` for full table)

| Task | Command |
|------|---------|
| Install frontend deps | `cd web && npm ci` |
| Download Go deps | `cd backend && go mod download` |
| Frontend lint | `cd web && npm run lint` |
| Frontend tests | `cd web && npm test` |
| Go tests | `cd backend && go test ./...` |
| Build + run (REQ-008) | `./scripts/run.sh` from repo root |
| Dev: Go server only | `cd backend && go run ./cmd/server` (serves on `:8080`) |
| Dev: Next.js hot-reload | `cd web && npm run dev` (`:3000`, proxies API to `:8080`) |

### Running the app

1. Build the frontend static export and copy it into the Go embed tree: `cd web && npm run release:sync`
2. Start the Go server: `cd backend && go run ./cmd/server`
3. The app is at `http://127.0.0.1:8080/`

Alternatively, `./scripts/run.sh` does both steps (plus `npm ci` if `node_modules` is missing).

### Gotchas

- **`go.mod` says `go 1.25.0`** — the default system Go (1.22) will refuse to build. The update script installs Go 1.25.x to `/usr/local/go`. Ensure `/usr/local/go/bin` is on `PATH`.
- **No external services** — SQLite is embedded (pure Go, no CGO). Database file auto-creates at `data/dlm.db`. No Docker, Redis, or Postgres needed.
- **Sample data** — on first start with an empty DB, 3 sample models (sphere, cube, cone) are auto-seeded. Deleting all models and restarting re-seeds them.
- **CSV upload field name** — the `POST /api/v1/models` endpoint expects the CSV file under field name `file` (not `csv`), and the CSV header must be exactly `id,x,y,z`. Light IDs must be **0-based sequential** integers (0, 1, 2, …); 1-based IDs will fail validation.
- **`next build`** is the slow step (~30–120 s). The `.next/` cache speeds up subsequent builds. Use `DLM_SKIP_NPM_CI=1 ./scripts/run.sh` to skip `npm ci` on repeat runs.
- **Next.js dev + SSE (live light updates)** — `npm run dev` rewrites buffer `text/event-stream`, so the UI opens EventSource against `http://127.0.0.1:8080` directly (`web/lib/sseUrl.ts`). Start Go with `CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000` (or set `NEXT_PUBLIC_DLM_API_ORIGIN` to match your API). `./scripts/run.sh` serves UI and API on the same origin and does not need this.
