import { EditorSelection, type EditorState, type Transaction } from "@codemirror/state";
import type { EditorView } from "@codemirror/view";

/**
 * REQ-024: when focused, insert at the active selection (replacing a non-empty
 * selection). Otherwise append at the end of the document.
 */
export function buildInsertSnippetTransaction(
  state: EditorState,
  hasFocus: boolean,
  text: string,
): Transaction {
  if (hasFocus) {
    const { from, to } = state.selection.main;
    return state.update({
      changes: { from, to, insert: text },
      selection: EditorSelection.cursor(from + text.length),
      scrollIntoView: true,
    });
  }
  const len = state.doc.length;
  const needsLeadNewline = len > 0 && state.doc.sliceString(len - 1, len) !== "\n";
  const insert = (needsLeadNewline ? "\n" : "") + text;
  return state.update({
    changes: { from: len, to: len, insert },
    selection: EditorSelection.cursor(len + insert.length),
    scrollIntoView: true,
  });
}

export function insertSnippetInPythonEditor(view: EditorView, text: string): void {
  view.dispatch(buildInsertSnippetTransaction(view.state, view.hasFocus, text));
}
