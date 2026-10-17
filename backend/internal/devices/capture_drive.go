package devices

import (
	"context"

	"example.com/dlm/backend/internal/store"
)

// wledLEDFrame is one WLED state write for segment 0.
//
// Each entry of "i" is an RGB colour. WLED paints those colours onto successive
// LEDs. A 4-number entry is RGBW, so the LED index must not sit in front of
// the colour: that index would be sent as the red channel and every "off" bulb
// would glow red.
func wledLEDFrame(on bool, bri int, colors [][3]int) map[string]any {
	leds := make([][]int, len(colors))
	for i, c := range colors {
		leds[i] = []int{c[0], c[1], c[2]}
	}
	return map[string]any{
		"on":  on,
		"bri": bri,
		"seg": []map[string]any{
			{"id": 0, "i": leds},
		},
	}
}

// DriveSingleLED sends a WLED frame with only litIdx lit (white, full brightness)
// and all other LEDs off. n is the total LED count for the device.
func (p *Pusher) DriveSingleLED(ctx context.Context, d store.Device, litIdx, n int) error {
	colors := make([][3]int, n)
	if litIdx >= 0 && litIdx < n {
		colors[litIdx] = [3]int{255, 255, 255}
	}
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, wledLEDFrame(true, 255, colors))
}

// DriveAllColor paints all n LEDs the same RGB colour at full brightness.
func (p *Pusher) DriveAllColor(ctx context.Context, d store.Device, n, r, g, b int) error {
	colors := make([][3]int, n)
	for i := range colors {
		colors[i] = [3]int{r, g, b}
	}
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, wledLEDFrame(true, 255, colors))
}

// DriveAllOff sends a WLED frame with all n LEDs off.
func (p *Pusher) DriveAllOff(ctx context.Context, d store.Device, n int) error {
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, wledLEDFrame(false, 0, make([][3]int, n)))
}
