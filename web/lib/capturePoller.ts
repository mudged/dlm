import type { CaptureStatus } from "./devices";

const POLLING_INTERVAL_MS = 1000;
const MAX_CONSECUTIVE_ERRORS = 3;

export type CapturePollerCallbacks = {
  onStatus: (status: CaptureStatus) => void;
  onError: (message: string) => void;
};

/**
 * Starts a polling interval that calls `getCaptureStatusFn` every `intervalMs`.
 * Stops automatically when the returned state is no longer "running" or "stopping".
 * Stops and calls `onError` after `maxConsecutiveErrors` consecutive failures.
 * Returns a cleanup function that cancels the interval.
 */
export function createCapturePoller(
  deviceId: string,
  getCaptureStatusFn: (id: string) => Promise<CaptureStatus>,
  callbacks: CapturePollerCallbacks,
  intervalMs = POLLING_INTERVAL_MS,
  maxConsecutiveErrors = MAX_CONSECUTIVE_ERRORS,
): () => void {
  let stopped = false;
  let consecutiveErrors = 0;

  const timer = setInterval(() => {
    void (async () => {
      if (stopped) return;
      try {
        const s = await getCaptureStatusFn(deviceId);
        if (stopped) return;
        consecutiveErrors = 0;
        callbacks.onStatus(s);
        if (s.state !== "running" && s.state !== "stopping") {
          clearInterval(timer);
          stopped = true;
        }
      } catch (e) {
        if (stopped) return;
        consecutiveErrors += 1;
        if (consecutiveErrors >= maxConsecutiveErrors) {
          callbacks.onError(e instanceof Error ? e.message : "Polling failed.");
          clearInterval(timer);
          stopped = true;
        }
      }
    })();
  }, intervalMs);

  return () => {
    stopped = true;
    clearInterval(timer);
  };
}
