package devices

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/dlm/backend/internal/store"
)

func rgbFromLogical(on bool, hexColor string, brightnessPct float64) (r, g, b int) {
	if !on {
		return 0, 0, 0
	}
	hexColor = strings.TrimSpace(strings.ToLower(hexColor))
	if !strings.HasPrefix(hexColor, "#") || len(hexColor) != 7 {
		return 0, 0, 0
	}
	var rr, gg, bb int
	_, err := fmt.Sscanf(hexColor, "#%02x%02x%02x", &rr, &gg, &bb)
	if err != nil {
		return 0, 0, 0
	}
	br := brightnessPct
	if br < 0 {
		br = 0
	}
	if br > 100 {
		br = 100
	}
	scale := br / 100.0
	return int(float64(rr) * scale + 0.5), int(float64(gg) * scale + 0.5), int(float64(bb) * scale + 0.5)
}

// buildWLEDState builds a /json/state body for per-LED colours on segment 0 (1:1 idx → LED).
func buildWLEDState(states []store.LightStateDTO) map[string]any {
	leds := make([][]int, 0, len(states))
	for _, s := range states {
		r, g, b := rgbFromLogical(s.On, s.Color, s.BrightnessPct)
		leds = append(leds, []int{s.ID, r, g, b})
	}
	return map[string]any{
		"on":  true,
		"bri": 255,
		"seg": []map[string]any{
			{
				"id": 0,
				"i":  leds,
			},
		},
	}
}

func postJSONState(ctx context.Context, client *http.Client, baseURL, password string, body map[string]any) error {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/json/state"
	if password != "" {
		q := url.Values{}
		q.Set("PW", password)
		u = u + "?" + q.Encode()
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		slurp, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("wled http %d: %s", res.StatusCode, strings.TrimSpace(string(slurp)))
	}
	return nil
}
