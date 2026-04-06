import { describe, expect, it } from "vitest";
import * as THREE from "three";
import {
  colorFromHexAndBrightness,
  emissiveColorFromHex,
  emissiveIntensityFromBrightness,
  meshStandardMaterialForOnLight,
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

describe("emissiveIntensityFromBrightness (REQ-028)", () => {
  it("is zero at 0% and monotonic through 100%", () => {
    expect(emissiveIntensityFromBrightness(0)).toBe(0);
    const a = emissiveIntensityFromBrightness(25);
    const b = emissiveIntensityFromBrightness(50);
    const c = emissiveIntensityFromBrightness(100);
    expect(b).toBeGreaterThan(a);
    expect(c).toBeGreaterThan(b);
  });
});

describe("emissiveColorFromHex", () => {
  it("keeps full chroma unlike brightness-scaled base color", () => {
    const e = emissiveColorFromHex("#ff0000");
    expect(e.r).toBeCloseTo(1, 5);
    const base = colorFromHexAndBrightness("#ff0000", 50);
    expect(base.r).toBeCloseTo(0.5, 5);
  });
});

describe("meshStandardMaterialForOnLight (REQ-028)", () => {
  it("uses emissive tied to hex and intensity from brightness", () => {
    const m = meshStandardMaterialForOnLight("#00ff00", 100);
    expect(m.metalness).toBe(0);
    expect(m.roughness).toBe(0.35);
    expect(m.emissiveIntensity).toBe(emissiveIntensityFromBrightness(100));
    expect(m.emissive.g).toBeGreaterThan(0.9);
    m.dispose();
  });
});
