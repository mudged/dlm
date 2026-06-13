import { useEffect } from "react";
import { createCapturePoller } from "./capturePoller";
import type { CaptureStatus } from "./devices";
import { getCaptureStatus } from "./devices";

/**
 * Starts capture-status polling whenever `state` is "running" or "stopping".
 * Polling stops automatically when the state transitions to a terminal value,
 * on N consecutive errors, or on unmount.
 */
export function useCapturePolling(
  deviceId: string | null,
  state: CaptureStatus["state"] | undefined,
  setCaptureStatus: (s: CaptureStatus) => void,
  setCaptureError: (err: string | null) => void,
) {
  const shouldPoll = state === "running" || state === "stopping";

  useEffect(() => {
    if (!deviceId || !shouldPoll) return;

    const stop = createCapturePoller(deviceId, getCaptureStatus, {
      onStatus: setCaptureStatus,
      onError: (msg) => setCaptureError(msg),
    });

    return stop;
  }, [deviceId, shouldPoll, setCaptureStatus, setCaptureError]);
}
