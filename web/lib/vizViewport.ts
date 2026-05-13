import * as THREE from "three";

/** REQ-019: fixed dark-grey WebGL backdrop; independent of shell light/dark (see docs/architecture §4.7). */
export const VIZ_VIEWPORT_BG = 0x262626;

export const VIZ_VIEWPORT_BG_CSS = "#262626";

/**
 * REQ-034 rule 2: fixed boundary cuboid margin applied on each axis (both directions) in the
 * model three.js view. The scene view uses the scene's persisted `margin_m` per REQ-015 BR 12.
 * Keep in sync with `backend.DefaultSceneBoundaryMarginM` (also 0.3 m) so legacy DBs migrating
 * via `ALTER TABLE` and fresh installs render identically on the model side.
 */
export const MODEL_BOUNDARY_MARGIN_M = 0.3;

/** REQ-028: sRGB output + tone mapping so emissive spheres do not blow out (architecture §4.7). */
export function configureVizWebGLRenderer(renderer: THREE.WebGLRenderer): void {
  renderer.outputColorSpace = THREE.SRGBColorSpace;
  renderer.toneMapping = THREE.ACESFilmicToneMapping;
  renderer.toneMappingExposure = 1.12;
}

/** One axis-aligned point (xyz, SI metres). */
export type LightPoint = { x: number; y: number; z: number };

/**
 * Result of {@link boundaryCornersForLights}: the padded AABB used to draw the REQ-034
 * faint boundary cuboid and to drive REQ-016 default-framing distance.
 */
export type BoundaryBox = {
  /** Lower corner after expanding the tight AABB by `margin` on every axis. */
  min: { x: number; y: number; z: number };
  /** Upper corner after expanding the tight AABB by `margin` on every axis. */
  max: { x: number; y: number; z: number };
  /** Centre point of the padded box (used as `Mesh.position` for the LineSegments). */
  center: { x: number; y: number; z: number };
  /** Per-axis extent `max - min`. Always ≥ `2 * margin` even when all lights share a position. */
  size: { x: number; y: number; z: number };
  /** Largest of `size.x/y/z` — convenient for camera-framing math. */
  maxAxis: number;
};

/**
 * REQ-034 boundary geometry helper. Returns the axis-aligned bounding box of `lights`
 * expanded by `margin` on every axis in both directions, plus its centre/size.
 *
 * - `lights` may be empty: the helper returns a degenerate cube `[-margin, +margin]` so callers
 *   can still build a visible cuboid for an empty model view (matches the architecture §3.15
 *   empty-scene policy for dimensions).
 * - Non-finite coordinates are ignored per REQ-034 rule 1 ("ignore non-finite values").
 * - `margin` is treated as `Math.max(0, margin)` defensively; callers must pre-validate finite.
 */
export function boundaryCornersForLights(
  lights: readonly LightPoint[],
  margin: number,
): BoundaryBox {
  const m = Number.isFinite(margin) && margin > 0 ? margin : 0;
  let mnX = Infinity;
  let mnY = Infinity;
  let mnZ = Infinity;
  let mxX = -Infinity;
  let mxY = -Infinity;
  let mxZ = -Infinity;
  let any = false;
  for (const L of lights) {
    if (!Number.isFinite(L.x) || !Number.isFinite(L.y) || !Number.isFinite(L.z)) {
      continue;
    }
    any = true;
    if (L.x < mnX) mnX = L.x;
    if (L.y < mnY) mnY = L.y;
    if (L.z < mnZ) mnZ = L.z;
    if (L.x > mxX) mxX = L.x;
    if (L.y > mxY) mxY = L.y;
    if (L.z > mxZ) mxZ = L.z;
  }
  if (!any) {
    mnX = 0;
    mnY = 0;
    mnZ = 0;
    mxX = 0;
    mxY = 0;
    mxZ = 0;
  }
  const min = { x: mnX - m, y: mnY - m, z: mnZ - m };
  const max = { x: mxX + m, y: mxY + m, z: mxZ + m };
  const size = {
    x: max.x - min.x,
    y: max.y - min.y,
    z: max.z - min.z,
  };
  const center = {
    x: (min.x + max.x) / 2,
    y: (min.y + max.y) / 2,
    z: (min.z + max.z) / 2,
  };
  const maxAxis = Math.max(size.x, size.y, size.z);
  return { min, max, center, size, maxAxis };
}
