import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import type { IconDefinition } from "@fortawesome/fontawesome-svg-core";
import Link from "next/link";
import type { ComponentProps, ReactNode } from "react";

type Props = Omit<ComponentProps<typeof Link>, "className" | "children"> & {
  icon: IconDefinition;
  children: ReactNode;
  className?: string;
};

/** Button-styled Next.js Link with Font Awesome icon (REQ-018). */
export function ButtonLink({
  icon,
  children,
  className = "",
  ...linkProps
}: Props) {
  return (
    <Link
      {...linkProps}
      className={`inline-flex min-h-11 min-w-11 items-center justify-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2.5 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800 ${className}`}
    >
      <FontAwesomeIcon icon={icon} className="h-[1em] w-[1em] shrink-0" aria-hidden />
      <span>{children}</span>
    </Link>
  );
}
