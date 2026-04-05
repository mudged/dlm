/**
 * Pyodide worker for REQ-022: ast.parse diagnostics and black (fallback: textwrap) formatting.
 * Same CDN as dlm-python-scene-worker.mjs per architecture §3.17.
 */
import { loadPyodide } from "https://cdn.jsdelivr.net/pyodide/v0.26.4/full/pyodide.mjs";

const INDEX_URL = "https://cdn.jsdelivr.net/pyodide/v0.26.4/full/";

let pyodide = null;
let blackReady = false;

async function getPyodide() {
  if (!pyodide) {
    pyodide = await loadPyodide({ indexURL: INDEX_URL });
  }
  return pyodide;
}

async function ensureBlack(p) {
  if (blackReady) {
    return;
  }
  await p.loadPackage("micropip");
  await p.runPythonAsync(`
import micropip
await micropip.install("black", keep_going=True)
`);
  blackReady = true;
}

self.onmessage = async (ev) => {
  const msg = ev.data;
  if (!msg || typeof msg.id !== "number") {
    return;
  }
  const { id } = msg;
  try {
    const p = await getPyodide();
    if (msg.type === "lint") {
      const source = typeof msg.source === "string" ? msg.source : "";
      p.globals.set("_dlm_lint_src", source);
      const proxy = await p.runPythonAsync(`
import ast
src = _dlm_lint_src
try:
    ast.parse(src)
    diagnostics = []
except SyntaxError as e:
    ln = e.lineno or 1
    col = e.offset if e.offset is not None else 1
    diagnostics = [{"line": ln, "column": col, "message": e.msg or "Syntax error"}]
diagnostics
`);
      let diagnostics = [];
      try {
        diagnostics = proxy.toJs({ dict_converter: Object });
      } finally {
        proxy.destroy?.();
      }
      self.postMessage({ type: "lintResult", id, diagnostics });
      return;
    }
    if (msg.type === "format") {
      const source = typeof msg.source === "string" ? msg.source : "";
      let text = source;
      let usedBlack = false;
      try {
        await ensureBlack(p);
        p.globals.set("_dlm_fmt_src", source);
        const out = await p.runPythonAsync(`
import black
black.format_str(_dlm_fmt_src, mode=black.Mode())
`);
        text = String(out);
        usedBlack = true;
      } catch {
        p.globals.set("_dlm_fmt_src2", source);
        const out = await p.runPythonAsync(`
import textwrap
raw = _dlm_fmt_src2.replace("\\t", "    ")
w = textwrap.dedent(raw)
if not w.endswith("\\n"):
    w = w + "\\n"
w
`);
        text = String(out);
        usedBlack = false;
      }
      self.postMessage({
        type: "formatResult",
        id,
        ok: true,
        text,
        usedBlack,
      });
      return;
    }
  } catch (e) {
    if (msg.type === "lint") {
      self.postMessage({
        type: "lintResult",
        id,
        diagnostics: [
          {
            line: 1,
            column: 1,
            message: String(e),
          },
        ],
      });
    } else if (msg.type === "format") {
      self.postMessage({
        type: "formatResult",
        id,
        ok: false,
        error: String(e),
      });
    }
  }
};
