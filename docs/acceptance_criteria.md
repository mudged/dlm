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

Feature: Wire light model definition

  Scenario: REQ-005 caps lights at 1000 and requires sequential ids from zero
    Parent requirement: REQ-005
    Given docs/requirements.md exists
    When requirement REQ-005 is read
    Then a model must allow at most 1000 lights
    And light indices must be a contiguous sequence starting at 0 with no gaps or duplicates

  Scenario: REQ-005 defines CSV columns and metadata
    Parent requirement: REQ-005
    Given docs/requirements.md exists
    When requirement REQ-005 is read
    Then the interchange CSV must use fields id, x, y, and z
    And model metadata must include name and creation date

Feature: Model management in the application

  Scenario: REQ-006 requires list, view, delete, and CSV upload create
    Parent requirement: REQ-006
    Given docs/requirements.md exists
    When requirement REQ-006 is read
    Then the application must support listing all models
    And the application must support viewing a single model
    And the application must support deleting a model
    And the application must support adding a new model by uploading a CSV file

  Scenario: REQ-006 binds UI expectations to responsive device classes
    Parent requirement: REQ-006
    Given docs/requirements.md exists
    When the REQ-006 responsive UX notes are read
    Then expectations are stated for mobile, tablet, and desktop for list, detail, and upload flows

Feature: CSV upload validation

  Scenario: Valid minimal CSV is accepted
    Parent requirement: REQ-007
    Given a CSV file with header row "id,x,y,z"
    And row "0,0,0,0"
    When the user uploads the file to create a model
    Then the upload validation succeeds
    And no validation error is shown for id sequence or numeric coordinates

  Scenario: Reject when id sequence is not contiguous from zero
    Parent requirement: REQ-007
    Given a CSV file with header row "id,x,y,z"
    And row "0,0,0,0"
    And row "2,1,1,1"
    When the user uploads the file to create a model
    Then the upload is rejected
    And the user sees feedback that ids must be sequential starting at 0

  Scenario: Reject when more than 1000 lights
    Parent requirement: REQ-007
    Given a CSV file with header row "id,x,y,z"
    And the file contains 1001 data rows with ids 0 through 1000 and valid numeric coordinates
    When the user uploads the file to create a model
    Then the upload is rejected
    And the user sees feedback referencing the 1000 light limit

  Scenario: Reject non-numeric coordinate
    Parent requirement: REQ-007
    Given a CSV file with header row "id,x,y,z"
    And row "0,not-a-number,0,0"
    When the user uploads the file to create a model
    Then the upload is rejected
    And the user sees feedback indicating invalid numeric values

  Scenario: Reject wrong or missing columns
    Parent requirement: REQ-007
    Given a CSV file with header row "idx,x,y,z"
    When the user uploads the file to create a model
    Then the upload is rejected
    And the user sees feedback indicating the file format is incorrect

Feature: Single command build and run (REQ-008)

  Scenario: README documents the one-shot build-and-run command
    Parent requirement: REQ-008
    Given docs/requirements.md defines REQ-008
    When README.md is read
    Then it states the exact command or script path to build the static UI and run the Go server in one invocation
    And AGENTS.md references REQ-008 so the workflow stays documented for agents

  Scenario: Requirements mandate no extra manual step for standard local run
    Parent requirement: REQ-008
    Given docs/requirements.md exists
    When REQ-008 business rules are read
    Then a single documented invocation must complete static UI build for embed and start the server without requiring a second manual step for that standard session

Feature: Default geometric sample models (REQ-009)

  Scenario: Fresh data set exposes three shape samples
    Parent requirement: REQ-009
    Given the application store has no user-created models and default seeding applies
    When the user opens the model list
    Then exactly three predefined models are visible
    And their names identify a sphere, a cube, and a cone respectively

  Scenario: Consecutive lights are 10 cm apart in meters
    Parent requirement: REQ-009
    Given a default sample model with at least two lights
    When lights are ordered by id ascending
    Then the Euclidean distance between each pair of consecutive lights is 0.1 meters

  Scenario: Each sample shape is about 2 meters tall
    Parent requirement: REQ-009
    Given the default sphere, cube, and cone sample models
    When their geometry is evaluated per REQ-009 scope
    Then the sphere has diameter about 2 meters
    And the cube has edge length about 2 meters
    And the cone has height about 2 meters

  Scenario: Samples respect the 1000 light cap
    Parent requirement: REQ-009
    Given docs/requirements.md defines REQ-005 and REQ-009
    When each default sample model is inspected
    Then each has at most 1000 lights
```

---

**Next step:** After you approve these documents, invoke the `@architect` agent to update `docs/architecture.md` so implementation can proceed. When the feature is implemented, invoke the `@verifier` agent to audit, run tests, and update `docs/traceability_matrix.md`.
