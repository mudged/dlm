/**
 * REQ-034: 12 edges of an axis-aligned box [min,max] for LineSegments (pairs of vertices).
 * Returns null when there are no points.
 */
export function paddedAabbWireframePositions(
  points: readonly { x: number; y: number; z: number }[],
  padM: number,
): Float32Array | null {
  if (points.length === 0) {
    return null;
  }
  let mx = points[0]!.x;
  let my = points[0]!.y;
  let mz = points[0]!.z;
  let Mx = mx;
  let My = my;
  let Mz = mz;
  for (let i = 1; i < points.length; i++) {
    const p = points[i]!;
    mx = Math.min(mx, p.x);
    my = Math.min(my, p.y);
    mz = Math.min(mz, p.z);
    Mx = Math.max(Mx, p.x);
    My = Math.max(My, p.y);
    Mz = Math.max(Mz, p.z);
  }
  const p = Math.max(0, padM);
  mx -= p;
  my -= p;
  mz -= p;
  Mx += p;
  My += p;
  Mz += p;
  // 12 edges × 2 vertices × 3 components
  const out = new Float32Array(12 * 2 * 3);
  let o = 0;
  const push = (
    ax: number,
    ay: number,
    az: number,
    bx: number,
    by: number,
    bz: number,
  ) => {
    out[o++] = ax;
    out[o++] = ay;
    out[o++] = az;
    out[o++] = bx;
    out[o++] = by;
    out[o++] = bz;
  };
  // bottom face z = mz
  push(mx, my, mz, Mx, my, mz);
  push(Mx, my, mz, Mx, My, mz);
  push(Mx, My, mz, mx, My, mz);
  push(mx, My, mz, mx, my, mz);
  // top face z = Mz
  push(mx, my, Mz, Mx, my, Mz);
  push(Mx, my, Mz, Mx, My, Mz);
  push(Mx, My, Mz, mx, My, Mz);
  push(mx, My, Mz, mx, my, Mz);
  // verticals
  push(mx, my, mz, mx, my, Mz);
  push(Mx, my, mz, Mx, my, Mz);
  push(Mx, My, mz, Mx, My, Mz);
  push(mx, My, mz, mx, My, Mz);
  return out;
}
