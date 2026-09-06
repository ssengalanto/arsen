import { NextResponse } from "next/server";
import { setSessionCookies, type SessionTokens } from "@/lib/auth/cookies";
import { UPSTREAM_URL, relayError } from "@/lib/api/bff";
import type { UserProfile } from "@/features/auth";

interface LoginBody {
  email: string;
  password: string;
}

// POST /api/auth/login → Go POST /api/sessions. On success the token pair is
// converted to httpOnly cookies and only the profile is returned — never a token.
export async function POST(request: Request): Promise<NextResponse> {
  const body = (await request.json()) as LoginBody;

  const res = await fetch(`${UPSTREAM_URL}/api/sessions`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  if (!res.ok) return relayError(res);

  const session = (await res.json()) as SessionTokens;
  await setSessionCookies(session);

  const meRes = await fetch(`${UPSTREAM_URL}/api/users/me`, {
    headers: { authorization: `Bearer ${session.accessToken}` },
    cache: "no-store",
  });
  const profile = (await meRes.json()) as UserProfile;
  const user: UserProfile = {
    id: profile.id,
    email: profile.email,
    emailVerified: profile.emailVerified,
    createdAt: profile.createdAt,
  };
  return NextResponse.json({ user });
}
