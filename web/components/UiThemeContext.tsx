"use client";

import { createContext, useContext } from "react";
import type { ThemeChoice } from "@/lib/theme";

export const UiThemeContext = createContext<ThemeChoice>("light");

export function useUiTheme(): ThemeChoice {
  return useContext(UiThemeContext);
}
