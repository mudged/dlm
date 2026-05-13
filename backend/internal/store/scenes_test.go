package store

import (
	"context"
	"errors"
	"math"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

// floatEq compares two float64s with a tight epsilon, suitable for assertions involving
// the scene margin_m fan-out (DefaultSceneBoundaryMarginM = 0.3, sub-millimetre precision).
func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestPatchSceneLightsSceneAndBatch(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "m1", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.PatchSceneLightsScene(ctx, sc.ID, LightStatePatch{
		On: ptrBool(true),
	})
	if err != nil || res.UpdatedCount != 2 {
		t.Fatalf("scene patch %+v err %v", res, err)
	}
	c0 := "#aabbcc"
	c1 := "#112233"
	br := 100.0
	on := true
	batch, err := s.PatchSceneLightsBatch(ctx, sc.ID, []SceneBatchLightUpdate{
		{ModelID: sum.ID, LightID: 0, Patch: LightStatePatch{Color: &c0, On: &on, BrightnessPct: &br}},
		{ModelID: sum.ID, LightID: 1, Patch: LightStatePatch{Color: &c1, On: &on, BrightnessPct: &br}},
	})
	if err != nil || batch.UpdatedCount != 2 {
		t.Fatalf("batch %+v err %v", batch, err)
	}
	d, err := s.Get(ctx, sum.ID)
	if err != nil || !d.Lights[0].On || d.Lights[0].Color != "#aabbcc" {
		t.Fatalf("light0 %+v err %v", d.Lights[0], err)
	}
	if d.Lights[1].Color != "#112233" {
		t.Fatalf("light1 color %q", d.Lights[1].Color)
	}

	res2, err := s.PatchSceneLightsScene(ctx, sc.ID, LightStatePatch{
		On: ptrBool(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res2.UpdatedCount != 0 || !res2.UnchangedAll || len(res2.States) != 2 {
		t.Fatalf("idempotent scene patch want 0 writes unchanged_all, got %+v", res2)
	}
}

func TestPatchSceneLightsBatchRejectsForeignLight(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	a, err := s.Create(ctx, "a", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create(ctx, "b", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{a.ID})
	if err != nil {
		t.Fatal(err)
	}
	on := true
	br := 100.0
	col := "#ff0000"
	_, err = s.PatchSceneLightsBatch(ctx, sc.ID, []SceneBatchLightUpdate{
		{ModelID: b.ID, LightID: 0, Patch: LightStatePatch{On: &on, Color: &col, BrightnessPct: &br}},
	})
	if !errors.Is(err, ErrSceneLightNotInScene) {
		t.Fatalf("want ErrSceneLightNotInScene got %v", err)
	}
}

func TestRoutinesCreateStartStopDelete(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.CreateRoutine(ctx, "fx", "d", RoutineTypePythonSceneScript, "pass", "")
	if err != nil {
		t.Fatal(err)
	}
	runID, err := s.StartRoutineRun(ctx, sc.ID, r.ID)
	if err != nil || runID == "" {
		t.Fatalf("start %+v %v", runID, err)
	}
	runs, err := s.ListRunningRoutineRunsForScene(ctx, sc.ID)
	if err != nil || len(runs) != 1 || runs[0].ID != runID {
		t.Fatalf("want one running run %+v err %v", runs, err)
	}
	if err := s.StopRoutineRun(ctx, sc.ID, runID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRoutine(ctx, r.ID); err != nil {
		t.Fatal(err)
	}
}

func TestStartRoutineRun_secondStartSameRoutine_returnsConflict(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.CreateRoutine(ctx, "fx", "d", RoutineTypePythonSceneScript, "pass", "")
	if err != nil {
		t.Fatal(err)
	}
	runID, err := s.StartRoutineRun(ctx, sc.ID, r.ID)
	if err != nil || runID == "" {
		t.Fatalf("first start: %v %q", err, runID)
	}
	_, err = s.StartRoutineRun(ctx, sc.ID, r.ID)
	var c *SceneRoutineConflictError
	if !errors.As(err, &c) || c.RunID != runID || c.RoutineID != r.ID {
		t.Fatalf("want SceneRoutineConflictError for duplicate start, got err=%v conflict=%+v", err, c)
	}
}

func TestScenes_CreateGetDelete(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}

	sc, err := s.CreateScene(ctx, "living", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	if sc.ModelCount != 1 {
		t.Fatalf("model count %d", sc.ModelCount)
	}

	list, err := s.ListScenes(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list %+v err %v", list, err)
	}

	d, err := s.GetScene(ctx, sc.ID)
	if err != nil || len(d.Items) != 1 || d.Items[0].Lights[0].Sx != 0 {
		t.Fatalf("detail %+v err %v", d, err)
	}

	if err := s.Delete(ctx, sum.ID); err == nil {
		t.Fatal("expected model delete blocked by scene")
	} else {
		var mu *ModelInUseError
		if !errors.As(err, &mu) || len(mu.Scenes) != 1 {
			t.Fatalf("want ModelInUseError, got %v", err)
		}
	}

	if err := s.DeleteScene(ctx, sc.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, sum.ID); err != nil {
		t.Fatal(err)
	}
}

func TestScenes_CreateSceneAutoOffsetsNegativeModelCoords(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: -1, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatalf("create scene: %v", err)
	}
	d, err := s.GetScene(ctx, sc.ID)
	if err != nil || len(d.Items) != 1 {
		t.Fatalf("detail %+v err %v", d, err)
	}
	L := d.Items[0].Lights[0]
	if L.Sx < -sceneCoordEps || L.Sy < -sceneCoordEps || L.Sz < -sceneCoordEps {
		t.Fatalf("expected non-negative scene coords, got sx=%v sy=%v sz=%v", L.Sx, L.Sy, L.Sz)
	}
	if d.Items[0].OffsetX != 1 {
		t.Fatalf("want first-model offset_x=1 for x=-1 light, got %d", d.Items[0].OffsetX)
	}
}

func TestScenes_LastModelRemoval(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	err = s.RemoveSceneModel(ctx, sc.ID, sum.ID)
	if err != ErrSceneLastModel {
		t.Fatalf("want ErrSceneLastModel got %v", err)
	}
}

func TestScenes_CreateScenePreservesModelOrder(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	a, err := s.Create(ctx, "alpha", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create(ctx, "beta", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	// UUID order differs from creation order; response must follow request array order (beta then alpha).
	sc, err := s.CreateScene(ctx, "ordered", []string{b.ID, a.ID})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.GetScene(ctx, sc.ID)
	if err != nil || len(d.Items) != 2 {
		t.Fatalf("detail %+v err %v", d, err)
	}
	if d.Items[0].ModelID != b.ID || d.Items[1].ModelID != a.ID {
		t.Fatalf("want order [beta, alpha], got %q then %q", d.Items[0].ModelID, d.Items[1].ModelID)
	}
}

func TestScenes_SpatialQueriesAndDimensions(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "spatial", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 1, Z: 1},
		{ID: 2, X: 2, Y: 2, Z: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "scene-spatial", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PatchSceneModelOffsets(ctx, sc.ID, sum.ID, 10, 0, 0); err != nil {
		t.Fatal(err)
	}

	dims, err := s.GetSceneDimensions(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	// AABB of lights (10..12, 0..2, 0..2) padded by the scene's persisted margin (default
	// DefaultSceneBoundaryMarginM = 0.3 per REQ-015 BR 12 / REQ-034 rule 3).
	// Y/Z min stay clamped at 0 because 0 - 0.3 < 0.
	const m = DefaultSceneBoundaryMarginM
	if !floatEq(dims.Origin.X, 10-m) || dims.Origin.Y != 0 || dims.Origin.Z != 0 {
		t.Fatalf("origin = %+v", dims.Origin)
	}
	if !floatEq(dims.Max.X, 12+m) || !floatEq(dims.Max.Y, 2+m) || !floatEq(dims.Max.Z, 2+m) {
		t.Fatalf("max = %+v", dims.Max)
	}
	if !floatEq(dims.Size.Width, 2+2*m) || !floatEq(dims.Size.Height, 2+m) || !floatEq(dims.Size.Depth, 2+m) {
		t.Fatalf("size = %+v", dims.Size)
	}
	if dims.MarginM != m {
		t.Fatalf("margin_m = %v want %v", dims.MarginM, m)
	}

	allLights, err := s.ListSceneLights(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allLights) != 3 {
		t.Fatalf("all lights len = %d", len(allLights))
	}
	if allLights[0].Sx != 10 || allLights[0].Sy != 0 || allLights[0].Sz != 0 {
		t.Fatalf("first scene-space light = %+v", allLights[0])
	}

	cuboidLights, err := s.QuerySceneLightsCuboid(ctx, sc.ID, SceneCuboid{
		Position: ScenePoint{X: 10, Y: 0, Z: 0},
		Dimensions: SceneDimensionsSize{
			Width:  1,
			Height: 1,
			Depth:  1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cuboidLights) != 2 {
		t.Fatalf("cuboid lights len = %d", len(cuboidLights))
	}
	if cuboidLights[0].LightID != 0 || cuboidLights[1].LightID != 1 {
		t.Fatalf("cuboid lights = %+v", cuboidLights)
	}

	sphereLights, err := s.QuerySceneLightsSphere(ctx, sc.ID, SceneSphere{
		Center: ScenePoint{X: 10, Y: 0, Z: 0},
		Radius: 0.1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sphereLights) != 1 || sphereLights[0].LightID != 0 {
		t.Fatalf("sphere lights = %+v", sphereLights)
	}
}

func TestScenes_SpatialBulkUpdateAndInvalidGeometry(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "spatial-update", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 1, Z: 1},
		{ID: 2, X: 2, Y: 2, Z: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "scene-spatial-update", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PatchSceneModelOffsets(ctx, sc.ID, sum.ID, 10, 0, 0); err != nil {
		t.Fatal(err)
	}

	on := true
	color := "#112233"
	brightness := 42.0
	updated, err := s.PatchSceneLightsCuboid(ctx, sc.ID, SceneCuboid{
		Position: ScenePoint{X: 10, Y: 0, Z: 0},
		Dimensions: SceneDimensionsSize{
			Width:  1,
			Height: 1,
			Depth:  1,
		},
	}, LightStatePatch{
		On:            &on,
		Color:         &color,
		BrightnessPct: &brightness,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.UpdatedCount != 2 || len(updated.States) != 2 {
		t.Fatalf("updated = %+v", updated)
	}

	detail, err := s.Get(ctx, sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !detail.Lights[0].On || !detail.Lights[1].On || detail.Lights[2].On {
		t.Fatalf("unexpected on/off values: %+v", detail.Lights)
	}
	if detail.Lights[0].Color != "#112233" || detail.Lights[1].Color != "#112233" || detail.Lights[2].Color != "#ffffff" {
		t.Fatalf("unexpected colors: %+v", detail.Lights)
	}
	if detail.Lights[0].BrightnessPct != 42 || detail.Lights[1].BrightnessPct != 42 || detail.Lights[2].BrightnessPct != 100 {
		t.Fatalf("unexpected brightness: %+v", detail.Lights)
	}
	if detail.Lights[0].X != 0 || detail.Lights[1].X != 1 || detail.Lights[2].X != 2 {
		t.Fatalf("canonical model coordinates changed: %+v", detail.Lights)
	}

	off := false
	_, err = s.PatchSceneLightsSphere(ctx, sc.ID, SceneSphere{
		Center: ScenePoint{X: 10, Y: 0, Z: 0},
		Radius: -1,
	}, LightStatePatch{On: &off})
	if !errors.Is(err, ErrSceneInvalidGeometry) {
		t.Fatalf("want ErrSceneInvalidGeometry, got %v", err)
	}

	detailAfter, err := s.Get(ctx, sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !detailAfter.Lights[0].On || !detailAfter.Lights[1].On || detailAfter.Lights[2].On {
		t.Fatalf("state changed after invalid geometry: %+v", detailAfter.Lights)
	}
}

// TestScenes_PatchMarginM_DefaultsAndRoundTrip covers REQ-015 BR 12 / REQ-034 rule 3:
// new scenes default to 0.3 m, GET payloads expose margin_m, PATCH persists, and
// GetSceneDimensions reflects the per-scene value.
func TestScenes_PatchMarginM_DefaultsAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "margin-default", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 1, Z: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "scene-margin", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !floatEq(sc.MarginM, DefaultSceneBoundaryMarginM) {
		t.Fatalf("create margin = %v, want %v", sc.MarginM, DefaultSceneBoundaryMarginM)
	}

	d, err := s.GetScene(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !floatEq(d.MarginM, DefaultSceneBoundaryMarginM) {
		t.Fatalf("detail margin = %v, want %v", d.MarginM, DefaultSceneBoundaryMarginM)
	}
	list, err := s.ListScenes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !floatEq(list[0].MarginM, DefaultSceneBoundaryMarginM) {
		t.Fatalf("list margin = %+v", list)
	}

	updated, err := s.PatchSceneMarginM(ctx, sc.ID, 0.5)
	if err != nil {
		t.Fatalf("patch valid margin: %v", err)
	}
	if !floatEq(updated.MarginM, 0.5) || updated.ID != sc.ID || updated.ModelCount != 1 {
		t.Fatalf("patch echo unexpected: %+v", updated)
	}

	d2, err := s.GetScene(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !floatEq(d2.MarginM, 0.5) {
		t.Fatalf("detail margin after patch = %v", d2.MarginM)
	}

	dims, err := s.GetSceneDimensions(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Lights at sx 0..1, sy 0..1, sz 0..1; padded by 0.5 → [-0.5..1.5] clamped to [0..1.5].
	if !floatEq(dims.MarginM, 0.5) {
		t.Fatalf("dims margin = %v", dims.MarginM)
	}
	if !floatEq(dims.Max.X, 1.5) || !floatEq(dims.Max.Y, 1.5) || !floatEq(dims.Max.Z, 1.5) {
		t.Fatalf("dims max = %+v", dims.Max)
	}
	if dims.Origin.X != 0 || dims.Origin.Y != 0 || dims.Origin.Z != 0 {
		t.Fatalf("dims origin clamp = %+v", dims.Origin)
	}
}

// TestScenes_PatchMarginM_InvalidValues exercises REQ-015 BR 13 validation:
// reject NaN / Inf / negative / > MaxSceneBoundaryMarginM and leave the stored value untouched.
func TestScenes_PatchMarginM_InvalidValues(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	sum, err := s.Create(ctx, "margin-invalid", []wiremodel.Light{
		{ID: 0, X: 0, Y: 0, Z: 0},
		{ID: 1, X: 1, Y: 1, Z: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "scene-margin-invalid", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}

	cases := []float64{-1, -0.0001, MaxSceneBoundaryMarginM + 0.0001, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, v := range cases {
		if _, err := s.PatchSceneMarginM(ctx, sc.ID, v); !errors.Is(err, ErrInvalidMarginM) {
			t.Fatalf("expect ErrInvalidMarginM for %v, got %v", v, err)
		}
	}
	d, err := s.GetScene(ctx, sc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !floatEq(d.MarginM, DefaultSceneBoundaryMarginM) {
		t.Fatalf("margin mutated after invalid PATCH: %v", d.MarginM)
	}

	// Boundary values 0 and MaxSceneBoundaryMarginM are accepted.
	if _, err := s.PatchSceneMarginM(ctx, sc.ID, 0); err != nil {
		t.Fatalf("patch 0 should succeed: %v", err)
	}
	if _, err := s.PatchSceneMarginM(ctx, sc.ID, MaxSceneBoundaryMarginM); err != nil {
		t.Fatalf("patch %v should succeed: %v", MaxSceneBoundaryMarginM, err)
	}

	if _, err := s.PatchSceneMarginM(ctx, "no-such-scene", 0.42); !errors.Is(err, ErrSceneNotFound) {
		t.Fatalf("expect ErrSceneNotFound for unknown id, got %v", err)
	}
}
