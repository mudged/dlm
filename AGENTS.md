# Agent workflow

This repo uses the **[Superpowers](https://github.com/obra/superpowers)** methodology for AI-assisted
development. Superpowers is a set of composable skills that make the agent work like a disciplined
engineer: it brainstorms a spec, writes a plan, implements with test-driven development, reviews, and
verifies before claiming done. Project-specific knowledge lives under [`docs/`](docs/) and is
referenced below.

> **Note:** This repository previously used a five-persona pipeline (`@requirements`, `@architect`,
> `@planner`, `@implementor`, `@verifier`) with work items and a traceability matrix. That approach
> has been retired in favor of Superpowers; the knowledge those personas produced now lives in
> [`docs/`](docs/).

## Product context

- **Backend:** Go (HTTP API/service), single self-contained binary, pure Go, no cgo.
- **Frontend:** Next.js with Tailwind CSS; UI must be **responsive** (mobile, tablet, desktop). It is
  built to a static export and embedded into the Go binary.
- **Deployment target:** Raspberry Pi 4 Model B (ARM64, constrained CPU/RAM).

## Project knowledge map (`docs/`)

Read the relevant area before making changes. Index: [`docs/README.md`](docs/README.md).

| Area | Use it for |
|------|-----------|
| [`docs/requirements/`](docs/requirements/) | What to build: a plain-English feature tour and a "try this → you should see" checklist (feature codes `REQ-*` are cross-referenced in an index). |
| [`docs/design/`](docs/design/) | How it is designed, split into junior-readable topic files. Start at [`architecture.md`](docs/design/architecture.md) (big picture + `§`→file index); details in service/frontend/deployment/request-flow pages, with a [`glossary`](docs/design/glossary.md) and Mermaid diagrams. |
| [`docs/engineering/`](docs/engineering/) | How to build/generate code: [coding standards](docs/engineering/coding-standards.md), [build and run](docs/engineering/build-and-run.md), [CI and release](docs/engineering/ci-and-release.md), [CV runtime](docs/engineering/cv-runtime.md), [environment and API](docs/engineering/environment-and-api.md). |
| [`docs/userguide/`](docs/userguide/) | How the end user operates the software. |

Requirements and acceptance criteria are **authoritative**. If implementation reveals a genuine
conflict with the design, update the specification first (see
[`docs/engineering/coding-standards.md`](docs/engineering/coding-standards.md)).

Requirements are written for a **non-technical reader** (aim for a ~15-year-old): `requirements.md` is
a plain-English, themed feature tour and `acceptance-criteria.md` is a "try this → you should see"
checklist — no templated fields or Gherkin. Feature codes (`REQ-*`) stay out of the prose and instead
live in the **feature code index** at the bottom of `requirements.md`; when adding a requirement,
append a new `REQ-NNN` (never reuse or renumber) and add its index row. Full convention:
[`docs/engineering/coding-standards.md`](docs/engineering/coding-standards.md) → "Writing requirements".

Design docs in [`docs/design/`](docs/design/) are written for a **junior developer** (knows basic
Go/JS/HTTP, but not this project's domain): explain jargon on first use, keep bold sparing, lead each
section with a plain-language summary, and use Mermaid diagrams where they help. Preserve the
stability contract — **never renumber or reuse a `§N.N` section number or a `REQ-*` code** (both are
cited from source-code comments and other docs). Full convention:
[`docs/engineering/coding-standards.md`](docs/engineering/coding-standards.md) → "Writing design docs".

## Local build and run

The supported one-command build-and-run is **`./scripts/run.sh`** from the repo root. It builds the
Next.js static export into the Go embed tree and starts the Go server on
[http://127.0.0.1:8080/](http://127.0.0.1:8080/). Full detail (prerequisites, env vars, dev
two-process workflow, cross-compilation) is in
[`docs/engineering/build-and-run.md`](docs/engineering/build-and-run.md).

**`README.md`** is the hobbyist-facing landing page and must stay approachable: document
`./scripts/run.sh` and point at the [user guide](docs/userguide/); do **not** put internal
requirement IDs (`REQ-*`) in it — those belong in `docs/`.

## Key commands

| Task | Command |
|------|---------|
| Build + run (embed UI, start server) | `./scripts/run.sh` from repo root |
| Install frontend deps | `cd web && npm ci` |
| Download Go deps | `cd backend && go mod download` |
| Frontend lint | `cd web && npm run lint` |
| Frontend tests | `cd web && npm test` |
| Go tests | `cd backend && go test ./...` |
| Dev: Go server only | `cd backend && go run ./cmd/server` (serves on `:8080`) |
| Dev: Next.js hot-reload | `cd web && npm run dev` (`:3000`, proxies API to `:8080`) |

## Environment and gotchas

- **Go ≥ 1.25** is required (`backend/go.mod` declares `go 1.25.0`); the default system Go (1.22) will
  refuse to build. Ensure a 1.25.x toolchain is on `PATH` (`/usr/local/go/bin` on Cursor Cloud).
- **No external services** — SQLite is embedded (pure Go). The DB auto-creates at `data/dlm.db`.
- **Sample data** — a fresh/empty DB seeds 3 models (sphere, cube, cone) and 3 Python routines.
- **CSV upload** — `POST /api/v1/models` expects field name `file`; header exactly `id,x,y,z`; light
  IDs 0-based sequential.
- **`next build`** is the slow step; use `DLM_SKIP_NPM_CI=1 ./scripts/run.sh` to skip `npm ci` on
  repeat runs.

Full detail: [`docs/engineering/build-and-run.md`](docs/engineering/build-and-run.md) and
[`docs/engineering/environment-and-api.md`](docs/engineering/environment-and-api.md).
