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

/** Full-chroma linear colour from hex for emissive tint (REQ-028); intensity scales with brightness separately. */
export function emissiveColorFromHex(hex: string): THREE.Color {
  const c = new THREE.Color();
  c.set(normalizeLightHex(hex));
  return c;
}

/** REQ-028: monotonic emissive strength vs brightness 0–100%; tuned for dark-grey viewport (§4.7). */
export function emissiveIntensityFromBrightness(brightnessPct: number): number {
  const b = Math.max(0, Math.min(100, brightnessPct)) / 100;
  // k = emissiveIntensity at 100%; cap reduces blow-out when many lights are fully on.
  const k = 5.5;
  const cap = 8;
  return Math.min(k * b, cap);
}

/** §4.7 / REQ-028: standard material with base albedo × brightness and emissive glow × brightness. */
export function meshStandardMaterialForOnLight(
  hex: string,
  brightnessPct: number,
): THREE.MeshStandardMaterial {
  return new THREE.MeshStandardMaterial({
    color: colorFromHexAndBrightness(hex, brightnessPct),
    emissive: emissiveColorFromHex(hex),
    emissiveIntensity: emissiveIntensityFromBrightness(brightnessPct),
    metalness: 0,
    roughness: 0.25,
  });
}

/**
 * Additive outer shell so on-lights read as glowing (tone mapping dulls emissive-only cues).
 * `toneMapped: false` keeps the halo vivid; not used for picking.
 */
export function additiveGlowShellMaterialForOnLight(
  hex: string,
  brightnessPct: number,
): THREE.MeshBasicMaterial {
  const b = Math.max(0, Math.min(100, brightnessPct)) / 100;
  const col = emissiveColorFromHex(hex).clone();
  col.multiplyScalar(0.5 + 1.1 * b);
  return new THREE.MeshBasicMaterial({
    color: col,
    transparent: true,
    opacity: 0.22 + 0.68 * b,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
    toneMapped: false,
  });
}
