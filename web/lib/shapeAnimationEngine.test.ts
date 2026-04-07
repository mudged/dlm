import { describe, expect, it } from "vitest";
import {
  buildBatchUpdatesFromSim,
  initShapeAnimationSim,
  makeRng,
  tickShapeAnimationSim,
} from "./shapeAnimationEngine";
import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "./shapeAnimationDefault";

describe("shapeAnimationEngine", () => {
  it("assigns inside sphere and background outside", () => {
    const def = JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const dims = { max: { x: 4, y: 4, z: 4 } };
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

  it("tick advances without throwing", () => {
    const def = JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const dims = { max: { x: 4, y: 4, z: 4 } };
    const rng = makeRng(1);
    const sim = initShapeAnimationSim(def, dims, rng);
    for (let i = 0; i < 5; i++) {
      tickShapeAnimationSim(sim, dims, rng);
    }
    expect(sim.shapes[0]!.active).toBe(true);
  });
});
