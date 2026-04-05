import { autocompletion, type CompletionContext } from "@codemirror/autocomplete";
import { linter, type Diagnostic } from "@codemirror/lint";
import type { Text } from "@codemirror/state";
import { lintPythonSource, type PyLintDiagnostic } from "@/lib/pythonEditorWorker";

/** Architecture §3.17 / §4.13 — static `scene` API surface for completion. */
export const SCENE_API_COMPLETIONS: {
  label: string;
  detail: string;
  type: string;
}[] = [
  { label: "height", detail: "float (m)", type: "property" },
  {
    label: "get_all_lights",
    detail: "() → await",
    type: "function",
  },
  {
    label: "get_lights_within_sphere",
    detail: "(center, radius) → await",
    type: "function",
  },
  {
    label: "get_lights_within_cuboid",
    detail: "(position, dimensions) → await",
    type: "function",
  },
  {
    label: "set_all_lights",
    detail: "(patch) → await",
    type: "function",
  },
  {
    label: "set_lights_in_sphere",
    detail: "(center, radius, patch) → await",
    type: "function",
  },
  {
    label: "set_lights_in_cuboid",
    detail: "(position, dimensions, patch) → await",
    type: "function",
  },
  {
    label: "update_lights_batch",
    detail: "(updates) → await",
    type: "function",
  },
];

function sceneCompletions(context: CompletionContext) {
  const before = context.matchBefore(/scene\.[\w]*/);
  if (!before) {
    return null;
  }
  if (before.from === before.to && !context.explicit) {
    return null;
  }
  const dot = before.text.indexOf(".");
  if (dot < 0) {
    return null;
  }
  const afterDot = before.text.slice(dot + 1);
  const from = before.from + dot + 1;
  const lower = afterDot.toLowerCase();
  const options = SCENE_API_COMPLETIONS.filter(
    (o) =>
      o.label.startsWith(afterDot) ||
      o.label.toLowerCase().startsWith(lower),
  ).map((o) => ({
    label: o.label,
    detail: o.detail,
    type: o.type,
  }));
  if (options.length === 0) {
    return null;
  }
  return { from, options, validFor: /^[\w]*$/ };
}

function pyDiagToCm(doc: Text, d: PyLintDiagnostic): Diagnostic {
  const lineNo = Math.min(Math.max(1, d.line), doc.lines);
  const ln = doc.line(lineNo);
  let col0 = Math.max(0, d.column - 1);
  if (col0 > ln.length) {
    col0 = ln.length;
  }
  const from = ln.from + col0;
  const to = Math.min(from + 1, ln.to);
  return {
    from,
    to,
    message: d.message,
    severity: "error",
  };
}

let lintGeneration = 0;

export function pythonRoutineLinter() {
  return linter(
    async (view) => {
      const gen = ++lintGeneration;
      const text = view.state.doc.toString();
      let raw: PyLintDiagnostic[];
      try {
        raw = await lintPythonSource(text);
      } catch {
        return [];
      }
      if (gen !== lintGeneration) {
        return [];
      }
      return raw.map((d) => pyDiagToCm(view.state.doc, d));
    },
    { delay: 450 },
  );
}

export function pythonSceneAutocompletion() {
  return autocompletion({ override: [sceneCompletions] });
}
