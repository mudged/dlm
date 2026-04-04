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

export async function fetchScenes(): Promise<SceneSummary[]> {
  const res = await fetch("/api/v1/scenes", { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`scenes list failed (${res.status})`);
  }
  return res.json() as Promise<SceneSummary[]>;
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
