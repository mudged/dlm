package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/store"
)

func postModel(t *testing.T, srv *httptest.Server, name, csv string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("name", name); err != nil {
		t.Fatal(err)
	}
	part, err := w.CreateFormFile("file", "model.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(csv)); err != nil {
		t.Fatal(err)
	}
	ct := w.FormDataContentType()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(srv.URL+"/api/v1/models", ct, &buf)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestModels_createValidCSV(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n"
	res := postModel(t, srv, "m1", csv)
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	if sum.Name != "m1" || sum.LightCount != 1 || sum.ID == "" {
		t.Fatalf("sum %+v", sum)
	}
}

func TestModels_rejectBadIds(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n2,1,1,1\n"
	res := postModel(t, srv, "bad", csv)
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	if !strings.Contains(strings.ToLower(string(b)), "sequential") {
		t.Fatalf("body = %s", b)
	}
}

func TestModels_rejectWrongHeader(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	csv := "idx,x,y,z\n0,0,0,0\n"
	res := postModel(t, srv, "h", csv)
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestModels_duplicateNameConflict(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n"
	res1 := postModel(t, srv, "dup", csv)
	res1.Body.Close()
	if res1.StatusCode != http.StatusCreated {
		t.Fatalf("first status = %d", res1.StatusCode)
	}
	res2 := postModel(t, srv, "dup", csv)
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("second status = %d", res2.StatusCode)
	}
}

func TestModels_listGetDelete(t *testing.T) {
	st := testStore(t)
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st))
	t.Cleanup(srv.Close)

	csv := "id,x,y,z\n0,0,0,0\n"
	res := postModel(t, srv, "listed", csv)
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	res, err := http.Get(srv.URL + "/api/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var list []store.Summary
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d", len(list))
	}

	res, err = http.Get(srv.URL + "/api/v1/models/" + sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", res.StatusCode)
	}

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/models/"+sum.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", res.StatusCode)
	}

	res, err = http.Get(srv.URL + "/api/v1/models/" + sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete = %d", res.StatusCode)
	}
}
