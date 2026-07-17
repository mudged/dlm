# Requirements

Product requirements and acceptance criteria.

- [`requirements.md`](requirements.md) — user stories and business rules, one immutable `REQ-NNN` per
  requirement.
- [`acceptance-criteria.md`](acceptance-criteria.md) — Gherkin scenarios; every scenario declares its
  parent requirement (`Parent requirement: REQ-xxx`).
- [`templates/`](templates/) — templates for adding new requirements and acceptance criteria.

Requirement IDs are **immutable** — never reuse or renumber them. The Go acceptance suite in
`backend/internal/acceptance/` reads these files by path, so keep filenames and cross-references
consistent when editing.
