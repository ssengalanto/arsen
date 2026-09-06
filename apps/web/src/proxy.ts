import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE } from "@/lib/auth/cookies";

// Presence-based guard for the protected `(app)` route group. The server proxy
// is the authority for token validity — this only checks that a session cookie
// exists and redirects to /login otherwise. No signature verification (R-4).
export function proxy(request: NextRequest): NextResponse {
  const hasSession =
    request.cookies.has(ACCESS_COOKIE) || request.cookies.has(REFRESH_COOKIE);

  if (hasSession) return NextResponse.next();

  const loginUrl = new URL("/login", request.url);
  const { pathname, search } = request.nextUrl;
  loginUrl.searchParams.set("next", `${pathname}${search}`);
  return NextResponse.redirect(loginUrl);
}

export const config = {
  matcher: ["/dashboard/:path*", "/resources/:path*"],
};
