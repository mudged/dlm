# WI-01 — Device `light_count` field (store + API)

- **Model:** Small/fast (Composer 2.5)
- **Depends on:** none
- **Area:** Backend Go (`backend/`)
- **Requirements:** REQ-047 (see `docs/requirements.md`)
- **Architecture:** `docs/architecture.md` §3.3 (devices schema), §3.2 (devices routes)

## Context

Repo `/workspaces/dlm` is a monorepo: Go module under `backend/`, Next.js under `web/`.
The product is a single Go binary using a **pure-Go** SQLite driver (`modernc.org/sqlite`,
no cgo). Devices (WLED controllers) are already implemented. This item adds a persisted
`light_count` to a device so a later "capture light sequence" (WI-02) knows how many lights
to sweep **even when the device is not assigned to any model**.

Key existing files:
- `backend/internal/store/devices.go` — `Device` struct, `ensureDeviceTable`, `scanDevice`, `CreateDevice`, `PatchDevice`, list/get queries.
- `backend/internal/httpapi/devices.go` — HTTP handlers for `/api/v1/devices…`.
- Tests: `backend/internal/store/devices_test.go`, `backend/internal/httpapi/devices_api_test.go`.

## Tasks

1. **Schema + migration** in `backend/internal/store/devices.go`:
   - Add column `light_count INTEGER NOT NULL DEFAULT 0` to the `CREATE TABLE IF NOT EXISTS devices` statement in `ensureDeviceTable`.
   - Add an **idempotent migration** for existing DBs: detect whether the `light_count` column exists (e.g. query `PRAGMA table_info(devices)`), and if missing run `ALTER TABLE devices ADD COLUMN light_count INTEGER NOT NULL DEFAULT 0`. Legacy rows must inherit `0`.
2. **Struct + DTO fields:**
   - Add `LightCount int \`json:"light_count"\`` to `Device`.
   - Add `LightCount int` to `DeviceCreate`.
   - Add `LightCount *int` to `DevicePatch` (nil = leave unchanged).
3. **Read paths:** add `light_count` to **all three** SELECT column lists (`ListDevices`, `GetDevice`, `GetDeviceForModel`) and to `scanDevice` (scan into `d.LightCount`). Keep column order consistent across query and scan.
4. **Write paths:**
   - `CreateDevice`: validate `0 <= LightCount <= 1000` (return a clear error otherwise — reuse the existing error style; the HTTP layer maps it to 400). Insert the value.
   - `PatchDevice`: when `LightCount != nil`, validate the same range and add `light_count = ?` to the update set.
5. **HTTP** in `backend/internal/httpapi/devices.go`:
   - Parse `light_count` from the POST and PATCH JSON bodies and pass through to the store types.
   - Include `light_count` in the device JSON responses (it is already on `Device`, so this is automatic if you serialize `Device`; verify the response shape).
   - On out-of-range `light_count`, return `400` with the standard envelope, code `validation_failed` (match the existing validation error code used for name/type).

## Acceptance / tests

- Extend `backend/internal/store/devices_test.go`:
  - Create with `light_count = 50` round-trips through get/list.
  - Patch updates `light_count`; omitting it leaves it unchanged.
  - Out-of-range (`-1`, `1001`) is rejected.
  - **Migration:** opening a store whose `devices` table predates the column yields existing rows with `light_count == 0` (simulate by creating the table without the column, or rely on the `ALTER` path).
- Extend `backend/internal/httpapi/devices_api_test.go`: POST/PATCH accept `light_count`; GET returns it; out-of-range → `400` `validation_failed`.
- `cd backend && go test ./...` passes.

## Out of scope

- The capture sweep behaviour and endpoints (WI-02).
- Any UI (WI-03).

## Definition of done

`light_count` is persisted, migrated, validated (0–1000), and visible in device API
responses; all backend tests green.
