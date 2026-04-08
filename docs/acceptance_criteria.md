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

  Scenario: REQ-011 requires authoritative in-memory state after successful writes
    Parent requirement: REQ-011
    Given docs/requirements.md defines REQ-011 and REQ-039
    When the REQ-011 business rules are read
    Then successful writes must update authoritative in-memory per-light state on the server
    And per-light operational state must not be stored in durable application storage for reload after restart
    And other API clients and connected visualization clients must be able to observe the updated state without a process restart

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
    And sphere appearance must match the latest authoritative server state per REQ-039 without indefinite staleness after a successful write

  Scenario: REQ-012 preserves REQ-010 segments and hover coordinates behavior
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-010 and REQ-012
    When REQ-012 business rules are read
    Then REQ-010 segments and hover or touch id and coordinates behavior remain in force

  Scenario: REQ-012 applies defaults for lights without established authoritative state
    Parent requirement: REQ-012
    Given docs/requirements.md defines REQ-011 REQ-014 REQ-012 and REQ-039
    When the REQ-012 business rule about missing stored state is read
    Then lights without established authoritative state in the running service must use the REQ-011 default aligned with REQ-014
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

  Scenario: REQ-014 binds reset to authoritative state and visualization timeliness
    Parent requirement: REQ-014
    Given docs/requirements.md defines REQ-011 REQ-012 REQ-013 REQ-014 and REQ-039
    When the REQ-014 business rules about successful reset are read
    Then reset must update authoritative in-memory state per REQ-011 and REQ-039
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
    And activating camera reset must not change persisted models scenes placements or authoritative per-light state

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
    Given docs/requirements.md defines REQ-017 REQ-033 and REQ-032
    When the REQ-017 business rules about confirmation are read
    Then factory reset must show a blocking prompt before irreversible effects begin
    And the prompt must warn that all models scenes registered devices routines including Python and shape animation routines and related data will be permanently removed
    And only default sample models and default sample Python routines will remain after completion
    And cancel or dismiss must leave data unchanged

  Scenario: REQ-017 requires post-reset state to match fresh samples
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-009 REQ-011 REQ-014 REQ-017 REQ-021 REQ-022 REQ-033 and REQ-032
    When the REQ-017 business rules about outcomes are read
    Then after confirmed factory reset no user-created models or scenes may remain in listings
    And no registered devices or device-model assignments may remain per REQ-035 and REQ-036
    And no routine definitions beyond the three default Python sample routines from REQ-032 may remain including no user shape animation definitions per REQ-033 and no persisted routine run state from prior entities may remain
    And the model list must satisfy REQ-009 expectations for a fresh seed with three identifiable samples
    And the routine definition list must satisfy REQ-032 expectations for exactly three default Python sample routines
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

Feature: Scene routines Python definitions run stop and scene API (REQ-021)

  Scenario: REQ-021 requires routine definitions with name description and Python or shape animation kind plus list and delete
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-021 REQ-022 and REQ-033
    When the REQ-021 scope and business rules about definitions are read
    Then the product must support creating a routine with name description and routine kind either Python per REQ-022 or shape animation per REQ-033
    And the product must support listing all routine definitions
    And the product must support deleting a routine definition
    And name is required at create time

  Scenario: REQ-021 requires starting and stopping a routine against a scene
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-021
    When the REQ-021 business rules about run lifecycle are read
    Then the product must support starting a routine run scoped to exactly one existing scene
    And the product must support stopping an active run
    And start must fail with clear actionable errors when the scene does not exist or is not usable

  Scenario: REQ-021 binds running automation to scene API for light state changes
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-011 REQ-020 REQ-021 REQ-022 and REQ-033
    When the REQ-021 business rule about scene API usage during an active run is read
    Then automated changes to on off hex colour or brightness for lights in that scene from Python runs must be effected through the Python scene binding and underlying REQ-020 scene API surface
    And automated changes from shape animation runs must use REQ-020 equivalent native paths that preserve REQ-011 semantics without executing user Python
    And canonical stored model coordinates must not be rewritten by routine automation

  Scenario: REQ-021 coordinates with device requirements for physical output
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-021 REQ-035 REQ-036 REQ-038 and REQ-039
    When the REQ-021 scope about physical devices is read
    Then physical output for models with assigned devices must follow REQ-035 through REQ-038 without REQ-021 restating device registry or WLED protocol details

  Scenario: REQ-021 allows volumetric targeting in scene space when scripts use regions
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-020 and REQ-021
    When the REQ-021 business rules about volumetric targeting are read
    Then inclusion for region-scoped script logic must evaluate using scene-space positions
    And cuboid and or sphere geometry must be consistent with REQ-020
    And invalid region geometry must be rejected per REQ-020 expectations

  Scenario: REQ-032 default random colour cycle routine matches former all-lights test routine behavior
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-011 REQ-015 REQ-020 REQ-021 REQ-022 and REQ-032
    When the REQ-032 business rules for the random colour cycle all scene lights routine are read
    Then starting that seeded Python routine on a scene must set every light in the scene on with brightness 100 percent and valid hex colour per REQ-011
    And while the run remains active at most once per elapsed SI second each light must receive a new independently uniformly random REQ-011-valid hex colour
    And the approximate one-second cadence must be documented in docs/architecture.md

  Scenario: REQ-032 stopping the random colour cycle routine ends automation without implicit reset to defaults
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-011 REQ-014 REQ-021 REQ-022 and REQ-032
    When the REQ-032 business rules for stopping the random colour cycle routine are read
    Then stopping must cease further automated updates promptly
    And lights retain the last successfully applied authoritative in-memory state per REQ-011 and REQ-039
    And stopping must not by itself reset lights to REQ-014 defaults

  Scenario: REQ-021 ties routine UI to responsive non-hover-only use when exposed
    Parent requirement: REQ-021
    Given docs/requirements.md defines REQ-002 and REQ-021
    When the REQ-021 responsive UX notes are read
    Then any UI for list create delete start and stop must be usable on mobile tablet and desktop without hover-only essential steps

  Scenario: REQ-017 factory reset removes routine data per REQ-021 REQ-022 and REQ-033 scope
    Parent requirement: REQ-017
    Given docs/requirements.md defines REQ-017 REQ-021 REQ-022 REQ-033 and REQ-032
    When the REQ-017 scope about factory reset data removal is read
    Then factory reset must remove persisted scene routine definitions including shape animation per REQ-033 and any persisted routine run state together with models scenes and related data
    And factory reset must re-seed exactly the three default Python sample routines defined in REQ-032

Feature: Python scene routines editor API docs and execution (REQ-022)

  Scenario: REQ-022 requires in-browser Python editor with highlighting checking completion and formatting
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 scope and business rules about the editor are read
    Then the product must provide an in-browser Python code editor with full syntax highlighting
    And the Python authoring editor must use CodeMirror major version 6 in the browser for the editable buffer
    And the editor must surface syntax or static issues to the user without requiring a separate desktop tool
    And code completion appropriate for Python and the scene API must be enabled by default where technically feasible
    And automatic code formatting must be available and enabled by default where the product supports it

  Scenario: REQ-022 mandates CodeMirror 6 for browser Python editing
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 business rules about the in-browser editor stack are read
    Then the application must use CodeMirror 6 for Python editing in the browser on the Python routine authoring surface

  Scenario: REQ-022 requires save load duplicate and delete for Python routine definitions
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 business rules about persistence are read
    Then the application must provide save load duplicate and delete for Python routine definitions

  Scenario: REQ-022 requires documented Python scene library on the same page as the editor
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022 REQ-020 REQ-024 and REQ-011
    When the REQ-022 business rules about documentation are read
    Then reference documentation for the Python scene library must appear on the same page as the editor
    And the documentation must satisfy REQ-024 for placement below the code editor and for selectable entries commented samples and insert-at-caret-or-end behavior
    And the documentation must target novices with plain language and parameter descriptions
    And the Python API must map to scene capabilities consistent with REQ-020 and REQ-011 light-state semantics

  Scenario: REQ-022 requires instructional wording for a twelve-year-old beginner in Python
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 business rule 10 about user-visible instructional text is read
    Then the Python routine authoring surface must require wording understandable to a twelve-year-old who has just started learning Python for headings labels primary tooltips empty states and short inline help
    And that wording must use short sentences everyday words and brief explanations for unavoidable specialist terms without long expository paragraphs in the chrome around the editor

  Scenario: REQ-022 requires illustrative scene API names documented in architecture and on-page reference
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 scope about the Python library is read
    Then illustrative examples include a documented attribute or method for scene vertical extent akin to scene height
    And illustrative examples include a documented method for sphere-filtered light retrieval in scene space akin to scene getLightsWithinSphere
    And docs architecture.md and the on-page reference must list canonical names if they differ from those examples

  Scenario: REQ-022 requires continuous loop execution while run is active
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 business rules about starting a Python routine against a scene are read
    Then while the run remains active the implementation must repeatedly execute the user script in a loop
    And docs architecture.md must document iteration timing and fairness with other runs

  Scenario: REQ-022 requires cooperative stop plus forcible termination
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-022
    When the REQ-022 business rules about stopping a Python routine run are read
    Then stopping must cease further loop iterations promptly under normal conditions
    And the implementation must support forcible termination when the routine does not respond to cooperative stop within architecture-defined bounds
    And docs architecture.md must document those bounds and the termination mechanism

  Scenario: REQ-022 forbids rewriting canonical model coordinates from Python automation
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-005 REQ-015 and REQ-022
    When the REQ-022 business rule about automation and the scene API is read
    Then Python routine automation must affect lights only through the documented scene API surface
    And it must not rewrite canonical stored model coordinates

  Scenario: REQ-022 ties editor docs and run controls to responsive non-hover-only use
    Parent requirement: REQ-022
    Given docs/requirements.md defines REQ-002 and REQ-022
    When the REQ-022 business rule 11 and responsive UX notes are read
    Then the editor documentation and run and stop controls must remain usable on mobile tablet and desktop
    And essential steps must not rely on hover-only affordances

Feature: Create routine kind choice Python or shape animation (REQ-023)

  Scenario: REQ-023 requires new routine creation to offer Python or shape animation only
    Parent requirement: REQ-023
    Given docs/requirements.md defines REQ-021 REQ-022 REQ-033 and REQ-023
    When the REQ-023 scope and business rules about new routine creation are read
    Then the user-facing flow to create a new routine definition must let the user choose Python per REQ-022 or shape animation per REQ-033
    And the product must not offer any third creatable routine kind beyond those two

  Scenario: REQ-023 limits type control options to the two defined kinds
    Parent requirement: REQ-023
    Given docs/requirements.md defines REQ-023
    When the REQ-023 business rule 2 is read
    Then if the UI shows a type or kind control its options must map only to Python and shape animation

  Scenario: REQ-023 forbids redundant duplicate primary create actions for the same flow
    Parent requirement: REQ-023
    Given docs/requirements.md defines REQ-023
    When the REQ-023 business rule 3 is read
    Then the UI must not present redundant standalone primary actions that start the same authoring flow

  Scenario: REQ-023 ties create flow controls to responsive non-hover-only use
    Parent requirement: REQ-023
    Given docs/requirements.md defines REQ-002 and REQ-023
    When the REQ-023 business rule 4 and responsive UX notes are read
    Then any create control must be operable on mobile tablet and desktop without hover-only essential steps

Feature: Python routine API reference below editor (REQ-024)

  Scenario: REQ-024 requires API reference directly below the code editor
    Parent requirement: REQ-024
    Given docs/requirements.md defines REQ-002 REQ-022 and REQ-024
    When the REQ-024 scope and business rules are read
    Then the Python routine view must place the API reference section directly below the code editor in vertical document order with no other primary workflow block between editor and reference except minimal separators or headings per architecture
    And the catalog must enumerate every Python-exposed public scene API surface element for routines with no deliberate omissions

  Scenario: REQ-024 requires selectable API entries with detail sample usage commented snippets and insert control
    Parent requirement: REQ-024
    Given docs/requirements.md defines REQ-022 and REQ-024
    When the REQ-024 business rules about catalog interaction and samples are read
    Then the user must be able to choose one function method or documented attribute at a time to view expanded detail and sample usage for that item
    And every sample code snippet shown in the reference must include Python hash comments that briefly describe what the code does without being verbose
    And the product must expose a control that inserts the currently shown example into the editor at the caret when the caret is active in the editor otherwise at the end of the buffer

  Scenario: REQ-024 catalog documents API items and does not replace REQ-032 default routines
    Parent requirement: REQ-024
    Given docs/requirements.md defines REQ-024 and REQ-032
    When the REQ-024 business rule 7 is read
    Then per-entry samples document individual API items
    And the three full default Python sample routines from REQ-032 are not required to appear as whole-script REQ-024 catalog entries

  Scenario: REQ-024 API reference remains usable on all device classes
    Parent requirement: REQ-024
    Given docs/requirements.md defines REQ-002 REQ-022 and REQ-024
    When the REQ-024 business rule 6 and responsive UX notes are read
    Then the reference region selector and insert control must remain usable on mobile tablet and desktop without hover-only essential steps
    And the section must remain readable and scrollable without horizontal scrolling for primary snippet content where avoidable

Feature: Python routine default sphere colour template (REQ-025)

  Scenario: REQ-025 requires new Python routines to open with sphere-region colour change template
    Parent requirement: REQ-025
    Given docs/requirements.md defines REQ-011 REQ-020 REQ-022 REQ-024 and REQ-025
    When the REQ-025 scope and business rules are read
    Then newly created Python routine definitions must open with default template code whose primary illustrated behavior is changing colours for lights inside a sphere region in scene space
    And the template must use sphere targeting consistent with REQ-020 and must not rewrite canonical model coordinates
    And the template must include Python hash comments that briefly explain each main step to the same brevity standard as REQ-024 samples

Feature: Python scene binding width depth height (REQ-026)

  Scenario: REQ-026 requires scene width depth and height on the Python binding
    Parent requirement: REQ-026
    Given docs/requirements.md defines REQ-020 REQ-022 REQ-024 and REQ-026
    When the REQ-026 business rules are read
    Then the Python scene object must expose height and must also expose width and depth or equivalent documented names for all three axis-aligned extents
    And values must align with REQ-020 dimension semantics for the same scene snapshot
    And docs architecture.md the REQ-022 on-page reference and the REQ-024 API reference must document all three attributes including which world axis each maps to

Feature: Python routine unified run-in-scene and live viewport (REQ-027)

  Scenario: REQ-027 requires one unified region for scene target run stop and live three.js viewport
    Parent requirement: REQ-027
    Given docs/requirements.md defines REQ-002 REQ-010 REQ-012 REQ-015 REQ-019 REQ-022 and REQ-027
    When the REQ-027 scope and business rules are read
    Then the product must present exactly one unified run scene viewport region on the Python routine authoring surface
    And that region must combine scene selection for routine execution run or stop controls tied to that scene and the live three.js viewport showing the same scene
    And the product must not split run in scene from visual debug into parallel sections with duplicate scene pickers or viewports for the same workflow
    And light state changes from the routine must become visible in the viewport within the same class of timeliness as REQ-012 after successful writes
    And the viewport must follow REQ-010 REQ-012 REQ-015 and REQ-019 visual rules for scene composite views as applicable

  Scenario: REQ-027 requires reset scene lights and reset camera controls
    Parent requirement: REQ-027
    Given docs/requirements.md defines REQ-011 REQ-014 REQ-016 REQ-018 and REQ-027
    When the REQ-027 business rules about reset actions are read
    Then a reset scene lights control must set every light in the selected scene to REQ-014 defaults and update authoritative in-memory state per REQ-011 and REQ-039 without changing scene membership placements or canonical model coordinates
    And a reset camera control must restore default client navigation for that viewport only per REQ-016 semantics without altering persisted models scenes placements or authoritative per-light state
    And REQ-018 applies where reset actions are implemented as buttons

  Scenario: REQ-027 unified region remains usable without hover-only essential steps
    Parent requirement: REQ-027
    Given docs/requirements.md defines REQ-002 and REQ-027
    When the REQ-027 business rule 6 and responsive UX notes are read
    Then scene selection run stop viewport and both reset actions must be reachable on mobile tablet and desktop without hover-only essential steps

Feature: Three.js emissive glow scaled by brightness (REQ-028)

  Scenario: REQ-028 requires emissive appearance for on light spheres in three.js views
    Parent requirement: REQ-028
    Given docs/requirements.md defines REQ-010 REQ-012 REQ-015 REQ-027 and REQ-028
    When the REQ-028 scope and business rules are read
    Then on lights must use a material with a clear emissive light-emitting component in the single-model view scene composite view and Python routine unified live viewport where REQ-012 spheres apply
    And the sphere must read as emitting light not only as a diffuse tinted surface

  Scenario: REQ-028 ties glow strength to REQ-011 brightness with strong appearance at 100 percent
    Parent requirement: REQ-028
    Given docs/requirements.md defines REQ-011 and REQ-028
    When the REQ-028 business rules about brightness scaling are read
    Then emissive strength or documented equivalent must scale with brightness percentage from zero through one hundred
    And at 100 percent brightness the glow must be visibly strong
    And at lower percents the glow must be weaker in a perceptibly dimmer way

  Scenario: REQ-028 requires monotonic glow versus brightness for the same hex colour
    Parent requirement: REQ-028
    Given docs/requirements.md defines REQ-028
    When the REQ-028 business rule about ordering is read
    Then for two on lights with the same hex colour the higher brightness must not appear less glowing than the lower brightness

  Scenario: REQ-028 preserves REQ-012 off appearance without glow like on lights
    Parent requirement: REQ-028
    Given docs/requirements.md defines REQ-010 REQ-012 and REQ-028
    When the REQ-028 business rules for off lights are read
    Then off lights must keep REQ-012 dim grey transparent styling
    And emissive contribution for off lights must remain negligible so they do not glow like on lights or outshine wire segments

  Scenario: REQ-028 binds visualization timeliness and architecture documentation
    Parent requirement: REQ-028
    Given docs/requirements.md defines REQ-003 REQ-011 REQ-012 and REQ-028
    When the REQ-028 business rules about updates and architecture are read
    Then glow must follow authoritative server light state per REQ-039 without indefinite staleness after successful writes consistent with REQ-012
    And docs architecture.md must describe the three.js material approach and brightness mapping including Pi WebGL notes where relevant

Feature: High-throughput light updates (REQ-029)

  Scenario: REQ-029 states scale assumptions for many lights and frequent updates
    Parent requirement: REQ-029
    Given docs/requirements.md defines REQ-005 REQ-011 and REQ-029
    When the REQ-029 scope is read
    Then the design target includes on the order of hundreds of lights consistent with REQ-005 scale
    And multiple aggregate update cycles per second across writes and or viewer refresh are expected

  Scenario: REQ-029 requires architecture to document write path observer path and transport considerations
    Parent requirement: REQ-029
    Given docs/requirements.md defines REQ-003 REQ-011 REQ-012 REQ-020 and REQ-029
    When the REQ-029 business rules are read
    Then docs architecture.md must describe how high-throughput light updates are met including write path and observer path
    And connection reuse or HTTP features and batch bulk or push versus polling rationale must be addressed per REQ-029 rules

  Scenario: REQ-029 requires aggregate update paths while preserving REQ-011 per-light operations
    Parent requirement: REQ-029
    Given docs/requirements.md defines REQ-011 REQ-020 and REQ-029
    When the REQ-029 business rule about integrator update paths is read
    Then documented aggregate update paths must exist so integrators are not limited to one HTTP request per light as the only option for high-frequency multi-light changes
    And REQ-011 per-light read and write operations remain required for granular control

  Scenario: REQ-029 ties high-throughput expectations to Pi deployment constraints
    Parent requirement: REQ-029
    Given docs/requirements.md defines REQ-003 and REQ-029
    When the REQ-029 scope is read
    Then solutions for high-throughput light updates must remain plausible on Raspberry Pi 4 constraints from REQ-003

Feature: Python scene API random hex colour helper (REQ-030)

  Scenario: REQ-030 requires a documented random colour helper on the Python scene binding
    Parent requirement: REQ-030
    Given docs/requirements.md defines REQ-022 REQ-011 and REQ-030
    When the REQ-030 scope and business rules are read
    Then the Python routine scene surface must expose exactly one primary documented callable for a random hex colour suitable for light state color fields
    And docs architecture.md must name the callable and its sync or async semantics

  Scenario: REQ-030 matches uniform 24-bit hex distribution and REQ-011 string shape
    Parent requirement: REQ-030
    Given docs/requirements.md defines REQ-011 and REQ-030
    When the REQ-030 business rule about return value formatting is read
    Then each call must return a string equivalent to formatting one integer uniformly from 0 through 0xFFFFFF as "#%06x" % that integer
    And the string must be valid as REQ-011 color with hash plus six hex digits

  Scenario: REQ-030 requires REQ-024 catalog completion and worker editor alignment
    Parent requirement: REQ-030
    Given docs/requirements.md defines REQ-022 REQ-024 and REQ-030
    When the REQ-030 business rules about documentation and tooling are read
    Then the REQ-024 API catalog must list the callable with a commented sample per REQ-024
    And CodeMirror completions and the scene worker must stay aligned with the chosen Python name and async or sync semantics

Feature: Redundant light-state skips (visualization, authoritative state, device traffic) (REQ-031)

  Scenario: REQ-031 requires skipping unnecessary visualization work when state is unchanged
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-031
    When the REQ-031 business rules are read
    Then the client must compare incoming or locally applied per-light state to the last applied effective rendering state
    And when on off hex colour and brightness are equivalent after documented normalization the client must not perform a full visualization rebuild solely to reflect that same state again

  Scenario: REQ-031 should reduce redundant updates when authoritative state would not change
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-011 REQ-020 REQ-031 and REQ-039
    When the REQ-031 business rules about authoritative state updates are read
    Then the product should avoid applying redundant per-light state updates when the proposed state is equivalent to authoritative in-memory state after documented normalization
    And observable API behavior for such no-op cases must be documented in docs/architecture.md

  Scenario: REQ-031 aligns device traffic skips with REQ-035 through REQ-038
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-031 REQ-035 REQ-036 REQ-038 and REQ-039
    When the REQ-031 business rules about physical devices are read
    Then docs/architecture.md must describe how assigned devices receive only changes not equivalent to last applied on the device or channel
    And equivalence rules must align with REQ-031 rules for visualization and server state where possible

  Scenario: REQ-031 encourages in-memory last-applied or last-known state with invalidation
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-031
    When the REQ-031 business rules about caching are read
    Then the product should maintain in-memory or equivalent records of last-applied or last-known effective per-light state on client and or server as architecture defines
    And those records must be cleared or resynchronized on navigation model or scene change or architecture-defined invalidation events

  Scenario: REQ-031 preserves REQ-012 when state actually changes
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-012 and REQ-031
    When per-light state differs from the last applied effective state
    Then the visualization must update without indefinite staleness after successful writes the client knows about
    And REQ-010 REQ-015 and REQ-027 drawing rules remain in force
    And authoritative in-memory state and API read results must reflect real updates without skipping them due to stale caches

  Scenario: REQ-031 requires architecture documentation of equivalence cache and no-op semantics
    Parent requirement: REQ-031
    Given docs/requirements.md defines REQ-029 and REQ-031
    When docs/architecture.md is read after the architect pass
    Then it describes where equivalence is evaluated what is cached invalidation rules and documented no-op update behavior including device push skips where applicable
    And it aligns with REQ-029 observer and refresh strategy where relevant

Feature: Default seeded Python sample routines (REQ-032)

  Scenario: REQ-032 seeds three default Python routines on fresh install and factory reset
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-017 REQ-021 REQ-022 and REQ-032
    When the REQ-032 scope and business rules about automatic creation are read
    Then exactly three Python routine definitions must be created on first start with an empty routine store
    And exactly three Python routine definitions must exist after confirmed factory reset per REQ-017
    And the user must be able to recognize growing sphere sweeping cuboid and random colour cycle all scene lights from names or descriptions
    And each seeded routine must be editable and duplicable like any other Python routine per REQ-022
    And the primary delivery must not be as whole-script entries in the REQ-024 API reference catalog

  Scenario: REQ-032 mandates three novice-oriented routines using public scene API only
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-032
    When the REQ-032 scope and business rules are read
    Then the product must ship three complete default Python routines for growing sphere sweeping cuboid and random colour cycle all scene lights
    And each routine must use only the public Python scene API and documented helpers such as REQ-030 as named in architecture and REQ-024
    And each routine must include frequent short hash comments aimed at novice readers consistent with REQ-024 sample comment style

  Scenario: Growing sphere routine centers fills scene over ten seconds and loops with new random colour
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-011 REQ-020 REQ-022 REQ-026 REQ-030 and REQ-032
    When the growing sphere routine behavior described in REQ-032 business rule 3 is read
    Then each cycle must use a new independent random REQ-011 valid hex colour
    And the sphere must be centered at the geometric center of the scene axis aligned extent per REQ-026 and REQ-020
    And over ten SI seconds the sphere radius must increase monotonically from a small positive value until every scene light lies inside or on the closed sphere per REQ-020 inclusion semantics
    And while growing every light inside the current closed sphere must be on with brightness 100 percent and the cycle hex colour
    And after growth completes a new cycle must begin immediately with a new small sphere and new random colour while the run remains active

  Scenario: Sweeping cuboid routine spans scene width and depth with twenty cm height and turns off exited lights
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-011 REQ-020 REQ-022 REQ-026 REQ-030 and REQ-032
    When the sweeping cuboid routine behavior described in REQ-032 business rule 4 is read
    Then each cycle must use a new independent random REQ-011 valid hex colour
    And the cuboid must have width and depth equal to scene width and depth and height exactly 0.2 meters
    And each cycle must start with the cuboid at the bottom of the scene volume and over ten SI seconds move monotonically to the top without leaving scene bounds
    And at each update every light inside or on the closed cuboid must be on with brightness 100 percent and the cycle colour
    And any light that was inside the cuboid on a prior update in the same cycle but is no longer inside must be set off per REQ-011
    And after reaching the top a new cycle must start at the bottom with a new random colour while the run remains active

  Scenario: REQ-032 default routines must not rewrite canonical model coordinates
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-005 REQ-015 REQ-020 and REQ-032
    When the REQ-032 business rule about scene API only is read
    Then all three default routines must affect lights only through scene space operations
    And canonical stored model coordinates must not be rewritten by those routines

  Scenario: Architecture documents where REQ-032 seed content is defined
    Parent requirement: REQ-032
    Given docs/requirements.md defines REQ-032 and REQ-025
    When docs/architecture.md is read after the architect pass
    Then it names where the REQ-032 initial seed content is defined and how it relates to the default new Python routine template if applicable

Feature: Shape animation routines declarative authoring and run (REQ-033)

  Scenario: REQ-033 defines second routine kind with name description and parameters
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-033
    When the REQ-033 scope about persisted definitions is read
    Then shape animation is a routine kind alongside Python per REQ-021
    And each definition has required name optional description per architecture and structured parameters in business rules

  Scenario: REQ-033 add and edit includes unified run on scene viewport per REQ-027
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-027 and REQ-033
    When the REQ-033 scope about the authoring surface is read
    Then the shape animation add or edit flow must include the unified region with scene selection run stop live three.js viewport reset scene lights and reset camera
    And the product must not duplicate scene picker or viewport for that workflow

  Scenario: REQ-033 constrains shapes count type size colour brightness speed position motion and edge behavior
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-033 REQ-020 and REQ-026
    When the REQ-033 business rules 2 through 7 are read
    Then the definition allows between 1 and 20 shapes each sphere or axis-aligned cuboid in scene space
    And size may be fixed or random uniform between user lower and upper bounds per shape
    And shape colour may be fixed REQ-011 hex or random with REQ-030 distribution intent
    And each shape has user-set brightness percent 0 through 100 per REQ-011 for lights inside that shape
    And motion uses a normalized direction from signed dx dy dz not all zero and scalar speed in SI meters per second where the UI may use centimeters per second with consistent conversion
    And speed may be fixed positive or random uniform between positive lower and upper bounds on run start and each loop cycle per business rule 10
    And initial position may be explicit scene coordinates or random against a chosen scene face top bottom left right back or front with whole shape inside the scene volume
    And per shape edge behavior is one of Pac-Man wrap stop and disappear deflect random angle or deflect inflection angle against scene boundary

  Scenario: REQ-033 assigns lights inside shapes to shape colour brightness and others to background or off
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-011 REQ-033 and REQ-020
    When the REQ-033 business rules 1 4 8 and 9 are read
    Then lights inside at least one active shape receive that shapes winning colour and per-shape brightness per deterministic overlap precedence
    And lights outside all active shapes receive background colour and brightness or are set off when background mode is none

  Scenario: REQ-033 animation loops until stop or all shapes terminal
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-033
    When the REQ-033 business rule 10 is read
    Then while the run is active the animation repeats in a loop until the user stops or no active shapes remain when all used stop and disappear

  Scenario: REQ-033 preserves canonical model coordinates and uses scene API semantics
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-005 REQ-015 REQ-020 REQ-021 and REQ-033
    When the REQ-033 scope and business rules about persistence are read
    Then shape animation runs must update light state via REQ-020 equivalent operations with REQ-011 semantics
    And canonical stored model coordinates must not be rewritten

  Scenario: REQ-033 authoring remains responsive without hover-only essential steps
    Parent requirement: REQ-033
    Given docs/requirements.md defines REQ-002 and REQ-033
    When the REQ-033 responsive UX notes are read
    Then parameter forms and unified viewport controls must be usable on mobile tablet and desktop without hover-only essential steps

Feature: Faint scene boundary cuboid in three.js views (REQ-034)

  Scenario: REQ-034 defines axis-aligned boundary from light extremes plus symmetric margin
    Parent requirement: REQ-034
    Given docs/requirements.md defines REQ-034 REQ-010 and REQ-015
    When the REQ-034 business rules about geometry are read
    Then the tight boundary must be the axis-aligned min and max of every light position in the viewport coordinate space
    And models need not be regular cuboids because only light positions define the tight box
    And the model view must expand the tight box by exactly 0.3 meters on each axis in both directions
    And the scene view must expand the tight box by the scene persisted margin m on each axis in both directions

  Scenario: REQ-015 stores scene boundary margin m defaulting to 30 cm and allows edit
    Parent requirement: REQ-034
    Given docs/requirements.md defines REQ-015 and REQ-034
    When REQ-015 business rules 12 and 13 are read
    Then each scene must persist one non-negative finite margin m in SI meters for the visualization boundary
    And new scenes must default m to 0.3 meters
    And legacy scenes without m must behave as m equals 0.3 after migration or read fallback
    And the scene management UI must expose a control to view and change m without hover-only essential apply steps

  Scenario: REQ-034 applies to model view and scene view including embedded scene canvases
    Parent requirement: REQ-034
    Given docs/requirements.md defines REQ-034 REQ-010 REQ-015 REQ-027 and REQ-033
    When the REQ-034 scope is read
    Then the model three.js view must show the boundary per REQ-034 in model local coordinates
    And the scene three.js view must show the boundary in derived scene space sx sy sz using that scene current m
    And embedded scene previews that reuse the same canvas pattern must show the boundary using the same scene m

  Scenario: REQ-034 visual prominence matches faint inter-light wire guidance
    Parent requirement: REQ-034
    Given docs/requirements.md defines REQ-010 REQ-019 and REQ-034
    When the REQ-034 business rules about appearance are read
    Then the boundary must be faint and subtle similar in visual weight to REQ-010 inter-light segments
    And it must not be more prominent than those segments or the light spheres

  Scenario: REQ-010 and REQ-015 reference the boundary cuboid requirement
    Parent requirement: REQ-034
    Given docs/requirements.md defines REQ-010 REQ-015 and REQ-034
    When REQ-010 scope and REQ-015 business rule 8 are read
    Then the model view must require the REQ-034 boundary alongside spheres and chain segments
    And the scene view must require the REQ-034 boundary alongside per-model chain segments

Feature: Physical devices WLED first and extensibility (REQ-035)

  Scenario: REQ-035 defines device and mandates WLED as first type
    Parent requirement: REQ-035
    Given docs/requirements.md defines REQ-035
    When requirement REQ-035 is read
    Then a device must be a physical controller that maps model light indices to hardware outputs
    And the first supported device type must be WLED suitable for ESP32-class individually addressable strings
    And architecture must document mapping from REQ-005 light id order to WLED segments or LEDs

  Scenario: REQ-035 excludes durable storage of per-light output state
    Parent requirement: REQ-035
    Given docs/requirements.md defines REQ-035 and REQ-039
    When the REQ-035 business rules about persistence are read
    Then device registry metadata and connection parameters may be persisted
    And per-light on off colour and brightness must not be persisted as application durable state

  Scenario: REQ-035 mandates manual device registration for MVP
    Parent requirement: REQ-035
    Given docs/requirements.md defines REQ-035 and REQ-037
    When the REQ-035 business rules are read
    Then implementations must support adding a device using operator-supplied connection parameters
    And automated network discovery may be deferred until architecture implements it without blocking the manual add path

Feature: Device model one-to-one assignment (REQ-036)

  Scenario: REQ-036 requires exclusive at-most-one assignment each side
    Parent requirement: REQ-036
    Given docs/requirements.md defines REQ-036
    When the REQ-036 business rules are read
    Then a device must be assigned to at most one model at a time
    And a model must have at most one assigned device at a time
    And unassigned devices and unassigned models must both be valid states

  Scenario: REQ-036 forbids silent reassignment when a link already exists
    Parent requirement: REQ-036
    Given docs/requirements.md defines REQ-036
    When the REQ-036 business rule about conflicting assignment is read
    Then assigning when the device or model is already linked must be rejected with a clear error or require an explicit user flow
    And silent reassignment that breaks an existing third link without user intent must not occur

  Scenario: REQ-036 requires assign and unassign via Devices management
    Parent requirement: REQ-036
    Given docs/requirements.md defines REQ-036 and REQ-037
    When the REQ-036 business rules are read
    Then assign and unassign must be available through the Devices management flows described in REQ-037

Feature: Device management UI (REQ-037)

  Scenario: REQ-037 requires a Devices section and core flows
    Parent requirement: REQ-037
    Given docs/requirements.md defines REQ-002 REQ-018 REQ-035 REQ-037
    When the REQ-037 scope and business rules are read
    Then the UI must expose Devices as a first-class navigation destination
    And the user must be able to list devices add a device using manual connection parameters per REQ-035 assign and unassign a model per REQ-036
    And when architecture provides discovery the user must be able to initiate or refresh it and add a chosen candidate
    And the user must be able to view and edit at least device name

  Scenario: REQ-037 essential actions meet responsive baseline
    Parent requirement: REQ-037
    Given docs/requirements.md defines REQ-002 and REQ-037
    When the REQ-037 responsive UX notes are read
    Then essential device actions must not rely on hover-only steps on mobile tablet or desktop

  Scenario: REQ-037 removing a device clears its model assignment
    Parent requirement: REQ-037
    Given docs/requirements.md defines REQ-037 REQ-035 and REQ-036
    When the REQ-037 business rules about deleting or forgetting a device are read
    Then removing a device from the registry must clear any model assignment for that device in the same logical operation

Feature: Routines sync physical lights and run on server (REQ-038)

  Scenario: REQ-038 requires backend routine execution independent of browser
    Parent requirement: REQ-038
    Given docs/requirements.md defines REQ-021 REQ-038 and REQ-039
    When the REQ-038 business rules about server-side execution are read
    Then starting a routine must run automation on the backend until stop or failure
    And a web browser session must not be required for the run to progress

  Scenario: REQ-038 requires physical sync for models with assigned devices
    Parent requirement: REQ-038
    Given docs/requirements.md defines REQ-015 REQ-021 REQ-035 REQ-036 REQ-038 and REQ-039
    When the REQ-038 business rules about physical sync are read
    Then when a routine run updates logical light state for a model that has an assigned device those updates must be applied to the device per architecture
    And models without assigned devices require no physical traffic for those updates

  Scenario: REQ-038 maps device to model local indices not scene placement
    Parent requirement: REQ-038
    Given docs/requirements.md defines REQ-015 and REQ-038
    When the REQ-038 scope about mapping is read
    Then device output must use model local light indices zero through n minus one
    And scene placement offsets must not change how model lights map to hardware

  Scenario: REQ-038 headless control uses the same documented API as the UI
    Parent requirement: REQ-038
    Given docs/requirements.md defines REQ-038
    When the REQ-038 business rules are read
    Then routine start and stop must be available through the same documented HTTP API surface the web UI uses
    And architecture may add an explicit documented integrator alias only if semantics match the UI-facing surface

Feature: In-memory light authority and sync (REQ-039)

  Scenario: REQ-039 forbids durable per-light state in application storage
    Parent requirement: REQ-039
    Given docs/requirements.md defines REQ-011 and REQ-039
    When the REQ-039 business rules are read
    Then per-light on off colour and brightness must not be stored in SQLite or equivalent application store for reload after restart
    And REQ-011 read APIs must reflect current authoritative in-memory state on the running server

  Scenario: REQ-039 defines startup and assignment synchronization
    Parent requirement: REQ-039
    Given docs/requirements.md defines REQ-014 REQ-035 REQ-036 and REQ-039
    When the REQ-039 business rules about sync events are read
    Then on service start models without an assigned device must initialize to all lights off per REQ-014 defaults
    And models with an assigned device must run a documented sync with that device at startup
    And newly assigning a device to a model must trigger sync immediately after the link succeeds

  Scenario: REQ-039 states product priority headless over device UI polish
    Parent requirement: REQ-039
    Given docs/requirements.md defines REQ-037 REQ-038 and REQ-039
    When the REQ-039 scope and priority business rules are read
    Then routine and device correctness and throughput are primary over device UI depth and authoring polish
    And baseline REQ-002 usability where UI exists must still hold

  Scenario: REQ-039 requires architecture to document startup and unassign policies
    Parent requirement: REQ-039
    Given docs/requirements.md defines REQ-039
    When the REQ-039 business rules are read
    Then docs/architecture.md must document one consistent policy when startup sync observes device state that differs from REQ-014 defaults
    And docs/architecture.md must document physical output behavior when a device is unassigned from a model
```

---

**Next step:** After you approve these documents, invoke the `@architect` agent to update `docs/architecture.md` so implementation can proceed. When the feature is implemented, invoke the `@verifier` agent to audit, run tests, and update `docs/traceability_matrix.md`.
