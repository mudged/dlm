package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// sphereFibN is the number of Fibonacci-sphere seeds before great-circle chord subdivision (REQ-009).
// Tuned so final polyline length stays in [500, 1000].
const sphereFibN = 268

// fibonacciUnitSphere returns n quasi-uniform points on the unit sphere (golden-angle spiral).
func fibonacciUnitSphere(n int) [][3]float64 {
	if n <= 0 {
		return nil
	}
	pts := make([][3]float64, n)
	golden := (1 + math.Sqrt(5)) / 2
	for i := 0; i < n; i++ {
		y := 1 - 2*(float64(i)+0.5)/float64(n)
		r := math.Sqrt(math.Max(0, 1-y*y))
		theta := 2 * math.Pi * float64(i) / golden
		pts[i] = [3]float64{math.Cos(theta) * r, y, math.Sin(theta) * r}
	}
	return pts
}

func scaleToRadius(pts [][3]float64, R float64) [][3]float64 {
	out := make([][3]float64, len(pts))
	for i, p := range pts {
		n := math.Hypot(p[0], math.Hypot(p[1], p[2]))
		if n < 1e-15 {
			out[i] = [3]float64{0, R, 0}
			continue
		}
		s := R / n
		out[i] = [3]float64{p[0] * s, p[1] * s, p[2] * s}
	}
	return out
}

func greedyNNOrder(pts [][3]float64, start int) [][3]float64 {
	if len(pts) == 0 {
		return nil
	}
	used := make([]bool, len(pts))
	out := make([][3]float64, 0, len(pts))
	cur := start
	for len(out) < len(pts) {
		used[cur] = true
		out = append(out, pts[cur])
		best := -1
		bestD := 0.0
		for j := range pts {
			if used[j] {
				continue
			}
			d := dist3(pts[cur], pts[j])
			if best < 0 || d < bestD || (math.Abs(d-bestD) < 1e-12 && j < best) {
				bestD = d
				best = j
			}
		}
		if best < 0 {
			break
		}
		cur = best
	}
	return out
}

func findStartMaxZ(pts [][3]float64) int {
	if len(pts) == 0 {
		return 0
	}
	bi := 0
	mz := pts[0][2]
	for i := 1; i < len(pts); i++ {
		if pts[i][2] > mz {
			mz = pts[i][2]
			bi = i
		}
	}
	return bi
}

// SphereLights returns lights on ‖p‖=sphereR with area-even Fibonacci seeds, greedy NN ordering,
// and great-circle subdivision into the REQ-009 chord band (architecture §3.8).
func SphereLights() []wiremodel.Light {
	raw := scaleToRadius(fibonacciUnitSphere(sphereFibN), sphereR)
	order := greedyNNOrder(raw, findStartMaxZ(raw))
	var dst [][3]float64
	for i := 0; i < len(order)-1; i++ {
		seg := subdivideGreatCircleArc(order[i], order[i+1], sphereR)
		dst = appendChain(dst, seg)
	}
	return assignIDs(dst)
}

// NameSphere is the canonical English sample name (REQ-009).
const NameSphere = "Sample sphere"
