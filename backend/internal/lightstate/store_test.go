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

func ptrBool(b bool) *bool { return &b }
