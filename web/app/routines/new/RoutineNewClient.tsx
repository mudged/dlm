"use client";

import { faArrowLeft, faPlus } from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { PYTHON_ROUTINE_DEFAULT_SOURCE } from "@/lib/pythonSceneApiCatalog";
import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "@/lib/shapeAnimationDefault";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_SHAPE_ANIMATION,
  createRoutine,
} from "@/lib/routines";

export default function RoutineNewClient() {
  const router = useRouter();
  const sp = useSearchParams();
  const kindParam = sp.get("kind");
  const initialKind: "python" | "shape" =
    kindParam === "shape" ? "shape" : "python";

  const [kind, setKind] = useState<"python" | "shape">(initialKind);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const n = name.trim();
      if (!n) {
        throw new Error("Name is required.");
      }
      if (kind === "python") {
        const created = await createRoutine({
          name: n,
          description,
          type: ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
          python_source: PYTHON_ROUTINE_DEFAULT_SOURCE,
        });
        router.replace(`/routines/python?id=${encodeURIComponent(created.id)}`);
      } else {
        const created = await createRoutine({
          name: n,
          description,
          type: ROUTINE_TYPE_SHAPE_ANIMATION,
          definition_json: SHAPE_ANIMATION_DEFAULT_DEFINITION,
        });
        router.replace(`/routines/shape?id=${encodeURIComponent(created.id)}`);
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
          Choose Python (script in Pyodide) or shape animation (JSON definition,
          runs in the browser). Name your routine, then edit it on the next page.
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
        <fieldset className="space-y-2">
          <legend className="text-xs font-medium text-slate-700 dark:text-slate-300">
            Kind
          </legend>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="radio"
              name="kind"
              checked={kind === "python"}
              onChange={() => setKind("python")}
              className="h-4 w-4"
            />
            Python scene script
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="radio"
              name="kind"
              checked={kind === "shape"}
              onChange={() => setKind("shape")}
              className="h-4 w-4"
            />
            Shape animation
          </label>
        </fieldset>
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
