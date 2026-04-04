import * as THREE from "three";

/** Parse #RRGGBB and apply brightness 0–100% as linear RGB scale (architecture §4.7). */
export function colorFromHexAndBrightness(
  hex: string,
  brightnessPct: number,
): THREE.Color {
  const c = new THREE.Color(hex);
  const s = Math.max(0, Math.min(100, brightnessPct)) / 100;
  c.multiplyScalar(s);
  return c;
}
