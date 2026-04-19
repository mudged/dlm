import { describe, expect, it } from "vitest";
import { apiOriginForEventSource, eventSourceUrl } from "./sseUrl";

describe("eventSourceUrl", () => {
  it("keeps leading slash and uses relative path when origin is empty", () => {
    expect(eventSourceUrl("/api/v1/scenes/a/lights/events")).toBe(
      "/api/v1/scenes/a/lights/events",
    );
  });

});

describe("apiOriginForEventSource", () => {
  it("returns empty when not in development and no explicit origin", () => {
    expect(apiOriginForEventSource()).toBe("");
  });
});
