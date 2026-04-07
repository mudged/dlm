import { describe, expect, it } from "vitest";
import {
  PYTHON_ROUTINE_DEFAULT_SOURCE,
  PYTHON_SCENE_API_CATALOG_FULL,
  SCENE_API_CATALOG,
} from "./pythonSceneApiCatalog";

describe("pythonSceneApiCatalog", () => {
  it("includes width height depth max_* and all worker methods", () => {
    const labels = SCENE_API_CATALOG.map((e) => e.label);
    expect(labels).toContain("scene.width");
    expect(labels).toContain("scene.height");
    expect(labels).toContain("scene.depth");
    expect(labels).toContain("scene.max_x");
    expect(labels).toContain("scene.max_y");
    expect(labels).toContain("scene.max_z");
    expect(labels).toContain("random_hex_colour");
    expect(labels).toContain("set_lights_in_sphere");
    expect(labels).toContain("update_lights_batch");
  });

  it("full catalog appends REQ-032 sample entries without duplicating API rows", () => {
    const apiIds = new Set(SCENE_API_CATALOG.map((e) => e.id));
    const full = PYTHON_SCENE_API_CATALOG_FULL;
    expect(full.length).toBeGreaterThan(SCENE_API_CATALOG.length);
    const sample = full.filter((e) => e.kind === "sample");
    expect(sample.map((s) => s.id)).toEqual([
      "sample-growing-sphere",
      "sample-sweeping-cuboid",
      "sample-random-colour-cycle",
    ]);
    for (const e of SCENE_API_CATALOG) {
      expect(full.some((x) => x.id === e.id)).toBe(true);
    }
    for (const s of sample) {
      expect(apiIds.has(s.id)).toBe(false);
    }
  });

  it("default template demonstrates sphere colour update", () => {
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("set_lights_in_sphere");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("await");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("scene.random_hex_colour()");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).not.toMatch(/^\s*import random\s*$/m);
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE.split("\n").filter((l) => l.trim().startsWith("#")).length).toBeGreaterThanOrEqual(3);
  });

  it("each API catalog snippet includes at least one Python comment for learners", () => {
    for (const e of SCENE_API_CATALOG) {
      expect(e.snippet, e.id).toMatch(/^\s*#/m);
    }
  });

  it("REQ-032 sample snippets are commented and use scene API only", () => {
    const samples = PYTHON_SCENE_API_CATALOG_FULL.filter((e) => e.kind === "sample");
    for (const e of samples) {
      expect(e.snippet.split("\n").filter((l) => l.trim().startsWith("#")).length).toBeGreaterThanOrEqual(
        3,
      );
      expect(e.snippet).toContain("scene.random_hex_colour()");
    }
    expect(samples[0]!.snippet).toContain("set_lights_in_sphere");
    expect(samples[1]!.snippet).toContain("get_lights_within_cuboid");
    expect(samples[1]!.snippet).toContain("update_lights_batch");
  });
});
