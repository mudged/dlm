"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import type { ModelSummary } from "@/lib/models";
import { createScene } from "@/lib/scenes";

type Row = { key: string; modelId: string };

function newRow(): Row {
  return { key: crypto.randomUUID(), modelId: "" };
}

export function NewSceneClient() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [models, setModels] = useState<ModelSummary[] | null>(null);
  const [rows, setRows] = useState<Row[]>(() => [newRow()]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const loadModels = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/models", { cache: "no-store" });
      if (!res.ok) {
        setModels([]);
        return;
      }
      setModels((await res.json()) as ModelSummary[]);
    } catch {
      setModels([]);
    }
  }, []);

  useEffect(() => {
    void loadModels();
  }, [loadModels]);

  function addRow() {
    setRows((r) => [...r, newRow()]);
  }

  function moveRow(from: number, to: number) {
    if (to < 0 || to >= rows.length) return;
    setRows((rs) => {
      const next = [...rs];
      const [x] = next.splice(from, 1);
      next.splice(to, 0, x);
      return next;
    });
  }

  async function submit() {
    setError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Scene name is required.");
      return;
    }
    const ordered = rows
      .map((r) => r.modelId.trim())
      .filter((id) => id !== "");
    if (ordered.length === 0) {
      setError("Add at least one model with a selected id.");
      return;
    }
    const seen = new Set<string>();
    for (const id of ordered) {
      if (seen.has(id)) {
        setError("Each model can only appear once.");
        return;
      }
      seen.add(id);
    }
    setBusy(true);
    try {
      const sum = await createScene({
        name: trimmed,
        models: ordered.map((model_id) => ({ model_id })),
      });
      router.push(`/scenes/detail?id=${encodeURIComponent(sum.id)}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          New scene
        </h1>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
          Choose a name and one or more models in order. The server places each
          model in scene space automatically (non-negative coordinates, spacing
          to the +X side after the first model).
        </p>
      </header>

      {error ? (
        <p
          className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
          role="alert"
        >
          {error}
        </p>
      ) : null}

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-slate-700 dark:text-slate-300">
          Scene name
        </span>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 dark:border-slate-600 dark:bg-slate-900"
          placeholder="e.g. Living room"
        />
      </label>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          Models (order = placement order)
        </h2>
        {rows.map((row, i) => (
          <div
            key={row.key}
            className="flex flex-col gap-2 rounded-lg border border-slate-200 p-3 dark:border-slate-700 sm:flex-row sm:flex-wrap sm:items-end"
          >
            <label className="flex min-w-[10rem] flex-1 flex-col gap-1 text-xs">
              <span className="text-slate-600 dark:text-slate-400">Model</span>
              <select
                className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
                value={row.modelId}
                onChange={(e) => {
                  const v = e.target.value;
                  setRows((rs) =>
                    rs.map((r, j) => (j === i ? { ...r, modelId: v } : r)),
                  );
                }}
                disabled={!models?.length}
              >
                <option value="">
                  {models === null ? "Loading…" : "Select…"}
                </option>
                {(models ?? []).map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name}
                  </option>
                ))}
              </select>
            </label>
            <div className="flex flex-wrap gap-1">
              <button
                type="button"
                className="inline-flex min-h-9 min-w-11 items-center justify-center rounded-lg border border-slate-300 bg-white px-2 text-xs font-medium text-slate-800 hover:bg-slate-50 disabled:opacity-40 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
                disabled={i === 0}
                onClick={() => moveRow(i, i - 1)}
              >
                Up
              </button>
              <button
                type="button"
                className="inline-flex min-h-9 min-w-11 items-center justify-center rounded-lg border border-slate-300 bg-white px-2 text-xs font-medium text-slate-800 hover:bg-slate-50 disabled:opacity-40 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
                disabled={i === rows.length - 1}
                onClick={() => moveRow(i, i + 1)}
              >
                Down
              </button>
            </div>
          </div>
        ))}
        <Button
          type="button"
          className="min-h-11 w-full bg-slate-600 hover:bg-slate-500 dark:bg-slate-600 dark:hover:bg-slate-500 sm:w-auto"
          onClick={addRow}
        >
          Add another model row
        </Button>
      </section>

      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          className="min-h-11"
          disabled={busy}
          onClick={() => void submit()}
        >
          {busy ? "Creating…" : "Create scene"}
        </Button>
        <Link
          href="/scenes"
          className="inline-flex min-h-11 items-center rounded-lg border border-slate-300 px-4 text-sm font-medium text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800"
        >
          Cancel
        </Link>
      </div>
    </div>
  );
}
