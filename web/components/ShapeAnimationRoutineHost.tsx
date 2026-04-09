"use client";

import { useEffect, useRef } from "react";
import {
  SHAPE_ANIMATION_DT_SEC,
  buildBatchUpdatesFromSim,
  initShapeAnimationSim,
  makeRng,
  sceneDimensionsFromApiResponse,
  tickShapeAnimationSim,
  type BatchLightUpdate,
  type SceneDimensions,
  type SceneLightFlat,
  type ShapeAnimationSim,
} from "@/lib/shapeAnimationEngine";
import {
  fetchSceneDimensions,
  fetchSceneLightsFlat,
  patchSceneLightsStateBatch,
} from "@/lib/scenes";
import { stopSceneRoutineRun } from "@/lib/routines";

/**
 * Runs shape animation in the browser while a scene routine is active (REQ-033).
 */
export function ShapeAnimationRoutineHost(props: {
  sceneId: string;
  runId: string;
  definitionJson: string;
  onSceneRefresh?: () => void;
  onError?: (message: string) => void;
  onStopped?: () => void;
  /** Called after each physics tick with current sim (e.g. editor ghost overlays). */
  onSimTick?: (sim: ShapeAnimationSim) => void;
  /**
   * Called synchronously each tick with the same updates sent to the API so parents
   * can merge into local scene state for smooth 3D updates without waiting for refetch.
   */
  onLightsPreview?: (updates: BatchLightUpdate[]) => void;
  /** Only refetch full scene from API every N ticks (reduces jank). Default 3. */
  sceneRefreshStride?: number;
}) {
  const {
    sceneId,
    runId,
    definitionJson,
    onSceneRefresh,
    onError,
    onStopped,
    onSimTick,
    onLightsPreview,
    sceneRefreshStride = 3,
  } = props;
  const stoppedRef = useRef(false);

  useEffect(() => {
    stoppedRef.current = false;
    let timer: ReturnType<typeof setInterval> | null = null;

    const rng = makeRng(Date.now() % 0xffffffff);

    void (async () => {
      let dims: SceneDimensions;
      try {
        const d = await fetchSceneDimensions(sceneId);
        dims = sceneDimensionsFromApiResponse(d);
      } catch (e) {
        onError?.(e instanceof Error ? e.message : "Could not load scene size");
        return;
      }

      let sim: ReturnType<typeof initShapeAnimationSim>;
      try {
        sim = initShapeAnimationSim(definitionJson, dims, rng);
      } catch (e) {
        onError?.(e instanceof Error ? e.message : "Bad shape routine JSON");
        return;
      }

      const tickMs = Math.round(SHAPE_ANIMATION_DT_SEC * 1000);
      let tickCount = 0;
      let lightsCache: SceneLightFlat[] | null = null;

      const runTick = () => {
        if (stoppedRef.current) {
          return;
        }
        void (async () => {
          if (stoppedRef.current) {
            return;
          }
          try {
            const { allShapesStopped } = tickShapeAnimationSim(sim, dims, rng);
            onSimTick?.(sim);
            tickCount += 1;
            const needFetch = lightsCache === null || tickCount % sceneRefreshStride === 0;
            if (needFetch) {
              lightsCache = (await fetchSceneLightsFlat(sceneId)) as SceneLightFlat[];
            }
            const updates = buildBatchUpdatesFromSim(sim, lightsCache!);
            onLightsPreview?.(updates);
            if (stoppedRef.current) {
              return;
            }
            if (updates.length > 0) {
              await patchSceneLightsStateBatch(sceneId, updates);
            }
            if (stoppedRef.current) {
              return;
            }
            if (tickCount % sceneRefreshStride === 0) {
              onSceneRefresh?.();
            }
            if (allShapesStopped) {
              stoppedRef.current = true;
              if (timer) {
                clearInterval(timer);
                timer = null;
              }
              await stopSceneRoutineRun(sceneId, runId);
              onStopped?.();
            }
          } catch (e) {
            onError?.(e instanceof Error ? e.message : "Shape animation tick failed");
            stoppedRef.current = true;
            if (timer) {
              clearInterval(timer);
              timer = null;
            }
          }
        })();
      };

      runTick();
      timer = setInterval(runTick, tickMs);
    })();

    return () => {
      stoppedRef.current = true;
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [
    sceneId,
    runId,
    definitionJson,
    onSceneRefresh,
    onError,
    onStopped,
    onSimTick,
    onLightsPreview,
    sceneRefreshStride,
  ]);

  return (
    <p className="text-xs text-emerald-800 dark:text-emerald-200">
      Shape animation is running in your browser. It updates lights through the scene
      API about {Math.round(1 / SHAPE_ANIMATION_DT_SEC)} times per second. Stop the
      routine to end it.
    </p>
  );
}
