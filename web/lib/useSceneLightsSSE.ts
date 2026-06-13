"use client";

import type { Dispatch, MutableRefObject, SetStateAction } from "react";
import { useEffect, useRef } from "react";
import {
  applySceneLightDeltas,
  parseLightsSSEMessage,
} from "@/lib/lightDeltas";
import type { SceneDetail } from "@/lib/scenes";
import { eventSourceUrl } from "@/lib/sseUrl";
import { manageSSEConnection } from "@/lib/sseReconnect";

/**
 * Core SSE handler factory — extracted so it can be unit-tested without React.
 * @internal exported for tests only
 */
export function createSceneSSEHandlers(params: {
  seqRef: { current: number | null };
  sseLiveRef?: { current: boolean };
  sceneId: string;
  setScene: Dispatch<SetStateAction<SceneDetail | null>>;
  onReload: () => void | Promise<void>;
  closeEs: () => void;
}): {
  onopen: () => void;
  onmessage: (ev: { data: string }) => void;
  onerror: () => void;
} {
  const { seqRef, sseLiveRef, sceneId, setScene, onReload, closeEs } = params;

  return {
    onopen() {
      if (sseLiveRef) {
        sseLiveRef.current = true;
      }
    },

    onmessage(ev) {
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
    },

    onerror() {
      closeEs();
      if (sseLiveRef) {
        sseLiveRef.current = false;
      }
      seqRef.current = null;
      void onReload();
    },
  };
}

/**
 * Subscribes to scene light SSE (REQ-041) and merges deltas into React state.
 * On sequence gap or connection error, calls onReload for a full scene fetch.
 * Reconnects automatically with bounded exponential backoff after transient errors.
 * Optionally updates `sseLiveRef` so callers can avoid redundant polling while SSE is healthy.
 * Optionally accepts a caller-owned `seqRef` so the caller can read the latest applied
 * SSE sequence number (e.g. to guard background refetches against stale overwrites).
 */
export function useSceneLightsSSE(
  sceneId: string | undefined,
  setScene: Dispatch<SetStateAction<SceneDetail | null>>,
  onReload: () => void | Promise<void>,
  options?: {
    enabled?: boolean;
    sseLiveRef?: MutableRefObject<boolean>;
    seqRef?: MutableRefObject<number | null>;
  },
): void {
  const enabled = options?.enabled ?? true;
  const sseLiveRef = options?.sseLiveRef;
  const internalSeqRef = useRef<number | null>(null);
  const seqRef = options?.seqRef ?? internalSeqRef;

  useEffect(() => {
    seqRef.current = null;
    if (sseLiveRef) {
      sseLiveRef.current = false;
    }
  }, [sceneId, seqRef, sseLiveRef]);

  useEffect(() => {
    if (!sceneId || typeof window === "undefined" || !enabled) {
      return;
    }
    const session = manageSSEConnection({
      makeEs: () =>
        new EventSource(
          eventSourceUrl(
            `/api/v1/scenes/${encodeURIComponent(sceneId)}/lights/events`,
          ),
        ),
      attach(es, { scheduleReconnect, resetBackoff }) {
        const handlers = createSceneSSEHandlers({
          seqRef,
          sseLiveRef,
          sceneId,
          setScene,
          onReload,
          closeEs: () => es.close(),
        });
        es.onopen = () => {
          resetBackoff();
          handlers.onopen();
        };
        es.onmessage = handlers.onmessage;
        es.onerror = () => {
          handlers.onerror();
          scheduleReconnect();
        };
      },
    });
    return () => {
      session.destroy();
      if (sseLiveRef) {
        sseLiveRef.current = false;
      }
    };
  }, [sceneId, enabled, setScene, onReload, seqRef, sseLiveRef]);
}
