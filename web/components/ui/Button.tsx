import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import type { IconDefinition } from "@fortawesome/fontawesome-svg-core";
import type { ButtonHTMLAttributes, ReactNode } from "react";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon: IconDefinition;
  children: ReactNode;
};

/** Touch-friendly minimum hit target (min 44px) per REQ-002. REQ-018: visible Font Awesome icon. */
export function Button({
  className = "",
  icon,
  children,
  type = "button",
  ...props
}: Props) {
  return (
    <button
      type={type}
      className={`inline-flex min-h-11 min-w-11 items-center justify-center gap-2 rounded-lg bg-slate-800 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 active:bg-slate-900 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white ${className}`}
      {...props}
    >
      <FontAwesomeIcon icon={icon} className="h-[1em] w-[1em] shrink-0" aria-hidden />
      <span>{children}</span>
    </button>
  );
}
