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

/** Video panel phase on mount when restoring from URL after a tab switch. */
export function initialVideoPhaseFromJobId(
  jobId: string | null,
): InitialVideoPhase {
  return jobId ? { kind: "polling", jobId } : { kind: "idle" };
}
