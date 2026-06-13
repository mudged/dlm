# Printable fiducial marker assets (REQ-049)

Static ArUco markers served by `GET /api/v1/capture/marker`. Optional for video
reconstruction; when printed and placed in shot they improve pose estimation and
metric scale (REQ-048).

## Default marker

| Property | Value |
|----------|-------|
| Dictionary | ArUco 4×4, 50 codes (`DICT_4X4_50`) |
| Marker ID | 0 |
| Printed edge length | **100 mm** (0.1 m) — black square outer edge |
| Files | `fiducial_marker_aruco4x4_50_id0_100mm.pdf` (default), `.png` |

The printed edge length must match `edge_length_m` passed to the reconstruction
job (`0.1` for the default asset). Regenerate assets with OpenCV if these
parameters change; keep in sync with `internal/cvruntime` ArUco detection.
