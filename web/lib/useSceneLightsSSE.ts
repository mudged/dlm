"use client";

import type { Dispatch, MutableRefObject, SetStateAction } from "react";
import { useEffect, useRef } from "react";
import {
  applySceneLightDeltas,
  parseLightsSSEMessage,
} from "@/lib/lightDeltas";
import type { SceneDetail } from "@/lib/scenes";
import { eventSourceUrl } from "@/lib/sseUrl";

/**
 * Subscribes to scene light SSE (REQ-041) and merges deltas into React state.
 * On sequence gap or connection error, calls onReload for a full scene fetch.
 * Optionally updates `sseLiveRef` so callers can avoid redundant polling while SSE is healthy.
 */
export function useSceneLightsSSE(
  sceneId: string | undefined,
  setScene: Dispatch<SetStateAction<SceneDetail | null>>,
  onReload: () => void | Promise<void>,
  options?: {
    enabled?: boolean;
    sseLiveRef?: MutableRefObject<boolean>;
  },
): void {
  const enabled = options?.enabled ?? true;
  const sseLiveRef = options?.sseLiveRef;
  const seqRef = useRef<number | null>(null);

  useEffect(() => {
    seqRef.current = null;
    if (sseLiveRef) {
      sseLiveRef.current = false;
    }
  }, [sceneId, sseLiveRef]);

  useEffect(() => {
    if (!sceneId || typeof window === "undefined" || !enabled) {
      return;
    }
    const es = new EventSource(
      eventSourceUrl(
        `/api/v1/scenes/${encodeURIComponent(sceneId)}/lights/events`,
      ),
    );
    es.onopen = () => {
      if (sseLiveRef) {
        sseLiveRef.current = true;
      }
    };
    es.onmessage = (ev) => {
      const msg = parseLightsSSEMessage(ev.data);
      if (!msg) {
        if (process.env.NODE_ENV === "development") {
          const preview =
            typeof ev.data === "string"
              ? ev.data.length > 220
                ? `${ev.data.slice(0, 220)}…`
                : ev.data
              : "";
          // Wrong endpoint or dev tooling often sends JSON without `seq` (e.g. {"key":"started"}).
          console.warn(
            "[dlm] scene lights SSE: ignored message (expected JSON with numeric seq). Preview:",
            preview,
          );
        }
        return;
      }
      const prev = seqRef.current;
      if (prev !== null && msg.seq !== prev + 1) {
        // Lost or reordered events: resync from GET …/scenes/{id}. Do not advance seqRef to
        // this message's seq — we did not apply its deltas; after reload, accept the next seq.
        seqRef.current = null;
        void Promise.resolve(onReload());
        return;
      }
      seqRef.current = msg.seq;
      const deltas = msg.deltas ?? [];
      if (deltas.length === 0) {
        return;
      }
      setScene((s) => {
        if (!s || s.id !== sceneId) {
          return s;
        }
        return { ...s, items: applySceneLightDeltas(s.items, deltas) };
      });
    };
    es.onerror = () => {
      es.close();
      if (sseLiveRef) {
        sseLiveRef.current = false;
      }
      seqRef.current = null;
      void onReload();
    };
    return () => {
      es.close();
      if (sseLiveRef) {
        sseLiveRef.current = false;
      }
    };
  }, [sceneId, enabled, setScene, onReload, sseLiveRef]);
}
