import type { Light } from "@/lib/models";

/** Sphere radius (m); diameter = 2 cm per REQ-010. */
export const SPHERE_RADIUS_M = 0.01;

/** Vertex pairs for LineSegments: (x0,y0,z0,x1,y1,z1, ...) along ascending `id`. */
export function buildWireSegmentPositions(lights: Light[]): Float32Array {
  if (lights.length < 2) {
    return new Float32Array(0);
  }
  const sorted = [...lights].sort((a, b) => a.id - b.id);
  const out = new Float32Array((sorted.length - 1) * 6);
  let o = 0;
  for (let i = 0; i < sorted.length - 1; i++) {
    const a = sorted[i];
    const b = sorted[i + 1];
    out[o++] = a.x;
    out[o++] = a.y;
    out[o++] = a.z;
    out[o++] = b.x;
    out[o++] = b.y;
    out[o++] = b.z;
  }
  return out;
}
