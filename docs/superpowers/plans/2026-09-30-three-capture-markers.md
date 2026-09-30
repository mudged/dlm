# Three printable capture markers — implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Offer three printable markers and, when the operator says they used one, join video clips only through those markers.

**Architecture:** Embedded PNG/PDF assets for ArUco ids 0, 1, and 2 are served by `GET /api/v1/capture/marker?id=`. Checking the marker box sends `ids: [0, 1, 2]` on the existing capture job. `reconstruct.py` detects those ids, keeps the largest chain of clips that share a marker, poses that group in the lowest id's frame, and names every other usable clip. It does not line cameras up from the blinking lights when a marker spec is present.

**Tech Stack:** Go 1.25 HTTP handlers, embedded assets, Python 3, OpenCV (`cv2`), NumPy, Next.js/Tailwind. Product docs in Markdown.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-30-three-capture-markers-design.md` (REQ-052). REQ-048 and REQ-049 stay. Do not renumber them.
- Marker 1, 2, and 3 are ArUco `DICT_4X4_50` ids 0, 1, and 2. Printed black outer edge is 100 mm (`edge_length_m` 0.1).
- `NOT_JOINED_REASON` is exactly `this clip does not share a marker with the clips used for the model`.
- Joined-footage failure is exactly `Not enough joined footage. At least two clips need to share a marker, or be linked by a clip that shows two markers. Dropped: {dropped}.` with `{dropped}` as `file (reason)` joined by `"; "`, in upload order.
- That failure runs only when at least two clips already survived the colour-flash checks and the marker join then leaves fewer than two. Fewer than two survivors before the join keeps today's flash or quiet-room message, with no `NOT_JOINED_REASON`.
- Largest joined group wins. Equal size: the group that contains the earliest uploaded file.
- Anchor is the lowest marker id in the chosen group. Several clips that could supply the same link: use the earliest uploaded clip. A camera that sees the anchor uses that pose. Otherwise use the lowest id it sees whose frame is already in the anchor frame.
- When `marker` is present, do not fall back to essential-matrix poses. When `marker` is absent, do not run this scan. `scale_hint` is not applied on the marker path. Metric scale is 1.0. Do not multiply coordinates by the edge length again.
- Other ArUco ids are ignored. Omitted or null `ids` means `[0, 1, 2]`. An explicit empty `ids` list accepts nothing.
- Scan the first `MARKER_SCAN_SECS` (5.0) seconds. A clip sees an id only after pose estimation succeeds. Keep the first success per id. The camera is still.
- No new capture-result JSON fields. No new review-page layout.
- Id 0's printed pattern stays today's pattern (`cv2.aruco.generateImageMarker` at 800 px matches the current PNG byte for byte).

## File structure

- `backend/internal/httpapi/assets/gen_markers.py` — regenerates the three PNGs and PDFs. One responsibility: printable files.
- `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id{0,1,2}_100mm.{png,pdf}` — embedded downloads.
- `backend/internal/httpapi/marker_assets.go` — embeds those six files.
- `backend/internal/httpapi/capture_marker.go` — `id` query and content type.
- `backend/internal/cvruntime/contract.go` — `Marker.IDs`.
- `backend/internal/httpapi/capture_models.go` — fills ids when `marker=true`.
- `backend/internal/cvruntime/src/reconstruct.py` — join, anchor poses, detection, job outcome.
- `web/app/models/new/NewModelClient.tsx` — three links and placement text.
- Requirements, design, user guide, and the asset README — REQ-052 wording.

---

### Task 1: Three downloadable markers

**Files:**
- Create: `backend/internal/httpapi/assets/gen_markers.py`
- Create: `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id1_100mm.png`
- Create: `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id1_100mm.pdf`
- Create: `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id2_100mm.png`
- Create: `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id2_100mm.pdf`
- Modify: `backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id0_100mm.pdf` (same marker image, plus the placement note)
- Modify: `backend/internal/httpapi/assets/README.md`
- Modify: `backend/internal/httpapi/marker_assets.go`
- Modify: `backend/internal/httpapi/capture_marker.go`
- Test: `backend/internal/httpapi/capture_marker_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `GET /api/v1/capture/marker?id=0|1|2&type=pdf|png`. Omitted `id` is id 0. Any other id is HTTP 400 `bad_request` / `marker id must be 0, 1, or 2`.

- [ ] **Step 1: Write the failing test**

Add these tests to `capture_marker_test.go`. Add `"bytes"` to the imports.

```go
func TestGetCaptureMarker_idSelectsDistinctPDF(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	bodies := map[string][]byte{}
	for _, id := range []string{"0", "1", "2"} {
		res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=" + id)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("id %s status = %d", id, res.StatusCode)
		}
		name := "fiducial_marker_aruco4x4_50_id" + id + "_100mm.pdf"
		if cd := res.Header.Get("Content-Disposition"); !strings.Contains(cd, name) {
			t.Fatalf("id %s Content-Disposition = %q", id, cd)
		}
		if !bytes.Contains(body, []byte("stick them on different sides")) {
			t.Fatalf("id %s pdf missing placement note", id)
		}
		bodies[id] = body
	}
	if bytes.Equal(bodies["0"], bodies["1"]) || bytes.Equal(bodies["1"], bodies["2"]) {
		t.Fatal("marker pdfs are not distinct")
	}
}

func TestGetCaptureMarker_idPNGDistinct(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	var prev []byte
	for _, id := range []string{"0", "1", "2"} {
		res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=" + id + "&type=png")
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("id %s status = %d", id, res.StatusCode)
		}
		if res.Header.Get("Content-Type") != "image/png" {
			t.Fatalf("id %s Content-Type = %q", id, res.Header.Get("Content-Type"))
		}
		if len(body) < 8 || body[0] != 0x89 || string(body[1:4]) != "PNG" {
			t.Fatalf("id %s is not a png", id)
		}
		if prev != nil && bytes.Equal(prev, body) {
			t.Fatalf("id %s png matches the previous id", id)
		}
		prev = body
	}
}

func TestGetCaptureMarker_badID(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=3")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "bad_request" || payload.Error.Message != "marker id must be 0, 1, or 2" {
		t.Fatalf("error = %#v", payload.Error)
	}
}
```

Add `"encoding/json"` to the imports.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/httpapi/ -run 'TestGetCaptureMarker_id|TestGetCaptureMarker_badID' -count=1`

Expected: FAIL. The id query is ignored, or id 1 has no embedded file.

- [ ] **Step 3: Generate the assets and serve them**

Write `backend/internal/httpapi/assets/gen_markers.py`:

```python
#!/usr/bin/env python3
"""Regenerate the three printable ArUco markers (DICT_4X4_50, ids 0–2, 800 px)."""

import zlib
from pathlib import Path

import cv2

ROOT = Path(__file__).resolve().parent
EDGE_PX = 800
NOTES = [
    "Marker {n} of 3. Print all three and stick them on different sides of what you are wrapping.",
    "Keep them flat and do not move them between clips.",
    "Each clip should show at least one marker. To join two sides that do not share a marker,",
    "at least one clip must show two markers at the same time.",
]


def _escape(text: str) -> str:
    return text.replace("\\", "\\\\").replace("(", "\\(").replace(")", "\\)")


def _pdf(image: bytes, width: int, height: int, lines: list[str]) -> bytes:
    content = "q 283.46 0 0 283.46 155.77 438.54 cm /Im0 Do Q\nBT /F1 11 Tf 50 390 Td 14 TL\n"
    for i, line in enumerate(lines):
        if i:
            content += "T* "
        content += f"({_escape(line)}) Tj\n"
    content += "ET\n"
    content_b = content.encode("latin1")
    objects = [
        b"<< /Type /Catalog /Pages 2 0 R >>",
        b"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
        b"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] "
        b"/Resources << /XObject << /Im0 5 0 R >> /Font << /F1 6 0 R >> >> "
        b"/Contents 4 0 R >>",
        b"<< /Length %d >>\nstream\n%s\nendstream" % (len(content_b), content_b),
        b"<< /Type /XObject /Subtype /Image /Width %d /Height %d "
        b"/ColorSpace /DeviceGray /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n"
        % (width, height, len(image))
        + image
        + b"\nendstream",
        b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
    ]
    out = bytearray(b"%PDF-1.4\n")
    offsets = []
    for i, body in enumerate(objects, start=1):
        offsets.append(len(out))
        out += f"{i} 0 obj\n".encode()
        out += body
        out += b"\nendobj\n"
    xref = len(out)
    out += f"xref\n0 {len(objects) + 1}\n".encode()
    out += b"0000000000 65535 f \n"
    for off in offsets:
        out += f"{off:010d} 00000 n \n".encode()
    out += (
        f"trailer\n<< /Size {len(objects) + 1} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n"
    ).encode()
    return bytes(out)


def main() -> None:
    dictionary = cv2.aruco.getPredefinedDictionary(cv2.aruco.DICT_4X4_50)
    for marker_id in (0, 1, 2):
        image = cv2.aruco.generateImageMarker(dictionary, marker_id, EDGE_PX)
        stem = f"fiducial_marker_aruco4x4_50_id{marker_id}_100mm"
        cv2.imwrite(str(ROOT / f"{stem}.png"), image)
        raw = image.tobytes()
        lines = [line.format(n=marker_id + 1) for line in NOTES]
        (ROOT / f"{stem}.pdf").write_bytes(_pdf(zlib.compress(raw), EDGE_PX, EDGE_PX, lines))


if __name__ == "__main__":
    main()
```

Run: `python3 backend/internal/httpapi/assets/gen_markers.py`

Expected: six asset files. Id 0 PNG bytes stay identical to the file already in git before this regeneration. Check with `git diff -- backend/internal/httpapi/assets/fiducial_marker_aruco4x4_50_id0_100mm.png` and expect no diff for that PNG.

Replace `marker_assets.go` with:

```go
package httpapi

import _ "embed"

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
```

Import `"strconv"` in `marker_assets.go`.

Replace `getCaptureMarker` in `capture_marker.go`:

```go
func (a *apiDeps) getCaptureMarker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id, ok := markerQueryID(r.URL.Query().Get("id"))
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "marker id must be 0, 1, or 2")
		return
	}
	format := r.URL.Query().Get("type")
	png := format == "png"
	switch format {
	case "", "pdf", "aruco", "png":
	default:
		png = false
	}
	data, contentType, filename := markerAsset(id, png)
	serveMarkerAsset(w, data, contentType, filename)
}

func markerQueryID(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id < 0 || id > 2 {
		return 0, false
	}
	return id, true
}
```

`capture_marker.go` already imports `strconv`.

Replace the table in `assets/README.md` so it lists ids 0, 1, and 2, says Marker 1 is id 0, and says each PDF includes the three placement sentences (different sides, keep still, one clip must show two markers). Point regeneration at `python3 gen_markers.py` from this directory.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/httpapi/ -run 'TestGetCaptureMarker' -count=1`

Expected: PASS, including the original PDF and PNG tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/httpapi/assets backend/internal/httpapi/marker_assets.go backend/internal/httpapi/capture_marker.go backend/internal/httpapi/capture_marker_test.go
git commit -m "$(cat <<'EOF'
Offer three printable capture markers for download.

EOF
)"
```

---

### Task 2: Send all three marker ids with the capture job

**Files:**
- Modify: `backend/internal/cvruntime/contract.go` (`Marker` struct)
- Modify: `backend/internal/httpapi/capture_models.go` (the `marker=true` assignment, about line 171)
- Test: `backend/internal/httpapi/capture_models_test.go` (`TestAPIv1CaptureModels_postWithMarkerAndScaleHint_forwardsCreateParams`)

**Interfaces:**
- Consumes: nothing from Task 1
- Produces: `cvruntime.Marker.IDs []int` with JSON `ids`. When the form field `marker` is `true` or `1`, the value is `[]int{0, 1, 2}` together with dictionary `DICT_4X4_50` and edge `0.1`.

- [ ] **Step 1: Write the failing assertion**

In `TestAPIv1CaptureModels_postWithMarkerAndScaleHint_forwardsCreateParams`, after the edge-length check, add:

```go
	if len(fake.last.Marker.IDs) != 3 || fake.last.Marker.IDs[0] != 0 || fake.last.Marker.IDs[1] != 1 || fake.last.Marker.IDs[2] != 2 {
		t.Fatalf("Marker.IDs = %v, want [0 1 2]", fake.last.Marker.IDs)
	}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/httpapi/ -run TestAPIv1CaptureModels_postWithMarkerAndScaleHint_forwardsCreateParams -count=1`

Expected: FAIL compiling (`IDs` undefined) or FAIL the new assertion.

- [ ] **Step 3: Add the field and set it**

In `contract.go`:

```go
type Marker struct {
	Dictionary  string  `json:"dictionary"`
	EdgeLengthM float64 `json:"edge_length_m"`
	IDs         []int   `json:"ids,omitempty"`
}
```

In `capture_models.go`, replace the marker assignment with:

```go
		params.Marker = &cvruntime.Marker{
			Dictionary:  "DICT_4X4_50",
			EdgeLengthM: 0.1,
			IDs:         []int{0, 1, 2},
		}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend && go test ./internal/httpapi/ -run TestAPIv1CaptureModels_postWithMarkerAndScaleHint_forwardsCreateParams -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/contract.go backend/internal/httpapi/capture_models.go backend/internal/httpapi/capture_models_test.go
git commit -m "$(cat <<'EOF'
Send all three capture marker ids when the operator used a marker.

EOF
)"
```

---

### Task 3: Join clips and express poses in one marker frame

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (next to the ArUco helpers, around `_poses_from_aruco`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py`

**Interfaces:**
- Consumes: nothing
- Produces:
  - `NOT_JOINED_REASON: str`
  - `_joined_footage_message(rejected_feeds: list[dict]) -> str`
  - `_marker_ids(marker_spec: dict) -> list[int]`
  - `_choose_joined_feeds(sightings: list[set[int]]) -> list[int]`
  - `_anchor_poses(detections: list[dict[int, tuple[np.ndarray, np.ndarray]]], chosen: list[int]) -> list[tuple[np.ndarray, np.ndarray]]`

`_choose_joined_feeds` returns feed indexes into `sightings`, sorted ascending. `_anchor_poses` returns one `(R, t)` per chosen feed, in that same order. `t` is shape `(3, 1)`.

- [ ] **Step 1: Write the failing test**

Append this class at the end of `test_reconstruct.py`:

```python
class TestMarkerJoin(unittest.TestCase):
    """REQ-052: which clips share a marker frame."""

    @classmethod
    def setUpClass(cls):
        cls.m = _load_reconstruct_module()

    def test_reason_and_failure_sentence(self):
        dropped = self.m._joined_footage_message([
            {"file": "a.avi", "reason": self.m.NOT_JOINED_REASON},
            {"file": "b.avi", "reason": "no start or end signal found"},
        ])
        self.assertEqual(
            dropped,
            "Not enough joined footage. At least two clips need to share a marker, "
            "or be linked by a clip that shows two markers. "
            "Dropped: a.avi (this clip does not share a marker with the clips used for the model); "
            "b.avi (no start or end signal found).",
        )

    def test_omitted_ids_mean_zero_one_two(self):
        self.assertEqual(self.m._marker_ids({"dictionary": "DICT_4X4_50", "edge_length_m": 0.1}), [0, 1, 2])
        self.assertEqual(self.m._marker_ids({"ids": []}), [])

    def test_chain_joins_three_clips(self):
        chosen = self.m._choose_joined_feeds([{0}, {0, 1}, {1}])
        self.assertEqual(chosen, [0, 1, 2])

    def test_outsider_is_left_out(self):
        chosen = self.m._choose_joined_feeds([{0}, {0}, {1}])
        self.assertEqual(chosen, [0, 1])

    def test_equal_groups_keep_the_earliest_file(self):
        chosen = self.m._choose_joined_feeds([{0}, {0}, {1}, {1}])
        self.assertEqual(chosen, [0, 1])

    def test_larger_group_wins(self):
        chosen = self.m._choose_joined_feeds([{1}, {1}, {0}, {0}, {0}])
        self.assertEqual(chosen, [2, 3, 4])

    def test_anchor_pose_uses_the_earliest_bridge(self):
        eye = np.eye(3, dtype=np.float64)

        def T(x):
            return (eye, np.array([[x], [0.0], [0.0]], dtype=np.float64))

        detections = [
            {0: T(0.0)},
            {0: T(-0.2), 1: T(-0.5)},
            {1: T(-0.5)},
        ]
        poses = self.m._anchor_poses(detections, [0, 1, 2])
        self.assertEqual(len(poses), 3)
        np.testing.assert_allclose(poses[2][1], [[-0.2], [0.0], [0.0]], atol=1e-9)
        np.testing.assert_allclose(poses[0][1], [[0.0], [0.0], [0.0]], atol=1e-9)
```

`_load_reconstruct_module` already exists in this test file. `numpy` is already imported as `np`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestMarkerJoin -v`

Expected: FAIL with `AttributeError` for `_choose_joined_feeds` or `NOT_JOINED_REASON`.

- [ ] **Step 3: Implement the helpers**

Add this block in `reconstruct.py` immediately above `_poses_from_aruco`:

```python
NOT_JOINED_REASON = (
    "this clip does not share a marker with the clips used for the model"
)


def _joined_footage_message(rejected_feeds: list[dict]) -> str:
    dropped = "; ".join(f"{r['file']} ({r['reason']})" for r in rejected_feeds)
    return (
        "Not enough joined footage. At least two clips need to share a marker, "
        "or be linked by a clip that shows two markers. "
        f"Dropped: {dropped}."
    )


def _marker_ids(marker_spec: dict) -> list[int]:
    raw = marker_spec.get("ids", None)
    if raw is None:
        return [0, 1, 2]
    return [int(i) for i in raw]


def _choose_joined_feeds(sightings: list[set[int]]) -> list[int]:
    """Largest set of feeds connected by shared marker ids.

    Equal sizes keep the set that contains the earliest feed. Returned
    indexes are sorted ascending.
    """
    n = len(sightings)
    parent = list(range(n))

    def find(i: int) -> int:
        while parent[i] != i:
            parent[i] = parent[parent[i]]
            i = parent[i]
        return i

    def union(a: int, b: int) -> None:
        ra, rb = find(a), find(b)
        if ra != rb:
            parent[rb] = ra

    seen_by_id: dict[int, list[int]] = {}
    for i, ids in enumerate(sightings):
        for marker_id in ids:
            seen_by_id.setdefault(marker_id, []).append(i)
    for feeds in seen_by_id.values():
        for other in feeds[1:]:
            union(feeds[0], other)

    components: dict[int, list[int]] = {}
    for i in range(n):
        components.setdefault(find(i), []).append(i)
    best = min(components.values(), key=lambda members: (-len(members), min(members)))
    return sorted(best)


def _anchor_poses(
    detections: list[dict[int, tuple[np.ndarray, np.ndarray]]],
    chosen: list[int],
) -> list[tuple[np.ndarray, np.ndarray]]:
    """Camera poses for *chosen* feeds, in the lowest marker id's frame.

    detections[i][marker_id] is (R, t) with x_cam = R @ x_marker + t.
    A marker frame already known as x_m = R_am @ x_anchor + t_am is linked
    by the earliest chosen feed that sees both that marker and a new one.
    """
    present: set[int] = set()
    for i in chosen:
        present.update(detections[i])
    anchor = min(present)
    known: dict[int, tuple[np.ndarray, np.ndarray]] = {
        anchor: (np.eye(3, dtype=np.float64), np.zeros((3, 1), dtype=np.float64)),
    }
    progressed = True
    while len(known) < len(present) and progressed:
        progressed = False
        for i in chosen:
            seen = detections[i]
            known_ids = [m for m in seen if m in known]
            unknown_ids = [m for m in seen if m not in known]
            if not known_ids or not unknown_ids:
                continue
            k = min(known_ids)
            R_k, t_k = seen[k]
            R_ak, t_ak = known[k]
            for u in sorted(unknown_ids):
                if u in known:
                    continue
                R_u, t_u = seen[u]
                R_au = R_u.T @ R_k @ R_ak
                t_au = R_u.T @ (R_k @ t_ak + t_k - t_u)
                known[u] = (R_au, t_au)
                progressed = True
    poses = []
    for i in chosen:
        seen = detections[i]
        marker_id = anchor if anchor in seen else min(m for m in seen if m in known)
        R_m, t_m = seen[marker_id]
        R_am, t_am = known[marker_id]
        poses.append((R_m @ R_am, R_m @ t_am + t_m))
    return poses
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestMarkerJoin -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
Join capture clips that share a printable marker.

EOF
)"
```

---

### Task 4: Use the joined group during reconstruction

**Files:**
- Modify: `backend/internal/cvruntime/src/reconstruct.py` (`_poses_from_aruco`, `_aruco_pose_single_feed`, `estimate_poses`, `main`)
- Test: `backend/internal/cvruntime/src/test_reconstruct.py` (`test_with_marker_spec_succeeds_and_finds_lights`, `test_aruco_no_double_scaling`, new `TestMarkerJoinVideos`)

**Interfaces:**
- Consumes: `NOT_JOINED_REASON`, `_joined_footage_message`, `_marker_ids`, `_choose_joined_feeds`, `_anchor_poses` from Task 3
- Produces:
  - `_detect_markers(feed_paths: list[str], Ks: list[np.ndarray], marker_spec: dict) -> list[dict[int, tuple[np.ndarray, np.ndarray]]]`
  - `_poses_from_detections(detections: list[dict[int, tuple[np.ndarray, np.ndarray]]]) -> tuple[list, float]`
  - `estimate_poses` marker branch calls those two and returns scale `1.0` without the essential-matrix fallback
  - `main` drops outsiders and fails with `_joined_footage_message` before triangulation

- [ ] **Step 1: Write the failing tests**

Replace `test_with_marker_spec_succeeds_and_finds_lights` with:

```python
    def test_with_marker_spec_succeeds_and_finds_lights(self):
        """A marker spec with no marker in the picture is not enough joined footage."""
        n = 6
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=n, seed=7)
            spec = _spec(d, gt)
            spec["marker"] = {"dictionary": "DICT_4X4_50", "edge_length_m": 0.05}
            res = _reconstruct(spec)

        self.assertEqual(res["status"], "failed", res)
        self.assertTrue(
            res["error"].startswith("Not enough joined footage."),
            res["error"],
        )
```

In `test_aruco_no_double_scaling`, replace the stub and the `estimate_poses` call with:

```python
        def _stub_detect(feed_paths, Ks, marker_spec):
            return [{0: (R0, t0)}, {0: (R1, t1)}]

        original = m._detect_markers
        m._detect_markers = _stub_detect
        try:
            poses, scale = m.estimate_poses(
                feed_paths=["a.mp4", "b.mp4"],
                Ks=[K, K],
                by_light={},
                marker_spec={"dictionary": "DICT_4X4_50", "edge_length_m": 0.05},
                scale_hint_m=None,
            )
        finally:
            m._detect_markers = original
```

Leave the rest of that test (scale `1.0` and the triangulation check) as it is.

Append `TestMarkerJoinVideos` to `test_reconstruct.py`. The helper draws each requested marker only on the leading 0.4 s, so the colour flashes and the bulb are not covered by the white marker square. Billboards are enough for join/reject tests. The distance test warps the marker with the same corner order `solvePnP` uses.

```python
def _write_avi(path: str, frames: list, fps: int = 30) -> None:
    h, w = frames[0].shape[:2]
    out = cv2.VideoWriter(path, cv2.VideoWriter_fourcc(*"XVID"), fps, (w, h))
    for frame in frames:
        out.write(frame)
    out.release()


def _blank(w: int, h: int):
    return np.zeros((h, w, 3), dtype=np.uint8)


def _billboard(frame, marker_id: int, center: tuple[int, int], size: int = 80) -> None:
    dictionary = cv2.aruco.getPredefinedDictionary(cv2.aruco.DICT_4X4_50)
    image = cv2.aruco.generateImageMarker(dictionary, marker_id, 64)
    cx, cy = center
    half = size // 2
    margin = 20
    h, w = frame.shape[:2]
    x0, y0 = max(0, cx - half - margin), max(0, cy - half - margin)
    x1, y1 = min(w, cx + half + margin), min(h, cy + half + margin)
    frame[y0:y1, x0:x1] = (255, 255, 255)
    patch = cv2.cvtColor(cv2.resize(image, (size, size), interpolation=cv2.INTER_NEAREST), cv2.COLOR_GRAY2BGR)
    frame[cy - half:cy + half, cx - half:cx + half] = patch


def _write_joined_clip(path: str, marker_ids: list[int], blink_xy: tuple[int, int] = (320, 300)) -> None:
    """One still clip: markers in the opening quiet moment, then one bulb and both bookends."""
    fps = 30
    w, h = 640, 480
    lead = 12  # 0.4 s
    cue = 6    # 0.2 s
    settle = 15
    dwell = 30
    frames = []
    centers = {0: (140, 140), 1: (500, 140), 2: (320, 80)}

    def base(with_markers: bool):
        frame = _blank(w, h)
        if with_markers:
            for marker_id in marker_ids:
                _billboard(frame, marker_id, centers.get(marker_id, (320, 80)))
        return frame

    for _ in range(lead):
        frames.append(base(True))
    colours = ((0, 0, 255), (255, 0, 0), (0, 255, 0))

    def bookend():
        for i, colour in enumerate(colours):
            if i:
                frames.extend(base(False) for _ in range(cue))
            for _ in range(cue):
                frame = base(False)
                cv2.circle(frame, blink_xy, 14, colour, -1)
                frames.append(frame)

    bookend()
    frames.extend(base(False) for _ in range(settle))
    for _ in range(dwell):
        frame = base(False)
        cv2.circle(frame, blink_xy, 14, (255, 255, 255), -1)
        frames.append(frame)
    frames.extend(base(False) for _ in range(settle))
    bookend()
    frames.extend(base(False) for _ in range(9))
    _write_avi(path, frames, fps)


def _spec_from_paths(paths: list[str], marker: bool) -> dict:
    spec = {
        "feeds": [{"path": p, "name": os.path.basename(p)} for p in paths],
        "dwell_ms": 1000,
    }
    if marker:
        spec["marker"] = {"dictionary": "DICT_4X4_50", "edge_length_m": 0.1, "ids": [0, 1, 2]}
    return spec


class TestMarkerJoinVideos(unittest.TestCase):
    """REQ-052 job outcomes. Clips are synthetic and the camera does not move."""

    def test_shared_marker_uses_the_pair_and_names_the_outsider(self):
        with tempfile.TemporaryDirectory() as d:
            paths = []
            for name, ids in (("a.avi", [0]), ("b.avi", [0]), ("c.avi", [1])):
                path = os.path.join(d, name)
                _write_joined_clip(path, ids)
                paths.append(path)
            res = _reconstruct(_spec_from_paths(paths, marker=True))
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(res["rejected_feeds"], [{
            "file": "c.avi",
            "reason": "this clip does not share a marker with the clips used for the model",
        }])

    def test_different_markers_without_a_bridge_fail(self):
        with tempfile.TemporaryDirectory() as d:
            a = os.path.join(d, "a.avi")
            b = os.path.join(d, "b.avi")
            _write_joined_clip(a, [0])
            _write_joined_clip(b, [1])
            res = _reconstruct(_spec_from_paths([a, b], marker=True))
        self.assertEqual(res["status"], "failed")
        self.assertTrue(res["error"].startswith("Not enough joined footage."), res["error"])
        self.assertIn("a.avi (this clip does not share a marker with the clips used for the model)", res["error"])
        self.assertIn("b.avi (this clip does not share a marker with the clips used for the model)", res["error"])

    def test_equal_pairs_keep_the_earlier_upload(self):
        with tempfile.TemporaryDirectory() as d:
            names = (("a.avi", [0]), ("b.avi", [0]), ("c.avi", [1]), ("d.avi", [1]))
            paths = []
            for name, ids in names:
                path = os.path.join(d, name)
                _write_joined_clip(path, ids)
                paths.append(path)
            res = _reconstruct(_spec_from_paths(paths, marker=True))
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        files = [row["file"] for row in res["rejected_feeds"]]
        self.assertEqual(files, ["c.avi", "d.avi"])

    def test_larger_group_is_used(self):
        with tempfile.TemporaryDirectory() as d:
            names = (("a.avi", [0]), ("b.avi", [0]), ("c.avi", [0]), ("d.avi", [1]), ("e.avi", [1]))
            paths = []
            for name, ids in names:
                path = os.path.join(d, name)
                _write_joined_clip(path, ids)
                paths.append(path)
            res = _reconstruct(_spec_from_paths(paths, marker=True))
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(
            [row["file"] for row in res["rejected_feeds"]],
            ["d.avi", "e.avi"],
        )

    def test_unknown_aruco_id_does_not_join(self):
        with tempfile.TemporaryDirectory() as d:
            a = os.path.join(d, "a.avi")
            b = os.path.join(d, "b.avi")
            c = os.path.join(d, "c.avi")
            _write_joined_clip(a, [0])
            _write_joined_clip(b, [0])
            _write_joined_clip(c, [7])
            res = _reconstruct(_spec_from_paths([a, b, c], marker=True))
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(res["rejected_feeds"], [{
            "file": "c.avi",
            "reason": "this clip does not share a marker with the clips used for the model",
        }])

    def test_missing_flashes_are_not_blamed_on_markers(self):
        with tempfile.TemporaryDirectory() as d:
            gt = _gen(d, n_lights=2, no_bookend=True, seed=1)
            spec = _spec(d, gt)
            spec["marker"] = {"dictionary": "DICT_4X4_50", "edge_length_m": 0.1, "ids": [0, 1, 2]}
            res = _reconstruct(spec)
        self.assertEqual(res["status"], "failed")
        self.assertIn("missed the start and end flashes", res["error"])
        self.assertFalse(res["error"].startswith("Not enough joined footage."))

    def test_bridge_puts_all_three_clips_in_one_metric_frame(self):
        """Two lights 0.36 m apart. A sees marker 0, C sees marker 1, B sees both."""
        with tempfile.TemporaryDirectory() as d:
            paths = _write_bridge_scene(d)
            res = _reconstruct(_spec_from_paths(paths, marker=True))
        self.assertEqual(res["status"], "succeeded", res.get("error"))
        self.assertEqual(res["rejected_feeds"], [])
        self.assertGreaterEqual(len(res["lights"]), 2, res)
        pts = [(p["x"], p["y"], p["z"]) for p in res["lights"]]
        dist = float(np.linalg.norm(np.array(pts[0]) - np.array(pts[1])))
        self.assertAlmostEqual(dist, 0.36, delta=0.36 * 0.20, msg=pts)
```

Add `_write_bridge_scene` next to `_write_joined_clip`. World frame is marker 0. Marker 1 sits 0.50 m to the right on the same plane. Cameras look toward the plane from z = 0.7. Lights sit between the markers, 0.36 m apart in y.

```python
def _write_bridge_scene(directory: str) -> list[str]:
    fps = 30
    width = height = 640
    f = 500.0
    K = np.array([[f, 0, width / 2], [0, f, height / 2], [0, 0, 1]], dtype=np.float64)
    R = np.array([[1, 0, 0], [0, -1, 0], [0, 0, -1]], dtype=np.float64)
    edge = 0.1
    half = edge / 2
    corners = np.array([
        [-half, half, 0],
        [half, half, 0],
        [half, -half, 0],
        [-half, -half, 0],
    ], dtype=np.float64)
    marker_1 = corners + np.array([0.50, 0, 0])
    lights = [np.array([0.25, 0.18, 0.0]), np.array([0.25, -0.18, 0.0])]
    cameras = {
        "a.avi": np.array([0.00, 0, 0.7]),
        "b.avi": np.array([0.25, 0, 0.7]),
        "c.avi": np.array([0.50, 0, 0.7]),
    }
    visible = {
        "a.avi": [0],
        "b.avi": [0, 1],
        "c.avi": [1],
    }
    dictionary = cv2.aruco.getPredefinedDictionary(cv2.aruco.DICT_4X4_50)
    images = {i: cv2.aruco.generateImageMarker(dictionary, i, 120) for i in (0, 1)}

    def project(point, t):
        cam = R @ point.reshape(3, 1) + t
        pix = K @ cam
        return np.array([pix[0, 0] / pix[2, 0], pix[1, 0] / pix[2, 0]], dtype=np.float32)

    def draw_marker(frame, marker_id, world_corners, t):
        dst = np.stack([project(p, t) for p in world_corners])
        if np.any(dst[:, 0] < 0) or np.any(dst[:, 0] >= width) or np.any(dst[:, 1] < 0) or np.any(dst[:, 1] >= height):
            return
        src = np.array([[0, 0], [119, 0], [119, 119], [0, 119]], dtype=np.float32)
        bigger = (dst - dst.mean(axis=0)) * 1.45 + dst.mean(axis=0)
        white = cv2.warpPerspective(
            np.full((120, 120), 255, np.uint8),
            cv2.getPerspectiveTransform(src, bigger.astype(np.float32)),
            (width, height),
        )
        gray = cv2.warpPerspective(
            images[marker_id],
            cv2.getPerspectiveTransform(src, dst.astype(np.float32)),
            (width, height),
        )
        mask = white > 0
        frame[mask] = 255
        ink = gray > 0
        colour = cv2.cvtColor(gray, cv2.COLOR_GRAY2BGR)
        frame[ink] = colour[ink]

    paths = []
    for name, center in cameras.items():
        t = -R @ center.reshape(3, 1)
        frames = []
        lead = 12
        cue = 6
        settle = 15
        dwell = 30

        def base(with_markers: bool):
            frame = _blank(width, height)
            if with_markers:
                if 0 in visible[name]:
                    draw_marker(frame, 0, corners, t)
                if 1 in visible[name]:
                    draw_marker(frame, 1, marker_1, t)
            return frame

        for _ in range(lead):
            frames.append(base(True))
        colours = ((0, 0, 255), (255, 0, 0), (0, 255, 0))
        light_px = [project(p, t) for p in lights]

        def bookend():
            for i, colour in enumerate(colours):
                if i:
                    frames.extend(base(False) for _ in range(cue))
                for _ in range(cue):
                    frame = base(False)
                    for pix in light_px:
                        cv2.circle(frame, (int(round(pix[0])), int(round(pix[1]))), 12, colour, -1)
                    frames.append(frame)

        bookend()
        frames.extend(base(False) for _ in range(settle))
        for pix in light_px:
            for _ in range(dwell):
                frame = base(False)
                cv2.circle(frame, (int(round(pix[0])), int(round(pix[1]))), 12, (255, 255, 255), -1)
                frames.append(frame)
            frames.extend(base(False) for _ in range(2))
        frames.extend(base(False) for _ in range(settle))
        bookend()
        frames.extend(base(False) for _ in range(9))
        path = os.path.join(directory, name)
        _write_avi(path, frames, fps)
        paths.append(path)
    return paths
```

`cv2` is already imported in `test_reconstruct.py`. If it is not, import it at the top of the test file the same way `reconstruct.py` imports it.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestMarkerJoinVideos test_with_marker_spec_succeeds_and_finds_lights test_aruco_no_double_scaling -v`

Expected: FAIL. The no-marker-in-picture job still succeeds via the essential matrix, `_detect_markers` is missing, and the video jobs do not set clips aside.

- [ ] **Step 3: Detect every allowed id and stop the essential-matrix fallback**

Replace `_aruco_pose_single_feed` with a detector that keeps every allowed id. Add `_detect_markers` and `_poses_from_detections` next to it:

```python
def _detect_markers(
    feed_paths: list[str],
    Ks: list[np.ndarray],
    marker_spec: dict,
) -> list[dict[int, tuple[np.ndarray, np.ndarray]]]:
    allowed = set(_marker_ids(marker_spec))
    adict = _aruco_dict(marker_spec["dictionary"])
    params = cv2.aruco.DetectorParameters()
    try:
        detector = cv2.aruco.ArucoDetector(adict, params)

        def _detect(gray):
            corners, ids, _ = detector.detectMarkers(gray)
            return corners, ids
    except AttributeError:
        def _detect(gray):
            corners, ids, _ = cv2.aruco.detectMarkers(gray, adict, parameters=params)
            return corners, ids

    edge_m = float(marker_spec["edge_length_m"])
    dist = np.zeros((4, 1), dtype=np.float64)
    found = []
    for fi, (path, K) in enumerate(zip(feed_paths, Ks)):
        found.append(_detect_markers_one_feed(path, K, dist, _detect, edge_m, allowed, fi))
    return found


def _detect_markers_one_feed(
    video_path: str,
    K: np.ndarray,
    dist: np.ndarray,
    detect_fn,
    edge_m: float,
    allowed: set[int],
    feed_idx: int,
) -> dict[int, tuple[np.ndarray, np.ndarray]]:
    cap = cv2.VideoCapture(video_path)
    poses: dict[int, tuple[np.ndarray, np.ndarray]] = {}
    try:
        if not cap.isOpened():
            return poses
        fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
        max_frames = max(1, int(MARKER_SCAN_SECS * fps))
        for _ in range(max_frames):
            if allowed and allowed <= set(poses):
                break
            ret, frame = cap.read()
            if not ret:
                break
            gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            corners, ids = detect_fn(gray)
            if ids is None:
                continue
            for corner, marker_id in zip(corners, ids.flatten().tolist()):
                marker_id = int(marker_id)
                if marker_id not in allowed or marker_id in poses:
                    continue
                rvecs, tvecs, _ = _estimate_pose_single(
                    [corner], edge_m, K, dist
                )
                poses[marker_id] = (_rodrigues(rvecs[0]), tvecs[0].reshape(3, 1))
                _log(f"    feed {feed_idx}: marker {marker_id} found")
    finally:
        cap.release()
    if not poses:
        _log(f"    feed {feed_idx}: no marker in first {MARKER_SCAN_SECS:.0f}s")
    return poses


def _poses_from_detections(
    detections: list[dict[int, tuple[np.ndarray, np.ndarray]]],
) -> tuple[list, float]:
    """Poses in the anchor frame when every feed is in one joined group.

    Returns all-None poses and scale 1.0 when they are not. Does not use
    the essential matrix.
    """
    n = len(detections)
    chosen = _choose_joined_feeds([set(d) for d in detections])
    if chosen != list(range(n)) or n < 2:
        return [None] * n, 1.0
    return _anchor_poses(detections, chosen), 1.0
```

In `estimate_poses`, replace the whole ArUco block (the `if marker_spec is not None` try/except that falls back when not every camera is localised) with:

```python
    if marker_spec is not None:
        _log("  trying ArUco marker pose estimation …")
        try:
            detections = _detect_markers(feed_paths, Ks, marker_spec)
            aruco_poses, scale = _poses_from_detections(detections)
            if all(p is not None for p in aruco_poses):
                _log(f"  ArUco: all {n} cameras localised")
                return aruco_poses, scale
            _log("  ArUco: feeds are not one joined group; not using E-matrix")
            return aruco_poses, scale
        except Exception as exc:
            _log(f"  ArUco failed ({exc}); not using E-matrix")
            return [None] * n, 1.0
```

Delete `_poses_from_aruco` if nothing else calls it. Do not leave a path that calls `_essential_matrix_pose` when `marker_spec` is not None.

In `main`, after the `len(usable) < 2` failure and before `compact = ...`, insert:

```python
        if marker_spec is not None:
            detections = _detect_markers(
                [feeds[fi]["path"] for fi in usable],
                [all_Ks[fi] for fi in usable],
                marker_spec,
            )
            chosen = _choose_joined_feeds([set(d) for d in detections])
            if len(chosen) < 2:
                for fi in usable:
                    rejected[fi] = NOT_JOINED_REASON
                rejected_feeds = _rejected_feeds(scans, rejected)
                _emit_failure(_joined_footage_message(rejected_feeds), rejected_feeds)
            chosen_set = set(chosen)
            for local_i, fi in enumerate(usable):
                if local_i not in chosen_set:
                    rejected[fi] = NOT_JOINED_REASON
            rejected_feeds = _rejected_feeds(scans, rejected)
            usable = [fi for local_i, fi in enumerate(usable) if local_i in chosen_set]
```

The existing `by_light` comprehension assumes every detection's feed is still in `usable`. After the join drops outsiders, skip those feeds. Replace that comprehension with:

```python
        by_light = {}
        for lid, dets in numbered.items():
            kept = [(compact[fi], cx, cy) for fi, cx, cy in dets if fi in compact]
            if kept:
                by_light[lid] = kept
```

Leave the later `estimate_poses(...)` call in place. It scans the already-joined feeds a second time and, because those feeds are one group, returns metric poses. Do not pass the outsiders in that list. Immediately after that call, if a marker spec was present and any pose is `None`, the second scan did not confirm the group. Fail instead of triangulating empty poses:

```python
        if marker_spec is not None and any(p is None for p in poses):
            for fi in usable:
                rejected[fi] = NOT_JOINED_REASON
            rejected_feeds = _rejected_feeds(scans, rejected)
            _emit_failure(_joined_footage_message(rejected_feeds), rejected_feeds)
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend/internal/cvruntime/src && python3 test_reconstruct.py TestMarkerJoin TestMarkerJoinVideos test_with_marker_spec_succeeds_and_finds_lights test_aruco_no_double_scaling test_most_lights_recovered -v`

Expected: PASS. `test_most_lights_recovered` is the unchanged no-marker path.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/cvruntime/src/reconstruct.py backend/internal/cvruntime/src/test_reconstruct.py
git commit -m "$(cat <<'EOF'
Reconstruct from the clips a shared marker can join.

EOF
)"
```

---

### Task 5: Three download links on the create-from-video page

**Files:**
- Modify: `web/app/models/new/NewModelClient.tsx` (the fiducial marker box, about lines 368–402)

**Interfaces:**
- Consumes: `GET /api/v1/capture/marker?id=` from Task 1
- Produces: links `Download marker 1`, `Download marker 2`, and `Download marker 3`. The checkbox label stays `I've placed a printed marker in the scene`.

- [ ] **Step 1: Replace the single link**

There is no component test for this page. Replace the single `<a href="/api/v1/capture/marker">` with three links and the placement list. Keep the checkbox.

```tsx
          <ul className="list-disc space-y-1 pl-4 text-xs text-slate-600 dark:text-slate-400">
            <li>Print all three and stick them on different sides of what you are wrapping.</li>
            <li>Keep them flat and do not move them between clips.</li>
            <li>
              Each clip should show at least one marker. To join two sides that do not share a
              marker, at least one clip must show two markers at the same time.
            </li>
            <li>You can still leave the marker box unchecked and build a model without them.</li>
          </ul>
          {(
            [
              ["0", "Download marker 1"],
              ["1", "Download marker 2"],
              ["2", "Download marker 3"],
            ] as const
          ).map(([id, label]) => (
            <a
              key={id}
              href={`/api/v1/capture/marker?id=${id}`}
              download
              className="inline-flex min-h-11 w-fit items-center gap-1.5 text-xs text-sky-600 underline-offset-2 hover:underline dark:text-sky-400"
            >
              {label}
            </a>
          ))}
```

The old download icon can go. The links stay in the existing column, so they stack on a phone. `min-h-11` keeps the tap target at least 44 px.

- [ ] **Step 2: Lint**

Run: `cd web && npm run lint`

Expected: PASS

- [ ] **Step 3: Check the page in the browser**

Start the app if it is not already running (`DLM_SKIP_NPM_CI=1 ./scripts/run.sh` from the repo root when dependencies are installed). Open `http://127.0.0.1:8080/models/new`.

Confirm:

- The create-from-video section shows `Download marker 1`, `Download marker 2`, and `Download marker 3`.
- The four placement sentences are visible.
- The checkbox still says `I've placed a printed marker in the scene`.
- Marker 1's link downloads a PDF (`/api/v1/capture/marker?id=0`). On a narrow viewport the three links stack and remain tappable.

- [ ] **Step 4: Commit**

```bash
git add web/app/models/new/NewModelClient.tsx
git commit -m "$(cat <<'EOF'
Show a download for each of the three capture markers.

EOF
)"
```

---

### Task 6: Document REQ-052

**Files:**
- Modify: `docs/requirements/requirements.md` (§11 marker paragraph and the feature code index)
- Modify: `docs/requirements/acceptance-criteria.md` (Building a model from video)
- Modify: `docs/design/overview.md` (add a REQ-052 row after REQ-051)
- Modify: `docs/design/appendix-traceability.md` (add a REQ-052 bullet after REQ-051)
- Modify: `docs/design/backend-lights-and-automation.md` (§3.23 pose bullet and §3.23.2)
- Modify: `docs/design/backend-service.md` (the capture marker and `POST /models/capture` rows)
- Modify: `docs/design/frontend.md` (§4.17 marker bullet)
- Modify: `docs/design/glossary.md` (fiducial marker entry)
- Modify: `docs/userguide/build-model-from-video.md` (section 5)

**Interfaces:**
- Consumes: the behaviour from Tasks 1–5
- Produces: REQ-052 in the index only. The readable requirements and the user guide do not contain `REQ-052`.

- [ ] **Step 1: Update the feature tour and the checklist**

In `requirements.md` §11, replace the paragraph that begins `There's also an **optional** printable **marker**` with:

```markdown
There's also an **optional** set of three printable **markers** (special printed patterns, each one
different). Print all three and stick them on different sides of what you are wrapping, such as a
tree. Keep them flat and do not move them between clips. Each clip should show at least one. Clips
that see the same marker go together. Clips that see different markers go together only when some
clip shows two markers at once, or a chain of clips links them. A clip that cannot be joined is
named and left out. If fewer than two clips can be joined, the app says there is not enough joined
footage and does not save a model. You can skip the markers and the app still works.
```

Append this index row. Do not renumber existing rows.

```markdown
| REQ-052 | Three printable markers so wrapped lights can still be joined | §11 Building a model from video |
```

In `acceptance-criteria.md`, change the review bullet's sentence `A printable marker is offered but isn't required.` to `Three printable markers are offered but aren't required.`

Add these two checks after that review item:

```markdown
**A clip that does not share a marker is named.**
- Try: print two different markers and film two clips that share one of them, plus a third clip that shows only the other marker. Upload all three with the marker box checked.
- You should see: a model you can review, and the third file set aside because this clip does not share a marker with the clips used for the model.

**Too little joined footage stops the job.**
- Try: upload two clips that show different markers, with no clip showing both, and the marker box checked.
- You should see: a message that starts with "Not enough joined footage." You are not offered Confirm.
```

- [ ] **Step 2: Update the design pages and the user guide**

Add to `overview.md` after the REQ-051 row:

```markdown
| REQ-052 | §3.23.2, §4.17: three printable ArUco markers (ids 0, 1, 2, 100 mm) download from the create-from-video page; with `marker=true`, clips join only through shared ids, including a chain from a clip that sees two ids; the largest group is reconstructed and other usable clips are `rejected_feeds`; fewer than two joined clips fails with "Not enough joined footage"; the unchecked path does not scan for markers. |
```

Add to `appendix-traceability.md` after the REQ-051 bullet:

```markdown
- **REQ-052** — three printable capture markers. Clips join only when they share one, or a clip shows two and links them. A clip that cannot be joined is named. Too little joined footage fails the job. See §3.23.2, §4.17.
```

In `backend-lights-and-automation.md`, replace the marker sentences inside the camera-pose bullet with:

```markdown
When the job includes a marker spec, detect ArUco ids 0, 1, and 2 (`marker.ids` when that list is present; omitted ids mean those three). A clip sees an id only when pose estimation succeeds in the first 5 seconds. Clips that share an id are one group. A clip that sees two ids joins those groups, including through a longer chain. Reconstruct the largest group. If two groups are the same size, use the one that contains the earliest uploaded file. Pose that group in the lowest id's frame. The earliest uploaded clip that sees two markers supplies the link between those frames. Scale is 1.0 because the printed edge is 0.1 m. Do not run the essential-matrix path when a marker spec is present. When the spec is absent, keep today's essential-matrix path and do not scan for markers.
```

Replace the body of §3.23.2 (keep the heading and section number) with:

```markdown
**In plain terms:** three different printed patterns can sit on different sides of a tree. A camera only has to see one of them. Clips join when they saw the same pattern, or when some clip saw two and links the sides. You can still build a model without printing any.

- `GET /api/v1/capture/marker?id=0|1|2` returns that marker. Omitting `id` returns Marker 1 (id 0). Any other id is **400** `bad_request` with message `marker id must be 0, 1, or 2`. `type=png` returns a PNG. Omitted `type`, `pdf`, or `aruco` returns a PDF. Any other `type` returns the PDF for that id.
- Filenames are `fiducial_marker_aruco4x4_50_id0_100mm.pdf` (and `.png`), and the same with `id1` and `id2`. Dictionary `DICT_4X4_50`, black outer edge 100 mm. Each PDF says to put the three markers on different sides, keep them still, and have one clip show two markers to join sides that do not share one.
- `marker=true` on `POST /api/v1/models/capture` sends `dictionary: "DICT_4X4_50"`, `edge_length_m: 0.1`, `ids: [0, 1, 2]`. Patterns outside that list are ignored.
- A usable clip outside the chosen group is a `rejected_feeds` row whose reason is `this clip does not share a marker with the clips used for the model`. The review list already shows that row.
- If the colour-flash checks left at least two clips and the join then leaves fewer than two, the job fails. `error` is `Not enough joined footage. At least two clips need to share a marker, or be linked by a clip that shows two markers. Dropped: {dropped}.` `{dropped}` is `file (reason)` joined with `"; "`, in upload order. Confirm is not offered. If fewer than two clips survived the colour-flash checks, keep today's flash or quiet-room error and do not use the marker reason.
- Printing a marker is still optional. With the marker box unchecked, reconstruction does not look for these patterns.
```

In `backend-service.md`, change the `GET /api/v1/capture/marker` cell so it says query `id` is `0`, `1`, or `2` (omitted means id 0), any other id is **400** `marker id must be 0, 1, or 2`, and `type=png` or the default PDF applies to that id. Change the `POST /api/v1/models/capture` marker sentence so `marker=true` fills `Dictionary: "DICT_4X4_50"`, `EdgeLengthM: 0.1`, and `IDs: [0, 1, 2]`.

In `frontend.md` §4.17, replace the single "Download printable marker" sentence with the three link labels and the four placement sentences. Say the review layout is unchanged.

In `glossary.md`, add to the fiducial-marker entry: video capture offers three patterns so a wrap can hide one of them from a given camera.

Replace user-guide section `### 5. Optional: a printable marker` with:

```markdown
### 5. Optional: three printable markers

If you want, you can download and print three special **markers** (printed square patterns, each one
different) from the upload screen. Stick them on different sides of what you are wrapping, such as a
tree. Keep them flat, and do not move them between clips.

Each clip should show at least one marker. If two clips see different markers, they can still be
joined when one clip shows two markers at the same time. A clip that cannot be joined is named and
left out. If the page says there is not enough joined footage, film again so at least two clips share
a marker, or add a clip that shows two of them.

This is totally optional. Leave the marker box unchecked and you can still build a model.
```

Do not put `REQ-052` in the user guide.

- [ ] **Step 3: Commit**

```bash
git add docs/requirements/requirements.md docs/requirements/acceptance-criteria.md docs/design/overview.md docs/design/appendix-traceability.md docs/design/backend-lights-and-automation.md docs/design/backend-service.md docs/design/frontend.md docs/design/glossary.md docs/userguide/build-model-from-video.md
git commit -m "$(cat <<'EOF'
Document the three printable capture markers.

EOF
)"
```

---

## Self-review notes

Spec coverage: downloads and the 400 id (Task 1), `ids` on upload (Task 2), group choice and anchor math (Task 3), job success/failure and no essential-matrix fallback (Task 4), page copy (Task 5), REQ-052 docs (Task 6). The bridge distance test is `test_bridge_puts_all_three_clips_in_one_metric_frame`. `test_most_lights_recovered` guards the unchecked path.

Placeholder scan: §3.23.2 is pasted in full under its existing heading. No unfinished steps.

Type check: `_detect_markers` returns `list[dict[int, tuple]]`. `_choose_joined_feeds` takes `list[set[int]]`. `_anchor_poses` and `_poses_from_detections` use that dict list. `Marker.IDs` / JSON `ids` is the only new Go field. The page links use `id=0|1|2`, matching Task 1.
