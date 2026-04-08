/** Minimal valid definition_json for POST shape_animation (REQ-033 / §3.17.2). */
export const SHAPE_ANIMATION_DEFAULT_DEFINITION = {
  version: 1,
  background: {
    mode: "lights_on" as const,
    color: "#1a1a2e",
    brightness_pct: 40,
  },
  shapes: [
    {
      kind: "sphere" as const,
      size: { mode: "fixed" as const, radius_m: 0.35 },
      color: { mode: "fixed" as const, color: "#00ff88" },
      brightness_pct: 100,
      placement: {
        mode: "fixed" as const,
        center_m: { x: 1, y: 1, z: 1 },
      },
      motion: {
        direction: { dx: 1, dy: 0.3, dz: 0 },
        speed: { mode: "fixed" as const, m_s: 0.15 },
      },
      edge_behavior: "wrap" as const,
    },
  ],
};
