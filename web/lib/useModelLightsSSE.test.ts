import { describe, expect, it, vi } from "vitest";
import { createModelSSEHandlers } from "./useModelLightsSSE";
import type { ModelDetail } from "@/lib/models";

/** Minimal fake for SSE event dispatch — no React or browser needed. */
function makeFakeES() {
  let onopen: (() => void) | null = null;
  let onmessage: ((ev: { data: string }) => void) | null = null;
  let onerror: (() => void) | null = null;
  let closed = false;

  return {
    get onopen() {
      return onopen;
    },
    set onopen(fn: (() => void) | null) {
      onopen = fn;
    },
    get onmessage() {
      return onmessage;
    },
    set onmessage(fn: ((ev: { data: string }) => void) | null) {
      onmessage = fn;
    },
    get onerror() {
      return onerror;
    },
    set onerror(fn: (() => void) | null) {
      onerror = fn;
    },
    dispatch(data: unknown) {
      onmessage?.({ data: JSON.stringify(data) });
    },
    triggerOpen() {
      onopen?.();
    },
    triggerError() {
      onerror?.();
    },
    close() {
      closed = true;
    },
    get closed() {
      return closed;
    },
  };
}

function makeDelta(lightId: number) {
  return { light_id: lightId, on: true, color: "#ff0000", brightness_pct: 75 };
}

function makeHandlers(
  seqRef: { current: number | null },
  onReload: () => void,
  setModel?: (updater: (m: ModelDetail | null) => ModelDetail | null) => void,
  es?: ReturnType<typeof makeFakeES>,
) {
  const fakeEs = es ?? makeFakeES();
  const handlers = createModelSSEHandlers({
    seqRef,
    modelId: "model-1",
    setModel: setModel ?? vi.fn(),
    onReload,
    closeEs: () => fakeEs.close(),
  });
  fakeEs.onopen = handlers.onopen;
  fakeEs.onmessage = handlers.onmessage;
  fakeEs.onerror = handlers.onerror;
  return { fakeEs, handlers };
}

describe("createModelSSEHandlers", () => {
  describe("sequential events", () => {
    it("applies deltas for seq=1, seq=2 without calling onReload", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setModel = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload, setModel);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      fakeEs.dispatch({ seq: 2, deltas: [makeDelta(1)] });

      expect(onReload).not.toHaveBeenCalled();
      expect(setModel).toHaveBeenCalledTimes(2);
      expect(seqRef.current).toBe(2);
    });

    it("treats the first message as baseline regardless of seq value", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload);

      fakeEs.dispatch({ seq: 42, deltas: [makeDelta(0)] });

      expect(onReload).not.toHaveBeenCalled();
      expect(seqRef.current).toBe(42);
    });
  });

  describe("sequence gap", () => {
    it("calls onReload once and clears seqRef on gap (seq=1 then seq=3)", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      expect(seqRef.current).toBe(1);

      fakeEs.dispatch({ seq: 3, deltas: [makeDelta(1)] });

      expect(onReload).toHaveBeenCalledTimes(1);
      expect(seqRef.current).toBeNull();
    });

    it("does NOT apply deltas of the gap message", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setModel = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload, setModel);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      setModel.mockClear();

      fakeEs.dispatch({ seq: 3, deltas: [makeDelta(1)] });

      expect(setModel).not.toHaveBeenCalled();
    });

    it("accepts the next event after the gap as new baseline without another onReload", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setModel = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload, setModel);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      fakeEs.dispatch({ seq: 3, deltas: [makeDelta(1)] }); // gap → reload, seq cleared
      onReload.mockClear();
      setModel.mockClear();

      fakeEs.dispatch({ seq: 99, deltas: [makeDelta(2)] }); // new baseline

      expect(onReload).not.toHaveBeenCalled();
      expect(setModel).toHaveBeenCalledTimes(1);
      expect(seqRef.current).toBe(99);
    });
  });

  describe("onerror", () => {
    it("calls onReload once, clears seqRef, and closes the EventSource", () => {
      const seqRef = { current: 5 as number | null };
      const onReload = vi.fn();
      const fakeEs = makeFakeES();
      const { } = makeHandlers(seqRef, onReload, undefined, fakeEs);

      fakeEs.triggerError();

      expect(onReload).toHaveBeenCalledTimes(1);
      expect(seqRef.current).toBeNull();
      expect(fakeEs.closed).toBe(true);
    });

    it("accepts subsequent events as new baseline after onerror (no second onReload)", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setModel = vi.fn();
      const fakeEs = makeFakeES();
      makeHandlers(seqRef, onReload, setModel, fakeEs);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      fakeEs.triggerError(); // closes es, reloads, clears seq
      onReload.mockClear();
      setModel.mockClear();

      // Simulate a new connection's first event (new EventSource would be created, but
      // we reuse the handler logic — seqRef is null so next seq is treated as baseline).
      const seqRef2 = { current: null as number | null };
      const fakeEs2 = makeFakeES();
      makeHandlers(seqRef2, onReload, setModel, fakeEs2);

      fakeEs2.dispatch({ seq: 7, deltas: [makeDelta(1)] });

      expect(onReload).not.toHaveBeenCalled();
      expect(setModel).toHaveBeenCalledTimes(1);
      expect(seqRef2.current).toBe(7);
    });
  });

  describe("onopen", () => {
    it("sets sseLiveRef to true when provided", () => {
      const seqRef = { current: null as number | null };
      const sseLiveRef = { current: false };
      const fakeEs = makeFakeES();
      const handlers = createModelSSEHandlers({
        seqRef,
        sseLiveRef,
        modelId: "model-1",
        setModel: vi.fn(),
        onReload: vi.fn(),
        closeEs: () => fakeEs.close(),
      });
      fakeEs.onopen = handlers.onopen;
      fakeEs.onmessage = handlers.onmessage;
      fakeEs.onerror = handlers.onerror;

      fakeEs.triggerOpen();

      expect(sseLiveRef.current).toBe(true);
    });

    it("sets sseLiveRef to false on onerror", () => {
      const seqRef = { current: null as number | null };
      const sseLiveRef = { current: true };
      const fakeEs = makeFakeES();
      const handlers = createModelSSEHandlers({
        seqRef,
        sseLiveRef,
        modelId: "model-1",
        setModel: vi.fn(),
        onReload: vi.fn(),
        closeEs: () => fakeEs.close(),
      });
      fakeEs.onopen = handlers.onopen;
      fakeEs.onmessage = handlers.onmessage;
      fakeEs.onerror = handlers.onerror;

      fakeEs.triggerError();

      expect(sseLiveRef.current).toBe(false);
    });
  });

  describe("empty deltas", () => {
    it("does not call setModel when deltas array is empty", () => {
      const seqRef = { current: null as number | null };
      const setModel = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, vi.fn(), setModel);

      fakeEs.dispatch({ seq: 1, deltas: [] });

      expect(setModel).not.toHaveBeenCalled();
      expect(seqRef.current).toBe(1);
    });

    it("does not call setModel when deltas field is absent", () => {
      const seqRef = { current: null as number | null };
      const setModel = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, vi.fn(), setModel);

      fakeEs.dispatch({ seq: 1 });

      expect(setModel).not.toHaveBeenCalled();
    });
  });

  describe("invalid messages", () => {
    it("ignores non-JSON messages without throwing", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload);

      fakeEs.onmessage?.({ data: "not-json" });

      expect(onReload).not.toHaveBeenCalled();
      expect(seqRef.current).toBeNull();
    });

    it("ignores messages without a numeric seq", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload);

      fakeEs.onmessage?.({ data: JSON.stringify({ key: "started" }) });

      expect(onReload).not.toHaveBeenCalled();
    });
  });
});
