/**
 * Client-side shape animation (REQ-033 / architecture §3.17.2).
 * Call initShapeAnimationSim once per run/cycle, then tickShapeAnimation each frame.
 */

/** Scene AABB for motion and placement (from GET …/dimensions: origin = min, max = max). */
export type SceneDimensions = {
  min: { x: number; y: number; z: number };
  max: { x: number; y: number; z: number };
};

/** Build sim bounds from API response (backward compatible if origin missing). */
export function sceneDimensionsFromApiResponse(d: {
  origin?: { x: number; y: number; z: number };
  max: { x: number; y: number; z: number };
}): SceneDimensions {
  return {
    min: {
      x: d.origin?.x ?? 0,
      y: d.origin?.y ?? 0,
      z: d.origin?.z ?? 0,
    },
    max: { x: d.max.x, y: d.max.y, z: d.max.z },
  };
}

export type SceneLightFlat = {
  scene_id: string;
  model_id: string;
  light_id: number;
  sx: number;
  sy: number;
  sz: number;
};

export type BatchLightUpdate = {
  model_id: string;
  light_id: number;
  on: boolean;
  color: string;
  brightness_pct: number;
};

type RNG = () => number;

function randomHex(rng: RNG): string {
  const n = Math.floor(rng() * 0x1000000);
  return `#${n.toString(16).padStart(6, "0")}`;
}

function randomUniform(rng: RNG, lo: number, hi: number): number {
  return lo + (hi - lo) * rng();
}

function normalizeDir(dx: number, dy: number, dz: number): { ux: number; uy: number; uz: number } {
  const len = Math.hypot(dx, dy, dz);
  if (len <= 0) {
    return { ux: 1, uy: 0, uz: 0 };
  }
  return { ux: dx / len, uy: dy / len, uz: dz / len };
}

function randomUnitVector(rng: RNG): { ux: number; uy: number; uz: number } {
  const t = 2 * Math.PI * rng();
  const u = 2 * rng() - 1;
  const s = Math.sqrt(Math.max(0, 1 - u * u));
  return { ux: s * Math.cos(t), uy: s * Math.sin(t), uz: u };
}

function reflectSpecular(
  vx: number,
  vy: number,
  vz: number,
  hitX: boolean,
  hitY: boolean,
  hitZ: boolean,
): { vx: number; vy: number; vz: number } {
  let nx = vx;
  let ny = vy;
  let nz = vz;
  if (hitX) nx = -vx;
  if (hitY) ny = -vy;
  if (hitZ) nz = -vz;
  return { vx: nx, vy: ny, vz: nz };
}

type BackgroundSpec = {
  mode: "lights_on" | "lights_off";
  color: string;
  brightness_pct: number;
};

type SimShape = {
  kind: "sphere" | "cuboid";
  edge: string;
  brightness_pct: number;
  colorMode: "fixed" | "random";
  fixedColor: string;
  currentColor: string;
  px: number;
  py: number;
  pz: number;
  radius: number;
  w: number;
  h: number;
  d: number;
  vx: number;
  vy: number;
  vz: number;
  active: boolean;
};

export type ShapeAnimationSim = {
  background: BackgroundSpec;
  shapes: SimShape[];
};

/** Physics step (s). 60 Hz for smoother motion than 20 Hz; architecture default tick. */
const DT = 1 / 60;

export function makeRng(seed: number): RNG {
  let s = seed >>> 0;
  return () => {
    s = (1664525 * s + 1013904223) >>> 0;
    return s / 0x100000000;
  };
}

function initShapesFromDefinition(
  raw: unknown,
  rng: RNG,
  dims: SceneDimensions,
): { background: BackgroundSpec; shapes: SimShape[] } {
  const o = raw as Record<string, unknown>;
  if (o.version !== 1) {
    throw new Error("definition version must be 1");
  }
  const bg = o.background as Record<string, unknown>;
  const bgMode = bg.mode as string;
  let background: BackgroundSpec;
  if (bgMode === "lights_on") {
    background = {
      mode: "lights_on",
      color: String(bg.color ?? "#ffffff"),
      brightness_pct: Number(bg.brightness_pct ?? 100),
    };
  } else {
    background = { mode: "lights_off", color: "#000000", brightness_pct: 0 };
  }

  const minX = dims.min.x;
  const minY = dims.min.y;
  const minZ = dims.min.z;
  const maxX = dims.max.x;
  const maxY = dims.max.y;
  const maxZ = dims.max.z;

  const shapesIn = o.shapes as unknown[];
  const shapes: SimShape[] = [];
  for (const s of shapesIn) {
    const sh = s as Record<string, unknown>;
    const kind = sh.kind as string;
    const edge = String(sh.edge_behavior);
    const br = Number(sh.brightness_pct);
    const col = sh.color as Record<string, unknown>;
    const colorMode = col.mode === "random" ? "random" : "fixed";
    const fixedColor =
      colorMode === "fixed" ? String((col as { color?: string }).color ?? "#ffffff") : "#ffffff";
    const currentColor = colorMode === "random" ? randomHex(rng) : fixedColor;

    const sz = sh.size as Record<string, unknown>;
    let radius = 0.1;
    let w = 0.2;
    let h = 0.2;
    let d = 0.2;
    if (sz.mode === "fixed") {
      if (kind === "sphere") {
        radius = Number(sz.radius_m);
      } else {
        w = Number(sz.width_m);
        h = Number(sz.height_m);
        d = Number(sz.depth_m);
      }
    } else {
      if (kind === "sphere") {
        radius = randomUniform(rng, Number(sz.radius_min_m), Number(sz.radius_max_m));
      } else {
        w = randomUniform(rng, Number(sz.width_min_m), Number(sz.width_max_m));
        h = randomUniform(rng, Number(sz.height_min_m), Number(sz.height_max_m));
        d = randomUniform(rng, Number(sz.depth_min_m), Number(sz.depth_max_m));
      }
    }

    const mot = sh.motion as Record<string, unknown>;
    const dir = mot.direction as Record<string, unknown>;
    const { ux, uy, uz } = normalizeDir(
      Number(dir.dx),
      Number(dir.dy),
      Number(dir.dz),
    );
    const sp = mot.speed as Record<string, unknown>;
    let speed = 0.1;
    if (sp.mode === "fixed") {
      speed = Number(sp.m_s);
    } else {
      speed = randomUniform(rng, Number(sp.min_m_s), Number(sp.max_m_s));
    }

    let px = 0;
    let py = 0;
    let pz = 0;
    const pl = sh.placement as Record<string, unknown>;
    if (pl.mode === "fixed") {
      if (kind === "sphere") {
        const c = pl.center_m as Record<string, unknown>;
        px = Number(c.x);
        py = Number(c.y);
        pz = Number(c.z);
      } else {
        const c = pl.min_corner_m as Record<string, unknown>;
        px = Number(c.x);
        py = Number(c.y);
        pz = Number(c.z);
      }
    } else {
      const face = String(pl.face);
      if (kind === "sphere") {
        if (face === "left") {
          px = minX + radius + 1e-6;
          py = randomUniform(rng, minY + radius, maxY - radius);
          pz = randomUniform(rng, minZ + radius, maxZ - radius);
        } else if (face === "right") {
          px = maxX - radius - 1e-6;
          py = randomUniform(rng, minY + radius, maxY - radius);
          pz = randomUniform(rng, minZ + radius, maxZ - radius);
        } else if (face === "bottom") {
          py = minY + radius + 1e-6;
          px = randomUniform(rng, minX + radius, maxX - radius);
          pz = randomUniform(rng, minZ + radius, maxZ - radius);
        } else if (face === "top") {
          py = maxY - radius - 1e-6;
          px = randomUniform(rng, minX + radius, maxX - radius);
          pz = randomUniform(rng, minZ + radius, maxZ - radius);
        } else if (face === "back") {
          pz = minZ + radius + 1e-6;
          px = randomUniform(rng, minX + radius, maxX - radius);
          py = randomUniform(rng, minY + radius, maxY - radius);
        } else {
          pz = maxZ - radius - 1e-6;
          px = randomUniform(rng, minX + radius, maxX - radius);
          py = randomUniform(rng, minY + radius, maxY - radius);
        }
      } else {
        if (face === "left") {
          px = minX;
          py = randomUniform(rng, minY, Math.max(minY, maxY - h));
          pz = randomUniform(rng, minZ, Math.max(minZ, maxZ - d));
        } else if (face === "right") {
          px = Math.max(minX, maxX - w);
          py = randomUniform(rng, minY, Math.max(minY, maxY - h));
          pz = randomUniform(rng, minZ, Math.max(minZ, maxZ - d));
        } else if (face === "bottom") {
          py = minY;
          px = randomUniform(rng, minX, Math.max(minX, maxX - w));
          pz = randomUniform(rng, minZ, Math.max(minZ, maxZ - d));
        } else if (face === "top") {
          py = Math.max(minY, maxY - h);
          px = randomUniform(rng, minX, Math.max(minX, maxX - w));
          pz = randomUniform(rng, minZ, Math.max(minZ, maxZ - d));
        } else if (face === "back") {
          pz = minZ;
          px = randomUniform(rng, minX, Math.max(minX, maxX - w));
          py = randomUniform(rng, minY, Math.max(minY, maxY - h));
        } else {
          pz = Math.max(minZ, maxZ - d);
          px = randomUniform(rng, minX, Math.max(minX, maxX - w));
          py = randomUniform(rng, minY, Math.max(minY, maxY - h));
        }
      }
    }

    shapes.push({
      kind: kind as "sphere" | "cuboid",
      edge,
      brightness_pct: br,
      colorMode,
      fixedColor,
      currentColor,
      px,
      py,
      pz,
      radius,
      w,
      h,
      d,
      vx: ux * speed,
      vy: uy * speed,
      vz: uz * speed,
      active: true,
    });
  }
  return { background, shapes };
}

/** Create simulation state for a run or loop cycle (new random draws). Reuse returned rng for tickShapeAnimationSim. */
export function initShapeAnimationSim(
  definitionJson: string,
  dims: SceneDimensions,
  rng: RNG,
): ShapeAnimationSim {
  const def = JSON.parse(definitionJson) as unknown;
  return initShapesFromDefinition(def, rng, dims);
}

function integrateShape(
  s: SimShape,
  minX: number,
  minY: number,
  minZ: number,
  maxX: number,
  maxY: number,
  maxZ: number,
  rng: RNG,
): void {
  if (!s.active) {
    return;
  }
  let px = s.px + s.vx * DT;
  let py = s.py + s.vy * DT;
  let pz = s.pz + s.vz * DT;
  const vx0 = s.vx;
  const vy0 = s.vy;
  const vz0 = s.vz;

  const spanX = Math.max(maxX - minX, 1e-9);
  const spanY = Math.max(maxY - minY, 1e-9);
  const spanZ = Math.max(maxZ - minZ, 1e-9);

  let hitX = false;
  let hitY = false;
  let hitZ = false;

  if (s.kind === "sphere") {
    const r = s.radius;
    if (px - r < minX || px + r > maxX) hitX = true;
    if (py - r < minY || py + r > maxY) hitY = true;
    if (pz - r < minZ || pz + r > maxZ) hitZ = true;
  } else {
    if (px < minX || px + s.w > maxX) hitX = true;
    if (py < minY || py + s.h > maxY) hitY = true;
    if (pz < minZ || pz + s.d > maxZ) hitZ = true;
  }

  const violated = hitX || hitY || hitZ;

  if (!violated) {
    s.px = px;
    s.py = py;
    s.pz = pz;
    return;
  }

  if (s.edge === "stop") {
    s.active = false;
    return;
  }
  if (s.edge === "wrap") {
    if (s.kind === "sphere") {
      const r = s.radius;
      while (px - r < minX) px += spanX;
      while (px + r > maxX) px -= spanX;
      while (py - r < minY) py += spanY;
      while (py + r > maxY) py -= spanY;
      while (pz - r < minZ) pz += spanZ;
      while (pz + r > maxZ) pz -= spanZ;
    } else {
      while (px < minX) px += spanX;
      while (px + s.w > maxX) px -= spanX;
      while (py < minY) py += spanY;
      while (py + s.h > maxY) py -= spanY;
      while (pz < minZ) pz += spanZ;
      while (pz + s.d > maxZ) pz -= spanZ;
    }
    s.px = px;
    s.py = py;
    s.pz = pz;
    return;
  }
  if (s.edge === "deflect_random") {
    const sp = Math.hypot(vx0, vy0, vz0);
    const u = randomUnitVector(rng);
    s.vx = u.ux * sp;
    s.vy = u.uy * sp;
    s.vz = u.uz * sp;
    return;
  }
  if (s.edge === "deflect_specular") {
    const r = reflectSpecular(vx0, vy0, vz0, hitX, hitY, hitZ);
    s.vx = r.vx;
    s.vy = r.vy;
    s.vz = r.vz;
  }
}

function lightInShape(L: SceneLightFlat, s: SimShape): boolean {
  if (!s.active) {
    return false;
  }
  const { sx, sy, sz } = L;
  if (s.kind === "sphere") {
    const d = Math.hypot(sx - s.px, sy - s.py, sz - s.pz);
    return d <= s.radius + 1e-9;
  }
  return (
    sx >= s.px - 1e-9 &&
    sx <= s.px + s.w + 1e-9 &&
    sy >= s.py - 1e-9 &&
    sy <= s.py + s.h + 1e-9 &&
    sz >= s.pz - 1e-9 &&
    sz <= s.pz + s.d + 1e-9
  );
}

/** Advance physics one step. Pass rng only for deflect_random (use same seeded rng per run). */
export function tickShapeAnimationSim(
  sim: ShapeAnimationSim,
  dims: SceneDimensions,
  rng: RNG,
): { allShapesStopped: boolean } {
  const { min, max } = dims;
  for (const s of sim.shapes) {
    integrateShape(s, min.x, min.y, min.z, max.x, max.y, max.z, rng);
  }
  const allStopped = sim.shapes.every((s) => !s.active);
  return { allShapesStopped: allStopped };
}

/** Build per-light updates from current sim + light list. */
export function buildBatchUpdatesFromSim(
  sim: ShapeAnimationSim,
  lights: SceneLightFlat[],
): BatchLightUpdate[] {
  const updates: BatchLightUpdate[] = [];
  const { background, shapes } = sim;
  for (const L of lights) {
    let winner = -1;
    for (let i = 0; i < shapes.length; i++) {
      if (lightInShape(L, shapes[i]!)) {
        winner = i;
        break;
      }
    }
    if (winner >= 0) {
      const sh = shapes[winner]!;
      updates.push({
        model_id: L.model_id,
        light_id: L.light_id,
        on: true,
        color: sh.currentColor,
        brightness_pct: sh.brightness_pct,
      });
    } else if (background.mode === "lights_on") {
      updates.push({
        model_id: L.model_id,
        light_id: L.light_id,
        on: true,
        color: background.color,
        brightness_pct: background.brightness_pct,
      });
    } else {
      updates.push({
        model_id: L.model_id,
        light_id: L.light_id,
        on: false,
        color: "#ffffff",
        brightness_pct: 100,
      });
    }
  }
  return updates;
}

export const SHAPE_ANIMATION_DT_SEC = DT;

/** Semi-transparent editor/run overlay meshes (scene space). */
export type GhostShapeOverlay = {
  kind: "sphere" | "cuboid";
  px: number;
  py: number;
  pz: number;
  radius: number;
  w: number;
  h: number;
  d: number;
  color: string;
};

function overlaysFromSim(sim: ShapeAnimationSim): GhostShapeOverlay[] {
  return sim.shapes
    .filter((s) => s.active)
    .map((s) => ({
      kind: s.kind,
      px: s.px,
      py: s.py,
      pz: s.pz,
      radius: s.radius,
      w: s.w,
      h: s.h,
      d: s.d,
      color: s.currentColor,
    }));
}

/** Initial shape positions/sizes for the shape routine editor preview (deterministic RNG). */
export function ghostShapesFromDefinition(
  definitionJson: string,
  dims: SceneDimensions,
): GhostShapeOverlay[] {
  const rng = makeRng(1);
  try {
    const sim = initShapeAnimationSim(definitionJson, dims, rng);
    return overlaysFromSim(sim);
  } catch {
    return [];
  }
}

/** Live positions during an active run (updated each tick by `ShapeAnimationRoutineHost`). */
export function ghostOverlaysFromSim(sim: ShapeAnimationSim): GhostShapeOverlay[] {
  return overlaysFromSim(sim);
}
