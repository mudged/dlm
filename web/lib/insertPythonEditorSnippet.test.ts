import { EditorSelection, EditorState } from "@codemirror/state";
import { describe, expect, it } from "vitest";
import { buildInsertSnippetTransaction } from "./insertPythonEditorSnippet";

describe("buildInsertSnippetTransaction", () => {
  it("appends at end with leading newline when not focused", () => {
    const state = EditorState.create({ doc: "a" });
    const tr = buildInsertSnippetTransaction(state, false, "b");
    expect(tr.state.doc.toString()).toBe("a\nb");
  });

  it("appends without extra newline when document already ends with newline", () => {
    const state = EditorState.create({ doc: "a\n" });
    const tr = buildInsertSnippetTransaction(state, false, "b");
    expect(tr.state.doc.toString()).toBe("a\nb");
  });

  it("replaces selection when focused", () => {
    const state = EditorState.create({
      doc: "hello",
      selection: EditorSelection.create([EditorSelection.range(1, 4)]),
    });
    const tr = buildInsertSnippetTransaction(state, true, "XX");
    expect(tr.state.doc.toString()).toBe("hXXo");
  });
});
