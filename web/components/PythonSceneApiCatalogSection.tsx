"use client";

import { SCENE_API_CATALOG } from "@/lib/pythonSceneApiCatalog";

export function PythonSceneApiCatalogSection() {
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
        Scene API reference (complete)
      </h2>
      <p className="mt-1 text-xs text-slate-600 dark:text-slate-400">
        Every <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">scene</code> name
        you can use in your Python here. Width, height, and depth are the room size in metres.
      </p>
      <ul className="mt-4 space-y-6">
        {SCENE_API_CATALOG.map((e) => (
          <li key={e.id} id={e.id} className="border-t border-slate-100 pt-4 first:border-t-0 first:pt-0 dark:border-slate-800">
            <h3 className="font-mono text-sm font-semibold text-sky-800 dark:text-sky-300">
              {e.label}
              <span className="ml-2 font-sans text-xs font-normal text-slate-500">
                ({e.kind})
              </span>
            </h3>
            <p className="mt-1 text-xs text-slate-700 dark:text-slate-300">{e.description}</p>
            <pre className="mt-2 max-w-full overflow-x-auto rounded-lg bg-slate-900 p-3 text-xs text-slate-100">
              <code>{e.snippet.trimEnd()}</code>
            </pre>
          </li>
        ))}
      </ul>
    </section>
  );
}
