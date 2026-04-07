import { describe, expect, it } from "vitest";
import {
  PYTHON_SAMPLE_GROWING_SPHERE_SOURCE,
  PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE,
  PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE,
} from "./pythonRoutineSamples";

describe("pythonRoutineSamples", () => {
  it("growing sphere uses centre, max corners, and sphere bulk update", () => {
    expect(PYTHON_SAMPLE_GROWING_SPHERE_SOURCE).toContain("scene.max_x");
    expect(PYTHON_SAMPLE_GROWING_SPHERE_SOURCE).toContain("scene.width / 2");
    expect(PYTHON_SAMPLE_GROWING_SPHERE_SOURCE).toContain("set_lights_in_sphere");
    expect(PYTHON_SAMPLE_GROWING_SPHERE_SOURCE).toContain("time.monotonic()");
  });

  it("sweeping cuboid uses full width depth, 0.2 m height, and batch off for exits", () => {
    expect(PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE).toContain("SLAB_H = 0.2");
    expect(PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE).toContain("scene.max_y");
    expect(PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE).toContain("get_lights_within_cuboid");
    expect(PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE).toContain("update_lights_batch");
  });

  it("random colour cycle uses set_all_lights, get_all_lights, and ~1 s cadence", () => {
    expect(PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE).toContain(
      "set_all_lights",
    );
    expect(PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE).toContain(
      "get_all_lights",
    );
    expect(PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE).toContain(
      "CYCLE_SEC = 1.0",
    );
    expect(PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE).toContain(
      "scene.random_hex_colour()",
    );
  });
});
