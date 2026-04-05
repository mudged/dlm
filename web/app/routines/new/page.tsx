import { Suspense } from "react";
import RoutineNewClient from "./RoutineNewClient";

function RoutineNewFallback() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <p className="text-sm text-slate-500">Loading…</p>
    </div>
  );
}

export default function RoutineNewPage() {
  return (
    <Suspense fallback={<RoutineNewFallback />}>
      <RoutineNewClient />
    </Suspense>
  );
}
