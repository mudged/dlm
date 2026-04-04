import * as THREE from "three";

/** Normalize API/UI hex to a string `THREE.Color` accepts (`#RRGGBB`). */
export function normalizeLightHex(hex: string): string {
  const t = hex.trim();
  if (/^#[0-9A-Fa-f]{6}$/.test(t)) {
    return t.toLowerCase();
  }
  if (/^[0-9A-Fa-f]{6}$/.test(t)) {
    return `#${t.toLowerCase()}`;
  }
  return "#ffffff";
}

/** Parse #RRGGBB and apply brightness 0–100% as linear RGB scale (architecture §4.7). */
export function colorFromHexAndBrightness(
  hex: string,
  brightnessPct: number,
): THREE.Color {
  const c = new THREE.Color();
  c.set(normalizeLightHex(hex));
  const s = Math.max(0, Math.min(100, brightnessPct)) / 100;
  c.multiplyScalar(s);
  return c;
}
