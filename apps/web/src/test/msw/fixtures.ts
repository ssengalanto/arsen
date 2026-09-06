export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export const userProfileFixture = {
  id: "11111111-1111-1111-1111-111111111111",
  email: "verified@example.com",
  emailVerified: true,
  createdAt: "2026-01-01T00:00:00Z",
};

export const sessionResponseFixture = {
  self: "/api/sessions",
  kind: "Session",
  accessToken: "access-token-jwt",
  refreshToken: "refresh-token-opaque",
  tokenType: "Bearer",
  expiresIn: 900,
};

export const validationProblemFixture = {
  type: "https://arsen.dev/problems/validation",
  title: "Validation Failed",
  status: 400,
  detail: "The request body is invalid.",
  instance: "/api/sessions",
  errors: [{ field: "email", detail: "is required" }],
};

export function problem(status: number, detail: string) {
  return {
    type: "about:blank",
    title: detail,
    status,
    detail,
    instance: "/",
  };
}
