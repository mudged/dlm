package wiremodel

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseLightsCSV_validMinimal(t *testing.T) {
	csv := "id,x,y,z\n0,0,0,0\n"
	lights, err := ParseLightsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(lights) != 1 || lights[0].ID != 0 {
		t.Fatalf("lights = %+v", lights)
	}
}

func TestParseLightsCSV_emptyModel(t *testing.T) {
	csv := "id,x,y,z\n"
	lights, err := ParseLightsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(lights) != 0 {
		t.Fatalf("expected 0 lights, got %d", len(lights))
	}
}

func TestParseLightsCSV_wrongHeader(t *testing.T) {
	csv := "idx,x,y,z\n0,0,0,0\n"
	_, err := ParseLightsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("want ParseError, got %T %v", err, err)
	}
	if !strings.Contains(strings.ToLower(pe.Message), "header") {
		t.Fatalf("message = %q", pe.Message)
	}
}

func TestParseLightsCSV_nonContiguousIDs(t *testing.T) {
	// Two rows with IDs 0 and 2: ID 2 is out of range [0,2), so the error
	// should identify id 2 specifically and still mention "sequential".
	csv := "id,x,y,z\n0,0,0,0\n2,1,1,1\n"
	_, err := ParseLightsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("want ParseError, got %T", err)
	}
	if !strings.Contains(strings.ToLower(pe.Message), "sequential") {
		t.Fatalf("message should mention 'sequential', got: %q", pe.Message)
	}
	if !strings.Contains(pe.Message, "2") {
		t.Fatalf("message should name the offending id 2, got: %q", pe.Message)
	}
}

func TestParseLightsCSV_duplicateID(t *testing.T) {
	// Three rows with IDs 0, 1, 0: duplicate 0 should produce a specific error.
	csv := "id,x,y,z\n0,0,0,0\n1,1,1,1\n0,2,2,2\n"
	_, err := ParseLightsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for duplicate id")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("want ParseError, got %T", err)
	}
	if !strings.Contains(pe.Message, "0") {
		t.Fatalf("message should name the duplicate id 0, got: %q", pe.Message)
	}
}

func TestMissingIDsError_message(t *testing.T) {
	err := MissingIDsError([]int{2, 5})
	if !strings.Contains(err.Message, "2") || !strings.Contains(err.Message, "5") {
		t.Fatalf("MissingIDsError should name ids 2 and 5, got: %q", err.Message)
	}
	if !strings.Contains(strings.ToLower(err.Message), "sequential") {
		t.Fatalf("MissingIDsError should mention 'sequential', got: %q", err.Message)
	}
}

func TestParseLightsCSV_nonNumericCoordinate(t *testing.T) {
	csv := "id,x,y,z\n0,not-a-number,0,0\n"
	_, err := ParseLightsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("want ParseError, got %T", err)
	}
	if !strings.Contains(strings.ToLower(pe.Message), "finite") && !strings.Contains(strings.ToLower(pe.Message), "number") {
		t.Fatalf("message = %q", pe.Message)
	}
}

func TestParseLightsCSV_over1000Rows(t *testing.T) {
	var b strings.Builder
	b.WriteString("id,x,y,z\n")
	for i := 0; i <= 1000; i++ {
		b.WriteString(strconv.Itoa(i))
		b.WriteString(",0,0,0\n")
	}
	_, err := ParseLightsCSV(strings.NewReader(b.String()))
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("want ParseError, got %T", err)
	}
	if !strings.Contains(strings.ToLower(pe.Message), "1000") {
		t.Fatalf("message = %q", pe.Message)
	}
}
