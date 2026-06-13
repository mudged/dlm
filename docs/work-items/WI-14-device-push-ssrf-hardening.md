# WI-14 — Harden device push against redirect-based SSRF

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** Backend Go (`backend/internal/devices/wled.go`)
- **Source:** Code review finding (Medium, security).

## Context

Device `base_url` is validated at creation for `http`/`https` scheme + host
(`backend/internal/store/devices.go` ~34–49), but the outbound HTTP client that pushes light state to
WLED (`backend/internal/devices/wled.go` ~60–88) uses Go's default redirect-following with no
`CheckRedirect` hook. A malicious or compromised endpoint can answer with a redirect to an arbitrary
URL, causing the Pi to issue its POST to an internal LAN address — an SSRF amplification beyond the
operator-intended WLED host.

Key existing files:
- `backend/internal/devices/wled.go` — the `http.Client` and the request/POST to the device (~60–88).
- `backend/internal/store/devices.go` — existing base_url validation (~34–49) for reference.

## Tasks

1. Give the device `http.Client` a `CheckRedirect` that **refuses redirects** (return
   `http.ErrUseLastResponse` or an error). WLED's JSON/state API does not need to follow redirects;
   a redirect from a light controller is a red flag.
2. Set a sensible request **timeout** on the client if not already present (avoid hanging on a
   malicious host), and keep it shared/reused rather than per-request if that matches current style.
3. (Optional, note in PR if you add it) Validate the resolved host is not obviously off-target, but the
   core fix is disabling redirects.

## Acceptance / tests

- Unit test with an `httptest.Server` that responds `302` to an attacker URL: the client does **not**
  follow the redirect (request fails or returns the 3xx without re-issuing to the new location).
- Normal `200` push still works.
- `cd backend && go test ./...` passes.

## Out of scope

- Authentication on the API itself; broader network policy.

## Definition of done

The device push client never follows HTTP redirects and has a request timeout; covered by a test.
