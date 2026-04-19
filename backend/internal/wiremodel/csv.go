package wiremodel

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Light is one indexed position on the wire (REQ-005).
type Light struct {
	ID int
	X  float64
	Y  float64
	Z  float64
}

// ParseError describes CSV validation failure (REQ-007).
type ParseError struct {
	Message string
}

func (e *ParseError) Error() string { return e.Message }

var wantHeader = []string{"id", "x", "y", "z"}

// ParseLightsCSV reads a UTF-8 CSV with header id,x,y,z and validates rows per docs/agentic-development/architecture §3.6.
func ParseLightsCSV(r io.Reader) ([]Light, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = 4
	cr.ReuseRecord = true

	header, err := cr.Read()
	if err == io.EOF {
		return nil, &ParseError{Message: "CSV is empty or missing header row"}
	}
	if err != nil {
		return nil, &ParseError{Message: fmt.Sprintf("invalid CSV header: %v", err)}
	}
	if len(header) != 4 {
		return nil, &ParseError{Message: "CSV must have exactly four columns: id,x,y,z"}
	}
	for i := range wantHeader {
		if strings.TrimSpace(header[i]) != wantHeader[i] {
			return nil, &ParseError{Message: "CSV header must be exactly id,x,y,z in that order"}
		}
	}

	var lights []Light
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &ParseError{Message: fmt.Sprintf("invalid CSV row: %v", err)}
		}
		id, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			return nil, &ParseError{Message: "each id must be an integer"}
		}
		x, err := parseFiniteFloat(rec[1])
		if err != nil {
			return nil, &ParseError{Message: "x, y, and z must be finite numbers"}
		}
		y, err := parseFiniteFloat(rec[2])
		if err != nil {
			return nil, &ParseError{Message: "x, y, and z must be finite numbers"}
		}
		z, err := parseFiniteFloat(rec[3])
		if err != nil {
			return nil, &ParseError{Message: "x, y, and z must be finite numbers"}
		}
		lights = append(lights, Light{ID: id, X: x, Y: y, Z: z})
	}

	n := len(lights)
	if n > 1000 {
		return nil, &ParseError{Message: "a model may have at most 1000 lights"}
	}
	if err := validateContiguousIDs(lights, n); err != nil {
		return nil, err
	}

	return lights, nil
}

func parseFiniteFloat(s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, strconv.ErrSyntax
	}
	return v, nil
}

func validateContiguousIDs(lights []Light, n int) error {
	if n == 0 {
		return nil
	}
	seen := make([]bool, n)
	for _, L := range lights {
		id := L.ID
		if id < 0 || id >= n {
			return &ParseError{Message: "light ids must be sequential starting at 0 with no gaps (expected ids 0 through n-1 for each row)"}
		}
		if seen[id] {
			return &ParseError{Message: "light ids must be sequential starting at 0 with no gaps (duplicate id)"}
		}
		seen[id] = true
	}
	for i := range seen {
		if !seen[i] {
			return &ParseError{Message: "light ids must be sequential starting at 0 with no gaps"}
		}
	}
	return nil
}
