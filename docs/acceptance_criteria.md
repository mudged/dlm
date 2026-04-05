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

  Scenario: REQ-005 defines chain adjacency at most two neighbors
    Parent requirement: REQ-005
    Given docs/requirements.md defines REQ-005
    When the REQ-005 business rule about adjacency along the wire is read
    Then each light has at most two logical neighbors by consecutive id only
    And the first and last light each have exactly one neighbor when n is greater than 1
    And there is no adjacency between non-consecutive ids

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

  Scenario: REQ-006 blocks model delete when model is used in scenes and explains why
    Parent requirement: REQ-006
    Given docs/requirements.md defines REQ-006 and REQ-015
    When the REQ-006 business rule about deletion and scene membership is read
    Then the system must refuse to delete a model that is assigned to one or more scenes
    And the user must be clearly informed that the model is in use by one or more scenes
    And the explanation must indicate what the user can do next such as remove the model from those scenes

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

  Scenario: Consecutive sample lights are between 5 cm and 10 cm apart
    Parent requirement: REQ-009
    Given a default sample model with at least two lights
    When lights are ordered by id ascending
    Then each Euclidean distance between consecutive lights is at least 0.05 meters
    And each such distance is at most 0.10 meters

  Scenario: Each sample has between 500 and 1000 lights
    Parent requirement: REQ-009
    Given docs/requirements.md defines REQ-005 and REQ-009
    When each default sample model is inspected
    Then each has at least 500 lights
    And each has at most 1000 lights

  Scenario: Sample lights lie on or near the outside surface not inside
    Parent requirement: REQ-009
    Given docs/requirements.md defines REQ-009
    When the REQ-009 scope and business rules about surface placement are read
    Then lights must be on the outside surface of the nominal solid or within 0.03 meters of it
    And lights must not be inside the nominal solid

  Scenario: Cube sample uses face planes with even distribution not edge-only
    Parent requirement: REQ-009
    Given docs/requirements.md defines REQ-009
    When the REQ-009 scope and business rules about the cube are read
    Then lights must be placed on the six nominal face planes
    And lights must be evenly distributed over those face areas
    And placement must not be confined to edges or vertices only

  Scenario: Sphere and cone samples evenly cover surfaces not single-curve-only
    Parent requirement: REQ-009
    Given docs/requirements.md defines REQ-009
    When the REQ-009 scope about sphere and cone coverage is read
    Then the sphere must have lights evenly distributed over spherical surface area
    And the cone must cover lateral surface and base with even distribution
    And layouts must not be edge-only or single narrow curve only

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

Feature: Three.js model view (REQ-010)

  Scenario: Requirements mandate three.js for the model view
    Parent requirement: REQ-010
    Given docs/requirements.md exists
    When requirement REQ-010 is read
    Then viewing a single model must include a 3D visualization of light positions
    And that visualization must use the three.js library as a direct front-end dependency

  Scenario: REQ-010 ties the 3D view to REQ-006 and responsive UX
    Parent requirement: REQ-010
    Given docs/requirements.md exists
    When REQ-010 scope, business rules, and responsive notes are read
    Then the three.js view is required in the same model view flow as REQ-006
    And mobile, tablet, and desktop usability expectations are stated for the 3D view

  Scenario: REQ-010 requires 2 cm spheres with white default or REQ-012 state appearance
    Parent requirement: REQ-010
    Given docs/requirements.md defines REQ-010 and REQ-012
    When the REQ-010 business rule about sphere diameter and colour is read
    Then each light must be shown as a sphere with diameter 0.02 meters
    And colour and on off appearance must follow REQ-012 when per light state is available
    And otherwise the sphere must appear white with solid fill

  Scenario: REQ-010 requires D0D0D0 light grey 85 percent transparent segments between consecutive ids
    Parent requirement: REQ-010
    Given docs/requirements.md defines REQ-005 and REQ-010
    When the REQ-010 scope and business rules about line segments are read
    Then straight segments must connect only consecutive lights i and i plus 1 for ascending ids
    And segments must use hex colour D0D0D0 as canonical light grey with 85 percent transparency meaning 15 percent opacity
    And segments must be thin barely visible and less prominent than the light spheres

  Scenario: REQ-010 requires every light drawn with previous and next connectivity
    Parent requirement: REQ-010
    Given docs/requirements.md exists
    When REQ-010 scope and business rules are read
    Then every light in the model must be drawn as a 2 cm diameter sphere per REQ-010 and REQ-012
    And interior lights along the wire must connect to previous and next via those segments

  Scenario: REQ-010 forbids omitting lights when n is at most 1000
    Parent requirement: REQ-010
    Given docs/requirements.md exists
    When REQ-010 business rule about all lights drawn is read
    Then the renderer must not skip or merge lights for performance when n is at most 1000

  Scenario: REQ-010 requires hover or touch equivalent for id and coordinates
    Parent requirement: REQ-010
    Given docs/requirements.md exists
    When the REQ-010 business rules and responsive notes are read
    Then pointer hover over a light sphere must show id and x y z
    And touch-first devices must have an equivalent to show id and coordinates

Feature: Per-light state REST API (REQ-011)

  Scenario: Requirements mandate REST API to query light state
    Parent requirement: REQ-011
    Given docs/requirements.md exists
    When requirement REQ-011 is read
    Then the product must expose REST operations to read current light state for a model
    And the API must support reading state for all lights in one response and for a single light by id

  Scenario: Requirements mandate REST API to update each light individually
    Parent requirement: REQ-011
    Given docs/requirements.md exists
    When the REQ-011 business rules are read
    Then the API must allow updating each light individually by id
    And updates must support on or off hex colour and brightness percentage
    And partial field updates must be supported where REST semantics allow

  Scenario: REQ-011 requires hex colour validation and brightness percent range
    Parent requirement: REQ-011
    Given docs/requirements.md defines REQ-011
    When the REQ-011 business rules about colour and brightness are read
    Then hex colour must use a canonical form defined in architecture
    And invalid hex values must be rejected with a clear error
    And brightness must be a percent with 0 minimum and 100 maximum for the on appearance

  Scenario: REQ-011 requires persisted state after successful writes
    Parent requirement: REQ-011
    Given docs/requirements.md defines REQ-011
    When the REQ-011 business rules are read
    Then successful writes must persist with the model for reloads and other clients

  Scenario: REQ-011 binds default state to REQ-014 and cross-surface consistency
    Parent requirement: REQ-011
    Given docs/requirements.md defines REQ-011 and REQ-014
    When the REQ-011 business rules about defaults are read
    Then default state for lights without prior state must match REQ-014
    And defaults must be consistent across API and UI

Feature: Visualization reflects light state (REQ-012)

  Scenario: On lights appear filled with hex colour and brightness
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-012
    When the REQ-012 business rules for an on light are read
    Then the sphere must appear filled opaque surface fill
    And the appearance must use the current hex colour and brightness from REQ-011

  Scenario: Off lights use D0D0D0 and 85 percent transparency like wire segments
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-010 and REQ-012
    When the REQ-012 business rules for an off light are read
    Then the sphere must use hex colour D0D0D0 with 85 percent transparency meaning 15 percent opacity
    And the appearance must be consistent with REQ-010 segment styling
    And off lights must not appear more prominent than on lights or than wire segments

  Scenario: Visualization updates after API state changes
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-012
    When the REQ-012 business rules about updates are read
    Then the 3D view must update when light state changes via REQ-011 while viewing the model
    And sphere appearance must match the latest persisted state without indefinite staleness after a successful write

  Scenario: REQ-012 preserves REQ-010 segments and hover coordinates behavior
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-010 and REQ-012
    When REQ-012 business rules are read
    Then REQ-010 segments and hover or touch id and coordinates behavior remain in force

  Scenario: REQ-012 applies defaults for lights without stored state
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-011 REQ-014 and REQ-012
    When the REQ-012 business rule about missing stored state is read
    Then lights without stored state must use the REQ-011 default aligned with REQ-014
    And they must still render per on and off rules

Feature: Model view light list pagination and bulk settings (REQ-013)

  Scenario: REQ-013 requires paginated light list and page size control
    Parent requirement: REQ-013
    Given docs/requirements.md defines REQ-013
    When the REQ-013 scope and business rules are read
    Then the model view must include a paginated list of lights when the model has more than one light
    And the user must be able to change how many lights are shown per page with at least three distinct choices between 1 and 1000 inclusive

  Scenario: REQ-013 requires jump to light by id with validation
    Parent requirement: REQ-013
    Given docs/requirements.md defines REQ-013
    When the REQ-013 business rules about jumping by id are read
    Then the user must be able to navigate to the page containing a given light id
    And invalid or out of range ids must not change the page silently
    And the user must receive actionable feedback for invalid or out of range ids

  Scenario: REQ-013 requires multi-select and bulk apply of REQ-011 fields
    Parent requirement: REQ-013
    Given docs/requirements.md defines REQ-011 and REQ-013
    When the REQ-013 business rules about multi-select are read
    Then the user must be able to select multiple lights from the list
    And pointer keyboard and touch-equivalent patterns must be addressed per REQ-002
    And the user must be able to apply the same on off hex colour and brightness to every selected light
    And validation for colour and brightness must match REQ-011

  Scenario: REQ-013 binds bulk updates to visualization timeliness
    Parent requirement: REQ-013
    Given docs/requirements.md defines REQ-012 and REQ-013
    When the REQ-013 business rule about successful bulk apply is read
    Then after a successful bulk apply the 3D view and list must reflect updated state without indefinite staleness

  Scenario: REQ-013 ties pagination and multi-select to responsive UX
    Parent requirement: REQ-013
    Given docs/requirements.md defines REQ-002 and REQ-013
    When the REQ-013 responsive UX notes are read
    Then pagination and multi-select must remain usable on mobile tablet and desktop
    And essential actions must not rely on hover-only affordances

Feature: Default light state and model reset (REQ-014)

  Scenario: REQ-014 mandates initial off state with 100 percent white
    Parent requirement: REQ-014
    Given docs/requirements.md defines REQ-014
    When the REQ-014 business rules about initial state are read
    Then every light must start in the off state with on false
    And brightness must be 100 percent
    And hex colour must be white FFFFFF in canonical six-digit form

  Scenario: REQ-014 requires a reset control on the model view
    Parent requirement: REQ-014
    Given docs/requirements.md defines REQ-006 and REQ-014
    When the REQ-014 scope and business rules are read
    Then the model view must expose a reset affordance such as a Reset button
    And the control must restore all lights in the current model in one user action

  Scenario: REQ-014 binds reset to persistence and visualization timeliness
    Parent requirement: REQ-014
    Given docs/requirements.md defines REQ-011 REQ-012 REQ-013 and REQ-014
    When the REQ-014 business rules about successful reset are read
    Then reset must persist state per REQ-011
    And the 3D view and light list must update without indefinite staleness after success

  Scenario: REQ-014 requires non-hover-only reset on all device classes
    Parent requirement: REQ-014
    Given docs/requirements.md defines REQ-002 and REQ-014
    When the REQ-014 responsive UX notes are read
    Then the reset control must be reachable on mobile tablet and desktop
    And essential use of reset must not rely on hover-only affordances

Feature: Scenes composite 3D space and management (REQ-015)

  Scenario: REQ-015 requires create with one or more models plus list delete and open
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 scope and business rules about lifecycle are read
    Then creating a scene must require a name and one or more models in the same creation flow
    And a persisted scene must always have at least one model until the whole scene is deleted
    And the application must list all scenes
    And the application must allow deleting a scene
    And the application must allow opening or selecting a scene for viewing

  Scenario: REQ-015 requires automatic offsets on create so models fit fully in scene including boundary
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 business rules about scene creation are read
    Then the system must automatically compute integer placement offsets for each model chosen at create time
    And every light of each model must lie fully within the non-negative scene region after composition
    And the create flow must not require the user to enter numeric placement offsets for initial placement
    And automatic calculation must satisfy display and framing boundary rules together with containment

  Scenario: REQ-015 keeps stored model coordinates separate from derived scene coordinates
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-005 REQ-006 and REQ-015
    When the REQ-015 business rules about coordinate systems are read
    Then persisted model light coordinates must not be rewritten when a model is used in a scene
    And scene visualization and scene-oriented API data must use positions derived from canonical coordinates plus offsets
    And the single-model view must continue to reflect the original model coordinates per REQ-006

  Scenario: REQ-015 forbids negative placement and any light outside the non-negative scene region
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 business rules about placement and containment are read
    Then placement offsets x y z must be integers greater than or equal to zero
    And every light in scene space must have coordinates greater than or equal to zero on each axis
    And the system must reject or block placements that would put any part of any model outside that region
    And the user must receive clear feedback when a placement is invalid

  Scenario: REQ-015 requires scene volume at least one meter beyond combined model bounds
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 business rule about automatic scene sizing is read
    Then the scene display volume must automatically size to enclose all placed models
    And the margin beyond the tight combined axis-aligned extent must be at least one SI meter in the sense stated in REQ-015

  Scenario: REQ-015 requires confirmation before deleting scene when removing the last model
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 business rules about removing the last model are read
    Then removing the last remaining model must not silently delete the scene
    And the user must confirm after a clear explanation that the entire scene will be deleted
    And on confirm the scene must be deleted and on cancel the scene must remain unchanged
    And the confirmation flow must not rely on hover-only essential steps per REQ-002

  Scenario: REQ-015 defaults new model to the right and allows reposition with automatic bounds update
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-015
    When the REQ-015 business rules about adding models and adjusting placement are read
    Then adding a model when at least one model already exists must default placement to the right of the existing layout per architecture
    And the scene volume must adjust automatically to satisfy containment and margin rules
    And the user must be able to change placements of models already in the scene subject to non-negative containment
    And successful placement changes must persist

  Scenario: REQ-015 requires three.js scene view analogous to model view
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-010 REQ-012 and REQ-015
    When the REQ-015 business rules about viewing a scene are read
    Then the scene view must use three.js as a direct front-end dependency
    And every light of every assigned model must be drawn with positions relative to the scene including placement
    And segments must connect consecutive ids only within the same model
    And per-light state from REQ-011 and REQ-012 must apply in the scene view

  Scenario: REQ-015 requires add and remove model from scene in scene view
    Parent requirement: REQ-015
    Given docs/requirements.md defines REQ-006 and REQ-015
    When the REQ-015 business rules about editing membership are read
    Then the scene view must allow adding an existing model to the scene
    And the scene view must allow removing a model from the scene subject to last-model confirmation rules
    And changes must persist after success
    And essential controls must not rely on hover-only interaction per REQ-002

Feature: Camera reset for 3D views (REQ-016)

  Scenario: REQ-016 requires camera reset on model and scene three.js views
    Parent requirement: REQ-016
    Given docs/requirements.md defines REQ-016
    When the REQ-016 scope and business rules are read
    Then the single-model view with three.js must expose a camera reset affordance
    And the scene view with three.js must expose a camera reset affordance
    And activating camera reset must restore the default framing per architecture
    And activating camera reset must not change persisted models scenes placements or per-light state

  Scenario: REQ-016 binds camera reset to responsive non-hover-only use
    Parent requirement: REQ-016
    Given docs/requirements.md defines REQ-002 and REQ-016
    When the REQ-016 responsive UX notes and business rules are read
    Then camera reset must be reachable on mobile tablet and desktop
    And essential use of camera reset must not rely on hover-only affordances

Feature: Options factory reset with confirmation (REQ-017)

  Scenario: REQ-017 requires an Options section with factory reset
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-017
    When the REQ-017 scope and business rules are read
    Then the product must expose an Options section or equivalent discoverable area
    And that area must include a factory reset action with unambiguous labeling

  Scenario: REQ-017 requires prompt and warning before any destructive factory reset
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-017
    When the REQ-017 business rules about confirmation are read
    Then factory reset must show a blocking prompt before irreversible effects begin
    And the prompt must warn that all models scenes routines and related data will be permanently removed
    And only default sample models will remain after completion
    And cancel or dismiss must leave data unchanged

  Scenario: REQ-017 requires post-reset state to match fresh samples
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-009 REQ-011 REQ-014 REQ-017 and REQ-021
    When the REQ-017 business rules about outcomes are read
    Then after confirmed factory reset no user-created models or scenes may remain in listings
    And no routine definitions or persisted routine run state from REQ-021 may remain
    And the model list must satisfy REQ-009 expectations for a fresh seed with three identifiable samples
    And per-light defaults for present models must align with REQ-014 and REQ-011

  Scenario: REQ-017 factory reset flow is usable without hover-only steps
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-002 and REQ-017
    When the REQ-017 business rule about REQ-002 is read
    Then opening Options starting factory reset and confirming or canceling must not require hover-only essential steps

Feature: Application shell themes navigation branding and Font Awesome (REQ-018)

  Scenario: REQ-018 mandates light and dark themes with specified surfaces
    Parent requirement: REQ-018
    Given docs/requirements.md defines REQ-018
    When requirement REQ-018 is read
    Then the product must expose both light and dark themes with a discoverable way to switch between them
    And light theme must use a white or white-equivalent main application background with dark primary text
    And dark theme must use a dark grey main application background with white or near-white primary text

  Scenario: REQ-018 requires collapsible left navigation toggled by a burger control
    Parent requirement: REQ-018
    Given docs/requirements.md defines REQ-002 and REQ-018
    When the REQ-018 scope and business rules about navigation are read
    Then primary navigation must be in a left region that collapses and expands
    And a burger button must toggle collapse and expand
    And the menu and burger must remain usable on touch devices per REQ-002

  Scenario: REQ-018 fixes application title and Font Awesome regular lightbulb logo
    Parent requirement: REQ-018
    Given docs/requirements.md defines REQ-018
    When the REQ-018 business rules about branding are read
    Then the visible application title must be exactly Domestic Light & Magic
    And the application logo must use the Font Awesome classic regular lightbulb icon as linked from Font Awesome lightbulb classic regular

  Scenario: REQ-018 requires Font Awesome icons on buttons
    Parent requirement: REQ-018
    Given docs/requirements.md defines REQ-018
    When the REQ-018 business rule about buttons is read
    Then button elements and button-styled action controls must show a Font Awesome icon in the visible UI
    And essential theme navigation and menu actions must not rely on hover-only steps per REQ-002

  Scenario: REQ-018 defaults theme to system preference until user overrides
    Parent requirement: REQ-018
    Given docs/requirements.md defines REQ-018
    When the REQ-018 business rule 1 about default versus override is read
    Then on first load the application must follow the user's system light versus dark preference where available for example prefers-color-scheme
    And after the user manually chooses light or dark that choice must persist across sessions and override the system signal until the user changes it again or a documented reset path exists

Feature: Three.js fixed dark-grey viewport (REQ-019)

  Scenario: REQ-019 requires dark-grey three.js backdrop for model and scene in both shell themes
    Parent requirement: REQ-019
    Given docs/requirements.md defines REQ-010 REQ-015 REQ-018 and REQ-019
    When the REQ-019 scope and business rules are read
    Then the WebGL rendering surface for single-model detail and scene composite detail must use a dark grey background not white or near-white
    And that policy must apply when REQ-018 shell theme is light and when it is dark

  Scenario: REQ-019 keeps three.js viewport responsive per REQ-002
    Parent requirement: REQ-019
    Given docs/requirements.md defines REQ-002 and REQ-019
    When the REQ-019 business rule 3 and responsive notes are read
    Then the 3D viewport and its controls must remain usable on mobile tablet and desktop without hover-only essential steps

Feature: Scene spatial API dimensions filters and bulk updates (REQ-020)

  Scenario: REQ-020 exposes scene dimensions
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-020
    When the REQ-020 business rules are read
    Then the scene API must include a read operation that returns scene dimensions
    And dimension values must be unambiguous in numeric meaning and units policy

  Scenario: REQ-020 returns all scene lights in scene coordinates
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-005 REQ-015 and REQ-020
    When the REQ-020 business rules for all-lights retrieval are read
    Then the scene API must include a read operation returning all lights in a scene
    And each returned coordinate must be in scene space derived from model coordinates plus scene placement
    And original canonical model coordinates must not be rewritten by this operation

  Scenario: REQ-020 supports cuboid-based light retrieval in scene space
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-020
    When the REQ-020 business rules for cuboid retrieval are read
    Then the scene API must include a read operation that accepts a cuboid position and dimensions in scene space
    And only lights within that cuboid are returned

  Scenario: REQ-020 supports sphere-based light retrieval in scene space
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-020
    When the REQ-020 business rules for sphere retrieval are read
    Then the scene API must include a read operation that accepts a sphere in scene space
    And only lights within that sphere are returned

  Scenario: REQ-020 supports cuboid-based bulk light updates in scene space
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-011 and REQ-020
    When the REQ-020 business rules for cuboid bulk update are read
    Then the scene API must include a bulk update operation for all lights within a cuboid in scene space
    And updated light-state fields must follow REQ-011 semantics and validation

  Scenario: REQ-020 supports sphere-based bulk light updates in scene space
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-011 and REQ-020
    When the REQ-020 business rules for sphere bulk update are read
    Then the scene API must include a bulk update operation for all lights within a sphere in scene space
    And updated light-state fields must follow REQ-011 semantics and validation

  Scenario: REQ-020 rejects invalid region geometry without partial update
    Parent requirement: REQ-020
    Given docs/requirements.md defines REQ-020
    When region input contains non-finite values or non-positive dimensions or radius
    Then the API request must be rejected with a clear actionable error
    And no partial updates may be persisted

Feature: Scene routines definitions run stop and first random colour type (REQ-021)

  Scenario: REQ-021 requires routine definitions with name description and type plus list and delete
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-021
    When the REQ-021 scope and business rules about definitions are read
    Then the product must support creating a routine with name description and type
    And the product must support listing all routine definitions
    And the product must support deleting a routine definition
    And name and type are required at create time

  Scenario: REQ-021 requires starting and stopping a routine against a scene
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-021
    When the REQ-021 business rules about run lifecycle are read
    Then the product must support starting a routine run scoped to exactly one existing scene
    And the product must support stopping an active run
    And start must fail with clear actionable errors when the scene does not exist or is not usable

  Scenario: REQ-021 binds running automation to scene API for light state changes
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-011 REQ-020 and REQ-021
    When the REQ-021 business rule about scene API usage during an active run is read
    Then automated changes to on off hex colour or brightness for lights in that scene must use the scene API surface from REQ-020
    And canonical stored model coordinates must not be rewritten by routine automation

  Scenario: REQ-021 allows volumetric targeting in scene space for types that need it
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-020 and REQ-021
    When the REQ-021 business rules about volumetric targeting are read
    Then routine types that limit effects must evaluate inclusion using scene-space positions
    And they must use cuboid and or sphere geometry consistent with REQ-020
    And invalid region geometry must be rejected per REQ-020 expectations

  Scenario: REQ-021 first routine type turns on all scene lights and randomizes colours about every second
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-011 REQ-015 REQ-020 and REQ-021
    When the REQ-021 business rules for the first routine type are read
    Then starting that type on a scene must set every light in the scene on with brightness 100 percent and valid hex colour per REQ-011
    And while the run remains active at most once per elapsed SI second each light must receive a new independently uniformly random REQ-011-valid hex colour
    And the approximate one-second cadence must be documented in docs/architecture.md

  Scenario: REQ-021 stopping the first type ends automation without implicit reset to defaults
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-011 REQ-014 and REQ-021
    When the REQ-021 business rules about stopping the first type are read
    Then stopping must cease further automated updates promptly
    And lights retain the last successfully persisted state per REQ-011
    And stopping must not by itself reset lights to REQ-014 defaults

  Scenario: REQ-021 ties routine UI to responsive non-hover-only use when exposed
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-002 and REQ-021
    When the REQ-021 responsive UX notes are read
    Then any UI for list create delete start and stop must be usable on mobile tablet and desktop without hover-only essential steps

  Scenario: REQ-017 factory reset removes routine data per REQ-021 scope
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-017 and REQ-021
    When the REQ-017 scope about factory reset data removal is read
    Then factory reset must remove persisted scene routine definitions and any persisted routine run state together with models scenes and related data
```

---

**Next step:** After you approve these documents, invoke the `@architect` agent to update `docs/architecture.md` so implementation can proceed. When the feature is implemented, invoke the `@verifier` agent to audit, run tests, and update `docs/traceability_matrix.md`.
