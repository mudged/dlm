"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/Button";

export default function NewModelPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!name.trim()) {
      setError("Name is required.");
      return;
    }
    if (!file) {
      setError("Choose a CSV file.");
      return;
    }
    setSubmitting(true);
    try {
      const fd = new FormData();
      fd.set("name", name.trim());
      fd.set("file", file);
      const res = await fetch("/api/v1/models", {
        method: "POST",
        body: fd,
      });
      const j = (await res.json().catch(() => null)) as {
        error?: { message?: string };
        id?: string;
      };
      if (!res.ok) {
        setError(j?.error?.message ?? `Upload failed (${res.status})`);
        setSubmitting(false);
        return;
      }
      const id = j && "id" in j && typeof j.id === "string" ? j.id : null;
      if (id) {
        router.push(`/models/detail?id=${encodeURIComponent(id)}`);
      } else {
        router.push("/models");
      }
    } catch {
      setError("Could not reach the API.");
      setSubmitting(false);
    }
  }

  return (
    <main className="mx-auto flex max-w-lg flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <nav className="flex flex-wrap gap-3 text-sm">
        <Link
          href="/"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Home
        </Link>
        <Link
          href="/models"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Models
        </Link>
        <Link
          href="/scenes"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Scenes
        </Link>
      </nav>

      <header>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Upload model
        </h1>
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
          CSV must start with header{" "}
          <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
            id,x,y,z
          </code>
          . Ids must be sequential from 0 with no gaps (max 1000 rows).
        </p>
      </header>

      <form className="flex flex-col gap-4" onSubmit={(e) => void onSubmit(e)}>
        <div className="flex flex-col gap-1">
          <label htmlFor="model-name" className="text-sm font-medium">
            Name
          </label>
          <input
            id="model-name"
            name="name"
            type="text"
            autoComplete="off"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="min-h-11 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900"
            placeholder="e.g. Living room string"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="model-csv" className="text-sm font-medium">
            CSV file
          </label>
          <input
            id="model-csv"
            name="file"
            type="file"
            accept=".csv,text/csv"
            className="min-h-11 text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-slate-200 file:px-3 file:py-2 file:text-sm dark:file:bg-slate-700"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
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
          <Button type="submit" disabled={submitting} className="w-full sm:w-auto">
            {submitting ? "Uploading…" : "Upload"}
          </Button>
          <Button
            type="button"
            className="w-full bg-slate-500 hover:bg-slate-600 dark:bg-slate-600 dark:hover:bg-slate-500 sm:ml-2 sm:w-auto"
            onClick={() => router.push("/models")}
          >
            Cancel
          </Button>
        </div>
      </form>
    </main>
  );
}
