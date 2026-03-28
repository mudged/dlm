# Acceptance criteria

Gherkin scenarios trace to requirements via `Parent requirement: REQ-xxx` on every scenario.

```gherkin
Feature: Full-stack composition (Go, Next.js, Tailwind)

  Scenario: Requirements mandate Go backend and Next.js with Tailwind
    Parent requirement: REQ-001
    Given docs/requirements.md exists
    When the requirement REQ-001 is read
    Then the product must include a Golang service component
    And the product must include a Next.js front end with Tailwind CSS

  Scenario: REQ-001 identifies actors and priority
    Parent requirement: REQ-001
    Given docs/requirements.md exists
    When the REQ-001 metadata table is read
    Then the priority is Must
    And actors include end user and operator or maintainer

Feature: Responsive and reactive experience

  Scenario: Requirements define mobile, tablet, and desktop expectations
    Parent requirement: REQ-002
    Given docs/requirements.md exists
    When the REQ-002 responsive UX notes are read
    Then expectations are stated for mobile viewports
    And expectations are stated for tablet viewports
    And expectations are stated for desktop viewports

  Scenario: Requirements require reactive UI behavior
    Parent requirement: REQ-002
    Given docs/requirements.md exists
    When the REQ-002 business rules are read
    Then client-side interactivity appropriate for Next.js is required
    And primary user-facing flows must remain usable across device classes

Feature: Raspberry Pi 4 Model B deployment

  Scenario: Requirements name Pi 4 Model B as deployment target
    Parent requirement: REQ-003
    Given docs/requirements.md exists
    When REQ-003 is read
    Then Raspberry Pi 4 Model B is stated as a deployment target

  Scenario: Requirements bind architecture to Pi-oriented documentation
    Parent requirement: REQ-003
    Given docs/requirements.md exists
    When the REQ-003 business rules are read
    Then they require docs/architecture.md to describe Go and Next.js on Pi 4 Model B after the architect pass
    And they require documentation to acknowledge ARM64 together with typical Pi CPU and RAM limits

Feature: Single executable packaging without Docker

  Scenario: REQ-004 mandates one executable per release target
    Parent requirement: REQ-004
    Given docs/requirements.md exists
    When the requirement REQ-004 is read
    Then the product must ship or document one runnable executable file per supported release target
    And a separate Node.js runtime binary must not be required from the distribution for routine operation

  Scenario: REQ-004 defers Docker and mandatory container packaging
    Parent requirement: REQ-004
    Given docs/requirements.md exists
    When the REQ-004 scope and business rules are read
    Then Docker OCI images and docker-compose as required packaging are out of scope at this stage
    And documentation must not present containers as the only or required way to run the product

  Scenario: REQ-004 requires architecture reconciliation with Pi and UI requirements
    Parent requirement: REQ-004
    Given docs/requirements.md exists
    When the REQ-004 business rules are read
    Then they require docs/architecture.md to be updated for a consistent single-binary model with REQ-003 and REQ-001
```

---

**Next step:** Invoke the `@verifier` agent to audit the codebase against this document and `docs/architecture.md`, run tests, and update `docs/traceability_matrix.md`.
