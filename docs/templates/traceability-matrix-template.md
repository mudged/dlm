# Traceability matrix

The verifier appends rows to the table below. **Do not remove or reorder columns.** Add new rows after the header separator line.

| ID | Functional Requirement | System Component | Test Case Number | Status | Verification |
|----|------------------------|------------------|------------------|--------|--------------|
| TRC-001 | REQ-000 — <short title> | <e.g. backend: `package/path`, frontend: `app/...`> | <e.g. TC-REQ-000-01 or test file:line> | Pending \| Pass \| Fail | <e.g. verified date / verifier note> |

## Column definitions

- **ID**: Stable trace row ID (e.g. TRC-001), unique in this file.
- **Functional Requirement**: REQ ID plus short title, matching `docs/requirements.md`.
- **System Component**: Primary code location or service name implementing the requirement.
- **Test Case Number**: Automated test identifier or stable name that maps to a scenario in `docs/acceptance_criteria.md`.
- **Status**: `Pending` until verification; then `Pass` or `Fail`.
- **Verification**: Brief evidence (e.g. test run id, date, manual check note).
