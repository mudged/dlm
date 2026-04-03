package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

const (
	chordTarget = 0.1 // 10 cm (REQ-009)
	sphereR     = 1.0 // diameter 2 m
	cubeHalf    = 1.0 // edge length 2 m, vertices at ±1
	coneH       = 2.0 // height 2 m
	coneR       = 1.0 // base radius 1 m
)

// ChordTolerance is the max allowed deviation from chordTarget in tests. Straight
// segments whose length is not a multiple of 0.1 m (e.g. cone slant √5) use equal
// subdivisions, so step length may differ slightly from chordTarget.
const ChordTolerance = 2e-3

func dist(a, b wiremodel.Light) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func assignIDs(pts [][3]float64) []wiremodel.Light {
	out := make([]wiremodel.Light, len(pts))
	for i := range pts {
		out[i] = wiremodel.Light{ID: i, X: pts[i][0], Y: pts[i][1], Z: pts[i][2]}
	}
	return out
}

// subdivideEdge emits points along the closed segment [a,b] inclusive, step ~chordTarget Euclidean.
func subdivideEdge(ax, ay, az, bx, by, bz float64) [][3]float64 {
	dx := bx - ax
	dy := by - ay
	dz := bz - az
	L := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if L < 1e-12 {
		return [][3]float64{{ax, ay, az}}
	}
	n := int(math.Round(L / chordTarget))
	if n < 1 {
		n = 1
	}
	var out [][3]float64
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		out = append(out, [3]float64{
			ax + t*dx,
			ay + t*dy,
			az + t*dz,
		})
	}
	return out
}

func appendChain(dst [][3]float64, chain [][3]float64) [][3]float64 {
	if len(chain) == 0 {
		return dst
	}
	if len(dst) == 0 {
		return append(dst, chain...)
	}
	// skip duplicate joint vertex
	if samePoint(dst[len(dst)-1], chain[0]) {
		return append(dst, chain[1:]...)
	}
	return append(dst, chain...)
}

func samePoint(a, b [3]float64) bool {
	return math.Abs(a[0]-b[0]) < 1e-9 && math.Abs(a[1]-b[1]) < 1e-9 && math.Abs(a[2]-b[2]) < 1e-9
}
