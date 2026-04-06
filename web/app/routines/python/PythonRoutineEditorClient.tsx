"use client";

// Python routine editor — CodeMirror, in-browser run, visual debug, bottom command list.

import { python } from "@codemirror/lang-python";
import { lintGutter } from "@codemirror/lint";
import dynamic from "next/dynamic";
import { PythonCodeMirrorEditor } from "@/components/PythonCodeMirrorEditor";
import { PythonRoutineHost } from "@/components/PythonRoutineHost";
import { PythonSceneApiCatalogSection } from "@/components/PythonSceneApiCatalogSection";
import {
  faArrowsRotate,
  faBook,
  faCirclePlay,
  faCircleStop,
  faCopy,
  faFloppyDisk,
  faLightbulb,
  faRotate,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/Button";
import { formatPythonSource } from "@/lib/pythonEditorWorker";
import {
  pythonRoutineLinter,
  pythonSceneAutocompletion,
} from "@/lib/pythonRoutineCodemirror";
import { PYTHON_ROUTINE_DEFAULT_SOURCE } from "@/lib/pythonSceneApiCatalog";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
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

export default function PythonRoutineEditorClient() {
  const router = useRouter();
  const search = useSearchParams();
  const idParam = search.get("id");
  const sceneFromQuery = search.get("scene");

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [code, setCode] = useState(PYTHON_ROUTINE_DEFAULT_SOURCE);
  const [routineId, setRoutineId] = useState<string | null>(idParam);
  const [loaded, setLoaded] = useState(() => Boolean(idParam));
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const [scenesList, setScenesList] = useState<SceneSummary[] | null>(null);
  const [runSceneId, setRunSceneId] = useState("");
  const [debugSceneId, setDebugSceneId] = useState("");
  const [debugScene, setDebugScene] = useState<SceneDetail | null>(null);
  const [activeRun, setActiveRun] = useState<{
    run_id: string;
    scene_id: string;
  } | null>(null);
  const [cameraResetVersion, setCameraResetVersion] = useState(0);

  const [tab, setTab] = useState<"code" | "help">("code");

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
      router.replace("/routines/new?type=python_scene_script");
      return;
    }
    void load(idParam);
  }, [idParam, load, router]);

  useEffect(() => {
    void (async () => {
      try {
        const list = await fetchScenes();
        setScenesList(list);
        setDebugSceneId((prev) => {
          if (sceneFromQuery && list.some((s) => s.id === sceneFromQuery)) {
            return sceneFromQuery;
          }
          if (prev) {
            return prev;
          }
          return list.length > 0 ? list[0]!.id : "";
        });
        setRunSceneId((prev) => {
          if (sceneFromQuery && list.some((s) => s.id === sceneFromQuery)) {
            return sceneFromQuery;
          }
          if (prev) {
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
    if (!debugSceneId) {
      setDebugScene(null);
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const s = await fetchScene(debugSceneId);
        if (!cancelled) {
          setDebugScene(s);
        }
      } catch {
        if (!cancelled) {
          setDebugScene(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [debugSceneId]);

  const refreshDebugScene = useCallback(async () => {
    if (!debugSceneId) {
      return;
    }
    try {
      const s = await fetchScene(debugSceneId);
      setDebugScene(s);
    } catch {
      /* ignore */
    }
  }, [debugSceneId]);

  async function onSave() {
    setBusy(true);
    setError(null);
    try {
      const n = name.trim();
      if (!n) {
        throw new Error("Name is required.");
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
    if (!window.confirm("Delete this Python routine permanently?")) {
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
        "Replace the editor contents with the default sphere colour demo template?",
      )
    ) {
      return;
    }
    setCode(PYTHON_ROUTINE_DEFAULT_SOURCE);
  }

  async function onStartRun() {
    if (!routineId || !runSceneId) {
      setError("Choose a scene to run against.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await startSceneRoutine(runSceneId, routineId);
      setActiveRun({ run_id: res.run_id, scene_id: res.scene_id });
      if (!debugSceneId) {
        setDebugSceneId(runSceneId);
      }
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
      await refreshDebugScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Stop failed");
    } finally {
      setBusy(false);
    }
  }

  async function onResetSceneLights() {
    if (!debugSceneId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await patchSceneLightsStateScene(debugSceneId, {
        on: false,
        color: "#ffffff",
        brightness_pct: 100,
      });
      await refreshDebugScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Reset scene lights failed");
    } finally {
      setBusy(false);
    }
  }

  const workerSource = activeRun ? code : "";
  const showWorker =
    activeRun &&
    activeRun.scene_id === runSceneId &&
    Boolean(workerSource.trim());

  const onWorkerIteration = useCallback(
    (sid: string) => {
      if (sid === debugSceneId) {
        void refreshDebugScene();
      }
    },
    [debugSceneId, refreshDebugScene],
  );

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
            Python scene routine
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            Edit and save Python that runs in your browser (Pyodide) against a scene. Use
            visual debug below to preview a scene; start a run to drive lights from your
            script. Full API reference is at the bottom of this page.
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
          <span className="text-slate-600 dark:text-slate-400">Description</span>
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
            Reset to template
          </Button>
        </div>
      </div>

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="run-routine-heading"
      >
        <h2
          id="run-routine-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          Run on scene
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          Start begins your routine on the server and runs this copy of your script in the
          browser. Stop ends the run and stops the script.
        </p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
          <label className="flex min-w-[200px] flex-1 flex-col gap-1 text-xs">
            <span className="text-slate-600 dark:text-slate-400">Scene for run</span>
            <select
              value={runSceneId}
              onChange={(e) => {
                const v = e.target.value;
                setRunSceneId(v);
              }}
              className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            >
              <option value="">Select scene…</option>
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
              disabled={busy || !routineId || !runSceneId}
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
        </div>
        {showWorker ? (
          <div className="mt-3">
            <PythonRoutineHost
              sceneId={activeRun.scene_id}
              source={workerSource}
              onWorkerMessage={(msg) => setError(msg)}
              onIterationComplete={onWorkerIteration}
            />
          </div>
        ) : null}
      </section>

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="visual-debug-heading"
      >
        <h2
          id="visual-debug-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          Visual debug
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          Inspect the room in 3D. When your run uses the same scene as the one you pick
          here, the picture updates after each time your script runs. Reset scene lights
          sets every light in that room back to the usual defaults (off, white, full
          brightness) and does not stop an active run.
        </p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
          <label className="flex min-w-[200px] flex-1 flex-col gap-1 text-xs">
            <span className="text-slate-600 dark:text-slate-400">Debug scene</span>
            <select
              value={debugSceneId}
              onChange={(e) => setDebugSceneId(e.target.value)}
              className="min-h-11 rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            >
              <option value="">Select scene…</option>
              {(scenesList ?? []).map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          <Button
            type="button"
            icon={faLightbulb}
            disabled={busy || !debugSceneId}
            onClick={() => void onResetSceneLights()}
          >
            Reset scene lights
          </Button>
          <Button
            type="button"
            icon={faArrowsRotate}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={!debugScene || debugScene.items.length === 0}
            onClick={() => setCameraResetVersion((v) => v + 1)}
          >
            Reset camera
          </Button>
        </div>
        {debugScene && debugScene.items.length > 0 ? (
          <div className="mt-4 min-h-[280px] w-full sm:min-h-[320px]">
            <SceneLightsCanvas
              items={debugScene.items}
              cameraPersistenceKey={`python-debug-${debugScene.id}`}
              cameraResetVersion={cameraResetVersion}
            />
          </div>
        ) : debugSceneId ? (
          <p className="mt-3 text-xs text-slate-500">Could not load scene or scene is empty.</p>
        ) : (
          <p className="mt-3 text-xs text-slate-500">Select a scene to show the 3D view.</p>
        )}
      </section>

      <div className="lg:grid lg:grid-cols-2 lg:gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex gap-2 border-b border-slate-200 dark:border-slate-700 lg:hidden">
            <button
              type="button"
              className={`px-3 py-2 text-sm font-medium ${
                tab === "code"
                  ? "border-b-2 border-sky-600 text-sky-700 dark:text-sky-300"
                  : "text-slate-600 dark:text-slate-400"
              }`}
              onClick={() => setTab("code")}
            >
              Code
            </button>
            <button
              type="button"
              className={`px-3 py-2 text-sm font-medium ${
                tab === "help"
                  ? "border-b-2 border-sky-600 text-sky-700 dark:text-sky-300"
                  : "text-slate-600 dark:text-slate-400"
              }`}
              onClick={() => setTab("help")}
            >
              Scene API help
            </button>
            <a
              href="#python-scene-api-catalog"
              className="px-3 py-2 text-sm font-medium text-slate-600 dark:text-slate-400"
            >
              Full API ↓
            </a>
          </div>
          <div className={tab === "help" ? "hidden lg:block" : "block"}>
            <p className="mb-2 hidden text-xs text-slate-500 lg:block">
              Python editor: syntax highlighting, debounced syntax check (Pyodide{" "}
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">ast.parse</code>
              ), <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene.</code>{" "}
              completions, and format on save (
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">black</code> when
              available). Use top-level{" "}
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">await</code> for
              async methods.
            </p>
            <PythonCodeMirrorEditor
              value={code}
              onChange={setCode}
              extensions={extensions}
            />
          </div>
        </div>

        <aside
          className={`mt-4 space-y-3 text-sm text-slate-700 dark:text-slate-300 lg:mt-0 ${
            tab === "code" ? "hidden lg:block" : "block"
          }`}
        >
          <h2 className="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
            <FontAwesomeIcon icon={faBook} className="h-4 w-4 text-sky-600" />
            Scene API (Python) — short guide
          </h2>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            The object{" "}
            <code className="text-sky-700 dark:text-sky-300">scene</code> stands for the
            room you chose when you started the run. Use it to read sizes and change lights
            in that room.
          </p>
          <ul className="list-inside list-disc space-y-2 text-xs">
            <li>
              <code>scene.width</code>, <code>scene.height</code>, <code>scene.depth</code> —
              room size in metres (width, height, depth).
            </li>
            <li>
              <code>await scene.get_all_lights()</code> — list with <code>sx</code>,{" "}
              <code>sy</code>, <code>sz</code> and state.
            </li>
            <li>
              <code>await scene.get_lights_within_sphere(center, radius)</code> —{" "}
              <code>center</code> is <code>{"{x,y,z}"}</code>.
            </li>
            <li>
              <code>await scene.get_lights_within_cuboid(position, dimensions)</code> —{" "}
              <code>dimensions</code>: width, height, depth.
            </li>
            <li>
              <code>await scene.set_all_lights(...)</code> — fields <code>on</code>,{" "}
              <code>color</code> (<code>#RRGGBB</code>), <code>brightness_pct</code>.
            </li>
            <li>
              <code>await scene.set_lights_in_sphere(center, radius, patch)</code> and cuboid
              variant.
            </li>
            <li>
              <code>await scene.update_lights_batch(updates)</code> — change several specific
              lights in one list.
            </li>
          </ul>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            The script runs about every 50 ms between iterations (plus network). See the
            complete list with snippets at the bottom of this page (
            <a
              href="#python-scene-api-catalog"
              className="text-sky-700 underline dark:text-sky-300"
            >
              Scene API reference
            </a>
            ).
          </p>
        </aside>

      </div>

      <PythonSceneApiCatalogSection />
    </div>
  );
}
