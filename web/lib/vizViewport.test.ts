import { describe, expect, it } from "vitest";
import * as THREE from "three";
import {
  MODEL_BOUNDARY_MARGIN_M,
  VIZ_VIEWPORT_BG,
  VIZ_VIEWPORT_BG_CSS,
  boundaryCornersForLights,
  configureVizWebGLRenderer,
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

  it("MODEL_BOUNDARY_MARGIN_M matches REQ-034 rule 2 fixed model padding", () => {
    expect(MODEL_BOUNDARY_MARGIN_M).toBe(0.3);
  });
});

describe("boundaryCornersForLights", () => {
  it("expands a tight AABB symmetrically on every axis (REQ-034 rule 2)", () => {
    const box = boundaryCornersForLights(
      [
        { x: -1, y: 0, z: 2 },
        { x: 1, y: 5, z: 4 },
        { x: 0.5, y: 2, z: 3 },
      ],
      0.3,
    );
    expect(box.min).toEqual({ x: -1.3, y: -0.3, z: 1.7 });
    expect(box.max).toEqual({ x: 1.3, y: 5.3, z: 4.3 });
    expect(box.size.x).toBeCloseTo(2.6, 9);
    expect(box.size.y).toBeCloseTo(5.6, 9);
    expect(box.size.z).toBeCloseTo(2.6, 9);
    expect(box.maxAxis).toBeCloseTo(5.6, 9);
    expect(box.center).toEqual({ x: 0, y: 2.5, z: 3 });
  });

  it("returns a 2*margin cube for an empty light list (REQ-034 rule 1 degenerate case)", () => {
    const box = boundaryCornersForLights([], 0.3);
    expect(box.min).toEqual({ x: -0.3, y: -0.3, z: -0.3 });
    expect(box.max).toEqual({ x: 0.3, y: 0.3, z: 0.3 });
    expect(box.size).toEqual({ x: 0.6, y: 0.6, z: 0.6 });
  });

  it("collapses to a 2*margin cube when all lights share a position", () => {
    const box = boundaryCornersForLights(
      [
        { x: 1, y: 1, z: 1 },
        { x: 1, y: 1, z: 1 },
        { x: 1, y: 1, z: 1 },
      ],
      0.5,
    );
    expect(box.center).toEqual({ x: 1, y: 1, z: 1 });
    expect(box.size.x).toBeCloseTo(1, 9);
    expect(box.size.y).toBeCloseTo(1, 9);
    expect(box.size.z).toBeCloseTo(1, 9);
  });

  it("treats zero margin as a tight box (still finite size)", () => {
    const box = boundaryCornersForLights(
      [
        { x: 0, y: 0, z: 0 },
        { x: 2, y: 4, z: 8 },
      ],
      0,
    );
    expect(box.min).toEqual({ x: 0, y: 0, z: 0 });
    expect(box.max).toEqual({ x: 2, y: 4, z: 8 });
    expect(box.maxAxis).toBe(8);
  });

  it("ignores non-finite light coordinates (REQ-034 rule 1)", () => {
    const box = boundaryCornersForLights(
      [
        { x: 0, y: 0, z: 0 },
        { x: Number.NaN, y: 100, z: 100 },
        { x: 2, y: 1, z: 1 },
        { x: Number.POSITIVE_INFINITY, y: 1, z: 1 },
      ],
      0.1,
    );
    expect(box.min).toEqual({ x: -0.1, y: -0.1, z: -0.1 });
    expect(box.max).toEqual({ x: 2.1, y: 1.1, z: 1.1 });
  });

  it("clamps a negative or non-finite margin to zero", () => {
    const box = boundaryCornersForLights(
      [{ x: 0, y: 0, z: 0 }],
      -1,
    );
    expect(box.size).toEqual({ x: 0, y: 0, z: 0 });
    const nanBox = boundaryCornersForLights(
      [{ x: 0, y: 0, z: 0 }],
      Number.NaN,
    );
    expect(nanBox.size).toEqual({ x: 0, y: 0, z: 0 });
  });
});
