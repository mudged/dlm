package store

import (
	"context"
	"errors"
	"testing"

	"example.com/dlm/backend/internal/wiremodel"
)

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
