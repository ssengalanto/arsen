import { ApiError } from "./error";

// Browser-side fetch against the relative BFF routes. Cookies are httpOnly and
// travel automatically via `credentials: "include"`; this never talks to the Go
// backend directly.
export async function apiFetch<T>(input: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(input, {
    ...init,
    credentials: "include",
    headers: {
      "content-type": "application/json",
      ...init.headers,
    },
  });

  if (!res.ok) throw await ApiError.fromResponse(res);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export function apiPost<T>(input: string, body?: unknown): Promise<T> {
  return apiFetch<T>(input, {
    method: "POST",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}
