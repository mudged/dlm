import { afterEach, describe, expect, it, vi } from "vitest";
import { clientId } from "./clientId";

describe("clientId", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses crypto.randomUUID when available", () => {
    const uuid = "11111111-2222-4333-8444-555555555555";
    vi.stubGlobal("crypto", {
      randomUUID: () => uuid,
    });
    expect(clientId()).toBe(uuid);
  });

  it("falls back when crypto.randomUUID is missing (non-secure context)", () => {
    vi.stubGlobal("crypto", {});
    const a = clientId();
    const b = clientId();
    expect(a).toMatch(/^[0-9a-z-]+$/i);
    expect(b).toMatch(/^[0-9a-z-]+$/i);
    expect(a).not.toBe(b);
  });

  it("falls back when crypto is undefined", () => {
    vi.stubGlobal("crypto", undefined);
    expect(clientId()).toMatch(/^[0-9a-z-]+$/i);
  });
});
