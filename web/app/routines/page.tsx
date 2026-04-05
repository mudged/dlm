"use client";

import {
  faPenToSquare,
  faPlus,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  deleteRoutine,
  fetchRoutines,
  type RoutineDefinition,
} from "@/lib/routines";

export default function RoutinesPage() {
  const router = useRouter();
  const [list, setList] = useState<RoutineDefinition[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

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
          detail page. Create a routine and pick its type—including Python—from
          one form. Python routines use the in-browser editor and Pyodide runner
          (see help text on the editor page).
        </p>
        <p className="mt-4">
          <Button
            type="button"
            icon={faPlus}
            className="min-h-11"
            onClick={() => router.push("/routines/new")}
          >
            New routine
          </Button>
        </p>
      </header>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400" role="alert">
          {error}
        </p>
      ) : null}

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
                <div className="flex flex-wrap gap-2">
                  {r.type === ROUTINE_TYPE_PYTHON_SCENE_SCRIPT ? (
                    <Button
                      type="button"
                      icon={faPenToSquare}
                      className="min-h-11"
                      disabled={busy}
                      onClick={() =>
                        router.push(
                          `/routines/python?id=${encodeURIComponent(r.id)}`,
                        )
                      }
                    >
                      Edit Python
                    </Button>
                  ) : null}
                  <Button
                    type="button"
                    icon={faTrash}
                    className="min-h-11 bg-red-900 hover:bg-red-800 dark:bg-red-950"
                    disabled={busy}
                    onClick={() => void onDelete(r.id)}
                  >
                    Delete
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
