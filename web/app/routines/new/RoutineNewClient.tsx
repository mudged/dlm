"use client";

import { faArrowLeft, faPlus } from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/Button";
import { PYTHON_ROUTINE_DEFAULT_SOURCE } from "@/lib/pythonSceneApiCatalog";
import {
  ROUTINE_TYPE_CREATE_OPTIONS,
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_RANDOM_COLOUR_ALL,
  createRoutine,
  isCreatableRoutineType,
} from "@/lib/routines";

function initialTypeFromQuery(raw: string | null): string {
  if (raw && isCreatableRoutineType(raw)) {
    return raw;
  }
  return ROUTINE_TYPE_RANDOM_COLOUR_ALL;
}

export default function RoutineNewClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const typeFromUrl = useMemo(
    () => initialTypeFromQuery(searchParams.get("type")),
    [searchParams],
  );

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [routineType, setRoutineType] = useState(typeFromUrl);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setRoutineType(typeFromUrl);
  }, [typeFromUrl]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const n = name.trim();
      if (!n) {
        throw new Error("Name is required.");
      }
      if (!isCreatableRoutineType(routineType)) {
        throw new Error("Select a valid routine type.");
      }
      const body =
        routineType === ROUTINE_TYPE_PYTHON_SCENE_SCRIPT
          ? {
              name: n,
              description,
              type: routineType,
              python_source: PYTHON_ROUTINE_DEFAULT_SOURCE,
            }
          : { name: n, description, type: routineType };
      const created = await createRoutine(body);
      if (routineType === ROUTINE_TYPE_PYTHON_SCENE_SCRIPT) {
        router.replace(
          `/routines/python?id=${encodeURIComponent(created.id)}`,
        );
      } else {
        router.push("/routines");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <header>
        <Link
          href="/routines"
          className="inline-flex items-center gap-2 text-sm text-sky-700 underline dark:text-sky-300"
        >
          <span aria-hidden>←</span> All routines
        </Link>
        <h1 className="mt-4 text-2xl font-bold tracking-tight sm:text-3xl">
          New routine
        </h1>
        <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
          Choose the routine type, then name it. Python routines open in the
          in-browser editor next. Colour-cycle routines are ready to run from a
          scene detail page.
        </p>
      </header>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400" role="alert">
          {error}
        </p>
      ) : null}

      <form
        onSubmit={(e) => void onSubmit(e)}
        className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
      >
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
          <span className="text-slate-600 dark:text-slate-400" id="routine-type-label">
            Routine type
          </span>
          <select
            id="routine-type"
            aria-labelledby="routine-type-label"
            className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            value={routineType}
            onChange={(e) => setRoutineType(e.target.value)}
          >
            {ROUTINE_TYPE_CREATE_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
        <div className="flex flex-wrap gap-2 pt-1">
          <Button type="submit" icon={faPlus} disabled={busy || !name.trim()}>
            Create routine
          </Button>
          <Button
            type="button"
            icon={faArrowLeft}
            className="bg-slate-200 text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 dark:hover:bg-slate-600"
            disabled={busy}
            onClick={() => router.push("/routines")}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
