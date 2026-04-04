package samples

import (
	"math"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

const minLights, maxLights = 500, 1000

func assertChordBand(t *testing.T, name string, lights []wiremodel.Light) {
	t.Helper()
	if len(lights) < 2 {
		t.Fatalf("%s: need at least 2 lights", name)
	}
	for i := 1; i < len(lights); i++ {
		d := dist(lights[i-1], lights[i])
		if d < chordMin-SpatialTestTol || d > chordMax+SpatialTestTol {
			t.Fatalf("%s: chord %d->%d = %v want in [%v,%v]", name, i-1, i, d, chordMin, chordMax)
		}
	}
}

func assertLightCount(t *testing.T, name string, lights []wiremodel.Light) {
	t.Helper()
	n := len(lights)
	if n < minLights || n > maxLights {
		t.Fatalf("%s: len %d want %d..%d", name, n, minLights, maxLights)
	}
}

func TestSphereLights_surfaceChordCount(t *testing.T) {
	L := SphereLights()
	assertLightCount(t, "sphere", L)
	assertChordBand(t, "sphere", L)
	for _, p := range L {
		r := math.Hypot(p.X, math.Hypot(p.Y, p.Z))
		if math.Abs(r-sphereR) > 5e-3 { // on surface; small tol for float
			t.Fatalf("point off sphere r=%v want %v", r, sphereR)
		}
	}
	assertSphereHemisphereBound(t, L, 0.56)
}

// assertSphereHemisphereBound: no closed hemisphere (through origin) contains more than maxFrac of lights.
func assertSphereHemisphereBound(t *testing.T, lights []wiremodel.Light, maxFrac float64) {
	t.Helper()
	n := len(lights)
	if n == 0 {
		return
	}
	const ndir = 48
	golden := (1 + math.Sqrt(5)) / 2
	worst := 0
	for k := 0; k < ndir; k++ {
		y := 1 - 2*(float64(k)+0.5)/float64(ndir)
		rho := math.Sqrt(math.Max(0, 1-y*y))
		th := 2 * math.Pi * float64(k) / golden
		dx := math.Cos(th) * rho
		dy := y
		dz := math.Sin(th) * rho
		cnt := 0
		for _, p := range lights {
			if p.X*dx+p.Y*dy+p.Z*dz >= -1e-6 {
				cnt++
			}
		}
		if cnt > worst {
			worst = cnt
		}
	}
	if float64(worst)/float64(n) > maxFrac {
		t.Fatalf("sphere: hemisphere occupancy %d/%d > %v", worst, n, maxFrac)
	}
}

func TestCubeLights_surfaceChordCount(t *testing.T) {
	L := CubeLights()
	assertLightCount(t, "cube", L)
	assertChordBand(t, "cube", L)
	assertCubeLightsOnClosedSurface(t, L)
	assertCubeHelixZSpan(t, L)
}

func assertCubeLightsOnClosedSurface(t *testing.T, lights []wiremodel.Light) {
	t.Helper()
	for _, p := range lights {
		if !pointOnCubeSurface([3]float64{p.X, p.Y, p.Z}) {
			t.Fatalf("cube: point (%v,%v,%v) not on closed surface", p.X, p.Y, p.Z)
		}
	}
}

func assertCubeHelixZSpan(t *testing.T, lights []wiremodel.Light) {
	t.Helper()
	if len(lights) == 0 {
		return
	}
	z0 := lights[0].Z
	z1 := lights[len(lights)-1].Z
	const zTol = 5e-3
	if math.Abs(z0-(-cubeHalf)) > zTol {
		t.Fatalf("cube: helix start z %v want ~%v", z0, -cubeHalf)
	}
	if math.Abs(z1-cubeHalf) > zTol {
		t.Fatalf("cube: helix end z %v want ~%v", z1, cubeHalf)
	}
}

func TestConeLights_surfaceChordCount(t *testing.T) {
	L := ConeLights()
	assertLightCount(t, "cone", L)
	assertChordBand(t, "cone", L)
	for _, p := range L {
		if !pointOnConeLateral(p.X, p.Y, p.Z) && !pointOnConeBase(p.X, p.Y, p.Z) {
			t.Fatalf("point not on cone lateral/base/apex: (%v,%v,%v)", p.X, p.Y, p.Z)
		}
	}
	assertConeLateralBaseSplit(t, L)
}

func assertConeLateralBaseSplit(t *testing.T, lights []wiremodel.Light) {
	t.Helper()
	var nLat, nBase int
	for _, p := range lights {
		if pointOnConeBase(p.X, p.Y, p.Z) {
			nBase++
		}
		if pointOnConeLateral(p.X, p.Y, p.Z) {
			nLat++
		}
	}
	if nBase < 80 {
		t.Fatalf("cone: too few base-plane lights %d", nBase)
	}
	if nLat < 350 {
		t.Fatalf("cone: too few lateral lights %d", nLat)
	}
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
