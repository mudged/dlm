import { normalizeLightHex } from "@/lib/lightAppearance";
import type { Light } from "@/lib/models";
import type { SceneItem } from "@/lib/scenes";

function tripleKey(on: boolean, color: string, brightnessPct: number): string {
  const hex = normalizeLightHex(color);
  const br = Number.isFinite(brightnessPct) ? brightnessPct : 100;
  return `${on ? 1 : 0}|${hex}|${br}`;
}

/** Stable string for REQ-031: rebuild three.js only when positions or effective light state changes. */
export function modelLightsVizSignature(lights: Light[]): string {
  const sorted = [...lights].sort((a, b) => a.id - b.id);
  const pos = sorted
    .map((L) => `${L.id}:${L.x},${L.y},${L.z}`)
    .join(";");
  const state = sorted
    .map((L) => {
      const on = L.on !== false;
      const c = L.color ?? "#ffffff";
      const br =
        typeof L.brightness_pct === "number" && Number.isFinite(L.brightness_pct)
          ? L.brightness_pct
          : 100;
      return `${L.id}:${tripleKey(on, c, br)}`;
    })
    .join(";");
  return `${pos}|${state}`;
}

function sceneSpaceLight(L: SceneItem["lights"][number]): Light {
  return {
    id: L.id,
    x: L.sx,
    y: L.sy,
    z: L.sz,
    on: L.on,
    color: L.color,
    brightness_pct: L.brightness_pct,
  };
}

export function sceneItemsVizSignature(items: SceneItem[]): string {
  const sortedItems = [...items].sort((a, b) => a.model_id.localeCompare(b.model_id));
  return sortedItems
    .map((it) => {
      const lights = it.lights.map(sceneSpaceLight);
      return `${it.model_id}:${modelLightsVizSignature(lights)}`;
    })
    .join("||");
}
