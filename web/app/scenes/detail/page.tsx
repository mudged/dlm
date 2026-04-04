import { Suspense } from "react";
import Link from "next/link";
import { SceneDetailClient } from "./SceneDetailClient";

export default function SceneDetailPage() {
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <nav className="flex flex-wrap gap-3 text-sm">
        <Link
          href="/"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Home
        </Link>
        <Link
          href="/scenes"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Scenes
        </Link>
        <Link
          href="/models"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Models
        </Link>
        <Link
          href="/options"
          className="text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Options
        </Link>
      </nav>

      <Suspense
        fallback={<p className="text-sm text-slate-500">Loading…</p>}
      >
        <SceneDetailClient />
      </Suspense>
    </main>
  );
}
