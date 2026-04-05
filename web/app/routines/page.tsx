"use client";

import { faPlus, faTrash } from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import {
  ROUTINE_TYPE_RANDOM_COLOUR_ALL,
  createRoutine,
  deleteRoutine,
  fetchRoutines,
  type RoutineDefinition,
} from "@/lib/routines";

export default function RoutinesPage() {
  const [list, setList] = useState<RoutineDefinition[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const load = useCallback(async () => {
    setError(null);
    try {
      const data = await fetchRoutines();
      setList(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load routines");
      setList([]);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await createRoutine({
        name: name.trim(),
        description,
        type: ROUTINE_TYPE_RANDOM_COLOUR_ALL,
      });
      setName("");
      setDescription("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  async function onDelete(id: string) {
    setBusy(true);
    setError(null);
    try {
      await deleteRoutine(id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Delete failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <header>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Routines
        </h1>
        <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
          Save named routines, then start them from a{" "}
          <Link href="/scenes" className="text-sky-700 underline dark:text-sky-300">
            scene
          </Link>{" "}
          detail page.
        </p>
      </header>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400" role="alert">
          {error}
        </p>
      ) : null}

      <form
        onSubmit={onCreate}
        className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
      >
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          New routine
        </h2>
        <label className="flex flex-col gap-1 text-xs">
          <span className="text-slate-600 dark:text-slate-400">Name</span>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs">
          <span className="text-slate-600 dark:text-slate-400">
            Description (optional)
          </span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs">
          <span className="text-slate-600 dark:text-slate-400">Type</span>
          <select
            className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            value={ROUTINE_TYPE_RANDOM_COLOUR_ALL}
            disabled
          >
            <option value={ROUTINE_TYPE_RANDOM_COLOUR_ALL}>
              Random colour cycle (all lights)
            </option>
          </select>
        </label>
        <Button type="submit" icon={faPlus} disabled={busy || !name.trim()}>
          Create routine
        </Button>
      </form>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          Saved routines
        </h2>
        {!list ? (
          <p className="text-sm text-slate-500">Loading…</p>
        ) : list.length === 0 ? (
          <p className="text-sm text-slate-500">No routines yet.</p>
        ) : (
          <ul className="divide-y divide-slate-200 rounded-xl border border-slate-200 dark:divide-slate-700 dark:border-slate-700">
            {list.map((r) => (
              <li
                key={r.id}
                className="flex flex-col gap-2 p-4 sm:flex-row sm:items-center sm:justify-between"
              >
                <div>
                  <p className="font-medium text-slate-900 dark:text-slate-100">
                    {r.name}
                  </p>
                  {r.description ? (
                    <p className="text-xs text-slate-600 dark:text-slate-400">
                      {r.description}
                    </p>
                  ) : null}
                  <p className="text-xs text-slate-500">{r.type}</p>
                </div>
                <Button
                  type="button"
                  icon={faTrash}
                  className="min-h-11 bg-red-900 hover:bg-red-800 dark:bg-red-950"
                  disabled={busy}
                  onClick={() => void onDelete(r.id)}
                >
                  Delete
                </Button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
