/**
 * Scene command list for the Python routine UI — keep in sync with
 * `public/dlm-python-scene-worker.mjs` and CodeMirror completions.
 */
import {
  PYTHON_SAMPLE_GROWING_SPHERE_SOURCE,
  PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE,
  PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE,
} from "@/lib/pythonRoutineSamples";

export type SceneApiCatalogEntry = {
  /** Stable id for anchors / keys */
  id: string;
  label: string;
  kind: "property" | "method" | "sample";
  description: string;
  /** Copy-paste example; may use top-level await */
  snippet: string;
  /** When set, overrides completion list detail (e.g. synchronous methods). */
  completionDetail?: string;
};

export const PYTHON_ROUTINE_DEFAULT_SOURCE = `# Demo: change colours of lights inside a sphere in scene space.
# Adjust center, radius, and colours for your layout.

# Middle of a typical room (metres)
cx, cy, cz = 1.0, 1.0, 1.0
# How big the bubble is (metres)
radius = 0.8
# Random colour for the lights (# + six hex digits, REQ-030)
colour = scene.random_hex_colour()
# Turn on lights inside the sphere and paint them
await scene.set_lights_in_sphere(
    {"x": cx, "y": cy, "z": cz},
    radius,
    {"on": True, "color": colour, "brightness_pct": 100},
)
`;

export const SCENE_API_CATALOG: SceneApiCatalogEntry[] = [
  {
    id: "scene-width",
    label: "scene.width",
    kind: "property",
    description: "How wide the room is in metres (left to right).",
    snippet: `# Room width in metres (left to right)
w = scene.width`,
  },
  {
    id: "scene-height",
    label: "scene.height",
    kind: "property",
    description: "How tall the room is in metres (floor to ceiling).",
    snippet: `# How tall the room is (floor to ceiling, metres)
h = scene.height`,
  },
  {
    id: "scene-depth",
    label: "scene.depth",
    kind: "property",
    description: "How deep the room is in metres (front to back).",
    snippet: `# Front-to-back size of the room (metres)
d = scene.depth`,
  },
  {
    id: "scene-max-x",
    label: "scene.max_x",
    kind: "property",
    description:
      "Far corner of the room along +X (metres). With max_y and max_z it bounds where lights can sit.",
    snippet: `# How far lights reach along +X (metres)
mx = scene.max_x`,
  },
  {
    id: "scene-max-y",
    label: "scene.max_y",
    kind: "property",
    description: "Far corner along +Y — ceiling direction in the default 3D view (metres).",
    snippet: `# How high lights can go (metres, +Y up)
my = scene.max_y`,
  },
  {
    id: "scene-max-z",
    label: "scene.max_z",
    kind: "property",
    description: "Far corner along +Z (depth into the screen in the default view, metres).",
    snippet: `# How far lights reach along +Z (metres)
mz = scene.max_z`,
  },
  {
    id: "random-hex-colour",
    label: "random_hex_colour",
    kind: "method",
    description:
      "Returns a random #rrggbb colour string you can use for light color (no import random needed).",
    completionDetail: "() → str",
    snippet: `# Pick a random colour string for a light
colour = scene.random_hex_colour()`,
  },
  {
    id: "get-all-lights",
    label: "get_all_lights",
    kind: "method",
    description:
      "Returns every light in the scene with sx, sy, sz and state fields.",
    snippet: `# Get every light and its on/off, colour, and position
lights = await scene.get_all_lights()`,
  },
  {
    id: "get-lights-sphere",
    label: "get_lights_within_sphere",
    kind: "method",
    description: "Query lights whose scene-space position lies inside a sphere.",
    snippet: `# Lights inside a ball around a point (centre + radius in metres)
found = await scene.get_lights_within_sphere(
    {"x": 0, "y": 0, "z": 0},
    1.0,
)`,
  },
  {
    id: "get-lights-cuboid",
    label: "get_lights_within_cuboid",
    kind: "method",
    description: "Query lights inside an axis-aligned cuboid in scene space.",
    snippet: `# Lights inside a box: corner + width, height, depth
found = await scene.get_lights_within_cuboid(
    {"x": 0, "y": 0, "z": 0},
    {"width": 2, "height": 2, "depth": 2},
)`,
  },
  {
    id: "set-all-lights",
    label: "set_all_lights",
    kind: "method",
    description: "Apply the same state patch to every light in the scene.",
    snippet: `# Same look for every light (off, white, full brightness)
await scene.set_all_lights(
    {"on": False, "color": "#ffffff", "brightness_pct": 100},
)`,
  },
  {
    id: "set-lights-sphere",
    label: "set_lights_in_sphere",
    kind: "method",
    description: "Bulk-update state for lights inside a sphere (scene coordinates).",
    snippet: `# Turn on lights in a ball and set their colour
await scene.set_lights_in_sphere(
    {"x": 1, "y": 1, "z": 1},
    0.5,
    {"on": True, "color": "#ff6600", "brightness_pct": 100},
)`,
  },
  {
    id: "set-lights-cuboid",
    label: "set_lights_in_cuboid",
    kind: "method",
    description: "Bulk-update state for lights inside a cuboid.",
    snippet: `# Change lights inside a box-shaped region
await scene.set_lights_in_cuboid(
    {"x": 0, "y": 0, "z": 0},
    {"width": 1, "height": 1, "depth": 1},
    {"on": True, "color": "#00ff00", "brightness_pct": 80},
)`,
  },
  {
    id: "update-lights-batch",
    label: "update_lights_batch",
    kind: "method",
    description: "Per-light updates; each entry needs model_id, light_id, and state fields.",
    snippet: `# Change only specific lights (replace model id with a real one)
await scene.update_lights_batch(
    [
        {
            "model_id": "<uuid>",
            "light_id": 0,
            "on": True,
            "color": "#ffffff",
            "brightness_pct": 100,
        },
    ],
)`,
  },
];

/** REQ-032 full routine bodies — same strings as toolbar “Load sample”. */
export const PYTHON_ROUTINE_SAMPLE_CATALOG_ENTRIES: SceneApiCatalogEntry[] = [
  {
    id: "sample-growing-sphere",
    label: "Sample: growing sphere",
    kind: "sample",
    description:
      "Full script: a sphere in the middle grows for 10 seconds, colours lights inside, then repeats with a new colour. Uses a small saved state dict so Stop still works.",
    snippet: PYTHON_SAMPLE_GROWING_SPHERE_SOURCE,
  },
  {
    id: "sample-sweeping-cuboid",
    label: "Sample: sweeping cuboid",
    kind: "sample",
    description:
      "Full script: a 20 cm slab covers the floor and slides to the ceiling in 10 seconds; lights inside turn on with a random colour; lights that leave turn off; then it repeats.",
    snippet: PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE,
  },
  {
    id: "sample-random-colour-cycle",
    label: "Sample: random colour cycle (all lights)",
    kind: "sample",
    description:
      "Full script: turns every light on, then about once per second gives each light a new random colour (replaces the old server-only colour cycle).",
    snippet: PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE,
  },
];

/** REQ-024 picker: API members plus REQ-032 samples (no duplicate snippet sources). */
export const PYTHON_SCENE_API_CATALOG_FULL: SceneApiCatalogEntry[] = [
  ...SCENE_API_CATALOG,
  ...PYTHON_ROUTINE_SAMPLE_CATALOG_ENTRIES,
];

/** CodeMirror completion options — labels are member names after `scene.` */
export const SCENE_API_COMPLETIONS: {
  label: string;
  detail: string;
  type: string;
}[] = SCENE_API_CATALOG.map((e) => ({
  label: e.label.startsWith("scene.")
    ? e.label.slice("scene.".length)
    : e.label,
  detail:
    e.completionDetail ??
    (e.kind === "property" ? "property" : "() → await"),
  type: e.kind === "property" ? "property" : "function",
}));
