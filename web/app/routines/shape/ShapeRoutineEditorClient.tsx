"use client";

import {
  faArrowsRotate,
  faCirclePlay,
  faCircleStop,
  faCopy,
  faFloppyDisk,
  faLightbulb,
  faRotate,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import dynamic from "next/dynamic";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ShapeAnimationRoutineHost } from "@/components/ShapeAnimationRoutineHost";
import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "@/lib/shapeAnimationDefault";
import {
  ROUTINE_TYPE_SHAPE_ANIMATION,
  createRoutine,
  deleteRoutine,
  fetchRoutine,
  patchRoutine,
  startSceneRoutine,
  stopSceneRoutineRun,
  type RoutineDefinition,
} from "@/lib/routines";
import {
  fetchScene,
  fetchScenes,
  patchSceneLightsStateScene,
  type SceneDetail,
  type SceneSummary,
} from "@/lib/scenes";

const SceneLightsCanvas = dynamic(() => import("@/components/SceneLightsCanvas"), {
  ssr: false,
});

export default function ShapeRoutineEditorClient() {
  const router = useRouter();
  const search = useSearchParams();
  const idParam = search.get("id");
  const sceneFromQuery = search.get("scene");

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [jsonText, setJsonText] = useState(
    JSON.stringify(SHAPE_ANIMATION_DEFAULT_DEFINITION, null, 2),
  );
  const [routineId, setRoutineId] = useState<string | null>(idParam);
  const [loaded, setLoaded] = useState(() => Boolean(idParam));
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const [scenesList, setScenesList] = useState<SceneSummary[] | null>(null);
  const [targetSceneId, setTargetSceneId] = useState("");
  const [targetScene, setTargetScene] = useState<SceneDetail | null>(null);
  const [activeRun, setActiveRun] = useState<{
    run_id: string;
    scene_id: string;
  } | null>(null);
  const [cameraResetVersion, setCameraResetVersion] = useState(0);

  const load = useCallback(async (rid: string) => {
    setError(null);
    setLoaded(false);
    try {
      const r: RoutineDefinition = await fetchRoutine(rid);
      if (r.type !== ROUTINE_TYPE_SHAPE_ANIMATION) {
        throw new Error("This routine is not a shape animation.");
      }
      setRoutineId(r.id);
      setName(r.name);
      setDescription(r.description);
      if (r.definition_json != null) {
        setJsonText(JSON.stringify(r.definition_json, null, 2));
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Load failed");
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (!idParam) {
      router.replace("/routines/new?kind=shape");
      return;
    }
    void load(idParam);
  }, [idParam, load, router]);

  useEffect(() => {
    void (async () => {
      try {
        const list = await fetchScenes();
        setScenesList(list);
        setTargetSceneId((prev) => {
          if (sceneFromQuery && list.some((s) => s.id === sceneFromQuery)) {
            return sceneFromQuery;
          }
          if (prev && list.some((s) => s.id === prev)) {
            return prev;
          }
          return list.length > 0 ? list[0]!.id : "";
        });
      } catch {
        setScenesList([]);
      }
    })();
  }, [sceneFromQuery]);

  useEffect(() => {
    if (!targetSceneId) {
      setTargetScene(null);
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const s = await fetchScene(targetSceneId);
        if (!cancelled) {
          setTargetScene(s);
        }
      } catch {
        if (!cancelled) {
          setTargetScene(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [targetSceneId]);

  const refreshTargetScene = useCallback(async () => {
    if (!targetSceneId) {
      return;
    }
    try {
      const s = await fetchScene(targetSceneId);
      setTargetScene(s);
    } catch {
      /* ignore */
    }
  }, [targetSceneId]);

  function parseDefinitionOrThrow(): unknown {
    return JSON.parse(jsonText) as unknown;
  }

  async function onSave() {
    setBusy(true);
    setError(null);
    try {
      const n = name.trim();
      if (!n) {
        throw new Error("Please enter a name.");
      }
      const parsed = parseDefinitionOrThrow();
      if (!routineId) {
        throw new Error("Routine id missing.");
      }
      await patchRoutine(routineId, {
        name: n,
        description,
        definition_json: parsed,
      });
    } catch (e) {
      setError(
        e instanceof SyntaxError
          ? "Invalid JSON in definition."
          : e instanceof Error
            ? e.message
            : "Save failed",
      );
    } finally {
      setBusy(false);
    }
  }

  async function onDuplicate() {
    if (!routineId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const parsed = parseDefinitionOrThrow();
      const dup = await createRoutine({
        name: `${name.trim() || "routine"} (copy)`,
        description,
        type: ROUTINE_TYPE_SHAPE_ANIMATION,
        definition_json: parsed,
      });
      router.push(`/routines/shape?id=${encodeURIComponent(dup.id)}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Duplicate failed");
    } finally {
      setBusy(false);
    }
  }

  async function onDelete() {
    if (!routineId) {
      return;
    }
    if (!window.confirm("Delete this routine for good? You cannot undo this.")) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await deleteRoutine(routineId);
      router.push("/routines");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setBusy(false);
    }
  }

  async function onStartRun() {
    if (!routineId || !targetSceneId) {
      setError("Pick a room first.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await startSceneRoutine(targetSceneId, routineId);
      setActiveRun({ run_id: res.run_id, scene_id: res.scene_id });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Start failed");
    } finally {
      setBusy(false);
    }
  }

  async function onStopRun() {
    if (!activeRun) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await stopSceneRoutineRun(activeRun.scene_id, activeRun.run_id);
      setActiveRun(null);
      await refreshTargetScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Stop failed");
    } finally {
      setBusy(false);
    }
  }

  async function onResetSceneLights() {
    if (!targetSceneId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await patchSceneLightsStateScene(targetSceneId, {
        on: false,
        color: "#ffffff",
        brightness_pct: 100,
      });
      await refreshTargetScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Reset room lights failed");
    } finally {
      setBusy(false);
    }
  }

  const showShapeHost = activeRun && activeRun.scene_id === targetSceneId;

  let definitionJsonForHost = "";
  try {
    definitionJsonForHost = JSON.stringify(parseDefinitionOrThrow());
  } catch {
    definitionJsonForHost = "";
  }

  if (!loaded) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8">
        <p className="text-sm text-slate-500">Loading…</p>
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-4 px-4 py-6 sm:px-6 lg:px-8">
      <header className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Shape animation routine
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            Edit the JSON definition (version, background, shapes). Run from here
            or from a scene page. Animation runs in your browser and updates lights
            via the scene API.
          </p>
        </div>
        <Link
          href="/routines"
          className="text-sm text-sky-700 underline dark:text-sky-300"
        >
          ← All routines
        </Link>
      </header>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400" role="alert">
          {error}
        </p>
      ) : null}

      <div className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <label className="flex flex-col gap-1 text-xs">
          <span className="text-slate-600 dark:text-slate-400">Name</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs">
          <span className="text-slate-600 dark:text-slate-400">Description (optional)</span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </label>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            icon={faFloppyDisk}
            disabled={busy}
            onClick={() => void onSave()}
          >
            Save
          </Button>
          <Button
            type="button"
            icon={faCopy}
            disabled={busy || !routineId}
            onClick={() => void onDuplicate()}
          >
            Duplicate
          </Button>
          <Button
            type="button"
            icon={faTrash}
            className="bg-red-900 hover:bg-red-800 dark:bg-red-950"
            disabled={busy}
            onClick={() => void onDelete()}
          >
            Delete
          </Button>
        </div>
      </div>

      <section className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-200">
          Definition JSON
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          Must match the server schema (version 1, background, shapes array). Invalid JSON
          or fields will fail on save.
        </p>
        <textarea
          value={jsonText}
          onChange={(e) => setJsonText(e.target.value)}
          spellCheck={false}
          className="mt-3 h-72 w-full rounded border border-slate-300 bg-slate-50 p-2 font-mono text-xs dark:border-slate-600 dark:bg-slate-950"
        />
        <div className="mt-2">
          <Button
            type="button"
            icon={faRotate}
            className="bg-slate-200 text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 dark:hover:bg-slate-600"
            onClick={() => {
              try {
                const o = JSON.parse(jsonText) as unknown;
                setJsonText(JSON.stringify(o, null, 2));
              } catch {
                setError("Format: fix JSON syntax first.");
              }
            }}
          >
            Format JSON
          </Button>
        </div>
      </section>

      <section className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-200">
          Run and watch the room
        </h2>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
          <label className="flex min-w-[200px] flex-1 flex-col gap-1 text-xs">
            <span className="text-slate-600 dark:text-slate-400">Room</span>
            <select
              value={targetSceneId}
              onChange={(e) => setTargetSceneId(e.target.value)}
              className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            >
              <option value="">Select a room…</option>
              {(scenesList ?? []).map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          {!activeRun ? (
            <Button
              type="button"
              icon={faCirclePlay}
              disabled={busy || !routineId || !targetSceneId}
              onClick={() => void onStartRun()}
            >
              Start
            </Button>
          ) : (
            <Button
              type="button"
              icon={faCircleStop}
              className="bg-amber-800 hover:bg-amber-700 dark:bg-amber-900"
              disabled={busy}
              onClick={() => void onStopRun()}
            >
              Stop
            </Button>
          )}
          <Button
            type="button"
            icon={faLightbulb}
            disabled={busy || !targetSceneId}
            onClick={() => void onResetSceneLights()}
          >
            Reset room lights
          </Button>
          <Button
            type="button"
            icon={faArrowsRotate}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={!targetScene || targetScene.items.length === 0}
            onClick={() => setCameraResetVersion((v) => v + 1)}
          >
            Reset camera
          </Button>
        </div>
        {showShapeHost && activeRun && definitionJsonForHost ? (
          <div className="mt-3" key={activeRun.run_id}>
            <ShapeAnimationRoutineHost
              sceneId={activeRun.scene_id}
              runId={activeRun.run_id}
              definitionJson={definitionJsonForHost}
              onSceneRefresh={() => void refreshTargetScene()}
              onError={(m) => setError(m)}
              onStopped={() => {
                setActiveRun(null);
                void refreshTargetScene();
              }}
            />
          </div>
        ) : null}
        {targetScene && targetScene.items.length > 0 ? (
          <div className="mt-4 min-h-[280px] w-full sm:min-h-[320px]">
            <SceneLightsCanvas
              items={targetScene.items}
              cameraPersistenceKey={`shape-unified-${targetScene.id}`}
              cameraResetVersion={cameraResetVersion}
            />
          </div>
        ) : targetSceneId ? (
          <p className="mt-3 text-xs text-slate-500">Could not load this room or it is empty.</p>
        ) : (
          <p className="mt-3 text-xs text-slate-500">Pick a room to see the 3D view.</p>
        )}
      </section>
    </div>
  );
}
