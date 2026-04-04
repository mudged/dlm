/** @vitest-environment jsdom */
import { describe, expect, it, beforeEach, vi } from "vitest";
import {
  THEME_STORAGE_KEY,
  applyThemeClass,
  isDarkDom,
  persistTheme,
  readStoredTheme,
  resolveInitialTheme,
} from "./theme";

function mockMatchMediaDark(matches: boolean) {
  vi.stubGlobal(
    "matchMedia",
    vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  );
}

describe("theme helpers", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark");
    vi.unstubAllGlobals();
  });

  it("readStoredTheme returns null when empty", () => {
    expect(readStoredTheme()).toBeNull();
  });

  it("persistTheme and readStoredTheme round-trip", () => {
    persistTheme("dark");
    expect(readStoredTheme()).toBe("dark");
    persistTheme("light");
    expect(readStoredTheme()).toBe("light");
  });

  it("applyThemeClass toggles dark class", () => {
    applyThemeClass("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    expect(isDarkDom()).toBe(true);
    applyThemeClass("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(isDarkDom()).toBe(false);
  });

  it("uses expected storage key", () => {
    expect(THEME_STORAGE_KEY).toBe("dlm-theme");
  });

  it("resolveInitialTheme uses stored light or dark when set", () => {
    persistTheme("dark");
    expect(resolveInitialTheme(readStoredTheme())).toBe("dark");
    persistTheme("light");
    expect(resolveInitialTheme(readStoredTheme())).toBe("light");
  });

  it("resolveInitialTheme follows prefers-color-scheme when storage unset", () => {
    mockMatchMediaDark(true);
    expect(resolveInitialTheme(null)).toBe("dark");
    mockMatchMediaDark(false);
    expect(resolveInitialTheme(null)).toBe("light");
  });

  it("resolveInitialTheme defaults to light when matchMedia throws", () => {
    vi.stubGlobal("matchMedia", () => {
      throw new Error("unsupported");
    });
    expect(resolveInitialTheme(null)).toBe("light");
  });
});
