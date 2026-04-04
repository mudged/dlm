"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/Button";
import type { ModelSummary } from "@/lib/models";
import {
  addSceneModel,
  deleteScene,
  fetchScene,
  patchSceneModelOffsets,
  removeSceneModel,
  type SceneDetail,
} from "@/lib/scenes";

const SceneLightsCanvas = dynamic(
  () => import("@/components/SceneLightsCanvas"),
  { ssr: false },
);

export function SceneDetailClient() {
  const router = useRouter();
  const search = useSearchParams();
  const id = search.get("id");
  const [scene, setScene] = useState<SceneDetail | null>(null);
  const [models, setModels] = useState<ModelSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [addModelId, setAddModelId] = useState("");
  const [cameraResetVersion, setCameraResetVersion] = useState(0);

  const load = useCallback(async () => {
    if (!id) {
      return;
    }
    setError(null);
    try {
      const [s, mRes] = await Promise.all([
        fetchScene(id),
        fetch("/api/v1/models", { cache: "no-store" }),
      ]);
      setScene(s);
      if (mRes.ok) {
        setModels((await mRes.json()) as ModelSummary[]);
      } else {
        setModels([]);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load scene");
      setScene(null);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  const inSceneIds = useMemo(() => {
    if (!scene) {
      return new Set<string>();
    }
    return new Set(scene.items.map((i) => i.model_id));
  }, [scene]);

  const addableModels = useMemo(() => {
    if (!models) {
      return [];
    }
    return models.filter((m) => !inSceneIds.has(m.id));
  }, [models, inSceneIds]);

  async function onAddModel() {
    if (!id || !addModelId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await addSceneModel(id, { model_id: addModelId });
      setAddModelId("");
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Add failed");
    } finally {
      setBusy(false);
    }
  }

  async function onRemoveModel(modelId: string) {
    if (!id) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const r = await removeSceneModel(id, modelId);
      if (r === "last_model") {
        const ok = window.confirm(
          "Removing the last model deletes the entire scene. This cannot be undone. Continue?",
        );
        if (!ok) {
          setBusy(false);
          return;
        }
        await deleteScene(id);
        router.push("/scenes");
        return;
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Remove failed");
    } finally {
      setBusy(false);
    }
  }

  async function onPatchOffsets(
    modelId: string,
    ox: number,
    oy: number,
    oz: number,
  ) {
    if (!id) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await patchSceneModelOffsets(id, modelId, ox, oy, oz);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Update failed");
    } finally {
      setBusy(false);
    }
  }

  async function onDeleteScene() {
    if (!id) {
      return;
    }
    if (!window.confirm("Delete this entire scene?")) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await deleteScene(id);
      router.push("/scenes");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setBusy(false);
    }
  }

  if (!id) {
    return (
      <p className="text-sm text-amber-800 dark:text-amber-200">
        Missing scene id. Open a scene from the list.
      </p>
    );
  }

  if (!scene && !error) {
    return <p className="text-sm text-slate-500">Loading…</p>;
  }

  if (!scene) {
    return (
      <p className="text-sm text-red-600 dark:text-red-400" role="alert">
        {error}
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {scene.name}
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          {scene.items.length} model{scene.items.length === 1 ? "" : "s"} ·{" "}
          {new Date(scene.created_at).toLocaleString()}
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

      <section className="space-y-2" aria-labelledby="scene-3d-heading">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <h2
            id="scene-3d-heading"
            className="text-sm font-semibold text-slate-800 dark:text-slate-200"
          >
            3D view
          </h2>
          <Button
            type="button"
            className="h-9 w-full px-3 py-0 text-sm sm:w-auto"
            onClick={() => setCameraResetVersion((v) => v + 1)}
          >
            Reset camera
          </Button>
        </div>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Drag to rotate; scroll or pinch to zoom. Tap a light to pin details.
        </p>
        <SceneLightsCanvas
          items={scene.items}
          cameraPersistenceKey={scene.id}
          cameraResetVersion={cameraResetVersion}
        />
      </section>

      <section className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          Models in scene
        </h2>
        <ul className="space-y-4">
          {scene.items.map((it) => (
            <li
              key={it.model_id}
              className="flex flex-col gap-3 border-b border-slate-100 pb-4 last:border-0 dark:border-slate-800"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <Link
                  href={`/models/detail?id=${encodeURIComponent(it.model_id)}`}
                  className="font-medium text-sky-800 hover:underline dark:text-sky-300"
                >
                  {it.name}
                </Link>
                <Button
                  type="button"
                  className="min-h-11 bg-red-800 hover:bg-red-700 dark:bg-red-900 dark:hover:bg-red-800"
                  disabled={busy}
                  onClick={() => void onRemoveModel(it.model_id)}
                >
                  Remove from scene
                </Button>
              </div>
              <PlacementEditor
                initial={{
                  ox: it.offset_x,
                  oy: it.offset_y,
                  oz: it.offset_z,
                }}
                disabled={busy}
                onApply={(ox, oy, oz) =>
                  void onPatchOffsets(it.model_id, ox, oy, oz)
                }
              />
            </li>
          ))}
        </ul>
      </section>

      <section className="space-y-2 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          Add model
        </h2>
        <p className="text-xs text-slate-600 dark:text-slate-400">
          New models are placed to the right (+X) of the current layout by default
          when you do not specify offsets.
        </p>
        <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
          <label className="flex min-w-[12rem] flex-col gap-1 text-xs">
            <span className="text-slate-600 dark:text-slate-400">Model</span>
            <select
              className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
              value={addModelId}
              onChange={(e) => setAddModelId(e.target.value)}
              disabled={busy || addableModels.length === 0}
            >
              <option value="">
                {addableModels.length === 0
                  ? "All models are already in this scene"
                  : "Select a model…"}
              </option>
              {addableModels.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name}
                </option>
              ))}
            </select>
          </label>
          <Button
            type="button"
            className="min-h-11 w-full sm:w-auto"
            disabled={busy || !addModelId}
            onClick={() => void onAddModel()}
          >
            Add with default placement
          </Button>
        </div>
      </section>

      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          className="min-h-11 bg-slate-600 dark:bg-slate-600"
          disabled={busy}
          onClick={() => void load()}
        >
          Refresh
        </Button>
        <Button
          type="button"
          className="min-h-11 bg-red-900 hover:bg-red-800 dark:bg-red-950 dark:hover:bg-red-900"
          disabled={busy}
          onClick={() => void onDeleteScene()}
        >
          Delete entire scene
        </Button>
      </div>
    </div>
  );
}

function PlacementEditor({
  initial,
  disabled,
  onApply,
}: {
  initial: { ox: number; oy: number; oz: number };
  disabled: boolean;
  onApply: (ox: number, oy: number, oz: number) => void;
}) {
  const [ox, setOx] = useState(initial.ox);
  const [oy, setOy] = useState(initial.oy);
  const [oz, setOz] = useState(initial.oz);

  useEffect(() => {
    setOx(initial.ox);
    setOy(initial.oy);
    setOz(initial.oz);
  }, [initial.ox, initial.oy, initial.oz]);

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
      <span className="text-xs text-slate-600 dark:text-slate-400">
        Offsets (m, integers ≥ 0)
      </span>
      <label className="flex items-center gap-1 text-xs">
        ox
        <input
          type="number"
          min={0}
          step={1}
          value={ox}
          onChange={(e) => setOx(Number(e.target.value))}
          className="w-20 rounded border border-slate-300 bg-white px-1 py-1 font-mono dark:border-slate-600 dark:bg-slate-900"
        />
      </label>
      <label className="flex items-center gap-1 text-xs">
        oy
        <input
          type="number"
          min={0}
          step={1}
          value={oy}
          onChange={(e) => setOy(Number(e.target.value))}
          className="w-20 rounded border border-slate-300 bg-white px-1 py-1 font-mono dark:border-slate-600 dark:bg-slate-900"
        />
      </label>
      <label className="flex items-center gap-1 text-xs">
        oz
        <input
          type="number"
          min={0}
          step={1}
          value={oz}
          onChange={(e) => setOz(Number(e.target.value))}
          className="w-20 rounded border border-slate-300 bg-white px-1 py-1 font-mono dark:border-slate-600 dark:bg-slate-900"
        />
      </label>
      <Button
        type="button"
        className="h-9 px-3 text-xs"
        disabled={disabled}
        onClick={() => onApply(ox, oy, oz)}
      >
        Apply placement
      </Button>
    </div>
  );
}
