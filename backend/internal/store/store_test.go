package store

import (
	"context"
	"path/filepath"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

func testDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_CreateListGetDelete(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 1, Y: 2, Z: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if sum.LightCount != 1 || sum.Name != "m1" {
		t.Fatalf("summary %+v", sum)
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].LightCount != 1 {
		t.Fatalf("list %+v", list)
	}

	d, err := s.Get(ctx, sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Lights) != 1 || d.Lights[0].ID != 0 || d.Lights[0].X != 1 {
		t.Fatalf("detail %+v", d)
	}
	if !d.Lights[0].On || d.Lights[0].Color != "#ffffff" || d.Lights[0].BrightnessPct != 100 {
		t.Fatalf("want default light state, got %+v", d.Lights[0])
	}

	if err := s.Delete(ctx, sum.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, sum.ID); err != ErrNotFound {
		t.Fatalf("get after delete: %v", err)
	}
}

func TestStore_DuplicateName(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	if _, err := s.Create(ctx, "same", nil); err != nil {
		t.Fatal(err)
	}
	_, err := s.Create(ctx, "same", nil)
	if err != ErrDuplicateName {
		t.Fatalf("want ErrDuplicateName, got %v", err)
	}
}

func TestStore_EmptyLights(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "empty", nil)
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Get(ctx, sum.ID)
	if err != nil || len(d.Lights) != 0 {
		t.Fatalf("detail %+v err %v", d, err)
	}
}

func TestStore_SeedDefaultSamples_idempotent(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	if err := s.SeedDefaultSamples(ctx); err != nil {
		t.Fatal(err)
	}
	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("after seed want 3 models, got %d", len(list))
	}
	if err := s.SeedDefaultSamples(ctx); err != nil {
		t.Fatal(err)
	}
	list2, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list2) != 3 {
		t.Fatalf("second seed must not duplicate, got %d", len(list2))
	}
}

func TestStore_LightStateListPatch(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "stateful", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	states, err := s.ListLightStates(ctx, sum.ID)
	if err != nil || len(states) != 2 {
		t.Fatalf("ListLightStates: %v %+v", err, states)
	}
	if !states[0].On || states[0].Color != "#ffffff" {
		t.Fatalf("state[0] %+v", states[0])
	}

	f := false
	blue := "#0000ff"
	patched, err := s.PatchLightState(ctx, sum.ID, 0, LightStatePatch{
		On:    &f,
		Color: &blue,
	})
	if err != nil {
		t.Fatal(err)
	}
	if patched.On || patched.Color != "#0000ff" {
		t.Fatalf("patched %+v", patched)
	}

	one, err := s.GetLightState(ctx, sum.ID, 1)
	if err != nil || !one.On {
		t.Fatalf("light 1 %+v err %v", one, err)
	}

	_, err = s.PatchLightState(ctx, sum.ID, 99, LightStatePatch{On: &f})
	if err != ErrInvalidLightIndex {
		t.Fatalf("want ErrInvalidLightIndex, got %v", err)
	}
}

func TestStore_BatchPatchLightStates(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "batchy", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 0, Z: 0},
		{ID: 2, X: 2, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.BatchPatchLightStates(ctx, sum.ID, nil, LightStatePatch{On: ptrBool(false)})
	if err != ErrBatchEmptyIDs {
		t.Fatalf("empty ids: want ErrBatchEmptyIDs, got %v", err)
	}

	_, err = s.BatchPatchLightStates(ctx, sum.ID, []int{0, 0}, LightStatePatch{On: ptrBool(false)})
	if err != ErrBatchDuplicateIDs {
		t.Fatalf("dup ids: want ErrBatchDuplicateIDs, got %v", err)
	}

	_, err = s.BatchPatchLightStates(ctx, sum.ID, []int{0}, LightStatePatch{})
	if err == nil {
		t.Fatal("empty patch fields: want error")
	}

	_, err = s.BatchPatchLightStates(ctx, sum.ID, []int{99}, LightStatePatch{On: ptrBool(false)})
	if err != ErrInvalidLightIndex {
		t.Fatalf("out of range: want ErrInvalidLightIndex, got %v", err)
	}

	f := false
	states, err := s.BatchPatchLightStates(ctx, sum.ID, []int{2, 0}, LightStatePatch{On: &f})
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 2 || states[0].ID != 0 || states[1].ID != 2 {
		t.Fatalf("want sorted ids 0,2 got %+v", states)
	}
	if states[0].On || states[1].On {
		t.Fatalf("expected both off %+v", states)
	}

	one, err := s.GetLightState(ctx, sum.ID, 1)
	if err != nil || !one.On {
		t.Fatalf("light 1 should be unchanged %+v", one)
	}
}

func ptrBool(b bool) *bool { return &b }
