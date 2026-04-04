"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { fetchScenes, type SceneSummary } from "@/lib/scenes";

export default function ScenesListPage() {
  const router = useRouter();
  const [scenes, setScenes] = useState<SceneSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const data = await fetchScenes();
      setScenes(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load scenes");
      setScenes([]);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <nav className="text-sm">
        <Link
          href="/"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Home
        </Link>
        <span className="text-slate-400"> · </span>
        <Link
          href="/models"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Models
        </Link>
      </nav>

      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Scenes
          </h1>
          <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
            Composite 3D layouts: multiple models with integer placement in
            scene coordinates (meters).
          </p>
        </div>
        <Button
          type="button"
          className="w-full sm:w-auto"
          onClick={() => router.push("/scenes/new")}
        >
          New scene
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

      {scenes === null ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : scenes.length === 0 ? (
        <p className="text-sm text-slate-600 dark:text-slate-400">
          No scenes yet.{" "}
          <Link
            href="/scenes/new"
            className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
          >
            Create one
          </Link>{" "}
          with at least one model.
        </p>
      ) : (
        <ul className="divide-y divide-slate-200 rounded-xl border border-slate-200 bg-white dark:divide-slate-700 dark:border-slate-700 dark:bg-slate-900/40">
          {scenes.map((s) => (
            <li
              key={s.id}
              className="flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0">
                <Link
                  href={`/scenes/detail?id=${encodeURIComponent(s.id)}`}
                  className="font-medium text-sky-800 hover:underline dark:text-sky-300"
                >
                  {s.name}
                </Link>
                <p className="truncate text-xs text-slate-500 dark:text-slate-400">
                  {s.model_count} model{s.model_count === 1 ? "" : "s"} ·{" "}
                  {new Date(s.created_at).toLocaleString()}
                </p>
              </div>
              <Button
                type="button"
                className="min-h-11 w-full sm:w-auto"
                onClick={() =>
                  router.push(`/scenes/detail?id=${encodeURIComponent(s.id)}`)
                }
              >
                Open
              </Button>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
