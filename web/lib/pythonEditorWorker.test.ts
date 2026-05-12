import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

type WorkerMessage = {
  type: string;
  id: number;
  source?: string;
};

class MockWorker {
  static instances: MockWorker[] = [];

  onmessage: ((ev: MessageEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  onmessageerror: ((ev: MessageEvent) => void) | null = null;
  readonly messages: WorkerMessage[] = [];
  terminated = false;

  constructor(
    readonly url: URL,
    readonly options?: WorkerOptions,
  ) {
    MockWorker.instances.push(this);
  }

  postMessage(msg: WorkerMessage) {
    this.messages.push(msg);
  }

  terminate() {
    this.terminated = true;
  }

  simulateResponse(id: number, payload: Record<string, unknown>) {
    this.onmessage?.({ data: { id, ...payload } } as MessageEvent);
  }

  simulateError(message: string) {
    this.onerror?.({ type: "error", message } as Event);
  }

  simulateMessageError(message: string) {
    this.onmessageerror?.({
      type: "messageerror",
      data: message,
    } as MessageEvent);
  }
}

async function importClient() {
  return import("./pythonEditorWorker");
}

beforeEach(() => {
  vi.resetModules();
  MockWorker.instances = [];
  vi.stubGlobal("window", { location: { origin: "http://localhost:3000" } });
  vi.stubGlobal("Worker", MockWorker);
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("pythonEditorWorker", () => {
  it("resolves responses by matching request id", async () => {
    const { lintPythonSource } = await importClient();

    const result = lintPythonSource("print('ok')");
    const worker = MockWorker.instances[0];
    expect(worker.messages[0]).toMatchObject({ type: "lint", id: 1 });

    worker.simulateResponse(1, {
      type: "lintResult",
      diagnostics: [{ line: 1, column: 1, message: "Syntax error" }],
    });

    await expect(result).resolves.toEqual([
      { line: 1, column: 1, message: "Syntax error" },
    ]);
  });

  it("rejects pending requests on worker error and respawns on the next request", async () => {
    const { lintPythonSource } = await importClient();

    const failed = lintPythonSource("bad");
    const firstWorker = MockWorker.instances[0];
    firstWorker.simulateError("boom");

    await expect(failed).rejects.toThrow(
      "python editor worker failed: boom",
    );
    expect(firstWorker.terminated).toBe(true);

    const recovered = lintPythonSource("print('ok')");
    expect(MockWorker.instances).toHaveLength(2);
    const secondWorker = MockWorker.instances[1];
    secondWorker.simulateResponse(2, {
      type: "lintResult",
      diagnostics: [],
    });

    await expect(recovered).resolves.toEqual([]);
  });

  it("rejects pending requests on message error and respawns on the next request", async () => {
    const { formatPythonSource } = await importClient();

    const failed = formatPythonSource("x=1");
    const firstWorker = MockWorker.instances[0];
    firstWorker.simulateMessageError("could not deserialize response");

    await expect(failed).rejects.toThrow(
      "python editor worker failed: could not deserialize response",
    );
    expect(firstWorker.terminated).toBe(true);

    const recovered = formatPythonSource("x=1");
    expect(MockWorker.instances).toHaveLength(2);
    const secondWorker = MockWorker.instances[1];
    secondWorker.simulateResponse(2, {
      type: "formatResult",
      ok: true,
      text: "x = 1\n",
      usedBlack: true,
    });

    await expect(recovered).resolves.toEqual({
      ok: true,
      text: "x = 1\n",
      usedBlack: true,
    });
  });

  it("rejects only the timed-out request and keeps the worker alive", async () => {
    vi.useFakeTimers();
    const { lintPythonSource } = await importClient();

    const timedOut = lintPythonSource("bad");
    const worker = MockWorker.instances[0];

    vi.advanceTimersByTime(15_001);
    await expect(timedOut).rejects.toThrow("python editor worker timed out");
    expect(worker.terminated).toBe(false);

    const recovered = lintPythonSource("print('ok')");
    expect(MockWorker.instances).toHaveLength(1);
    worker.simulateResponse(2, {
      type: "lintResult",
      diagnostics: [],
    });

    await expect(recovered).resolves.toEqual([]);
  });

  it("leaves other pending requests untouched when one request times out", async () => {
    vi.useFakeTimers();
    const { lintPythonSource } = await importClient();

    const first = lintPythonSource("first");
    const second = lintPythonSource("second");
    const worker = MockWorker.instances[0];

    vi.advanceTimersByTime(14_999);
    worker.simulateResponse(2, {
      type: "lintResult",
      diagnostics: [],
    });
    vi.advanceTimersByTime(2);

    await expect(first).rejects.toThrow("python editor worker timed out");
    await expect(second).resolves.toEqual([]);
  });
});
