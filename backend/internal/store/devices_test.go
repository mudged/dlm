package store

import (
	"context"
	"errors"
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
