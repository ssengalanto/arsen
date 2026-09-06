import { apiFetch, apiPost } from "@/lib/api/client";
import type { LoginInput } from "../schemas/login";
import type { RegisterInput } from "../schemas/register";
import type { UserProfile } from "../types";

// Reads flow fetcher → hook → component. Every call hits a relative BFF URL;
// the browser never talks to the Go backend directly.

export function login(input: LoginInput): Promise<{ user: UserProfile }> {
  return apiPost("/api/auth/login", input);
}

export function register(input: RegisterInput): Promise<{ message: string }> {
  return apiPost("/api/auth/register", input);
}

export function logout(): Promise<void> {
  return apiFetch("/api/auth/logout", { method: "DELETE" });
}

export function me(): Promise<UserProfile> {
  return apiFetch("/api/me");
}
