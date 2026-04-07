import { describe, expect, it } from "vitest";
import {
  PYTHON_ROUTINE_DEFAULT_SOURCE,
  SCENE_API_CATALOG,
} from "./pythonSceneApiCatalog";

describe("pythonSceneApiCatalog", () => {
  it("includes width height depth and all worker methods", () => {
    const labels = SCENE_API_CATALOG.map((e) => e.label);
    expect(labels).toContain("scene.width");
    expect(labels).toContain("scene.height");
    expect(labels).toContain("scene.depth");
    expect(labels).toContain("random_hex_colour");
    expect(labels).toContain("set_lights_in_sphere");
    expect(labels).toContain("update_lights_batch");
  });

  it("default template demonstrates sphere colour update", () => {
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("set_lights_in_sphere");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("await");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("scene.random_hex_colour()");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).not.toMatch(/^\s*import random\s*$/m);
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE.split("\n").filter((l) => l.trim().startsWith("#")).length).toBeGreaterThanOrEqual(3);
  });

  it("each catalog snippet includes at least one Python comment for learners", () => {
    for (const e of SCENE_API_CATALOG) {
      expect(e.snippet, e.id).toMatch(/^\s*#/m);
    }
  });
});
