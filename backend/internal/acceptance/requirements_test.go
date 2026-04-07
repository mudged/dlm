package acceptance

import (
	"os"
	"path/filepath"
	"regexp"
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
	return rest[:len(prefix)+next]
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

func TestAcceptance_REQ014_defaultLightStateAndReset(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-014")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-014 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "off") {
		t.Fatal("REQ-014 must require lights to start off")
	}
	if !strings.Contains(lower, "100") {
		t.Fatal("REQ-014 must require 100 percent default brightness")
	}
	if !strings.Contains(block, "#FFFFFF") && !strings.Contains(block, "#ffffff") {
		t.Fatal("REQ-014 must require default hex colour #FFFFFF")
	}
	if !strings.Contains(lower, "reset") {
		t.Fatal("REQ-014 must require a reset control")
	}
	if !strings.Contains(lower, "all") {
		t.Fatal("REQ-014 must require reset to apply to all lights")
	}
	if !strings.Contains(lower, "persist") {
		t.Fatal("REQ-014 must require reset persistence")
	}
	if !strings.Contains(lower, "req-011") || !strings.Contains(lower, "req-012") || !strings.Contains(lower, "req-013") {
		t.Fatal("REQ-014 must reference REQ-011, REQ-012, and REQ-013")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-014 must reference REQ-002 for non-hover-only operability")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-014 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-014 must forbid hover-only reset interaction")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-014 metadata must state priority Must")
	}
}

func TestAcceptance_REQ015_scenesCompositeSpaceAndManagement(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-015")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-015 section")
	}
	lower := strings.ToLower(block)
	for _, kw := range []string{"scene", "create", "list", "delete", "open"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-015 must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "automatically") || !strings.Contains(lower, "offset") {
		t.Fatal("REQ-015 must require automatic placement offsets on create")
	}
	if !strings.Contains(lower, "integer") {
		t.Fatal("REQ-015 must require integer placement offsets")
	}
	if !strings.Contains(lower, "non-negative") {
		t.Fatal("REQ-015 must require non-negative scene-space containment")
	}
	if !strings.Contains(lower, "1 meter") && !strings.Contains(lower, "1 m") {
		t.Fatal("REQ-015 must require at least 1 meter margin")
	}
	if !strings.Contains(lower, "three.js") {
		t.Fatal("REQ-015 must require three.js scene rendering")
	}
	if !strings.Contains(lower, "derived") || !strings.Contains(lower, "canonical") {
		t.Fatal("REQ-015 must distinguish derived scene coordinates from canonical model coordinates")
	}
	if !strings.Contains(lower, "confirm") || !strings.Contains(lower, "last") {
		t.Fatal("REQ-015 must require confirmation when removing the last model")
	}
	if !strings.Contains(lower, "to the") || !strings.Contains(lower, "right") {
		t.Fatal("REQ-015 must require default placement to the right when adding models")
	}
	for _, dep := range []string{"req-002", "req-005", "req-006", "req-010", "req-011", "req-012"} {
		if !strings.Contains(lower, dep) {
			t.Fatalf("REQ-015 dependencies must include %s", strings.ToUpper(dep))
		}
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-015 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-015 metadata must state priority Must")
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

func TestAcceptance_REQ019_fixedDarkGreyThreeJSViewport(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-019")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-019 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "three.js") {
		t.Fatal("REQ-019 must reference three.js model and scene views")
	}
	if !strings.Contains(lower, "dark") || (!strings.Contains(lower, "grey") && !strings.Contains(lower, "gray")) {
		t.Fatal("REQ-019 must require a dark-grey viewport background")
	}
	if !strings.Contains(lower, "light") || !strings.Contains(lower, "dark") {
		t.Fatal("REQ-019 must require same viewport policy in light and dark shell modes")
	}
	if !strings.Contains(lower, "req-018") {
		t.Fatal("REQ-019 must reference REQ-018 shell theme behavior")
	}
	if !strings.Contains(lower, "req-010") || !strings.Contains(lower, "req-015") {
		t.Fatal("REQ-019 dependencies must include REQ-010 and REQ-015")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-019 must reference REQ-002 for responsive controls")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-019 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-019 must forbid hover-only essential 3D interactions")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-019 metadata must state priority Must")
	}
}

func TestAcceptance_REQ020_sceneSpatialAPI(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-020")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-020 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "dimension") {
		t.Fatal("REQ-020 must require a scene dimensions read operation")
	}
	if !strings.Contains(lower, "all lights") {
		t.Fatal("REQ-020 must require retrieving all scene lights")
	}
	if !strings.Contains(lower, "cuboid") {
		t.Fatal("REQ-020 must require cuboid query/update operations")
	}
	if !strings.Contains(lower, "sphere") {
		t.Fatal("REQ-020 must require sphere query/update operations")
	}
	if !strings.Contains(lower, "bulk") || !strings.Contains(lower, "update") {
		t.Fatal("REQ-020 must require bulk update operations")
	}
	if !strings.Contains(lower, "scene space") {
		t.Fatal("REQ-020 must require scene-space coordinates")
	}
	if !strings.Contains(lower, "must not") || !strings.Contains(lower, "canonical") {
		t.Fatal("REQ-020 must require canonical model coordinates to remain unchanged")
	}
	if !strings.Contains(lower, "invalid geometry") || !strings.Contains(lower, "non-positive") {
		t.Fatal("REQ-020 must require invalid geometry rejection")
	}
	if !strings.Contains(lower, "partial") {
		t.Fatal("REQ-020 must forbid partial updates on invalid geometry")
	}
	for _, dep := range []string{"req-005", "req-011", "req-015"} {
		if !strings.Contains(lower, dep) {
			t.Fatalf("REQ-020 dependencies must include %s", strings.ToUpper(dep))
		}
	}
	if !strings.Contains(lower, "n/a") || !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-020 must include API-only responsive notes with REQ-002 linkage")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-020 metadata must state priority Must")
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

func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestAcceptance_REQ022_pythonSceneRoutinesAndCodeMirror6(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-022")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-022 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "codemirror") {
		t.Fatal("REQ-022 must require CodeMirror for the in-browser Python editor")
	}
	if !strings.Contains(lower, "@codemirror") {
		t.Fatal("REQ-022 must reference the @codemirror package ecosystem")
	}
	if !strings.Contains(lower, "browser") {
		t.Fatal("REQ-022 must state browser editing for Python routines")
	}
	if !strings.Contains(lower, "req-021") {
		t.Fatal("REQ-022 must reference REQ-021 for routine coexistence")
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-022 must reference REQ-002 for responsive / non-hover-only editor UX")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-022 metadata must state priority Must")
	}

	root := repoRoot(t)
	pkgPath := filepath.Join(root, "web", "package.json")
	pkgBytes, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("read %s: %v", pkgPath, err)
	}
	pkg := string(pkgBytes)
	pkgLower := strings.ToLower(pkg)
	if strings.Contains(pkgLower, "monaco") {
		t.Fatal("web/package.json must not depend on Monaco for REQ-022 (CodeMirror 6 only)")
	}
	for _, dep := range []string{
		`"@codemirror/lang-python"`,
		`"@codemirror/lint"`,
		`"@codemirror/autocomplete"`,
		`"codemirror"`,
	} {
		if !strings.Contains(pkg, dep) {
			t.Fatalf("web/package.json must include dependency %s for REQ-022 CodeMirror 6 stack", dep)
		}
	}
	if !strings.Contains(pkg, `"@codemirror/lang-python": "^6`) && !strings.Contains(pkg, `"@codemirror/lang-python": "6`) {
		t.Fatal("web/package.json must pin @codemirror/lang-python to major version 6")
	}
	if strings.Contains(pkgLower, "@uiw/react-codemirror") || strings.Contains(pkg, `"@uiw/react-codemirror"`) {
		t.Fatal("web/package.json must not depend on @uiw/react-codemirror when using the codemirror package + EditorView")
	}

	editorPath := filepath.Join(root, "web", "app", "routines", "python", "PythonRoutineEditorClient.tsx")
	ed, err := os.ReadFile(editorPath)
	if err != nil {
		t.Fatalf("read %s: %v", editorPath, err)
	}
	edStr := string(ed)
	if strings.Contains(edStr, "@uiw/react-codemirror") {
		t.Fatal("PythonRoutineEditorClient must not import @uiw/react-codemirror")
	}
	if !strings.Contains(edStr, "PythonCodeMirrorEditor") {
		t.Fatal("PythonRoutineEditorClient must render PythonCodeMirrorEditor (CodeMirror 6 EditorView host)")
	}
	if !strings.Contains(edStr, "@codemirror/lang-python") {
		t.Fatal("PythonRoutineEditorClient must use @codemirror/lang-python")
	}

	cmPath := filepath.Join(root, "web", "components", "PythonCodeMirrorEditor.tsx")
	cmBytes, err := os.ReadFile(cmPath)
	if err != nil {
		t.Fatalf("read %s: %v", cmPath, err)
	}
	cmStr := string(cmBytes)
	if !strings.Contains(cmStr, "EditorView") {
		t.Fatal("PythonCodeMirrorEditor must use CodeMirror 6 EditorView")
	}
	if !strings.Contains(cmStr, `from "codemirror"`) {
		t.Fatal("PythonCodeMirrorEditor must import from the codemirror package (e.g. basicSetup)")
	}
}

func TestAcceptance_REQ024_pythonRoutineApiReferenceBelowEditor(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-024")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-024 section")
	}
	lower := strings.ToLower(block)
	for _, kw := range []string{
		"below", "editor", "codemirror", "insert", "sample", "catalog",
	} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-024 must mention %q (API reference / editor integration)", kw)
		}
	}
	for _, dep := range []string{"req-002", "req-018", "req-022", "req-027"} {
		if !strings.Contains(lower, dep) {
			t.Fatalf("REQ-024 dependencies must reference %s", dep)
		}
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-024 metadata must state priority Must")
	}

	root := repoRoot(t)
	editorPath := filepath.Join(root, "web", "app", "routines", "python", "PythonRoutineEditorClient.tsx")
	ed, err := os.ReadFile(editorPath)
	if err != nil {
		t.Fatalf("read %s: %v", editorPath, err)
	}
	edStr := string(ed)
	idxEditor := strings.Index(edStr, "<PythonCodeMirrorEditor")
	idxCatalog := strings.Index(edStr, "<PythonSceneApiCatalogSection")
	if idxEditor < 0 || idxCatalog < 0 || idxEditor > idxCatalog {
		t.Fatal("REQ-024: PythonCodeMirrorEditor must appear before PythonSceneApiCatalogSection in document order")
	}
	if !strings.Contains(edStr, "editorViewRef={editorViewRef}") {
		t.Fatal("REQ-024: editor must expose EditorView ref for caret-based insert")
	}
	if !strings.Contains(edStr, "insertSnippetInPythonEditor") {
		t.Fatal("REQ-024: routine editor must use insertSnippetInPythonEditor for snippet insertion")
	}
	if !strings.Contains(edStr, "onInsertSnippet={onInsertSnippet}") {
		t.Fatal("REQ-024: catalog section must receive onInsertSnippet")
	}

	catPath := filepath.Join(root, "web", "components", "PythonSceneApiCatalogSection.tsx")
	catBytes, err := os.ReadFile(catPath)
	if err != nil {
		t.Fatalf("read %s: %v", catPath, err)
	}
	catStr := string(catBytes)
	if !strings.Contains(catStr, `id="python-scene-api-catalog"`) {
		t.Fatal("PythonSceneApiCatalogSection must expose stable id python-scene-api-catalog")
	}
	if !strings.Contains(catStr, "<select") {
		t.Fatal("REQ-024: catalog must provide a selectable control (e.g. select) for entries")
	}
	if !strings.Contains(catStr, "Put this example in your code") {
		t.Fatal("REQ-024: catalog must expose an insert affordance for the shown example")
	}

	insertPath := filepath.Join(root, "web", "lib", "insertPythonEditorSnippet.ts")
	insertBytes, err := os.ReadFile(insertPath)
	if err != nil {
		t.Fatalf("read %s: %v", insertPath, err)
	}
	if !strings.Contains(string(insertBytes), "insertSnippetInPythonEditor") {
		t.Fatal("insertPythonEditorSnippet.ts must export insertSnippetInPythonEditor")
	}

	catalogSrcPath := filepath.Join(root, "web", "lib", "pythonSceneApiCatalog.ts")
	catalogSrc, err := os.ReadFile(catalogSrcPath)
	if err != nil {
		t.Fatalf("read %s: %v", catalogSrcPath, err)
	}
	cs := string(catalogSrc)
	snippetRe := regexp.MustCompile("(?s)snippet:\\s+`([^`]+)`")
	matches := snippetRe.FindAllStringSubmatch(cs, -1)
	if len(matches) < 5 {
		t.Fatalf("pythonSceneApiCatalog.ts must define multiple catalog snippets (got %d)", len(matches))
	}
	for i, m := range matches {
		if len(m) < 2 || !strings.Contains(m[1], "#") {
			t.Fatalf("REQ-024: catalog snippet %d must include at least one Python # comment", i)
		}
	}
}

func TestAcceptance_REQ027_pythonRoutineUnifiedRunAndViewport(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-027")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-027 section")
	}
	lower := strings.ToLower(block)
	for _, kw := range []string{"unified", "viewport", "three.js", "scene", "req-012", "req-014", "req-016"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-027 must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "req-022") {
		t.Fatal("REQ-027 must reference REQ-022 for the authoring surface")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-027 metadata must state priority Must")
	}

	root := repoRoot(t)
	editorPath := filepath.Join(root, "web", "app", "routines", "python", "PythonRoutineEditorClient.tsx")
	ed, err := os.ReadFile(editorPath)
	if err != nil {
		t.Fatalf("read %s: %v", editorPath, err)
	}
	edStr := string(ed)
	if strings.Count(edStr, "value={targetSceneId}") != 1 {
		t.Fatal("REQ-027: exactly one scene selector must bind targetSceneId (run + viewport)")
	}
	if strings.Count(edStr, "<SceneLightsCanvas") != 1 {
		t.Fatal("REQ-027: unified region must render a single SceneLightsCanvas for the live view")
	}
	if !strings.Contains(edStr, "startSceneRoutine(targetSceneId") {
		t.Fatal("REQ-027: Start must target the same scene id as the viewport")
	}
	if !strings.Contains(edStr, "Run your script and watch the room") {
		t.Fatal("REQ-027: unified run/viewport section must be clearly labeled for beginners")
	}
	if !strings.Contains(edStr, "Reset room lights") || !strings.Contains(edStr, "Reset camera") {
		t.Fatal("REQ-027: unified region must expose reset scene lights and reset camera controls")
	}
	if strings.Contains(edStr, "debugSceneId") || strings.Contains(edStr, "runSceneId") {
		t.Fatal("REQ-027: must not split run vs debug with separate scene id state")
	}
}

func TestAcceptance_REQ030_pythonSceneRandomHexColour(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-030")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-030 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "random") && !strings.Contains(lower, "randrange") {
		t.Fatal("REQ-030 must describe random hex colour behaviour")
	}
	if !strings.Contains(lower, "req-011") || !strings.Contains(lower, "req-022") || !strings.Contains(lower, "req-024") {
		t.Fatal("REQ-030 must list REQ-011, REQ-022, and REQ-024 in dependencies or body")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-030 metadata must state priority Must")
	}

	root := repoRoot(t)
	workerPath := filepath.Join(root, "web", "public", "dlm-python-scene-worker.mjs")
	wb, err := os.ReadFile(workerPath)
	if err != nil {
		t.Fatalf("read %s: %v", workerPath, err)
	}
	ws := string(wb)
	if !strings.Contains(ws, "random_hex_colour") {
		t.Fatal("dlm-python-scene-worker.mjs must expose random_hex_colour on the scene shim")
	}
	if !strings.Contains(ws, "randrange") {
		t.Fatal("REQ-030 worker must use Python random.randrange for distribution parity")
	}

	catalogPath := filepath.Join(root, "web", "lib", "pythonSceneApiCatalog.ts")
	cb, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("read %s: %v", catalogPath, err)
	}
	cs := string(cb)
	if !strings.Contains(cs, "random_hex_colour") {
		t.Fatal("pythonSceneApiCatalog must list random_hex_colour for REQ-024")
	}
	if !strings.Contains(cs, "scene.random_hex_colour()") {
		t.Fatal("pythonSceneApiCatalog must show scene.random_hex_colour() in template or snippet")
	}

	archPath := filepath.Join(root, "docs", "architecture.md")
	archBytes, err := os.ReadFile(archPath)
	if err != nil {
		t.Fatalf("read %s: %v", archPath, err)
	}
	archStr := string(archBytes)
	if !strings.Contains(archStr, "random_hex_colour") {
		t.Fatal("docs/architecture.md must document scene.random_hex_colour (REQ-030)")
	}
}

func TestAcceptance_REQ029_highThroughputLightUpdates(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-029")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-029 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "integrator") {
		t.Fatal("REQ-029 must name integrators as actors")
	}
	if !strings.Contains(lower, "hundreds") {
		t.Fatal("REQ-029 must state a hundreds-of-lights scale assumption")
	}
	if !strings.Contains(lower, "req-011") {
		t.Fatal("REQ-029 must reference REQ-011 for persisted per-light state semantics")
	}
	if !strings.Contains(lower, "req-012") {
		t.Fatal("REQ-029 must reference REQ-012 for viewer freshness")
	}
	if !strings.Contains(lower, "req-015") || !strings.Contains(lower, "req-020") {
		t.Fatal("REQ-029 scope must reference REQ-015 and REQ-020 for model/scene contexts")
	}
	if !strings.Contains(lower, "raspberry pi") || !strings.Contains(lower, "req-003") {
		t.Fatal("REQ-029 must tie plausible solutions to Pi / REQ-003 constraints")
	}
	if !strings.Contains(lower, "server-sent events") && !strings.Contains(lower, "websocket") {
		t.Fatal("REQ-029 scope must mention Server-Sent Events and/or WebSocket as illustrative mechanisms")
	}
	if !strings.Contains(lower, "batch") || !strings.Contains(lower, "bulk") {
		t.Fatal("REQ-029 must mention batch or bulk write APIs")
	}
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-029 business rule 1 must require docs/architecture.md")
	}
	if !strings.Contains(lower, "write path") {
		t.Fatal("REQ-029 must require the architecture to describe the write path")
	}
	if !strings.Contains(lower, "observer path") {
		t.Fatal("REQ-029 must require the architecture to describe the observer path")
	}
	if !strings.Contains(lower, "server-push") || !strings.Contains(lower, "polling") {
		t.Fatal("REQ-029 must address server-push vs polling in the observer path")
	}
	if !strings.Contains(lower, "aggregate") {
		t.Fatal("REQ-029 business rule 2 must require documented aggregate update paths")
	}
	if !strings.Contains(lower, "one http request per light") {
		t.Fatal("REQ-029 must forbid one-request-per-light as the only high-frequency path")
	}
	if !strings.Contains(lower, "http/2") && !strings.Contains(lower, "keep-alive") {
		t.Fatal("REQ-029 connection reuse rule should mention HTTP/2 or keep-alive")
	}
	if !strings.Contains(lower, "pool") {
		t.Fatal("REQ-029 should mention client pooling or equivalent connection reuse guidance")
	}
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-029 metadata must state priority Must")
	}
	for _, dep := range []string{"req-003", "req-011", "req-012", "req-015", "req-020"} {
		if !strings.Contains(lower, dep) {
			t.Fatalf("REQ-029 dependencies must include %s", strings.ToUpper(dep))
		}
	}
	if !strings.Contains(lower, "req-002") {
		t.Fatal("REQ-029 responsive notes must reference REQ-002")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-029 responsive notes must mention %q", kw)
		}
	}
}

func TestAcceptance_REQ029_architectureDocumentsThroughputAndSSE(t *testing.T) {
	root := repoRoot(t)
	archPath := filepath.Join(root, "docs", "architecture.md")
	archBytes, err := os.ReadFile(archPath)
	if err != nil {
		t.Fatalf("read %s: %v", archPath, err)
	}
	arch := string(archBytes)
	archLower := strings.ToLower(arch)
	if !strings.Contains(arch, "### 3.18 High-throughput light state updates") {
		t.Fatal("architecture must contain §3.18 High-throughput light state updates (REQ-029)")
	}
	for _, fragment := range []string{
		"**Write path",
		"**Observer path",
		"PATCH /api/v1/models/{id}/lights/state/batch",
		"GET /api/v1/models/{id}/lights/events",
		"GET /api/v1/scenes/{id}/lights/events",
		"text/event-stream",
		"EventSource",
	} {
		if !strings.Contains(arch, fragment) {
			t.Fatalf("architecture §3.18 / API tables must mention %q", fragment)
		}
	}
	if !strings.Contains(archLower, "server-sent") || !strings.Contains(archLower, "sse") {
		t.Fatal("architecture must document Server-Sent Events (SSE) for high-throughput observers")
	}
	if !strings.Contains(archLower, "keep-alive") {
		t.Fatal("architecture §3.18 must document HTTP keep-alive / connection reuse")
	}
	if !strings.Contains(archLower, "http/2") {
		t.Fatal("architecture §3.18 must mention HTTP/2 for reverse-proxy multiplexing")
	}

	readmePath := filepath.Join(root, "README.md")
	readmeBytes, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read %s: %v", readmePath, err)
	}
	readme := strings.ToLower(string(readmeBytes))
	if !strings.Contains(readme, "req-029") {
		t.Fatal("README must document REQ-029 (high-throughput light updates)")
	}
	if !strings.Contains(readme, "/lights/events") {
		t.Fatal("README must mention the /lights/events SSE URLs for integrators")
	}

	for _, rel := range []string{
		"web/app/models/detail/ModelDetailClient.tsx",
		"web/app/scenes/detail/SceneDetailClient.tsx",
	} {
		p := filepath.Join(root, rel)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		s := string(b)
		if !strings.Contains(s, "EventSource") {
			t.Fatalf("%s must use EventSource for REQ-029 observer path", rel)
		}
		if !strings.Contains(s, "/lights/events") {
			t.Fatalf("%s must subscribe to /lights/events", rel)
		}
	}

	routerPath := filepath.Join(root, "backend", "internal", "httpapi", "router.go")
	routerBytes, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatalf("read %s: %v", routerPath, err)
	}
	routerSrc := string(routerBytes)
	if !strings.Contains(routerSrc, `GET /models/{id}/lights/events`) {
		t.Fatal("router must register GET /models/{id}/lights/events for SSE (REQ-029)")
	}
	if !strings.Contains(routerSrc, `GET /scenes/{id}/lights/events`) {
		t.Fatal("router must register GET /scenes/{id}/lights/events for SSE (REQ-029)")
	}
	// Route order: literal "events" before patterns that treat a segment as light id.
	if strings.Index(routerSrc, `GET /models/{id}/lights/events`) >= strings.Index(routerSrc, `GET /models/{id}/lights/{lightId}/state`) {
		t.Fatal("router must register GET .../lights/events before GET .../lights/{lightId}/state")
	}
}

func TestAcceptance_REQ021_sceneRoutinesPythonAndSceneAPI(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-021")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-021 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-021 metadata must state priority Must")
	}
	for _, kw := range []string{"python", "req-022", "req-020", "req-011", "req-015"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-021 must reference %q (scope, rules, or dependencies)", kw)
		}
	}
	if !strings.Contains(lower, "start") || !strings.Contains(lower, "stop") {
		t.Fatal("REQ-021 must require starting and stopping routine runs")
	}
	if !strings.Contains(lower, "scene") {
		t.Fatal("REQ-021 must scope automation to scenes / scene API")
	}
	if !strings.Contains(lower, "req-032") {
		t.Fatal("REQ-021 must reference REQ-032 for default seeded Python behaviors")
	}
	if !strings.Contains(lower, "non-python") && !strings.Contains(lower, "python-only") {
		t.Fatal("REQ-021 must clarify Python-only routine definitions vs non-Python types")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-021 responsive notes must mention %q", kw)
		}
	}

	root := repoRoot(t)

	routerPath := filepath.Join(root, "backend", "internal", "httpapi", "router.go")
	routerBytes, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatalf("read %s: %v", routerPath, err)
	}
	routerSrc := string(routerBytes)
	for _, needle := range []string{
		`GET /routines`,
		`POST /routines`,
		`/scenes/{id}/routines/{routineId}/start`,
		`/scenes/{id}/routines/runs/{runId}/stop`,
		`/scenes/{id}/routines/runs`,
	} {
		if !strings.Contains(routerSrc, needle) {
			t.Fatalf("router.go must register %q for REQ-021 routine lifecycle", needle)
		}
	}

	mainPath := filepath.Join(root, "backend", "cmd", "server", "main.go")
	mainBytes, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read %s: %v", mainPath, err)
	}
	mainSrc := string(mainBytes)
	if strings.Contains(mainSrc, "StartRoutineScheduler") {
		t.Fatal("REQ-021 / architecture §3.16: server must not run StartRoutineScheduler (no Go routine ticker)")
	}

	storeRoutinesPath := filepath.Join(root, "backend", "internal", "store", "routines.go")
	routinesGo, err := os.ReadFile(storeRoutinesPath)
	if err != nil {
		t.Fatalf("read %s: %v", storeRoutinesPath, err)
	}
	routinesGoStr := string(routinesGo)
	if strings.Contains(routinesGoStr, "TickRoutineRuns") {
		t.Fatal("store/routines.go must not define TickRoutineRuns (automation runs in the browser: Pyodide and/or shape animation engine)")
	}

	sceneClientPath := filepath.Join(root, "web", "app", "scenes", "detail", "SceneDetailClient.tsx")
	sceneBytes, err := os.ReadFile(sceneClientPath)
	if err != nil {
		t.Fatalf("read %s: %v", sceneClientPath, err)
	}
	sceneSrc := string(sceneBytes)
	if !strings.Contains(sceneSrc, "startSceneRoutine") || !strings.Contains(sceneSrc, "stopSceneRoutineRun") {
		t.Fatal("SceneDetailClient must call startSceneRoutine and stopSceneRoutineRun (REQ-021 UI)")
	}

	workerPath := filepath.Join(root, "web", "public", "dlm-python-scene-worker.mjs")
	workerBytes, err := os.ReadFile(workerPath)
	if err != nil {
		t.Fatalf("read %s: %v", workerPath, err)
	}
	ws := string(workerBytes)
	if !strings.Contains(ws, "/api/v1/scenes/") || !strings.Contains(ws, "runPythonAsync") {
		t.Fatal("Pyodide worker must drive scene API via fetch to /api/v1/scenes/... (REQ-021 rule 3)")
	}

	archPath := filepath.Join(root, "docs", "architecture.md")
	archBytes, err := os.ReadFile(archPath)
	if err != nil {
		t.Fatalf("read %s: %v", archPath, err)
	}
	archStr := string(archBytes)
	archLower := strings.ToLower(archStr)
	if !strings.Contains(archLower, "pyodide") || !strings.Contains(archLower, "req-021") {
		t.Fatal("docs/architecture.md must document Pyodide execution and REQ-021 for scene routines")
	}
	if !strings.Contains(archStr, "time.Ticker") {
		t.Fatal("docs/architecture.md must explicitly address time.Ticker vs browser automation (REQ-021 / §3.16)")
	}
}

func TestAcceptance_REQ032_threeSeededPythonSampleRoutines(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-032")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-032 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-032 metadata must state priority Must")
	}
	for _, kw := range []string{"three", "sphere", "cuboid", "random", "req-017", "req-022", "python"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-032 must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "factory") || !strings.Contains(lower, "reset") {
		t.Fatal("REQ-032 must tie re-seeding to factory reset (REQ-017)")
	}
	if !strings.Contains(lower, "req-024") {
		t.Fatal("REQ-032 must reference REQ-024 (catalog vs whole-script delivery)")
	}
	for _, kw := range []string{"req-011", "req-020", "req-021", "req-026", "req-030"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-032 dependencies or body must reference %s", kw)
		}
	}

	root := repoRoot(t)

	samplesPath := filepath.Join(root, "web", "lib", "pythonRoutineSamples.ts")
	sb, err := os.ReadFile(samplesPath)
	if err != nil {
		t.Fatalf("read %s: %v", samplesPath, err)
	}
	samples := string(sb)
	for _, export := range []string{
		"PYTHON_SAMPLE_GROWING_SPHERE_SOURCE",
		"PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE",
		"PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE",
	} {
		if !strings.Contains(samples, "export const "+export) {
			t.Fatalf("pythonRoutineSamples.ts must export %s (REQ-032 single source of truth)", export)
		}
	}

	genScript := filepath.Join(root, "scripts", "gen-python-routine-samples-go.mjs")
	if _, err := os.Stat(genScript); err != nil {
		t.Fatalf("REQ-032 TS→Go sync: %s must exist: %v", genScript, err)
	}

	seedGenPath := filepath.Join(root, "backend", "internal", "seed", "python_routine_samples_gen.go")
	seedBytes, err := os.ReadFile(seedGenPath)
	if err != nil {
		t.Fatalf("read %s: %v", seedGenPath, err)
	}
	seedStr := string(seedBytes)
	if !strings.Contains(seedStr, "DefaultPythonRoutineRows") || !strings.Contains(seedStr, "growingSphereSource") {
		t.Fatal("generated Go seed must define DefaultPythonRoutineRows and sample bodies (REQ-032)")
	}

	storePath := filepath.Join(root, "backend", "internal", "store", "store.go")
	storeBytes, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("read %s: %v", storePath, err)
	}
	storeStr := string(storeBytes)
	if !strings.Contains(storeStr, "SeedDefaultPythonRoutines") || !strings.Contains(storeStr, "seedDefaultPythonRoutinesTx") {
		t.Fatal("store.go must implement SeedDefaultPythonRoutines for REQ-032 fresh install + factory reset")
	}

	archPath := filepath.Join(root, "docs", "architecture.md")
	archBytes, err := os.ReadFile(archPath)
	if err != nil {
		t.Fatalf("read %s: %v", archPath, err)
	}
	archStr := string(archBytes)
	if !strings.Contains(archStr, "pythonRoutineSamples.ts") || !strings.Contains(archStr, "REQ-032") {
		t.Fatal("docs/architecture.md must name pythonRoutineSamples.ts and REQ-032 seed placement (business rule 7)")
	}

	catalogPath := filepath.Join(root, "web", "lib", "pythonSceneApiCatalog.ts")
	catBytes, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("read %s: %v", catalogPath, err)
	}
	catStr := string(catBytes)
	if !strings.Contains(catStr, "sample-growing-sphere") || !strings.Contains(catStr, "sample-sweeping-cuboid") || !strings.Contains(catStr, "sample-random-colour-cycle") {
		t.Fatal("pythonSceneApiCatalog must list all three REQ-032 sample catalog entry ids")
	}

	editorPath := filepath.Join(root, "web", "app", "routines", "python", "PythonRoutineEditorClient.tsx")
	edBytes, err := os.ReadFile(editorPath)
	if err != nil {
		t.Fatalf("read %s: %v", editorPath, err)
	}
	edStr := string(edBytes)
	if !strings.Contains(edStr, "Load growing sphere sample") || !strings.Contains(edStr, "Load sweeping cuboid sample") || !strings.Contains(edStr, "Load random colour cycle sample") {
		t.Fatal("PythonRoutineEditorClient must expose Load sample actions for all three REQ-032 scripts")
	}
}

func TestAcceptance_REQ033_shapeAnimationRoutines(t *testing.T) {
	doc := readRequirements(t)
	block := requirementBlock(doc, "REQ-033")
	if block == "" {
		t.Fatal("requirements must contain ### REQ-033 section")
	}
	lower := strings.ToLower(block)
	if !strings.Contains(block, "| **Priority** | Must |") {
		t.Fatal("REQ-033 metadata must state priority Must")
	}
	for _, kw := range []string{
		"req-021", "req-020", "req-027", "req-011", "req-023", "req-026",
		"sphere", "cuboid", "brightness",
	} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-033 must reference or require %q", kw)
		}
	}
	if !strings.Contains(lower, "m/s") && !strings.Contains(lower, "meter") {
		t.Fatal("REQ-033 must document SI speed (m/s or meters per second)")
	}
	if !strings.Contains(lower, "docs/architecture.md") {
		t.Fatal("REQ-033 rule 12 must require docs/architecture.md for persistence and simulation")
	}
	if !strings.Contains(lower, "shape") && !strings.Contains(lower, "animation") {
		t.Fatal("REQ-033 must name shape animation in scope or rules")
	}
	for _, kw := range []string{"mobile", "tablet", "desktop"} {
		if !strings.Contains(lower, kw) {
			t.Fatalf("REQ-033 responsive notes must mention %q", kw)
		}
	}
	if !strings.Contains(lower, "hover-only") && !strings.Contains(lower, "hover only") {
		t.Fatal("REQ-033 must forbid hover-only essential steps where UX notes apply")
	}

	root := repoRoot(t)

	validatePath := filepath.Join(root, "backend", "internal", "store", "shape_animation_def.go")
	vb, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("read %s: %v", validatePath, err)
	}
	vstr := string(vb)
	if !strings.Contains(vstr, "ValidateShapeAnimationDefinitionJSON") {
		t.Fatal("shape_animation_def.go must export ValidateShapeAnimationDefinitionJSON (REQ-033 server validation)")
	}

	storePath := filepath.Join(root, "backend", "internal", "store", "store.go")
	sb, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("read %s: %v", storePath, err)
	}
	storeStr := string(sb)
	if !strings.Contains(storeStr, "ensureDefinitionJSONColumn") || !strings.Contains(storeStr, "definition_json") {
		t.Fatal("store.go must migrate routines.definition_json for REQ-033")
	}

	routinesPath := filepath.Join(root, "backend", "internal", "store", "routines.go")
	rb, err := os.ReadFile(routinesPath)
	if err != nil {
		t.Fatalf("read %s: %v", routinesPath, err)
	}
	routinesStr := string(rb)
	if !strings.Contains(routinesStr, `RoutineTypeShapeAnimation`) || !strings.Contains(routinesStr, `"shape_animation"`) {
		t.Fatal("routines.go must define RoutineTypeShapeAnimation / shape_animation (REQ-033)")
	}

	httpRoutinesPath := filepath.Join(root, "backend", "internal", "httpapi", "routines.go")
	hb, err := os.ReadFile(httpRoutinesPath)
	if err != nil {
		t.Fatalf("read %s: %v", httpRoutinesPath, err)
	}
	hr := string(hb)
	if !strings.Contains(hr, "definition_json") || !strings.Contains(hr, "DefinitionJSON") {
		t.Fatal("httpapi/routines.go must accept definition_json on create/patch for shape_animation (REQ-033)")
	}

	enginePath := filepath.Join(root, "web", "lib", "shapeAnimationEngine.ts")
	eb, err := os.ReadFile(enginePath)
	if err != nil {
		t.Fatalf("read %s: %v", enginePath, err)
	}
	es := string(eb)
	for _, needle := range []string{"initShapeAnimationSim", "tickShapeAnimationSim", "buildBatchUpdatesFromSim"} {
		if !strings.Contains(es, needle) {
			t.Fatalf("shapeAnimationEngine.ts must define %s (REQ-033 client simulation)", needle)
		}
	}

	hostPath := filepath.Join(root, "web", "components", "ShapeAnimationRoutineHost.tsx")
	hostBytes, err := os.ReadFile(hostPath)
	if err != nil {
		t.Fatalf("read %s: %v", hostPath, err)
	}
	hostSrc := string(hostBytes)
	if !strings.Contains(hostSrc, "patchSceneLightsStateBatch") {
		t.Fatal("ShapeAnimationRoutineHost must call patchSceneLightsStateBatch for scene light writes (REQ-033)")
	}

	scenesLibPath := filepath.Join(root, "web", "lib", "scenes.ts")
	scb, err := os.ReadFile(scenesLibPath)
	if err != nil {
		t.Fatalf("read %s: %v", scenesLibPath, err)
	}
	scenesLib := string(scb)
	if !strings.Contains(scenesLib, "patchSceneLightsStateBatch") || !strings.Contains(scenesLib, "fetchSceneDimensions") {
		t.Fatal("scenes.ts must expose patchSceneLightsStateBatch and fetchSceneDimensions for REQ-033")
	}

	defPath := filepath.Join(root, "web", "lib", "shapeAnimationDefault.ts")
	db, err := os.ReadFile(defPath)
	if err != nil {
		t.Fatalf("read %s: %v", defPath, err)
	}
	if !strings.Contains(string(db), "SHAPE_ANIMATION_DEFAULT_DEFINITION") {
		t.Fatal("shapeAnimationDefault.ts must export SHAPE_ANIMATION_DEFAULT_DEFINITION (REQ-033 create default)")
	}

	newRoutinePath := filepath.Join(root, "web", "app", "routines", "new", "RoutineNewClient.tsx")
	nb, err := os.ReadFile(newRoutinePath)
	if err != nil {
		t.Fatalf("read %s: %v", newRoutinePath, err)
	}
	newRoutineSrc := string(nb)
	if !strings.Contains(newRoutineSrc, "ROUTINE_TYPE_SHAPE_ANIMATION") || !strings.Contains(newRoutineSrc, "SHAPE_ANIMATION_DEFAULT_DEFINITION") {
		t.Fatal("RoutineNewClient must offer shape_animation create path with default definition (REQ-023 + REQ-033)")
	}

	shapeEditorPath := filepath.Join(root, "web", "app", "routines", "shape", "ShapeRoutineEditorClient.tsx")
	seb, err := os.ReadFile(shapeEditorPath)
	if err != nil {
		t.Fatalf("read %s: %v", shapeEditorPath, err)
	}
	shapeEd := string(seb)
	if !strings.Contains(shapeEd, "SceneLightsCanvas") || !strings.Contains(shapeEd, "ShapeAnimationRoutineHost") {
		t.Fatal("ShapeRoutineEditorClient must embed SceneLightsCanvas and ShapeAnimationRoutineHost (REQ-033 + REQ-027 viewport)")
	}

	sceneDetailPath := filepath.Join(root, "web", "app", "scenes", "detail", "SceneDetailClient.tsx")
	sdb, err := os.ReadFile(sceneDetailPath)
	if err != nil {
		t.Fatalf("read %s: %v", sceneDetailPath, err)
	}
	sceneDetail := string(sdb)
	if !strings.Contains(sceneDetail, "ShapeAnimationRoutineHost") || !strings.Contains(sceneDetail, "ROUTINE_TYPE_SHAPE_ANIMATION") {
		t.Fatal("SceneDetailClient must host ShapeAnimationRoutineHost when routine_type is shape_animation (REQ-021 + REQ-033)")
	}

	archPath := filepath.Join(root, "docs", "architecture.md")
	ab, err := os.ReadFile(archPath)
	if err != nil {
		t.Fatalf("read %s: %v", archPath, err)
	}
	arch := string(ab)
	archLower := strings.ToLower(arch)
	if !strings.Contains(archLower, "req-033") || !strings.Contains(archLower, "3.17.2") {
		t.Fatal("docs/architecture.md must document REQ-033 and §3.17.2 shape animation")
	}
	if !strings.Contains(archLower, "definition_json") {
		t.Fatal("docs/architecture.md must document definition_json for shape routines (REQ-033)")
	}
}
