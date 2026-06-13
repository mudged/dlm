/**
 * Returns the state to apply after a background resync fetch.
 *
 * Rules:
 * - If `current` is null (initial page load, state never populated), always
 *   apply the fetched state — there is nothing to protect.
 * - If the SSE sequence advanced between when the fetch was issued
 *   (`seqAtStart`) and when it resolved (`seqNow`), the fetched snapshot
 *   predates the latest SSE-merged state → keep `current` unchanged so the
 *   stale GET does not overwrite live updates.
 * - Otherwise (no SSE events arrived during the fetch) apply the fetched state.
 *
 * @param current     Current React state (null means never populated).
 * @param seqAtStart  `seqRef.current` captured just before the fetch was issued.
 * @param seqNow      `seqRef.current` at the moment the fetch result is being applied.
 * @param fetched     The state returned by the GET request.
 */
export function applyFetchIfNotStale<T>(
  current: T | null,
  seqAtStart: number | null,
  seqNow: number | null,
  fetched: T,
): T | null {
  if (current === null) return fetched;
  if (seqNow !== seqAtStart) return current;
  return fetched;
}
