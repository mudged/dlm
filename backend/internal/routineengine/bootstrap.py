"""DLM server-side Python routine runner (REQ-022 / architecture §3.17)."""
import ast
import asyncio
import json
import os
import random
import sys
import urllib.error
import urllib.request

API_BASE = os.environ.get("DLM_API_BASE", "http://127.0.0.1:8080").rstrip("/")
SCENE_ID = os.environ["DLM_SCENE_ID"]
SOURCE_PATH = os.environ["DLM_USER_SOURCE_PATH"]
ITER_GAP_SEC = float(os.environ.get("DLM_ITER_GAP_SEC", "0.05"))


def _http_json_sync(method, path, body=None):
    url = API_BASE + path
    data = None
    headers = {}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            if resp.status == 204:
                return None
            raw = resp.read()
            if not raw:
                return None
            return json.loads(raw.decode("utf-8"))
    except urllib.error.HTTPError as e:
        try:
            detail = e.read().decode("utf-8", errors="replace")
        except Exception:
            detail = str(e)
        raise RuntimeError(f"HTTP {e.code}: {detail}") from e


async def _http_json(method, path, body=None):
    return await asyncio.to_thread(_http_json_sync, method, path, body)


_dims: dict = {}


async def _refresh_dims():
    global _dims
    path = f"/api/v1/scenes/{SCENE_ID}/dimensions"
    _dims = await _http_json("GET", path, None) or {}


class Scene:
    """Async scene shim — same routes as §3.15 / dlm-python-scene-worker.mjs."""

    def random_hex_colour(self) -> str:
        return "#%06x" % random.randrange(0x1000000)

    @property
    def width(self) -> float:
        return float((_dims.get("size") or {}).get("width") or 0)

    @property
    def height(self) -> float:
        return float((_dims.get("size") or {}).get("height") or 0)

    @property
    def depth(self) -> float:
        return float((_dims.get("size") or {}).get("depth") or 0)

    @property
    def max_x(self) -> float:
        return float((_dims.get("max") or {}).get("x") or 0)

    @property
    def max_y(self) -> float:
        return float((_dims.get("max") or {}).get("y") or 0)

    @property
    def max_z(self) -> float:
        return float((_dims.get("max") or {}).get("z") or 0)

    def _base(self) -> str:
        return f"/api/v1/scenes/{SCENE_ID}"

    async def get_all_lights(self):
        return await _http_json("GET", f"{self._base()}/lights", None)

    async def get_lights_within_sphere(self, center, radius):
        c = {"x": float(center["x"]), "y": float(center["y"]), "z": float(center["z"])}
        return await _http_json(
            "POST",
            f"{self._base()}/lights/query/sphere",
            {"center": c, "radius": float(radius)},
        )

    async def get_lights_within_cuboid(self, position, dimensions):
        pos = {"x": float(position["x"]), "y": float(position["y"]), "z": float(position["z"])}
        dim = {
            "width": float(dimensions["width"]),
            "height": float(dimensions["height"]),
            "depth": float(dimensions["depth"]),
        }
        return await _http_json(
            "POST",
            f"{self._base()}/lights/query/cuboid",
            {"position": pos, "dimensions": dim},
        )

    async def set_all_lights(self, patch):
        p = dict(patch)
        return await _http_json("PATCH", f"{self._base()}/lights/state/scene", p)

    async def set_lights_in_sphere(self, center, radius, patch):
        c = {"x": float(center["x"]), "y": float(center["y"]), "z": float(center["z"])}
        body = {"center": c, "radius": float(radius)}
        body.update(patch)
        return await _http_json("PATCH", f"{self._base()}/lights/state/sphere", body)

    async def set_lights_in_cuboid(self, position, dimensions, patch):
        pos = {"x": float(position["x"]), "y": float(position["y"]), "z": float(position["z"])}
        dim = {
            "width": float(dimensions["width"]),
            "height": float(dimensions["height"]),
            "depth": float(dimensions["depth"]),
        }
        body = {"position": pos, "dimensions": dim}
        body.update(patch)
        return await _http_json("PATCH", f"{self._base()}/lights/state/cuboid", body)

    async def update_lights_batch(self, updates):
        return await _http_json(
            "PATCH",
            f"{self._base()}/lights/state/batch",
            {"updates": list(updates)},
        )


async def _main() -> None:
    with open(SOURCE_PATH, encoding="utf-8") as f:
        user_source = f.read()

    scene = Scene()
    g: dict = {
        "scene": scene,
        "asyncio": asyncio,
        "math": __import__("math"),
        "time": __import__("time"),
    }

    while True:
        await _refresh_dims()
        code = compile(
            user_source,
            "<user_routine>",
            "exec",
            flags=ast.PyCF_ALLOW_TOP_LEVEL_AWAIT,
        )
        res = eval(code, g, g)
        if asyncio.iscoroutine(res):
            await res

        await asyncio.sleep(ITER_GAP_SEC)

        data = await _http_json("GET", f"/api/v1/scenes/{SCENE_ID}/routines/runs", None)
        runs = (data or {}).get("runs") or []
        if not runs:
            break


if __name__ == "__main__":
    try:
        asyncio.run(_main())
    except KeyboardInterrupt:
        sys.exit(130)
