import type { Light } from "@/lib/models";

/** Center and largest axis-aligned extent for framing a camera around light positions. */
export function boundingFromLights(lights: Light[]): {
  center: [number, number, number];
  maxDim: number;
} {
  if (lights.length === 0) {
    return { center: [0, 0, 0], maxDim: 1 };
  }
  let minX = lights[0].x;
  let maxX = lights[0].x;
  let minY = lights[0].y;
  let maxY = lights[0].y;
  let minZ = lights[0].z;
  let maxZ = lights[0].z;
  for (let i = 1; i < lights.length; i++) {
    const L = lights[i];
    minX = Math.min(minX, L.x);
    maxX = Math.max(maxX, L.x);
    minY = Math.min(minY, L.y);
    maxY = Math.max(maxY, L.y);
    minZ = Math.min(minZ, L.z);
    maxZ = Math.max(maxZ, L.z);
  }
  const cx = (minX + maxX) / 2;
  const cy = (minY + maxY) / 2;
  const cz = (minZ + maxZ) / 2;
  const maxDim = Math.max(maxX - minX, maxY - minY, maxZ - minZ, 1e-9);
  return { center: [cx, cy, cz], maxDim };
}
