export const THEME_STORAGE_KEY = "dlm-theme";

export type ThemeChoice = "light" | "dark";

/** Read persisted theme from localStorage; returns null if unset or unreadable. */
export function readStoredTheme(): ThemeChoice | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const v = localStorage.getItem(THEME_STORAGE_KEY);
    if (v === "dark" || v === "light") {
      return v;
    }
  } catch {
    /* ignore */
  }
  return null;
}

/**
 * Effective shell theme before React runs: persisted choice wins; otherwise
 * prefers-color-scheme; otherwise light (REQ-018).
 */
export function resolveInitialTheme(stored: ThemeChoice | null): ThemeChoice {
  if (stored === "dark" || stored === "light") {
    return stored;
  }
  if (typeof window === "undefined") {
    return "light";
  }
  try {
    if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      return "dark";
    }
    return "light";
  } catch {
    return "light";
  }
}

export function applyThemeClass(mode: ThemeChoice): void {
  if (typeof document === "undefined") {
    return;
  }
  const root = document.documentElement;
  if (mode === "dark") {
    root.classList.add("dark");
  } else {
    root.classList.remove("dark");
  }
}

export function persistTheme(mode: ThemeChoice): void {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, mode);
  } catch {
    /* ignore */
  }
}

/** Whether the app UI is using dark surfaces (class on html). */
export function isDarkDom(): boolean {
  if (typeof document === "undefined") {
    return false;
  }
  return document.documentElement.classList.contains("dark");
}
