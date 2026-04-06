"use client";

// REQ-022–REQ-027 / architecture §4.13 — CodeMirror 6 + visual debug + bottom API catalog.

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
  /** Single scene for both Pyodide run target and 3D visual debug (combined section). */
  const [selectedSceneId, setSelectedSceneId] = useState("");
  const [selectedScene, setSelectedScene] = useState<SceneDetail | null>(null);
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
        setSelectedSceneId((prev) => {
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
    if (!selectedSceneId) {
      setSelectedScene(null);
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const s = await fetchScene(selectedSceneId);
        if (!cancelled) {
          setSelectedScene(s);
        }
      } catch {
        if (!cancelled) {
          setSelectedScene(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [selectedSceneId]);

  const refreshSelectedScene = useCallback(async () => {
    if (!selectedSceneId) {
      return;
    }
    try {
      const s = await fetchScene(selectedSceneId);
      setSelectedScene(s);
    } catch {
      /* ignore */
    }
  }, [selectedSceneId]);

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
        "Replace your code with the starter demo (coloured lights in a ball)?",
      )
    ) {
      return;
    }
    setCode(PYTHON_ROUTINE_DEFAULT_SOURCE);
  }

  async function onStartRun() {
    if (!routineId || !selectedSceneId) {
      setError("Choose a scene to run against.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      // Start/stop only toggles routine run state — never resets light state (REQ-014 reset is explicit button only).
      const res = await startSceneRoutine(selectedSceneId, routineId);
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
      // Refetch scene to reflect final script state — does not PATCH lights (no implicit reset).
      await refreshSelectedScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Stop failed");
    } finally {
      setBusy(false);
    }
  }

  async function onResetSceneLights() {
    if (!selectedSceneId) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await patchSceneLightsStateScene(selectedSceneId, {
        on: false,
        color: "#ffffff",
        brightness_pct: 100,
      });
      await refreshSelectedScene();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Reset scene lights failed");
    } finally {
      setBusy(false);
    }
  }

  const workerSource = activeRun ? code : "";
  const showWorker =
    activeRun &&
    activeRun.scene_id === selectedSceneId &&
    Boolean(workerSource.trim());

  const onWorkerIteration = useCallback(
    (sid: string) => {
      if (sid === selectedSceneId) {
        void refreshSelectedScene();
      }
    },
    [selectedSceneId, refreshSelectedScene],
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
            Python lights routine
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            Type your Python code up here. Pick a room (scene) lower down, press Start, and
            watch the lights change in the picture. Start and Stop do not turn lights back to
            normal — use Reset scene lights for that. Scroll to the bottom for examples of
            every <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene</code>{" "}
            command.
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
            Start over with demo code
          </Button>
        </div>
      </div>

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
              Tips
            </button>
            <a
              href="#python-scene-api-catalog"
              className="px-3 py-2 text-sm font-medium text-slate-600 dark:text-slate-400"
            >
              All examples ↓
            </a>
          </div>
          <div className={tab === "help" ? "hidden lg:block" : "block"}>
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
            Quick tips for your code
          </h2>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            In Python, <code className="text-sky-700 dark:text-sky-300">scene</code> is the
            room you picked when you press Start. You ask it for lights, change colours, and
            so on.
          </p>
          <ul className="list-inside list-disc space-y-2 text-xs">
            <li>
              <code>scene.width</code>, <code>scene.height</code>, <code>scene.depth</code> —
              how big the room is in metres (length, height, depth).
            </li>
            <li>
              <code>await scene.get_all_lights()</code> — get every light and where it sits.
            </li>
            <li>
              <code>await scene.get_lights_within_sphere(…)</code> — lights inside a ball
              shape; the middle point looks like <code>{"{x, y, z}"}</code>.
            </li>
            <li>
              <code>await scene.get_lights_within_cuboid(…)</code> — lights inside a box.
            </li>
            <li>
              <code>await scene.set_all_lights(…)</code> — set every light the same way: on or
              off, colour like <code>#ff6600</code>, and how bright (0–100).
            </li>
            <li>
              <code>await scene.set_lights_in_sphere(…)</code> or{" "}
              <code>set_lights_in_cuboid(…)</code> — change only the lights in that shape.
            </li>
            <li>
              <code>await scene.update_lights_batch(…)</code> — change several specific lights
              in one go.
            </li>
          </ul>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            Your code runs again and again in a short loop while the routine is on. Copy-paste
            examples for each command are at the bottom (
            <a
              href="#python-scene-api-catalog"
              className="text-sky-700 underline dark:text-sky-300"
            >
              All scene commands
            </a>
            ).
          </p>
        </aside>

      </div>

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="run-and-debug-heading"
      >
        <h2
          id="run-and-debug-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          Try it on a room
        </h2>
        <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
          Choose the same room for running your code and for the 3D picture. Start runs your
          code; Stop stops it. Neither button clears your light colours — tap Reset scene
          lights if you want everything back to the usual defaults (the routine can keep
          running).
        </p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
          <label className="flex min-w-[200px] flex-1 flex-col gap-1 text-xs">
            <span className="text-slate-600 dark:text-slate-400">Scene</span>
            <select
              value={selectedSceneId}
              onChange={(e) => setSelectedSceneId(e.target.value)}
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
              disabled={busy || !routineId || !selectedSceneId}
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
            disabled={busy || !selectedSceneId}
            onClick={() => void onResetSceneLights()}
          >
            Reset scene lights
          </Button>
          <Button
            type="button"
            icon={faArrowsRotate}
            className="bg-slate-600 hover:bg-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600"
            disabled={!selectedScene || selectedScene.items.length === 0}
            onClick={() => setCameraResetVersion((v) => v + 1)}
          >
            Reset camera
          </Button>
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
        {selectedScene && selectedScene.items.length > 0 ? (
          <div className="mt-4 min-h-[280px] w-full sm:min-h-[320px]">
            <SceneLightsCanvas
              items={selectedScene.items}
              cameraPersistenceKey={`python-debug-${selectedScene.id}`}
              cameraResetVersion={cameraResetVersion}
            />
          </div>
        ) : selectedSceneId ? (
          <p className="mt-3 text-xs text-slate-500">
            This room could not be loaded, or it has no lights to show.
          </p>
        ) : (
          <p className="mt-3 text-xs text-slate-500">
            Pick a room from the list to see it in 3D.
          </p>
        )}
      </section>

      <PythonSceneApiCatalogSection />
    </div>
  );
}
