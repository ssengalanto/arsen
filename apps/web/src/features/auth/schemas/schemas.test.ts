import { describe, it, expect } from "vitest";
import { loginSchema } from "./login";
import { registerSchema } from "./register";

describe("loginSchema", () => {
  it("accepts a valid email and non-empty password", () => {
    const result = loginSchema.safeParse({ email: "a@b.com", password: "x" });
    expect(result.success).toBe(true);
  });

  it("rejects an invalid email", () => {
    expect(loginSchema.safeParse({ email: "nope", password: "x" }).success).toBe(false);
  });

  it("rejects an empty password", () => {
    expect(loginSchema.safeParse({ email: "a@b.com", password: "" }).success).toBe(false);
  });
});

describe("registerSchema", () => {
  it("accepts a valid email and an 8+ char password", () => {
    const result = registerSchema.safeParse({ email: "a@b.com", password: "supersecret" });
    expect(result.success).toBe(true);
  });

  it("rejects a password shorter than 8 characters", () => {
    expect(registerSchema.safeParse({ email: "a@b.com", password: "short" }).success).toBe(
      false,
    );
  });

  it("rejects an invalid email", () => {
    expect(registerSchema.safeParse({ email: "nope", password: "supersecret" }).success).toBe(
      false,
    );
  });
});
