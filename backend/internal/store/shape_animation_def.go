package store

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// ValidateShapeAnimationDefinitionJSON validates REQ-033 / architecture §3.17.2 definition_json.
func ValidateShapeAnimationDefinitionJSON(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("definition_json is required for shape_animation")
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return fmt.Errorf("definition_json: invalid JSON: %w", err)
	}
	var ver float64
	if b, ok := root["version"]; !ok {
		return fmt.Errorf("definition_json: version is required")
	} else if err := json.Unmarshal(b, &ver); err != nil || ver != 1 {
		return fmt.Errorf("definition_json: version must be 1")
	}
	bg, ok := root["background"]
	if !ok {
		return fmt.Errorf("definition_json: background is required")
	}
	var bgObj struct {
		Mode           string   `json:"mode"`
		Color          *string  `json:"color"`
		BrightnessPct  *float64 `json:"brightness_pct"`
	}
	if err := json.Unmarshal(bg, &bgObj); err != nil {
		return fmt.Errorf("definition_json.background: %w", err)
	}
	switch bgObj.Mode {
	case "lights_on":
		if bgObj.Color == nil || *bgObj.Color == "" {
			return fmt.Errorf("definition_json.background: color required when mode is lights_on")
		}
		if _, err := ValidateColor(*bgObj.Color); err != nil {
			return fmt.Errorf("definition_json.background.color: %w", err)
		}
		if bgObj.BrightnessPct == nil {
			return fmt.Errorf("definition_json.background: brightness_pct required when mode is lights_on")
		}
		if err := ValidateBrightnessPct(*bgObj.BrightnessPct); err != nil {
			return fmt.Errorf("definition_json.background.brightness_pct: %w", err)
		}
	case "lights_off":
	default:
		return fmt.Errorf("definition_json.background.mode must be lights_on or lights_off")
	}

	shapesRaw, ok := root["shapes"]
	if !ok {
		return fmt.Errorf("definition_json: shapes is required")
	}
	var shapes []json.RawMessage
	if err := json.Unmarshal(shapesRaw, &shapes); err != nil {
		return fmt.Errorf("definition_json.shapes: %w", err)
	}
	if len(shapes) == 0 || len(shapes) > 20 {
		return fmt.Errorf("definition_json.shapes: need between 1 and 20 shapes")
	}
	for i, sh := range shapes {
		if err := validateShapeElement(i, sh); err != nil {
			return err
		}
	}
	return nil
}

func validateShapeElement(idx int, raw json.RawMessage) error {
	var o map[string]json.RawMessage
	if err := json.Unmarshal(raw, &o); err != nil {
		return fmt.Errorf("definition_json.shapes[%d]: %w", idx, err)
	}
	var kind string
	if b, ok := o["kind"]; ok {
		_ = json.Unmarshal(b, &kind)
	}
	if kind != "sphere" && kind != "cuboid" {
		return fmt.Errorf("definition_json.shapes[%d].kind must be sphere or cuboid", idx)
	}
	if err := validateSize(idx, o["size"], kind); err != nil {
		return err
	}
	if err := validateColorSpec(idx, o["color"]); err != nil {
		return err
	}
	var br float64
	if b, ok := o["brightness_pct"]; !ok {
		return fmt.Errorf("definition_json.shapes[%d]: brightness_pct is required", idx)
	} else if err := json.Unmarshal(b, &br); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].brightness_pct: %w", idx, err)
	} else if err := ValidateBrightnessPct(br); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].brightness_pct: %w", idx, err)
	}
	if err := validatePlacement(idx, o["placement"], kind); err != nil {
		return err
	}
	if err := validateMotion(idx, o["motion"]); err != nil {
		return err
	}
	var edge string
	if b, ok := o["edge_behavior"]; !ok {
		return fmt.Errorf("definition_json.shapes[%d]: edge_behavior is required", idx)
	} else if err := json.Unmarshal(b, &edge); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].edge_behavior: %w", idx, err)
	}
	switch edge {
	case "wrap", "stop", "deflect_random", "deflect_specular":
	default:
		return fmt.Errorf("definition_json.shapes[%d].edge_behavior: invalid value", idx)
	}
	return nil
}

func validateSize(idx int, raw json.RawMessage, kind string) error {
	if len(raw) == 0 {
		return fmt.Errorf("definition_json.shapes[%d]: size is required", idx)
	}
	var sz struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(raw, &sz); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].size: %w", idx, err)
	}
	switch sz.Mode {
	case "fixed":
		if kind == "sphere" {
			var s struct {
				RadiusM *float64 `json:"radius_m"`
			}
			if err := json.Unmarshal(raw, &s); err != nil {
				return err
			}
			if s.RadiusM == nil || !positiveFinite(*s.RadiusM) {
				return fmt.Errorf("definition_json.shapes[%d].size: sphere needs positive finite radius_m", idx)
			}
		} else {
			var c struct {
				WidthM  *float64 `json:"width_m"`
				HeightM *float64 `json:"height_m"`
				DepthM  *float64 `json:"depth_m"`
			}
			if err := json.Unmarshal(raw, &c); err != nil {
				return err
			}
			if c.WidthM == nil || c.HeightM == nil || c.DepthM == nil ||
				!positiveFinite(*c.WidthM) || !positiveFinite(*c.HeightM) || !positiveFinite(*c.DepthM) {
				return fmt.Errorf("definition_json.shapes[%d].size: cuboid needs positive width_m height_m depth_m", idx)
			}
		}
	case "random_uniform":
		if kind == "sphere" {
			var s struct {
				RadiusMinM *float64 `json:"radius_min_m"`
				RadiusMaxM *float64 `json:"radius_max_m"`
			}
			if err := json.Unmarshal(raw, &s); err != nil {
				return err
			}
			if s.RadiusMinM == nil || s.RadiusMaxM == nil ||
				!positiveFinite(*s.RadiusMinM) || !positiveFinite(*s.RadiusMaxM) ||
				*s.RadiusMinM > *s.RadiusMaxM {
				return fmt.Errorf("definition_json.shapes[%d].size: invalid sphere random bounds", idx)
			}
		} else {
			var c struct {
				WidthMinM  *float64 `json:"width_min_m"`
				WidthMaxM  *float64 `json:"width_max_m"`
				HeightMinM *float64 `json:"height_min_m"`
				HeightMaxM *float64 `json:"height_max_m"`
				DepthMinM  *float64 `json:"depth_min_m"`
				DepthMaxM  *float64 `json:"depth_max_m"`
			}
			if err := json.Unmarshal(raw, &c); err != nil {
				return err
			}
			if c.WidthMinM == nil || c.WidthMaxM == nil || c.HeightMinM == nil || c.HeightMaxM == nil ||
				c.DepthMinM == nil || c.DepthMaxM == nil {
				return fmt.Errorf("definition_json.shapes[%d].size: cuboid random needs all min/max pairs", idx)
			}
			if !positiveFinite(*c.WidthMinM) || !positiveFinite(*c.WidthMaxM) || *c.WidthMinM > *c.WidthMaxM ||
				!positiveFinite(*c.HeightMinM) || !positiveFinite(*c.HeightMaxM) || *c.HeightMinM > *c.HeightMaxM ||
				!positiveFinite(*c.DepthMinM) || !positiveFinite(*c.DepthMaxM) || *c.DepthMinM > *c.DepthMaxM {
				return fmt.Errorf("definition_json.shapes[%d].size: invalid cuboid random bounds", idx)
			}
		}
	default:
		return fmt.Errorf("definition_json.shapes[%d].size.mode must be fixed or random_uniform", idx)
	}
	return nil
}

func validateColorSpec(idx int, raw json.RawMessage) error {
	if len(raw) == 0 {
		return fmt.Errorf("definition_json.shapes[%d]: color is required", idx)
	}
	var c struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].color: %w", idx, err)
	}
	switch c.Mode {
	case "random":
	case "fixed":
		var f struct {
			Color string `json:"color"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			return err
		}
		if _, err := ValidateColor(f.Color); err != nil {
			return fmt.Errorf("definition_json.shapes[%d].color: %w", idx, err)
		}
	default:
		return fmt.Errorf("definition_json.shapes[%d].color.mode must be fixed or random", idx)
	}
	return nil
}

func validatePlacement(idx int, raw json.RawMessage, kind string) error {
	if len(raw) == 0 {
		return fmt.Errorf("definition_json.shapes[%d]: placement is required", idx)
	}
	var p struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].placement: %w", idx, err)
	}
	switch p.Mode {
	case "fixed":
		var fm map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fm); err != nil {
			return err
		}
		if kind == "sphere" {
			cm, ok := fm["center_m"]
			if !ok {
				return fmt.Errorf("definition_json.shapes[%d].placement: center_m required for sphere", idx)
			}
			var pt struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
				Z float64 `json:"z"`
			}
			if err := json.Unmarshal(cm, &pt); err != nil {
				return fmt.Errorf("definition_json.shapes[%d].placement.center_m: %w", idx, err)
			}
			if !finite(pt.X) || !finite(pt.Y) || !finite(pt.Z) {
				return fmt.Errorf("definition_json.shapes[%d].placement.center_m: non-finite", idx)
			}
		} else {
			mc, ok := fm["min_corner_m"]
			if !ok {
				return fmt.Errorf("definition_json.shapes[%d].placement: min_corner_m required for cuboid", idx)
			}
			var pt struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
				Z float64 `json:"z"`
			}
			if err := json.Unmarshal(mc, &pt); err != nil {
				return fmt.Errorf("definition_json.shapes[%d].placement.min_corner_m: %w", idx, err)
			}
			if !finite(pt.X) || !finite(pt.Y) || !finite(pt.Z) {
				return fmt.Errorf("definition_json.shapes[%d].placement.min_corner_m: non-finite", idx)
			}
		}
	case "random_face":
		var rf struct {
			Face string `json:"face"`
		}
		if err := json.Unmarshal(raw, &rf); err != nil {
			return err
		}
		switch rf.Face {
		case "top", "bottom", "left", "right", "back", "front":
		default:
			return fmt.Errorf("definition_json.shapes[%d].placement.face: invalid", idx)
		}
	default:
		return fmt.Errorf("definition_json.shapes[%d].placement.mode must be fixed or random_face", idx)
	}
	return nil
}

func validateMotion(idx int, raw json.RawMessage) error {
	if len(raw) == 0 {
		return fmt.Errorf("definition_json.shapes[%d]: motion is required", idx)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].motion: %w", idx, err)
	}
	dirRaw, ok := m["direction"]
	if !ok {
		return fmt.Errorf("definition_json.shapes[%d].motion: direction is required", idx)
	}
	var d struct {
		Dx float64 `json:"dx"`
		Dy float64 `json:"dy"`
		Dz float64 `json:"dz"`
	}
	if err := json.Unmarshal(dirRaw, &d); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].motion.direction: %w", idx, err)
	}
	if !finite(d.Dx) || !finite(d.Dy) || !finite(d.Dz) {
		return fmt.Errorf("definition_json.shapes[%d].motion.direction: non-finite", idx)
	}
	if d.Dx == 0 && d.Dy == 0 && d.Dz == 0 {
		return fmt.Errorf("definition_json.shapes[%d].motion.direction: cannot be all zero", idx)
	}
	spRaw, ok := m["speed"]
	if !ok {
		return fmt.Errorf("definition_json.shapes[%d].motion: speed is required", idx)
	}
	var sp struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(spRaw, &sp); err != nil {
		return fmt.Errorf("definition_json.shapes[%d].motion.speed: %w", idx, err)
	}
	switch sp.Mode {
	case "fixed":
		var f struct {
			MS float64 `json:"m_s"`
		}
		if err := json.Unmarshal(spRaw, &f); err != nil {
			return err
		}
		if !positiveFinite(f.MS) {
			return fmt.Errorf("definition_json.shapes[%d].motion.speed: m_s must be positive finite", idx)
		}
	case "random_uniform":
		var r struct {
			MinMS float64 `json:"min_m_s"`
			MaxMS float64 `json:"max_m_s"`
		}
		if err := json.Unmarshal(spRaw, &r); err != nil {
			return err
		}
		if !positiveFinite(r.MinMS) || !positiveFinite(r.MaxMS) || r.MinMS > r.MaxMS {
			return fmt.Errorf("definition_json.shapes[%d].motion.speed: invalid random m/s bounds", idx)
		}
	default:
		return fmt.Errorf("definition_json.shapes[%d].motion.speed.mode must be fixed or random_uniform", idx)
	}
	return nil
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func positiveFinite(v float64) bool {
	return finite(v) && v > 0
}
