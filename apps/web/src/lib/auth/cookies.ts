import { cookies } from "next/headers";

export const ACCESS_COOKIE = "arsen_access";
export const REFRESH_COOKIE = "arsen_refresh";

const DEFAULT_REFRESH_MAX_AGE = 60 * 60 * 24 * 30; // 30 days

function refreshMaxAge(): number {
  const parsed = Number.parseInt(process.env.SESSION_REFRESH_MAX_AGE ?? "", 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : DEFAULT_REFRESH_MAX_AGE;
}

export interface SessionTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export async function setSessionCookies(tokens: SessionTokens): Promise<void> {
  const store = await cookies();
  const base = {
    httpOnly: true,
    sameSite: "lax" as const,
    path: "/",
    secure: process.env.NODE_ENV === "production",
  };
  store.set(ACCESS_COOKIE, tokens.accessToken, { ...base, maxAge: tokens.expiresIn });
  store.set(REFRESH_COOKIE, tokens.refreshToken, { ...base, maxAge: refreshMaxAge() });
}

export async function clearSessionCookies(): Promise<void> {
  const store = await cookies();
  store.delete(ACCESS_COOKIE);
  store.delete(REFRESH_COOKIE);
}

export async function getAccessToken(): Promise<string | undefined> {
  return (await cookies()).get(ACCESS_COOKIE)?.value;
}

export async function getRefreshToken(): Promise<string | undefined> {
  return (await cookies()).get(REFRESH_COOKIE)?.value;
}
