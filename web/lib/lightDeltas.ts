import type { Light } from "@/lib/models";
import type { SceneItem } from "@/lib/scenes";

/** Payload on each SSE `data:` line (REQ-041). */
export type LightsSSEMessage = {
  seq: number;
  deltas?: Array<{
    model_id?: string;
    light_id: number;
    on: boolean;
    color: string;
    brightness_pct: number;
  }>;
};

export function parseLightsSSEMessage(raw: string): LightsSSEMessage | null {
  try {
    const j = JSON.parse(raw) as LightsSSEMessage;
    if (typeof j.seq !== "number" || !Number.isFinite(j.seq)) {
      return null;
    }
    return j;
  } catch {
    return null;
  }
}

export function applyModelLightDeltas<L extends Light>(lights: L[], deltas: LightsSSEMessage["deltas"]): L[] {
  if (!deltas || deltas.length === 0) {
    return lights;
  }
  const byId = new Map(deltas.map((d) => [d.light_id, d]));
  return lights.map((L) => {
    const d = byId.get(L.id);
    if (!d) {
      return L;
    }
    return {
      ...L,
      on: d.on,
      color: d.color,
      brightness_pct: d.brightness_pct,
    };
  });
}

type SceneLightDelta = NonNullable<LightsSSEMessage["deltas"]>[number];

/** Resolve model id / light id (snake_case from Go; tolerate camelCase). */
function sceneDeltaKeys(d: SceneLightDelta): { modelId: string; lightId: number } | null {
  const raw = d as SceneLightDelta & {
    modelId?: string;
    lightId?: number;
  };
  const mid = raw.model_id ?? raw.modelId ?? "";
  const lid = raw.light_id ?? raw.lightId;
  if (!mid || typeof lid !== "number" || !Number.isFinite(lid)) {
    return null;
  }
  return { modelId: mid, lightId: lid };
}

export function applySceneLightDeltas(
  items: SceneItem[],
  deltas: NonNullable<LightsSSEMessage["deltas"]>,
): SceneItem[] {
  if (deltas.length === 0) {
    return items;
  }
  const byModel = new Map<string, Map<number, SceneLightDelta>>();
  for (const d of deltas) {
    const keys = sceneDeltaKeys(d);
    if (!keys) {
      continue;
    }
    let m = byModel.get(keys.modelId);
    if (!m) {
      m = new Map();
      byModel.set(keys.modelId, m);
    }
    m.set(keys.lightId, d);
  }
  return items.map((it) => {
    const m = byModel.get(it.model_id);
    if (!m || m.size === 0) {
      return it;
    }
    return {
      ...it,
      lights: it.lights.map((L) => {
        const d = m!.get(L.id);
        if (!d) {
          return L;
        }
        return {
          ...L,
          on: d.on,
          color: d.color,
          brightness_pct: d.brightness_pct,
        };
      }),
    };
  });
}
