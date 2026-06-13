import { afterEach, describe, expect, it, vi } from "vitest";
import {
  confirmCaptureJob,
  createCaptureJob,
  discardCaptureJob,
  getCaptureJob,
} from "./models";

const MOCK_JOB = {
  job_id: "job-abc",
  status: "pending" as const,
};

const MOCK_JOB_SUCCEEDED = {
  job_id: "job-abc",
  status: "succeeded" as const,
  result: {
    light_count: 3,
    lights: [
      { id: 0, x: 0.1, y: 0.2, z: 0.3 },
      { id: 1, x: 0.4, y: 0.5, z: 0.6 },
      { id: 2, x: 0.7, y: 0.8, z: 0.9 },
    ],
    missing: [],
    low_confidence: [],
  },
};

describe("createCaptureJob (REQ-049)", () => {
  const realFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("POSTs to /api/v1/models/capture with FormData files", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(JSON.stringify(MOCK_JOB), {
        status: 202,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const f1 = new File(["a"], "clip1.mp4", { type: "video/mp4" });
    const f2 = new File(["b"], "clip2.mp4", { type: "video/mp4" });
    const result = await createCaptureJob([f1, f2]);

    expect(result.job_id).toBe("job-abc");
    expect(calls).toHaveLength(1);
    expect(calls[0].url).toBe("/api/v1/models/capture");
    expect(calls[0].init.method).toBe("POST");
    const body = calls[0].init.body as FormData;
    expect(body).toBeInstanceOf(FormData);
    expect(body.getAll("files")).toHaveLength(2);
  });

  it("appends marker=true when marker param is set", async () => {
    let captured: FormData | null = null;
    globalThis.fetch = vi.fn(async (_url, init) => {
      captured = init?.body as FormData;
      return new Response(JSON.stringify(MOCK_JOB), {
        status: 202,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const f1 = new File(["a"], "a.mp4", { type: "video/mp4" });
    const f2 = new File(["b"], "b.mp4", { type: "video/mp4" });
    await createCaptureJob([f1, f2], { marker: true });

    expect(captured).not.toBeNull();
    expect((captured as FormData).get("marker")).toBe("true");
  });

  it("surfaces error.message and error.code on 400", async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          error: {
            code: "too_few_files",
            message: "At least 2 video files are required",
          },
        }),
        { status: 400, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const f = new File(["x"], "lone.mp4", { type: "video/mp4" });
    await expect(createCaptureJob([f])).rejects.toMatchObject({
      message: "At least 2 video files are required",
      code: "too_few_files",
    });
  });
});

describe("getCaptureJob (REQ-049)", () => {
  const realFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("GETs /api/v1/models/capture/{jobId} and returns the job", async () => {
    const calls: string[] = [];
    globalThis.fetch = vi.fn(async (url) => {
      calls.push(url as string);
      return new Response(JSON.stringify(MOCK_JOB_SUCCEEDED), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    const result = await getCaptureJob("job-abc");

    expect(result.status).toBe("succeeded");
    expect(result.result?.light_count).toBe(3);
    expect(calls[0]).toBe("/api/v1/models/capture/job-abc");
  });

  it("URL-encodes the jobId", async () => {
    let captured = "";
    globalThis.fetch = vi.fn(async (url) => {
      captured = url as string;
      return new Response(JSON.stringify(MOCK_JOB), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }) as typeof globalThis.fetch;

    await getCaptureJob("job/with spaces");
    expect(captured).toBe("/api/v1/models/capture/job%2Fwith%20spaces");
  });

  it("surfaces error message on 404", async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({ error: { message: "Job not found", code: "not_found" } }),
        { status: 404, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    await expect(getCaptureJob("missing")).rejects.toMatchObject({
      message: "Job not found",
      code: "not_found",
    });
  });
});

describe("confirmCaptureJob (REQ-049)", () => {
  const realFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("POSTs to /api/v1/models/capture/{jobId}/confirm with { name }", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url, init) => {
      calls.push({ url: url as string, init: init as RequestInit });
      return new Response(
        JSON.stringify({ id: "model-xyz", name: "My model" }),
        { status: 201, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    const result = await confirmCaptureJob("job-abc", "My model");

    expect(result.id).toBe("model-xyz");
    expect(calls[0].url).toBe("/api/v1/models/capture/job-abc/confirm");
    expect(calls[0].init.method).toBe("POST");
    const body = JSON.parse((calls[0].init.body as string) ?? "{}") as unknown;
    expect(body).toEqual({ name: "My model" });
  });

  it("surfaces 409 conflict message (duplicate name)", async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          error: {
            code: "duplicate_name",
            message: "A model with that name already exists",
          },
        }),
        { status: 409, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    await expect(confirmCaptureJob("job-abc", "Existing")).rejects.toMatchObject({
      code: "duplicate_name",
    });
  });
});

describe("discardCaptureJob (REQ-049)", () => {
  const realFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("DELETEs /api/v1/models/capture/{jobId} and resolves on 204", async () => {
    const calls: string[] = [];
    globalThis.fetch = vi.fn(async (url) => {
      calls.push(url as string);
      return new Response(null, { status: 204 });
    }) as typeof globalThis.fetch;

    await expect(discardCaptureJob("job-abc")).resolves.toBeUndefined();
    expect(calls[0]).toBe("/api/v1/models/capture/job-abc");
  });

  it("URL-encodes the jobId on delete", async () => {
    let captured = "";
    globalThis.fetch = vi.fn(async (url) => {
      captured = url as string;
      return new Response(null, { status: 204 });
    }) as typeof globalThis.fetch;

    await discardCaptureJob("job with spaces");
    expect(captured).toBe("/api/v1/models/capture/job%20with%20spaces");
  });

  it("throws on non-204 non-ok responses", async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({ error: { message: "Already discarded", code: "gone" } }),
        { status: 410, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof globalThis.fetch;

    await expect(discardCaptureJob("job-gone")).rejects.toMatchObject({
      message: "Already discarded",
      code: "gone",
    });
  });
});
