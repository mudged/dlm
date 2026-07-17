# Requirements

What Domestic Light & Magic is meant to do, in plain language anyone can follow.

- [`requirements.md`](requirements.md) — a friendly tour of every feature, grouped by theme, with a few
  diagrams. Start here.
- [`acceptance-criteria.md`](acceptance-criteria.md) — a "try this → you should see" checklist for
  confirming each feature actually works.

Both are written for a general reader, not just developers. You'll notice short reference codes like
`REQ-011` scattered through the project's source code and design docs; the bottom of
[`requirements.md`](requirements.md) has a lookup table that maps each code to the feature it describes.

**Adding or changing requirements?** Keep this plain-English, non-technical style (no templated
fields, no Gherkin), append a new `REQ-NNN` (never reuse or renumber) and add its row to the feature
code index. The full convention is in
[`../engineering/coding-standards.md`](../engineering/coding-standards.md) → "Writing requirements".
