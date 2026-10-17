package devices

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/dlm/backend/internal/store"
)

func captureStateBody(t *testing.T, drive func(baseURL string) error) map[string]any {
	t.Helper()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/state" {
			t.Errorf("path: got %s want /json/state", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	if err := drive(srv.URL); err != nil {
		t.Fatalf("drive: %v", err)
	}
	return got
}

func segmentColors(t *testing.T, body map[string]any) [][]int {
	t.Helper()
	seg, ok := body["seg"].([]any)
	if !ok || len(seg) != 1 {
		t.Fatalf("seg: %#v", body["seg"])
	}
	item, ok := seg[0].(map[string]any)
	if !ok {
		t.Fatalf("seg[0]: %#v", seg[0])
	}
	raw, ok := item["i"].([]any)
	if !ok {
		t.Fatalf("i: %#v", item["i"])
	}
	colors := make([][]int, len(raw))
	for i, entry := range raw {
		arr, ok := entry.([]any)
		if !ok {
			t.Fatalf("i[%d]: %#v", i, entry)
		}
		colors[i] = make([]int, len(arr))
		for j, n := range arr {
			f, ok := n.(float64)
			if !ok {
				t.Fatalf("i[%d][%d]: %#v", i, j, n)
			}
			colors[i][j] = int(f)
		}
	}
	return colors
}

func TestDriveSingleLED_onlyLitBulbIsWhite(t *testing.T) {
	t.Parallel()
	p := &Pusher{Client: deviceHTTPClient()}
	body := captureStateBody(t, func(baseURL string) error {
		return p.DriveSingleLED(context.Background(), store.Device{BaseURL: baseURL}, 1, 3)
	})
	if body["on"] != true || body["bri"] != float64(255) {
		t.Fatalf("on/bri: %#v", body)
	}
	want := [][]int{{0, 0, 0}, {255, 255, 255}, {0, 0, 0}}
	if got := segmentColors(t, body); !reflect.DeepEqual(got, want) {
		t.Fatalf("colors: got %#v want %#v", got, want)
	}
}

func TestDriveAllColor_paintsEachBulbThatColour(t *testing.T) {
	t.Parallel()
	p := &Pusher{Client: deviceHTTPClient()}
	body := captureStateBody(t, func(baseURL string) error {
		return p.DriveAllColor(context.Background(), store.Device{BaseURL: baseURL}, 2, 255, 0, 0)
	})
	want := [][]int{{255, 0, 0}, {255, 0, 0}}
	if got := segmentColors(t, body); !reflect.DeepEqual(got, want) {
		t.Fatalf("colors: got %#v want %#v", got, want)
	}
}
