export type ModelSummary = {
  id: string;
  name: string;
  created_at: string;
  light_count: number;
};

export type Light = {
  id: number;
  x: number;
  y: number;
  z: number;
  on: boolean;
  color: string;
  brightness_pct: number;
};

export type ModelDetail = ModelSummary & { lights: Light[] };

// ── Capture-job types (REQ-049) ──────────────────────────────────────────────

export type CaptureJobLight = { id: number; x: number; y: number; z: number };

export type CaptureJobRejectedFeed = { file: string; reason: string };

export type CaptureJobResult = {
  light_count: number;
  lights: CaptureJobLight[];
  missing: number[];
  low_confidence: number[];
  rejected_feeds?: CaptureJobRejectedFeed[];
};

export const captureLightCountRangeMessage =
  "Light count must be a whole number from 1 to 1000.";

/** Empty means omit light_count. Otherwise a whole number from 1 to 1000. */
export function parseOptionalCaptureLightCount(raw: string): number | undefined {
  const trimmed = raw.trim();
  if (trimmed === "") return undefined;
  if (!/^[0-9]+$/.test(trimmed)) {
    throw new Error(captureLightCountRangeMessage);
  }
  const n = Number.parseInt(trimmed, 10);
  if (n < 1 || n > 1000) {
    throw new Error(captureLightCountRangeMessage);
  }
  return n;
}

export type CaptureJobStatus =
  | "pending"
  | "running"
  | "succeeded"
  | "failed";

export type CaptureJob = {
  job_id: string;
  status: CaptureJobStatus;
  progress?: number;
  result?: CaptureJobResult;
  error?: { message?: string; code?: string };
};

type CaptureApiError = Error & { code?: string };

function captureApiError(
  j: { error?: string | { message?: string; code?: string } } | null,
  fallback: string,
): CaptureApiError {
  let message = fallback;
  let code: string | undefined;
  if (typeof j?.error === "string" && j.error.trim()) {
    message = j.error;
  } else if (j?.error && typeof j.error === "object") {
    if (j.error.message) message = j.error.message;
    code = j.error.code;
  }
  const err = new Error(message) as CaptureApiError;
  if (code) err.code = code;
  return err;
}

/** Combined video upload limit. Must match maxCaptureUploadBytes in the Go handler. */
export const maxCaptureUploadBytes = 2 * 1024 * 1024 * 1024;

const uploadConnectionMessage =
  "The upload stopped before the server could reply. The connection may have timed out or dropped. Check the app log on the machine running the server.";

function captureUploadLimitMessage(): string {
  const gb = maxCaptureUploadBytes / (1024 * 1024 * 1024);
  return `The videos are larger than ${gb} GB combined. Use shorter clips.`;
}

function isConnectionFailure(err: unknown): boolean {
  if (!(err instanceof Error)) return false;
  const msg = err.message.toLowerCase();
  return (
    err instanceof TypeError ||
    msg.includes("failed to fetch") ||
    msg.includes("networkerror") ||
    msg.includes("load failed") ||
    msg.includes("network request failed")
  );
}

/** POST /api/v1/models/capture — submit ≥ 2 video files, returns a new job. */
export async function createCaptureJob(
  files: File[],
  params?: { marker?: boolean; scale_hint?: number; light_count?: number },
): Promise<CaptureJob> {
  const totalBytes = files.reduce((sum, file) => sum + file.size, 0);
  if (totalBytes > maxCaptureUploadBytes) {
    throw new Error(captureUploadLimitMessage());
  }
  const fd = new FormData();
  for (const f of files) {
    fd.append("files", f);
  }
  if (params?.marker) fd.set("marker", "true");
  if (params?.scale_hint !== undefined)
    fd.set("scale_hint", String(params.scale_hint));
  if (
    params?.light_count !== undefined &&
    Number.isInteger(params.light_count) &&
    params.light_count >= 1 &&
    params.light_count <= 1000
  ) {
    fd.set("light_count", String(params.light_count));
  }
  let res: Response;
  try {
    res = await fetch("/api/v1/models/capture", {
      method: "POST",
      body: fd,
    });
  } catch (err) {
    if (isConnectionFailure(err)) {
      throw new Error(uploadConnectionMessage);
    }
    throw err;
  }
  const j = (await res.json().catch(() => null)) as
    | (CaptureJob & { error?: { message?: string; code?: string } })
    | null;
  if (!res.ok)
    throw captureApiError(j, `capture job create failed (${res.status})`);
  return j as CaptureJob;
}

type CaptureJobWire = Omit<CaptureJob, "error"> & {
  error?: string | { message?: string; code?: string };
};

/** The capture job API sends `error` as a string. Older callers also accept `{ message }`. */
function normalizeCaptureJob(j: CaptureJobWire): CaptureJob {
  const raw = j.error;
  let error: CaptureJob["error"];
  if (typeof raw === "string") {
    error = raw.trim() ? { message: raw } : undefined;
  } else {
    error = raw;
  }
  return { ...j, error };
}

/** GET /api/v1/models/capture/{jobId} — poll job status. */
export async function getCaptureJob(jobId: string): Promise<CaptureJob> {
  const res = await fetch(
    `/api/v1/models/capture/${encodeURIComponent(jobId)}`,
    { cache: "no-store" },
  );
  const j = (await res.json().catch(() => null)) as CaptureJobWire | null;
  if (!res.ok)
    throw captureApiError(j, `capture job fetch failed (${res.status})`);
  if (!j) throw new Error("capture job fetch returned no body");
  return normalizeCaptureJob(j);
}

/** POST /api/v1/models/capture/{jobId}/confirm — persist detected lights as a named model. */
export async function confirmCaptureJob(
  jobId: string,
  name: string,
): Promise<{ id: string }> {
  const res = await fetch(
    `/api/v1/models/capture/${encodeURIComponent(jobId)}/confirm`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    },
  );
  const j = (await res.json().catch(() => null)) as
    | ({ id?: string } & { error?: { message?: string; code?: string } })
    | null;
  if (!res.ok)
    throw captureApiError(j, `capture job confirm failed (${res.status})`);
  if (!j?.id) throw new Error("Confirm returned no model id.");
  return { id: j.id };
}

/** DELETE /api/v1/models/capture/{jobId} — discard a pending/succeeded job. */
export async function discardCaptureJob(jobId: string): Promise<void> {
  const res = await fetch(
    `/api/v1/models/capture/${encodeURIComponent(jobId)}`,
    { method: "DELETE" },
  );
  if (res.status === 204) return;
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string; code?: string };
    } | null;
    throw captureApiError(j, `capture job discard failed (${res.status})`);
  }
}
