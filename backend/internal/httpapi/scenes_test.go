package httpapi

import (
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

func TestScenes_createAndGet(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m1", "id,x,y,z\n0,0,0,0\n")
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("create model %d %s", res.StatusCode, b)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	body := `{"name":"s1","models":[{"model_id":"` + sum.ID + `"}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	cr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer cr.Body.Close()
	if cr.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(cr.Body)
		t.Fatalf("create scene %d %s", cr.StatusCode, b)
	}
	var sc store.SceneSummary
	if err := json.NewDecoder(cr.Body).Decode(&sc); err != nil {
		t.Fatal(err)
	}

	gr, err := http.Get(srv.URL + "/api/v1/scenes/" + sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Body.Close()
	if gr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(gr.Body)
		t.Fatalf("get scene %d %s", gr.StatusCode, b)
	}
}

func TestModels_delete409WhenInScene(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m1", "id,x,y,z\n0,0,0,0\n")
	defer res.Body.Close()
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	body := `{"name":"s1","models":[{"model_id":"` + sum.ID + `"}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	cr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = cr.Body.Close()

	dr, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/models/"+sum.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(dr)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 409 got %d %s", resp.StatusCode, b)
	}
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Details struct {
				Scenes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"scenes"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error.Code != "model_in_scenes" || len(env.Error.Details.Scenes) != 1 {
		t.Fatalf("unexpected body %+v", env)
	}
}

func TestScenes_createRejectsClientOffsets(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m1", "id,x,y,z\n0,0,0,0\n")
	defer res.Body.Close()
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	body := `{"name":"s1","models":[{"model_id":"` + sum.ID + `","offset_x":0}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	cr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer cr.Body.Close()
	if cr.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(cr.Body)
		t.Fatalf("want 400 got %d %s", cr.StatusCode, b)
	}
}
