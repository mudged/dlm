"use client";

import { faFloppyDisk, faXmark } from "@fortawesome/free-solid-svg-icons";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { ButtonLink } from "@/components/ui/ButtonLink";
import { createDevice } from "@/lib/devices";

export default function NewDevicePage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [baseURL, setBaseURL] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!name.trim()) {
      setError("Name is required.");
      return;
    }
    if (!baseURL.trim()) {
      setError("Base URL is required (for example http://wled.local).");
      return;
    }
    setSubmitting(true);
    try {
      const d = await createDevice({
        name: name.trim(),
        base_url: baseURL.trim(),
        wled_password: password.trim() || undefined,
      });
      router.push(`/devices/detail?id=${encodeURIComponent(d.id)}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create device.");
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-lg flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <header>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Add WLED device
        </h1>
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
          Use the device&apos;s reachable URL without a trailing slash. If the
          device uses an AP password, enter it here (stored locally in SQLite).
        </p>
      </header>

      <form className="flex flex-col gap-4" onSubmit={(e) => void onSubmit(e)}>
        <div className="flex flex-col gap-1">
          <label htmlFor="device-name" className="text-sm font-medium">
            Name
          </label>
          <input
            id="device-name"
            name="name"
            type="text"
            autoComplete="off"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="e.g. Living room strip"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="device-base-url" className="text-sm font-medium">
            Base URL
          </label>
          <input
            id="device-base-url"
            name="base_url"
            type="url"
            inputMode="url"
            autoComplete="off"
            value={baseURL}
            onChange={(e) => setBaseURL(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="http://192.168.1.50"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="device-pw" className="text-sm font-medium">
            WLED password (optional)
          </label>
          <input
            id="device-pw"
            name="wled_password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="Leave empty if none"
          />
        </div>

        {error ? (
          <p
            className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-100"
            role="alert"
          >
            {error}
          </p>
        ) : null}

        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            type="submit"
            icon={faFloppyDisk}
            disabled={submitting}
            className="w-full sm:w-auto"
          >
            {submitting ? "Saving…" : "Save device"}
          </Button>
          <ButtonLink
            href="/devices"
            icon={faXmark}
            className="w-full sm:ml-2 sm:w-auto"
          >
            Cancel
          </ButtonLink>
        </div>
      </form>
    </div>
  );
}
