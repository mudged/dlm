# Contributing

Thanks for helping improve Domestic Light & Magic. This page covers one-time setup for contributors.
For build/run detail, coding standards, CI, and the rest of the engineering docs, see
[`docs/engineering/`](docs/engineering/).

## Development methodology: Superpowers

This project follows the **[Superpowers](https://github.com/obra/superpowers)** methodology for
AI-assisted development — a set of composable skills that make a coding agent work like a disciplined
engineer (brainstorm a spec, write a plan, implement with test-driven development, review, and verify
before claiming done). Project-specific knowledge that the agent relies on lives under [`docs/`](docs/)
and is mapped from [`AGENTS.md`](AGENTS.md).

### Install Superpowers (one-time, per developer)

Superpowers is a Cursor plugin, installed client-side — it is **not** committed to this repository, so
each contributor installs it once in their own editor. In Cursor Agent chat:

```text
/add-plugin superpowers
```

Or search for "superpowers" in the Cursor plugin marketplace. Once installed, the skills trigger
automatically. See the [Superpowers README](https://github.com/obra/superpowers) for other hosts
(Claude Code, Codex, etc.) and details.

## Build, run, and test

The supported one-command build-and-run is **`./scripts/run.sh`** from the repo root. Prerequisites,
environment variables, the two-process dev workflow, and cross-compilation are documented in
[`docs/engineering/build-and-run.md`](docs/engineering/build-and-run.md). Before opening a pull
request, run the test and lint gates described in
[`docs/engineering/coding-standards.md`](docs/engineering/coding-standards.md) and
[`docs/engineering/ci-and-release.md`](docs/engineering/ci-and-release.md).
