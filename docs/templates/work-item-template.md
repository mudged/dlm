# Work item template

Copy this structure for each new `docs/work-items/WI-NN-<slug>.md` file. Replace placeholders; keep sections even when a field is "none". Work items are **self-contained bootstrap prompts** for a fresh **`@implementor`** chat.

---

## File header (title line)

```markdown
# WI-NN — <Short title>
```

---

## Full work item body

```markdown
# WI-NN — <Short title>

## Bootstrap

> **Human:** Start a **new** Cursor chat for this item only.
> 1. Set the chat **Model** to: **<Small/fast | Medium | Hard>** — see **Advised model** below.
> 2. Attach this file (or paste its contents) as context.
> 3. Invoke **`@implementor`** and ask it to implement this work item.

- **Persona:** `@implementor`
- **Advised model:** <Small/fast (Composer 2.5) | Medium (Sonnet) | Hard (Opus)>
- **Depends on:** <none | WI-XX, WI-YY>
- **Area:** <Backend Go | Frontend Next.js | CV / Python | Docs | Infra / CI>
- **Requirements:** <REQ-NNN, …> (see `docs/requirements.md`)
- **Architecture:** <`docs/architecture.md` §… citations>

## Context

<Why this item exists; repo layout reminder if helpful; key existing files and symbols the implementor should open first.>

## Tasks

1. <Concrete, ordered step with file paths and expected behavior.>
2. <…>

## Acceptance / tests

- <Specific test commands and scenarios that must pass.>
- `cd backend && go test ./...` and/or `cd web && npm test` / `npm run lint` as applicable.

## Out of scope

- <What this item must not touch — other WIs, settled docs, unrelated refactors.>

## Definition of done

<One short paragraph: observable outcome + tests green.>
```

---

## Conventions (every item)

- Repo root: monorepo with Go in `backend/`, Next.js in `web/`. See [`AGENTS.md`](../../AGENTS.md).
- Do **not** edit `docs/requirements.md` or `docs/architecture.md` during implementation; escalate conflicts to **`@architect`**.
- Add or adjust tests with each change; do not break existing tests.
