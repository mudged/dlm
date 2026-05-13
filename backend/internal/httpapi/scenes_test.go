package httpapi

import (
	"encoding/json"
	"io"
	"math"
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
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
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
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
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
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
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

func TestScenes_spatialEndpointsDimensionsAndQueries(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m-spatial", "id,x,y,z\n0,0,0,0\n1,1,1,1\n2,2,2,2\n")
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("create model %d %s", res.StatusCode, b)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	createSceneBody := `{"name":"s-spatial","models":[{"model_id":"` + sum.ID + `"}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(createSceneBody))
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

	patchPlacementBody := `{"offset_x":10,"offset_y":0,"offset_z":0}`
	req, _ = http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/"+sc.ID+"/models/"+sum.ID, strings.NewReader(patchPlacementBody))
	req.Header.Set("Content-Type", "application/json")
	pr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Body.Close()
	if pr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(pr.Body)
		t.Fatalf("patch placement %d %s", pr.StatusCode, b)
	}

	dr, err := http.Get(srv.URL + "/api/v1/scenes/" + sc.ID + "/dimensions")
	if err != nil {
		t.Fatal(err)
	}
	defer dr.Body.Close()
	if dr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(dr.Body)
		t.Fatalf("dimensions status %d %s", dr.StatusCode, b)
	}
	var dims struct {
		Origin struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"origin"`
		Size struct {
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
			Depth  float64 `json:"depth"`
		} `json:"size"`
		Max struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"max"`
		MarginM float64 `json:"margin_m"`
	}
	if err := json.NewDecoder(dr.Body).Decode(&dims); err != nil {
		t.Fatal(err)
	}
	// AABB of lights (10..12, 0..2, 0..2) padded by the scene's persisted margin_m
	// (default 0.3 per REQ-015 BR 12 / REQ-034 rule 3). Y/Z min clamp at 0.
	const m = 0.3
	if !approxEq(dims.Origin.X, 10-m) || dims.Origin.Y != 0 || dims.Origin.Z != 0 {
		t.Fatalf("origin = %+v", dims.Origin)
	}
	if !approxEq(dims.Max.X, 12+m) || !approxEq(dims.Max.Y, 2+m) || !approxEq(dims.Max.Z, 2+m) {
		t.Fatalf("max = %+v", dims.Max)
	}
	if !approxEq(dims.Size.Width, 2+2*m) || !approxEq(dims.Size.Height, 2+m) || !approxEq(dims.Size.Depth, 2+m) {
		t.Fatalf("size = %+v", dims.Size)
	}
	if dims.MarginM != m {
		t.Fatalf("margin_m = %v want %v", dims.MarginM, m)
	}

	lr, err := http.Get(srv.URL + "/api/v1/scenes/" + sc.ID + "/lights")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Body.Close()
	if lr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(lr.Body)
		t.Fatalf("scene lights status %d %s", lr.StatusCode, b)
	}
	var allLights []struct {
		ModelID string  `json:"model_id"`
		LightID int     `json:"light_id"`
		Sx      float64 `json:"sx"`
		Sy      float64 `json:"sy"`
		Sz      float64 `json:"sz"`
	}
	if err := json.NewDecoder(lr.Body).Decode(&allLights); err != nil {
		t.Fatal(err)
	}
	if len(allLights) != 3 {
		t.Fatalf("all lights len = %d", len(allLights))
	}
	if allLights[0].ModelID != sum.ID || allLights[0].LightID != 0 || allLights[0].Sx != 10 || allLights[0].Sy != 0 || allLights[0].Sz != 0 {
		t.Fatalf("first allLights = %+v", allLights[0])
	}

	cuboidBody := `{"position":{"x":10,"y":0,"z":0},"dimensions":{"width":1,"height":1,"depth":1}}`
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes/"+sc.ID+"/lights/query/cuboid", strings.NewReader(cuboidBody))
	req.Header.Set("Content-Type", "application/json")
	qr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer qr.Body.Close()
	if qr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(qr.Body)
		t.Fatalf("cuboid query status %d %s", qr.StatusCode, b)
	}
	var cuboidLights []struct {
		LightID int `json:"light_id"`
	}
	if err := json.NewDecoder(qr.Body).Decode(&cuboidLights); err != nil {
		t.Fatal(err)
	}
	if len(cuboidLights) != 2 || cuboidLights[0].LightID != 0 || cuboidLights[1].LightID != 1 {
		t.Fatalf("cuboid lights = %+v", cuboidLights)
	}

	sphereBody := `{"center":{"x":10,"y":0,"z":0},"radius":0.1}`
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes/"+sc.ID+"/lights/query/sphere", strings.NewReader(sphereBody))
	req.Header.Set("Content-Type", "application/json")
	sr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Body.Close()
	if sr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(sr.Body)
		t.Fatalf("sphere query status %d %s", sr.StatusCode, b)
	}
	var sphereLights []struct {
		LightID int `json:"light_id"`
	}
	if err := json.NewDecoder(sr.Body).Decode(&sphereLights); err != nil {
		t.Fatal(err)
	}
	if len(sphereLights) != 1 || sphereLights[0].LightID != 0 {
		t.Fatalf("sphere lights = %+v", sphereLights)
	}
}

func TestScenes_spatialBulkUpdateAndInvalidGeometry(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m-spatial-update", "id,x,y,z\n0,0,0,0\n1,1,1,1\n2,2,2,2\n")
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("create model %d %s", res.StatusCode, b)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	createSceneBody := `{"name":"s-spatial-update","models":[{"model_id":"` + sum.ID + `"}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(createSceneBody))
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

	patchPlacementBody := `{"offset_x":10,"offset_y":0,"offset_z":0}`
	req, _ = http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/"+sc.ID+"/models/"+sum.ID, strings.NewReader(patchPlacementBody))
	req.Header.Set("Content-Type", "application/json")
	pr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Body.Close()
	if pr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(pr.Body)
		t.Fatalf("patch placement %d %s", pr.StatusCode, b)
	}

	patchCuboidBody := `{"position":{"x":10,"y":0,"z":0},"dimensions":{"width":1,"height":1,"depth":1},"on":true,"color":"#abcdef","brightness_pct":12.5}`
	req, _ = http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/"+sc.ID+"/lights/state/cuboid", strings.NewReader(patchCuboidBody))
	req.Header.Set("Content-Type", "application/json")
	ur, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer ur.Body.Close()
	if ur.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(ur.Body)
		t.Fatalf("patch cuboid status %d %s", ur.StatusCode, b)
	}
	var updateOut struct {
		UpdatedCount int `json:"updated_count"`
		States       []struct {
			ModelID string  `json:"model_id"`
			ID      int     `json:"id"`
			On      bool    `json:"on"`
			Color   string  `json:"color"`
			Bright  float64 `json:"brightness_pct"`
		} `json:"states"`
	}
	if err := json.NewDecoder(ur.Body).Decode(&updateOut); err != nil {
		t.Fatal(err)
	}
	if updateOut.UpdatedCount != 2 || len(updateOut.States) != 2 {
		t.Fatalf("updateOut = %+v", updateOut)
	}
	if updateOut.States[0].ModelID != sum.ID || !updateOut.States[0].On || updateOut.States[0].Color != "#abcdef" {
		t.Fatalf("first update state = %+v", updateOut.States[0])
	}

	statesRes, err := http.Get(srv.URL + "/api/v1/models/" + sum.ID + "/lights/state")
	if err != nil {
		t.Fatal(err)
	}
	defer statesRes.Body.Close()
	if statesRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(statesRes.Body)
		t.Fatalf("list states status %d %s", statesRes.StatusCode, b)
	}
	var statesOut struct {
		States []store.LightStateDTO `json:"states"`
	}
	if err := json.NewDecoder(statesRes.Body).Decode(&statesOut); err != nil {
		t.Fatal(err)
	}
	if len(statesOut.States) != 3 {
		t.Fatalf("states len = %d", len(statesOut.States))
	}
	if !statesOut.States[0].On || !statesOut.States[1].On || statesOut.States[2].On {
		t.Fatalf("unexpected on/off state %+v", statesOut.States)
	}
	if statesOut.States[0].Color != "#abcdef" || statesOut.States[1].Color != "#abcdef" || statesOut.States[2].Color != "#ffffff" {
		t.Fatalf("unexpected colors %+v", statesOut.States)
	}

	invalidSphereBody := `{"center":{"x":10,"y":0,"z":0},"radius":-1,"on":false}`
	req, _ = http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/"+sc.ID+"/lights/state/sphere", strings.NewReader(invalidSphereBody))
	req.Header.Set("Content-Type", "application/json")
	ir, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer ir.Body.Close()
	if ir.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(ir.Body)
		t.Fatalf("invalid sphere status %d %s", ir.StatusCode, b)
	}
	var errEnv struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(ir.Body).Decode(&errEnv); err != nil {
		t.Fatal(err)
	}
	if errEnv.Error.Code != "validation_failed" {
		t.Fatalf("error code = %q", errEnv.Error.Code)
	}

	statesRes2, err := http.Get(srv.URL + "/api/v1/models/" + sum.ID + "/lights/state")
	if err != nil {
		t.Fatal(err)
	}
	defer statesRes2.Body.Close()
	if statesRes2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(statesRes2.Body)
		t.Fatalf("list states after invalid status %d %s", statesRes2.StatusCode, b)
	}
	var statesOut2 struct {
		States []store.LightStateDTO `json:"states"`
	}
	if err := json.NewDecoder(statesRes2.Body).Decode(&statesOut2); err != nil {
		t.Fatal(err)
	}
	if !statesOut2.States[0].On || !statesOut2.States[1].On || statesOut2.States[2].On {
		t.Fatalf("state changed after invalid geometry %+v", statesOut2.States)
	}
}

// approxEq is a sub-millimetre float equality used for assertions involving
// the persisted scene margin_m (REQ-015 BR 12 / REQ-034 rule 3).
func approxEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// TestScenes_patchMarginM covers PATCH /api/v1/scenes/{id} per REQ-015 BR 13 / REQ-034:
// happy path, invalid_margin_m envelope on out-of-range / non-finite, and 404 on unknown id.
func TestScenes_patchMarginM(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)

	res := postModel(t, srv, "m1", "id,x,y,z\n0,0,0,0\n1,1,1,1\n")
	defer res.Body.Close()
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}

	createBody := `{"name":"scene-margin","models":[{"model_id":"` + sum.ID + `"}]}`
	cReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/scenes", strings.NewReader(createBody))
	cReq.Header.Set("Content-Type", "application/json")
	cRes, err := http.DefaultClient.Do(cReq)
	if err != nil {
		t.Fatal(err)
	}
	defer cRes.Body.Close()
	if cRes.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(cRes.Body)
		t.Fatalf("create scene %d %s", cRes.StatusCode, b)
	}
	var sc store.SceneSummary
	if err := json.NewDecoder(cRes.Body).Decode(&sc); err != nil {
		t.Fatal(err)
	}
	if !approxEq(sc.MarginM, 0.3) {
		t.Fatalf("create margin = %v want 0.3", sc.MarginM)
	}

	doPatch := func(t *testing.T, payload string) *http.Response {
		t.Helper()
		pReq, perr := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/"+sc.ID, strings.NewReader(payload))
		if perr != nil {
			t.Fatal(perr)
		}
		pReq.Header.Set("Content-Type", "application/json")
		pRes, perr := http.DefaultClient.Do(pReq)
		if perr != nil {
			t.Fatal(perr)
		}
		return pRes
	}

	// Happy path: 0.5 m.
	okRes := doPatch(t, `{"margin_m":0.5}`)
	if okRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(okRes.Body)
		_ = okRes.Body.Close()
		t.Fatalf("patch 0.5 status %d %s", okRes.StatusCode, b)
	}
	var echo store.SceneSummary
	if err := json.NewDecoder(okRes.Body).Decode(&echo); err != nil {
		_ = okRes.Body.Close()
		t.Fatal(err)
	}
	_ = okRes.Body.Close()
	if !approxEq(echo.MarginM, 0.5) || echo.ID != sc.ID || echo.ModelCount != 1 {
		t.Fatalf("patch echo unexpected: %+v", echo)
	}

	// Persisted in subsequent GET.
	gRes, err := http.Get(srv.URL + "/api/v1/scenes/" + sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	var detail store.SceneDetail
	if err := json.NewDecoder(gRes.Body).Decode(&detail); err != nil {
		_ = gRes.Body.Close()
		t.Fatal(err)
	}
	_ = gRes.Body.Close()
	if !approxEq(detail.MarginM, 0.5) {
		t.Fatalf("GET after PATCH margin = %v", detail.MarginM)
	}

	// Invalid inputs — each returns 400 invalid_margin_m and does NOT mutate the scene.
	type apiErr struct {
		Error struct {
			Code    string  `json:"code"`
			Message string  `json:"message"`
			Details *struct {
				Min float64 `json:"min"`
				Max float64 `json:"max"`
			} `json:"details,omitempty"`
		} `json:"error"`
	}
	for _, payload := range []string{
		`{"margin_m":-0.1}`,
		`{"margin_m":5.0001}`,
		`{"margin_m":"oops"}`, // type mismatch → bad_request
	} {
		bad := doPatch(t, payload)
		if bad.StatusCode != http.StatusBadRequest {
			b, _ := io.ReadAll(bad.Body)
			_ = bad.Body.Close()
			t.Fatalf("payload %q expected 400, got %d %s", payload, bad.StatusCode, b)
		}
		var ae apiErr
		_ = json.NewDecoder(bad.Body).Decode(&ae)
		_ = bad.Body.Close()
		switch payload {
		case `{"margin_m":"oops"}`:
			if ae.Error.Code != "bad_request" {
				t.Fatalf("string payload error.code = %q want bad_request", ae.Error.Code)
			}
		default:
			if ae.Error.Code != "invalid_margin_m" || ae.Error.Details == nil || ae.Error.Details.Max != 5 {
				t.Fatalf("payload %q error = %+v", payload, ae.Error)
			}
		}
	}

	// Missing field — currently the only editable field.
	missing := doPatch(t, `{}`)
	if missing.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing margin_m expected 400, got %d", missing.StatusCode)
	}
	_ = missing.Body.Close()

	// Unknown field rejected.
	extra := doPatch(t, `{"margin_m":0.4,"name":"oops"}`)
	if extra.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown field expected 400, got %d", extra.StatusCode)
	}
	_ = extra.Body.Close()

	// 404 for unknown scene id.
	missReq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/scenes/no-such-scene", strings.NewReader(`{"margin_m":0.4}`))
	missReq.Header.Set("Content-Type", "application/json")
	missRes, err := http.DefaultClient.Do(missReq)
	if err != nil {
		t.Fatal(err)
	}
	if missRes.StatusCode != http.StatusNotFound {
		t.Fatalf("missing scene expected 404, got %d", missRes.StatusCode)
	}
	_ = missRes.Body.Close()

	// Re-verify margin still 0.5 after the failed PATCHes.
	gRes2, err := http.Get(srv.URL + "/api/v1/scenes/" + sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	var detail2 store.SceneDetail
	if err := json.NewDecoder(gRes2.Body).Decode(&detail2); err != nil {
		_ = gRes2.Body.Close()
		t.Fatal(err)
	}
	_ = gRes2.Body.Close()
	if !approxEq(detail2.MarginM, 0.5) {
		t.Fatalf("margin mutated by failed PATCH: %v", detail2.MarginM)
	}
}
