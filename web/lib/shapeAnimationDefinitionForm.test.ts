import { describe, expect, it } from "vitest";
import {
  definitionObjectFromForm,
  shapeFormStateFromDefault,
  shapeFormStateFromUnknown,
} from "./shapeAnimationDefinitionForm";
import { SHAPE_ANIMATION_DEFAULT_DEFINITION } from "./shapeAnimationDefault";

describe("shapeAnimationDefinitionForm", () => {
  it("round-trips default definition through form state", () => {
    const state = shapeFormStateFromUnknown(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    const obj = definitionObjectFromForm(state) as { version: number; shapes: unknown[] };
    expect(obj.version).toBe(1);
    expect(obj.shapes).toHaveLength(1);
  });

  it("defaultShapeFormStateFromDefault matches parse of constant", () => {
    const a = shapeFormStateFromDefault();
    const b = shapeFormStateFromUnknown(SHAPE_ANIMATION_DEFAULT_DEFINITION);
    expect(a.shapes.length).toBe(b.shapes.length);
    expect(a.backgroundMode).toBe(b.backgroundMode);
  });
});
