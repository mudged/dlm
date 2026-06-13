import { afterEach, describe, expect, it, vi } from "vitest";
import { createSceneSSEHandlers } from "./useSceneLightsSSE";
import { manageSSEConnection } from "./sseReconnect";
import type { SceneDetail } from "@/lib/scenes";

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
  return { light_id: lightId, on: true, color: "#00ff00", brightness_pct: 50 };
}

function makeHandlers(
  seqRef: { current: number | null },
  onReload: () => void,
  setScene?: (updater: (s: SceneDetail | null) => SceneDetail | null) => void,
  es?: ReturnType<typeof makeFakeES>,
) {
  const fakeEs = es ?? makeFakeES();
  const handlers = createSceneSSEHandlers({
    seqRef,
    sceneId: "scene-1",
    setScene: setScene ?? vi.fn(),
    onReload,
    closeEs: () => fakeEs.close(),
  });
  fakeEs.onopen = handlers.onopen;
  fakeEs.onmessage = handlers.onmessage;
  fakeEs.onerror = handlers.onerror;
  return { fakeEs, handlers };
}

// ---------------------------------------------------------------------------
// createSceneSSEHandlers unit tests
// ---------------------------------------------------------------------------

describe("createSceneSSEHandlers", () => {
  describe("sequential events", () => {
    it("applies deltas for seq=1, seq=2 without calling onReload", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setScene = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload, setScene);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      fakeEs.dispatch({ seq: 2, deltas: [makeDelta(1)] });

      expect(onReload).not.toHaveBeenCalled();
      expect(setScene).toHaveBeenCalledTimes(2);
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
      fakeEs.dispatch({ seq: 3, deltas: [makeDelta(1)] });

      expect(onReload).toHaveBeenCalledTimes(1);
      expect(seqRef.current).toBeNull();
    });

    it("does NOT apply deltas of the gap message", () => {
      const seqRef = { current: null as number | null };
      const onReload = vi.fn();
      const setScene = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, onReload, setScene);

      fakeEs.dispatch({ seq: 1, deltas: [makeDelta(0)] });
      setScene.mockClear();

      fakeEs.dispatch({ seq: 3, deltas: [makeDelta(1)] });

      expect(setScene).not.toHaveBeenCalled();
    });
  });

  describe("onerror", () => {
    it("calls onReload once, clears seqRef, and closes the EventSource", () => {
      const seqRef = { current: 5 as number | null };
      const onReload = vi.fn();
      const fakeEs = makeFakeES();
      makeHandlers(seqRef, onReload, undefined, fakeEs);

      fakeEs.triggerError();

      expect(onReload).toHaveBeenCalledTimes(1);
      expect(seqRef.current).toBeNull();
      expect(fakeEs.closed).toBe(true);
    });
  });

  describe("onopen", () => {
    it("sets sseLiveRef to true when provided", () => {
      const seqRef = { current: null as number | null };
      const sseLiveRef = { current: false };
      const fakeEs = makeFakeES();
      const handlers = createSceneSSEHandlers({
        seqRef,
        sseLiveRef,
        sceneId: "scene-1",
        setScene: vi.fn(),
        onReload: vi.fn(),
        closeEs: () => fakeEs.close(),
      });
      fakeEs.onopen = handlers.onopen;
      fakeEs.onerror = handlers.onerror;

      fakeEs.triggerOpen();

      expect(sseLiveRef.current).toBe(true);
    });

    it("sets sseLiveRef to false on onerror", () => {
      const seqRef = { current: null as number | null };
      const sseLiveRef = { current: true };
      const fakeEs = makeFakeES();
      const handlers = createSceneSSEHandlers({
        seqRef,
        sseLiveRef,
        sceneId: "scene-1",
        setScene: vi.fn(),
        onReload: vi.fn(),
        closeEs: () => fakeEs.close(),
      });
      fakeEs.onopen = handlers.onopen;
      fakeEs.onerror = handlers.onerror;

      fakeEs.triggerError();

      expect(sseLiveRef.current).toBe(false);
    });
  });

  describe("empty deltas", () => {
    it("does not call setScene when deltas array is empty", () => {
      const seqRef = { current: null as number | null };
      const setScene = vi.fn();
      const { fakeEs } = makeHandlers(seqRef, vi.fn(), setScene);

      fakeEs.dispatch({ seq: 1, deltas: [] });

      expect(setScene).not.toHaveBeenCalled();
      expect(seqRef.current).toBe(1);
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

// ---------------------------------------------------------------------------
// Reconnect with backoff — tests for manageSSEConnection + createSceneSSEHandlers
// ---------------------------------------------------------------------------

function makeReconnectSession(
  seqRef: { current: number | null },
  onReload: () => void,
  setScene?: (updater: (s: SceneDetail | null) => SceneDetail | null) => void,
  overrides?: { backoffStepsMs?: readonly number[]; jitterMs?: number },
) {
  const makeEs = vi.fn(() => makeFakeES());

  const session = manageSSEConnection({
    makeEs: makeEs as unknown as () => EventSource,
    attach(es, { scheduleReconnect, resetBackoff }) {
      const fakeEs = es as unknown as ReturnType<typeof makeFakeES>;
      const handlers = createSceneSSEHandlers({
        seqRef,
        sceneId: "scene-1",
        setScene: setScene ?? vi.fn(),
        onReload,
        closeEs: () => fakeEs.close(),
      });
      fakeEs.onopen = () => {
        resetBackoff();
        handlers.onopen();
      };
      fakeEs.onmessage = handlers.onmessage;
      fakeEs.onerror = () => {
        handlers.onerror();
        scheduleReconnect();
      };
    },
    backoffStepsMs: overrides?.backoffStepsMs ?? [100, 200, 500, 1_000],
    jitterMs: overrides?.jitterMs ?? 0,
  });

  const getEs = (callIndex = makeEs.mock.calls.length - 1) =>
    makeEs.mock.results[callIndex].value as ReturnType<typeof makeFakeES>;

  return { makeEs, session, getEs };
}

describe("SSE reconnect with backoff (manageSSEConnection + createSceneSSEHandlers)", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("creates a new EventSource after the first backoff delay", () => {
    vi.useFakeTimers();
    const { makeEs, session, getEs } = makeReconnectSession(
      { current: null },
      vi.fn(),
    );

    expect(makeEs).toHaveBeenCalledTimes(1);

    getEs(0).triggerError();
    expect(makeEs).toHaveBeenCalledTimes(1);

    vi.advanceTimersByTime(99);
    expect(makeEs).toHaveBeenCalledTimes(1);

    vi.advanceTimersByTime(1); // 100 ms total → fires
    expect(makeEs).toHaveBeenCalledTimes(2);

    session.destroy();
  });

  it("backoff delay grows on repeated failures", () => {
    vi.useFakeTimers();
    const { makeEs, session, getEs } = makeReconnectSession(
      { current: null },
      vi.fn(),
    );

    // attempt 0 → 100 ms
    getEs(0).triggerError();
    vi.advanceTimersByTime(100);
    expect(makeEs).toHaveBeenCalledTimes(2);

    // attempt 1 → 200 ms; 100 ms alone is not enough
    getEs(1).triggerError();
    vi.advanceTimersByTime(100);
    expect(makeEs).toHaveBeenCalledTimes(2);

    vi.advanceTimersByTime(100); // 200 ms total
    expect(makeEs).toHaveBeenCalledTimes(3);

    session.destroy();
  });

  it("resets backoff to base delay after a successful onopen", () => {
    vi.useFakeTimers();
    const { makeEs, session, getEs } = makeReconnectSession(
      { current: null },
      vi.fn(),
    );

    // First error → 100 ms
    getEs(0).triggerError();
    vi.advanceTimersByTime(100);
    expect(makeEs).toHaveBeenCalledTimes(2);

    // Successful open resets backoff
    getEs(1).triggerOpen();

    // Next error should use base delay again (100 ms)
    getEs(1).triggerError();
    vi.advanceTimersByTime(100);
    expect(makeEs).toHaveBeenCalledTimes(3);

    session.destroy();
  });

  it("cancels the pending reconnect timer on destroy", () => {
    vi.useFakeTimers();
    const { makeEs, session, getEs } = makeReconnectSession(
      { current: null },
      vi.fn(),
    );

    getEs(0).triggerError();
    session.destroy();

    vi.advanceTimersByTime(1_000);
    expect(makeEs).toHaveBeenCalledTimes(1);
  });

  it("does not reconnect after destroy even if error fires late", () => {
    vi.useFakeTimers();
    const { makeEs, session, getEs } = makeReconnectSession(
      { current: null },
      vi.fn(),
    );

    session.destroy();
    getEs(0).triggerError();

    vi.advanceTimersByTime(1_000);
    expect(makeEs).toHaveBeenCalledTimes(1);
  });

  it("calls onReload and clears seqRef on error before reconnecting", () => {
    vi.useFakeTimers();
    const seqRef = { current: 7 as number | null };
    const onReload = vi.fn();
    const { session, getEs } = makeReconnectSession(seqRef, onReload);

    getEs(0).triggerError();

    expect(onReload).toHaveBeenCalledTimes(1);
    expect(seqRef.current).toBeNull();

    session.destroy();
  });
});
