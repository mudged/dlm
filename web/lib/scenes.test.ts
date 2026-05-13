import { afterEach, describe, expect, it, vi } from "vitest";
import {
  SCENE_BOUNDARY_MARGIN_DEFAULT_M,
  SCENE_BOUNDARY_MARGIN_MAX_M,
  SCENE_BOUNDARY_MARGIN_MIN_M,
  patchSceneMargin,
} from "./scenes";

describe("scene boundary margin constants (REQ-015 BR 12 / REQ-034 rule 3)", () => {
  it("default 0.3 m, range [0, 5]", () => {
    expect(SCENE_BOUNDARY_MARGIN_DEFAULT_M).toBe(0.3);
    expect(SCENE_BOUNDARY_MARGIN_MIN_M).toBe(0);
    expect(SCENE_BOUNDARY_MARGIN_MAX_M).toBe(5);
  });
});

describe("patchSceneMargin (REQ-015 BR 13 / REQ-034)", () => {
  const realFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("sends PATCH /api/v1/scenes/{id} with margin_m and returns the echo", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(
        JSON.stringify({
          id: "sc-1",
          name: "Living room",
          created_at: "2026-05-13T08:00:00Z",
          model_count: 2,
          margin_m: 0.5,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const result = await patchSceneMargin("sc-1", 0.5);
    expect(result.margin_m).toBe(0.5);
    expect(calls).toHaveLength(1);
    expect(calls[0].url).toBe("/api/v1/scenes/sc-1");
    expect(calls[0].init.method).toBe("PATCH");
    const body = JSON.parse((calls[0].init.body as string) ?? "{}");
    expect(body).toEqual({ margin_m: 0.5 });
  });

  it("surfaces error.code=invalid_margin_m on 400 with a typed Error", async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          error: {
            code: "invalid_margin_m",
            message: "margin_m must be finite and within [0, 5]",
            details: { min: 0, max: 5 },
          },
        }),
        { status: 400, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    await expect(patchSceneMargin("sc-1", 9.9)).rejects.toMatchObject({
      code: "invalid_margin_m",
    });
  });

  it("encodes scene id segments (REQ-015 path safety)", async () => {
    let captured = "";
    globalThis.fetch = vi.fn(async (url) => {
      captured = url as string;
      return new Response(
        JSON.stringify({
          id: "with space",
          name: "x",
          created_at: "2026-05-13T08:00:00Z",
          model_count: 0,
          margin_m: 0.3,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;
    await patchSceneMargin("with space", 0.3);
    expect(captured).toBe("/api/v1/scenes/with%20space");
  });
});
