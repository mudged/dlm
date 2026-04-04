"use client";

import { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { boundingFromLights } from "@/lib/lightBounds";
import type { Light } from "@/lib/models";
import {
  colorFromHexAndBrightness,
  normalizeLightHex,
} from "@/lib/lightAppearance";
import {
  buildWireSegmentPositions,
  SPHERE_RADIUS_M,
} from "@/lib/wireSegments";
import { VIZ_VIEWPORT_BG, VIZ_VIEWPORT_BG_CSS } from "@/lib/vizViewport";

type Props = {
  lights: Light[];
  /** When set, orbit position/target are restored across `lights` updates with the same key (e.g. model id). */
  cameraPersistenceKey?: string;
  /** Increment to re-apply default framing without persisting orbit (REQ-016). */
  cameraResetVersion?: number;
};

type PickData = {
  id: number;
  x: number;
  y: number;
  z: number;
  on: boolean;
  color: string;
  brightness_pct: number;
};
/** Client coordinates for `position: fixed` tooltip (avoids overflow clipping). */
type TooltipState = { pick: PickData; cx: number; cy: number };

const TAP_PX = 10;
/** REQ-010: #D0D0D0 at 85% transparency (15% opacity); matches off-sphere styling (REQ-012). */
const VIZ_GREY = 0xd0d0d0;
const LINE_OPACITY = 0.15;
const OFF_SPHERE_OPACITY = 0.15;
const HOVER_DECIMALS = 4;

function lightOn(L: Light): boolean {
  return L.on !== false;
}

function lightColor(L: Light): string {
  return L.color ?? "#ffffff";
}

function lightBrightness(L: Light): number {
  const v = L.brightness_pct;
  return typeof v === "number" && Number.isFinite(v) ? v : 100;
}

/** REQ-019: fixed dark-grey backdrop regardless of shell theme. */
function sceneBackgroundColor(): THREE.Color {
  return new THREE.Color(VIZ_VIEWPORT_BG);
}

/** Low-contrast floor grid tuned for VIZ_VIEWPORT_BG (REQ-019). */
function createSubtleGridHelper(
  size: number,
  divisions: number,
): THREE.GridHelper {
  const centerLine = 0x2c3b4f;
  const gridLine = 0x1b2638;
  const grid = new THREE.GridHelper(size, divisions, centerLine, gridLine);
  const mats = grid.material;
  const list = Array.isArray(mats) ? mats : [mats];
  for (const m of list) {
    if (m instanceof THREE.LineBasicMaterial) {
      m.transparent = true;
      m.opacity = 0.38;
      m.depthWrite = false;
    }
  }
  return grid;
}

function formatCoord(n: number): string {
  return n.toFixed(HOVER_DECIMALS);
}

type SavedOrbit = {
  key: string;
  position: THREE.Vector3;
  target: THREE.Vector3;
};

function ModelLightsCanvas({
  lights,
  cameraPersistenceKey,
  cameraResetVersion = 0,
}: Props) {
  const wrapRef = useRef<HTMLDivElement>(null);
  const canvasHostRef = useRef<HTMLDivElement>(null);
  const orbitRef = useRef<SavedOrbit | null>(null);
  const prevResetVerRef = useRef(cameraResetVersion);
  const [pinned, setPinned] = useState<TooltipState | null>(null);
  const [hover, setHover] = useState<TooltipState | null>(null);
  const pinnedRef = useRef<TooltipState | null>(null);
  pinnedRef.current = pinned;

  useEffect(() => {
    setPinned(null);
    setHover(null);
  }, [lights, cameraPersistenceKey, cameraResetVersion]);

  useEffect(() => {
    const wrap = wrapRef.current;
    const container = canvasHostRef.current;
    if (!container || !wrap) {
      return;
    }

    const sorted = [...lights].sort((a, b) => a.id - b.id);
    const n = sorted.length;

    const scene = new THREE.Scene();
    scene.background = sceneBackgroundColor();

    const camera = new THREE.PerspectiveCamera(55, 1, 0.001, 1e7);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setClearColor(VIZ_VIEWPORT_BG, 1);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.screenSpacePanning = true;

    const margin = SPHERE_RADIUS_M * 2;
    const { center, maxDim } = boundingFromLights(lights);
    const [cx, cy, cz] = center;
    const framedDim = maxDim + margin;
    const target = new THREE.Vector3(cx, cy, cz);

    const forceDefaultFraming =
      cameraResetVersion !== prevResetVerRef.current;
    prevResetVerRef.current = cameraResetVersion;

    const saved =
      cameraPersistenceKey !== undefined
        ? orbitRef.current
        : null;
    const restoreOrbit =
      !forceDefaultFraming &&
      saved !== null &&
      cameraPersistenceKey !== undefined &&
      saved.key === cameraPersistenceKey;

    if (restoreOrbit) {
      camera.position.copy(saved.position);
      controls.target.copy(saved.target);
      camera.lookAt(controls.target);
    } else {
      controls.target.copy(target);
      const dist = Math.max(framedDim * 1.8, 0.5);
      camera.position.set(cx + dist * 0.85, cy + dist * 0.55, cz + dist * 0.85);
      camera.lookAt(target);
    }
    controls.update();

    const gridSize = Math.max(framedDim * 2.5, 1);
    const grid = createSubtleGridHelper(gridSize, 12);
    grid.position.set(cx, cy - framedDim * 0.5 - 1e-6, cz);
    scene.add(grid);

    const disposeGrid = () => {
      grid.geometry.dispose();
      const mat = grid.material;
      if (Array.isArray(mat)) {
        mat.forEach((m) => m.dispose());
      } else {
        mat.dispose();
      }
    };

    scene.add(new THREE.AmbientLight(0xffffff, 0.55));
    const dir = new THREE.DirectionalLight(0xffffff, 0.9);
    dir.position.set(1, 1.5, 0.8);
    scene.add(dir);

    const sphereGeom = new THREE.SphereGeometry(SPHERE_RADIUS_M, 20, 16);
    /** Per-(hex, brightness) materials; shared across lights with identical appearance (§4.7 option B). */
    const onMaterialCache = new Map<string, THREE.MeshBasicMaterial>();
    function basicMaterialForLight(L: Light): THREE.MeshBasicMaterial {
      const key = `${normalizeLightHex(lightColor(L))}:${lightBrightness(L)}`;
      let m = onMaterialCache.get(key);
      if (!m) {
        const col = colorFromHexAndBrightness(
          lightColor(L),
          lightBrightness(L),
        );
        m = new THREE.MeshBasicMaterial({ color: col });
        onMaterialCache.set(key, m);
      }
      return m;
    }
    const matOff = new THREE.MeshBasicMaterial({
      color: VIZ_GREY,
      transparent: true,
      opacity: OFF_SPHERE_OPACITY,
      depthWrite: false,
    });

    const onSortedIdx: number[] = [];
    const offSortedIdx: number[] = [];
    for (let i = 0; i < n; i++) {
      if (lightOn(sorted[i])) {
        onSortedIdx.push(i);
      } else {
        offSortedIdx.push(i);
      }
    }

    const onMeshes: THREE.Mesh[] = [];
    let instOff: THREE.InstancedMesh | null = null;
    const dummy = new THREE.Object3D();

    if (onSortedIdx.length > 0) {
      for (let j = 0; j < onSortedIdx.length; j++) {
        const si = onSortedIdx[j]!;
        const L = sorted[si]!;
        const mesh = new THREE.Mesh(sphereGeom, basicMaterialForLight(L));
        mesh.position.set(L.x, L.y, L.z);
        mesh.userData.sortedIdx = si;
        scene.add(mesh);
        onMeshes.push(mesh);
      }
    }

    if (offSortedIdx.length > 0) {
      const c = offSortedIdx.length;
      instOff = new THREE.InstancedMesh(sphereGeom, matOff, c);
      instOff.instanceMatrix.setUsage(THREE.DynamicDrawUsage);
      for (let j = 0; j < c; j++) {
        const L = sorted[offSortedIdx[j]!];
        dummy.position.set(L.x, L.y, L.z);
        dummy.updateMatrix();
        instOff.setMatrixAt(j, dummy.matrix);
      }
      instOff.instanceMatrix.needsUpdate = true;
      scene.add(instOff);
    }

    const pickTargets: THREE.Object3D[] = [...onMeshes];
    if (instOff) {
      pickTargets.push(instOff);
    }

    const segPos = buildWireSegmentPositions(lights);
    let lineSegments: THREE.LineSegments | null = null;
    if (segPos.length > 0) {
      const lg = new THREE.BufferGeometry();
      lg.setAttribute("position", new THREE.BufferAttribute(segPos, 3));
      const lm = new THREE.LineBasicMaterial({
        color: VIZ_GREY,
        transparent: true,
        opacity: LINE_OPACITY,
        depthWrite: false,
      });
      lineSegments = new THREE.LineSegments(lg, lm);
      scene.add(lineSegments);
    }

    const raycaster = new THREE.Raycaster();
    const ndc = new THREE.Vector2();

    function pickInstance(clientX: number, clientY: number): PickData | null {
      if (pickTargets.length === 0 || n === 0) {
        return null;
      }
      const canvas = renderer.domElement;
      const cr = canvas.getBoundingClientRect();
      ndc.x = ((clientX - cr.left) / cr.width) * 2 - 1;
      ndc.y = -((clientY - cr.top) / cr.height) * 2 + 1;
      raycaster.setFromCamera(ndc, camera);
      const hits = raycaster.intersectObjects(pickTargets, false);
      const hit = hits[0];
      if (hit === undefined) {
        return null;
      }
      let si: number | undefined;
      const fromUser = hit.object.userData.sortedIdx;
      if (typeof fromUser === "number") {
        si = fromUser;
      } else if (hit.object === instOff && hit.instanceId !== undefined) {
        si = offSortedIdx[hit.instanceId]!;
      }
      if (si === undefined) {
        return null;
      }
      const L = sorted[si]!;
      return {
        id: L.id,
        x: L.x,
        y: L.y,
        z: L.z,
        on: lightOn(L),
        color: lightColor(L),
        brightness_pct: lightBrightness(L),
      };
    }

    let rafHover = 0;
    let pendingHover: { cx: number; cy: number } | null = null;

    function applyHover(clientX: number, clientY: number) {
      if (pinnedRef.current) {
        return;
      }
      const pick = pickInstance(clientX, clientY);
      if (pick) {
        setHover({ pick, cx: clientX, cy: clientY });
      } else {
        setHover(null);
      }
    }

    function scheduleHover(clientX: number, clientY: number) {
      pendingHover = { cx: clientX, cy: clientY };
      if (rafHover) {
        return;
      }
      rafHover = requestAnimationFrame(() => {
        rafHover = 0;
        if (pendingHover) {
          const { cx, cy } = pendingHover;
          pendingHover = null;
          applyHover(cx, cy);
        }
      });
    }

    const ptr = {
      down: false,
      x: 0,
      y: 0,
      dragged: false,
    };

    function onPointerDown(e: PointerEvent) {
      ptr.down = true;
      ptr.x = e.clientX;
      ptr.y = e.clientY;
      ptr.dragged = false;
    }

    function onPointerMove(e: PointerEvent) {
      if (ptr.down) {
        const d = Math.hypot(e.clientX - ptr.x, e.clientY - ptr.y);
        if (d > TAP_PX) {
          ptr.dragged = true;
        }
      }
      if (e.pointerType === "mouse") {
        scheduleHover(e.clientX, e.clientY);
      }
    }

    function onPointerUp(e: PointerEvent) {
      if (ptr.down && !ptr.dragged) {
        const pick = pickInstance(e.clientX, e.clientY);
        if (pick) {
          setPinned({ pick, cx: e.clientX, cy: e.clientY });
          setHover(null);
        } else {
          setPinned(null);
          if (e.pointerType === "mouse") {
            applyHover(e.clientX, e.clientY);
          }
        }
      }
      ptr.down = false;
      ptr.dragged = false;
    }

    function onPointerLeave() {
      if (!pinnedRef.current) {
        setHover(null);
      }
    }

    const canvasEl = renderer.domElement;
    canvasEl.addEventListener("pointerdown", onPointerDown);
    canvasEl.addEventListener("pointermove", onPointerMove);
    canvasEl.addEventListener("pointerup", onPointerUp);
    canvasEl.addEventListener("pointercancel", onPointerUp);

    wrap.addEventListener("pointerleave", onPointerLeave);

    container.appendChild(canvasEl);

    let raf = 0;
    const tick = () => {
      controls.update();
      renderer.render(scene, camera);
      raf = requestAnimationFrame(tick);
    };

    const setSize = () => {
      const w = Math.max(container.clientWidth, 1);
      const h = Math.max(container.clientHeight, 1);
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h, false);
    };

    setSize();
    const ro = new ResizeObserver(() => setSize());
    ro.observe(container);
    tick();

    return () => {
      if (cameraPersistenceKey !== undefined) {
        orbitRef.current = {
          key: cameraPersistenceKey,
          position: camera.position.clone(),
          target: controls.target.clone(),
        };
      } else {
        orbitRef.current = null;
      }
      cancelAnimationFrame(raf);
      cancelAnimationFrame(rafHover);
      ro.disconnect();
      canvasEl.removeEventListener("pointerdown", onPointerDown);
      canvasEl.removeEventListener("pointermove", onPointerMove);
      canvasEl.removeEventListener("pointerup", onPointerUp);
      canvasEl.removeEventListener("pointercancel", onPointerUp);
      wrap.removeEventListener("pointerleave", onPointerLeave);
      controls.dispose();
      sphereGeom.dispose();
      for (const m of onMaterialCache.values()) {
        m.dispose();
      }
      onMaterialCache.clear();
      matOff.dispose();
      if (lineSegments) {
        lineSegments.geometry.dispose();
        (lineSegments.material as THREE.Material).dispose();
      }
      disposeGrid();
      renderer.dispose();
      if (canvasEl.parentNode === container) {
        container.removeChild(canvasEl);
      }
    };
  }, [lights, cameraPersistenceKey, cameraResetVersion]);

  const tip = pinned ?? hover;

  return (
    <div
      ref={wrapRef}
      className="relative h-[min(50vh,24rem)] w-full min-h-[240px] overflow-hidden rounded-xl border border-slate-600/40"
      style={{ backgroundColor: VIZ_VIEWPORT_BG_CSS }}
      role="img"
      aria-label="Three-dimensional view of light positions"
    >
      <div ref={canvasHostRef} className="absolute inset-0" />
      {tip ? (
        <div
          className="pointer-events-none fixed z-[100] max-w-[14rem] rounded-md border border-slate-300 bg-white/95 px-2 py-1.5 text-xs text-slate-900 shadow-md dark:border-slate-600 dark:bg-slate-900/95 dark:text-slate-100"
          style={{
            left: tip.cx + 12,
            top: tip.cy + 12,
          }}
        >
          <div className="font-semibold">id {tip.pick.id}</div>
          <div className="font-mono tabular-nums text-[0.7rem] leading-relaxed">
            x {formatCoord(tip.pick.x)}
            <br />
            y {formatCoord(tip.pick.y)}
            <br />
            z {formatCoord(tip.pick.z)}
            <br />
            <span className="text-slate-600 dark:text-slate-300">
              {tip.pick.on ? "on" : "off"} · {tip.pick.color} ·{" "}
              {tip.pick.brightness_pct}%
            </span>
          </div>
        </div>
      ) : null}
      {lights.length === 0 ? (
        <span className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center bg-black/35 text-sm text-slate-300">
          No lights — empty model
        </span>
      ) : null}
    </div>
  );
}

export { ModelLightsCanvas };
export default ModelLightsCanvas;
