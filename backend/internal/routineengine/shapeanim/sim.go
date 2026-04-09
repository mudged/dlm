// Package shapeanim ports web/lib/shapeAnimationEngine.ts simulation for server-side ticks (REQ-033 / §3.17.2).
package shapeanim

import (
	"encoding/json"
	"fmt"
	"math"

	"example.com/dlm/backend/internal/store"
)

// Dimensions is scene AABB in scene space (origin min corner, max inclusive upper bound).
type Dimensions struct {
	Min, Max struct {
		X, Y, Z float64
	}
}

// FromStore maps store.SceneDimensions to shapeanim.Dimensions.
func FromStore(d *store.SceneDimensions) Dimensions {
	var out Dimensions
	if d == nil {
		return out
	}
	out.Min.X = d.Origin.X
	out.Min.Y = d.Origin.Y
	out.Min.Z = d.Origin.Z
	out.Max.X = d.Max.X
	out.Max.Y = d.Max.Y
	out.Max.Z = d.Max.Z
	return out
}

// DT is simulation step (60 Hz), matches SHAPE_ANIMATION_DT_SEC in TypeScript.
const DT = 1.0 / 60.0

// Sim is runtime shape animation state.
type Sim struct {
	Background struct {
		Mode            string
		Color           string
		BrightnessPct   float64
	}
	Shapes []simShape
}

type simShape struct {
	Kind           string
	Edge           string
	BrightnessPct  float64
	ColorMode      string
	FixedColor     string
	CurrentColor   string
	Px, Py, Pz     float64
	Radius         float64
	W, H, D        float64
	Vx, Vy, Vz     float64
	Active         bool
}

// NewRng matches makeRng in TypeScript (LCG).
func NewRng(seed uint32) func() float64 {
	s := seed
	const inv = 1.0 / 4294967296.0
	return func() float64 {
		s = 1664525*s + 1013904223
		return float64(s) * inv
	}
}

func randomHex(rng func() float64) string {
	n := int(rng() * 0x1000000)
	if n < 0 {
		n = 0
	}
	if n > 0xffffff {
		n = 0xffffff
	}
	return fmt.Sprintf("#%06x", n)
}

func randomUniform(rng func() float64, lo, hi float64) float64 {
	return lo + (hi-lo)*rng()
}

func normalizeDir(dx, dy, dz float64) (ux, uy, uz float64) {
	l := math.Hypot(dx, math.Hypot(dy, dz))
	if l <= 0 {
		return 1, 0, 0
	}
	return dx / l, dy / l, dz / l
}

func randomUnitVector(rng func() float64) (ux, uy, uz float64) {
	t := 2 * math.Pi * rng()
	u := 2*rng() - 1
	s := math.Sqrt(math.Max(0, 1-u*u))
	return s * math.Cos(t), s * math.Sin(t), u
}

func reflectSpecular(vx, vy, vz float64, hitX, hitY, hitZ bool) (float64, float64, float64) {
	nx, ny, nz := vx, vy, vz
	if hitX {
		nx = -vx
	}
	if hitY {
		ny = -vy
	}
	if hitZ {
		nz = -vz
	}
	return nx, ny, nz
}

// ParseAndInit builds simulation state from definition_json (version 1) and dimensions.
func ParseAndInit(definitionJSON string, dims Dimensions, rng func() float64) (*Sim, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(definitionJSON), &root); err != nil {
		return nil, fmt.Errorf("definition_json: %w", err)
	}
	var ver float64
	if err := json.Unmarshal(root["version"], &ver); err != nil || ver != 1 {
		return nil, fmt.Errorf("definition version must be 1")
	}
	var bg map[string]json.RawMessage
	if err := json.Unmarshal(root["background"], &bg); err != nil {
		return nil, fmt.Errorf("background: %w", err)
	}
	var bgMode string
	_ = json.Unmarshal(bg["mode"], &bgMode)

	sim := &Sim{}
	switch bgMode {
	case "lights_on":
		var col string
		var br float64
		_ = json.Unmarshal(bg["color"], &col)
		_ = json.Unmarshal(bg["brightness_pct"], &br)
		sim.Background.Mode = "lights_on"
		sim.Background.Color = col
		sim.Background.BrightnessPct = br
	case "lights_off":
		sim.Background.Mode = "lights_off"
		sim.Background.Color = "#000000"
		sim.Background.BrightnessPct = 0
	default:
		return nil, fmt.Errorf("invalid background.mode")
	}

	var shapesRaw []json.RawMessage
	if err := json.Unmarshal(root["shapes"], &shapesRaw); err != nil {
		return nil, fmt.Errorf("shapes: %w", err)
	}

	minX, minY, minZ := dims.Min.X, dims.Min.Y, dims.Min.Z
	maxX, maxY, maxZ := dims.Max.X, dims.Max.Y, dims.Max.Z

	for _, shRaw := range shapesRaw {
		var sh map[string]json.RawMessage
		if err := json.Unmarshal(shRaw, &sh); err != nil {
			return nil, err
		}
		var kind, edge string
		_ = json.Unmarshal(sh["kind"], &kind)
		_ = json.Unmarshal(sh["edge_behavior"], &edge)
		var br float64
		_ = json.Unmarshal(sh["brightness_pct"], &br)

		var colObj map[string]json.RawMessage
		_ = json.Unmarshal(sh["color"], &colObj)
		var cmode string
		_ = json.Unmarshal(colObj["mode"], &cmode)
		colorMode := "fixed"
		if cmode == "random" {
			colorMode = "random"
		}
		fixedColor := "#ffffff"
		if colorMode == "fixed" {
			var fc string
			_ = json.Unmarshal(colObj["color"], &fc)
			if fc != "" {
				fixedColor = fc
			}
		}
		curColor := fixedColor
		if colorMode == "random" {
			curColor = randomHex(rng)
		}

		var sz map[string]json.RawMessage
		_ = json.Unmarshal(sh["size"], &sz)
		var szMode string
		_ = json.Unmarshal(sz["mode"], &szMode)
		radius, w, h, d := 0.1, 0.2, 0.2, 0.2
		if szMode == "fixed" {
			if kind == "sphere" {
				_ = json.Unmarshal(sz["radius_m"], &radius)
			} else {
				_ = json.Unmarshal(sz["width_m"], &w)
				_ = json.Unmarshal(sz["height_m"], &h)
				_ = json.Unmarshal(sz["depth_m"], &d)
			}
		} else {
			if kind == "sphere" {
				var rmin, rmax float64
				_ = json.Unmarshal(sz["radius_min_m"], &rmin)
				_ = json.Unmarshal(sz["radius_max_m"], &rmax)
				radius = randomUniform(rng, rmin, rmax)
			} else {
				var wmin, wmax, hmin, hmax, dmin, dmax float64
				_ = json.Unmarshal(sz["width_min_m"], &wmin)
				_ = json.Unmarshal(sz["width_max_m"], &wmax)
				_ = json.Unmarshal(sz["height_min_m"], &hmin)
				_ = json.Unmarshal(sz["height_max_m"], &hmax)
				_ = json.Unmarshal(sz["depth_min_m"], &dmin)
				_ = json.Unmarshal(sz["depth_max_m"], &dmax)
				w = randomUniform(rng, wmin, wmax)
				h = randomUniform(rng, hmin, hmax)
				d = randomUniform(rng, dmin, dmax)
			}
		}

		var mot map[string]json.RawMessage
		_ = json.Unmarshal(sh["motion"], &mot)
		var dir map[string]json.RawMessage
		_ = json.Unmarshal(mot["direction"], &dir)
		var dx, dy, dz float64
		_ = json.Unmarshal(dir["dx"], &dx)
		_ = json.Unmarshal(dir["dy"], &dy)
		_ = json.Unmarshal(dir["dz"], &dz)
		ux, uy, uz := normalizeDir(dx, dy, dz)
		var sp map[string]json.RawMessage
		_ = json.Unmarshal(mot["speed"], &sp)
		var spMode string
		_ = json.Unmarshal(sp["mode"], &spMode)
		speed := 0.1
		if spMode == "fixed" {
			_ = json.Unmarshal(sp["m_s"], &speed)
		} else {
			var smin, smax float64
			_ = json.Unmarshal(sp["min_m_s"], &smin)
			_ = json.Unmarshal(sp["max_m_s"], &smax)
			speed = randomUniform(rng, smin, smax)
		}

		px, py, pz := 0.0, 0.0, 0.0
		var pl map[string]json.RawMessage
		_ = json.Unmarshal(sh["placement"], &pl)
		var plMode string
		_ = json.Unmarshal(pl["mode"], &plMode)
		if plMode == "fixed" {
			if kind == "sphere" {
				var c map[string]float64
				_ = json.Unmarshal(pl["center_m"], &c)
				px, py, pz = c["x"], c["y"], c["z"]
			} else {
				var c map[string]float64
				_ = json.Unmarshal(pl["min_corner_m"], &c)
				px, py, pz = c["x"], c["y"], c["z"]
			}
		} else {
			var face string
			_ = json.Unmarshal(pl["face"], &face)
			if kind == "sphere" {
				switch face {
				case "left":
					px = minX + radius + 1e-6
					py = randomUniform(rng, minY+radius, maxY-radius)
					pz = randomUniform(rng, minZ+radius, maxZ-radius)
				case "right":
					px = maxX - radius - 1e-6
					py = randomUniform(rng, minY+radius, maxY-radius)
					pz = randomUniform(rng, minZ+radius, maxZ-radius)
				case "bottom":
					py = minY + radius + 1e-6
					px = randomUniform(rng, minX+radius, maxX-radius)
					pz = randomUniform(rng, minZ+radius, maxZ-radius)
				case "top":
					py = maxY - radius - 1e-6
					px = randomUniform(rng, minX+radius, maxX-radius)
					pz = randomUniform(rng, minZ+radius, maxZ-radius)
				case "back":
					pz = minZ + radius + 1e-6
					px = randomUniform(rng, minX+radius, maxX-radius)
					py = randomUniform(rng, minY+radius, maxY-radius)
				default: // front
					pz = maxZ - radius - 1e-6
					px = randomUniform(rng, minX+radius, maxX-radius)
					py = randomUniform(rng, minY+radius, maxY-radius)
				}
			} else {
				switch face {
				case "left":
					px = minX
					py = randomUniform(rng, minY, math.Max(minY, maxY-h))
					pz = randomUniform(rng, minZ, math.Max(minZ, maxZ-d))
				case "right":
					px = math.Max(minX, maxX-w)
					py = randomUniform(rng, minY, math.Max(minY, maxY-h))
					pz = randomUniform(rng, minZ, math.Max(minZ, maxZ-d))
				case "bottom":
					py = minY
					px = randomUniform(rng, minX, math.Max(minX, maxX-w))
					pz = randomUniform(rng, minZ, math.Max(minZ, maxZ-d))
				case "top":
					py = math.Max(minY, maxY-h)
					px = randomUniform(rng, minX, math.Max(minX, maxX-w))
					pz = randomUniform(rng, minZ, math.Max(minZ, maxZ-d))
				case "back":
					pz = minZ
					px = randomUniform(rng, minX, math.Max(minX, maxX-w))
					py = randomUniform(rng, minY, math.Max(minY, maxY-h))
				default:
					pz = math.Max(minZ, maxZ-d)
					px = randomUniform(rng, minX, math.Max(minX, maxX-w))
					py = randomUniform(rng, minY, math.Max(minY, maxY-h))
				}
			}
		}

		sim.Shapes = append(sim.Shapes, simShape{
			Kind: kind, Edge: edge, BrightnessPct: br,
			ColorMode: colorMode, FixedColor: fixedColor, CurrentColor: curColor,
			Px: px, Py: py, Pz: pz, Radius: radius, W: w, H: h, D: d,
			Vx: ux * speed, Vy: uy * speed, Vz: uz * speed,
			Active: true,
		})
	}
	return sim, nil
}

func integrateShape(s *simShape, minX, minY, minZ, maxX, maxY, maxZ float64, rng func() float64) {
	if !s.Active {
		return
	}
	px := s.Px + s.Vx*DT
	py := s.Py + s.Vy*DT
	pz := s.Pz + s.Vz*DT
	vx0, vy0, vz0 := s.Vx, s.Vy, s.Vz

	spanX := math.Max(maxX-minX, 1e-9)
	spanY := math.Max(maxY-minY, 1e-9)
	spanZ := math.Max(maxZ-minZ, 1e-9)

	hitX, hitY, hitZ := false, false, false
	if s.Kind == "sphere" {
		r := s.Radius
		if px-r < minX || px+r > maxX {
			hitX = true
		}
		if py-r < minY || py+r > maxY {
			hitY = true
		}
		if pz-r < minZ || pz+r > maxZ {
			hitZ = true
		}
	} else {
		if px < minX || px+s.W > maxX {
			hitX = true
		}
		if py < minY || py+s.H > maxY {
			hitY = true
		}
		if pz < minZ || pz+s.D > maxZ {
			hitZ = true
		}
	}
	violated := hitX || hitY || hitZ
	if !violated {
		s.Px, s.Py, s.Pz = px, py, pz
		return
	}
	switch s.Edge {
	case "stop":
		s.Active = false
	case "wrap":
		if s.Kind == "sphere" {
			r := s.Radius
			for px-r < minX {
				px += spanX
			}
			for px+r > maxX {
				px -= spanX
			}
			for py-r < minY {
				py += spanY
			}
			for py+r > maxY {
				py -= spanY
			}
			for pz-r < minZ {
				pz += spanZ
			}
			for pz+r > maxZ {
				pz -= spanZ
			}
		} else {
			for px < minX {
				px += spanX
			}
			for px+s.W > maxX {
				px -= spanX
			}
			for py < minY {
				py += spanY
			}
			for py+s.H > maxY {
				py -= spanY
			}
			for pz < minZ {
				pz += spanZ
			}
			for pz+s.D > maxZ {
				pz -= spanZ
			}
		}
		s.Px, s.Py, s.Pz = px, py, pz
	case "deflect_random":
		sp := math.Hypot(vx0, math.Hypot(vy0, vz0))
		ux, uy, uz := randomUnitVector(rng)
		s.Vx, s.Vy, s.Vz = ux*sp, uy*sp, uz*sp
	case "deflect_specular":
		nx, ny, nz := reflectSpecular(vx0, vy0, vz0, hitX, hitY, hitZ)
		s.Vx, s.Vy, s.Vz = nx, ny, nz
	}
}

// Tick advances all shapes one step. Returns true if every shape has stopped.
func Tick(sim *Sim, dims Dimensions, rng func() float64) bool {
	minX, minY, minZ := dims.Min.X, dims.Min.Y, dims.Min.Z
	maxX, maxY, maxZ := dims.Max.X, dims.Max.Y, dims.Max.Z
	for i := range sim.Shapes {
		integrateShape(&sim.Shapes[i], minX, minY, minZ, maxX, maxY, maxZ, rng)
	}
	for _, s := range sim.Shapes {
		if s.Active {
			return false
		}
	}
	return true
}

func lightInShape(L store.SceneLightFlat, s *simShape) bool {
	if !s.Active {
		return false
	}
	sx, sy, sz := L.Sx, L.Sy, L.Sz
	if s.Kind == "sphere" {
		d := math.Hypot(sx-s.Px, math.Hypot(sy-s.Py, sz-s.Pz))
		return d <= s.Radius+1e-9
	}
	return sx >= s.Px-1e-9 && sx <= s.Px+s.W+1e-9 &&
		sy >= s.Py-1e-9 && sy <= s.Py+s.H+1e-9 &&
		sz >= s.Pz-1e-9 && sz <= s.Pz+s.D+1e-9
}

// BuildBatchUpdates mirrors buildBatchUpdatesFromSim in TypeScript.
func BuildBatchUpdates(sim *Sim, lights []store.SceneLightFlat) []store.SceneBatchLightUpdate {
	out := make([]store.SceneBatchLightUpdate, 0, len(lights))
	for _, L := range lights {
		winner := -1
		for i := range sim.Shapes {
			if lightInShape(L, &sim.Shapes[i]) {
				winner = i
				break
			}
		}
		if winner >= 0 {
			sh := &sim.Shapes[winner]
			on := true
			c := sh.CurrentColor
			br := sh.BrightnessPct
			out = append(out, store.SceneBatchLightUpdate{
				ModelID: L.ModelID,
				LightID: L.LightID,
				Patch: store.LightStatePatch{
					On:            &on,
					Color:         &c,
					BrightnessPct: &br,
				},
			})
		} else if sim.Background.Mode == "lights_on" {
			on := true
			c := sim.Background.Color
			br := sim.Background.BrightnessPct
			out = append(out, store.SceneBatchLightUpdate{
				ModelID: L.ModelID,
				LightID: L.LightID,
				Patch: store.LightStatePatch{
					On:            &on,
					Color:         &c,
					BrightnessPct: &br,
				},
			})
		} else {
			on := false
			c := "#ffffff"
			br := 100.0
			out = append(out, store.SceneBatchLightUpdate{
				ModelID: L.ModelID,
				LightID: L.LightID,
				Patch: store.LightStatePatch{
					On:            &on,
					Color:         &c,
					BrightnessPct: &br,
				},
			})
		}
	}
	return out
}
