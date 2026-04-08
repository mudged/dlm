import type { SceneItem } from "@/lib/scenes";

export type SceneLightBatchPatch = {
  model_id: string;
  light_id: number;
  on: boolean;
  color: string;
  brightness_pct: number;
};

/** Apply batch light updates to scene items (scene-space coords unchanged). */
export function mergeSceneLightBatchIntoItems(
  items: SceneItem[],
  updates: SceneLightBatchPatch[],
): SceneItem[] {
  if (updates.length === 0) {
    return items;
  }
  const map = new Map<string, SceneLightBatchPatch>();
  for (const u of updates) {
    map.set(`${u.model_id}:${u.light_id}`, u);
  }
  return items.map((it) => ({
    ...it,
    lights: it.lights.map((L) => {
      const u = map.get(`${it.model_id}:${L.id}`);
      if (!u) {
        return L;
      }
      return {
        ...L,
        on: u.on,
        color: u.color,
        brightness_pct: u.brightness_pct,
      };
    }),
  }));
}
