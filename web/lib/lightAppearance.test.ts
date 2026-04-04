import { describe, expect, it } from "vitest";
import * as THREE from "three";
import {
  colorFromHexAndBrightness,
  normalizeLightHex,
} from "./lightAppearance";

describe("normalizeLightHex", () => {
  it("adds # for six hex digits", () => {
    expect(normalizeLightHex("ff0000")).toBe("#ff0000");
  });

  it("lowercases canonical input", () => {
    expect(normalizeLightHex("#00AAFF")).toBe("#00aaff");
  });

  it("falls back for invalid input", () => {
    expect(normalizeLightHex("")).toBe("#ffffff");
    expect(normalizeLightHex("nope")).toBe("#ffffff");
  });
});

describe("colorFromHexAndBrightness", () => {
  it("scales white at 50% to mid gray linear", () => {
    const c = colorFromHexAndBrightness("#ffffff", 50);
    expect(c.r).toBeCloseTo(0.5, 5);
    expect(c.g).toBeCloseTo(0.5, 5);
    expect(c.b).toBeCloseTo(0.5, 5);
  });

  it("clamps brightness beyond 0–100", () => {
    const lo = colorFromHexAndBrightness("#ff0000", -10);
    const hi = colorFromHexAndBrightness("#ff0000", 200);
    expect(lo.r + lo.g + lo.b).toBe(0);
    expect(hi.equals(new THREE.Color(1, 0, 0))).toBe(true);
  });
});
