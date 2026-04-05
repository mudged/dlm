export type RoutineDefinition = {
  id: string;
  name: string;
  description: string;
  type: string;
  python_source?: string;
  created_at: string;
};

export type RoutineRun = {
  id: string;
  routine_id: string;
  routine_name: string;
  /** Present for API ≥ python routines; absent on older responses. */
  routine_type?: string;
  status: string;
};

export type StartRoutineResponse = {
  run_id: string;
  scene_id: string;
  routine_id: string;
  status: string;
};

export const ROUTINE_TYPE_RANDOM_COLOUR_ALL = "random_colour_cycle_all";
export const ROUTINE_TYPE_PYTHON_SCENE_SCRIPT = "python_scene_script";

/** Every type the UI may offer when creating a new routine (REQ-023). */
export const ROUTINE_TYPE_CREATE_OPTIONS: readonly {
  value: string;
  label: string;
}[] = [
  {
    value: ROUTINE_TYPE_RANDOM_COLOUR_ALL,
    label: "Random colour cycle (all lights)",
  },
  {
    value: ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
    label: "Python (scene API)",
  },
];

export function isCreatableRoutineType(type: string): boolean {
  return ROUTINE_TYPE_CREATE_OPTIONS.some((o) => o.value === type);
}

export async function fetchRoutines(): Promise<RoutineDefinition[]> {
  const res = await fetch("/api/v1/routines", { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`routines list failed (${res.status})`);
  }
  return res.json() as Promise<RoutineDefinition[]>;
}

export async function fetchRoutine(id: string): Promise<RoutineDefinition> {
  const res = await fetch(`/api/v1/routines/${encodeURIComponent(id)}`, {
    cache: "no-store",
  });
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `routine load failed (${res.status})`);
  }
  return res.json() as Promise<RoutineDefinition>;
}

export async function createRoutine(body: {
  name: string;
  description: string;
  type: string;
  python_source?: string;
}): Promise<RoutineDefinition> {
  const res = await fetch("/api/v1/routines", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `create routine failed (${res.status})`);
  }
  return res.json() as Promise<RoutineDefinition>;
}

export async function patchRoutine(
  id: string,
  body: {
    name?: string;
    description?: string;
    python_source?: string;
  },
): Promise<RoutineDefinition> {
  const res = await fetch(`/api/v1/routines/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const j = (await res.json().catch(() => null)) as {
      error?: { message?: string };
    };
    throw new Error(j?.error?.message ?? `save routine failed (${res.status})`);
  }
  return res.json() as Promise<RoutineDefinition>;
}

export async function deleteRoutine(id: string): Promise<void> {
  const res = await fetch(`/api/v1/routines/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (res.status === 204) {
    return;
  }
  const j = (await res.json().catch(() => null)) as {
    error?: { message?: string; code?: string };
  };
  throw new Error(j?.error?.message ?? `delete routine failed (${res.status})`);
}

export async function fetchSceneRoutineRuns(sceneId: string): Promise<RoutineRun[]> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/routines/runs`,
    { cache: "no-store" },
  );
  if (!res.ok) {
    throw new Error(`routine runs failed (${res.status})`);
  }
  const data = (await res.json()) as { runs: RoutineRun[] };
  return data.runs ?? [];
}

export async function startSceneRoutine(
  sceneId: string,
  routineId: string,
): Promise<StartRoutineResponse> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/routines/${encodeURIComponent(routineId)}/start`,
    { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" },
  );
  if (res.status === 201 || res.status === 200) {
    return res.json() as Promise<StartRoutineResponse>;
  }
  const j = (await res.json().catch(() => null)) as {
    error?: { message?: string; code?: string };
  };
  throw new Error(j?.error?.message ?? `start routine failed (${res.status})`);
}

export async function stopSceneRoutineRun(
  sceneId: string,
  runId: string,
): Promise<void> {
  const res = await fetch(
    `/api/v1/scenes/${encodeURIComponent(sceneId)}/routines/runs/${encodeURIComponent(runId)}/stop`,
    { method: "POST" },
  );
  if (res.ok) {
    return;
  }
  const j = (await res.json().catch(() => null)) as {
    error?: { message?: string };
  };
  throw new Error(j?.error?.message ?? `stop routine failed (${res.status})`);
}
