import { describe, expect, it } from "vitest";
import { applyFetchIfNotStale } from "./sseResyncGuard";

type State = { lights: number[] };
const A: State = { lights: [1, 2, 3] };
const B: State = { lights: [4, 5, 6] };

describe("applyFetchIfNotStale", () => {
  describe("initial load (current is null)", () => {
    it("applies fetched state when current is null and seq has not moved", () => {
      expect(applyFetchIfNotStale(null, null, null, B)).toBe(B);
    });

    it("applies fetched state even when SSE advanced during the fetch (current is null)", () => {
      // SSE fired during initial load but model was null, so deltas were no-ops.
      // The GET result is still the first populated state — always apply.
      expect(applyFetchIfNotStale(null, null, 10, B)).toBe(B);
    });
  });

  describe("resync fetch (current is non-null, seqAtStart is null after gap reset)", () => {
    it("applies fetched state when no SSE fired during the fetch", () => {
      // seqAtStart = null (gap handler reset), seqNow = null (no SSE yet)
      expect(applyFetchIfNotStale(A, null, null, B)).toBe(B);
    });

    it("keeps current state when SSE advanced during the fetch", () => {
      // SSE fired seq=15 while GET was in flight — current state is newer
      expect(applyFetchIfNotStale(A, null, 15, B)).toBe(A);
    });
  });

  describe("subsequent resync fetch (seqAtStart is non-null)", () => {
    it("applies fetched state when seq is unchanged (SSE quiet during fetch)", () => {
      expect(applyFetchIfNotStale(A, 7, 7, B)).toBe(B);
    });

    it("keeps current state when SSE advanced during the fetch", () => {
      expect(applyFetchIfNotStale(A, 7, 12, B)).toBe(A);
    });

    it("keeps current state when seq regressed (should not happen, but safe to guard)", () => {
      // seqNow < seqAtStart would be very unusual; still different → skip
      expect(applyFetchIfNotStale(A, 10, 5, B)).toBe(A);
    });
  });
});
