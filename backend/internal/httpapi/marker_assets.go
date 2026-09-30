package httpapi

import (
	_ "embed"
	"strconv"
)

// Printable fiducial markers: ArUco DICT_4X4_50, ids 0, 1, and 2, 100 mm edge.
// Marker 1 is id 0. See assets/README.md.

//go:embed assets/fiducial_marker_aruco4x4_50_id0_100mm.pdf
var fiducialMarkerPDF0 []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id0_100mm.png
var fiducialMarkerPNG0 []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id1_100mm.pdf
var fiducialMarkerPDF1 []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id1_100mm.png
var fiducialMarkerPNG1 []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id2_100mm.pdf
var fiducialMarkerPDF2 []byte

//go:embed assets/fiducial_marker_aruco4x4_50_id2_100mm.png
var fiducialMarkerPNG2 []byte

func markerAsset(id int, png bool) (data []byte, contentType, filename string) {
	stem := "fiducial_marker_aruco4x4_50_id" + strconv.Itoa(id) + "_100mm"
	pdfs := [][]byte{fiducialMarkerPDF0, fiducialMarkerPDF1, fiducialMarkerPDF2}
	pngs := [][]byte{fiducialMarkerPNG0, fiducialMarkerPNG1, fiducialMarkerPNG2}
	if png {
		return pngs[id], "image/png", stem + ".png"
	}
	return pdfs[id], "application/pdf", stem + ".pdf"
}
