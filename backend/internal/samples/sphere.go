package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// SphereLights returns lights on a sphere of diameter 2 m (R=1 m) as a single
// polyline on the surface with consecutive chord length chordTarget (REQ-009).
func SphereLights() []wiremodel.Light {
	beta := 2 * math.Asin(chordTarget/(2*sphereR))
	var pts [][3]float64
	p := [3]float64{0, 0, -sphereR}
	twist := 0.04 // rad: small rotation about Z each step for coverage
	const maxPts = 620
	for len(pts) < maxPts {
		pts = append(pts, p)
		px, py, pz := p[0]/sphereR, p[1]/sphereR, p[2]/sphereR
		kx := -py
		ky := px
		kz := 0.0
		kn := math.Hypot(kx, ky)
		if kn < 1e-9 {
			kx, ky, kn = 1, 0, 1
		}
		kx /= kn
		ky /= kn
		qx, qy, qz := rodrigues(px, py, pz, kx, ky, kz, beta)
		p = [3]float64{qx * sphereR, qy * sphereR, qz * sphereR}
		ct, st := math.Cos(twist), math.Sin(twist)
		x, y := p[0]*ct-p[1]*st, p[0]*st+p[1]*ct
		p = [3]float64{x, y, p[2]}
		n := math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
		if n < 1e-9 {
			break
		}
		p[0], p[1], p[2] = p[0]/n*sphereR, p[1]/n*sphereR, p[2]/n*sphereR
		if len(pts) > 20 && p[2] > sphereR-0.02 {
			break
		}
	}
	return assignIDs(pts)
}

func rodrigues(px, py, pz, kx, ky, kz, ang float64) (float64, float64, float64) {
	dot := kx*px + ky*py + kz*pz
	cx := ky*pz - kz*py
	cy := kz*px - kx*pz
	cz := kx*py - ky*px
	cb, sb := math.Cos(ang), math.Sin(ang)
	return px*cb + cx*sb + kx*dot*(1 - cb),
		py*cb + cy*sb + ky*dot*(1 - cb),
		pz*cb + cz*sb + kz*dot*(1 - cb)
}

// NameSphere is the canonical English sample name (REQ-009).
const NameSphere = "Sample sphere"
