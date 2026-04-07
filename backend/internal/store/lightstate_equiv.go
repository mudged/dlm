package store

import (
	"math"
	"strings"
)

// equivBrightness compares brightness_pct per architecture §3.19 (REQ-031).
func equivBrightness(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

// EquivLightStateTriple compares effective (on, color, brightness) after normalization.
// Color strings must already be valid stored/API form (lowercase hex from ValidateColor).
func EquivLightStateTriple(on1 bool, color1 string, br1 float64, on2 bool, color2 string, br2 float64) bool {
	if on1 != on2 {
		return false
	}
	if strings.ToLower(strings.TrimSpace(color1)) != strings.ToLower(strings.TrimSpace(color2)) {
		return false
	}
	return equivBrightness(br1, br2)
}
