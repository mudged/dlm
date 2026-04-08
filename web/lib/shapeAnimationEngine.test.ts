import { describe, expect, it } from "vitest";
import {
  buildBatchUpdatesFromSim,
  ghostShapesFromDefinition,
  initShapeAnimationSim,
  makeRng,
  sceneDimensionsFromApiResponse,
  tickShapeAnimationSim,
} from "./shapeAnimationEngine";
import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "./shapeAnimationDefault";

describe("shapeAnimationEngine", () => {
  it("assigns inside sphere and background outside", () => {
    const def = JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const dims = sceneDimensionsFromApiResponse({
      origin: { x: 0, y: 0, z: 0 },
      max: { x: 4, y: 4, z: 4 },
    });
    const rng = makeRng(42);
    const sim = initShapeAnimationSim(def, dims, rng);
    const lights = [
      {
        scene_id: "s",
        model_id: "m",
        light_id: 0,
        sx: 1,
        sy: 1,
        sz: 1,
      },
      {
        scene_id: "s",
        model_id: "m",
        light_id: 1,
        sx: 3,
        sy: 3,
        sz: 3,
      },
    ];
    const updates = buildBatchUpdatesFromSim(sim, lights);
    expect(updates).toHaveLength(2);
    expect(updates[0]!.on && updates[0]!.color === sim.shapes[0]!.currentColor).toBe(
      true,
    );
    expect(updates[1]!.color).toBe(sim.background.color);
  });

  it("ghostShapesFromDefinition returns overlays for default definition", () => {
    const def = JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const dims = sceneDimensionsFromApiResponse({
      origin: { x: 0, y: 0, z: 0 },
      max: { x: 4, y: 4, z: 4 },
    });
    const ghosts = ghostShapesFromDefinition(def, dims);
    expect(ghosts.length).toBe(1);
    expect(ghosts[0]!.kind).toBe("sphere");
    expect(ghosts[0]!.radius).toBeGreaterThan(0);
  });

  it("uses negative origin when API returns padded min below zero (matches REQ-034 wireframe)", () => {
    const dims = sceneDimensionsFromApiResponse({
      origin: { x: 0, y: -0.3, z: -0.3 },
      max: { x: 3, y: 2, z: 2 },
    });
    expect(dims.min.y).toBe(-0.3);
    expect(dims.min.z).toBe(-0.3);
  });

  it("uses non-zero origin for axis-aligned scene bounds (random_face right)", () => {
    const def = JSON.stringify({
      version: 1,
      background: { mode: "lights_off" },
      shapes: [
        {
          kind: "sphere",
          size: { mode: "fixed", radius_m: 0.2 },
          color: { mode: "fixed", color: "#ffffff" },
          brightness_pct: 100,
          placement: { mode: "random_face", face: "right" },
          motion: {
            direction: { dx: 1, dy: 0, dz: 0 },
            speed: { mode: "fixed", m_s: 0.1 },
          },
          edge_behavior: "wrap",
        },
      ],
    });
    const dims = sceneDimensionsFromApiResponse({
      origin: { x: 9.7, y: 0, z: 0 },
      max: { x: 12.3, y: 2.3, z: 2.3 },
    });
    const rng = makeRng(1);
    const sim = initShapeAnimationSim(def, dims, rng);
    expect(sim.shapes[0]!.px).toBeGreaterThan(11.9);
    expect(sim.shapes[0]!.px).toBeLessThanOrEqual(12.3);
  });

  it("tick advances without throwing", () => {
    const def = JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const dims = sceneDimensionsFromApiResponse({
      origin: { x: 0, y: 0, z: 0 },
      max: { x: 4, y: 4, z: 4 },
    });
    const rng = makeRng(1);
    const sim = initShapeAnimationSim(def, dims, rng);
    for (let i = 0; i < 5; i++) {
      tickShapeAnimationSim(sim, dims, rng);
    }
    expect(sim.shapes[0]!.active).toBe(true);
  });
});
