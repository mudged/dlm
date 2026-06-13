# WI-15 — Factory reset confirmation guard

- **Model:** Medium (Sonnet)
- **Depends on:** none (complements WI-11)
- **Area:** Backend Go (`backend/internal/httpapi/system.go`), Frontend (`web/app/options`)
- **Source:** Code review finding (Medium, security).

## Context

`POST /api/v1/system/factory-reset` (`backend/internal/httpapi/system.go` ~8–23) is unauthenticated
and takes no confirmation token. The app targets a trusted hobbyist LAN, but if the Pi is ever exposed
(port forward, guest Wi-Fi), any client can wipe all models, scenes, devices, and routines. The
product has no auth system, so the proportionate mitigation is a **confirmation token**, not full
authn.

Check the current UI first: `web/app/options` already has a factory-reset disclosure (a recent commit
mentions "align factory-reset disclosure with REQ-017 BR-2"). Align the guard with REQ-017 — read
`docs/requirements.md` REQ-017 before changing behavior; if REQ-017 specifies the exact confirmation
contract, follow it and do not contradict it (flag conflicts instead of inventing).

Key existing files:
- `backend/internal/httpapi/system.go` — the factory-reset handler.
- `web/app/options/*` — the options page with the existing factory-reset control/disclosure.
- `docs/requirements.md` — REQ-017 (factory reset) and its business rules.

## Tasks

1. Require an explicit confirmation in the request body, e.g. `{ "confirm": "FACTORY RESET" }` (or the
   exact phrase REQ-017 specifies). Reject with `400`/`422` standard envelope (code e.g.
   `confirmation_required`) when missing or mismatched. Keep it a `POST`.
2. Update `web/app/options` so the existing disclosure flow sends the confirmation value (typed phrase
   or explicit checkbox → value). Keep the UI accessible and responsive.
3. Add the body to `http.MaxBytesReader` with a small limit (it's a tiny JSON).

## Acceptance / tests

- Backend test: reset without/with-wrong confirmation → rejected; correct confirmation → proceeds
  (use a fake/empty store; assert it reaches the reset path).
- Frontend test (if a test exists for options): the action submits the confirmation value.
- `cd backend && go test ./...` and `cd web && npm test && npm run lint` pass.

## Out of scope

- Stopping background workers during reset (WI-11) — coordinate ordering but implement separately.
- A general authentication system.

## Definition of done

Factory reset requires an explicit, documented confirmation token before destroying data, aligned with
REQ-017; UI sends it; covered by tests.
