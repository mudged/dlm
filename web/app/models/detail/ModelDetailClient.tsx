"use client";

import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Button } from "@/components/ui/Button";
import type { Light, ModelDetail } from "@/lib/models";

const PAGE_SIZE_OPTIONS = [25, 50, 100] as const;

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

function mergeBatchLightStatesApi(
  model: ModelDetail,
  states: { id: number; on: boolean; color: string; brightness_pct: number }[],
): ModelDetail {
  const byId = new Map(states.map((s) => [s.id, s]));
  return {
    ...model,
    lights: model.lights.map((L) => {
      const st = byId.get(L.id);
      if (!st) {
        return L;
      }
      return {
        ...L,
        on: st.on,
        color: st.color,
        brightness_pct: st.brightness_pct,
      };
    }),
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

function BulkLightStatePanel({
  modelId,
  model,
  selectedIds,
  onApplied,
}: {
  modelId: string;
  model: ModelDetail;
  selectedIds: number[];
  onApplied: (
    states: {
      id: number;
      on: boolean;
      color: string;
      brightness_pct: number;
    }[],
  ) => void;
}) {
  const [on, setOn] = useState(true);
  const [color, setColor] = useState("#ffffff");
  const [bp, setBp] = useState(100);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const seedKey = selectedIds.join(",");
  useEffect(() => {
    if (selectedIds.length === 0) {
      return;
    }
    const id = selectedIds[0];
    const L = model.lights.find((l) => l.id === id);
    if (L) {
      setOn(L.on);
      setColor(L.color);
      setBp(L.brightness_pct);
      setErr(null);
    }
  }, [seedKey, model.lights, selectedIds]);

  const apply = async () => {
    setSaving(true);
    setErr(null);
    try {
      const res = await fetch(
        `/api/v1/models/${encodeURIComponent(modelId)}/lights/state/batch`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            ids: selectedIds,
            on,
            color: color.trim(),
            brightness_pct: bp,
          }),
        },
      );
      const j = (await res.json().catch(() => null)) as {
        states?: {
          id: number;
          on: boolean;
          color: string;
          brightness_pct: number;
        }[];
        error?: { message?: string };
      };
      if (!res.ok) {
        setErr(j?.error?.message ?? `Bulk update failed (${res.status})`);
        return;
      }
      if (!j.states || !Array.isArray(j.states)) {
        setErr("Invalid response from server.");
        return;
      }
      onApplied(j.states);
    } catch {
      setErr("Could not reach the API.");
    } finally {
      setSaving(false);
    }
  };

  if (selectedIds.length === 0) {
    return null;
  }

  return (
    <div
      className="flex flex-col gap-2 rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-600 dark:bg-slate-800/50"
      role="region"
      aria-label="Bulk light settings"
    >
      <p className="text-xs font-medium text-slate-700 dark:text-slate-300">
        Apply to {selectedIds.length} selected light
        {selectedIds.length === 1 ? "" : "s"}
      </p>
      <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
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
          aria-label="Bulk color"
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
          onClick={() => void apply()}
        >
          {saving ? "…" : "Apply to selected"}
        </Button>
      </div>
      {err ? (
        <p className="text-xs text-red-600 dark:text-red-400" role="alert">
          {err}
        </p>
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

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] =
    useState<(typeof PAGE_SIZE_OPTIONS)[number]>(50);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [goToIdInput, setGoToIdInput] = useState("");
  const [goToIdErr, setGoToIdErr] = useState<string | null>(null);
  const shiftAnchorRef = useRef<number | null>(null);
  const headerSelectRef = useRef<HTMLInputElement>(null);

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

  // Reset list UI only when navigating to a different model, not on every model object refresh (e.g. PATCH).
  useEffect(() => {
    if (!model) {
      return;
    }
    setSelectedIds([]);
    setPage(1);
    setGoToIdInput("");
    setGoToIdErr(null);
    shiftAnchorRef.current = null;
    // eslint-disable-next-line react-hooks/exhaustive-deps -- model.id is the navigation key; full model updates in place
  }, [model?.id]);

  const lightCount = model?.lights.length ?? 0;
  const totalPages =
    lightCount > 0 ? Math.max(1, Math.ceil(lightCount / pageSize)) : 1;
  const pageClamped = Math.min(Math.max(1, page), totalPages);

  // Clamp page when result set size or page size changes (not on every model object identity change).
  useEffect(() => {
    if (!model || model.lights.length === 0) {
      return;
    }
    const tp = Math.max(1, Math.ceil(model.lights.length / pageSize));
    setPage((p) => Math.min(Math.max(1, p), tp));
  }, [model?.lights.length, pageSize]); // eslint-disable-line react-hooks/exhaustive-deps -- model from closure

  const pageLights = useMemo(() => {
    if (!model?.lights.length) {
      return [];
    }
    const start = (pageClamped - 1) * pageSize;
    return model.lights.slice(start, start + pageSize);
  }, [model, pageClamped, pageSize]);

  const pageIdList = useMemo(
    () => pageLights.map((L) => L.id),
    [pageLights],
  );

  useEffect(() => {
    const el = headerSelectRef.current;
    if (!el || pageIdList.length === 0) {
      return;
    }
    const c = pageIdList.filter((i) => selectedIds.includes(i)).length;
    el.indeterminate = c > 0 && c < pageIdList.length;
  }, [pageIdList, selectedIds]);

  const allOnPageSelected =
    pageIdList.length > 0 && pageIdList.every((i) => selectedIds.includes(i));

  const toggleSelectPage = () => {
    setSelectedIds((prev) => {
      const s = new Set(prev);
      if (allOnPageSelected) {
        pageIdList.forEach((i) => s.delete(i));
      } else {
        pageIdList.forEach((i) => s.add(i));
      }
      return [...s].sort((a, b) => a - b);
    });
  };

  const toggleRowSelected = (lightId: number, shiftKey: boolean) => {
    setSelectedIds((prev) => {
      if (shiftKey && shiftAnchorRef.current !== null) {
        const anchor = shiftAnchorRef.current;
        const lo = Math.min(anchor, lightId);
        const hi = Math.max(anchor, lightId);
        const s = new Set(prev);
        for (const pid of pageIdList) {
          if (pid >= lo && pid <= hi) {
            s.add(pid);
          }
        }
        shiftAnchorRef.current = lightId;
        return [...s].sort((a, b) => a - b);
      }
      const s = new Set(prev);
      if (s.has(lightId)) {
        s.delete(lightId);
      } else {
        s.add(lightId);
      }
      shiftAnchorRef.current = lightId;
      return [...s].sort((a, b) => a - b);
    });
  };

  const submitGoToId = () => {
    if (!model) {
      return;
    }
    const raw = goToIdInput.trim();
    if (!/^\d+$/.test(raw)) {
      setGoToIdErr("Enter a non-negative integer light id.");
      return;
    }
    const lid = Number(raw);
    const n = model.lights.length;
    if (n === 0 || lid < 0 || lid > n - 1) {
      setGoToIdErr(`Light id must be between 0 and ${Math.max(0, n - 1)}.`);
      return;
    }
    setGoToIdErr(null);
    const newPage = Math.floor(lid / pageSize) + 1;
    const tp = Math.max(1, Math.ceil(n / pageSize));
    setPage(Math.min(newPage, tp));
    requestAnimationFrame(() => {
      document
        .getElementById(`light-row-${lid}`)
        ?.scrollIntoView({ block: "nearest", behavior: "smooth" });
    });
  };

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
        <section className="space-y-3" aria-labelledby="lights-table-heading">
          <h2
            id="lights-table-heading"
            className="text-sm font-semibold text-slate-800 dark:text-slate-200"
          >
            Lights
          </h2>
          <BulkLightStatePanel
            modelId={model.id}
            model={model}
            selectedIds={selectedIds}
            onApplied={(states) =>
              setModel((m) =>
                m ? mergeBatchLightStatesApi(m, states) : m,
              )
            }
          />
          {lightCount > 1 ? (
            <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
              <label className="flex flex-col gap-0.5 text-xs text-slate-600 dark:text-slate-400">
                <span>Per page</span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(
                      Number(e.target.value) as (typeof PAGE_SIZE_OPTIONS)[number],
                    );
                    setPage(1);
                  }}
                  className="rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-600 dark:bg-slate-900"
                >
                  {PAGE_SIZE_OPTIONS.map((sz) => (
                    <option key={sz} value={sz}>
                      {sz}
                    </option>
                  ))}
                </select>
              </label>
              <div className="flex flex-wrap items-center gap-2">
                <Button
                  type="button"
                  className="h-8 px-2 py-0 text-xs"
                  disabled={pageClamped <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  Previous
                </Button>
                <span className="text-xs text-slate-600 tabular-nums dark:text-slate-400">
                  Page {pageClamped} of {totalPages}
                </span>
                <Button
                  type="button"
                  className="h-8 px-2 py-0 text-xs"
                  disabled={pageClamped >= totalPages}
                  onClick={() =>
                    setPage((p) => Math.min(totalPages, p + 1))
                  }
                >
                  Next
                </Button>
              </div>
              <form
                className="flex flex-col gap-1 sm:flex-row sm:items-end"
                onSubmit={(e) => {
                  e.preventDefault();
                  submitGoToId();
                }}
              >
                <label className="flex flex-col gap-0.5 text-xs text-slate-600 dark:text-slate-400">
                  <span>Go to light id</span>
                  <input
                    type="text"
                    inputMode="numeric"
                    value={goToIdInput}
                    onChange={(e) => {
                      setGoToIdInput(e.target.value);
                      setGoToIdErr(null);
                    }}
                    className="w-28 rounded border border-slate-300 bg-white px-2 py-1 font-mono text-sm dark:border-slate-600 dark:bg-slate-900"
                    aria-invalid={goToIdErr ? true : undefined}
                  />
                </label>
                <Button type="submit" className="h-8 px-2 py-0 text-xs sm:ml-1">
                  Go
                </Button>
              </form>
              {selectedIds.length > 0 ? (
                <Button
                  type="button"
                  className="h-8 min-h-0 min-w-0 bg-slate-200 px-2 py-0 text-xs text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 dark:hover:bg-slate-600"
                  onClick={() => {
                    setSelectedIds([]);
                    shiftAnchorRef.current = null;
                  }}
                >
                  Clear selection ({selectedIds.length})
                </Button>
              ) : null}
            </div>
          ) : null}
          {goToIdErr ? (
            <p className="text-xs text-red-600 dark:text-red-400" role="alert">
              {goToIdErr}
            </p>
          ) : null}
          <div className="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
            <table className="w-full min-w-[32rem] text-left text-sm">
              <thead className="bg-slate-100 dark:bg-slate-800/80">
                <tr>
                  <th className="w-10 px-2 py-2">
                    <span className="sr-only">Select</span>
                    {lightCount > 1 ? (
                      <input
                        ref={headerSelectRef}
                        type="checkbox"
                        className="rounded border-slate-300 dark:border-slate-600"
                        checked={allOnPageSelected}
                        onChange={() => toggleSelectPage()}
                        aria-label="Select all lights on this page"
                      />
                    ) : null}
                  </th>
                  <th className="px-3 py-2 font-medium">id</th>
                  <th className="px-3 py-2 font-medium">x</th>
                  <th className="px-3 py-2 font-medium">y</th>
                  <th className="px-3 py-2 font-medium">z</th>
                  <th className="px-3 py-2 font-medium">state</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 dark:divide-slate-700">
                {pageLights.map((L) => (
                  <tr
                    key={L.id}
                    id={`light-row-${L.id}`}
                    className="bg-white dark:bg-slate-900/30"
                  >
                    <td className="px-2 py-2">
                      <input
                        type="checkbox"
                        className="rounded border-slate-300 dark:border-slate-600"
                        checked={selectedIds.includes(L.id)}
                        onChange={(e) => {
                          const ne = e.nativeEvent as MouseEvent;
                          toggleRowSelected(L.id, ne.shiftKey === true);
                        }}
                        aria-label={`Select light ${L.id}`}
                      />
                    </td>
                    <td className="px-3 py-2 font-mono tabular-nums">{L.id}</td>
                    <td className="px-3 py-2 font-mono tabular-nums">{L.x}</td>
                    <td className="px-3 py-2 font-mono tabular-nums">{L.y}</td>
                    <td className="px-3 py-2 font-mono tabular-nums">{L.z}</td>
                    <td className="px-3 py-2">
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
        </section>
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
