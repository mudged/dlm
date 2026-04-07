"use client";

import { Button } from "@/components/ui/Button";
import { PYTHON_SCENE_API_CATALOG_FULL } from "@/lib/pythonSceneApiCatalog";
import { faSquarePlus } from "@fortawesome/free-solid-svg-icons";
import { useMemo, useState } from "react";

type Props = {
  /** REQ-024: inserts the currently shown example into the editor. REQ-032 samples replace the whole buffer. */
  onInsertSnippet: (snippet: string, options?: { replaceAll?: boolean }) => void;
};

export function PythonSceneApiCatalogSection({ onInsertSnippet }: Props) {
  const firstId = PYTHON_SCENE_API_CATALOG_FULL[0]?.id ?? "";
  const [selectedId, setSelectedId] = useState(firstId);

  const entry = useMemo(
    () =>
      PYTHON_SCENE_API_CATALOG_FULL.find((e) => e.id === selectedId) ??
      PYTHON_SCENE_API_CATALOG_FULL[0],
    [selectedId],
  );

  return (
    <section
      id="python-scene-api-catalog"
      className="scroll-mt-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
      aria-labelledby="python-scene-api-catalog-heading"
    >
      <h2
        id="python-scene-api-catalog-heading"
        className="text-lg font-semibold text-slate-900 dark:text-slate-100"
      >
        Scene API — pick one to read more
      </h2>
      <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
        <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene</code> is the room
        your script runs on after you press Start. Below is every name you can use, with a
        short example. Comments in the example explain each part.
      </p>

      <label className="mt-4 flex flex-col gap-1 text-xs">
        <span className="text-slate-600 dark:text-slate-400">Choose a name</span>
        <select
          value={selectedId}
          onChange={(e) => setSelectedId(e.target.value)}
          className="min-h-11 max-w-xl rounded border border-slate-300 bg-white px-2 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
        >
          {PYTHON_SCENE_API_CATALOG_FULL.map((e) => (
            <option key={e.id} value={e.id}>
              {e.label} ({e.kind})
            </option>
          ))}
        </select>
      </label>

      {entry ? (
        <div className="mt-4 space-y-3">
          <p className="text-sm text-slate-700 dark:text-slate-300">{entry.description}</p>
          <pre className="max-w-full overflow-x-auto rounded-lg bg-slate-900 p-3 text-xs text-slate-100">
            <code>{entry.snippet.trimEnd()}</code>
          </pre>
          <Button
            type="button"
            icon={faSquarePlus}
            onClick={() =>
              onInsertSnippet(entry.kind === "sample" ? entry.snippet : entry.snippet.trimEnd(), {
                replaceAll: entry.kind === "sample",
              })
            }
          >
            {entry.kind === "sample"
              ? "Replace code with this full sample"
              : "Put this example in your code"}
          </Button>
        </div>
      ) : null}
    </section>
  );
}
