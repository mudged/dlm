"use client";

import {
  faArrowLeft,
  faFloppyDisk,
  faLink,
  faTrash,
  faUnlink,
} from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ButtonLink } from "@/components/ui/ButtonLink";
import type { Device } from "@/lib/devices";
import {
  assignDevice,
  deleteDevice,
  fetchDevice,
  patchDevice,
  unassignDevice,
} from "@/lib/devices";
import type { ModelSummary } from "@/lib/models";

export function DeviceDetailClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const rawId = searchParams.get("id");

  const [device, setDevice] = useState<Device | null>(null);
  const [models, setModels] = useState<ModelSummary[] | null>(null);
  const [name, setName] = useState("");
  const [baseURL, setBaseURL] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [assignModelId, setAssignModelId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [assignBusy, setAssignBusy] = useState(false);

  const load = useCallback(async () => {
    if (!rawId) {
      return;
    }
    setError(null);
    try {
      const [d, mRes] = await Promise.all([
        fetchDevice(rawId),
        fetch("/api/v1/models", { cache: "no-store" }),
      ]);
      setDevice(d);
      setName(d.name);
      setBaseURL(d.base_url);
      setNewPassword("");
      if (!mRes.ok) {
        setModels([]);
        return;
      }
      const m = (await mRes.json()) as ModelSummary[];
      setModels(Array.isArray(m) ? m : []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load device.");
      setDevice(null);
      setModels([]);
    }
  }, [rawId]);

  useEffect(() => {
    void load();
  }, [load]);

  const modelNameById = useMemo(() => {
    const m = new Map<string, string>();
    for (const x of models ?? []) {
      m.set(x.id, x.name);
    }
    return m;
  }, [models]);

  if (!rawId) {
    return (
      <p className="text-sm text-amber-800 dark:text-amber-200" role="alert">
        Missing device id.{" "}
        <Link href="/devices" className="underline">
          Back to devices
        </Link>
      </p>
    );
  }

  const deviceId = rawId;

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSaving(true);
    try {
      const patch: {
        name?: string;
        base_url?: string;
        wled_password?: string;
      } = {};
      if (name.trim() !== device?.name) {
        patch.name = name.trim();
      }
      if (baseURL.trim() !== device?.base_url) {
        patch.base_url = baseURL.trim();
      }
      if (newPassword.trim()) {
        patch.wled_password = newPassword.trim();
      }
      if (Object.keys(patch).length === 0) {
        setSaving(false);
        return;
      }
      const d = await patchDevice(deviceId, patch);
      setDevice(d);
      setNewPassword("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function onClearPassword() {
    if (!window.confirm("Remove the stored WLED password for this device?")) {
      return;
    }
    setError(null);
    setSaving(true);
    try {
      const d = await patchDevice(deviceId, { wled_password: "" });
      setDevice(d);
      setNewPassword("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not clear password.");
    } finally {
      setSaving(false);
    }
  }

  async function onAssign() {
    if (!assignModelId) {
      setError("Choose a model to assign.");
      return;
    }
    setError(null);
    setAssignBusy(true);
    try {
      const d = await assignDevice(deviceId, assignModelId);
      setDevice(d);
      setAssignModelId("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Assign failed.");
    } finally {
      setAssignBusy(false);
    }
  }

  async function onUnassign() {
    setError(null);
    setAssignBusy(true);
    try {
      const d = await unassignDevice(deviceId);
      setDevice(d);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unassign failed.");
    } finally {
      setAssignBusy(false);
    }
  }

  async function onDelete() {
    if (!window.confirm("Delete this device? This cannot be undone.")) {
      return;
    }
    setError(null);
    try {
      await deleteDevice(deviceId);
      router.push("/devices");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed.");
    }
  }

  if (device === null && error && !saving) {
    return (
      <>
        <p className="text-sm text-amber-800 dark:text-amber-200" role="alert">
          {error}
        </p>
        <ButtonLink href="/devices" icon={faArrowLeft} className="mt-4 w-fit">
          Back
        </ButtonLink>
      </>
    );
  }

  if (device === null) {
    return <p className="text-sm text-slate-500">Loading…</p>;
  }

  const assignedName = device.model_id
    ? modelNameById.get(device.model_id) ?? device.model_id
    : null;

  return (
    <>
      <header className="flex flex-col gap-3 border-b border-slate-200 pb-6 dark:border-slate-700">
        <ButtonLink
          href="/devices"
          icon={faArrowLeft}
          className="w-fit border-0 bg-transparent px-0 hover:bg-transparent dark:bg-transparent dark:hover:bg-transparent"
        >
          All devices
        </ButtonLink>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {device.name}
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          WLED · {device.base_url}
        </p>
      </header>

      {error ? (
        <p
          className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
          role="alert"
        >
          {error}
        </p>
      ) : null}

      <form className="flex flex-col gap-4" onSubmit={(e) => void onSave(e)}>
        <div className="flex flex-col gap-1">
          <label htmlFor="edit-name" className="text-sm font-medium">
            Name
          </label>
          <input
            id="edit-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="edit-url" className="text-sm font-medium">
            Base URL
          </label>
          <input
            id="edit-url"
            type="url"
            inputMode="url"
            value={baseURL}
            onChange={(e) => setBaseURL(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="edit-pw" className="text-sm font-medium">
            New WLED password (optional)
          </label>
          <input
            id="edit-pw"
            type="password"
            autoComplete="new-password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="Leave blank to keep current"
          />
          <button
            type="button"
            className="self-start text-xs font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
            onClick={() => void onClearPassword()}
          >
            Remove stored password
          </button>
        </div>
        <Button
          type="submit"
          icon={faFloppyDisk}
          disabled={saving}
          className="w-full sm:w-auto"
        >
          {saving ? "Saving…" : "Save changes"}
        </Button>
      </form>

      <section className="mt-8 flex flex-col gap-3 rounded-xl border border-slate-200 p-4 dark:border-slate-700">
        <h2 className="text-sm font-semibold text-slate-900 dark:text-white">
          Model assignment
        </h2>
        <p className="text-xs text-slate-600 dark:text-slate-400">
          One device can drive at most one model, and each model can have at
          most one device.
        </p>
        {device.model_id ? (
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm">
              Assigned to{" "}
              <span className="font-medium">{assignedName}</span>
            </p>
            <Button
              type="button"
              icon={faUnlink}
              disabled={assignBusy}
              className="w-full bg-amber-800 hover:bg-amber-700 sm:w-auto dark:bg-amber-900 dark:hover:bg-amber-800"
              onClick={() => void onUnassign()}
            >
              Unassign
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <label htmlFor="assign-model" className="text-sm font-medium">
                Assign to model
              </label>
              <select
                id="assign-model"
                value={assignModelId}
                onChange={(e) => setAssignModelId(e.target.value)}
                className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
              >
                <option value="">Select a model…</option>
                {(models ?? []).map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.light_count} lights)
                  </option>
                ))}
              </select>
            </div>
            <Button
              type="button"
              icon={faLink}
              disabled={assignBusy || !(models ?? []).length}
              className="w-full sm:w-auto"
              onClick={() => void onAssign()}
            >
              Assign
            </Button>
          </div>
        )}
      </section>

      <div className="mt-8 border-t border-slate-200 pt-6 dark:border-slate-700">
        <Button
          type="button"
          icon={faTrash}
          className="w-full bg-red-800 hover:bg-red-700 sm:w-auto dark:bg-red-900 dark:hover:bg-red-800"
          onClick={() => void onDelete()}
        >
          Delete device
        </Button>
      </div>
    </>
  );
}
