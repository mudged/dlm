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
    expect(labels).toContain("set_lights_in_sphere");
    expect(labels).toContain("update_lights_batch");
  });

  it("default template demonstrates sphere colour update", () => {
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("set_lights_in_sphere");
    expect(PYTHON_ROUTINE_DEFAULT_SOURCE).toContain("await");
  });
});
