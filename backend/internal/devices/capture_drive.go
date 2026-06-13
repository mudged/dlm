package devices

import (
	"context"

	"example.com/dlm/backend/internal/store"
)

// DriveSingleLED sends a WLED frame with only litIdx lit (white, full brightness)
// and all other LEDs off. n is the total LED count for the device.
func (p *Pusher) DriveSingleLED(ctx context.Context, d store.Device, litIdx, n int) error {
	leds := make([][]int, n)
	for i := 0; i < n; i++ {
		if i == litIdx {
			leds[i] = []int{i, 255, 255, 255}
		} else {
			leds[i] = []int{i, 0, 0, 0}
		}
	}
	payload := map[string]any{
		"on":  true,
		"bri": 255,
		"seg": []map[string]any{
			{"id": 0, "i": leds},
		},
	}
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, payload)
}

// DriveAllOff sends a WLED frame with all n LEDs off.
func (p *Pusher) DriveAllOff(ctx context.Context, d store.Device, n int) error {
	leds := make([][]int, n)
	for i := 0; i < n; i++ {
		leds[i] = []int{i, 0, 0, 0}
	}
	payload := map[string]any{
		"on":  false,
		"bri": 0,
		"seg": []map[string]any{
			{"id": 0, "i": leds},
		},
	}
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, payload)
}
