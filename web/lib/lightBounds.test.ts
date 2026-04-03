import { describe, expect, it } from "vitest";
import { boundingFromLights } from "./lightBounds";

describe("boundingFromLights", () => {
  it("returns default frame when there are no lights", () => {
    expect(boundingFromLights([])).toEqual({
      center: [0, 0, 0],
      maxDim: 1,
    });
  });

  it("frames a single point with a small non-zero extent", () => {
    const b = boundingFromLights([{ id: 0, x: 5, y: -2, z: 1 }]);
    expect(b.center).toEqual([5, -2, 1]);
    expect(b.maxDim).toBe(1e-9);
  });

  it("uses the largest axis span across multiple lights", () => {
    const b = boundingFromLights([
      { id: 0, x: 0, y: 0, z: 0 },
      { id: 1, x: 3, y: 0, z: 0 },
      { id: 2, x: 0, y: 0, z: 1 },
    ]);
    expect(b.center[0]).toBeCloseTo(1.5);
    expect(b.center[1]).toBe(0);
    expect(b.center[2]).toBeCloseTo(0.5);
    expect(b.maxDim).toBe(3);
  });
});
