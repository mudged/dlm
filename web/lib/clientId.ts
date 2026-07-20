/**
 * Client-side id for React list keys and similar UI-only use.
 * Prefer crypto.randomUUID when available; fall back for non-secure
 * contexts (e.g. http://LAN-IP on a Pi) where randomUUID is missing.
 */
export function clientId(): string {
  const c = globalThis.crypto;
  if (c && typeof c.randomUUID === "function") {
    return c.randomUUID();
  }
  return `id-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 11)}`;
}
