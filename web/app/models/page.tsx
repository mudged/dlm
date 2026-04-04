"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import type { ModelSummary } from "@/lib/models";

export default function ModelsListPage() {
  const router = useRouter();
  const [models, setModels] = useState<ModelSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const res = await fetch("/api/v1/models", { cache: "no-store" });
      if (!res.ok) {
        const j = (await res.json().catch(() => null)) as {
          error?: { message?: string };
        };
        setError(j?.error?.message ?? `Request failed (${res.status})`);
        setModels([]);
        return;
      }
      const data = (await res.json()) as ModelSummary[];
      setModels(data);
    } catch {
      setError("Could not reach the API.");
      setModels([]);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function remove(id: string) {
    if (!window.confirm("Delete this model?")) return;
    setBusyId(id);
    setError(null);
    try {
      const res = await fetch(`/api/v1/models/${encodeURIComponent(id)}`, {
        method: "DELETE",
      });
      if (res.status !== 204) {
        const j = (await res.json().catch(() => null)) as {
          error?: {
            message?: string;
            code?: string;
            details?: { scenes?: { id: string; name: string }[] };
          };
        };
        let msg = j?.error?.message ?? `Delete failed (${res.status})`;
        if (
          res.status === 409 &&
          j?.error?.code === "model_in_scenes" &&
          j.error.details?.scenes?.length
        ) {
          const names = j.error.details.scenes
            .map((s) => s.name || s.id)
            .join(", ");
          msg = `${msg} Scenes: ${names}. Remove the model from those scenes first.`;
        }
        setError(msg);
        setBusyId(null);
        return;
      }
      await load();
    } catch {
      setError("Could not reach the API.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <nav className="text-sm">
        <Link
          href="/"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Home
        </Link>
      </nav>

      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Light models
          </h1>
          <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
            Up to 1000 lights per model. Upload a CSV with columns{" "}
            <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
              id,x,y,z
            </code>
            .
          </p>
        </div>
        <Button
          type="button"
          className="w-full sm:w-auto"
          onClick={() => router.push("/models/new")}
        >
          Upload CSV
        </Button>
      </header>

      {error ? (
        <p
          className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
          role="alert"
        >
          {error}
        </p>
      ) : null}

      {models === null ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : models.length === 0 ? (
        <p className="text-sm text-slate-600 dark:text-slate-400">
          No models yet.{" "}
          <Link
            href="/models/new"
            className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
          >
            Upload a CSV
          </Link>{" "}
          to add one.
        </p>
      ) : (
        <ul className="divide-y divide-slate-200 rounded-xl border border-slate-200 bg-white dark:divide-slate-700 dark:border-slate-700 dark:bg-slate-900/40">
          {models.map((m) => (
            <li
              key={m.id}
              className="flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0">
                <Link
                  href={`/models/detail?id=${encodeURIComponent(m.id)}`}
                  className="font-medium text-sky-800 hover:underline dark:text-sky-300"
                >
                  {m.name}
                </Link>
                <p className="truncate text-xs text-slate-500 dark:text-slate-400">
                  {m.light_count} lights ·{" "}
                  {new Date(m.created_at).toLocaleString()}
                </p>
              </div>
              <div className="flex shrink-0 gap-2">
                <Button
                  type="button"
                  className="min-h-11 bg-slate-600 dark:bg-slate-600"
                  onClick={() =>
                    router.push(
                      `/models/detail?id=${encodeURIComponent(m.id)}`,
                    )
                  }
                >
                  View
                </Button>
                <Button
                  type="button"
                  className="min-h-11 bg-red-800 hover:bg-red-700 dark:bg-red-900 dark:hover:bg-red-800"
                  disabled={busyId === m.id}
                  onClick={() => void remove(m.id)}
                >
                  Delete
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
