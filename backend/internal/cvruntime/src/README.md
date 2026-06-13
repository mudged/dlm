# CV runtime source (`src/`)

Python source for the DLM light-position reconstruction pipeline (REQ-048).

| File | Purpose |
|------|---------|
| `reconstruct.py` | Production entry-point — reads a `JobSpec` JSON, writes a `Result` JSON to stdout |
| `gen_fixtures.py` | **Dev-only** — generates synthetic video clips + ground-truth JSON for testing |
| `test_reconstruct.py` | Integration tests that run the full pipeline on synthetic fixtures |

---

## Standalone development setup

```bash
# 1. Create and activate a virtual environment
python3 -m venv .venv
source .venv/bin/activate   # Windows: .venv\Scripts\activate

# 2. Install runtime deps (same as the bundled runtime)
pip install opencv-python-headless numpy
```

---

## Run reconstruct.py manually

```bash
# Write a minimal spec file
cat > spec.json <<'EOF'
{
  "feeds":    [{"path": "feed_0.mp4"}, {"path": "feed_1.mp4"}],
  "dwell_ms": 1000
}
EOF

python3 reconstruct.py spec.json
```

Output is JSON on stdout; diagnostics go to stderr.

**With an ArUco marker and scale hint:**

```json
{
  "feeds": [{"path": "feed_0.mp4"}, {"path": "feed_1.mp4"}],
  "marker": {"dictionary": "DICT_4X4_50", "edge_length_m": 0.05},
  "dwell_ms": 1000
}
```

**With a scale hint only** (no marker; supply the actual camera baseline in metres):

```json
{
  "feeds": [{"path": "feed_0.mp4"}, {"path": "feed_1.mp4"}],
  "scale_hint_m": 0.30,
  "dwell_ms": 1000
}
```

---

## Generate synthetic fixtures

```bash
# 6 lights, two feeds, no marker
python3 gen_fixtures.py --output-dir /tmp/fixtures --n-lights 6

# 4 lights with perspective-correct ArUco marker
python3 gen_fixtures.py --output-dir /tmp/fixtures --n-lights 4 --with-marker

# Occlude light 2 in feed 1 (test missing detection)
python3 gen_fixtures.py --output-dir /tmp/fixtures --n-lights 5 \
        --occlude-in-feed 1 2
```

Output files:

```
fixtures/
  feed_0.mp4
  feed_1.mp4
  ground_truth.json     # true 3D positions and camera poses
```

---

## Run the tests

```bash
python3 test_reconstruct.py -v
```

Each test generates tiny clips on-the-fly in a temp directory; no persistent
fixtures are required.  A full run takes roughly 10–30 seconds depending on
machine speed.

---

## JobSpec / Result contract

These match the Go structs in `../contract.go` (package `cvruntime`):

```
JobSpec {
  feeds:        [{path: string}, ...]          // ≥ 2 required
  marker?:      {dictionary: string,           // e.g. "DICT_4X4_50"
                 edge_length_m: number}
  scale_hint_m?: number                        // camera baseline in metres
  dwell_ms:     number                         // blink window from REQ-047
}

Result {
  status:         "succeeded" | "failed"
  light_count:    number                       // max seen index + 1
  lights:         [{id, x, y, z}, ...]         // metres, 0-based IDs
  missing:        [number, ...]                // indices not triangulable
  low_confidence: [number, ...]                // high reprojection error
  error:          string | null
}
```

Light IDs are **0-based integers** matching the sweep ordinal (REQ-047 /
REQ-048 BR 2).  The script **never fabricates** coordinates for undetected
lights (REQ-048 BR 7); they appear in `missing` or `low_confidence` instead.
