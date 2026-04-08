package store

import (
	"testing"

	"example.com/dlm/backend/internal/lightstate"
)

func TestEquivLightStateTriple(t *testing.T) {
	if !lightstate.EquivLightStateTriple(true, "#aabbcc", 50, true, "#AABBCC", 50) {
		t.Fatal("case and spacing-insensitive color should match")
	}
	if lightstate.EquivLightStateTriple(true, "#aabbcc", 50, false, "#aabbcc", 50) {
		t.Fatal("on differs")
	}
	if lightstate.EquivLightStateTriple(true, "#aabbcc", 50, true, "#aabbcd", 50) {
		t.Fatal("color differs")
	}
	if lightstate.EquivLightStateTriple(true, "#aabbcc", 50, true, "#aabbcc", 50+2e-9) {
		t.Fatal("brightness outside tolerance")
	}
	if !lightstate.EquivLightStateTriple(true, "#aabbcc", 50, true, "#aabbcc", 50+5e-10) {
		t.Fatal("brightness within tolerance should match")
	}
}
