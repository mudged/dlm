package lightstate

import (
	"testing"
)

func TestStore_PatchEquiv(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 3)
	dto, unchanged, err := s.Patch("m1", 0, Patch{On: ptrBool(false)})
	if err != nil {
		t.Fatal(err)
	}
	if !unchanged || dto.On {
		t.Fatalf("expected noop default off, got unchanged=%v dto=%+v", unchanged, dto)
	}
	_, unchanged2, err := s.Patch("m1", 0, Patch{On: ptrBool(true)})
	if err != nil || unchanged2 {
		t.Fatalf("expected change, err=%v unchanged=%v", err, unchanged2)
	}
}

func TestStore_BatchDuplicateRejected(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 2)
	_, _, err := s.BatchPatch("m1", []int{0, 0}, Patch{On: ptrBool(true)})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestBatchPatch_ReturnsOnlyChangedLights(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 3)
	_, _, err := s.Patch("m1", 0, Patch{On: ptrBool(true)})
	if err != nil {
		t.Fatal(err)
	}
	on := true
	// light 0 already on — patching On:true is a no-op for that light
	// light 1 will change when On:true
	ids := []int{0, 1}
	out, allUnchanged, err := s.BatchPatch("m1", ids, Patch{On: &on})
	if err != nil {
		t.Fatal(err)
	}
	if allUnchanged {
		t.Fatal("expected at least one write")
	}
	if len(out) != 1 || out[0].ID != 1 || !out[0].On {
		t.Fatalf("want only changed light 1, got %+v", out)
	}
}

func TestResetAll_ReturnsOnlyChangedLights(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 2)
	_, _, err := s.Patch("m1", 0, Patch{On: ptrBool(true)})
	if err != nil {
		t.Fatal(err)
	}
	out, allDefault, err := s.ResetAll("m1")
	if err != nil {
		t.Fatal(err)
	}
	if allDefault {
		t.Fatal("expected at least one light to change")
	}
	if len(out) != 1 || out[0].ID != 0 || out[0].On {
		t.Fatalf("want only changed light 0 reset to default, got %+v", out)
	}
}

func TestBatchPatch_FullNoopReturnsEmpty(t *testing.T) {
	s := New()
	s.EnsureModel("m1", 2)
	off := false
	out, allUnchanged, err := s.BatchPatch("m1", []int{0, 1}, Patch{On: &off})
	if err != nil {
		t.Fatal(err)
	}
	if !allUnchanged {
		t.Fatal("expected all unchanged")
	}
	if len(out) != 0 {
		t.Fatalf("want empty slice on full noop, got %+v", out)
	}
}

func ptrBool(b bool) *bool { return &b }
