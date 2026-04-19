/**
 * Resolves the URL for browser `EventSource` (SSE).
 *
 * Next.js dev `rewrites()` proxy **buffers** `text/event-stream` responses, so light
 * updates never arrive in real time (see vercel/next.js#45048). In development we
 * therefore connect directly to the Go API (default `http://127.0.0.1:8080`).
 *
 * Set `NEXT_PUBLIC_DLM_API_ORIGIN` (and the same value as `DLM_BACKEND_ORIGIN` in
 * `next.config.ts` rewrites) if your Go server is not at `http://127.0.0.1:8080`.
 * For cross-origin dev, start Go with matching `CORS_ALLOWED_ORIGINS` (e.g.
 * `http://localhost:3000,http://127.0.0.1:3000`).
 */
export function apiOriginForEventSource(): string {
  if (typeof window === "undefined") {
    return "";
  }
  const explicit = process.env.NEXT_PUBLIC_DLM_API_ORIGIN?.trim();
  if (explicit) {
    return explicit.replace(/\/$/, "");
  }
  if (process.env.NODE_ENV === "development") {
    return "http://127.0.0.1:8080";
  }
  return "";
}

/** Path must start with `/` (e.g. `/api/v1/scenes/x/lights/events`). */
export function eventSourceUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  const origin = apiOriginForEventSource();
  return origin ? `${origin}${p}` : p;
}
