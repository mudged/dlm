# Design

Architecture and design of the application, written to be readable by a junior developer. Start with
[`architecture.md`](architecture.md) — it gives the big picture, a system diagram, and an index that
maps every section number (`§N.N`) to the file it lives in.

| File | What it covers |
|------|----------------|
| [`architecture.md`](architecture.md) | Entry point: what dlm is, the big-picture diagram, the key single-binary decision, the reading order, and the `§`→file section index. |
| [`overview.md`](overview.md) | §1 goals and constraints (per-requirement response table) and §2 repository layout. |
| [`backend-service.md`](backend-service.md) | §3.1–§3.14: the Go service — HTTP surface, persistence, build/embed, sample data, per-light state, scenes, factory reset. |
| [`backend-lights-and-automation.md`](backend-lights-and-automation.md) | §3.15–§3.23: scene geometry, routines (Python + shape animation), light-state push/elision, WLED devices, capture sweep, camera reconstruction. |
| [`frontend.md`](frontend.md) | §4: the Next.js + Tailwind + three.js UI. |
| [`deployment.md`](deployment.md) | §5–§6: UI/API coordination and Raspberry Pi deployment, release targets, CI/CD. |
| [`request-flows.md`](request-flows.md) | §7–§8: system-boundary flowchart and per-feature sequence diagrams (Mermaid). |
| [`security.md`](security.md) | §9: baseline security notes. |
| [`glossary.md`](glossary.md) | Plain-language definitions of terms and acronyms. |
| [`appendix-traceability.md`](appendix-traceability.md) | Dense `REQ-*` → section cross-reference blocks. |

**Stability contract:** section numbers (`§3.18`, `§8.24`) and feature codes (`REQ-*`) are cited from
source-code comments and other docs. Keep them stable when editing — never renumber or reuse them.
Requirements this design realizes live in [`../requirements/`](../requirements).
