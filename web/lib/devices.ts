export type Device = {
  id: string;
  type: string;
  name: string;
  base_url: string;
  model_id?: string;
  created_at: string;
};

export type DeviceListResponse = {
  devices: Device[];
};

type ApiErrorBody = {
  error?: { message?: string; code?: string };
};

function readErrorMessage(res: Response, data: unknown): string {
  const j = data as ApiErrorBody;
  return j?.error?.message ?? `Request failed (${res.status})`;
}

export async function fetchDevices(): Promise<Device[]> {
  const res = await fetch("/api/v1/devices", { cache: "no-store" });
  const data = (await res.json().catch(() => null)) as
    | DeviceListResponse
    | ApiErrorBody;
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  const list = (data as DeviceListResponse).devices;
  return Array.isArray(list) ? list : [];
}

export async function fetchDevice(id: string): Promise<Device> {
  const res = await fetch(`/api/v1/devices/${encodeURIComponent(id)}`, {
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  return data as Device;
}

export async function createDevice(input: {
  name: string;
  base_url: string;
  wled_password?: string;
}): Promise<Device> {
  const payload: Record<string, string> = {
    type: "wled",
    name: input.name.trim(),
    base_url: input.base_url.trim(),
  };
  const pw = input.wled_password?.trim();
  if (pw) {
    payload.wled_password = pw;
  }
  const res = await fetch("/api/v1/devices", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  return data as Device;
}

export async function patchDevice(
  id: string,
  patch: { name?: string; base_url?: string; wled_password?: string },
): Promise<Device> {
  const res = await fetch(`/api/v1/devices/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  return data as Device;
}

export async function deleteDevice(id: string): Promise<void> {
  const res = await fetch(`/api/v1/devices/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (res.status === 204) {
    return;
  }
  const data = await res.json().catch(() => null);
  throw new Error(readErrorMessage(res, data));
}

export async function assignDevice(
  deviceId: string,
  modelId: string,
): Promise<Device> {
  const res = await fetch(
    `/api/v1/devices/${encodeURIComponent(deviceId)}/assign`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ model_id: modelId }),
    },
  );
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  return data as Device;
}

export async function unassignDevice(deviceId: string): Promise<Device> {
  const res = await fetch(
    `/api/v1/devices/${encodeURIComponent(deviceId)}/unassign`,
    { method: "POST", headers: { "Content-Type": "application/json" } },
  );
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(readErrorMessage(res, data));
  }
  return data as Device;
}
