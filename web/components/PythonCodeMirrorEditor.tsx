"use client";

// CodeMirror 6 via `codemirror` + `@codemirror/view` `EditorView` (architecture §4.13).

import { indentWithTab } from "@codemirror/commands";
import {
  Compartment,
  EditorSelection,
  EditorState,
  type Extension,
} from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { basicSetup } from "codemirror";
import { oneDark } from "@codemirror/theme-one-dark";
import { useEffect, useRef, type MutableRefObject } from "react";
import { useUiTheme } from "@/components/UiThemeContext";
import type { ThemeChoice } from "@/lib/theme";

type Props = {
  value: string;
  onChange: (value: string) => void;
  extensions?: Extension[];
  className?: string;
  /** REQ-024: parent holds ref for “insert example” into the live buffer. */
  editorViewRef?: MutableRefObject<EditorView | null>;
};

function themeExtension(theme: ThemeChoice): Extension {
  return theme === "dark" ? oneDark : [];
}

export function PythonCodeMirrorEditor(props: Props) {
  const { value, onChange, extensions = [], className, editorViewRef } = props;
  const theme = useUiTheme();
  const parentRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const themeCompartmentRef = useRef(new Compartment());

  useEffect(() => {
    const parent = parentRef.current;
    if (!parent) {
      return;
    }
    const state = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        themeCompartmentRef.current.of(themeExtension(theme)),
        keymap.of([indentWithTab]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onChangeRef.current(update.state.doc.toString());
          }
        }),
        EditorView.theme({
          "&": {
            height: "min(50vh, 420px)",
            maxHeight: "50vh",
          },
          ".cm-scroller": {
            minHeight: "min(50vh, 420px)",
          },
          ".cm-editor.cm-focused": {
            outline: "none",
          },
        }),
        ...extensions,
      ],
    });
    const view = new EditorView({ state, parent });
    viewRef.current = view;
    if (editorViewRef) {
      editorViewRef.current = view;
    }
    return () => {
      view.destroy();
      viewRef.current = null;
      if (editorViewRef) {
        editorViewRef.current = null;
      }
    };
    // Intentionally run once on mount; `value` sync and theme reconfigure are handled below.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view) {
      return;
    }
    view.dispatch({
      effects: themeCompartmentRef.current.reconfigure(themeExtension(theme)),
    });
  }, [theme]);

  useEffect(() => {
    const view = viewRef.current;
    if (!view) {
      return;
    }
    const cur = view.state.doc.toString();
    if (cur !== value) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: value },
        selection: EditorSelection.cursor(value.length),
        scrollIntoView: true,
      });
    }
  }, [value]);

  return (
    <div
      ref={parentRef}
      className={
        className ??
        "overflow-hidden rounded-lg border border-slate-300 dark:border-slate-600"
      }
    />
  );
}
