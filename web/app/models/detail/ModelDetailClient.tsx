"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import type { ModelDetail } from "@/lib/models";

export function ModelDetailClient() {
  const router = useRouter();
  const params = useSearchParams();
  const id = params.get("id");
  const [model, setModel] = useState<ModelDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) {
      setError("Missing model id.");
      setLoading(false);
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/models/${encodeURIComponent(id)}`, {
        cache: "no-store",
      });
      const j = (await res.json().catch(() => null)) as ModelDetail & {
        error?: { message?: string };
      };
      if (!res.ok) {
        setError(j?.error?.message ?? `Could not load model (${res.status})`);
        setModel(null);
        setLoading(false);
        return;
      }
      setModel(j as ModelDetail);
    } catch {
      setError("Could not reach the API.");
      setModel(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  if (!id) {
    return (
      <p className="text-sm text-amber-800 dark:text-amber-200" role="alert">
        Missing model id. Return to the list and open a model again.
      </p>
    );
  }

  if (loading) {
    return <p className="text-sm text-slate-500">Loading…</p>;
  }

  if (error || !model) {
    return (
      <p
        className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
        role="alert"
      >
        {error ?? "Model not found."}
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {model.name}
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          {model.light_count} lights ·{" "}
          {new Date(model.created_at).toLocaleString()}
        </p>
      </header>

      {model.lights.length === 0 ? (
        <p className="text-sm text-slate-600 dark:text-slate-400">
          This model has no lights (header-only CSV).
        </p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
          <table className="w-full min-w-[28rem] text-left text-sm">
            <thead className="bg-slate-100 dark:bg-slate-800/80">
              <tr>
                <th className="px-3 py-2 font-medium">id</th>
                <th className="px-3 py-2 font-medium">x</th>
                <th className="px-3 py-2 font-medium">y</th>
                <th className="px-3 py-2 font-medium">z</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200 dark:divide-slate-700">
              {model.lights.map((L) => (
                <tr key={L.id} className="bg-white dark:bg-slate-900/30">
                  <td className="px-3 py-2 font-mono tabular-nums">{L.id}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.x}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.y}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.z}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="flex flex-col gap-2 sm:flex-row">
        <Button
          type="button"
          className="w-full bg-slate-600 dark:bg-slate-600 sm:w-auto"
          onClick={() => router.push("/models")}
        >
          Back to list
        </Button>
      </div>
    </div>
  );
}
