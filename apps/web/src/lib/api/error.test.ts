import { describe, it, expect } from "vitest";
import { ApiError } from "./error";

const problem = {
  type: "https://arsen.dev/problems/validation",
  title: "Validation Failed",
  status: 400,
  detail: "The request body is invalid.",
  instance: "/api/sessions",
  errors: [
    { field: "email", detail: "is required" },
    { field: "password", detail: "is too short" },
  ],
};

function problemResponse(body: unknown, status = 400): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/problem+json" },
  });
}

describe("ApiError", () => {
  it("parses a problem+json Response", async () => {
    const err = await ApiError.fromResponse(problemResponse(problem));
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(400);
    expect(err.title).toBe("Validation Failed");
    expect(err.detail).toBe("The request body is invalid.");
    expect(err.errors).toHaveLength(2);
  });

  it("exposes field errors as a map for form binding", async () => {
    const err = await ApiError.fromResponse(problemResponse(problem));
    expect(err.fieldErrors()).toEqual({ email: "is required", password: "is too short" });
  });

  it("falls back to a generic message when the body is not problem+json", async () => {
    const err = await ApiError.fromResponse(new Response("upstream exploded", { status: 502 }));
    expect(err.status).toBe(502);
    expect(err.message).toBeTruthy();
    expect(err.errors).toEqual([]);
  });

  it("never retains tokens or unknown fields from the body", async () => {
    const leaky = { ...problem, accessToken: "secret-jwt", refreshToken: "secret-opaque" };
    const err = await ApiError.fromResponse(problemResponse(leaky));
    const serialized = JSON.stringify({ ...err, message: err.message });
    expect(serialized).not.toContain("secret-jwt");
    expect(serialized).not.toContain("secret-opaque");
    expect((err as unknown as Record<string, unknown>).accessToken).toBeUndefined();
    expect((err as unknown as Record<string, unknown>).refreshToken).toBeUndefined();
  });
});
