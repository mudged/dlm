package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// Chord band per REQ-009 (consecutive lights along the polyline).
const (
	chordMin   = 0.05 // 5 cm
	chordMax   = 0.10 // 10 cm
	chordIdeal = 0.075
)

const (
	sphereR  = 1.0 // diameter 2 m
	cubeHalf = 1.0 // edge length 2 m
	coneH    = 2.0
	coneR    = 1.0
)

// SpatialTestTol relaxes chord band checks in tests (float + sampling).
const SpatialTestTol = 1e-3

// OnSurfaceTol: generated points on analytic surface should satisfy |error| < this.
const OnSurfaceTol = 1e-4

func dist(a, b wiremodel.Light) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Hypot(dx, math.Hypot(dy, dz))
}

func dist3(a, b [3]float64) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return math.Hypot(dx, math.Hypot(dy, dz))
}

func assignIDs(pts [][3]float64) []wiremodel.Light {
	out := make([]wiremodel.Light, len(pts))
	for i := range pts {
		out[i] = wiremodel.Light{ID: i, X: pts[i][0], Y: pts[i][1], Z: pts[i][2]}
	}
	return out
}

// segmentCountForChordBand returns the number of equal straight segments along an open
// edge of Euclidean length L so that each segment length lies in [chordMin, chordMax].
func segmentCountForChordBand(L float64) int {
	if L <= 1e-12 {
		return 0
	}
	if L >= chordMin && L <= chordMax {
		return 1
	}
	nLo := int(math.Ceil(L / chordMax))
	if nLo < 1 {
		nLo = 1
	}
	nHi := int(math.Floor(L / chordMin))
	if nHi < nLo {
		return nLo
	}
	ideal := L / chordIdeal
	n := int(math.Round(ideal))
	if n < nLo {
		n = nLo
	}
	if n > nHi {
		n = nHi
	}
	return n
}

// subdivideEdge emits points along the closed segment [a,b] inclusive; interior step
// lengths lie in [chordMin, chordMax] when L is large enough to allow it.
func subdivideEdge(ax, ay, az, bx, by, bz float64) [][3]float64 {
	dx := bx - ax
	dy := by - ay
	dz := bz - az
	L := math.Hypot(dx, math.Hypot(dy, dz))
	if L < 1e-12 {
		return [][3]float64{{ax, ay, az}}
	}
	n := segmentCountForChordBand(L)
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
	if samePoint(dst[len(dst)-1], chain[0]) {
		return append(dst, chain[1:]...)
	}
	return append(dst, chain...)
}

func samePoint(a, b [3]float64) bool {
	return math.Abs(a[0]-b[0]) < 1e-9 && math.Abs(a[1]-b[1]) < 1e-9 && math.Abs(a[2]-b[2]) < 1e-9
}

func norm3(v [3]float64) [3]float64 {
	n := math.Hypot(v[0], math.Hypot(v[1], v[2]))
	if n < 1e-15 {
		return v
	}
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}

func dot3(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// slerpUnit interpolates between unit vectors u0, u1 (t in [0,1]).
func slerpUnit(u0, u1 [3]float64, t float64) [3]float64 {
	d := clamp(dot3(u0, u1), -1, 1)
	if d > 1-1e-10 {
		x := u0[0] + t*(u1[0]-u0[0])
		y := u0[1] + t*(u1[1]-u0[1])
		z := u0[2] + t*(u1[2]-u0[2])
		return norm3([3]float64{x, y, z})
	}
	w := math.Acos(d)
	s0 := math.Sin((1-t)*w) / math.Sin(w)
	s1 := math.Sin(t*w) / math.Sin(w)
	return norm3([3]float64{
		s0*u0[0] + s1*u1[0],
		s0*u0[1] + s1*u1[1],
		s0*u0[2] + s1*u1[2],
	})
}

// subdivideGreatCircleArc emits points along the shorter great-circle arc from p0 to p1 on
// a sphere of radius R centered at the origin.
func subdivideGreatCircleArc(p0, p1 [3]float64, R float64) [][3]float64 {
	u0 := norm3(p0)
	u1 := norm3(p1)
	omega := math.Acos(clamp(dot3(u0, u1), -1, 1))
	arcLen := R * omega
	if arcLen < 1e-9 {
		return [][3]float64{{u0[0] * R, u0[1] * R, u0[2] * R}}
	}
	n := segmentCountForChordBand(arcLen)
	var out [][3]float64
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		u := slerpUnit(u0, u1, t)
		out = append(out, [3]float64{u[0] * R, u[1] * R, u[2] * R})
	}
	return out
}

// coneRingAtY returns samples on the cone lateral surface at height y in [0, coneH):
// radius r = coneR * (1 - y/coneH), full azimuth in the xz plane (Y-up, base on xz at y=0).
func coneRingAtY(y float64) [][3]float64 {
	if y >= coneH-1e-9 {
		return [][3]float64{{0, coneH, 0}}
	}
	r := coneR * (1 - y/coneH)
	if r < 1e-9 {
		return [][3]float64{{0, y, 0}}
	}
	C := 2 * math.Pi * r
	nLo := int(math.Ceil(C / chordMax))
	nHi := int(math.Floor(C / chordMin))
	if nHi < nLo {
		nHi = nLo
	}
	n := nLo
	ideal := int(math.Round(C / chordIdeal))
	if ideal >= nLo && ideal <= nHi {
		n = ideal
	} else if ideal > nHi {
		n = nHi
	}
	if n < 3 {
		n = 3
	}
	var out [][3]float64
	for k := 0; k < n; k++ {
		theta := 2 * math.Pi * float64(k) / float64(n)
		out = append(out, [3]float64{r * math.Cos(theta), y, r * math.Sin(theta)})
	}
	return out
}
