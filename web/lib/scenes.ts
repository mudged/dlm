import type { Light } from "@/lib/models";

export type SceneSummary = {
  id: string;
  name: string;
  created_at: string;
  model_count: number;
};

export type SceneLight = Light & {
  sx: number;
  sy: number;
  sz: number;
};

export type SceneItem = {
  model_id: string;
  name: string;
  offset_x: number;
  offset_y: number;
  offset_z: number;
  lights: SceneLight[];
};

export type SceneDetail = {
  id: string;
  name: string;
  created_at: string;
  items: SceneItem[];
};

export type SceneDimensionsResponse = {
  origin: { x: number; y: number; z: number };
  size: { width: number; height: number; depth: number };
  max: { x: number; y: number; z: number };
  margin_m: number;
};

export type SceneLightFlatRow = {
  scene_id: string;
  model_id: string;
  light_id: number;
  sx: number;
  sy: number;
  sz: number;
};

export async function fetchScenes(): Promise<SceneSummary[]> {
  const res = await fetch("/api/v1/scenes", { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`scenes list failed (${res.status})`);
  }
  return res.json() as Promise<SceneSummary[]>;
}

export async function fetchSceneDimensions(
  sceneId: string,
): Promise<SceneDimensionsResponse> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/dimensions`,
    { cache: "no-store" },
  );
  if (!res.ok) {
    throw new Error(`scene dimensions failed (${res.status})`);
  }
  return res.json() as Promise<SceneDimensionsResponse>;
}

export async function fetchSceneLightsFlat(
  sceneId: string,
): Promise<SceneLightFlatRow[]> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/lights`,
    { cache: "no-store" },
  );
  if (!res.ok) {
    throw new Error(`scene lights list failed (${res.status})`);
  }
  return res.json() as Promise<SceneLightFlatRow[]>;
}

export async function patchSceneLightsStateBatch(
  sceneId: string,
  updates: {
    model_id: string;
    light_id: number;
    on: boolean;
    color: string;
    brightness_pct: number;
  }[],
): Promise<void> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/lights/state/batch`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ updates }),
    },
  );
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(
      j?.error?.message ?? `scene batch update failed (${res.status})`,
    );
  }
}

export async function fetchScene(id: string): Promise<SceneDetail> {
  const res = await fetch(`/api/v1/scenes/${encodeURIComponent(id)}`, {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`scene load failed (${res.status})`);
  }
  return res.json() as Promise<SceneDetail>;
}

/** REQ-014 / REQ-027 — reset every light in the scene to default state. */
export async function patchSceneLightsStateScene(
  sceneId: string,
  body: { on: boolean; color: string; brightness_pct: number },
): Promise<void> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/lights/state/scene`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(
      j?.error?.message ?? `scene lights reset failed (${res.status})`,
    );
  }
}

export async function createScene(body: {
  name: string;
  models: { model_id: string }[];
}): Promise<SceneSummary> {
  const res = await fetch("/api/v1/scenes", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `create scene failed (${res.status})`);
  }
  return res.json() as Promise<SceneSummary>;
}

export async function deleteScene(id: string): Promise<void> {
  const res = await fetch(`/api/v1/scenes/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (res.status !== 204) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `delete scene failed (${res.status})`);
  }
}

export async function addSceneModel(
  sceneId: string,
  body: { model_id: string; offset_x?: number; offset_y?: number; offset_z?: number },
): Promise<void> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/models`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
  if (res.status !== 201) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `add model failed (${res.status})`);
  }
}

export async function patchSceneModelOffsets(
  sceneId: string,
  modelId: string,
  offset_x: number,
  offset_y: number,
  offset_z: number,
): Promise<void> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/models/${encodeURIComponent(modelId)}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ offset_x, offset_y, offset_z }),
    },
  );
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `patch placement failed (${res.status})`);
  }
}

export async function removeSceneModel(
  sceneId: string,
  modelId: string,
): Promise<"ok" | "last_model"> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/models/${encodeURIComponent(modelId)}`,
    { method: "DELETE" },
  );
  if (res.status === 204) {
    return "ok";
  }
  if (res.status === 409) {
    const j = (await res.json().catch(() => null)) as {
      error?: { code?: string };
    };
    if (j?.error?.code === "scene_last_model") {
      return "last_model";
    }
  }
  const j = (await res.json().catch(() => null)) as {
    error?: { message?: string };
  };
  throw new Error(j?.error?.message ?? `remove model failed (${res.status})`);
}
