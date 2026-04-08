"use client";

import {
  faPenToSquare,
  faPlus,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ButtonLink } from "@/components/ui/ButtonLink";
import type { Device } from "@/lib/devices";
import { deleteDevice, fetchDevices } from "@/lib/devices";

export default function DevicesPage() {
  const router = useRouter();
  const [devices, setDevices] = useState<Device[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const list = await fetchDevices();
      setDevices(list);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load devices.");
      setDevices([]);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function remove(id: string) {
    if (!window.confirm("Delete this device?")) return;
    setBusyId(id);
    setError(null);
    try {
      await deleteDevice(id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Devices
          </h1>
          <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
            Register WLED controllers and assign one device per light model.
            Discovery is not available yet; add the base URL manually (for
            example <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">http://wled.local</code>
            ).
          </p>
        </div>
        <ButtonLink
          href="/devices/new"
          icon={faPlus}
          className="w-full sm:w-auto"
        >
          Add device
        </ButtonLink>
      </header>

      {error ? (
        <p
          className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
          role="alert"
        >
          {error}
        </p>
      ) : null}

      {devices === null ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : devices.length === 0 ? (
        <p className="text-sm text-slate-600 dark:text-slate-400">
          No devices yet.{" "}
          <Link
            href="/devices/new"
            className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
          >
            Add a WLED device
          </Link>
          .
        </p>
      ) : (
        <ul className="divide-y divide-slate-200 rounded-xl border border-slate-200 bg-white dark:divide-slate-700 dark:border-slate-700 dark:bg-slate-900/40">
          {devices.map((d) => (
            <li
              key={d.id}
              className="flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0">
                <Link
                  href={`/devices/detail?id=${encodeURIComponent(d.id)}`}
                  className="font-medium text-sky-800 hover:underline dark:text-sky-300"
                >
                  {d.name}
                </Link>
                <p className="truncate text-xs text-slate-500 dark:text-slate-400">
                  {d.type} · {d.base_url}
                  {d.model_id ? ` · model ${d.model_id.slice(0, 8)}…` : ""}
                </p>
              </div>
              <div className="flex shrink-0 gap-2">
                <Button
                  type="button"
                  icon={faPenToSquare}
                  className="min-h-11 bg-slate-600 dark:bg-slate-600"
                  onClick={() =>
                    router.push(
                      `/devices/detail?id=${encodeURIComponent(d.id)}`,
                    )
                  }
                >
                  Edit
                </Button>
                <Button
                  type="button"
                  icon={faTrash}
                  className="min-h-11 bg-red-800 hover:bg-red-700 dark:bg-red-900 dark:hover:bg-red-800"
                  disabled={busyId === d.id}
                  onClick={() => void remove(d.id)}
                >
                  Delete
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
