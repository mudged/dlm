import { describe, expect, it } from "vitest";
import { SCENE_API_COMPLETIONS } from "./pythonRoutineCodemirror";

describe("pythonRoutineCodemirror", () => {
  it("lists scene API members for completion manifest", () => {
    const labels = SCENE_API_COMPLETIONS.map((o) => o.label);
    expect(labels).toContain("height");
    expect(labels).toContain("get_all_lights");
    expect(labels).toContain("update_lights_batch");
  });
});
