"use client";

// Python routine editor — CodeMirror, API reference below editor, unified run + 3D view (§4.13).

import type { EditorView } from "@codemirror/view";
import { python } from "@codemirror/lang-python";
import { lintGutter } from "@codemirror/lint";
import dynamic from "next/dynamic";
import { PythonCodeMirrorEditor } from "@/components/PythonCodeMirrorEditor";
import { PythonSceneApiCatalogSection } from "@/components/PythonSceneApiCatalogSection";
import {
  faArrowsRotate,
  faCirclePlay,
  faCircleStop,
  faCopy,
  faFileImport,
  faFloppyDisk,
  faLightbulb,
  faRotate,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/ui/Button";
import { formatPythonSource } from "@/lib/pythonEditorWorker";
import { insertSnippetInPythonEditor } from "@/lib/insertPythonEditorSnippet";
import {
  pythonRoutineLinter,
  pythonSceneAutocompletion,
} from "@/lib/pythonRoutineCodemirror";
import {
  PYTHON_ROUTINE_DEFAULT_SOURCE,
} from "@/lib/pythonSceneApiCatalog";
import {
  PYTHON_SAMPLE_GROWING_SPHERE_SOURCE,
  PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE,
  PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE,
} from "@/lib/pythonRoutineSamples";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  createRoutine,
  deleteRoutine,
  fetchRoutine,
  fetchSceneRoutineRuns,
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

export default function PythonRoutineEditorClient() {
  const router = useRouter();
  const search = useSearchParams();
  const idParam = search.get("id");
  const sceneFromQuery = search.get("scene");

  const editorViewRef = useRef<EditorView | null>(null);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [code, setCode] = useState(PYTHON_ROUTINE_DEFAULT_SOURCE);
  const [routineId, setRoutineId] = useState<string | null>(idParam);
  const [loaded, setLoaded] = useState(() => Boolean(idParam));
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const [scenesList, setScenesList] = useState<SceneSummary[] | null>(null);
  /** REQ-027: one scene for both Start/Stop and the 3D picture */
  const [targetSceneId, setTargetSceneId] = useState("");
  const [targetScene, setTargetScene] = useState<SceneDetail | null>(null);
  const [activeRun, setActiveRun] = useState<{
    run_id: string;
    scene_id: string;
  } | null>(null);
  const [cameraResetVersion, setCameraResetVersion] = useState(0);

  const extensions = useMemo(
    () => [
      python(),
      lintGutter(),
      pythonRoutineLinter(),
      pythonSceneAutocompletion(),
    ],
    [],
  );

  const load = useCallback(async (rid: string) => {
    setError(null);
    setLoaded(false);
    try {
      const r: RoutineDefinition = await fetchRoutine(rid);
      if (r.type !== ROUTINE_TYPE_PYTHON_SCENE_SCRIPT) {
        throw new Error("This routine is not a Python scene script.");
      }
      setRoutineId(r.id);
      setName(r.name);
      setDescription(r.description);
      setCode(r.python_source?.trim() ? r.python_source : PYTHON_ROUTINE_DEFAULT_SOURCE);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Load failed");
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (!idParam) {
      router.replace("/routines/new?kind=python");
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

  const onInsertSnippet = useCallback(
    (snippet: string, options?: { replaceAll?: boolean }) => {
      if (options?.replaceAll) {
        setCode(snippet);
        return;
      }
      const view = editorViewRef.current;
      if (view) {
        insertSnippetInPythonEditor(view, snippet);
      } else {
        setCode((prev) => {
          const needsNl = prev.length > 0 && !prev.endsWith("\n");
          return prev + (needsNl ? "\n" : "") + snippet;
        });
      }
    },
    [],
  );

  async function onSave() {
    setBusy(true);
    setError(null);
    try {
      const n = name.trim();
      if (!n) {
        throw new Error("Please enter a name.");
      }
      const formatted = await formatPythonSource(code);
      let sourceToSave = code;
      if (formatted.ok) {
        sourceToSave = formatted.text;
        setCode(formatted.text);
      }
      if (!routineId) {
        throw new Error("Routine id missing; create routines from Routines → New routine.");
      }
      await patchRoutine(routineId, {
        name: n,
        description,
        python_source: sourceToSave,
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Save failed");
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
      const dup = await createRoutine({
        name: `${name.trim() || "routine"} (copy)`,
        description,
        type: ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
        python_source: code,
      });
      router.push(`/routines/python?id=${encodeURIComponent(dup.id)}`);
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

  function onResetTemplate() {
    if (
      !window.confirm(
        "Replace your code with the starter demo (colours in a sphere)? Your current text will be lost unless you saved.",
      )
    ) {
      return;
    }
    setCode(PYTHON_ROUTINE_DEFAULT_SOURCE);
  }

  function loadGrowingSphereSample() {
    if (
      !window.confirm(
        "Replace your code with the full growing-sphere sample? Save first if you need a copy of what you have now.",
      )
    ) {
      return;
    }
    setCode(PYTHON_SAMPLE_GROWING_SPHERE_SOURCE);
  }

  function loadSweepingCuboidSample() {
    if (
      !window.confirm(
        "Replace your code with the full sweeping-cuboid sample? Save first if you need a copy of what you have now.",
      )
    ) {
      return;
    }
    setCode(PYTHON_SAMPLE_SWEEPING_CUBOID_SOURCE);
  }

  function loadRandomColourCycleSample() {
    if (
      !window.confirm(
        "Replace your code with the full random colour cycle sample? Save first if you need a copy of what you have now.",
      )
    ) {
      return;
    }
    setCode(PYTHON_SAMPLE_RANDOM_COLOUR_CYCLE_ALL_SOURCE);
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
      let runId =
        activeRun?.scene_id === targetSceneId ? activeRun.run_id : null;
      if (!runId) {
        const runs = await fetchSceneRoutineRuns(targetSceneId);
        if (runs.length > 0) {
          runId = runs[0].id;
        }
      }
      if (runId) {
        await stopSceneRoutineRun(targetSceneId, runId);
        setActiveRun(null);
      }
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
            Python room routine
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            Write short Python here. After you save and start, it runs on the server (python3)
            and changes lights in the room you pick. Under the code is a list of every command
            — open one, then press the button to copy the example into your script. Below that
            you start the run and watch the same room in 3D (server push).
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
            disabled={busy || !routineId}
            onClick={() => void onDelete()}
          >
            Delete
          </Button>
          <Button
            type="button"
            icon={faRotate}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={busy}
            onClick={onResetTemplate}
          >
            Reset to starter code
          </Button>
          <Button
            type="button"
            icon={faFileImport}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={busy}
            onClick={loadGrowingSphereSample}
          >
            Load growing sphere sample
          </Button>
          <Button
            type="button"
            icon={faFileImport}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={busy}
            onClick={loadSweepingCuboidSample}
          >
            Load sweeping cuboid sample
          </Button>
          <Button
            type="button"
            icon={faFileImport}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={busy}
            onClick={loadRandomColourCycleSample}
          >
            Load random colour cycle sample
          </Button>
        </div>
      </div>

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="python-code-heading"
      >
        <h2
          id="python-code-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          Your code
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          The editor checks Python as you type and can suggest words after{" "}
          <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene.</code> Save runs
          the formatter when it is available. You can use{" "}
          <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">await</code> at the top
          level for scene commands.
        </p>
        <div className="mt-3">
          <PythonCodeMirrorEditor
            value={code}
            onChange={setCode}
            extensions={extensions}
            editorViewRef={editorViewRef}
          />
        </div>
      </section>

      <PythonSceneApiCatalogSection onInsertSnippet={onInsertSnippet} />

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="run-watch-heading"
      >
        <h2
          id="run-watch-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          Run your script and watch the room
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          Choose one room for both the run and the 3D picture. Start tells the server you are
          running and plays your script in a loop here. Stop ends the loop. The picture updates
          when your script changes lights. Reset room lights sets every light back to off,
          white, and full brightness — it does not press Stop for you.
        </p>
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
        {activeRun && activeRun.scene_id === targetSceneId ? (
          <p className="mt-3 text-xs text-slate-500 dark:text-slate-400">
            Run is active on the server. The 3D view updates from the server; you do not need to
            keep this page open.
          </p>
        ) : null}
        {targetScene && targetScene.items.length > 0 ? (
          <div className="mt-4 min-h-[280px] w-full sm:min-h-[320px]">
            <SceneLightsCanvas
              items={targetScene.items}
              cameraPersistenceKey={`python-unified-${targetScene.id}`}
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
