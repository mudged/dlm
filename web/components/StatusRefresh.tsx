"use client";

import { faRotate } from "@fortawesome/free-solid-svg-icons";
import { useCallback, useState } from "react";
import { Button } from "@/components/ui/Button";

type StatusPayload = { service: string; version: string };

async function fetchStatus(): Promise<StatusPayload> {
  const res = await fetch("/api/v1/status", { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`status ${res.status}`);
  }
  return res.json() as Promise<StatusPayload>;
}

export function StatusRefresh() {
  const [data, setData] = useState<StatusPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onRefresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(await fetchStatus());
    } catch (e) {
      setError(e instanceof Error ? e.message : "request failed");
      setData(null);
    } finally {
      setLoading(false);
    }
  }, []);

  return (
    <div className="rounded-xl border border-slate-200 bg-white/80 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-900/60">
      <p className="text-sm font-medium text-slate-700 dark:text-slate-200">
        Manual refresh (same-origin API; dev uses Next rewrite)
      </p>
      <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
        <Button
          icon={faRotate}
          onClick={onRefresh}
          disabled={loading}
          aria-busy={loading}
        >
          {loading ? "Loading…" : "Refresh API status"}
        </Button>
        {error ? (
          <span className="text-sm text-red-600 dark:text-red-400" role="status">
            {error}
          </span>
        ) : null}
        {data ? (
          <span className="text-sm text-slate-600 dark:text-slate-300" role="status">
            {data.service} v{data.version}
          </span>
        ) : null}
      </div>
    </div>
  );
}
