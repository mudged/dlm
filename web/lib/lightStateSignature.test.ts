import { describe, expect, it } from "vitest";
import { modelLightsVizSignature } from "./lightStateSignature";
import type { Light } from "./models";

describe("modelLightsVizSignature", () => {
  const base: Light[] = [
    { id: 0, x: 0, y: 0, z: 0, on: false, color: "#ffffff", brightness_pct: 100 },
    { id: 1, x: 1, y: 0, z: 0, on: false, color: "#ffffff", brightness_pct: 100 },
  ];

  it("is stable for same logical state with different object identity", () => {
    const a = modelLightsVizSignature(base);
    const b = modelLightsVizSignature(
      base.map((L) => ({ ...L, color: "#FFFFFF" })),
    );
    expect(a).toBe(b);
  });

  it("changes when effective state changes", () => {
    const a = modelLightsVizSignature(base);
    const b = modelLightsVizSignature(
      base.map((L, i) => (i === 0 ? { ...L, on: true } : L)),
    );
    expect(a).not.toBe(b);
  });
});
