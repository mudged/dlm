"use client";

import { python } from "@codemirror/lang-python";
import { lintGutter } from "@codemirror/lint";
import CodeMirror from "@uiw/react-codemirror";
import {
  faBook,
  faCopy,
  faFloppyDisk,
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
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  createRoutine,
  deleteRoutine,
  fetchRoutine,
  patchRoutine,
  type RoutineDefinition,
} from "@/lib/routines";

const DEFAULT_SOURCE = `# Try top-level await with the scene API (examples in the help panel).
h = scene.height
lights = await scene.get_all_lights()
`;

export default function PythonRoutineEditorClient() {
  const router = useRouter();
  const search = useSearchParams();
  const idParam = search.get("id");

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [code, setCode] = useState(DEFAULT_SOURCE);
  const [routineId, setRoutineId] = useState<string | null>(idParam);
  const [loaded, setLoaded] = useState(!idParam);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
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
      setCode(r.python_source || DEFAULT_SOURCE);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Load failed");
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (idParam) {
      void load(idParam);
    } else {
      setRoutineId(null);
      setName("");
      setDescription("");
      setCode(DEFAULT_SOURCE);
      setLoaded(true);
    }
  }, [idParam, load]);

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
      if (routineId) {
        await patchRoutine(routineId, {
          name: n,
          description,
          python_source: sourceToSave,
        });
      } else {
        const created = await createRoutine({
          name: n,
          description,
          type: ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
          python_source: sourceToSave,
        });
        setRoutineId(created.id);
        router.replace(`/routines/python?id=${encodeURIComponent(created.id)}`);
      }
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
            Edit and save Python that runs in your browser (Pyodide) against the
            current scene when you start this routine from a{" "}
            <Link href="/scenes" className="text-sky-700 underline dark:text-sky-300">
              scene
            </Link>
            . Code and reference help stay on this page for beginners.
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
              className={`flex items-center gap-2 px-3 py-2 text-sm font-medium ${
                tab === "help"
                  ? "border-b-2 border-sky-600 text-sky-700 dark:text-sky-300"
                  : "text-slate-600 dark:text-slate-400"
              }`}
              onClick={() => setTab("help")}
            >
              <span>Scene API help</span>
            </button>
          </div>
          <div className={tab === "help" ? "hidden lg:block" : "block"}>
            <p className="mb-2 hidden text-xs text-slate-500 lg:block">
              Python editor: syntax highlighting, debounced syntax check (Pyodide{" "}
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">
                ast.parse
              </code>
              ), <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene.</code>{" "}
              completions, and format on save (
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">black</code>{" "}
              when available, else basic indent cleanup). Use top-level{" "}
              <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">
                await
              </code>{" "}
              for methods that call the server.
            </p>
            <div className="overflow-hidden rounded-lg border border-slate-300 dark:border-slate-600">
              <CodeMirror
                value={code}
                height="min(50vh, 420px)"
                theme="dark"
                extensions={extensions}
                onChange={(v) => setCode(v)}
                basicSetup={{
                  lineNumbers: true,
                  foldGutter: true,
                  closeBrackets: true,
                }}
              />
            </div>
          </div>
        </div>

        <aside
          className={`mt-4 space-y-3 text-sm text-slate-700 dark:text-slate-300 lg:mt-0 ${
            tab === "code" ? "hidden lg:block" : "block"
          }`}
        >
          <h2 className="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
            <FontAwesomeIcon icon={faBook} className="h-4 w-4 text-sky-600" />
            Scene API (Python)
          </h2>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            The object <code className="text-sky-700 dark:text-sky-300">scene</code>{" "}
            is bound to the scene you started the routine on. It talks to the same
            REST API as the rest of the app (no direct model URLs).
          </p>
          <ul className="list-inside list-disc space-y-2 text-xs">
            <li>
              <code>scene.height</code> — number, scene vertical size in meters from{" "}
              <code>GET …/dimensions</code> (<code>size.height</code>).
            </li>
            <li>
              <code>await scene.get_all_lights()</code> — list of lights with scene
              coordinates and state.
            </li>
            <li>
              <code>
                await scene.get_lights_within_sphere(center, radius)
              </code>{" "}
              — <code>center</code> is <code>{"{x,y,z}"}</code> in meters.
            </li>
            <li>
              <code>
                await scene.get_lights_within_cuboid(position, dimensions)
              </code>{" "}
              — <code>dimensions</code>: <code>width</code>, <code>height</code>,{" "}
              <code>depth</code>.
            </li>
            <li>
              <code>await scene.set_all_lights(...)</code> — JSON fields{" "}
              <code>on</code>, <code>color</code> (<code>#RRGGBB</code>),{" "}
              <code>brightness_pct</code>.
            </li>
            <li>
              <code>
                await scene.set_lights_in_sphere(center, radius, patch)
              </code>{" "}
              and cuboid variant — <code>patch</code> is a dict of state fields.
            </li>
            <li>
              <code>await scene.update_lights_batch(updates)</code> —{" "}
              <code>updates</code> matches <code>PATCH …/state/batch</code>.
            </li>
          </ul>
          <p className="text-xs text-slate-600 dark:text-slate-400">
            Your script is executed from the top on each loop (about every 50 ms
            between iterations, plus network time). Stop the routine from the scene
            page (that calls the API first, then signals the worker). If the worker
            does not exit cooperatively, the host terminates it after up to 5
            seconds (
            <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">
              T_force
            </code>
            ).
          </p>
        </aside>
      </div>
    </div>
  );
}
