import { describe, expect, it } from "vitest";
import { VIZ_VIEWPORT_BG, VIZ_VIEWPORT_BG_CSS } from "./vizViewport";

describe("vizViewport", () => {
  it("uses architecture default REQ-019 hex", () => {
    expect(VIZ_VIEWPORT_BG).toBe(0x262626);
    expect(VIZ_VIEWPORT_BG_CSS).toBe("#262626");
  });
});
