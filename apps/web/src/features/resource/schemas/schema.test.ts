import { describe, it, expect } from "vitest";
import { resourceSchema } from "./resource";

describe("resourceSchema", () => {
  it("accepts a valid resource with all fields", () => {
    const result = resourceSchema.safeParse({
      title: "Design doc",
      description: "A short description.",
      status: "active",
    });
    expect(result.success).toBe(true);
  });

  it("accepts a missing description (optional)", () => {
    expect(resourceSchema.safeParse({ title: "No desc", status: "draft" }).success).toBe(true);
  });

  it("rejects an empty title", () => {
    expect(resourceSchema.safeParse({ title: "", status: "draft" }).success).toBe(false);
  });

  it("rejects a title longer than 120 characters", () => {
    expect(
      resourceSchema.safeParse({ title: "a".repeat(121), status: "draft" }).success,
    ).toBe(false);
  });

  it("rejects a description longer than 2000 characters", () => {
    expect(
      resourceSchema.safeParse({
        title: "ok",
        description: "a".repeat(2001),
        status: "draft",
      }).success,
    ).toBe(false);
  });

  it("rejects an unknown status", () => {
    expect(resourceSchema.safeParse({ title: "ok", status: "nope" }).success).toBe(false);
  });
});
