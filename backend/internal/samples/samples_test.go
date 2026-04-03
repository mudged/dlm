package samples

import (
	"math"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

// openInteriorTol: classify points as lying on a nominal face plane.
const openInteriorTol = 1e-2

// looseInteriorMargin for the 85% quota (architecture §3.8).
const looseInteriorMargin = 0.02

// strictInteriorMargin for per-face evenness (excludes edge band and typical bridge legs).
const strictInteriorMargin = 0.07

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
	assertCubeFaceQuotas(t, L)
}

// cubeFaceOf assigns each surface point to one nominal face; x before y before z breaks ties on edges.
func cubeFaceOf(x, y, z float64) int {
	h := cubeHalf
	eps := openInteriorTol
	switch {
	case math.Abs(x-h) < eps && math.Abs(y) <= h+1e-6 && math.Abs(z) <= h+1e-6:
		return 2 // +X
	case math.Abs(x+h) < eps && math.Abs(y) <= h+1e-6 && math.Abs(z) <= h+1e-6:
		return 3 // -X
	case math.Abs(y-h) < eps && math.Abs(x) <= h+1e-6 && math.Abs(z) <= h+1e-6:
		return 4 // +Y
	case math.Abs(y+h) < eps && math.Abs(x) <= h+1e-6 && math.Abs(z) <= h+1e-6:
		return 5 // -Y
	case math.Abs(z-h) < eps && math.Abs(x) <= h+1e-6 && math.Abs(y) <= h+1e-6:
		return 0 // +Z
	case math.Abs(z+h) < eps && math.Abs(x) <= h+1e-6 && math.Abs(y) <= h+1e-6:
		return 1 // -Z
	default:
		return -1
	}
}

func cubeInteriorWithMargin(x, y, z float64, b float64) bool {
	h := cubeHalf
	switch cubeFaceOf(x, y, z) {
	case 0:
		return math.Abs(x) < h-b && math.Abs(y) < h-b
	case 1:
		return math.Abs(x) < h-b && math.Abs(y) < h-b
	case 2:
		return math.Abs(y) < h-b && math.Abs(z) < h-b
	case 3:
		return math.Abs(y) < h-b && math.Abs(z) < h-b
	case 4:
		return math.Abs(x) < h-b && math.Abs(z) < h-b
	case 5:
		return math.Abs(x) < h-b && math.Abs(z) < h-b
	default:
		return false
	}
}

func cubeOpenInteriorLoose(x, y, z float64) bool {
	return cubeInteriorWithMargin(x, y, z, looseInteriorMargin)
}

func cubeOpenInteriorStrict(x, y, z float64) bool {
	return cubeInteriorWithMargin(x, y, z, strictInteriorMargin)
}

func assertCubeFaceQuotas(t *testing.T, lights []wiremodel.Light) {
	t.Helper()
	n := len(lights)
	q := n / 6
	counts := [6]int{}
	interiorLoose := 0
	intFace := [6]int{}
	for _, p := range lights {
		f := cubeFaceOf(p.X, p.Y, p.Z)
		if f < 0 {
			t.Fatalf("cube: point (%v,%v,%v) not on a nominal face", p.X, p.Y, p.Z)
		}
		counts[f]++
		if cubeOpenInteriorLoose(p.X, p.Y, p.Z) {
			interiorLoose++
		}
		if cubeOpenInteriorStrict(p.X, p.Y, p.Z) {
			intFace[f]++
		}
	}
	minI, maxI := intFace[0], intFace[0]
	for f := 0; f < 6; f++ {
		c := intFace[f]
		if c < minI {
			minI = c
		}
		if c > maxI {
			maxI = c
		}
		if c < 20 {
			t.Fatalf("cube: strict-interior count on face %d is %d (want >= 20)", f, c)
		}
	}
	if maxI-minI > 8 {
		t.Fatalf("cube: strict-interior per-face counts %v differ by more than 8 (n=%d q=%d)", intFace, n, q)
	}
	need := int(0.85 * float64(n))
	if interiorLoose < need {
		t.Fatalf("cube: loose-interior lights %d want >= %d (85%% of %d)", interiorLoose, need, n)
	}
	_ = counts // all points classified; strict interior uses intFace
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
