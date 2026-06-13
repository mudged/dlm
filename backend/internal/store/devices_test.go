package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

func TestStore_DeviceCreateListAssignUnassignDelete(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}

	d, err := s.CreateDevice(ctx, DeviceCreate{Name: "strip", BaseURL: "http://wled.local"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Type != DeviceTypeWLED || d.Name != "strip" {
		t.Fatalf("device %+v", d)
	}
	if d.ModelID != nil {
		t.Fatal("expected no model_id")
	}

	list, err := s.ListDevices(ctx)
	if err != nil || len(list) != 1 || list[0].ID != d.ID {
		t.Fatalf("ListDevices: err=%v list=%+v", err, list)
	}

	got, err := s.GetDevice(ctx, d.ID)
	if err != nil || got.ID != d.ID {
		t.Fatalf("GetDevice: err=%v %+v", err, got)
	}

	if _, err := s.GetDeviceForModel(ctx, sum.ID); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("GetDeviceForModel before assign: err=%v", err)
	}

	if err := s.AssignDevice(ctx, d.ID, sum.ID); err != nil {
		t.Fatal(err)
	}
	after, err := s.GetDevice(ctx, d.ID)
	if err != nil || after.ModelID == nil || *after.ModelID != sum.ID {
		t.Fatalf("after assign %+v err=%v", after, err)
	}

	d2, err := s.CreateDevice(ctx, DeviceCreate{Name: "other", BaseURL: "http://other.local"})
	if err != nil {
		t.Fatal(err)
	}
	err = s.AssignDevice(ctx, d2.ID, sum.ID)
	if err == nil {
		t.Fatal("second device to same model must fail")
	}
	if !errors.Is(err, ErrDeviceAssignmentConflict) {
		t.Fatalf("want ErrDeviceAssignmentConflict, got %v", err)
	}

	err = s.AssignDevice(ctx, d.ID, sum.ID)
	if err == nil {
		t.Fatal("re-assign already-linked device must fail")
	}
	if !errors.Is(err, ErrDeviceAssignmentConflict) {
		t.Fatalf("want ErrDeviceAssignmentConflict, got %v", err)
	}

	if err := s.UnassignDevice(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	un, err := s.GetDevice(ctx, d.ID)
	if err != nil || un.ModelID != nil {
		t.Fatalf("after unassign %+v", un)
	}

	if err := s.AssignDevice(ctx, d2.ID, sum.ID); err != nil {
		t.Fatalf("assign second device after first unassigned: %v", err)
	}
	err = s.AssignDevice(ctx, d2.ID, sum.ID)
	if err == nil {
		t.Fatal("assign same device twice must fail")
	}
	if !errors.Is(err, ErrDeviceAssignmentConflict) {
		t.Fatalf("want ErrDeviceAssignmentConflict, got %v", err)
	}

	if err := s.DeleteDevice(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteDevice(ctx, d.ID); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestStore_CreateDevice_baseURLValidation(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	rejectCases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"file_scheme", "file:///etc/passwd"},
		{"ftp_scheme", "ftp://example.com/"},
		{"gopher_scheme", "gopher://example.com:70/"},
		{"javascript_scheme", "javascript:alert(1)"},
		{"http_no_host", "http:"},
		{"http_only_path", "http://"},
		{"not_a_url", "::::"},
	}
	for _, tc := range rejectCases {
		t.Run("reject_"+tc.name, func(t *testing.T) {
			_, err := s.CreateDevice(ctx, DeviceCreate{Name: "x", BaseURL: tc.in})
			if !errors.Is(err, ErrInvalidBaseURL) {
				t.Fatalf("want ErrInvalidBaseURL for %q, got %v", tc.in, err)
			}
		})
	}

	acceptCases := []struct {
		name string
		in   string
		want string
	}{
		{"http_hostname", "http://example.com", "http://example.com"},
		{"https_ip_port_trailing_slash", "https://1.2.3.4:8080/", "https://1.2.3.4:8080"},
		{"http_path_no_trailing_slash", "http://wled.local/api", "http://wled.local/api"},
		{"http_path_trailing_slash", "http://wled.local/api/", "http://wled.local/api"},
	}
	for _, tc := range acceptCases {
		t.Run("accept_"+tc.name, func(t *testing.T) {
			d, err := s.CreateDevice(ctx, DeviceCreate{Name: "x-" + tc.name, BaseURL: tc.in})
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if d.BaseURL != tc.want {
				t.Fatalf("base_url want %q got %q", tc.want, d.BaseURL)
			}
		})
	}
}

func TestStore_PatchDevice_baseURLValidation(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	d, err := s.CreateDevice(ctx, DeviceCreate{Name: "patch", BaseURL: "http://wled.local"})
	if err != nil {
		t.Fatal(err)
	}

	bad := "file:///etc/passwd"
	if _, err := s.PatchDevice(ctx, d.ID, DevicePatch{BaseURL: &bad}); !errors.Is(err, ErrInvalidBaseURL) {
		t.Fatalf("want ErrInvalidBaseURL on patch, got %v", err)
	}

	good := "https://wled.local:8080/"
	out, err := s.PatchDevice(ctx, d.ID, DevicePatch{BaseURL: &good})
	if err != nil {
		t.Fatal(err)
	}
	if out.BaseURL != "https://wled.local:8080" {
		t.Fatalf("base_url normalize want %q got %q", "https://wled.local:8080", out.BaseURL)
	}
}

func TestStore_DeviceLightCount_createGetListPatch(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	d, err := s.CreateDevice(ctx, DeviceCreate{Name: "strip", BaseURL: "http://wled.local", LightCount: 50})
	if err != nil {
		t.Fatal(err)
	}
	if d.LightCount != 50 {
		t.Fatalf("create LightCount = %d want 50", d.LightCount)
	}

	got, err := s.GetDevice(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.LightCount != 50 {
		t.Fatalf("get LightCount = %d want 50", got.LightCount)
	}

	list, err := s.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].LightCount != 50 {
		t.Fatalf("list LightCount = %+v", list)
	}

	patchVal := 120
	updated, err := s.PatchDevice(ctx, d.ID, DevicePatch{LightCount: &patchVal})
	if err != nil {
		t.Fatal(err)
	}
	if updated.LightCount != 120 {
		t.Fatalf("patch LightCount = %d want 120", updated.LightCount)
	}

	unchanged, err := s.PatchDevice(ctx, d.ID, DevicePatch{Name: strPtr("renamed")})
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Name != "renamed" || unchanged.LightCount != 120 {
		t.Fatalf("patch omit light_count: %+v", unchanged)
	}
}

func TestStore_DeviceLightCount_outOfRangeRejected(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	for _, n := range []int{-1, 1001} {
		_, err := s.CreateDevice(ctx, DeviceCreate{Name: "x", BaseURL: "http://wled.local", LightCount: n})
		if err == nil {
			t.Fatalf("create light_count=%d should fail", n)
		}
		if !strings.Contains(err.Error(), "light_count") {
			t.Fatalf("create light_count=%d err = %v", n, err)
		}
	}

	d, err := s.CreateDevice(ctx, DeviceCreate{Name: "ok", BaseURL: "http://wled.local", LightCount: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{-1, 1001} {
		patchN := n
		_, err := s.PatchDevice(ctx, d.ID, DevicePatch{LightCount: &patchN})
		if err == nil {
			t.Fatalf("patch light_count=%d should fail", n)
		}
		if !strings.Contains(err.Error(), "light_count") {
			t.Fatalf("patch light_count=%d err = %v", n, err)
		}
	}
}

func TestStore_DeviceLightCount_migrationFromLegacyTable(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "legacy_devices.db")

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE devices (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		base_url TEXT NOT NULL,
		wled_password TEXT,
		model_id TEXT UNIQUE,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO devices (id, type, name, base_url, wled_password, model_id, created_at)
		VALUES ('legacy-1', 'wled', 'legacy', 'http://legacy.local', NULL, NULL, '2020-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	d, err := s.GetDevice(ctx, "legacy-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.LightCount != 0 {
		t.Fatalf("migrated LightCount = %d want 0", d.LightCount)
	}
}

func strPtr(s string) *string { return &s }

func TestStore_FactoryReset_clearsDevices(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	if err := s.SeedDefaultSamples(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	list, err := s.List(ctx)
	if err != nil || len(list) < 1 {
		t.Fatalf("models: %v", err)
	}
	mid := list[0].ID

	d, err := s.CreateDevice(ctx, DeviceCreate{Name: "x", BaseURL: "http://x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AssignDevice(ctx, d.ID, mid); err != nil {
		t.Fatal(err)
	}
	devs, err := s.ListDevices(ctx)
	if err != nil || len(devs) != 1 {
		t.Fatalf("before reset devices: %v %+v", err, devs)
	}

	if err := s.FactoryReset(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := s.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 0 {
		t.Fatalf("after factory reset want 0 devices, got %d", len(after))
	}
}
