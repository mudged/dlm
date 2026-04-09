import * as THREE from "three";

/** REQ-010 / REQ-034: same faint wire as inter-light chain segments. */
const VIZ_GREY = 0xd0d0d0;
const LINE_OPACITY = 0.15;

export function createInterLightWireLineMaterial(): THREE.LineBasicMaterial {
  return new THREE.LineBasicMaterial({
    color: VIZ_GREY,
    transparent: true,
    opacity: LINE_OPACITY,
    depthWrite: false,
  });
}
