/**
 * Form state ↔ definition_json (REQ-033 / server schema) for shape animation authoring.
 */

import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "@/lib/shapeAnimationDefault";

export type EdgeBehavior =
  | "wrap"
  | "stop"
  | "deflect_random"
  | "deflect_specular";

export type FaceId = "top" | "bottom" | "left" | "right" | "back" | "front";

export type ShapeFormRow = {
  clientKey: string;
  kind: "sphere" | "cuboid";
  sizeMode: "fixed" | "random_uniform";
  radius_m: string;
  radius_min_m: string;
  radius_max_m: string;
  width_m: string;
  height_m: string;
  depth_m: string;
  width_min_m: string;
  width_max_m: string;
  height_min_m: string;
  height_max_m: string;
  depth_min_m: string;
  depth_max_m: string;
  colorMode: "fixed" | "random";
  color: string;
  brightness_pct: string;
  placementMode: "fixed" | "random_face";
  center_x: string;
  center_y: string;
  center_z: string;
  min_corner_x: string;
  min_corner_y: string;
  min_corner_z: string;
  face: FaceId;
  motion_dx: string;
  motion_dy: string;
  motion_dz: string;
  speedMode: "fixed" | "random_uniform";
  m_s: string;
  min_m_s: string;
  max_m_s: string;
  edge_behavior: EdgeBehavior;
};

export type ShapeAnimationFormState = {
  backgroundMode: "lights_on" | "lights_off";
  backgroundColor: string;
  backgroundBrightness_pct: string;
  shapes: ShapeFormRow[];
};

function newClientKey(): string {
  return `s-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function defaultShapeFormRow(): ShapeFormRow {
  return {
    clientKey: newClientKey(),
    kind: "sphere",
    sizeMode: "fixed",
    radius_m: "0.35",
    radius_min_m: "0.2",
    radius_max_m: "0.5",
    width_m: "0.4",
    height_m: "0.4",
    depth_m: "0.4",
    width_min_m: "0.2",
    width_max_m: "0.6",
    height_min_m: "0.2",
    height_max_m: "0.6",
    depth_min_m: "0.2",
    depth_max_m: "0.6",
    colorMode: "fixed",
    color: "#00ff88",
    brightness_pct: "100",
    placementMode: "fixed",
    center_x: "1",
    center_y: "1",
    center_z: "1",
    min_corner_x: "0",
    min_corner_y: "0",
    min_corner_z: "0",
    face: "front",
    motion_dx: "1",
    motion_dy: "0.3",
    motion_dz: "0",
    speedMode: "fixed",
    m_s: "0.15",
    min_m_s: "0.05",
    max_m_s: "0.3",
    edge_behavior: "wrap",
  };
}

export function defaultShapeAnimationFormState(): ShapeAnimationFormState {
  return {
    backgroundMode: "lights_on",
    backgroundColor: "#1a1a2e",
    backgroundBrightness_pct: "40",
    shapes: [defaultShapeFormRow()],
  };
}

function parseNum(label: string, s: string): number {
  const n = Number(String(s).trim());
  if (!Number.isFinite(n)) {
    throw new Error(`${label} must be a number`);
  }
  return n;
}

function parsePositive(label: string, s: string): number {
  const n = parseNum(label, s);
  if (n <= 0) {
    throw new Error(`${label} must be positive`);
  }
  return n;
}

function rowFromShapeJson(sh: Record<string, unknown>): ShapeFormRow {
  const kind = sh.kind === "cuboid" ? "cuboid" : "sphere";
  const size = (sh.size ?? {}) as Record<string, unknown>;
  const sizeMode = size.mode === "random_uniform" ? "random_uniform" : "fixed";
  const col = (sh.color ?? {}) as Record<string, unknown>;
  const colorMode = col.mode === "random" ? "random" : "fixed";
  const pl = (sh.placement ?? {}) as Record<string, unknown>;
  const placementMode = pl.mode === "random_face" ? "random_face" : "fixed";
  const mot = (sh.motion ?? {}) as Record<string, unknown>;
  const dir = (mot.direction ?? {}) as Record<string, unknown>;
  const sp = (mot.speed ?? {}) as Record<string, unknown>;
  const speedMode = sp.mode === "random_uniform" ? "random_uniform" : "fixed";
  const edge = String(sh.edge_behavior ?? "wrap") as EdgeBehavior;

  const center = (pl.center_m ?? {}) as Record<string, unknown>;
  const minC = (pl.min_corner_m ?? {}) as Record<string, unknown>;

  return {
    clientKey: newClientKey(),
    kind,
    sizeMode,
    radius_m: String(size.radius_m ?? "0.35"),
    radius_min_m: String(size.radius_min_m ?? "0.2"),
    radius_max_m: String(size.radius_max_m ?? "0.5"),
    width_m: String(size.width_m ?? "0.4"),
    height_m: String(size.height_m ?? "0.4"),
    depth_m: String(size.depth_m ?? "0.4"),
    width_min_m: String(size.width_min_m ?? "0.2"),
    width_max_m: String(size.width_max_m ?? "0.6"),
    height_min_m: String(size.height_min_m ?? "0.2"),
    height_max_m: String(size.height_max_m ?? "0.6"),
    depth_min_m: String(size.depth_min_m ?? "0.2"),
    depth_max_m: String(size.depth_max_m ?? "0.6"),
    colorMode,
    color: String((col as { color?: string }).color ?? "#ffffff"),
    brightness_pct: String(sh.brightness_pct ?? 100),
    placementMode,
    center_x: String(center.x ?? 0),
    center_y: String(center.y ?? 0),
    center_z: String(center.z ?? 0),
    min_corner_x: String(minC.x ?? 0),
    min_corner_y: String(minC.y ?? 0),
    min_corner_z: String(minC.z ?? 0),
    face: (String(pl.face ?? "front") as FaceId) || "front",
    motion_dx: String(dir.dx ?? 1),
    motion_dy: String(dir.dy ?? 0),
    motion_dz: String(dir.dz ?? 0),
    speedMode,
    m_s: String(sp.m_s ?? 0.1),
    min_m_s: String(sp.min_m_s ?? 0.05),
    max_m_s: String(sp.max_m_s ?? 0.2),
    edge_behavior: ["wrap", "stop", "deflect_random", "deflect_specular"].includes(edge)
      ? edge
      : "wrap",
  };
}

/** Parse API/JSON object into editable form state. */
export function shapeFormStateFromUnknown(raw: unknown): ShapeAnimationFormState {
  const o = raw as Record<string, unknown>;
  if (o.version !== 1) {
    throw new Error("definition version must be 1");
  }
  const bg = (o.background ?? {}) as Record<string, unknown>;
  const bgMode = bg.mode === "lights_off" ? "lights_off" : "lights_on";
  const shapesIn = o.shapes;
  if (!Array.isArray(shapesIn) || shapesIn.length === 0) {
    throw new Error("shapes must be a non-empty array");
  }
  if (shapesIn.length > 20) {
    throw new Error("at most 20 shapes");
  }
  return {
    backgroundMode: bgMode,
    backgroundColor: String(bg.color ?? "#1a1a2e"),
    backgroundBrightness_pct: String(bg.brightness_pct ?? 40),
    shapes: shapesIn.map((s) => rowFromShapeJson(s as Record<string, unknown>)),
  };
}

export function shapeFormStateFromDefault(): ShapeAnimationFormState {
  return shapeFormStateFromUnknown(SHAPE_ANIMATION_DEFAULT_DEFINITION);
}

function buildSize(row: ShapeFormRow): Record<string, unknown> {
  if (row.kind === "sphere") {
    if (row.sizeMode === "fixed") {
      return {
        mode: "fixed",
        radius_m: parsePositive("Sphere radius (m)", row.radius_m),
      };
    }
    const r0 = parsePositive("Sphere min radius (m)", row.radius_min_m);
    const r1 = parsePositive("Sphere max radius (m)", row.radius_max_m);
    if (r0 > r1) {
      throw new Error("Sphere min radius must be ≤ max");
    }
    return {
      mode: "random_uniform",
      radius_min_m: r0,
      radius_max_m: r1,
    };
  }
  if (row.sizeMode === "fixed") {
    return {
      mode: "fixed",
      width_m: parsePositive("Cuboid width (m)", row.width_m),
      height_m: parsePositive("Cuboid height (m)", row.height_m),
      depth_m: parsePositive("Cuboid depth (m)", row.depth_m),
    };
  }
  const w0 = parsePositive("Cuboid min width (m)", row.width_min_m);
  const w1 = parsePositive("Cuboid max width (m)", row.width_max_m);
  const h0 = parsePositive("Cuboid min height (m)", row.height_min_m);
  const h1 = parsePositive("Cuboid max height (m)", row.height_max_m);
  const d0 = parsePositive("Cuboid min depth (m)", row.depth_min_m);
  const d1 = parsePositive("Cuboid max depth (m)", row.depth_max_m);
  if (w0 > w1 || h0 > h1 || d0 > d1) {
    throw new Error("Cuboid random size: each min must be ≤ max");
  }
  return {
    mode: "random_uniform",
    width_min_m: w0,
    width_max_m: w1,
    height_min_m: h0,
    height_max_m: h1,
    depth_min_m: d0,
    depth_max_m: d1,
  };
}

function buildPlacement(row: ShapeFormRow): Record<string, unknown> {
  if (row.placementMode === "random_face") {
    return { mode: "random_face", face: row.face };
  }
  if (row.kind === "sphere") {
    return {
      mode: "fixed",
      center_m: {
        x: parseNum("Center X (m)", row.center_x),
        y: parseNum("Center Y (m)", row.center_y),
        z: parseNum("Center Z (m)", row.center_z),
      },
    };
  }
  return {
    mode: "fixed",
    min_corner_m: {
      x: parseNum("Min corner X (m)", row.min_corner_x),
      y: parseNum("Min corner Y (m)", row.min_corner_y),
      z: parseNum("Min corner Z (m)", row.min_corner_z),
    },
  };
}

function buildMotion(row: ShapeFormRow): Record<string, unknown> {
  const dx = parseNum("Direction dx", row.motion_dx);
  const dy = parseNum("Direction dy", row.motion_dy);
  const dz = parseNum("Direction dz", row.motion_dz);
  if (dx === 0 && dy === 0 && dz === 0) {
    throw new Error("Motion direction cannot be all zero");
  }
  let speed: Record<string, unknown>;
  if (row.speedMode === "fixed") {
    speed = { mode: "fixed", m_s: parsePositive("Speed (m/s)", row.m_s) };
  } else {
    const lo = parsePositive("Min speed (m/s)", row.min_m_s);
    const hi = parsePositive("Max speed (m/s)", row.max_m_s);
    if (lo > hi) {
      throw new Error("Min speed must be ≤ max speed");
    }
    speed = { mode: "random_uniform", min_m_s: lo, max_m_s: hi };
  }
  return {
    direction: { dx, dy, dz },
    speed,
  };
}

/** Build JSON-serializable definition object for PATCH/POST. */
export function definitionObjectFromForm(state: ShapeAnimationFormState): unknown {
  if (state.shapes.length === 0 || state.shapes.length > 20) {
    throw new Error("Need between 1 and 20 shapes");
  }
  const background: Record<string, unknown> =
    state.backgroundMode === "lights_off"
      ? { mode: "lights_off" }
      : {
          mode: "lights_on",
          color: state.backgroundColor.trim() || "#ffffff",
          brightness_pct: parseNum("Background brightness %", state.backgroundBrightness_pct),
        };
  if (state.backgroundMode === "lights_on") {
    const b = background.brightness_pct as number;
    if (b < 0 || b > 100) {
      throw new Error("Background brightness must be 0–100");
    }
  }
  const shapes = state.shapes.map((row) => {
    const br = parseNum("Shape brightness %", row.brightness_pct);
    if (br < 0 || br > 100) {
      throw new Error("Shape brightness must be 0–100");
    }
    const colorSpec: Record<string, unknown> =
      row.colorMode === "random" ? { mode: "random" } : { mode: "fixed", color: row.color.trim() };
    const size = buildSize(row);
    return {
      kind: row.kind,
      size,
      color: colorSpec,
      brightness_pct: br,
      placement: buildPlacement(row),
      motion: buildMotion(row),
      edge_behavior: row.edge_behavior,
    };
  });
  return { version: 1, background, shapes };
}

export function definitionJsonStringFromForm(state: ShapeAnimationFormState): string {
  return JSON.stringify(definitionObjectFromForm(state));
}
