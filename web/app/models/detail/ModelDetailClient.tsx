"use client";

import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import type { Light, ModelDetail } from "@/lib/models";

function mergeLightStateApi(
  model: ModelDetail,
  lightId: number,
  st: { on: boolean; color: string; brightness_pct: number },
): ModelDetail {
  return {
    ...model,
    lights: model.lights.map((L) =>
      L.id === lightId
        ? { ...L, on: st.on, color: st.color, brightness_pct: st.brightness_pct }
        : L,
    ),
  };
}

function LightStateEditor({
  modelId,
  light,
  onApplied,
}: {
  modelId: string;
  light: Light;
  onApplied: (st: {
    on: boolean;
    color: string;
    brightness_pct: number;
  }) => void;
}) {
  const [on, setOn] = useState(light.on);
  const [color, setColor] = useState(light.color);
  const [bp, setBp] = useState(light.brightness_pct);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setOn(light.on);
    setColor(light.color);
    setBp(light.brightness_pct);
    setErr(null);
  }, [light.id, light.on, light.color, light.brightness_pct]);

  const save = async () => {
    const body: Record<string, unknown> = {};
    if (on !== light.on) {
      body.on = on;
    }
    if (color.trim() !== light.color) {
      body.color = color.trim();
    }
    if (bp !== light.brightness_pct) {
      body.brightness_pct = bp;
    }
    if (Object.keys(body).length === 0) {
      return;
    }
    setSaving(true);
    setErr(null);
    try {
      const res = await fetch(
        `/api/v1/models/${encodeURIComponent(modelId)}/lights/${encodeURIComponent(String(light.id))}/state`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        },
      );
      const j = (await res.json().catch(() => null)) as {
        on?: boolean;
        color?: string;
        brightness_pct?: number;
        error?: { message?: string };
      };
      if (!res.ok) {
        setErr(j?.error?.message ?? `Update failed (${res.status})`);
        return;
      }
      if (
        j.on === undefined ||
        j.color === undefined ||
        j.brightness_pct === undefined
      ) {
        setErr("Invalid response from server.");
        return;
      }
      onApplied({
        on: j.on,
        color: j.color,
        brightness_pct: j.brightness_pct,
      });
    } catch {
      setErr("Could not reach the API.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:flex-wrap sm:items-center sm:gap-2">
      <label className="flex cursor-pointer items-center gap-1.5 text-xs">
        <input
          type="checkbox"
          className="rounded border-slate-300 dark:border-slate-600"
          checked={on}
          onChange={(e) => setOn(e.target.checked)}
        />
        on
      </label>
      <input
        type="color"
        value={color.length === 7 ? color : "#ffffff"}
        onChange={(e) => setColor(e.target.value)}
        className="h-8 w-12 cursor-pointer rounded border border-slate-300 bg-white dark:border-slate-600"
        aria-label={`Color for light ${light.id}`}
      />
      <label className="flex items-center gap-1 text-xs">
        <span className="text-slate-500 dark:text-slate-400">%</span>
        <input
          type="number"
          min={0}
          max={100}
          step={1}
          value={bp}
          onChange={(e) => setBp(Number(e.target.value))}
          className="w-14 rounded border border-slate-300 bg-white px-1 py-0.5 font-mono text-xs dark:border-slate-600 dark:bg-slate-900"
        />
      </label>
      <Button
        type="button"
        className="h-8 px-2 py-0 text-xs"
        disabled={saving}
        onClick={() => void save()}
      >
        {saving ? "…" : "Apply"}
      </Button>
      {err ? (
        <span className="text-xs text-red-600 dark:text-red-400">{err}</span>
      ) : null}
    </div>
  );
}

const ModelLightsCanvas = dynamic(() => import("@/components/ModelLightsCanvas"), {
    ssr: false,
    loading: () => (
      <div className="flex h-[min(50vh,24rem)] min-h-[240px] w-full items-center justify-center rounded-xl border border-slate-200 bg-slate-50 text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900/50 dark:text-slate-400">
        Preparing 3D view…
      </div>
    ),
  },
);

export function ModelDetailClient() {
  const router = useRouter();
  const params = useSearchParams();
  const id = params.get("id");
  const [model, setModel] = useState<ModelDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) {
      setError("Missing model id.");
      setLoading(false);
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/models/${encodeURIComponent(id)}`, {
        cache: "no-store",
      });
      const j = (await res.json().catch(() => null)) as ModelDetail & {
        error?: { message?: string };
      };
      if (!res.ok) {
        setError(j?.error?.message ?? `Could not load model (${res.status})`);
        setModel(null);
        setLoading(false);
        return;
      }
      setModel(j as ModelDetail);
    } catch {
      setError("Could not reach the API.");
      setModel(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  if (!id) {
    return (
      <p className="text-sm text-amber-800 dark:text-amber-200" role="alert">
        Missing model id. Return to the list and open a model again.
      </p>
    );
  }

  if (loading) {
    return <p className="text-sm text-slate-500">Loading…</p>;
  }

  if (error || !model) {
    return (
      <p
        className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
        role="alert"
      >
        {error ?? "Model not found."}
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {model.name}
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          {model.light_count} lights ·{" "}
          {new Date(model.created_at).toLocaleString()}
        </p>
      </header>

      <section className="space-y-2" aria-labelledby="model-3d-heading">
        <h2
          id="model-3d-heading"
          className="text-sm font-semibold text-slate-800 dark:text-slate-200"
        >
          3D view
        </h2>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Drag to rotate; scroll or pinch to zoom. Hover a sphere for id and
          coordinates (desktop). Tap a sphere to pin the label; tap empty space
          to clear (touch and mouse).
        </p>
        <ModelLightsCanvas lights={model.lights} />
      </section>

      {model.lights.length === 0 ? (
        <p className="text-sm text-slate-600 dark:text-slate-400">
          This model has no lights (header-only CSV).
        </p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
          <table className="w-full min-w-[28rem] text-left text-sm">
            <thead className="bg-slate-100 dark:bg-slate-800/80">
              <tr>
                <th className="px-3 py-2 font-medium">id</th>
                <th className="px-3 py-2 font-medium">x</th>
                <th className="px-3 py-2 font-medium">y</th>
                <th className="px-3 py-2 font-medium">z</th>
                <th className="px-3 py-2 font-medium">state</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200 dark:divide-slate-700">
              {model.lights.map((L) => (
                <tr key={L.id} className="bg-white dark:bg-slate-900/30">
                  <td className="px-3 py-2 font-mono tabular-nums">{L.id}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.x}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.y}</td>
                  <td className="px-3 py-2 font-mono tabular-nums">{L.z}</td>
                  <td className="px-3 py-2" colSpan={1}>
                    <LightStateEditor
                      modelId={model.id}
                      light={L}
                      onApplied={(st) =>
                        setModel((m) =>
                          m ? mergeLightStateApi(m, L.id, st) : m,
                        )
                      }
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="flex flex-col gap-2 sm:flex-row">
        <Button
          type="button"
          className="w-full bg-slate-600 dark:bg-slate-600 sm:w-auto"
          onClick={() => router.push("/models")}
        >
          Back to list
        </Button>
      </div>
    </div>
  );
}
