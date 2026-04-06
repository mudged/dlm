"use client";

import { useEffect, useRef, useState } from "react";

type WorkerOut =
  | { type: "ready" }
  | { type: "done" }
  | { type: "error"; message: string }
  | { type: "iterationComplete"; sceneId: string };

/** Architecture §3.17 — default `T_force` after cooperative stop. */
const T_FORCE_MS = 5000;

/**
 * Runs user Python in a Pyodide worker while a python_scene_script routine is active on the scene.
 * REQ-022 / architecture §3.17.
 */
export function PythonRoutineHost(props: {
  sceneId: string;
  source: string;
  onWorkerMessage?: (msg: string) => void;
  /** Fired after each successful script iteration (worker refreshed dims + ran user code). */
  onIterationComplete?: (sceneId: string) => void;
}) {
  const { sceneId, source, onWorkerMessage, onIterationComplete } = props;
  const [phase, setPhase] = useState<
    "loading" | "running" | "finished" | "error"
  >("loading");
  const workerRef = useRef<Worker | null>(null);
  const onMsgRef = useRef(onWorkerMessage);
  onMsgRef.current = onWorkerMessage;
  const onIterRef = useRef(onIterationComplete);
  onIterRef.current = onIterationComplete;

  useEffect(() => {
    setPhase("loading");
    const url = new URL(
      "/dlm-python-scene-worker.mjs",
      window.location.origin,
    );
    const w = new Worker(url, { type: "module" });
    workerRef.current = w;

    let forceTimer: number | null = null;
    let unmounted = false;

    const clearForceTimer = () => {
      if (forceTimer != null) {
        window.clearTimeout(forceTimer);
        forceTimer = null;
      }
    };

    const terminateWorker = () => {
      clearForceTimer();
      try {
        w.terminate();
      } catch {
        /* ignore */
      }
    };

    const armForceTerminate = () => {
      clearForceTimer();
      forceTimer = window.setTimeout(() => {
        forceTimer = null;
        try {
          w.terminate();
        } catch {
          /* ignore */
        }
      }, T_FORCE_MS);
    };

    w.onmessage = (ev: MessageEvent<WorkerOut>) => {
      const m = ev.data;
      if (m.type === "done") {
        terminateWorker();
        if (!unmounted) {
          setPhase("finished");
        }
        return;
      }
      if (m.type === "ready") {
        if (!unmounted) {
          setPhase("running");
        }
        return;
      }
      if (m.type === "iterationComplete") {
        onIterRef.current?.(m.sceneId);
        return;
      }
      if (m.type === "error") {
        if (!unmounted) {
          setPhase("error");
          onMsgRef.current?.(m.message);
        }
      }
    };

    w.postMessage({
      type: "init",
      sceneId,
      source,
    });

    return () => {
      unmounted = true;
      try {
        w.postMessage({ type: "stop" });
      } catch {
        /* ignore */
      }
      armForceTerminate();
      workerRef.current = null;
    };
  }, [sceneId, source]);

  if (phase === "loading") {
    return (
      <p className="text-xs text-slate-600 dark:text-slate-400">
        Loading Python runtime (Pyodide from CDN; first run may take a while)…
      </p>
    );
  }
  if (phase === "error") {
    return (
      <p className="text-xs text-amber-800 dark:text-amber-200">
        Python runner reported an error (see message above if shown). Stop and
        fix the script, then start again.
      </p>
    );
  }
  if (phase === "finished") {
    return (
      <p className="text-xs text-slate-600 dark:text-slate-400">
        Python routine stopped. Start the routine again from the list if you
        want another run.
      </p>
    );
  }
  return (
    <p className="text-xs text-emerald-800 dark:text-emerald-200">
      Python routine is executing in your browser. Stop the routine to end
      execution; if the worker does not exit cooperatively, it is terminated
      after up to {T_FORCE_MS / 1000} seconds (
      <code className="rounded bg-slate-200 px-0.5 dark:bg-slate-800">
        T_force
      </code>{" "}
      per architecture §3.17).
    </p>
  );
}
