import { describe, it, expect, beforeEach, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { API_URL, userProfileFixture } from "@/test/msw/fixtures";

const store = new Map<string, string>();

vi.mock("@/lib/auth/cookies", () => ({
  ACCESS_COOKIE: "arsen_access",
  REFRESH_COOKIE: "arsen_refresh",
  getAccessToken: () => Promise.resolve(store.get("access")),
  getRefreshToken: () => Promise.resolve(store.get("refresh")),
  setSessionCookies: (t: { accessToken: string; refreshToken: string }) => {
    store.set("access", t.accessToken);
    store.set("refresh", t.refreshToken);
    return Promise.resolve();
  },
  clearSessionCookies: () => {
    store.clear();
    return Promise.resolve();
  },
}));

import { serverFetch } from "./server";

beforeEach(() => {
  store.clear();
  store.set("access", "stale-access");
  store.set("refresh", "valid-refresh");
});

describe("serverFetch silent refresh", () => {
  it("refreshes once on 401, rewrites cookies, then retries successfully", async () => {
    let attempts = 0;
    server.use(
      http.get(`${API_URL}/api/users/me`, ({ request }) => {
        attempts += 1;
        if (request.headers.get("authorization") === "Bearer fresh-access") {
          return HttpResponse.json(userProfileFixture);
        }
        return HttpResponse.json({ title: "Unauthorized", status: 401 }, { status: 401 });
      }),
      http.post(`${API_URL}/api/tokens`, () =>
        HttpResponse.json({
          accessToken: "fresh-access",
          refreshToken: "rotated-refresh",
          expiresIn: 900,
        }),
      ),
    );

    const res = await serverFetch("/api/users/me");
    expect(res.status).toBe(200);
    expect(attempts).toBe(2);
    expect(store.get("access")).toBe("fresh-access");
    expect(store.get("refresh")).toBe("rotated-refresh");
  });

  it("clears cookies and surfaces 401 when refresh also fails", async () => {
    server.use(
      http.get(`${API_URL}/api/users/me`, () =>
        HttpResponse.json({ title: "Unauthorized", status: 401 }, { status: 401 }),
      ),
      http.post(`${API_URL}/api/tokens`, () =>
        HttpResponse.json({ title: "Unauthorized", status: 401 }, { status: 401 }),
      ),
    );

    await expect(serverFetch("/api/users/me")).rejects.toMatchObject({ status: 401 });
    expect(store.size).toBe(0);
  });
});
