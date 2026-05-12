/** Client for `public/dlm-python-editor-worker.mjs` (REQ-022 / §3.17). */

export type PyLintDiagnostic = {
  line: number;
  column: number;
  message: string;
};

export type FormatResult =
  | { ok: true; text: string; usedBlack: boolean }
  | { ok: false; error: string };

const PYTHON_EDITOR_WORKER_TIMEOUT_MS = 15_000;

let worker: Worker | null = null;
let nextId = 1;
const pending = new Map<
  number,
  {
    resolve: (v: unknown) => void;
    reject: (reason?: unknown) => void;
    timeoutId: ReturnType<typeof setTimeout>;
  }
>();

function describeWorkerFailure(ev: Event): string {
  const detail = ev as Event & {
    message?: unknown;
    error?: unknown;
    data?: unknown;
  };
  if (typeof detail.message === "string" && detail.message.length > 0) {
    return detail.message;
  }
  if (detail.error instanceof Error && detail.error.message.length > 0) {
    return detail.error.message;
  }
  if (detail.error !== undefined && detail.error !== null) {
    return String(detail.error);
  }
  if (detail.data !== undefined && detail.data !== null) {
    return String(detail.data);
  }
  return ev.type || "unknown";
}

function rejectAllPending(reason: string) {
  const err = new Error(`python editor worker failed: ${reason}`);
  for (const slot of pending.values()) {
    clearTimeout(slot.timeoutId);
    slot.reject(err);
  }
  pending.clear();
}

function handleWorkerFailure(ev: Event) {
  const failedWorker = worker;
  rejectAllPending(describeWorkerFailure(ev));
  worker = null;
  failedWorker?.terminate();
}

function getWorker(): Worker {
  if (typeof window === "undefined") {
    throw new Error("pythonEditorWorker requires a browser");
  }
  if (!worker) {
    const url = new URL(
      "/dlm-python-editor-worker.mjs",
      window.location.origin,
    );
    worker = new Worker(url, { type: "module" });
    worker.onmessage = (ev: MessageEvent) => {
      const m = ev.data as {
        type: string;
        id: number;
        diagnostics?: PyLintDiagnostic[];
        ok?: boolean;
        text?: string;
        usedBlack?: boolean;
        error?: string;
      };
      const slot = pending.get(m.id);
      if (!slot) {
        return;
      }
      pending.delete(m.id);
      clearTimeout(slot.timeoutId);
      if (m.type === "lintResult") {
        slot.resolve(m.diagnostics ?? []);
        return;
      }
      if (m.type === "formatResult") {
        if (m.ok) {
          slot.resolve({
            ok: true,
            text: m.text ?? "",
            usedBlack: Boolean(m.usedBlack),
          } satisfies FormatResult);
        } else {
          slot.resolve({
            ok: false,
            error: m.error ?? "Format failed",
          } satisfies FormatResult);
        }
      }
    };
    worker.onerror = handleWorkerFailure;
    worker.onmessageerror = handleWorkerFailure;
  }
  return worker;
}

function postWorkerRequest<T>(type: "lint" | "format", source: string): Promise<T> {
  const id = nextId++;
  const w = getWorker();
  return new Promise((resolve, reject) => {
    const timeoutId = setTimeout(() => {
      pending.delete(id);
      reject(new Error("python editor worker timed out"));
    }, PYTHON_EDITOR_WORKER_TIMEOUT_MS);
    pending.set(id, {
      resolve: resolve as (v: unknown) => void,
      reject,
      timeoutId,
    });
    try {
      w.postMessage({ type, id, source });
    } catch (e) {
      clearTimeout(timeoutId);
      pending.delete(id);
      reject(e);
    }
  });
}

export function lintPythonSource(source: string): Promise<PyLintDiagnostic[]> {
  return postWorkerRequest<PyLintDiagnostic[]>("lint", source);
}

export function formatPythonSource(source: string): Promise<FormatResult> {
  return postWorkerRequest<FormatResult>("format", source);
}
