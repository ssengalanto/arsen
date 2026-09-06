import { describe, it, expect } from "vitest";
import { NextRequest } from "next/server";
import { proxy } from "../../proxy";

function request(path: string, cookies: Record<string, string> = {}) {
  const req = new NextRequest(`http://localhost:3000${path}`);
  for (const [name, value] of Object.entries(cookies)) {
    req.cookies.set(name, value);
  }
  return req;
}

describe("proxy route protection", () => {
  it("redirects unauthenticated requests to /login with a next param", () => {
    const res = proxy(request("/dashboard"));
    expect(res.status).toBe(307);
    const location = res.headers.get("location");
    expect(location).not.toBeNull();
    const url = new URL(location!);
    expect(url.pathname).toBe("/login");
    expect(url.searchParams.get("next")).toBe("/dashboard");
  });

  it("preserves the full path and query in the next param", () => {
    const res = proxy(request("/resources/42?tab=details"));
    const url = new URL(res.headers.get("location")!);
    expect(url.pathname).toBe("/login");
    expect(url.searchParams.get("next")).toBe("/resources/42?tab=details");
  });

  it("allows the request through when a session cookie is present", () => {
    const res = proxy(request("/dashboard", { arsen_access: "token" }));
    expect(res.headers.get("location")).toBeNull();
  });

  it("allows access when only the refresh cookie is present", () => {
    const res = proxy(request("/dashboard", { arsen_refresh: "token" }));
    expect(res.headers.get("location")).toBeNull();
  });
});
