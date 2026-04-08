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

const inputCls =
  "min-h-9 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-900";
const labelCls = "text-slate-600 dark:text-slate-400";
const selectCls = `${inputCls} w-full`;

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
        <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <label className="flex flex-col gap-1 text-xs">
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
              <label className="flex flex-col gap-1 text-xs">
                <span className={labelCls}>Colour (hex)</span>
                <input
                  className={inputCls}
                  disabled={disabled}
                  value={state.backgroundColor}
                  onChange={(e) => onChange({ ...state, backgroundColor: e.target.value })}
                />
              </label>
              <label className="flex flex-col gap-1 text-xs">
                <span className={labelCls}>Brightness (%)</span>
                <input
                  type="number"
                  min={0}
                  max={100}
                  className={inputCls}
                  disabled={disabled}
                  value={state.backgroundBrightness_pct}
                  onChange={(e) =>
                    onChange({ ...state, backgroundBrightness_pct: e.target.value })
                  }
                />
              </label>
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

          <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <label className="flex flex-col gap-1 text-xs">
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
            <label className="flex flex-col gap-1 text-xs">
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
            <label className="flex flex-col gap-1 text-xs">
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
            <label className="flex flex-col gap-1 text-xs">
              <span className={labelCls}>Brightness (%)</span>
              <input
                type="number"
                min={0}
                max={100}
                className={inputCls}
                disabled={disabled}
                value={row.brightness_pct}
                onChange={(e) => updateRow(i, { brightness_pct: e.target.value })}
              />
            </label>
          </div>

          {row.colorMode === "fixed" ? (
            <label className="mt-3 flex max-w-xs flex-col gap-1 text-xs">
              <span className={labelCls}>Shape colour (hex)</span>
              <input
                className={inputCls}
                disabled={disabled}
                value={row.color}
                onChange={(e) => updateRow(i, { color: e.target.value })}
              />
            </label>
          ) : null}

          <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {row.kind === "sphere" ? (
              row.sizeMode === "fixed" ? (
                <label className="flex flex-col gap-1 text-xs sm:col-span-2">
                  <span className={labelCls}>Radius (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.radius_m}
                    onChange={(e) => updateRow(i, { radius_m: e.target.value })}
                  />
                </label>
              ) : (
                <>
                  <label className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>Min radius (m)</span>
                    <input
                      className={inputCls}
                      disabled={disabled}
                      value={row.radius_min_m}
                      onChange={(e) => updateRow(i, { radius_min_m: e.target.value })}
                    />
                  </label>
                  <label className="flex flex-col gap-1 text-xs">
                    <span className={labelCls}>Max radius (m)</span>
                    <input
                      className={inputCls}
                      disabled={disabled}
                      value={row.radius_max_m}
                      onChange={(e) => updateRow(i, { radius_max_m: e.target.value })}
                    />
                  </label>
                </>
              )
            ) : row.sizeMode === "fixed" ? (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Width (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.width_m}
                    onChange={(e) => updateRow(i, { width_m: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Height (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.height_m}
                    onChange={(e) => updateRow(i, { height_m: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Depth (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.depth_m}
                    onChange={(e) => updateRow(i, { depth_m: e.target.value })}
                  />
                </label>
              </>
            ) : (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>W min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.width_min_m}
                      onChange={(e) => updateRow(i, { width_min_m: e.target.value })}
                    />
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.width_max_m}
                      onChange={(e) => updateRow(i, { width_max_m: e.target.value })}
                    />
                  </div>
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>H min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.height_min_m}
                      onChange={(e) => updateRow(i, { height_min_m: e.target.value })}
                    />
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.height_max_m}
                      onChange={(e) => updateRow(i, { height_max_m: e.target.value })}
                    />
                  </div>
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>D min / max (m)</span>
                  <div className="flex gap-1">
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.depth_min_m}
                      onChange={(e) => updateRow(i, { depth_min_m: e.target.value })}
                    />
                    <input
                      className={`${inputCls} flex-1`}
                      disabled={disabled}
                      value={row.depth_max_m}
                      onChange={(e) => updateRow(i, { depth_max_m: e.target.value })}
                    />
                  </div>
                </label>
              </>
            )}
          </div>

          <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <label className="flex flex-col gap-1 text-xs">
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
              <label className="flex flex-col gap-1 text-xs">
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
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Center X (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.center_x}
                    onChange={(e) => updateRow(i, { center_x: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Center Y (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.center_y}
                    onChange={(e) => updateRow(i, { center_y: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Center Z (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.center_z}
                    onChange={(e) => updateRow(i, { center_z: e.target.value })}
                  />
                </label>
              </>
            ) : (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Min corner X (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.min_corner_x}
                    onChange={(e) => updateRow(i, { min_corner_x: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Min corner Y (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.min_corner_y}
                    onChange={(e) => updateRow(i, { min_corner_y: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Min corner Z (m)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.min_corner_z}
                    onChange={(e) => updateRow(i, { min_corner_z: e.target.value })}
                  />
                </label>
              </>
            )}
          </div>

          <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <label className="flex flex-col gap-1 text-xs">
              <span className={labelCls}>Direction dx</span>
              <input
                className={inputCls}
                disabled={disabled}
                value={row.motion_dx}
                onChange={(e) => updateRow(i, { motion_dx: e.target.value })}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs">
              <span className={labelCls}>Direction dy</span>
              <input
                className={inputCls}
                disabled={disabled}
                value={row.motion_dy}
                onChange={(e) => updateRow(i, { motion_dy: e.target.value })}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs">
              <span className={labelCls}>Direction dz</span>
              <input
                className={inputCls}
                disabled={disabled}
                value={row.motion_dz}
                onChange={(e) => updateRow(i, { motion_dz: e.target.value })}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs">
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
          <div className="mt-2 grid gap-3 sm:grid-cols-2">
            {row.speedMode === "fixed" ? (
              <label className="flex flex-col gap-1 text-xs">
                <span className={labelCls}>Speed (m/s)</span>
                <input
                  className={inputCls}
                  disabled={disabled}
                  value={row.m_s}
                  onChange={(e) => updateRow(i, { m_s: e.target.value })}
                />
                <span className="text-[0.65rem] text-slate-500">
                  ≈ {(Number(row.m_s) || 0) * 100} cm/s (when value is numeric)
                </span>
              </label>
            ) : (
              <>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Min speed (m/s)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.min_m_s}
                    onChange={(e) => updateRow(i, { min_m_s: e.target.value })}
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs">
                  <span className={labelCls}>Max speed (m/s)</span>
                  <input
                    className={inputCls}
                    disabled={disabled}
                    value={row.max_m_s}
                    onChange={(e) => updateRow(i, { max_m_s: e.target.value })}
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
