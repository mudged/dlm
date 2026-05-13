import { describe, expect, it } from "vitest";
import { FACTORY_RESET_DISCLOSURE, POST_RESET_FLASH } from "./factoryReset";

describe("FACTORY_RESET_DISCLOSURE (REQ-017 BR-2)", () => {
  it("mentions registered devices", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("devices");
  });

  it("mentions routines", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("routines");
  });

  it("mentions sample Python routines", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("sample Python routines");
  });

  it("mentions every model you uploaded", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("every model you uploaded");
  });

  it("mentions every scene", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("every scene");
  });

  it("mentions in-flight routine runs being stopped", () => {
    expect(FACTORY_RESET_DISCLOSURE).toContain("in-flight routine runs");
  });
});

describe("POST_RESET_FLASH (REQ-017 BR-2)", () => {
  it("mentions sample models", () => {
    expect(POST_RESET_FLASH).toContain("sample models");
  });

  it("mentions sample Python routines", () => {
    expect(POST_RESET_FLASH).toContain("sample Python routines");
  });
});
