import { describe, expect, it } from "vitest";
import { mergeSceneLightBatchIntoItems } from "./scenesMerge";
import type { SceneItem } from "./scenes";

describe("mergeSceneLightBatchIntoItems", () => {
  it("merges on color brightness for matching model and light id", () => {
    const items: SceneItem[] = [
      {
        model_id: "m1",
        name: "A",
        offset_x: 0,
        offset_y: 0,
        offset_z: 0,
        lights: [
          {
            id: 0,
            x: 0,
            y: 0,
            z: 0,
            sx: 1,
            sy: 1,
            sz: 1,
            on: false,
            color: "#ffffff",
            brightness_pct: 100,
          },
        ],
      },
    ];
    const next = mergeSceneLightBatchIntoItems(items, [
      { model_id: "m1", light_id: 0, on: true, color: "#ff0000", brightness_pct: 50 },
    ]);
    expect(next[0]!.lights[0]!.on).toBe(true);
    expect(next[0]!.lights[0]!.color).toBe("#ff0000");
    expect(next[0]!.lights[0]!.brightness_pct).toBe(50);
    expect(next[0]!.lights[0]!.sx).toBe(1);
  });
});
