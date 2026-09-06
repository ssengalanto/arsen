import { http, HttpResponse } from "msw";
import {
  API_URL,
  sessionResponseFixture,
  userProfileFixture,
  problem,
} from "./fixtures";

// Default happy-path handlers. Individual tests override with `server.use(...)`.
export const handlers = [
  // Upstream Go backend
  http.post(`${API_URL}/api/sessions`, () => HttpResponse.json(sessionResponseFixture)),
  http.post(`${API_URL}/api/tokens`, () =>
    HttpResponse.json({ ...sessionResponseFixture, kind: "TokenPair" }),
  ),
  http.delete(`${API_URL}/api/sessions/current`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${API_URL}/api/users`, () =>
    HttpResponse.json(
      { ...userProfileFixture, emailVerified: false, kind: "User" },
      { status: 201 },
    ),
  ),
  http.get(`${API_URL}/api/users/me`, () => HttpResponse.json(userProfileFixture)),

  // BFF route handlers (relative — resolved against the test origin)
  http.post("/api/auth/login", () => HttpResponse.json({ user: userProfileFixture })),
  http.post("/api/auth/register", () =>
    HttpResponse.json({ message: "Check your email to verify your account." }, { status: 201 }),
  ),
  http.post("/api/auth/refresh", () => new HttpResponse(null, { status: 204 })),
  http.delete("/api/auth/logout", () => new HttpResponse(null, { status: 204 })),
  http.get("/api/me", () => HttpResponse.json(userProfileFixture)),
  http.get("/api/resources", () => HttpResponse.json([])),
];

export { problem };
