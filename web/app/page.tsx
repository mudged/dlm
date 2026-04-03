import Link from "next/link";
import { StatusOnMount } from "@/components/StatusOnMount";
import { StatusRefresh } from "@/components/StatusRefresh";

export default function HomePage() {
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-8 px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <nav className="text-sm">
        <Link
          href="/models"
          className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
        >
          Light models
        </Link>
        <span className="text-slate-400"> · </span>
        <span className="text-slate-500">CSV upload & 3D layout</span>
      </nav>
      <header className="space-y-2">
        <p className="text-sm font-semibold uppercase tracking-wide text-sky-700 dark:text-sky-400">
          DLM
        </p>
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Domestic Light & Magic
        </h1>
        <p className="max-w-2xl text-base text-slate-600 dark:text-slate-400 sm:text-lg">
          Single binary: Next.js static export embedded in Go. Production: run{" "}
          <code className="rounded bg-slate-200 px-1.5 py-0.5 text-sm dark:bg-slate-800">
            dlm
          </code>{" "}
          only. Development: run Go on{" "}
          <code className="rounded bg-slate-200 px-1.5 py-0.5 text-sm dark:bg-slate-800">
            :8080
          </code>{" "}
          and{" "}
          <code className="rounded bg-slate-200 px-1.5 py-0.5 text-sm dark:bg-slate-800">
            npm run dev
          </code>{" "}
          in <code className="rounded bg-slate-200 px-1.5 py-0.5 text-sm dark:bg-slate-800">web/</code>{" "}
          (rewrites to the API).
        </p>
      </header>

      <section
        aria-labelledby="api-demo-heading"
        className="grid gap-6 md:grid-cols-2"
      >
        <span id="api-demo-heading" className="sr-only">
          API connectivity demos
        </span>
        <StatusOnMount />
        <StatusRefresh />
      </section>

      <footer className="text-center text-xs text-slate-500 dark:text-slate-500 sm:text-left">
        Layout reflows at <code className="font-mono">sm</code> and{" "}
        <code className="font-mono">md</code> breakpoints (Tailwind).
      </footer>
    </main>
  );
}
