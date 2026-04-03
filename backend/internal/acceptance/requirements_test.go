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
	if !strings.Contains(block, "0.01") {
		t.Fatal("REQ-010 must state 0.01 m (1 cm) sphere diameter")
	}
	if !strings.Contains(lower, "white") {
		t.Fatal("REQ-010 must require white spheres")
	}
	if !strings.Contains(lower, "every") && !strings.Contains(lower, "all") {
		t.Fatal("REQ-010 must require every/all lights to be drawn")
	}
	if !strings.Contains(lower, "transparent") || !strings.Contains(lower, "thin") {
		t.Fatal("REQ-010 must require thin transparent segments")
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
