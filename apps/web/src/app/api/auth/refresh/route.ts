import { NextResponse } from "next/server";
import {
  clearSessionCookies,
  getRefreshToken,
  setSessionCookies,
  type SessionTokens,
} from "@/lib/auth/cookies";
import { UPSTREAM_URL, relayError } from "@/lib/api/bff";

// POST /api/auth/refresh → Go POST /api/tokens. Reads the refresh cookie,
// rotates both cookies on success. On a terminal 401 the cookies are cleared.
export async function POST(): Promise<NextResponse> {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) {
    await clearSessionCookies();
    return new NextResponse(null, { status: 401 });
  }

  const res = await fetch(`${UPSTREAM_URL}/api/tokens`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ refreshToken }),
    cache: "no-store",
  });
  if (!res.ok) {
    if (res.status === 401) await clearSessionCookies();
    return relayError(res);
  }

  const session = (await res.json()) as SessionTokens;
  await setSessionCookies(session);
  return new NextResponse(null, { status: 204 });
}
