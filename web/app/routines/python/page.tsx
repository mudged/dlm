import { Suspense } from "react";
import PythonRoutineEditorClient from "./PythonRoutineEditorClient";

export default function PythonRoutineEditorPage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-6xl px-4 py-8">
          <p className="text-sm text-slate-500">Loading editor…</p>
        </div>
      }
    >
      <PythonRoutineEditorClient />
    </Suspense>
  );
}
