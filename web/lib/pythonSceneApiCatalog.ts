/**
 * Single manifest for Python `scene` API (REQ-024, REQ-026) — keep in sync with
 * `public/dlm-python-scene-worker.mjs` and CodeMirror completions.
 */
export type SceneApiCatalogEntry = {
  /** Stable id for anchors / keys */
  id: string;
  label: string;
  kind: "property" | "method";
  description: string;
  /** Copy-paste example; may use top-level await */
  snippet: string;
};

export const PYTHON_ROUTINE_DEFAULT_SOURCE = `# Demo: change colours of lights inside a sphere in scene space.
# Adjust center, radius, and colours for your layout.
import random

cx, cy, cz = 1.0, 1.0, 1.0
radius = 0.8
colour = "#%06x" % random.randrange(0x1000000)
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
    description:
      "Scene extent along +X (meters), from GET …/dimensions → size.width (architecture §3.15).",
    snippet: `w = scene.width
# Extent in meters along the scene +X axis`,
  },
  {
    id: "scene-height",
    label: "scene.height",
    kind: "property",
    description:
      "Scene extent along +Y (meters), from GET …/dimensions → size.height.",
    snippet: `h = scene.height`,
  },
  {
    id: "scene-depth",
    label: "scene.depth",
    kind: "property",
    description:
      "Scene extent along +Z (meters), from GET …/dimensions → size.depth.",
    snippet: `d = scene.depth`,
  },
  {
    id: "get-all-lights",
    label: "get_all_lights",
    kind: "method",
    description:
      "Returns every light in the scene with sx, sy, sz and state fields.",
    snippet: `lights = await scene.get_all_lights()`,
  },
  {
    id: "get-lights-sphere",
    label: "get_lights_within_sphere",
    kind: "method",
    description: "Query lights whose scene-space position lies inside a sphere.",
    snippet: `found = await scene.get_lights_within_sphere(
    {"x": 0, "y": 0, "z": 0},
    1.0,
)`,
  },
  {
    id: "get-lights-cuboid",
    label: "get_lights_within_cuboid",
    kind: "method",
    description: "Query lights inside an axis-aligned cuboid in scene space.",
    snippet: `found = await scene.get_lights_within_cuboid(
    {"x": 0, "y": 0, "z": 0},
    {"width": 2, "height": 2, "depth": 2},
)`,
  },
  {
    id: "set-all-lights",
    label: "set_all_lights",
    kind: "method",
    description: "Apply the same state patch to every light in the scene.",
    snippet: `await scene.set_all_lights(
    {"on": False, "color": "#ffffff", "brightness_pct": 100},
)`,
  },
  {
    id: "set-lights-sphere",
    label: "set_lights_in_sphere",
    kind: "method",
    description: "Bulk-update state for lights inside a sphere (scene coordinates).",
    snippet: `await scene.set_lights_in_sphere(
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
    snippet: `await scene.set_lights_in_cuboid(
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
    snippet: `await scene.update_lights_batch(
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

/** CodeMirror completion options — labels are member names after `scene.` */
export const SCENE_API_COMPLETIONS: {
  label: string;
  detail: string;
  type: string;
}[] = SCENE_API_CATALOG.map((e) => ({
  label: e.label.startsWith("scene.")
    ? e.label.slice("scene.".length)
    : e.label,
  detail: e.kind === "property" ? "property" : "() → await",
  type: e.kind === "property" ? "property" : "function",
}));
