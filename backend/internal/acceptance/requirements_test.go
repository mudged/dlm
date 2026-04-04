package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests trace to Gherkin in docs/acceptance_criteria.md (Given docs/requirements.md exists).

func requirementsPath(t *testing.T) string {
	t.Helper()
	// internal/acceptance -> repo root is ../.. from here; docs at ../../docs
	root := filepath.Join("..", "..", "..", "docs", "requirements.md")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func readRequirements(t *testing.T) string {
	t.Helper()
	p := requirementsPath(t)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

// requirementBlock returns the markdown subsection starting at "### REQ-xxx" up to (but not including) the next "### REQ-" heading.
func requirementBlock(doc, reqID string) string {
	prefix := "### " + reqID + " —"
	i := strings.Index(doc, prefix)
	if i < 0 {
		return ""
	}
	rest := doc[i:]
	next := strings.Index(rest[len(prefix):], "\n### REQ-")
	if next < 0 {
		return rest
	}
	return rest[: len(prefix)+next]
}

func TestAcceptance_REQ001_fullStackComposition(t *testing.T) {
	doc := readRequirements(t)
	lower := strings.ToLower(doc)
	if !strings.Contains(lower, "golang") && !strings.Contains(lower, "go ") {
		t.Fatal("requirements must mention the Go/Golang service (REQ-001)")
	}
	if !strings.Contains(lower, "next.js") {
		t.Fatal("requirements must mention Next.js (REQ-001)")
	}
	if !strings.Contains(lower, "tailwind") {
		t.Fatal("requirements must mention Tailwind (REQ-001)")
	}
	if !strings.Contains(doc, "| **Priority** | Must |") {
		t.Fatal("REQ-001 metadata must state priority Must")
	}
	if !strings.Contains(lower, "end user") {
		t.Fatal("REQ-001 actors must include end user")
	}
	if !strings.Contains(lower, "operator") && !strings.Contains(lower, "maintainer") {
		t.Fatal("REQ-001 actors must include operator or maintainer")
	}
}

func TestAcceptance_REQ002_responsiveAndReactive(t *testing.T) {
	doc := readRequirements(t)
	lower := strings.ToLower(doc)
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-002 must mention %q in responsive notes", kw)
		}
	}
	if !strings.Contains(lower, "reactive") {
		t.Fatal("REQ-002 must require reactive (client-side) UI")
	}
	if !strings.Contains(lower, "next.") && !strings.Contains(lower, "nextjs") {
		t.Fatal("REQ-002 must tie interactivity to Next.js")
	}
}

func TestAcceptance_REQ003_raspberryPiDeployment(t *testing.T) {
	doc := readRequirements(t)
	if !strings.Contains(doc, "Raspberry Pi 4") {
		t.Fatal("REQ-003 must name Raspberry Pi 4 Model B (or equivalent)")
	}
	lower := strings.ToLower(doc)
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-003 must require docs/architecture.md for deployment fit")
	}
	if !strings.Contains(lower, "arm64") {
		t.Fatal("REQ-003 must acknowledge ARM64")
	}
	if !strings.Contains(lower, "cpu") || !strings.Contains(lower, "ram") {
		t.Fatal("REQ-003 must acknowledge CPU and RAM limits")
	}
}

func TestAcceptance_REQ004_singleBinaryNoDocker(t *testing.T) {
	doc := readRequirements(t)
	lower := strings.ToLower(doc)
	if !strings.Contains(lower, "executable") {
		t.Fatal("REQ-004 must mandate an executable deliverable")
	}
	if !strings.Contains(lower, "node") {
		t.Fatal("REQ-004 must address Node.js runtime relative to the distribution")
	}
	if !strings.Contains(lower, "docker") {
		t.Fatal("REQ-004 must explicitly defer Docker / container packaging")
	}
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-004 must require architecture reconciliation (single-binary model)")
	}
}

func TestAcceptance_REQ005_wireLightModel(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-005")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-005 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "1000") {
		t.Fatal("REQ-005 must cap lights at 1000")
	}
	if !strings.Contains(lower, "csv") {
		t.Fatal("REQ-005 must mention CSV interchange")
	}
	for _, col := range []string{"id", "x", "y", "z"} {
		needle := "**" + col + "**"
		if !strings.Contains(block, needle) {
			t.Fatalf("REQ-005 must mention column/field %s in requirements text", needle)
		}
	}
	if !strings.Contains(lower, "metadata") {
		t.Fatal("REQ-005 must mention model metadata")
	}
	if !strings.Contains(lower, "name") {
		t.Fatal("REQ-005 must mention model name in metadata")
	}
	if !strings.Contains(lower, "creation date") {
		t.Fatal("REQ-005 must mention creation date in metadata")
	}
	if !strings.Contains(lower, "sequential") && !strings.Contains(lower, "contiguous") {
		t.Fatal("REQ-005 must require sequential/contiguous light indices")
	}
	if !strings.Contains(block, "starting at **0**") {
		t.Fatal("REQ-005 must state indices start at 0")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-005 metadata must state priority Must")
	}
}

func TestAcceptance_REQ006_modelCRUDAndCSVUpload(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-006")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-006 section")
	}
	lower := strings.ToLower(block)
	for _, verb := range []string{"list", "view", "delete"} {
		if !strings.Contains(lower, verb) {
			t.Fatalf("REQ-006 must mention %q", verb)
		}
	}
	if !strings.Contains(lower, "upload") {
		t.Fatal("REQ-006 must mention uploading a CSV")
	}
	if !strings.Contains(lower, "csv") {
		t.Fatal("REQ-006 must mention CSV for create")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-006 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "req-005") || !strings.Contains(lower, "req-001") {
		t.Fatal("REQ-006 dependencies must reference prior requirements (e.g. REQ-001, REQ-005)")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-006 metadata must state priority Must")
	}
}

func TestAcceptance_REQ007_validateCSVOnUpload(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-007")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-007 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "validat") {
		t.Fatal("REQ-007 must require validation on upload")
	}
	if !strings.Contains(lower, "csv") {
		t.Fatal("REQ-007 must mention CSV")
	}
	if !strings.Contains(lower, "sequential") || !strings.Contains(lower, "id") {
		t.Fatal("REQ-007 must mention sequential id rules")
	}
	if !strings.Contains(lower, "1000") {
		t.Fatal("REQ-007 must reference the 1000-light maximum")
	}
	if !strings.Contains(lower, "finite") {
		t.Fatal("REQ-007 must require finite numeric coordinates")
	}
	if !strings.Contains(lower, "actionable") || !strings.Contains(lower, "feedback") {
		t.Fatal("REQ-007 must require actionable feedback on rejection")
	}
	if !strings.Contains(lower, "partial") {
		t.Fatal("REQ-007 must forbid persisting a partial model on failed upload")
	}
	if !strings.Contains(lower, "req-005") && !strings.Contains(lower, "req-006") {
		t.Fatal("REQ-007 dependencies must reference REQ-005 or REQ-006")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-007 metadata must state priority Must")
	}
}

func readmePath(t *testing.T) string {
	t.Helper()
	root := filepath.Join("..", "..", "..", "README.md")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func readREADME(t *testing.T) string {
	t.Helper()
	p := readmePath(t)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func TestAcceptance_REQ008_singleCommandBuildAndRun(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-008")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-008 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "readme.md") {
		t.Fatal("REQ-008 must require README.md to document the command")
	}
	if !strings.Contains(lower, "go") {
		t.Fatal("REQ-008 must mention launching the Go server")
	}
	if !strings.Contains(lower, "static") && !strings.Contains(lower, "export") {
		t.Fatal("REQ-008 must mention static export or embedding the UI build")
	}
	if !strings.Contains(lower, "agents.md") {
		t.Fatal("REQ-008 must reference AGENTS.md for REQ-008 awareness")
	}
	if !strings.Contains(lower, "second manual step") {
		t.Fatal("REQ-008 must address the single-invocation workflow (no second manual step)")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-008 metadata must state priority Must")
	}
	readme := readREADME(t)
	rl := strings.ToLower(readme)
	if !strings.Contains(rl, "scripts/run.sh") && !strings.Contains(rl, "script") {
		t.Fatal("README should document the run script path (e.g. scripts/run.sh) per REQ-008")
	}
	if !strings.Contains(rl, "go") {
		t.Fatal("README should mention the Go server as part of the documented workflow")
	}
}

func TestAcceptance_REQ009_defaultSampleModels(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-009")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-009 section")
	}
	lower := strings.ToLower(block)
	for _, shape := range []string{"sphere", "cube", "cone"} {
		if !strings.Contains(lower, shape) {
			t.Fatalf("REQ-009 must mention sample shape %q", shape)
		}
	}
	if !strings.Contains(block, "500") || !strings.Contains(block, "1000") {
		t.Fatal("REQ-009 must state the 500–1000 light count band for samples")
	}
	if !strings.Contains(block, "0.05") || !strings.Contains(block, "0.10") {
		t.Fatal("REQ-009 must state consecutive spacing 0.05 m and 0.10 m (5–10 cm)")
	}
	if !strings.Contains(lower, "surface") {
		t.Fatal("REQ-009 must require lights on the outside surface")
	}
	if !strings.Contains(block, "0.03") {
		t.Fatal("REQ-009 must state the 0.03 m surface deviation tolerance")
	}
	if !strings.Contains(lower, "inside") {
		t.Fatal("REQ-009 must forbid interior placement (not inside the solid)")
	}
	if !strings.Contains(lower, "req-005") && !strings.Contains(lower, "req-006") {
		t.Fatal("REQ-009 dependencies must reference REQ-005 or REQ-006")
	}
	if !strings.Contains(lower, "evenly") {
		t.Fatal("REQ-009 must require even distribution over faces/surfaces")
	}
	if !strings.Contains(lower, "plane") {
		t.Fatal("REQ-009 must mention face planes (cube) / base plane (cone)")
	}
	if !strings.Contains(lower, "edge") {
		t.Fatal("REQ-009 must contrast face coverage with edge-only / wireframe layouts")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-009 metadata must state priority Must")
	}
}

func TestAcceptance_REQ010_threeJsModelView(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-010")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-010 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "three.js") && !strings.Contains(lower, "three") {
		t.Fatal("REQ-010 must require three.js")
	}
	if !strings.Contains(block, "0.02") {
		t.Fatal("REQ-010 must state 0.02 m (2 cm) sphere diameter")
	}
	if !strings.Contains(lower, "white") {
		t.Fatal("REQ-010 must require white spheres")
	}
	if !strings.Contains(lower, "every") && !strings.Contains(lower, "all") {
		t.Fatal("REQ-010 must require every/all lights to be drawn")
	}
	if !strings.Contains(strings.ToLower(block), "d0d0d0") {
		t.Fatal("REQ-010 must specify #D0D0D0 for wire segment grey")
	}
	if !strings.Contains(lower, "85%") && !strings.Contains(lower, "85 %") {
		t.Fatal("REQ-010 must state 85% transparency for segments")
	}
	if !strings.Contains(lower, "thin") {
		t.Fatal("REQ-010 must require thin segments")
	}
	if !strings.Contains(lower, "hover") {
		t.Fatal("REQ-010 must mention pointer hover for id and coordinates")
	}
	if !strings.Contains(lower, "touch") || !strings.Contains(lower, "tablet") {
		t.Fatal("REQ-010 must address touch or tablet for coordinate discovery")
	}
	if !strings.Contains(lower, "direct") {
		t.Fatal("REQ-010 must require three.js as a direct front-end dependency")
	}
	if !strings.Contains(lower, "manifest") && !strings.Contains(lower, "package") {
		t.Fatal("REQ-010 must mention the Next.js package manifest for the three dependency")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-010 dependencies or scope must reference REQ-002 for responsive UX")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-010 metadata must state priority Must")
	}
}

func TestAcceptance_REQ011_restLightStateAPI(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-011")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-011 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "rest") {
		t.Fatal("REQ-011 must require a REST API")
	}
	if !strings.Contains(lower, "query") && !strings.Contains(lower, "read") {
		t.Fatal("REQ-011 must allow querying or reading current light state")
	}
	if !strings.Contains(lower, "updat") && !strings.Contains(lower, "write") {
		t.Fatal("REQ-011 must allow updating or writing light state")
	}
	if !strings.Contains(lower, "individual") && !strings.Contains(lower, "each") {
		t.Fatal("REQ-011 must require per-light (individual or each) control")
	}
	if !strings.Contains(lower, "hex") {
		t.Fatal("REQ-011 must mention hex colour")
	}
	if !strings.Contains(lower, "brightness") || !strings.Contains(lower, "percent") {
		t.Fatal("REQ-011 must mention brightness as a percentage")
	}
	if !strings.Contains(lower, "**on**") || !strings.Contains(lower, "**off**") {
		t.Fatal("REQ-011 must state on and off light state (e.g. **on** / **off** in scope)")
	}
	if !strings.Contains(lower, "boolean") {
		t.Fatal("REQ-011 must describe on/off as boolean where stated in scope")
	}
	if !strings.Contains(lower, "patch") {
		t.Fatal("REQ-011 must mention PATCH-style partial updates")
	}
	if !strings.Contains(lower, "persist") {
		t.Fatal("REQ-011 must require successful writes to be persisted")
	}
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-011 must defer URL layout or defaults to docs/architecture.md")
	}
	if !strings.Contains(lower, "req-001") || !strings.Contains(lower, "req-005") || !strings.Contains(lower, "req-006") {
		t.Fatal("REQ-011 dependencies must reference REQ-001, REQ-005, and REQ-006")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-011 metadata must state priority Must")
	}
}

func TestAcceptance_REQ012_visualizationReflectsLightState(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-012")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-012 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "visual") && !strings.Contains(lower, "3d") {
		t.Fatal("REQ-012 must address visualization or 3D view")
	}
	if !strings.Contains(lower, "req-011") {
		t.Fatal("REQ-012 must reference REQ-011 for stored state")
	}
	if !strings.Contains(lower, "filled") {
		t.Fatal("REQ-012 must require a filled appearance for on lights")
	}
	if !strings.Contains(strings.ToLower(block), "d0d0d0") {
		t.Fatal("REQ-012 must specify #D0D0D0 for off-light appearance (aligned with segments)")
	}
	if !strings.Contains(lower, "85%") && !strings.Contains(lower, "85 %") {
		t.Fatal("REQ-012 must state 85% transparency for off lights")
	}
	if !strings.Contains(lower, "brightness") && !strings.Contains(lower, "colour") && !strings.Contains(lower, "color") {
		t.Fatal("REQ-012 must tie appearance to colour and/or brightness")
	}
	if !strings.Contains(lower, "reload") {
		t.Fatal("REQ-012 must address refresh without full page reload")
	}
	if !strings.Contains(lower, "req-010") {
		t.Fatal("REQ-012 must reference REQ-010 (segments / hover baseline)")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-012 responsive notes must reference REQ-002")
	}
	if !strings.Contains(block, "0.02") {
		t.Fatal("REQ-012 must preserve 0.02 m sphere diameter (REQ-010 alignment)")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-012 metadata must state priority Must")
	}
}

func TestAcceptance_REQ013_modelViewPaginationAndBulkSettings(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-013")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-013 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "paginat") {
		t.Fatal("REQ-013 must require a paginated light list")
	}
	if !strings.Contains(lower, "multi-select") || !strings.Contains(lower, "select multiple") {
		t.Fatal("REQ-013 must require multi-select or selecting multiple lights")
	}
	if !strings.Contains(lower, "page size") || !strings.Contains(lower, "three") {
		t.Fatal("REQ-013 must require at least three page-size choices")
	}
	if !strings.Contains(lower, "1000") {
		t.Fatal("REQ-013 must bound page size within 1 and 1000")
	}
	if !strings.Contains(lower, "jump") || !strings.Contains(lower, "id") {
		t.Fatal("REQ-013 must require jumping to a light by id")
	}
	if !strings.Contains(lower, "feedback") {
		t.Fatal("REQ-013 must require feedback for invalid or out-of-range ids")
	}
	if !strings.Contains(lower, "req-005") {
		t.Fatal("REQ-013 must reference REQ-005 for light id semantics")
	}
	if !strings.Contains(lower, "bulk") || !strings.Contains(lower, "apply") {
		t.Fatal("REQ-013 must require bulk or apply of settings to selected lights")
	}
	if !strings.Contains(lower, "req-011") {
		t.Fatal("REQ-013 must reference REQ-011 for on/off, colour, and brightness")
	}
	if !strings.Contains(lower, "persist") {
		t.Fatal("REQ-013 must require persistence per REQ-011 for successful updates")
	}
	if !strings.Contains(lower, "req-012") {
		t.Fatal("REQ-013 must reference REQ-012 for 3D view and list timeliness after bulk apply")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-013 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-013 must reference REQ-002 for responsive / touch / keyboard patterns")
	}
	if !strings.Contains(lower, "checkbox") {
		t.Fatal("REQ-013 must mention checkboxes or an equivalent multi-select affordance")
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-013 must forbid hover-only affordances for essential actions")
	}
	if !strings.Contains(lower, "req-006") {
		t.Fatal("REQ-013 dependencies or scope must reference REQ-006 (model view)")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-013 metadata must state priority Must")
	}
}

func TestAcceptance_REQ016_cameraResetModelAndSceneViews(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-016")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-016 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "camera reset") {
		t.Fatal("REQ-016 must require a camera reset control")
	}
	if !strings.Contains(lower, "model") {
		t.Fatal("REQ-016 must address the model 3D view")
	}
	if !strings.Contains(lower, "scene") {
		t.Fatal("REQ-016 must address the scene 3D view")
	}
	if !strings.Contains(lower, "three.js") || !strings.Contains(lower, "req-010") {
		t.Fatal("REQ-016 must tie model view to three.js / REQ-010")
	}
	if !strings.Contains(lower, "req-015") {
		t.Fatal("REQ-016 must reference REQ-015 for scene visualization")
	}
	if !strings.Contains(lower, "default") {
		t.Fatal("REQ-016 must require restoring default camera / framing")
	}
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-016 must reference docs/architecture.md for default framing")
	}
	if !strings.Contains(lower, "client-side") || !strings.Contains(lower, "navigation") {
		t.Fatal("REQ-016 must limit reset to client-side navigation")
	}
	if !strings.Contains(lower, "persist") {
		t.Fatal("REQ-016 must state that persisted models/scenes/state are not altered")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-016 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-016 must reference REQ-002 for non-hover-only operability")
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-016 must address hover-only affordances for the reset control")
	}
	if !strings.Contains(lower, "req-014") {
		t.Fatal("REQ-016 out-of-scope or rules must distinguish light reset (REQ-014) from camera reset")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-016 metadata must state priority Must")
	}
}

func TestAcceptance_REQ017_optionsFactoryResetWithConfirmation(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-017")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-017 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "options") {
		t.Fatal("REQ-017 must require an Options section or view")
	}
	if !strings.Contains(lower, "factory reset") {
		t.Fatal("REQ-017 must name factory reset as an action")
	}
	if !strings.Contains(lower, "req-009") {
		t.Fatal("REQ-017 must reference REQ-009 for re-seeded default samples")
	}
	if !strings.Contains(lower, "three") && !strings.Contains(lower, "sample") {
		t.Fatal("REQ-017 must require three default sample models after reset")
	}
	if !strings.Contains(lower, "confirm") {
		t.Fatal("REQ-017 must require explicit confirmation before destructive work")
	}
	if !strings.Contains(lower, "dialog") || !strings.Contains(lower, "prompt") {
		t.Fatal("REQ-017 must require a blocking prompt or dialog before irreversible effects")
	}
	if !strings.Contains(lower, "cancel") {
		t.Fatal("REQ-017 must allow cancel/dismissal without data loss")
	}
	if !strings.Contains(lower, "permanent") || !strings.Contains(lower, "irrevers") {
		t.Fatal("REQ-017 must warn of permanent or irreversible data loss")
	}
	if !strings.Contains(lower, "scene") {
		t.Fatal("REQ-017 must address scenes in wipe scope")
	}
	if !strings.Contains(lower, "req-015") {
		t.Fatal("REQ-017 must reference REQ-015 for scenes")
	}
	if !strings.Contains(lower, "req-011") {
		t.Fatal("REQ-017 must reference REQ-011 for per-light state")
	}
	if !strings.Contains(lower, "req-014") {
		t.Fatal("REQ-017 must reference REQ-014 for defaults after re-seed")
	}
	if !strings.Contains(lower, "req-006") {
		t.Fatal("REQ-017 must reference REQ-006")
	}
	if !strings.Contains(lower, "architecture") {
		t.Fatal("REQ-017 must defer store shape or navigation to architecture")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-017 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-017 must reference REQ-002 for the full flow (no hover-only)")
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-017 must forbid hover-only requirements for Options / factory reset flow")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-017 metadata must state priority Must")
	}
}

func TestAcceptance_REQ018_applicationShellThemesNavAndFontAwesome(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-018")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-018 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "light") || !strings.Contains(lower, "dark") {
		t.Fatal("REQ-018 must require both light and dark themes")
	}
	if !strings.Contains(lower, "theme") {
		t.Fatal("REQ-018 must use the term theme(s) for UI modes")
	}
	if !strings.Contains(lower, "white") {
		t.Fatal("REQ-018 must require white or white-equivalent background for light mode")
	}
	if !strings.Contains(lower, "grey") && !strings.Contains(lower, "gray") {
		t.Fatal("REQ-018 must require dark grey (or gray) background for dark mode")
	}
	if !strings.Contains(lower, "burger") && !strings.Contains(lower, "hamburger") {
		t.Fatal("REQ-018 must require a burger or hamburger control for the menu")
	}
	if !strings.Contains(lower, "left") {
		t.Fatal("REQ-018 must place primary navigation on the left")
	}
	if !strings.Contains(lower, "collaps") {
		t.Fatal("REQ-018 must require a collapsible navigation region")
	}
	if !strings.Contains(lower, "domestic light & magic") {
		t.Fatal("REQ-018 must require the exact application title Domestic Light & Magic")
	}
	if !strings.Contains(lower, "font awesome") {
		t.Fatal("REQ-018 must name Font Awesome for logo and buttons")
	}
	if !strings.Contains(lower, "lightbulb") {
		t.Fatal("REQ-018 must specify the lightbulb icon for the logo")
	}
	if !strings.Contains(lower, "regular") {
		t.Fatal("REQ-018 must specify classic/regular style for the lightbulb logo")
	}
	if !strings.Contains(lower, "fontawesome.com") {
		t.Fatal("REQ-018 must reference the Font Awesome catalog URL for the lightbulb icon")
	}
	if !strings.Contains(lower, "button") {
		t.Fatal("REQ-018 must require Font Awesome icons on buttons")
	}
	if !strings.Contains(lower, "req-001") {
		t.Fatal("REQ-018 must list REQ-001 in dependencies")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-018 must reference REQ-002 for responsive and non-hover-only use")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-018 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-018 must forbid hover-only requirements for menu or theme switching")
	}
	if !strings.Contains(lower, "touch") {
		t.Fatal("REQ-018 must require touch usability for the burger/menu (REQ-002 tie)")
	}
	if !strings.Contains(lower, "architecture") {
		t.Fatal("REQ-018 must defer some placement or tokens to architecture")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-018 metadata must state priority Must")
	}
}
