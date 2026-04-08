"use client";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faLightbulb } from "@fortawesome/free-regular-svg-icons";
import {
  faBars,
  faCubes,
  faGear,
  faHouse,
  faLayerGroup,
  faMoon,
  faNetworkWired,
  faSun,
  faWandMagicSparkles,
} from "@fortawesome/free-solid-svg-icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useEffect, useLayoutEffect, useState } from "react";
import { UiThemeContext } from "@/components/UiThemeContext";
import {
  applyThemeClass,
  persistTheme,
  readStoredTheme,
  resolveInitialTheme,
  type ThemeChoice,
} from "@/lib/theme";

const BRAND_TITLE = "Domestic Light & Magic";

type NavItem = { href: string; label: string; icon: typeof faHouse };

const NAV: NavItem[] = [
  { href: "/", label: "Home", icon: faHouse },
  { href: "/models", label: "Models", icon: faCubes },
  { href: "/devices", label: "Devices", icon: faNetworkWired },
  { href: "/scenes", label: "Scenes", icon: faLayerGroup },
  { href: "/routines", label: "Routines", icon: faWandMagicSparkles },
  { href: "/options", label: "Options", icon: faGear },
];

function isLgViewport(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  return window.matchMedia("(min-width: 1024px)").matches;
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [theme, setTheme] = useState<ThemeChoice>("light");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  useLayoutEffect(() => {
    const initial = resolveInitialTheme(readStoredTheme());
    setTheme(initial);
    applyThemeClass(initial);
  }, []);

  const toggleTheme = useCallback(() => {
    const next: ThemeChoice = theme === "light" ? "dark" : "light";
    setTheme(next);
    applyThemeClass(next);
    persistTheme(next);
  }, [theme]);

  const toggleNav = useCallback(() => {
    if (isLgViewport()) {
      setSidebarCollapsed((c) => !c);
    } else {
      setDrawerOpen((o) => !o);
    }
  }, []);

  useEffect(() => {
    setDrawerOpen(false);
  }, [pathname]);

  useEffect(() => {
    const mq = window.matchMedia("(min-width: 1024px)");
    const onChange = () => {
      if (mq.matches) {
        setDrawerOpen(false);
      }
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const showLabels = !sidebarCollapsed;

  return (
    <div className="flex min-h-screen bg-white text-gray-900 dark:bg-gray-900 dark:text-white">
      {drawerOpen ? (
        <button
          type="button"
          aria-label="Close menu"
          className="fixed inset-0 z-20 bg-black/40 lg:hidden"
          onClick={() => setDrawerOpen(false)}
        />
      ) : null}

      <aside
        className={`fixed left-0 top-0 z-30 flex h-full flex-col border-r border-slate-200 bg-white transition-all dark:border-slate-700 dark:bg-gray-900 lg:static lg:z-0 ${
          drawerOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        } ${sidebarCollapsed ? "w-16 lg:w-16" : "w-56"}`}
        aria-label="Primary navigation"
      >
        <div className="flex flex-1 flex-col gap-1 p-2 pt-4">
          {NAV.map(({ href, label, icon }) => {
            const active =
              href === "/"
                ? pathname === "/"
                : pathname === href || pathname.startsWith(`${href}/`);
            return (
              <Link
                key={href}
                href={href}
                title={label}
                className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                  active
                    ? "bg-slate-200 text-slate-900 dark:bg-slate-700 dark:text-white"
                    : "text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800"
                } ${!showLabels ? "justify-center px-2" : ""}`}
              >
                <FontAwesomeIcon icon={icon} className="h-5 w-5 shrink-0" aria-hidden />
                {showLabels ? <span>{label}</span> : null}
              </Link>
            );
          })}
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-10 flex items-center gap-3 border-b border-slate-200 bg-white px-3 py-3 dark:border-slate-700 dark:bg-gray-900">
          <button
            type="button"
            onClick={toggleNav}
            className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-lg border border-slate-300 text-slate-800 hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 dark:border-slate-600 dark:text-slate-100 dark:hover:bg-slate-800"
            aria-label="Toggle navigation menu"
          >
            <FontAwesomeIcon icon={faBars} className="h-5 w-5" aria-hidden />
          </button>
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <FontAwesomeIcon
              icon={faLightbulb}
              className="h-7 w-7 shrink-0 text-amber-500 dark:text-amber-400"
              aria-hidden
            />
            <span className="truncate text-base font-semibold tracking-tight sm:text-lg">
              {BRAND_TITLE}
            </span>
          </div>
          <button
            type="button"
            onClick={toggleTheme}
            className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-lg border border-slate-300 text-slate-800 hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 dark:border-slate-600 dark:text-slate-100 dark:hover:bg-slate-800"
            aria-label={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
            aria-pressed={theme === "dark"}
          >
            <FontAwesomeIcon
              icon={theme === "dark" ? faSun : faMoon}
              className="h-5 w-5"
              aria-hidden
            />
          </button>
        </header>

        <UiThemeContext.Provider value={theme}>
          <main className="min-w-0 flex-1 bg-white dark:bg-gray-900">
            {children}
          </main>
        </UiThemeContext.Provider>
      </div>
    </div>
  );
}
