/**
 * Pyodide worker for REQ-022: runs user Python against scene API via fetch.
 * Loaded from same origin as the app (static export / Go embed).
 */
import { loadPyodide } from "https://cdn.jsdelivr.net/pyodide/v0.26.4/full/pyodide.mjs";

let sceneId = "";
let stopped = false;
let pyodide = null;
let dimsCache = {};

async function apiJson(method, path, body) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const r = await fetch(path, opts);
  if (!r.ok) {
    const t = await r.text();
    throw new Error(t || String(r.status));
  }
  if (r.status === 204) return null;
  const ct = r.headers.get("content-type") || "";
  if (ct.includes("application/json")) return r.json();
  return null;
}

function point(p) {
  return { x: Number(p.x), y: Number(p.y), z: Number(p.z) };
}

async function refreshDims() {
  dimsCache = await apiJson("GET", `/api/v1/scenes/${sceneId}/dimensions`);
}

function buildScene() {
  const base = () => `/api/v1/scenes/${sceneId}`;
  return {
    /** REQ-030: same distribution as "#%06x" % random.randrange(0x1000000) in Pyodide. */
    random_hex_colour: () =>
      pyodide.runPython('"#%06x" % __import__("random").randrange(0x1000000)'),
    get width() {
      return dimsCache?.size?.width ?? 0;
    },
    get height() {
      return dimsCache?.size?.height ?? 0;
    },
    get depth() {
      return dimsCache?.size?.depth ?? 0;
    },
    get_all_lights: () => apiJson("GET", `${base()}/lights`),
    get_lights_within_sphere: (c, radius) =>
      apiJson("POST", `${base()}/lights/query/sphere`, {
        center: point(c),
        radius: Number(radius),
      }),
    get_lights_within_cuboid: (pos, dim) =>
      apiJson("POST", `${base()}/lights/query/cuboid`, {
        position: point(pos),
        dimensions: {
          width: Number(dim.width),
          height: Number(dim.height),
          depth: Number(dim.depth),
        },
      }),
    set_all_lights: (patch) => apiJson("PATCH", `${base()}/lights/state/scene`, patch),
    set_lights_in_sphere: (c, radius, patch) =>
      apiJson("PATCH", `${base()}/lights/state/sphere`, {
        center: point(c),
        radius: Number(radius),
        ...patch,
      }),
    set_lights_in_cuboid: (pos, dim, patch) =>
      apiJson("PATCH", `${base()}/lights/state/cuboid`, {
        position: point(pos),
        dimensions: {
          width: Number(dim.width),
          height: Number(dim.height),
          depth: Number(dim.depth),
        },
        ...patch,
      }),
    update_lights_batch: (updates) =>
      apiJson("PATCH", `${base()}/lights/state/batch`, { updates }),
  };
}

async function runUserLoop(source) {
  while (!stopped) {
    try {
      await refreshDims();
      await pyodide.runPythonAsync(source);
      self.postMessage({ type: "iterationComplete", sceneId });
    } catch (e) {
      self.postMessage({ type: "error", message: String(e) });
    }
    await new Promise((r) => setTimeout(r, 50));
    if (stopped) break;
    try {
      const data = await apiJson(
        "GET",
        `/api/v1/scenes/${sceneId}/routines/runs`,
      );
      if (!data?.runs?.length) {
        stopped = true;
        break;
      }
    } catch {
      stopped = true;
      break;
    }
  }
  self.postMessage({ type: "done" });
}

self.onmessage = async (ev) => {
  const msg = ev.data;
  if (msg?.type === "stop") {
    stopped = true;
    return;
  }
  if (msg?.type !== "init") return;

  sceneId = msg.sceneId;
  stopped = false;
  const source = msg.source || "";
  const indexURL =
    msg.indexURL || "https://cdn.jsdelivr.net/pyodide/v0.26.4/full/";

  try {
    pyodide = await loadPyodide({ indexURL });
    await refreshDims();
    pyodide.globals.set("scene", buildScene());
    self.postMessage({ type: "ready" });
    await runUserLoop(source);
  } catch (e) {
    self.postMessage({ type: "error", message: String(e) });
    self.postMessage({ type: "done" });
  }
};
