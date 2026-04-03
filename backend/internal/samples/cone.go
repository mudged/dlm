package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// ConeLights: base circle z=0, r=coneR, closed polygon with ~0.1 m chords; then slant
// from (coneR,0,0) to apex (0,0,coneH) (REQ-009).
func ConeLights() []wiremodel.Light {
	const N = 63 // 2*sin(π/N) ≈ 0.0993 m chord for R=1
	var pts [][3]float64
	for k := 0; k < N; k++ {
		theta := 2 * math.Pi * float64(k) / float64(N)
		pts = append(pts, [3]float64{
			coneR * math.Cos(theta),
			coneR * math.Sin(theta),
			0,
		})
	}
	first := pts[0]
	last := pts[len(pts)-1]
	closeSeg := subdivideEdge(last[0], last[1], last[2], first[0], first[1], first[2])
	pts = append(pts, closeSeg[1:]...)
	slant := subdivideEdge(first[0], first[1], first[2], 0, 0, coneH)
	pts = append(pts, slant[1:]...)
	return assignIDs(pts)
}

// NameCone is the canonical English sample name (REQ-009).
const NameCone = "Sample cone"
