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
