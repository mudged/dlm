import { describe, expect, it } from "vitest";
import { paddedAabbWireframePositions } from "./aabbWireframe";

describe("paddedAabbWireframePositions", () => {
  it("returns null for empty points", () => {
    expect(paddedAabbWireframePositions([], 0.3)).toBeNull();
  });

  it("expands single point by pad on all sides", () => {
    const pos = paddedAabbWireframePositions([{ x: 1, y: 2, z: 3 }], 0.3)!;
    expect(pos.length).toBe(12 * 2 * 3);
    // min (0.7,1.7,2.7) max (1.3,2.3,3.3) — one edge from min corner
    expect(pos[0]).toBeCloseTo(0.7);
    expect(pos[1]).toBeCloseTo(1.7);
    expect(pos[2]).toBeCloseTo(2.7);
    expect(pos[3]).toBeCloseTo(1.3);
    expect(pos[4]).toBeCloseTo(1.7);
    expect(pos[5]).toBeCloseTo(2.7);
  });
});
