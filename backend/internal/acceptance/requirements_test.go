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
