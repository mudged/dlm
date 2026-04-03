# Agent workflow (multi-agent development)

This repo uses a **spec-driven, multi-agent** process. Agent definitions live under [`.github/agents/`](.github/agents/). Each definition is a Markdown file with YAML frontmatter (`name`, `description`) and instructions the AI should follow when acting in that role.

## Product context

- **Backend:** Go (HTTP API/service).
- **Frontend:** Next.js with Tailwind CSS; UI must be **responsive** (mobile, tablet, desktop).
- **Deployment target:** Raspberry Pi 4 Model B (plan for ARM64, constrained CPU/RAM, and a clear run model for the Go process and the Next.js build).

## Agent pipeline (order)

| Order | Agent file | Cursor invoke | Responsibility |
|------:|------------|---------------|----------------|
| 1 | [`requirements.agent.md`](.github/agents/requirements.agent.md) | `@requirements` | Turn intent into `docs/requirements.md` and `docs/acceptance_criteria.md` (templates under `docs/templates/`). |
| 2 | [`architect.agent.md`](.github/agents/architect.agent.md) | `@architect` | Produce `docs/architecture.md` with Go + Next.js structure, Pi deployment notes, and **Mermaid** diagrams for boundaries and flows. |
| 3 | [`implementor.agent.md`](.github/agents/implementor.agent.md) | `@implementor` | Implement code and tests from architecture + Gherkin; TDD aligned with acceptance criteria. |
| 4 | [`verifier.agent.md`](.github/agents/verifier.agent.md) | `@verifier` | Audit vs architecture, run tests, report gaps or update `docs/traceability_matrix.md` on success. |

**Handoffs:** Each agent’s instructions end with who to invoke next; follow that chain unless you intentionally revisit an earlier stage.

## Local build and run (REQ-008)

**REQ-008** (see `docs/requirements.md`) requires a **single documented command** (script or Makefile target) that builds the Next static export for embed and starts the Go server. **`README.md` MUST stay aligned** with that command. When changing how the app is built or launched locally, update **requirements/architecture** if behavior changes, then **README**, then implementation.

## Supporting documents (expected paths)

Create and maintain these as the process runs (templates define shape):

- `docs/requirements.md`
- `docs/acceptance_criteria.md`
- `docs/architecture.md`
- `docs/traceability_matrix.md` (after successful verification)
- `docs/templates/requirement-template.md`
- `docs/templates/acceptance-criteria-template.md`
- `docs/templates/traceability-matrix-template.md`

If a template file is missing, add it before strict compliance with “MUST use template” steps is possible.

## Boundaries (by role)

- **Requirements:** No implementation; no Go/TS source or deployment manifests.
- **Architect:** No application source; specifications and diagrams only.
- **Implementor:** Code + tests; escalate spec conflicts to **@architect** rather than inventing divergent architecture.
- **Verifier:** No application code changes; audit, test execution, reports, and traceability updates only.

## Validation notes (agent files)

The agent files were aligned with this stack and deployment target: **Maven / Vert.x / Java** references in the architect role were replaced with **Go module layout, Next.js + Tailwind structure, and Raspberry Pi 4 deployment** planning. The requirements role no longer refers only to “Java.”

When editing `.github/agents/*.agent.md`, keep frontmatter valid, preserve the handoff at the end of each workflow, and update this file if the pipeline or document paths change.
