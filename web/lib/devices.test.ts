import { afterEach, describe, expect, it, vi } from "vitest";
import {
  CaptureError,
  createDevice,
  fetchDevice,
  getCaptureStatus,
  patchDevice,
  startCapture,
  stopCapture,
} from "./devices";

const DEVICE_FIXTURE = {
  id: "dev-1",
  type: "wled",
  name: "Studio strip",
  base_url: "http://wled.local",
  light_count: 120,
  created_at: "2026-06-01T00:00:00Z",
};

describe("Device.light_count (REQ-047)", () => {
  const realFetch = globalThis.fetch;
  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("fetchDevice returns light_count from the server", async () => {
    globalThis.fetch = vi.fn(async () =>
      new Response(JSON.stringify(DEVICE_FIXTURE), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    ) as typeof globalThis.fetch;

    const d = await fetchDevice("dev-1");
    expect(d.light_count).toBe(120);
  });

  it("createDevice sends light_count in the request body", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(JSON.stringify({ ...DEVICE_FIXTURE, light_count: 60 }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const d = await createDevice({
      name: "Studio strip",
      base_url: "http://wled.local",
      light_count: 60,
    });

    expect(d.light_count).toBe(60);
    expect(calls).toHaveLength(1);
    const body = JSON.parse((calls[0].init.body as string) ?? "{}");
    expect(body.light_count).toBe(60);
  });

  it("createDevice omits light_count when not provided", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(JSON.stringify(DEVICE_FIXTURE), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    await createDevice({ name: "Studio strip", base_url: "http://wled.local" });
    const body = JSON.parse((calls[0].init.body as string) ?? "{}");
    expect(body).not.toHaveProperty("light_count");
  });

  it("patchDevice sends light_count in the patch body", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(JSON.stringify({ ...DEVICE_FIXTURE, light_count: 200 }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const d = await patchDevice("dev-1", { light_count: 200 });
    expect(d.light_count).toBe(200);
    const body = JSON.parse((calls[0].init.body as string) ?? "{}");
    expect(body.light_count).toBe(200);
  });
});

describe("startCapture (REQ-047)", () => {
  const realFetch = globalThis.fetch;
  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("POSTs to /capture/start and returns CaptureStatus", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(
        JSON.stringify({ device_id: "dev-1", state: "running", light_count: 120, current_index: 0 }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const status = await startCapture("dev-1");
    expect(status.state).toBe("running");
    expect(status.light_count).toBe(120);
    expect(status.current_index).toBe(0);
    expect(calls[0].url).toBe("/api/v1/devices/dev-1/capture/start");
    expect(calls[0].init.method).toBe("POST");
  });

  it("throws CaptureError with code=capture_conflict on 409", async () => {
    globalThis.fetch = vi.fn(async () =>
      new Response(
        JSON.stringify({ error: { code: "capture_conflict", message: "already running" } }),
        { status: 409, headers: { "Content-Type": "application/json" } },
      ),
    ) as typeof globalThis.fetch;

    await expect(startCapture("dev-1")).rejects.toSatisfy(
      (e: unknown) => e instanceof CaptureError && e.code === "capture_conflict",
    );
  });

  it("throws CaptureError with code=capture_no_lights on 422", async () => {
    globalThis.fetch = vi.fn(async () =>
      new Response(
        JSON.stringify({ error: { code: "capture_no_lights", message: "no lights configured" } }),
        { status: 422, headers: { "Content-Type": "application/json" } },
      ),
    ) as typeof globalThis.fetch;

    await expect(startCapture("dev-1")).rejects.toSatisfy(
      (e: unknown) => e instanceof CaptureError && e.code === "capture_no_lights",
    );
  });
});

describe("stopCapture (REQ-047)", () => {
  const realFetch = globalThis.fetch;
  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("POSTs to /capture/stop and returns idle state", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(
        JSON.stringify({ device_id: "dev-1", state: "idle", light_count: 120 }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const status = await stopCapture("dev-1");
    expect(status.state).toBe("idle");
    expect(calls[0].url).toBe("/api/v1/devices/dev-1/capture/stop");
    expect(calls[0].init.method).toBe("POST");
  });
});

describe("getCaptureStatus (REQ-047)", () => {
  const realFetch = globalThis.fetch;
  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("GETs /capture and returns status", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(
        JSON.stringify({ state: "running", light_count: 120, current_index: 5 }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const status = await getCaptureStatus("dev-1");
    expect(status.state).toBe("running");
    expect(status.current_index).toBe(5);
    expect(calls[0].url).toBe("/api/v1/devices/dev-1/capture");
    expect(calls[0].init.method).toBeUndefined(); // GET has no explicit method
  });

  it("throws Error on non-ok response", async () => {
    globalThis.fetch = vi.fn(async () =>
      new Response(
        JSON.stringify({ error: { message: "device not found" } }),
        { status: 404, headers: { "Content-Type": "application/json" } },
      ),
    ) as typeof globalThis.fetch;

    await expect(getCaptureStatus("dev-missing")).rejects.toThrow("device not found");
  });
});
