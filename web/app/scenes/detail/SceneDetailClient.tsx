"use client";

import {
  faArrowsRotate,
  faCheck,
  faCirclePlay,
  faCircleStop,
  faPenToSquare,
  faPlus,
  faRightFromBracket,
  faRotate,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import dynamic from "next/dynamic";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
import { useSceneLightsSSE } from "@/lib/useSceneLightsSSE";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_SHAPE_ANIMATION,
  SceneRoutineConflictError,
  fetchRoutines,
  fetchSceneRoutineRuns,
  startSceneRoutine,
  stopSceneRoutineRun,
  type RoutineDefinition,
  type RoutineRun,
} from "@/lib/routines";

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
  const [routines, setRoutines] = useState<RoutineDefinition[] | null>(null);
  const [routineRuns, setRoutineRuns] = useState<RoutineRun[]>([]);
  const [selectedRoutineId, setSelectedRoutineId] = useState("");

  const load = useCallback(async () => {
    if (!id) {
      return;
    }
    setError(null);
    try {
      const [s, mRes, rList, runs] = await Promise.all([
        fetchScene(id),
        fetch("/api/v1/models", { cache: "no-store" }),
        fetchRoutines().catch(() => [] as RoutineDefinition[]),
        fetchSceneRoutineRuns(id).catch(() => [] as RoutineRun[]),
      ]);
      setScene(s);
      setRoutineRuns(runs);
      if (mRes.ok) {
        setModels((await mRes.json()) as ModelSummary[]);
      } else {
        setModels([]);
      }
      setRoutines(rList);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load scene");
      setScene(null);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  // REQ-041 / REQ-029: EventSource subscribes to GET /api/v1/scenes/{id}/lights/events via useSceneLightsSSE.
  const sceneSseLiveRef = useRef(false);
  useSceneLightsSSE(id ?? undefined, setScene, load, {
    sseLiveRef: sceneSseLiveRef,
  });

  useEffect(() => {
    if (!id || routineRuns.length === 0) {
      return;
    }
    const t = window.setInterval(() => {
      void (async () => {
        try {
          const runs = await fetchSceneRoutineRuns(id);
          setRoutineRuns(runs);
          if (!sceneSseLiveRef.current) {
            const s = await fetchScene(id);
            setScene(s);
          }
        } catch {
          /* ignore poll errors */
        }
      })();
    }, 1500);
    return () => window.clearInterval(t);
  }, [id, routineRuns.length]);

  const firstRun = routineRuns[0];
  const isPythonRun =
    firstRun?.routine_type === ROUTINE_TYPE_PYTHON_SCENE_SCRIPT;
  const isShapeRun =
    firstRun?.routine_type === ROUTINE_TYPE_SHAPE_ANIMATION;
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

  async function onStartRoutine() {
    if (!id || !selectedRoutineId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await startSceneRoutine(id, selectedRoutineId);
      await load();
    } catch (e) {
      if (e instanceof SceneRoutineConflictError) {
        await load();
        setError(
          e.requestedRoutineId !== e.existingRoutineId
            ? "Another routine is still running on this scene. Stop it before starting a different one."
            : null,
        );
      } else {
        setError(e instanceof Error ? e.message : "Start routine failed");
      }
    } finally {
      setBusy(false);
    }
  }

  async function onStopRoutine() {
    if (!id || routineRuns.length === 0) {
      return;
    }
    const run = routineRuns[0];
    setBusy(true);
    setError(null);
    try {
      await stopSceneRoutineRun(id, run.id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Stop routine failed");
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

      <section className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          Routines
        </h2>
        <p className="text-xs text-slate-600 dark:text-slate-400">
          Start a saved routine on this scene. The server runs Python (
          <code className="font-mono">python3</code>) or shape animation logic and
          updates lights via loopback HTTP to this API—those{" "}
          <code className="font-mono">PATCH</code> calls do not appear in the
          browser Network tab. Use the EventStream view on{" "}
          <code className="font-mono">lights/events</code> (or server logs) to
          confirm updates. Default samples include a ~1&nbsp;s random colour
          cycle and two geometry demos.
        </p>
        {routineRuns.length > 0 ? (
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-slate-800 dark:text-slate-200">
                Running:{" "}
                <span className="font-medium">
                  {routineRuns[0].routine_name}
                </span>
                {isPythonRun ? (
                  <span className="ml-2 text-xs text-slate-500">(script)</span>
                ) : null}
                {isShapeRun ? (
                  <span className="ml-2 text-xs text-slate-500">(shapes)</span>
                ) : null}
              </p>
              <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                {isPythonRun ? (
                  <Button
                    type="button"
                    icon={faPenToSquare}
                    className="min-h-11 w-full sm:w-auto"
                    disabled={busy}
                    onClick={() =>
                      router.push(
                        `/routines/python?id=${encodeURIComponent(routineRuns[0].routine_id)}&scene=${encodeURIComponent(id)}`,
                      )
                    }
                  >
                    Edit routine
                  </Button>
                ) : null}
                {isShapeRun ? (
                  <Button
                    type="button"
                    icon={faPenToSquare}
                    className="min-h-11 w-full sm:w-auto"
                    disabled={busy}
                    onClick={() =>
                      router.push(
                        `/routines/shape?id=${encodeURIComponent(routineRuns[0].routine_id)}&scene=${encodeURIComponent(id)}`,
                      )
                    }
                  >
                    Edit routine
                  </Button>
                ) : null}
                <Button
                  type="button"
                  icon={faCircleStop}
                  className="min-h-11 w-full bg-amber-800 hover:bg-amber-700 sm:w-auto"
                  disabled={busy}
                  onClick={() => void onStopRoutine()}
                >
                  Stop routine
                </Button>
              </div>
            </div>
            {(isPythonRun || isShapeRun) && firstRun ? (
              <p className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs text-slate-600 dark:border-slate-600 dark:bg-slate-950/50 dark:text-slate-400">
                This routine runs on the server. You can close this tab and lights keep updating;
                use Stop above or the API to end the run. Live changes appear in the 3D view via
                server push.
              </p>
            ) : null}
          </div>
        ) : (
          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
            <label className="flex min-w-[12rem] flex-col gap-1 text-xs">
              <span className="text-slate-600 dark:text-slate-400">Routine</span>
              <select
                className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
                value={selectedRoutineId}
                onChange={(e) => setSelectedRoutineId(e.target.value)}
                disabled={busy || !routines || routines.length === 0}
              >
                <option value="">
                  {!routines
                    ? "Loading…"
                    : routines.length === 0
                      ? "Create routines under Routines first"
                      : "Select a routine…"}
                </option>
                {(routines ?? []).map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name}
                  </option>
                ))}
              </select>
            </label>
            <Button
              type="button"
              icon={faCirclePlay}
              className="min-h-11 w-full sm:w-auto"
              disabled={busy || !selectedRoutineId}
              onClick={() => void onStartRoutine()}
            >
              Start on this scene
            </Button>
          </div>
        )}
      </section>

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
            icon={faArrowsRotate}
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
                  icon={faRightFromBracket}
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
            icon={faPlus}
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
          icon={faRotate}
          className="min-h-11 bg-slate-600 dark:bg-slate-600"
          disabled={busy}
          onClick={() => void load()}
        >
          Refresh
        </Button>
        <Button
          type="button"
          icon={faTrash}
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
        icon={faCheck}
        className="h-9 px-3 text-xs"
        disabled={disabled}
        onClick={() => onApply(ox, oy, oz)}
      >
        Apply placement
      </Button>
    </div>
  );
}
