"use client";

import { faPlus, faTrash } from "@fortawesome/free-solid-svg-icons";
import { Button } from "@/components/ui/Button";
import {
  defaultShapeFormRow,
  type EdgeBehavior,
  type FaceId,
  type ShapeAnimationFormState,
  type ShapeFormRow,
} from "@/lib/shapeAnimationDefinitionForm";

const EDGE_OPTIONS: { value: EdgeBehavior; label: string }[] = [
  { value: "wrap", label: "Pac-Man (wrap)" },
  { value: "stop", label: "Stop and disappear" },
  { value: "deflect_random", label: "Deflect random angle" },
  { value: "deflect_specular", label: "Deflect (specular)" },
];

const FACES: FaceId[] = ["top", "bottom", "left", "right", "back", "front"];

type Props = {
  state: ShapeAnimationFormState;
  onChange: (next: ShapeAnimationFormState) => void;
  disabled?: boolean;
};

/** Valid #rrggbb for <input type="color"> (fallback if hex invalid). */
function colorPickerValue(hex: string): string {
  const t = hex.trim();
  if (/^#[0-9A-Fa-f]{6}$/.test(t)) {
    return t.toLowerCase();
  }
  if (/^[0-9A-Fa-f]{6}$/.test(t)) {
    return `#${t.toLowerCase()}`;
  }
  return "#888888";
}

const labelCls = "text-slate-600 dark:text-slate-400";
const selectCls =
  "min-h-9 w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-900";

function ColorField(props: {
  label: string;
  value: string;
  onChange: (hex: string) => void;
  disabled?: boolean;
}) {
  const { label, value, onChange, disabled } = props;
  const pickerVal = colorPickerValue(value);
  return (
    <div className="flex flex-col gap-1 text-xs">
      <span className={labelCls}>{label}</span>
      <div className="flex flex-wrap items-center gap-2">
        <input
          type="color"
          aria-label={`${label} picker`}
          disabled={disabled}
          value={pickerVal}
          onChange={(e) => onChange(e.target.value)}
          className="h-9 w-14 cursor-pointer rounded border border-slate-300 bg-white p-0.5 dark:border-slate-600"
        />
        <input
          type="text"
          spellCheck={false}
          disabled={disabled}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          maxLength={7}
          className="h-9 w-[5.5rem] rounded border border-slate-300 bg-white px-1.5 font-mono text-sm dark:border-slate-600 dark:bg-slate-900"
        />
      </div>
    </div>
  );
}

function BrightnessField(props: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
}) {
  return (
    <label className="flex flex-col gap-1 text-xs">
      <span className={labelCls}>{props.label}</span>
      <input
        type="number"
        inputMode="numeric"
        min={0}
        max={100}
        step={1}
        size={3}
        maxLength={3}
        disabled={props.disabled}
        value={props.value}
        onChange={(e) => props.onChange(e.target.value)}
        className="h-9 w-[3.25rem] rounded border border-slate-300 bg-white px-1 text-center text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
      />
    </label>
  );
}

export function ShapeRoutineDefinitionForm({ state, onChange, disabled }: Props) {
  function setShapes(shapes: ShapeFormRow[]) {
    onChange({ ...state, shapes });
  }

  function updateRow(i: number, patch: Partial<ShapeFormRow>) {
    const next = state.shapes.map((r, j) => (j === i ? { ...r, ...patch } : r));
    setShapes(next);
  }

  return (
    <div className="space-y-6">
      <section className="rounded-lg border border-slate-200 p-3 dark:border-slate-600">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-700 dark:text-slate-300">
          Background
        </h3>
        <div className="mt-3 flex flex-wrap gap-4">
          <label className="flex min-w-[12rem] flex-col gap-1 text-xs">
            <span className={labelCls}>Mode</span>
            <select
              className={selectCls}
              disabled={disabled}
              value={state.backgroundMode}
              onChange={(e) =>
                onChange({
                  ...state,
                  backgroundMode: e.target.value === "lights_off" ? "lights_off" : "lights_on",
                })
              }
            >
              <option value="lights_on">Lights on (colour + brightness)</option>
              <option value="lights_off">Lights off outside shapes</option>
            </select>
          </label>
          {state.backgroundMode === "lights_on" ? (
            <>
              <ColorField
                label="Background colour"
                value={state.backgroundColor}
                disabled={disabled}
                onChange={(hex) => onChange({ ...state, backgroundColor: hex })}
              />
              <BrightnessField
                label="Brightness (%)"
                value={state.backgroundBrightness_pct}
                disabled={disabled}
                onChange={(v) => onChange({ ...state, backgroundBrightness_pct: v })}
              />
            </>
          ) : null}
        </div>
      </section>

      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-700 dark:text-slate-300">
          Shapes ({state.shapes.length} / 20)
        </h3>
        <Button
          type="button"
          icon={faPlus}
          disabled={disabled || state.shapes.length >= 20}
          className="bg-slate-200 text-slate-900 hover:bg-slate-300 dark:bg-slate-700 dark:text-slate-100 dark:hover:bg-slate-600"
          onClick={() => setShapes([...state.shapes, defaultShapeFormRow()])}
        >
          Add shape
        </Button>
      </div>

      {state.shapes.map((row, i) => (
        <section
          key={row.clientKey}
          className="rounded-lg border border-slate-200 p-3 dark:border-slate-600"
        >
          <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 pb-2 dark:border-slate-700">
            <span className="text-sm font-medium text-slate-800 dark:text-slate-200">
              Shape {i + 1}
            </span>
            <Button
              type="button"
              icon={faTrash}
              disabled={disabled || state.shapes.length <= 1}
              className="bg-red-900 hover:bg-red-800 dark:bg-red-950"
              onClick={() => setShapes(state.shapes.filter((_, j) => j !== i))}
            >
              Remove
            </Button>
          </div>

          <div className="mt-3 flex flex-wrap gap-3">
            <label className="flex min-w-[8rem] flex-col gap-1 text-xs">
              <span className={labelCls}>Kind</span>
              <select
                className={selectCls}
                disabled={disabled}
                value={row.kind}
                onChange={(e) =>
                  updateRow(i, { kind: e.target.value === "cuboid" ? "cuboid" : "sphere" })
                }
              >
                <option value="sphere">Sphere</option>
                <option value="cuboid">Cuboid</option>
              </select>
            </label>
            <label className="flex min-w-[10rem] flex-col gap-1 text-xs">
              <span className={labelCls}>Size mode</span>
              <select
                className={selectCls}
                disabled={disabled}
                value={row.sizeMode}
                onChange={(e) =>
                  updateRow(i, {
                    sizeMode: e.target.value === "random_uniform" ? "random_uniform" : "fixed",
                  })
                }
              >
                <option value="fixed">Fixed</option>
                <option value="random_uniform">Random (uniform)</option>
              </select>
            </label>
            <label className="flex min-w-[8rem] flex-col gap-1 text-xs">
              <span className={labelCls}>Colour</span>
              <select
                className={selectCls}
                disabled={disabled}
                value={row.colorMode}
                onChange={(e) =>
                  updateRow(i, { colorMode: e.target.value === "random" ? "random" : "fixed" })
                }
              >
                <option value="fixed">Fixed</option>
                <option value="random">Random</option>
              </select>
            </label>
            <BrightnessField
              label="Brightness (%)"
              value={row.brightness_pct}
              disabled={disabled}
              onChange={(v) => updateRow(i, { brightness_pct: v })}
            />
          </div>

          {row.colorMode === "fixed" ? (
            <div className="mt-3">
              <ColorField
                label="Shape colour"
                value={row.color}
                disabled={disabled}
                onChange={(hex) => updateRow(i, { color: hex })}
              />
            </div>
          ) : null}

          <div className="mt-3 flex flex-wrap gap-3">
            {row.kind === "sphere" ? (
              row.sizeMode === "fixed" ? (
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Radius (m)</span>
                  <input
                    type="number"
                    inputMode="decimal"
                    step="any"
                    disabled={disabled}
                    value={row.radius_m}
                    onChange={(e) => updateRow(i, { radius_m: e.target.value })}
                    className="h-9 w-20 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                  />
                </label>
              ) : (
                <>
                  <label className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>Min r (m)</span>
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.radius_min_m}
                      onChange={(e) => updateRow(i, { radius_min_m: e.target.value })}
                      className="h-9 w-20 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </label>
                  <label className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>Max r (m)</span>
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.radius_max_m}
                      onChange={(e) => updateRow(i, { radius_max_m: e.target.value })}
                      className="h-9 w-20 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </label>
                </>
              )
            ) : row.sizeMode === "fixed" ? (
              <>
                {(
                  [
                    ["width_m", "Width (m)", row.width_m],
                    ["height_m", "Height (m)", row.height_m],
                    ["depth_m", "Depth (m)", row.depth_m],
                  ] as const
                ).map(([key, lab, val]) => (
                  <label key={key} className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>{lab}</span>
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={val}
                      onChange={(e) => updateRow(i, { [key]: e.target.value })}
                      className="h-9 w-20 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </label>
                ))}
              </>
            ) : (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>W min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.width_min_m}
                      onChange={(e) => updateRow(i, { width_min_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.width_max_m}
                      onChange={(e) => updateRow(i, { width_max_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </div>
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>H min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.height_min_m}
                      onChange={(e) => updateRow(i, { height_min_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.height_max_m}
                      onChange={(e) => updateRow(i, { height_max_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </div>
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>D min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.depth_min_m}
                      onChange={(e) => updateRow(i, { depth_min_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={row.depth_max_m}
                      onChange={(e) => updateRow(i, { depth_max_m: e.target.value })}
                      className="h-9 w-[4.25rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </div>
                </label>
              </>
            )}
          </div>

          <div className="mt-3 flex flex-wrap gap-3">
            <label className="flex min-w-[12rem] flex-col gap-1 text-xs">
              <span className={labelCls}>Placement</span>
              <select
                className={selectCls}
                disabled={disabled}
                value={row.placementMode}
                onChange={(e) =>
                  updateRow(i, {
                    placementMode: e.target.value === "random_face" ? "random_face" : "fixed",
                  })
                }
              >
                <option value="fixed">Fixed coordinates</option>
                <option value="random_face">Random on scene face</option>
              </select>
            </label>
            {row.placementMode === "random_face" ? (
              <label className="flex min-w-[8rem] flex-col gap-1 text-xs">
                <span className={labelCls}>Face</span>
                <select
                  className={selectCls}
                  disabled={disabled}
                  value={row.face}
                  onChange={(e) => updateRow(i, { face: e.target.value as FaceId })}
                >
                  {FACES.map((f) => (
                    <option key={f} value={f}>
                      {f}
                    </option>
                  ))}
                </select>
              </label>
            ) : row.kind === "sphere" ? (
              <>
                {(
                  [
                    ["center_x", "Center X (m)", row.center_x],
                    ["center_y", "Center Y (m)", row.center_y],
                    ["center_z", "Center Z (m)", row.center_z],
                  ] as const
                ).map(([key, lab, val]) => (
                  <label key={key} className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>{lab}</span>
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={val}
                      onChange={(e) => updateRow(i, { [key]: e.target.value })}
                      className="h-9 w-[4.5rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </label>
                ))}
              </>
            ) : (
              <>
                {(
                  [
                    ["min_corner_x", "Min corner X (m)", row.min_corner_x],
                    ["min_corner_y", "Min corner Y (m)", row.min_corner_y],
                    ["min_corner_z", "Min corner Z (m)", row.min_corner_z],
                  ] as const
                ).map(([key, lab, val]) => (
                  <label key={key} className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>{lab}</span>
                    <input
                      type="number"
                      inputMode="decimal"
                      step="any"
                      disabled={disabled}
                      value={val}
                      onChange={(e) => updateRow(i, { [key]: e.target.value })}
                      className="h-9 w-[4.5rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                    />
                  </label>
                ))}
              </>
            )}
          </div>

          <div className="mt-3 flex flex-wrap gap-3">
            {(
              [
                ["motion_dx", "dx", row.motion_dx],
                ["motion_dy", "dy", row.motion_dy],
                ["motion_dz", "dz", row.motion_dz],
              ] as const
            ).map(([key, lab, val]) => (
              <label key={key} className="flex flex-col gap-1 text-xs">
                <span className={labelCls}>Direction {lab}</span>
                <input
                  type="number"
                  inputMode="decimal"
                  step="any"
                  disabled={disabled}
                  value={val}
                  onChange={(e) => updateRow(i, { [key]: e.target.value })}
                  className="h-9 w-[4.5rem] rounded border border-slate-300 bg-white px-1 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                />
              </label>
            ))}
            <label className="flex min-w-[10rem] flex-col gap-1 text-xs">
              <span className={labelCls}>Speed mode</span>
              <select
                className={selectCls}
                disabled={disabled}
                value={row.speedMode}
                onChange={(e) =>
                  updateRow(i, {
                    speedMode: e.target.value === "random_uniform" ? "random_uniform" : "fixed",
                  })
                }
              >
                <option value="fixed">Fixed (m/s)</option>
                <option value="random_uniform">Random (m/s)</option>
              </select>
            </label>
          </div>
          <div className="mt-2 flex flex-wrap gap-3">
            {row.speedMode === "fixed" ? (
              <label className="flex flex-col gap-1 text-xs">
                <span className={labelCls}>Speed (m/s)</span>
                <input
                  type="number"
                  inputMode="decimal"
                  step="any"
                  disabled={disabled}
                  value={row.m_s}
                  onChange={(e) => updateRow(i, { m_s: e.target.value })}
                  className="h-9 w-24 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                />
                <span className="text-[0.65rem] text-slate-500">
                  ≈ {(Number(row.m_s) || 0) * 100} cm/s (when numeric)
                </span>
              </label>
            ) : (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Min speed (m/s)</span>
                  <input
                    type="number"
                    inputMode="decimal"
                    step="any"
                    disabled={disabled}
                    value={row.min_m_s}
                    onChange={(e) => updateRow(i, { min_m_s: e.target.value })}
                    className="h-9 w-24 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Max speed (m/s)</span>
                  <input
                    type="number"
                    inputMode="decimal"
                    step="any"
                    disabled={disabled}
                    value={row.max_m_s}
                    onChange={(e) => updateRow(i, { max_m_s: e.target.value })}
                    className="h-9 w-24 rounded border border-slate-300 bg-white px-1.5 text-sm tabular-nums dark:border-slate-600 dark:bg-slate-900"
                  />
                </label>
              </>
            )}
          </div>

          <label className="mt-3 flex max-w-md flex-col gap-1 text-xs">
            <span className={labelCls}>When shape hits scene edge</span>
            <select
              className={selectCls}
              disabled={disabled}
              value={row.edge_behavior}
              onChange={(e) => updateRow(i, { edge_behavior: e.target.value as EdgeBehavior })}
            >
              {EDGE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </label>
        </section>
      ))}
    </div>
  );
}
