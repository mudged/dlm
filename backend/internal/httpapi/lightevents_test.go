package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/store"
)

func readFirstSSEDataLine(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			return strings.TrimPrefix(line, "data: "), nil
		}
	}
}

func TestModelLightsEvents_streamAfterPatch(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n"
	res := postModel(t, srv, "sse-model", csv)
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", res.StatusCode)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	modelID := sum.ID

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/models/"+modelID+"/lights/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	esRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer esRes.Body.Close()
	if esRes.StatusCode != http.StatusOK {
		t.Fatalf("sse status %d", esRes.StatusCode)
	}
	if ct := esRes.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("content-type %q", ct)
	}

	firstJSON, err := readFirstSSEDataLine(esRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	var first struct {
		Seq uint64 `json:"seq"`
	}
	if err := json.Unmarshal([]byte(firstJSON), &first); err != nil {
		t.Fatalf("decode first: %v body %q", err, firstJSON)
	}

	patchBody := `{"on":true,"color":"#ff0000","brightness_pct":100}`
	preq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/models/"+modelID+"/lights/0/state", strings.NewReader(patchBody))
	preq.Header.Set("Content-Type", "application/json")
	pres, err := http.DefaultClient.Do(preq)
	if err != nil {
		t.Fatal(err)
	}
	defer pres.Body.Close()
	_, _ = io.Copy(io.Discard, pres.Body)
	if pres.StatusCode != http.StatusOK {
		t.Fatalf("patch status %d", pres.StatusCode)
	}

	secondJSON, err := readFirstSSEDataLine(esRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	var second struct {
		Seq uint64 `json:"seq"`
	}
	if err := json.Unmarshal([]byte(secondJSON), &second); err != nil {
		t.Fatalf("decode second: %v body %q", err, secondJSON)
	}
	if second.Seq != first.Seq+1 {
		t.Fatalf("expected seq %d got %d", first.Seq+1, second.Seq)
	}
	cancel()
}

func TestSceneLightsEvents_streamAfterScenePatch(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n"
	res := postModel(t, srv, "scene-sse", csv)
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", res.StatusCode)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	modelID := sum.ID

	body := `{"name":"S","models":[{"model_id":"` + modelID + `"}]}`
	sres, err := http.Post(srv.URL+"/api/v1/scenes", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer sres.Body.Close()
	if sres.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(sres.Body)
		t.Fatalf("scene create %d %s", sres.StatusCode, b)
	}
	var scene struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(sres.Body).Decode(&scene); err != nil {
		t.Fatal(err)
	}
	sceneID := scene.ID

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/scenes/"+sceneID+"/lights/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	esRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer esRes.Body.Close()
	if esRes.StatusCode != http.StatusOK {
		t.Fatalf("sse status %d", esRes.StatusCode)
	}

	firstJSON, err := readFirstSSEDataLine(esRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	var first struct {
		Seq uint64 `json:"seq"`
	}
	if err := json.Unmarshal([]byte(firstJSON), &first); err != nil {
		t.Fatal(err)
	}

	patch := `{"on":true,"color":"#00ff00","brightness_pct":50}`
	purl := srv.URL + "/api/v1/scenes/" + sceneID + "/lights/state/scene"
	preq, _ := http.NewRequest(http.MethodPatch, purl, strings.NewReader(patch))
	preq.Header.Set("Content-Type", "application/json")
	pres, err := http.DefaultClient.Do(preq)
	if err != nil {
		t.Fatal(err)
	}
	defer pres.Body.Close()
	_, _ = io.Copy(io.Discard, pres.Body)
	if pres.StatusCode != http.StatusOK {
		t.Fatalf("patch status %d", pres.StatusCode)
	}

	secondJSON, err := readFirstSSEDataLine(esRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	var second struct {
		Seq uint64 `json:"seq"`
	}
	if err := json.Unmarshal([]byte(secondJSON), &second); err != nil {
		t.Fatal(err)
	}
	if second.Seq != first.Seq+1 {
		t.Fatalf("expected seq %d got %d", first.Seq+1, second.Seq)
	}
	cancel()
}
