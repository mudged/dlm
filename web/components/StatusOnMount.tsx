"use client";

import { useEffect, useState } from "react";

type StatusPayload = { service: string; version: string };

export function StatusOnMount() {
  const [data, setData] = useState<StatusPayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch("/api/v1/status", { cache: "no-store" });
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const json = (await res.json()) as StatusPayload;
        if (!cancelled) {
          setData(json);
          setError(null);
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "request failed");
          setData(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="rounded-xl border border-slate-200 bg-white/80 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-900/60">
      <h2 className="text-sm font-medium text-slate-700 dark:text-slate-200">
        Client load (mount → same-origin /api/v1)
      </h2>
      <p className="mt-2 text-sm leading-relaxed text-slate-600 dark:text-slate-400">
        {error ? (
          <span className="text-red-600 dark:text-red-400" role="status">
            {error}
          </span>
        ) : data ? (
          <>
            Connected to{" "}
            <strong className="text-slate-800 dark:text-slate-100">
              {data.service}
            </strong>{" "}
            version{" "}
            <strong className="text-slate-800 dark:text-slate-100">
              {data.version}
            </strong>
            .
          </>
        ) : (
          "Loading…"
        )}
      </p>
    </div>
  );
}
