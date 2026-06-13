import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createCapturePoller } from "./capturePoller";
import type { CaptureStatus } from "./devices";

function makeStatus(
  state: CaptureStatus["state"],
  current_index = 0,
  light_count = 5,
): CaptureStatus {
  return { state, current_index, light_count };
}

describe("createCapturePoller", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls getCaptureStatus on each tick and updates progress", async () => {
    const statuses: CaptureStatus[] = [
      makeStatus("running", 0),
      makeStatus("running", 1),
      makeStatus("running", 2),
    ];
    let call = 0;
    const getFn = vi.fn(async () => statuses[call++] ?? makeStatus("idle"));

    const received: CaptureStatus[] = [];
    const stop = createCapturePoller("dev-1", getFn, {
      onStatus: (s) => received.push(s),
      onError: vi.fn(),
    });

    // tick 1
    await vi.advanceTimersByTimeAsync(1000);
    expect(received).toHaveLength(1);
    expect(received[0].current_index).toBe(0);

    // tick 2
    await vi.advanceTimersByTimeAsync(1000);
    expect(received).toHaveLength(2);
    expect(received[1].current_index).toBe(1);

    // tick 3
    await vi.advanceTimersByTimeAsync(1000);
    expect(received).toHaveLength(3);
    expect(received[2].current_index).toBe(2);

    stop();
  });

  it("stops polling when state transitions to idle", async () => {
    const statuses: CaptureStatus[] = [
      makeStatus("running", 0),
      makeStatus("idle"),
    ];
    let call = 0;
    const getFn = vi.fn(async () => statuses[call++] ?? makeStatus("idle"));

    const received: CaptureStatus[] = [];
    const errorSpy = vi.fn();
    const stop = createCapturePoller("dev-1", getFn, {
      onStatus: (s) => received.push(s),
      onError: errorSpy,
    });

    await vi.advanceTimersByTimeAsync(1000);
    expect(received[0].state).toBe("running");

    await vi.advanceTimersByTimeAsync(1000);
    expect(received[1].state).toBe("idle");

    // no further calls after idle
    await vi.advanceTimersByTimeAsync(3000);
    expect(getFn).toHaveBeenCalledTimes(2);
    expect(errorSpy).not.toHaveBeenCalled();

    stop();
  });

  it("does not stop polling on 'stopping' state", async () => {
    const statuses: CaptureStatus[] = [
      makeStatus("running", 0),
      makeStatus("stopping", 1),
      makeStatus("idle"),
    ];
    let call = 0;
    const getFn = vi.fn(async () => statuses[call++] ?? makeStatus("idle"));

    const received: CaptureStatus[] = [];
    const stop = createCapturePoller("dev-1", getFn, {
      onStatus: (s) => received.push(s),
      onError: vi.fn(),
    });

    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(1000);
    expect(received[1].state).toBe("stopping");

    await vi.advanceTimersByTimeAsync(1000);
    expect(received[2].state).toBe("idle");

    await vi.advanceTimersByTimeAsync(2000);
    expect(getFn).toHaveBeenCalledTimes(3);

    stop();
  });

  it("calls onError and stops after maxConsecutiveErrors failures", async () => {
    const getFn = vi.fn(async () => {
      throw new Error("network error");
    });

    const statusSpy = vi.fn();
    const errorSpy = vi.fn();
    const stop = createCapturePoller(
      "dev-1",
      getFn,
      { onStatus: statusSpy, onError: errorSpy },
      1000,
      3,
    );

    // 3 failing ticks
    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(1000);

    expect(errorSpy).toHaveBeenCalledOnce();
    expect(errorSpy).toHaveBeenCalledWith("network error");
    expect(statusSpy).not.toHaveBeenCalled();

    // no further calls after stopping
    await vi.advanceTimersByTimeAsync(3000);
    expect(getFn).toHaveBeenCalledTimes(3);

    stop();
  });

  it("resets consecutive error count after a successful poll", async () => {
    let call = 0;
    const getFn = vi.fn(async () => {
      call++;
      if (call === 2) throw new Error("transient");
      return makeStatus("running", call - 1);
    });

    const errorSpy = vi.fn();
    const received: CaptureStatus[] = [];
    const stop = createCapturePoller(
      "dev-1",
      getFn,
      { onStatus: (s) => received.push(s), onError: errorSpy },
      1000,
      3,
    );

    await vi.advanceTimersByTimeAsync(1000); // call 1: success
    await vi.advanceTimersByTimeAsync(1000); // call 2: error (1 consecutive)
    await vi.advanceTimersByTimeAsync(1000); // call 3: success → resets counter
    await vi.advanceTimersByTimeAsync(1000); // call 4: success

    expect(errorSpy).not.toHaveBeenCalled();
    expect(received).toHaveLength(3);

    stop();
  });

  it("cleanup stops the interval immediately (no calls after stop)", async () => {
    const getFn = vi.fn(async () => makeStatus("running", 0));
    const statusSpy = vi.fn();

    const stop = createCapturePoller("dev-1", getFn, {
      onStatus: statusSpy,
      onError: vi.fn(),
    });

    await vi.advanceTimersByTimeAsync(1000);
    expect(statusSpy).toHaveBeenCalledTimes(1);

    stop();

    await vi.advanceTimersByTimeAsync(5000);
    expect(getFn).toHaveBeenCalledTimes(1);
  });

  it("ignores in-flight responses arriving after stop", async () => {
    let resolvePoll!: (s: CaptureStatus) => void;
    const getFn = vi.fn(
      () =>
        new Promise<CaptureStatus>((res) => {
          resolvePoll = res;
        }),
    );

    const statusSpy = vi.fn();
    const stop = createCapturePoller("dev-1", getFn, {
      onStatus: statusSpy,
      onError: vi.fn(),
    });

    await vi.advanceTimersByTimeAsync(1000); // tick fires, in-flight
    stop(); // cancel before response arrives
    resolvePoll(makeStatus("running", 0)); // resolve late
    await Promise.resolve(); // flush microtask

    expect(statusSpy).not.toHaveBeenCalled();
  });
});
