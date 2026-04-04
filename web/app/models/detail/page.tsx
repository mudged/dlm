import { Suspense } from "react";
import { ModelDetailClient } from "./ModelDetailClient";

export default function ModelDetailPage() {
  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <Suspense
        fallback={<p className="text-sm text-slate-500">Loading…</p>}
      >
        <ModelDetailClient />
      </Suspense>
    </div>
  );
}
