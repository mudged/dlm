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
