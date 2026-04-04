package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// Vertical helix on the cube boundary: (x, y) advances along the square
// cross-section perimeter (same piecewise map as the reference Python generator)
// while z marches linearly from -cubeHalf to +cubeHalf.
//
// Reference logic: s = (i * spacing) % perimeter with horizontal (x,y) from s;
// z = -half + i * (edge / (n-1)). Spacing is chosen in meters so straight 3D
// chords stay in [chordMin, chordMax] including worst-case 90° perimeter turns
// (corner chord ≈ spacing/√2 in the horizontal plane plus a small z component).
const (
	cubeHelixNumLights = 801
	// Perimeter advance per light: chordMin*sqrt(2) keeps the worst 90° corner
	// straight chord (in 3D with vertical dz) within the REQ-009 band; a raw
	// 0.05 m step would dip below chordMin at corners.
	cubeHelixPerimeterStep = chordMin * math.Sqrt2
)

const cubeSurfEps = 2e-3

func pointOnCubeSurface(p [3]float64) bool {
	x, y, z := p[0], p[1], p[2]
	h := cubeHalf
	if math.Abs(math.Abs(x)-h) <= cubeSurfEps && math.Abs(y) <= h+cubeSurfEps && math.Abs(z) <= h+cubeSurfEps {
		return true
	}
	if math.Abs(math.Abs(y)-h) <= cubeSurfEps && math.Abs(x) <= h+cubeSurfEps && math.Abs(z) <= h+cubeSurfEps {
		return true
	}
	if math.Abs(math.Abs(z)-h) <= cubeSurfEps && math.Abs(x) <= h+cubeSurfEps && math.Abs(y) <= h+cubeSurfEps {
		return true
	}
	return false
}

// cubePerimeterXY maps distance s ∈ [0, perimeter) along the square perimeter in the xy plane
// (counterclockwise from (-h,-h) along y = -h) to (x, y) on the boundary.
// sideLength is the full edge length (2 * cubeHalf).
func cubePerimeterXY(halfL, sideLength, s float64) (x, y float64) {
	s = math.Mod(s, 4*sideLength)
	if s < 0 {
		s += 4 * sideLength
	}
	switch {
	case s <= sideLength:
		x = -halfL + s
		y = -halfL
	case s <= 2*sideLength:
		x = halfL
		y = -halfL + (s - sideLength)
	case s <= 3*sideLength:
		x = halfL - (s - 2*sideLength)
		y = halfL
	default:
		x = -halfL
		y = halfL - (s - 3*sideLength)
	}
	return x, y
}

// CubeLights places lights on a dense vertical helix wrapping the four vertical
// face planes (plus corners on z = ±cubeHalf), ordered by id along the chain.
func CubeLights() []wiremodel.Light {
	halfL := cubeHalf
	sideLength := 2 * halfL
	perimeter := 4 * sideLength
	zStep := (2 * halfL) / float64(cubeHelixNumLights-1)

	pts := make([][3]float64, cubeHelixNumLights)
	for i := 0; i < cubeHelixNumLights; i++ {
		s := math.Mod(float64(i)*cubeHelixPerimeterStep, perimeter)
		x, y := cubePerimeterXY(halfL, sideLength, s)
		z := -halfL + float64(i)*zStep
		pts[i] = [3]float64{x, y, z}
	}
	return assignIDs(pts)
}

// NameCube is the canonical English sample name (REQ-009).
const NameCube = "Sample cube"
