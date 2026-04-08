package lightstate

import (
	"math"
	"strings"
)

// EquivLightStateTriple compares effective (on, color, brightness) after normalization
// (REQ-031). Color is compared case-insensitively; brightness within tiny float tolerance.
func EquivLightStateTriple(on1 bool, color1 string, br1 float64, on2 bool, color2 string, br2 float64) bool {
	if on1 != on2 {
		return false
	}
	if strings.ToLower(strings.TrimSpace(color1)) != strings.ToLower(strings.TrimSpace(color2)) {
		return false
	}
	return math.Abs(br1-br2) <= 1e-9
}
