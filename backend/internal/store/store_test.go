package store

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"example.com/dlm/backend/internal/samples"
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
	if d.Lights[0].On || d.Lights[0].Color != "#ffffff" || d.Lights[0].BrightnessPct != 100 {
		t.Fatalf("want default light state (off, white, 100%%), got %+v", d.Lights[0])
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

func TestStore_FactoryReset(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	if err := s.SeedDefaultSamples(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, "extra-user-model", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}}); err != nil {
		t.Fatal(err)
	}
	list, err := s.List(ctx)
	if err != nil || len(list) != 4 {
		t.Fatalf("want 4 models before reset, got %d err %v", len(list), err)
	}
	mid := list[0].ID
	if _, err := s.CreateScene(ctx, "scene-one", []string{mid}); err != nil {
		t.Fatal(err)
	}
	scenes, err := s.ListScenes(ctx)
	if err != nil || len(scenes) != 1 {
		t.Fatalf("want 1 scene before reset, got %d err %v", len(scenes), err)
	}

	if err := s.FactoryReset(ctx); err != nil {
		t.Fatal(err)
	}

	scenes2, err := s.ListScenes(ctx)
	if err != nil || len(scenes2) != 0 {
		t.Fatalf("want 0 scenes after reset, got %d err %v", len(scenes2), err)
	}
	list2, err := s.List(ctx)
	if err != nil || len(list2) != 3 {
		t.Fatalf("want 3 models after reset, got %d err %v", len(list2), err)
	}
	names := make([]string, 0, 3)
	for _, m := range list2 {
		names = append(names, m.Name)
	}
	slices.Sort(names)
	want := []string{samples.NameCone, samples.NameCube, samples.NameSphere}
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("model names = %v want %v", names, want)
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
	if states[0].On || states[0].Color != "#ffffff" {
		t.Fatalf("state[0] %+v", states[0])
	}

	f := false
	blue := "#0000ff"
	patched, unchanged, err := s.PatchLightState(ctx, sum.ID, 0, LightStatePatch{
		On:    &f,
		Color: &blue,
	})
	if err != nil {
		t.Fatal(err)
	}
	if unchanged {
		t.Fatal("expected first patch to write")
	}
	if patched.On || patched.Color != "#0000ff" {
		t.Fatalf("patched %+v", patched)
	}

	_, noop, err := s.PatchLightState(ctx, sum.ID, 0, LightStatePatch{Color: &blue})
	if err != nil {
		t.Fatal(err)
	}
	if !noop {
		t.Fatal("equivalent patch should skip UPDATE (REQ-031)")
	}

	one, err := s.GetLightState(ctx, sum.ID, 1)
	if err != nil || one.On {
		t.Fatalf("light 1 should stay default off %+v err %v", one, err)
	}

	_, _, err = s.PatchLightState(ctx, sum.ID, 99, LightStatePatch{On: &f})
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

	_, _, err = s.BatchPatchLightStates(ctx, sum.ID, nil, LightStatePatch{On: ptrBool(false)})
	if err != ErrBatchEmptyIDs {
		t.Fatalf("empty ids: want ErrBatchEmptyIDs, got %v", err)
	}

	_, _, err = s.BatchPatchLightStates(ctx, sum.ID, []int{0, 0}, LightStatePatch{On: ptrBool(false)})
	if err != ErrBatchDuplicateIDs {
		t.Fatalf("dup ids: want ErrBatchDuplicateIDs, got %v", err)
	}

	_, _, err = s.BatchPatchLightStates(ctx, sum.ID, []int{0}, LightStatePatch{})
	if err == nil {
		t.Fatal("empty patch fields: want error")
	}

	_, _, err = s.BatchPatchLightStates(ctx, sum.ID, []int{99}, LightStatePatch{On: ptrBool(false)})
	if err != ErrInvalidLightIndex {
		t.Fatalf("out of range: want ErrInvalidLightIndex, got %v", err)
	}

	f := false
	tOn := true
	states, allUnchanged, err := s.BatchPatchLightStates(ctx, sum.ID, []int{2, 0}, LightStatePatch{On: &tOn})
	if err != nil {
		t.Fatal(err)
	}
	if allUnchanged {
		t.Fatal("batch should have written")
	}
	if len(states) != 2 || states[0].ID != 0 || states[1].ID != 2 {
		t.Fatalf("want sorted ids 0,2 got %+v", states)
	}
	if !states[0].On || !states[1].On {
		t.Fatalf("expected both on %+v", states)
	}

	st2, _, err := s.BatchPatchLightStates(ctx, sum.ID, []int{0, 2}, LightStatePatch{On: &f})
	if err != nil {
		t.Fatal(err)
	}
	if len(st2) != 2 || st2[0].On || st2[1].On {
		t.Fatalf("want both off after batch %+v", st2)
	}
	st3, allNoop, err := s.BatchPatchLightStates(ctx, sum.ID, []int{0, 2}, LightStatePatch{On: &f})
	if err != nil {
		t.Fatal(err)
	}
	if !allNoop || len(st3) != 2 {
		t.Fatalf("idempotent batch want unchanged_all, got %+v unchanged=%v", st3, allNoop)
	}

	one, err := s.GetLightState(ctx, sum.ID, 1)
	if err != nil || one.On {
		t.Fatalf("light 1 should be unchanged (still off) %+v", one)
	}
}

func TestStore_ResetAllLightStates(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "resetme", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	tOn := true
	red := "#ff0000"
	if _, _, err := s.PatchLightState(ctx, sum.ID, 0, LightStatePatch{On: &tOn, Color: &red}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.PatchLightState(ctx, sum.ID, 1, LightStatePatch{BrightnessPct: ptrFloat(50)}); err != nil {
		t.Fatal(err)
	}

	states, _, err := s.ResetAllLightStates(ctx, sum.ID)
	if err != nil || len(states) != 2 {
		t.Fatalf("ResetAllLightStates: %v %+v", err, states)
	}
	for _, st := range states {
		if st.On || st.Color != "#ffffff" || st.BrightnessPct != 100 {
			t.Fatalf("want REQ-014 default, got %+v", st)
		}
	}
	if _, _, err := s.ResetAllLightStates(ctx, "00000000-0000-0000-0000-000000000000"); err != ErrNotFound {
		t.Fatalf("missing model: want ErrNotFound, got %v", err)
	}
}

func TestStore_ResetAllLightStates_noopWhenAlreadyDefault(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "alreadydef", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	states, unchangedAll, err := s.ResetAllLightStates(ctx, sum.ID)
	if err != nil || len(states) != 1 {
		t.Fatalf("ResetAllLightStates: %v %+v", err, states)
	}
	if !unchangedAll {
		t.Fatal("fresh model should already be at REQ-014 defaults")
	}
}

func ptrBool(b bool) *bool { return &b }

func ptrFloat(f float64) *float64 { return &f }
