package samples

import (
	"math"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

func assertChordSteps(t *testing.T, name string, lights []wiremodel.Light) {
	t.Helper()
	if len(lights) < 2 {
		t.Fatalf("%s: need at least 2 lights", name)
	}
	for i := 1; i < len(lights); i++ {
		d := dist(lights[i-1], lights[i])
		if math.Abs(d-chordTarget) > ChordTolerance {
			t.Fatalf("%s: chord %d->%d = %v want %v", name, i-1, i, d, chordTarget)
		}
	}
}

func TestSphereLights_chordAndCap(t *testing.T) {
	L := SphereLights()
	if len(L) > 1000 {
		t.Fatalf("len %d > 1000", len(L))
	}
	assertChordSteps(t, "sphere", L)
	for _, p := range L {
		r := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
		if math.Abs(r-sphereR) > 1e-4 {
			t.Fatalf("point not on sphere r=%v", r)
		}
	}
}

func TestCubeLights_chordAndCap(t *testing.T) {
	L := CubeLights()
	if len(L) > 1000 {
		t.Fatalf("len %d > 1000", len(L))
	}
	assertChordSteps(t, "cube", L)
}

func TestConeLights_chordAndCap(t *testing.T) {
	L := ConeLights()
	if len(L) > 1000 {
		t.Fatalf("len %d > 1000", len(L))
	}
	assertChordSteps(t, "cone", L)
}

func TestSampleNames(t *testing.T) {
	if NameSphere == "" || NameCube == "" || NameCone == "" {
		t.Fatal("sample names must be non-empty")
	}
}

func TestWiremodelLightIDsContiguous(t *testing.T) {
	for _, fn := range []func() []wiremodel.Light{SphereLights, CubeLights, ConeLights} {
		L := fn()
		for i := range L {
			if L[i].ID != i {
				t.Fatalf("id %d at index %d", L[i].ID, i)
			}
		}
	}
}
