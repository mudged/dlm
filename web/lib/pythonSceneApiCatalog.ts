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

export const PYTHON_ROUTINE_DEFAULT_SOURCE = `# Try this: pick random colours for lights inside a ball-shaped area.
# Change the middle point (cx, cy, cz), the radius, or add your own ideas.
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
    description: "How wide the room is in metres (left to right).",
    snippet: `w = scene.width
# Room width in metres`,
  },
  {
    id: "scene-height",
    label: "scene.height",
    kind: "property",
    description: "How tall the room is in metres (floor to ceiling).",
    snippet: `h = scene.height`,
  },
  {
    id: "scene-depth",
    label: "scene.depth",
    kind: "property",
    description: "How deep the room is in metres (front to back).",
    snippet: `d = scene.depth`,
  },
  {
    id: "get-all-lights",
    label: "get_all_lights",
    kind: "method",
    description:
      "Gives you a list of every light and its place in the room (sx, sy, sz) plus on/off, colour, and brightness.",
    snippet: `lights = await scene.get_all_lights()`,
  },
  {
    id: "get-lights-sphere",
    label: "get_lights_within_sphere",
    kind: "method",
    description:
      "Find lights inside a ball: you give the centre point and how big the ball is (radius).",
    snippet: `found = await scene.get_lights_within_sphere(
    {"x": 0, "y": 0, "z": 0},
    1.0,
)`,
  },
  {
    id: "get-lights-cuboid",
    label: "get_lights_within_cuboid",
    kind: "method",
    description:
      "Find lights inside a box: you give one corner’s position and how wide, tall, and deep the box is.",
    snippet: `found = await scene.get_lights_within_cuboid(
    {"x": 0, "y": 0, "z": 0},
    {"width": 2, "height": 2, "depth": 2},
)`,
  },
  {
    id: "set-all-lights",
    label: "set_all_lights",
    kind: "method",
    description:
      "Set every light the same way at once (on or off, colour, how bright).",
    snippet: `await scene.set_all_lights(
    {"on": False, "color": "#ffffff", "brightness_pct": 100},
)`,
  },
  {
    id: "set-lights-sphere",
    label: "set_lights_in_sphere",
    kind: "method",
    description:
      "Change only the lights inside a ball — handy for spot effects and demos.",
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
    description: "Change only the lights inside a box-shaped area.",
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
    description:
      "Change many specific lights in one list. Each line needs which model, which light number, and the new settings.",
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
