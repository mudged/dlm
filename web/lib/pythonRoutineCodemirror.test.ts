import { describe, expect, it } from "vitest";
import { SCENE_API_COMPLETIONS } from "./pythonRoutineCodemirror";

describe("pythonRoutineCodemirror", () => {
  it("lists scene API members for completion manifest", () => {
    const labels = SCENE_API_COMPLETIONS.map((o) => o.label);
    expect(labels).toContain("width");
    expect(labels).toContain("height");
    expect(labels).toContain("depth");
    expect(labels).toContain("max_x");
    expect(labels).toContain("max_y");
    expect(labels).toContain("max_z");
    expect(labels).toContain("random_hex_colour");
    expect(labels).toContain("get_all_lights");
    expect(labels).toContain("update_lights_batch");
  });
});
