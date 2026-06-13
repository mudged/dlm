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

export type CaptureJobResult = {
  light_count: number;
  lights: CaptureJobLight[];
  missing: number[];
  low_confidence: number[];
};

export type CaptureJobStatus =
  | "pending"
  | "processing"
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
  j: { error?: { message?: string; code?: string } } | null,
  fallback: string,
): CaptureApiError {
  const err = new Error(j?.error?.message ?? fallback) as CaptureApiError;
  if (j?.error?.code) err.code = j.error.code;
  return err;
}

/** POST /api/v1/models/capture — submit ≥ 2 video files, returns a new job. */
export async function createCaptureJob(
  files: File[],
  params?: { marker?: boolean; scale_hint?: number },
): Promise<CaptureJob> {
  const fd = new FormData();
  for (const f of files) {
    fd.append("files", f);
  }
  if (params?.marker) fd.set("marker", "true");
  if (params?.scale_hint !== undefined)
    fd.set("scale_hint", String(params.scale_hint));
  const res = await fetch("/api/v1/models/capture", {
    method: "POST",
    body: fd,
  });
  const j = (await res.json().catch(() => null)) as
    | (CaptureJob & { error?: { message?: string; code?: string } })
    | null;
  if (!res.ok)
    throw captureApiError(j, `capture job create failed (${res.status})`);
  return j as CaptureJob;
}

/** GET /api/v1/models/capture/{jobId} — poll job status. */
export async function getCaptureJob(jobId: string): Promise<CaptureJob> {
  const res = await fetch(
    `/api/v1/models/capture/${encodeURIComponent(jobId)}`,
    { cache: "no-store" },
  );
  const j = (await res.json().catch(() => null)) as
    | (CaptureJob & { error?: { message?: string; code?: string } })
    | null;
  if (!res.ok)
    throw captureApiError(j, `capture job fetch failed (${res.status})`);
  return j as CaptureJob;
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
