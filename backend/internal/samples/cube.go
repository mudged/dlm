package samples

import (
	"math"

	"example.com/dlm/backend/internal/wiremodel"
)

// cubeGridN waypoints per axis on each face (interior grid in local u,v).
// 4×4 → 15 edges/face × ~6 subdivisions + bridges ≈ 720 total.
const cubeGridN = 4

// cubeLim half-extent in face-local coordinates for waypoint grid (|u|,|v| < 1).
const cubeLim = 0.95

type cubeFace int

const (
	cubeFaceZPos cubeFace = iota // z = +cubeHalf
	cubeFaceXPos                 // x = +cubeHalf
	cubeFaceYPos                 // y = +cubeHalf
	cubeFaceZNeg                 // z = -cubeHalf
	cubeFaceXNeg                 // x = -cubeHalf
	cubeFaceYNeg                 // y = -cubeHalf
)

// uvToCube maps face-local (u,v) in [-lim,lim]² to a 3D point on that face.
func uvToCube(f cubeFace, u, v float64) [3]float64 {
	switch f {
	case cubeFaceZPos:
		return [3]float64{u, v, cubeHalf}
	case cubeFaceZNeg:
		return [3]float64{u, v, -cubeHalf}
	case cubeFaceXPos:
		return [3]float64{cubeHalf, u, v}
	case cubeFaceXNeg:
		return [3]float64{-cubeHalf, u, v}
	case cubeFaceYPos:
		return [3]float64{u, cubeHalf, v}
	case cubeFaceYNeg:
		return [3]float64{u, -cubeHalf, v}
	default:
		return [3]float64{u, v, cubeHalf}
	}
}

// faceSerpentineWaypoints returns nu×nv points in serpentine order on face f (local u then v).
func faceSerpentineWaypoints(f cubeFace, nu, nv int, lim float64) [][3]float64 {
	if nu < 2 || nv < 2 {
		return nil
	}
	du := (2 * lim) / float64(nu-1)
	dv := (2 * lim) / float64(nv-1)
	out := make([][3]float64, 0, nu*nv)
	for j := 0; j < nv; j++ {
		v := -lim + float64(j)*dv
		if j%2 == 0 {
			for i := 0; i < nu; i++ {
				u := -lim + float64(i)*du
				out = append(out, uvToCube(f, u, v))
			}
		} else {
			for i := nu - 1; i >= 0; i-- {
				u := -lim + float64(i)*du
				out = append(out, uvToCube(f, u, v))
			}
		}
	}
	return out
}

func reversePoints3(a [][3]float64) [][3]float64 {
	out := make([][3]float64, len(a))
	for i := range a {
		out[i] = a[len(a)-1-i]
	}
	return out
}

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

// axisSegmentOnCubeSurface checks samples along an axis-aligned segment stay on the closed cube surface.
func axisSegmentOnCubeSurface(a, b [3]float64) bool {
	const steps = 24
	for s := 1; s < steps; s++ {
		t := float64(s) / float64(steps)
		p := [3]float64{
			a[0] + t*(b[0]-a[0]),
			a[1] + t*(b[1]-a[1]),
			a[2] + t*(b[2]-a[2]),
		}
		if !pointOnCubeSurface(p) {
			return false
		}
	}
	return true
}

// appendCubeBridgePoints walks from `from` to `to` on the cube boundary using axis-aligned legs.
func appendCubeBridgePoints(dst [][3]float64, from, to [3]float64) [][3]float64 {
	cur := from
	guard := 0
	for dist3(cur, to) > 1e-6 && guard < 64 {
		guard++
		var nxt [3]float64
		found := false
		for _, cand := range [][3]float64{
			{to[0], cur[1], cur[2]},
			{cur[0], to[1], cur[2]},
			{cur[0], cur[1], to[2]},
		} {
			if !pointOnCubeSurface(cand) {
				continue
			}
			if !axisSegmentOnCubeSurface(cur, cand) {
				continue
			}
			if dist3(cur, cand) < 1e-9 {
				continue
			}
			nxt = cand
			found = true
			break
		}
		if !found {
			break
		}
		seg := subdivideEdge(cur[0], cur[1], cur[2], nxt[0], nxt[1], nxt[2])
		dst = appendChain(dst, seg)
		cur = nxt
	}
	return dst
}

// CubeLights places lights on all six cube faces using interior serpentines plus surface bridges.
func CubeLights() []wiremodel.Light {
	lim := cubeLim
	order := []cubeFace{
		cubeFaceZPos, cubeFaceXPos, cubeFaceYPos,
		cubeFaceZNeg, cubeFaceXNeg, cubeFaceYNeg,
	}
	var dst [][3]float64
	for fi, f := range order {
		wps := faceSerpentineWaypoints(f, cubeGridN, cubeGridN, lim)
		if fi > 0 {
			last := dst[len(dst)-1]
			if dist3(last, wps[len(wps)-1]) < dist3(last, wps[0]) {
				wps = reversePoints3(wps)
			}
			dst = appendCubeBridgePoints(dst, last, wps[0])
		}
		for wi := 0; wi < len(wps)-1; wi++ {
			a, b := wps[wi], wps[wi+1]
			seg := subdivideEdge(a[0], a[1], a[2], b[0], b[1], b[2])
			dst = appendChain(dst, seg)
		}
	}
	return assignIDs(dst)
}

// NameCube is the canonical English sample name (REQ-009).
const NameCube = "Sample cube"
