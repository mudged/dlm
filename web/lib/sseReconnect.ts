export const DEFAULT_BACKOFF_STEPS_MS: readonly number[] = [1_000, 2_000, 5_000, 15_000];
export const DEFAULT_JITTER_MS = 500;

/**
 * Returns the reconnect delay for the given attempt count.
 * Delay grows across `steps` and is capped at the last step value, plus random jitter.
 */
export function getBackoffDelayMs(
  attempt: number,
  steps: readonly number[] = DEFAULT_BACKOFF_STEPS_MS,
  jitter: number = DEFAULT_JITTER_MS,
): number {
  const base = steps[Math.min(attempt, steps.length - 1)];
  return base + Math.random() * jitter;
}

export interface ReconnectSession {
  /** Close the active connection and cancel any pending reconnect timer. Call on unmount. */
  destroy(): void;
}

/**
 * Opens an EventSource and automatically reconnects with bounded exponential backoff.
 *
 * Usage:
 * - In `attach`, wire up `onopen`/`onmessage`/`onerror`.
 * - Call `ctx.scheduleReconnect()` from within `onerror` to queue a backoff reconnect.
 * - Call `ctx.resetBackoff()` from within `onopen` to reset the attempt counter to 0.
 * - Call `session.destroy()` on effect cleanup to close the connection and cancel timers.
 */
export function manageSSEConnection(options: {
  makeEs: () => EventSource;
  attach: (
    es: EventSource,
    ctx: { scheduleReconnect: () => void; resetBackoff: () => void },
  ) => void;
  backoffStepsMs?: readonly number[];
  jitterMs?: number;
}): ReconnectSession {
  const {
    makeEs,
    attach,
    backoffStepsMs = DEFAULT_BACKOFF_STEPS_MS,
    jitterMs = DEFAULT_JITTER_MS,
  } = options;

  let destroyed = false;
  let attempt = 0;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let currentEs: EventSource | null = null;

  function scheduleReconnect(): void {
    if (destroyed) return;
    const delay = getBackoffDelayMs(attempt, backoffStepsMs, jitterMs);
    attempt += 1;
    timer = setTimeout(() => {
      timer = null;
      connect();
    }, delay);
  }

  function resetBackoff(): void {
    attempt = 0;
  }

  function connect(): void {
    if (destroyed) return;
    const es = makeEs();
    currentEs = es;
    attach(es, { scheduleReconnect, resetBackoff });
  }

  connect();

  return {
    destroy() {
      destroyed = true;
      if (timer !== null) {
        clearTimeout(timer);
        timer = null;
      }
      currentEs?.close();
      currentEs = null;
    },
  };
}
