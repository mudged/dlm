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
	return int(float64(rr)*scale + 0.5), int(float64(gg)*scale + 0.5), int(float64(bb)*scale + 0.5)
}

// buildWLEDState builds a /json/state body for per-LED colours on segment 0.
// Light id i is LED i. Each colour is red, green, blue only; the id is the
// position in the list, not a channel.
func buildWLEDState(states []store.LightStateDTO) map[string]any {
	n := 0
	for _, s := range states {
		if s.ID >= n {
			n = s.ID + 1
		}
	}
	colors := make([][3]int, n)
	for _, s := range states {
		if s.ID < 0 || s.ID >= n {
			continue
		}
		r, g, b := rgbFromLogical(s.On, s.Color, s.BrightnessPct)
		colors[s.ID] = [3]int{r, g, b}
	}
	return wledLEDFrame(true, 255, colors)
}

const deviceHTTPTimeout = 5 * time.Second

// deviceHTTPClient returns a shared outbound client for WLED pushes. Redirects are
// refused so a compromised device cannot SSRF via 3xx to an internal LAN host.
func deviceHTTPClient() *http.Client {
	return &http.Client{
		Timeout: deviceHTTPTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func postJSONState(ctx context.Context, client *http.Client, baseURL, password string, body map[string]any) error {
	if client == nil {
		client = deviceHTTPClient()
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
