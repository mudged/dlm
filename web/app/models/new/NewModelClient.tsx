"use client";

import {
  faCheck,
  faUpload,
  faXmark,
} from "@fortawesome/free-solid-svg-icons";
import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ButtonLink } from "@/components/ui/ButtonLink";
import {
  confirmCaptureJob,
  createCaptureJob,
  discardCaptureJob,
  getCaptureJob,
} from "@/lib/models";
import type { CaptureJob, CaptureJobLight, Light } from "@/lib/models";

// ── shared helpers ────────────────────────────────────────────────────────────

const POLL_INTERVAL_MS = 2_000;

function captureToPreviewLights(lights: CaptureJobLight[]): Light[] {
  return lights.map((l) => ({
    ...l,
    on: true,
    color: "#ffffff",
    brightness_pct: 100,
  }));
}

// ── lazy 3D canvas (SSR-safe) ────────────────────────────────────────────────

const ModelLightsCanvas = dynamic(
  () => import("@/components/ModelLightsCanvas"),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-[min(40vh,20rem)] min-h-[200px] w-full items-center justify-center rounded-xl border border-slate-200 bg-slate-50 text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900/50 dark:text-slate-400">
        Preparing 3D view…
      </div>
    ),
  },
);

// ── error banner ─────────────────────────────────────────────────────────────

function ErrorBanner({ msg }: { msg: string }) {
  return (
    <p
      className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
      role="alert"
    >
      {msg}
    </p>
  );
}

// ── CSV upload panel (original form) ─────────────────────────────────────────

function CsvPanel() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!name.trim()) {
      setError("Name is required.");
      return;
    }
    if (!file) {
      setError("Choose a CSV file.");
      return;
    }
    setSubmitting(true);
    try {
      const fd = new FormData();
      fd.set("name", name.trim());
      fd.set("file", file);
      const res = await fetch("/api/v1/models", { method: "POST", body: fd });
      const j = (await res.json().catch(() => null)) as {
        error?: { message?: string };
        id?: string;
      };
      if (!res.ok) {
        setError(j?.error?.message ?? `Upload failed (${res.status})`);
        setSubmitting(false);
        return;
      }
      const id = j && "id" in j && typeof j.id === "string" ? j.id : null;
      router.push(id ? `/models/detail?id=${encodeURIComponent(id)}` : "/models");
    } catch {
      setError("Could not reach the API.");
      setSubmitting(false);
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={(e) => void onSubmit(e)}>
      <div className="flex flex-col gap-1">
        <label htmlFor="csv-model-name" className="text-sm font-medium">
          Name
        </label>
        <input
          id="csv-model-name"
          name="name"
          type="text"
          autoComplete="off"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          placeholder="e.g. Living room string"
        />
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="csv-file" className="text-sm font-medium">
          CSV file
        </label>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Header must be{" "}
          <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
            id,x,y,z
          </code>
          . IDs must be sequential from 0 with no gaps (max 1000 rows).
        </p>
        <input
          id="csv-file"
          name="file"
          type="file"
          accept=".csv,text/csv"
          className="min-h-11 text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-slate-200 file:px-3 file:py-2 file:text-sm dark:file:bg-slate-700"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        />
      </div>
      {error ? <ErrorBanner msg={error} /> : null}
      <div className="flex flex-col gap-2 sm:flex-row">
        <Button
          type="submit"
          icon={faUpload}
          disabled={submitting}
          className="w-full sm:w-auto"
        >
          {submitting ? "Uploading…" : "Upload"}
        </Button>
        <ButtonLink
          href="/models"
          icon={faXmark}
          className="w-full sm:ml-2 sm:w-auto"
        >
          Cancel
        </ButtonLink>
      </div>
    </form>
  );
}

// ── video panel ───────────────────────────────────────────────────────────────

type VideoPhase =
  | { kind: "idle" }
  | { kind: "submitting" }
  | { kind: "polling"; jobId: string }
  | { kind: "review"; jobId: string; job: CaptureJob }
  | { kind: "confirming"; jobId: string; job: CaptureJob }
  | { kind: "failed"; jobId: string; errorMsg: string };

function VideoPanel({
  initialJobId,
}: {
  initialJobId: string | null;
}) {
  const router = useRouter();
  const params = useSearchParams();

  const [phase, setPhase] = useState<VideoPhase>(
    initialJobId ? { kind: "polling", jobId: initialJobId } : { kind: "idle" },
  );
  const [files, setFiles] = useState<File[]>([]);
  const [useMarker, setUseMarker] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const [confirmName, setConfirmName] = useState("");
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const [cameraResetVersion, setCameraResetVersion] = useState(0);

  const pollRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearPoll = () => {
    if (pollRef.current !== null) {
      clearTimeout(pollRef.current);
      pollRef.current = null;
    }
  };

  const setJobIdInUrl = useCallback(
    (jobId: string | null) => {
      const tab = params.get("tab") ?? "video";
      if (jobId) {
        router.replace(
          `/models/new?tab=${encodeURIComponent(tab)}&jobId=${encodeURIComponent(jobId)}`,
        );
      } else {
        router.replace(`/models/new?tab=${encodeURIComponent(tab)}`);
      }
    },
    [params, router],
  );

  const startPolling = useCallback(
    (jobId: string) => {
      clearPoll();
      const poll = async () => {
        try {
          const job = await getCaptureJob(jobId);
          if (job.status === "succeeded") {
            setPhase({ kind: "review", jobId, job });
            return;
          }
          if (job.status === "failed") {
            const msg =
              job.error?.message ?? "Reconstruction failed — please try again.";
            setPhase({ kind: "failed", jobId, errorMsg: msg });
            return;
          }
          setPhase({ kind: "polling", jobId });
          pollRef.current = setTimeout(() => void poll(), POLL_INTERVAL_MS);
        } catch (err) {
          const msg =
            err instanceof Error ? err.message : "Could not reach the API.";
          setPhase({ kind: "failed", jobId, errorMsg: msg });
        }
      };
      void poll();
    },
    [],
  );

  useEffect(() => {
    if (phase.kind === "polling") {
      startPolling(phase.jobId);
    }
    return clearPoll;
  }, [phase.kind === "polling" ? phase.jobId : null, startPolling]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    return () => clearPoll();
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setUploadError(null);
    if (files.length < 2) {
      setUploadError("Select at least 2 video files.");
      return;
    }
    setPhase({ kind: "submitting" });
    try {
      const job = await createCaptureJob(files, { marker: useMarker || undefined });
      setJobIdInUrl(job.job_id);
      if (job.status === "succeeded") {
        setPhase({ kind: "review", jobId: job.job_id, job });
      } else if (job.status === "failed") {
        const msg =
          job.error?.message ?? "Reconstruction failed — please try again.";
        setPhase({ kind: "failed", jobId: job.job_id, errorMsg: msg });
      } else {
        setPhase({ kind: "polling", jobId: job.job_id });
      }
    } catch (err) {
      const msg =
        err instanceof Error ? err.message : "Could not reach the API.";
      setUploadError(msg);
      setPhase({ kind: "idle" });
    }
  }

  async function onConfirm() {
    if (phase.kind !== "review") return;
    const { jobId, job } = phase;
    if (!confirmName.trim()) {
      setConfirmError("Name is required.");
      return;
    }
    setConfirmError(null);
    setPhase({ kind: "confirming", jobId, job });
    try {
      const { id } = await confirmCaptureJob(jobId, confirmName.trim());
      router.push(`/models/detail?id=${encodeURIComponent(id)}`);
    } catch (err) {
      const msg =
        err instanceof Error ? err.message : "Could not reach the API.";
      setConfirmError(msg);
      setPhase({ kind: "review", jobId, job });
    }
  }

  async function onCancel() {
    const jobId =
      phase.kind === "review" ||
      phase.kind === "failed" ||
      phase.kind === "confirming"
        ? phase.jobId
        : null;
    if (jobId) {
      try {
        await discardCaptureJob(jobId);
      } catch {
        // best-effort
      }
    }
    setJobIdInUrl(null);
    setPhase({ kind: "idle" });
    setFiles([]);
    setConfirmName("");
    setConfirmError(null);
  }

  // ── idle / upload form ────────────────────────────────────────────────────
  if (phase.kind === "idle" || phase.kind === "submitting") {
    return (
      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => void onSubmit(e)}
      >
        <div className="flex flex-col gap-1">
          <label htmlFor="video-files" className="text-sm font-medium">
            Video files{" "}
            <span className="font-normal text-slate-500">(≥ 2 required)</span>
          </label>
          <input
            id="video-files"
            name="files"
            type="file"
            accept="video/*"
            multiple
            className="min-h-11 text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-slate-200 file:px-3 file:py-2 file:text-sm dark:file:bg-slate-700"
            onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
            disabled={phase.kind === "submitting"}
          />
          {files.length > 0 ? (
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {files.length} file{files.length === 1 ? "" : "s"} selected
            </p>
          ) : null}
        </div>

        <div className="flex flex-col gap-2 rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-800/40">
          <p className="text-xs font-medium text-slate-700 dark:text-slate-300">
            Fiducial marker (optional)
          </p>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={useMarker}
              onChange={(e) => setUseMarker(e.target.checked)}
              disabled={phase.kind === "submitting"}
              className="rounded border-slate-300 dark:border-slate-600"
            />
            I&apos;ve placed a printed marker in the scene
          </label>
          <a
            href="/api/v1/capture/marker"
            download
            className="inline-flex w-fit items-center gap-1.5 text-xs text-sky-600 underline-offset-2 hover:underline dark:text-sky-400"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-3.5 w-3.5"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden
            >
              <path
                fillRule="evenodd"
                d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"
                clipRule="evenodd"
              />
            </svg>
            Download printable marker
          </a>
        </div>

        {uploadError ? <ErrorBanner msg={uploadError} /> : null}

        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            type="submit"
            icon={faUpload}
            disabled={phase.kind === "submitting"}
            className="w-full sm:w-auto"
          >
            {phase.kind === "submitting" ? "Uploading…" : "Start reconstruction"}
          </Button>
          <ButtonLink
            href="/models"
            icon={faXmark}
            className="w-full sm:ml-2 sm:w-auto"
          >
            Cancel
          </ButtonLink>
        </div>
      </form>
    );
  }

  // ── polling ───────────────────────────────────────────────────────────────
  if (phase.kind === "polling") {
    return (
      <div className="flex flex-col gap-4">
        <div
          className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-5 dark:border-slate-700 dark:bg-slate-800/40"
          role="status"
          aria-live="polite"
        >
          <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
            Reconstruction in progress…
          </p>
          <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
            Job&nbsp;
            <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
              {phase.jobId}
            </code>
            . You can close this tab and return — processing continues on the
            server.
          </p>
          <div
            className="mt-3 h-2 w-full overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700"
            aria-hidden
          >
            <div className="h-full animate-pulse rounded-full bg-sky-500" />
          </div>
        </div>
        <div>
          <Button
            type="button"
            icon={faXmark}
            className="w-full bg-slate-200 text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 sm:w-auto"
            onClick={() => void onCancel()}
          >
            Cancel job
          </Button>
        </div>
      </div>
    );
  }

  // ── failed ────────────────────────────────────────────────────────────────
  if (phase.kind === "failed") {
    return (
      <div className="flex flex-col gap-4">
        <ErrorBanner msg={phase.errorMsg} />
        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            type="button"
            icon={faUpload}
            className="w-full sm:w-auto"
            onClick={() => void onCancel()}
          >
            Try again
          </Button>
          <ButtonLink href="/models" icon={faXmark} className="w-full sm:ml-2 sm:w-auto">
            Back to models
          </ButtonLink>
        </div>
      </div>
    );
  }

  // ── review / confirming ────────────────────────────────────────────────────
  const { jobId, job } = phase;
  const result = job.result;
  const previewLights = result ? captureToPreviewLights(result.lights) : [];
  const isConfirming = phase.kind === "confirming";

  return (
    <div className="flex flex-col gap-6">
      {/* detection summary */}
      <section className="flex flex-col gap-3">
        <div className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-4 dark:border-slate-700 dark:bg-slate-800/40">
          <p className="text-sm font-semibold text-slate-800 dark:text-slate-200">
            {result?.light_count ?? 0} light
            {(result?.light_count ?? 0) === 1 ? "" : "s"} detected
          </p>
          {result && result.missing.length > 0 ? (
            <p className="mt-1.5 text-xs text-amber-700 dark:text-amber-400">
              Missing IDs: {result.missing.join(", ")}
            </p>
          ) : null}
          {result && result.low_confidence.length > 0 ? (
            <p className="mt-1 text-xs text-amber-600 dark:text-amber-500">
              Low-confidence IDs: {result.low_confidence.join(", ")}
            </p>
          ) : null}
          {result &&
          result.missing.length === 0 &&
          result.low_confidence.length === 0 ? (
            <p className="mt-1 text-xs text-emerald-700 dark:text-emerald-400">
              All lights detected with high confidence.
            </p>
          ) : null}
          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
            Job{" "}
            <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
              {jobId}
            </code>
          </p>
        </div>

        {/* 3D preview */}
        {previewLights.length > 0 ? (
          <section aria-label="3D preview of detected lights">
            <div className="mb-1 flex items-center justify-between">
              <p className="text-xs font-medium text-slate-600 dark:text-slate-400">
                3D preview (read-only)
              </p>
              <button
                type="button"
                onClick={() => setCameraResetVersion((v) => v + 1)}
                className="flex items-center gap-1 rounded px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                aria-label="Reset camera"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-3.5 w-3.5"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                  aria-hidden
                >
                  <path
                    fillRule="evenodd"
                    d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z"
                    clipRule="evenodd"
                  />
                </svg>
                Reset camera
              </button>
            </div>
            <ModelLightsCanvas
              lights={previewLights}
              cameraPersistenceKey={`capture-preview-${jobId}`}
              cameraResetVersion={cameraResetVersion}
            />
          </section>
        ) : null}
      </section>

      {/* confirm form */}
      <section className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor="confirm-name" className="text-sm font-medium">
            Model name
          </label>
          <input
            id="confirm-name"
            type="text"
            autoComplete="off"
            value={confirmName}
            onChange={(e) => {
              setConfirmName(e.target.value);
              setConfirmError(null);
            }}
            disabled={isConfirming}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="e.g. Living room string"
          />
        </div>
        {confirmError ? <ErrorBanner msg={confirmError} /> : null}
        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            type="button"
            icon={faCheck}
            disabled={isConfirming}
            className="w-full sm:w-auto"
            onClick={() => void onConfirm()}
          >
            {isConfirming ? "Saving…" : "Confirm model"}
          </Button>
          <Button
            type="button"
            icon={faXmark}
            disabled={isConfirming}
            className="w-full bg-slate-200 text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 sm:ml-2 sm:w-auto"
            onClick={() => void onCancel()}
          >
            Cancel
          </Button>
        </div>
      </section>
    </div>
  );
}

// ── tab bar ───────────────────────────────────────────────────────────────────

type Tab = "csv" | "video";

function TabBar({
  active,
  onSelect,
}: {
  active: Tab;
  onSelect: (t: Tab) => void;
}) {
  const base =
    "flex min-h-11 flex-1 items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition";
  const activeClass =
    "bg-white text-slate-900 shadow-sm dark:bg-slate-700 dark:text-white";
  const inactiveClass =
    "text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white";

  return (
    <div
      role="tablist"
      aria-label="Model creation method"
      className="flex rounded-xl bg-slate-100 p-1 dark:bg-slate-800"
    >
      <button
        role="tab"
        aria-selected={active === "csv"}
        type="button"
        className={`${base} ${active === "csv" ? activeClass : inactiveClass}`}
        onClick={() => onSelect("csv")}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 shrink-0"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path
            fillRule="evenodd"
            d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z"
            clipRule="evenodd"
          />
        </svg>
        CSV upload
      </button>
      <button
        role="tab"
        aria-selected={active === "video"}
        type="button"
        className={`${base} ${active === "video" ? activeClass : inactiveClass}`}
        onClick={() => onSelect("video")}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 shrink-0"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path d="M2 6a2 2 0 012-2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6zM14.553 7.106A1 1 0 0014 8v4a1 1 0 00.553.894l2 1A1 1 0 0018 13V7a1 1 0 00-1.447-.894l-2 1z" />
        </svg>
        Create from video
      </button>
    </div>
  );
}

// ── main export ───────────────────────────────────────────────────────────────

export function NewModelClient() {
  const router = useRouter();
  const params = useSearchParams();

  const rawTab = params.get("tab");
  const activeTab: Tab = rawTab === "video" ? "video" : "csv";
  const jobId = params.get("jobId");

  function switchTab(t: Tab) {
    router.push(`/models/new?tab=${t}`);
  }

  return (
    <div className="flex flex-col gap-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          New model
        </h1>
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
          Upload a CSV of light coordinates, or reconstruct from video.
        </p>
      </header>

      <TabBar active={activeTab} onSelect={switchTab} />

      <div role="tabpanel">
        {activeTab === "csv" ? (
          <CsvPanel />
        ) : (
          <VideoPanel initialJobId={jobId} />
        )}
      </div>
    </div>
  );
}
