import { describe, expect, it } from "vitest";
import {
  ROUTINE_TYPE_PYTHON_SCENE_SCRIPT,
  ROUTINE_TYPE_RANDOM_COLOUR_ALL,
  ROUTINE_TYPE_SHAPE_ANIMATION,
  SceneRoutineConflictError,
  isCreatableRoutineType,
} from "./routines";

describe("routine types (REQ-023)", () => {
  it("isCreatableRoutineType accepts python and shape_animation only", () => {
    expect(isCreatableRoutineType(ROUTINE_TYPE_PYTHON_SCENE_SCRIPT)).toBe(true);
    expect(isCreatableRoutineType(ROUTINE_TYPE_SHAPE_ANIMATION)).toBe(true);
    expect(isCreatableRoutineType(ROUTINE_TYPE_RANDOM_COLOUR_ALL)).toBe(false);
    expect(isCreatableRoutineType("unknown_future_type")).toBe(false);
  });
});

describe("SceneRoutineConflictError", () => {
  it("is instanceof Error", () => {
    const e = new SceneRoutineConflictError(
      "msg",
      "sc1",
      "req-rid",
      "run1",
      "exist-rid",
    );
    expect(e).toBeInstanceOf(Error);
    expect(e.existingRunId).toBe("run1");
    expect(e.code).toBe("scene_routine_conflict");
  });
});
