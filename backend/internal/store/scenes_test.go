package store

import (
	"context"
	"errors"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

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
	r, err := s.CreateRoutine(ctx, "fx", "d", RoutineTypeRandomColourCycleAll, "")
	if err != nil {
		t.Fatal(err)
	}
	runID, already, err := s.StartRoutineRun(ctx, sc.ID, r.ID)
	if err != nil || already || runID == "" {
		t.Fatalf("start %+v %v %v", runID, already, err)
	}
	d, err := s.Get(ctx, sum.ID)
	if err != nil || !d.Lights[0].On {
		t.Fatalf("after start want on, got %+v err %v", d.Lights[0], err)
	}
	if err := s.StopRoutineRun(ctx, sc.ID, runID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRoutine(ctx, r.ID); err != nil {
		t.Fatal(err)
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
	if dims.Origin.X != 0 || dims.Origin.Y != 0 || dims.Origin.Z != 0 {
		t.Fatalf("origin = %+v", dims.Origin)
	}
	if dims.Max.X != 13 || dims.Max.Y != 3 || dims.Max.Z != 3 {
		t.Fatalf("max = %+v", dims.Max)
	}
	if dims.Size.Width != 13 || dims.Size.Height != 3 || dims.Size.Depth != 3 {
		t.Fatalf("size = %+v", dims.Size)
	}
	if dims.MarginM != 1 {
		t.Fatalf("margin_m = %v", dims.MarginM)
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
