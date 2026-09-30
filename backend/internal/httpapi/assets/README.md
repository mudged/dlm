# Printable fiducial marker assets (REQ-049)

Static ArUco markers served by `GET /api/v1/capture/marker`. Optional for video
reconstruction; when printed and placed in shot they improve pose estimation and
metric scale (REQ-048).

## Markers (ids 0, 1, and 2)

| Property | Value |
|----------|-------|
| Dictionary | ArUco 4×4, 50 codes (`DICT_4X4_50`) |
| Marker IDs | 0, 1, and 2 (**Marker 1** in the user-facing set is **id 0**) |
| Printed edge length | **100 mm** (0.1 m) — black square outer edge |
| Files | `fiducial_marker_aruco4x4_50_id{N}_100mm.pdf` (default), `.png` for each id |

Each PDF includes three placement sentences: print all three and stick them on
different sides of what you are wrapping; keep them flat and do not move them
between clips; and when joining sides that do not share a marker, at least one
clip must show two markers at the same time.

The printed edge length must match `edge_length_m` passed to the reconstruction
job (`0.1` for these assets). Regenerate from this directory with:

```bash
python3 gen_markers.py
```

Keep assets in sync with `internal/cvruntime` ArUco detection.
