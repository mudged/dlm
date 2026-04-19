import { describe, expect, it } from "vitest";
import {
  applyModelLightDeltas,
  applySceneLightDeltas,
  parseLightsSSEMessage,
} from "./lightDeltas";

describe("parseLightsSSEMessage", () => {
  it("parses seq and deltas", () => {
    const m = parseLightsSSEMessage(
      JSON.stringify({
        seq: 2,
        deltas: [{ light_id: 0, on: true, color: "#ff0000", brightness_pct: 80 }],
      }),
    );
    expect(m?.seq).toBe(2);
    expect(m?.deltas?.[0]?.light_id).toBe(0);
  });

  it("returns null for invalid json", () => {
    expect(parseLightsSSEMessage("not json")).toBeNull();
  });
});

describe("applyModelLightDeltas", () => {
  it("merges by light_id", () => {
    const out = applyModelLightDeltas(
      [
        { id: 0, x: 0, y: 0, z: 0, on: false, color: "#ffffff", brightness_pct: 100 },
        { id: 1, x: 1, y: 0, z: 0, on: false, color: "#ffffff", brightness_pct: 100 },
      ],
      [{ light_id: 1, on: true, color: "#00ff00", brightness_pct: 50 }],
    );
    expect(out[0]?.on).toBe(false);
    expect(out[1]?.on).toBe(true);
    expect(out[1]?.color).toBe("#00ff00");
    expect(out[1]?.brightness_pct).toBe(50);
  });
});

describe("applySceneLightDeltas", () => {
  it("merges by model_id and light_id", () => {
    const out = applySceneLightDeltas(
      [
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
              sx: 0,
              sy: 0,
              sz: 0,
              on: false,
              color: "#ffffff",
              brightness_pct: 100,
            },
          ],
        },
      ],
      [
        {
          model_id: "m1",
          light_id: 0,
          on: true,
          color: "#0000ff",
          brightness_pct: 90,
        },
      ],
    );
    expect(out[0]?.lights[0]?.on).toBe(true);
    expect(out[0]?.lights[0]?.color).toBe("#0000ff");
  });

  it("merges when SSE uses camelCase modelId/lightId keys", () => {
    const out = applySceneLightDeltas(
      [
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
              sx: 0,
              sy: 0,
              sz: 0,
              on: false,
              color: "#ffffff",
              brightness_pct: 100,
            },
          ],
        },
      ],
      [
        {
          modelId: "m1",
          lightId: 0,
          on: true,
          color: "#00ff00",
          brightness_pct: 50,
        } as Parameters<typeof applySceneLightDeltas>[1][0],
      ],
    );
    expect(out[0]?.lights[0]?.color).toBe("#00ff00");
  });
});
