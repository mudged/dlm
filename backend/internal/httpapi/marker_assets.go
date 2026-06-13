package httpapi

import _ "embed"

// Default printable fiducial marker: ArUco DICT_4X4_50, id 0, 100 mm edge.
// See assets/README.md for dictionary, id, and scale reference.
//
//go:embed assets/fiducial_marker_aruco4x4_50_id0_100mm.pdf
var fiducialMarkerPDF []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id0_100mm.png
var fiducialMarkerPNG []byte
