import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { userProfileFixture, problem } from "@/test/msw/fixtures";
import { ApiError } from "@/lib/api/error";
import { login, register, logout, me } from "./index";

beforeEach(() => {
  server.resetHandlers();
});

describe("auth fetchers", () => {
  it("login POSTs credentials and returns the profile envelope", async () => {
    await expect(login({ email: "a@b.com", password: "secret123" })).resolves.toEqual({
      user: userProfileFixture,
    });
  });

  it("register POSTs and returns the anti-enumeration message", async () => {
    await expect(register({ email: "a@b.com", password: "secret123" })).resolves.toHaveProperty(
      "message",
    );
  });

  it("logout DELETEs and resolves void on 204", async () => {
    await expect(logout()).resolves.toBeUndefined();
  });

  it("me GETs the current profile", async () => {
    await expect(me()).resolves.toEqual(userProfileFixture);
  });

  it("throws an ApiError when the profile request is unauthorized", async () => {
    server.use(
      http.get("/api/me", () => HttpResponse.json(problem(401, "unauthorized"), { status: 401 })),
    );
    await expect(me()).rejects.toBeInstanceOf(ApiError);
  });
});
