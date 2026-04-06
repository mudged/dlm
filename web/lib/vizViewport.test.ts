import { describe, expect, it } from "vitest";
import * as THREE from "three";
import {
  configureVizWebGLRenderer,
  VIZ_VIEWPORT_BG,
  VIZ_VIEWPORT_BG_CSS,
} from "./vizViewport";

describe("vizViewport", () => {
  it("uses architecture default REQ-019 hex", () => {
    expect(VIZ_VIEWPORT_BG).toBe(0x262626);
    expect(VIZ_VIEWPORT_BG_CSS).toBe("#262626");
  });

  it("configureVizWebGLRenderer sets REQ-028 output and tone mapping", () => {
    const renderer = {
      outputColorSpace: THREE.NoColorSpace,
      toneMapping: THREE.NoToneMapping,
      toneMappingExposure: 1,
    } as THREE.WebGLRenderer;
    configureVizWebGLRenderer(renderer);
    expect(renderer.outputColorSpace).toBe(THREE.SRGBColorSpace);
    expect(renderer.toneMapping).toBe(THREE.ACESFilmicToneMapping);
    expect(renderer.toneMappingExposure).toBe(1.12);
  });
});
