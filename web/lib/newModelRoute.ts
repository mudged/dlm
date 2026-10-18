export type NewModelTab = "csv" | "video";

export type InitialVideoPhase =
  | { kind: "idle" }
  | { kind: "polling"; jobId: string };

/** Build `/models/new` URL with tab and optional in-progress capture job. */
export function buildNewModelUrl(
  tab: NewModelTab,
  jobId?: string | null,
): string {
  const base = `/models/new?tab=${encodeURIComponent(tab)}`;
  if (jobId) {
    return `${base}&jobId=${encodeURIComponent(jobId)}`;
  }
  return base;
}

/** Video is the usual path. Only an explicit csv tab opens the file form. */
export function activeNewModelTab(rawTab: string | null): NewModelTab {
  return rawTab === "csv" ? "csv" : "video";
}

/** Video panel phase on mount when restoring from URL after a tab switch. */
export function initialVideoPhaseFromJobId(
  jobId: string | null,
): InitialVideoPhase {
  return jobId ? { kind: "polling", jobId } : { kind: "idle" };
}
