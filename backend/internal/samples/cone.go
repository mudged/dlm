package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// coneTotalLights deterministic count in [500, 1000] (REQ-009).
const coneTotalLights = 720

// coneDZ steps height rings on the lateral surface.
const coneDZ = 0.11

// coneLateralSlant is lateral surface area π r ℓ.
func coneLateralSlant() float64 {
	return math.Pi * coneR * math.Sqrt(coneR*coneR+coneH*coneH)
}

func coneBaseArea() float64 {
	return math.Pi * coneR * coneR
}

// pointOnConeLateral reports whether (x,y,z) lies on the lateral surface (including apex) or is near it.
func pointOnConeLateral(x, y, z float64) bool {
	if math.Abs(z-coneH) < 1e-3 && math.Abs(x) < 1e-3 && math.Abs(y) < 1e-3 {
		return true
	}
	if z < -1e-3 || z > coneH+1e-3 {
		return false
	}
	r := math.Hypot(x, y)
	rExp := coneR * (1 - z/coneH)
	if rExp < 1e-6 {
		return r < 1e-3
	}
	return math.Abs(r-rExp) < 5e-3
}

// pointOnConeBase reports z≈0 and ρ≤coneR (base disk).
func pointOnConeBase(x, y, z float64) bool {
	if math.Abs(z) > 5e-3 {
		return false
	}
	return math.Hypot(x, y) <= coneR+5e-3
}

// coneBaseSpiral fills the base disk with a polyline from center toward rim (chord-bounded).
func coneBaseSpiral(nBudget int) [][3]float64 {
	if nBudget < 2 {
		nBudget = 2
	}
	const eps = 0.045
	var pts [][3]float64
	theta := 0.0
	r := 0.0
	pts = append(pts, [3]float64{0, 0, 0})
	for len(pts) < nBudget && r < coneR-eps {
		rNext := r + chordIdeal*0.42
		if rNext > coneR-eps {
			rNext = coneR - eps
		}
		theta += chordIdeal / math.Max(rNext, eps)
		x := rNext * math.Cos(theta)
		y := rNext * math.Sin(theta)
		last := pts[len(pts)-1]
		seg := subdivideEdge(last[0], last[1], last[2], x, y, 0)
		pts = appendChain(pts, seg[1:])
		r = rNext
	}
	return pts
}

// coneLateralToApex stacked rings + ruling to apex (lateral surface only).
func coneLateralToApex(nBudget int) [][3]float64 {
	var pts [][3]float64
	z := 0.0
	for z < coneH-coneDZ*0.4 && len(pts) < nBudget {
		ring := coneRingAtZ(z)
		if len(pts) == 0 {
			pts = append(pts, ring...)
		} else {
			arc := subdivideEdge(pts[len(pts)-1][0], pts[len(pts)-1][1], pts[len(pts)-1][2],
				ring[0][0], ring[0][1], ring[0][2])
			pts = appendChain(pts, arc[1:])
			pts = appendChain(pts, ring[1:])
		}
		z += coneDZ
	}
	apex := [3]float64{0, 0, coneH}
	if len(pts) == 0 {
		return [][3]float64{apex}
	}
	slant := subdivideEdge(pts[len(pts)-1][0], pts[len(pts)-1][1], pts[len(pts)-1][2],
		apex[0], apex[1], apex[2])
	pts = appendChain(pts, slant[1:])
	return pts
}

func trimConeChain(pts [][3]float64, target int) [][3]float64 {
	if len(pts) <= target {
		return pts
	}
	for len(pts) > target {
		best := 1
		bestScore := 1e9
		for i := 1; i < len(pts)-1; i++ {
			d0 := dist3(pts[i-1], pts[i])
			d1 := dist3(pts[i], pts[i+1])
			d2 := dist3(pts[i-1], pts[i+1])
			if d2 > chordMax+1e-6 {
				continue
			}
			score := d0 + d1
			if score < bestScore {
				bestScore = score
				best = i
			}
		}
		if bestScore >= 1e8 {
			break
		}
		pts = append(pts[:best], pts[best+1:]...)
	}
	return pts
}

// ConeLights covers the base disk and lateral surface with area-proportional budgets (REQ-009 §3.8).
func ConeLights() []wiremodel.Light {
	latA := coneLateralSlant()
	baseA := coneBaseArea()
	nBase := int(math.Round(float64(coneTotalLights) * baseA / (latA + baseA)))
	if nBase < 100 {
		nBase = 100
	}
	nLat := coneTotalLights - nBase
	if nLat < 350 {
		nLat = 350
		nBase = coneTotalLights - nLat
	}

	basePts := coneBaseSpiral(nBase)
	latPts := coneLateralToApex(nLat * 2)

	var dst [][3]float64
	dst = append(dst, basePts...)
	if len(dst) > 0 && len(latPts) > 0 {
		seg := subdivideEdge(dst[len(dst)-1][0], dst[len(dst)-1][1], dst[len(dst)-1][2],
			latPts[0][0], latPts[0][1], latPts[0][2])
		dst = appendChain(dst, seg[1:])
		dst = appendChain(dst, latPts[1:])
	} else {
		dst = append(dst, latPts...)
	}

	dst = trimConeChain(dst, coneTotalLights)
	return assignIDs(dst)
}

// NameCone is the canonical English sample name (REQ-009).
const NameCone = "Sample cone"
