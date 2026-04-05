import { redirect } from "next/navigation";

/** REQ-023 / architecture §4.13 — canonical create UX is `/routines/new`. */
export default function PythonRoutineNewRedirectPage() {
  redirect("/routines/new?type=python_scene_script");
}
