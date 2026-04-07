import { redirect } from "next/navigation";

/** New Python routines are created from the main “new routine” flow (`/routines/new`). */
export default function PythonRoutineNewRedirectPage() {
  redirect("/routines/new");
}
