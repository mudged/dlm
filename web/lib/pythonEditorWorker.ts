/** Client for `public/dlm-python-editor-worker.mjs` (REQ-022 / §3.17). */

export type PyLintDiagnostic = {
  line: number;
  column: number;
  message: string;
};

export type FormatResult =
  | { ok: true; text: string; usedBlack: boolean }
  | { ok: false; error: string };

let worker: Worker | null = null;
let nextId = 1;
const pending = new Map<
  number,
  {
    resolve: (v: unknown) => void;
  }
>();

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
  }
  return worker;
}

export function lintPythonSource(source: string): Promise<PyLintDiagnostic[]> {
  const id = nextId++;
  const w = getWorker();
  return new Promise((resolve) => {
    pending.set(id, { resolve: resolve as (v: unknown) => void });
    w.postMessage({ type: "lint", id, source });
  });
}

export function formatPythonSource(source: string): Promise<FormatResult> {
  const id = nextId++;
  const w = getWorker();
  return new Promise((resolve) => {
    pending.set(id, { resolve: resolve as (v: unknown) => void });
    w.postMessage({ type: "format", id, source });
  });
}
