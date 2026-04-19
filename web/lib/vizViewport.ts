import * as THREE from "three";

/** REQ-019: fixed dark-grey WebGL backdrop; independent of shell light/dark (see docs/agentic-development/architecture §4.7). */
export const VIZ_VIEWPORT_BG = 0x262626;

export const VIZ_VIEWPORT_BG_CSS = "#262626";

/** REQ-028: sRGB output + tone mapping so emissive spheres do not blow out (architecture §4.7). */
export function configureVizWebGLRenderer(renderer: THREE.WebGLRenderer): void {
  renderer.outputColorSpace = THREE.SRGBColorSpace;
  renderer.toneMapping = THREE.ACESFilmicToneMapping;
  renderer.toneMappingExposure = 1.12;
}
