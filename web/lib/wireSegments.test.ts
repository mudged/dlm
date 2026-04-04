import { describe, expect, it } from "vitest";
import { buildWireSegmentPositions, SPHERE_RADIUS_M } from "./wireSegments";

describe("buildWireSegmentPositions", () => {
  it("returns empty array for 0 or 1 light", () => {
    expect(buildWireSegmentPositions([]).length).toBe(0);
    expect(buildWireSegmentPositions([{ id: 0, x: 0, y: 0, z: 0 }]).length).toBe(
      0,
    );
  });

  it("connects consecutive sorted ids even if array order differs", () => {
    const positions = buildWireSegmentPositions([
      { id: 1, x: 1, y: 0, z: 0 },
      { id: 0, x: 0, y: 0, z: 0 },
    ]);
    expect(positions.length).toBe(6);
    expect(Array.from(positions)).toEqual([
      0, 0, 0, 1, 0, 0,
    ]);
  });

  it("builds a chain for three lights", () => {
    const positions = buildWireSegmentPositions([
      { id: 0, x: 0, y: 0, z: 0 },
      { id: 1, x: 1, y: 0, z: 0 },
      { id: 2, x: 1, y: 1, z: 0 },
    ]);
    expect(positions.length).toBe(12);
  });
});

describe("SPHERE_RADIUS_M", () => {
  it("is half of 2 cm diameter", () => {
    expect(SPHERE_RADIUS_M).toBe(0.01);
  });
});
