package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRoutines_PythonCreatePatchList(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	r, err := s.CreateRoutine(ctx, "py1", "desc", RoutineTypePythonSceneScript, "print(1)")
	if err != nil {
		t.Fatal(err)
	}
	if r.PythonSource != "print(1)" || r.Type != RoutineTypePythonSceneScript {
		t.Fatalf("routine %+v", r)
	}

	list, err := s.ListRoutines(ctx)
	if err != nil || len(list) != 1 || list[0].PythonSource != "print(1)" {
		t.Fatalf("list %+v err %v", list, err)
	}

	got, err := s.GetRoutine(ctx, r.ID)
	if err != nil || got.PythonSource != "print(1)" {
		t.Fatalf("get %+v err %v", got, err)
	}

	updated, err := s.PatchRoutine(ctx, r.ID, nil, nil, ptr("x = 1"))
	if err != nil || updated.PythonSource != "x = 1" {
		t.Fatalf("patch %+v err %v", updated, err)
	}

	if _, err := s.PatchRoutine(ctx, r.ID, ptr("py1b"), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRoutines_PythonPatchBlockedWhenRunning(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)

	sum, err := s.Create(ctx, "m1", nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := s.CreateScene(ctx, "s1", []string{sum.ID})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.CreateRoutine(ctx, "py", "", RoutineTypePythonSceneScript, "pass")
	if err != nil {
		t.Fatal(err)
	}
	runID, already, err := s.StartRoutineRun(ctx, sc.ID, r.ID)
	if err != nil || already || runID == "" {
		t.Fatalf("start %v %v %v", runID, already, err)
	}
	if _, err := s.PatchRoutine(ctx, r.ID, nil, nil, ptr("y=2")); err != ErrRoutineRunActive {
		t.Fatalf("want ErrRoutineRunActive, got %v", err)
	}
	if err := s.StopRoutineRun(ctx, sc.ID, runID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.PatchRoutine(ctx, r.ID, nil, nil, ptr("y=2")); err != nil {
		t.Fatal(err)
	}
}

func TestRoutines_PatchNonPythonNotEditable(t *testing.T) {
	ctx := context.Background()
	s := testDB(t)
	id := uuid.NewString()
	created := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO routines (id, name, description, type, python_source, created_at) VALUES (?, ?, ?, ?, ?, ?)
	`, id, "legacy", "", RoutineTypeRandomColourCycleAll, "", created); err != nil {
		t.Fatal(err)
	}
	if _, err := s.PatchRoutine(ctx, id, nil, nil, ptr("x")); err != ErrRoutineNotEditable {
		t.Fatalf("got %v", err)
	}
}

func ptr(s string) *string { return &s }
