import { ApiError } from "./error";
import {
  clearSessionCookies,
  getAccessToken,
  getRefreshToken,
  setSessionCookies,
} from "@/lib/auth/cookies";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface SessionResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

function authedFetch(path: string, init: RequestInit, access: string | undefined): Promise<Response> {
  const headers = new Headers(init.headers);
  if (access) headers.set("Authorization", `Bearer ${access}`);
  return fetch(`${API_URL}${path}`, { ...init, headers, cache: "no-store" });
}

async function refreshSession(): Promise<boolean> {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) return false;

  const res = await fetch(`${API_URL}/api/tokens`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ refreshToken }),
    cache: "no-store",
  });
  if (!res.ok) return false;

  const session = (await res.json()) as SessionResponse;
  await setSessionCookies(session);
  return true;
}

// Server-side fetch to the Go backend. Attaches the access cookie as a Bearer
// token; on a 401 it silently refreshes once (rotating both cookies) and retries.
// A terminal 401 clears the session cookies. Non-2xx responses throw `ApiError`.
export async function serverFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const first = await authedFetch(path, init, await getAccessToken());
  if (first.status !== 401) {
    if (!first.ok) throw await ApiError.fromResponse(first);
    return first;
  }

  const refreshed = await refreshSession();
  if (!refreshed) {
    await clearSessionCookies();
    throw await ApiError.fromResponse(first);
  }

  const retry = await authedFetch(path, init, await getAccessToken());
  if (retry.status === 401) {
    await clearSessionCookies();
    throw await ApiError.fromResponse(retry);
  }
  if (!retry.ok) throw await ApiError.fromResponse(retry);
  return retry;
}

export async function serverFetchJson<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await serverFetch(path, init);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
