"use client";

import type { Dispatch, MutableRefObject, SetStateAction } from "react";
import { useEffect, useRef } from "react";
import {
  applyModelLightDeltas,
  parseLightsSSEMessage,
} from "@/lib/lightDeltas";
import type { ModelDetail } from "@/lib/models";
import { eventSourceUrl } from "@/lib/sseUrl";
import { manageSSEConnection } from "@/lib/sseReconnect";

/**
 * Core SSE handler factory — extracted so it can be unit-tested without React.
 * @internal exported for tests only
 */
export function createModelSSEHandlers(params: {
  seqRef: { current: number | null };
  sseLiveRef?: { current: boolean };
  modelId: string;
  setModel: Dispatch<SetStateAction<ModelDetail | null>>;
  onReload: () => void | Promise<void>;
  closeEs: () => void;
}): {
  onopen: () => void;
  onmessage: (ev: { data: string }) => void;
  onerror: () => void;
} {
  const { seqRef, sseLiveRef, modelId, setModel, onReload, closeEs } = params;

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
          console.warn(
            "[dlm] model lights SSE: ignored message (expected JSON with numeric seq). Preview:",
            preview,
          );
        }
        return;
      }
      const prev = seqRef.current;
      if (prev !== null && msg.seq !== prev + 1) {
        // Lost or reordered events: resync from GET …/models/{id}. Do not advance seqRef to
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
      setModel((m) => {
        if (!m || m.id !== modelId) {
          return m;
        }
        return { ...m, lights: applyModelLightDeltas(m.lights, deltas) };
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
 * Subscribes to model light SSE (REQ-041) and merges deltas into React state.
 * On sequence gap or connection error, calls onReload for a full model fetch.
 * Reconnects automatically with bounded exponential backoff after transient errors.
 * Optionally updates `sseLiveRef` so callers can avoid redundant polling while SSE is healthy.
 * Optionally accepts a caller-owned `seqRef` so the caller can read the latest applied
 * SSE sequence number (e.g. to guard background refetches against stale overwrites).
 */
export function useModelLightsSSE(
  modelId: string | undefined,
  setModel: Dispatch<SetStateAction<ModelDetail | null>>,
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
  }, [modelId, seqRef, sseLiveRef]);

  useEffect(() => {
    if (!modelId || typeof window === "undefined" || !enabled) {
      return;
    }
    const session = manageSSEConnection({
      makeEs: () =>
        new EventSource(
          eventSourceUrl(
            `/api/v1/models/${encodeURIComponent(modelId)}/lights/events`,
          ),
        ),
      attach(es, { scheduleReconnect, resetBackoff }) {
        const handlers = createModelSSEHandlers({
          seqRef,
          sseLiveRef,
          modelId,
          setModel,
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
  }, [modelId, enabled, setModel, onReload, seqRef, sseLiveRef]);
}
