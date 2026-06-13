import { describe, expect, it } from "vitest";
import {
  buildNewModelUrl,
  initialVideoPhaseFromJobId,
} from "./newModelRoute";

describe("buildNewModelUrl", () => {
  it("includes tab only when no jobId", () => {
    expect(buildNewModelUrl("csv")).toBe("/models/new?tab=csv");
    expect(buildNewModelUrl("video")).toBe("/models/new?tab=video");
  });

  it("includes jobId when present", () => {
    expect(buildNewModelUrl("video", "job-abc")).toBe(
      "/models/new?tab=video&jobId=job-abc",
    );
  });

  it("URL-encodes tab and jobId", () => {
    expect(buildNewModelUrl("video", "job/with spaces")).toBe(
      "/models/new?tab=video&jobId=job%2Fwith%20spaces",
    );
  });

  it("preserves jobId when switching CSV ↔ video during reconstruction", () => {
    const jobId = "job-in-progress";

    const videoUrl = buildNewModelUrl("video", jobId);
    expect(videoUrl).toContain("jobId=job-in-progress");

    const csvUrl = buildNewModelUrl("csv", jobId);
    expect(csvUrl).toBe("/models/new?tab=csv&jobId=job-in-progress");

    const backToVideo = buildNewModelUrl("video", jobId);
    expect(backToVideo).toBe("/models/new?tab=video&jobId=job-in-progress");
  });
});

describe("initialVideoPhaseFromJobId", () => {
  it("returns idle when jobId is absent", () => {
    expect(initialVideoPhaseFromJobId(null)).toEqual({ kind: "idle" });
  });

  it("returns polling when jobId is present so the panel resumes the job", () => {
    expect(initialVideoPhaseFromJobId("job-abc")).toEqual({
      kind: "polling",
      jobId: "job-abc",
    });
  });
});
