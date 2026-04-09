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

export function applySceneLightDeltas(
  items: SceneItem[],
  deltas: NonNullable<LightsSSEMessage["deltas"]>,
): SceneItem[] {
  if (deltas.length === 0) {
    return items;
  }
  const byModel = new Map<string, Map<number, (typeof deltas)[number]>>();
  for (const d of deltas) {
    const mid = d.model_id ?? "";
    if (!mid) {
      continue;
    }
    let m = byModel.get(mid);
    if (!m) {
      m = new Map();
      byModel.set(mid, m);
    }
    m.set(d.light_id, d);
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
