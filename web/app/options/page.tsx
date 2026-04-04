"use client";

import {
  faBan,
  faTriangleExclamation,
  faTrash,
} from "@fortawesome/free-solid-svg-icons";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { Button } from "@/components/ui/Button";

export default function OptionsPage() {
  const router = useRouter();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  function openDialog() {
    setErr(null);
    dialogRef.current?.showModal();
  }

  function closeDialog() {
    dialogRef.current?.close();
  }

  async function confirmFactoryReset() {
    setBusy(true);
    setErr(null);
    try {
      const res = await fetch("/api/v1/system/factory-reset", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      const j = (await res.json().catch(() => null)) as {
        ok?: boolean;
        error?: { message?: string };
      };
      if (!res.ok || !j?.ok) {
        setErr(
          j?.error?.message ??
            (res.status >= 500
              ? "Factory reset failed. Try again."
              : `Request failed (${res.status})`),
        );
        setBusy(false);
        return;
      }
      closeDialog();
      router.replace("/models?factoryReset=1");
    } catch {
      setErr("Could not reach the API.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-lg flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Options
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          Maintenance actions that affect all stored data on this device.
        </p>
      </header>

      <section
        className="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900/40"
        aria-labelledby="factory-reset-heading"
      >
        <h2
          id="factory-reset-heading"
          className="text-sm font-semibold text-slate-900 dark:text-slate-100"
        >
          Factory reset
        </h2>
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
          Remove every model and scene, then restore only the three default
          sample models (sphere, cube, cone). This cannot be undone.
        </p>
        <Button
          type="button"
          icon={faTriangleExclamation}
          className="mt-4 min-h-11 w-full bg-red-800 hover:bg-red-700 dark:bg-red-900 dark:hover:bg-red-800 sm:w-auto"
          onClick={openDialog}
        >
          Factory reset…
        </Button>
      </section>

      <dialog
        ref={dialogRef}
        className="w-[min(100%,28rem)] rounded-xl border border-slate-200 bg-white p-0 text-slate-900 shadow-xl backdrop:bg-black/40 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
        onCancel={(e) => {
          if (busy) {
            e.preventDefault();
          }
        }}
      >
        <div className="border-b border-slate-200 px-4 py-3 dark:border-slate-700">
          <h3 className="text-lg font-semibold">Erase all data?</h3>
        </div>
        <div className="space-y-3 px-4 py-4 text-sm text-slate-700 dark:text-slate-300">
          <p>
            All models you uploaded and every scene will be permanently deleted.
            Custom light settings will be lost. After this completes, only the
            three default sample models will remain.
          </p>
          <p className="text-amber-900 dark:text-amber-200">
            This action cannot be undone.
          </p>
          {err ? (
            <p className="text-red-600 dark:text-red-400" role="alert">
              {err}
            </p>
          ) : null}
        </div>
        <div className="flex flex-col-reverse gap-2 border-t border-slate-200 px-4 py-3 sm:flex-row sm:justify-end dark:border-slate-700">
          <Button
            type="button"
            icon={faBan}
            className="min-h-11 w-full bg-slate-500 hover:bg-slate-600 dark:bg-slate-600 sm:w-auto"
            disabled={busy}
            onClick={closeDialog}
          >
            Cancel
          </Button>
          <Button
            type="button"
            icon={faTrash}
            className="min-h-11 w-full bg-red-800 hover:bg-red-700 dark:bg-red-900 sm:w-auto"
            disabled={busy}
            onClick={() => void confirmFactoryReset()}
          >
            {busy ? "Working…" : "Erase everything"}
          </Button>
        </div>
      </dialog>
    </div>
  );
}
