import { describe, expect, it } from "vitest";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_RANDOM_COLOUR_ALL,
  isCreatableRoutineType,
} from "./routines";

describe("routine types (REQ-023)", () => {
  it("isCreatableRoutineType accepts only python_scene_script", () => {
    expect(isCreatableRoutineType(ROUTINE_TYPE_PYTHON_SCENE_SCRIPT)).toBe(true);
    expect(isCreatableRoutineType(ROUTINE_TYPE_RANDOM_COLOUR_ALL)).toBe(false);
    expect(isCreatableRoutineType("unknown_future_type")).toBe(false);
  });
});
