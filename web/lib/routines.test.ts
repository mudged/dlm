import { describe, expect, it } from "vitest";
import {
  ROUTINE_TYPE_CREATE_OPTIONS,
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_RANDOM_COLOUR_ALL,
  isCreatableRoutineType,
} from "./routines";

describe("routine types (REQ-023)", () => {
  it("lists built-in and Python as creatable options", () => {
    const values = ROUTINE_TYPE_CREATE_OPTIONS.map((o) => o.value);
    expect(values).toContain(ROUTINE_TYPE_RANDOM_COLOUR_ALL);
    expect(values).toContain(ROUTINE_TYPE_PYTHON_SCENE_SCRIPT);
    expect(ROUTINE_TYPE_CREATE_OPTIONS.length).toBe(2);
  });

  it("isCreatableRoutineType accepts only listed types", () => {
    expect(isCreatableRoutineType(ROUTINE_TYPE_RANDOM_COLOUR_ALL)).toBe(true);
    expect(isCreatableRoutineType(ROUTINE_TYPE_PYTHON_SCENE_SCRIPT)).toBe(
      true,
    );
    expect(isCreatableRoutineType("unknown_future_type")).toBe(false);
  });
});
