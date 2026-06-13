// Package cvruntime manages the bundled OpenCV+Python child process used for
// light-position reconstruction.  The Go binary itself stays pure-Go / cgo-free;
// all CV work runs in a separate supervised child (see run.go).
//
// Packaging mechanism: B — the runtime bundle ships as a sibling directory to the
// binary (dist/cvruntime/<goos>_<goarch>/) rather than being embedded inside the
// binary.  This keeps the binary sizes manageable while still requiring no operator
// Python install (REQ-048 BR5).  Mechanism A (go:embed) is a possible future
// upgrade; keeping the resolver pluggable (resolve.go) makes the switch cheap.
//
// For bundle build instructions and on-disk footprint estimates see:
// docs/advanced-setup.md §CV runtime bundle.
// WI-09 will expand that section with operator-facing setup notes.
package cvruntime

// JobSpec is the JSON payload written to a temp file and passed to the CV child
// process as its first argument.  Mirrors the contract in docs/work-items/WI-04.
type JobSpec struct {
	Feeds     []FeedRef `json:"feeds"`
	Marker    *Marker   `json:"marker,omitempty"`
	ScaleHint *float64  `json:"scale_hint_m,omitempty"`
	// DwellMS is the blink-detection window from REQ-047; forwarded to the child.
	DwellMS int `json:"dwell_ms"`
}

// FeedRef is a reference to a single video-feed file.
type FeedRef struct {
	Path string `json:"path"`
}

// Marker describes an optional ArUco marker used for scale/orientation reference.
type Marker struct {
	Dictionary  string  `json:"dictionary"`
	EdgeLengthM float64 `json:"edge_length_m"`
}

// Result is the JSON payload the CV child writes to stdout.
// Status is "succeeded" or "failed"; when "failed", Error is non-nil.
type Result struct {
	Status        string       `json:"status"`
	LightCount    int          `json:"light_count"`
	Lights        []LightPoint `json:"lights"`
	Missing       []int        `json:"missing"`
	LowConfidence []int        `json:"low_confidence"`
	Error         *string      `json:"error"`
}

// LightPoint is a reconstructed light position in SI metres, 0-based ID.
type LightPoint struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
}
