import { NextResponse } from "next/server";
import { clearSessionCookies, getAccessToken } from "@/lib/auth/cookies";
import { UPSTREAM_URL } from "@/lib/api/bff";

// DELETE /api/auth/logout → Go DELETE /api/sessions/current. The cookies are
// ALWAYS cleared afterward, even if the upstream revoke fails (FR-004 ordering).
export async function DELETE(): Promise<NextResponse> {
  const access = await getAccessToken();
  if (access) {
    try {
      await fetch(`${UPSTREAM_URL}/api/sessions/current`, {
        method: "DELETE",
        headers: { authorization: `Bearer ${access}` },
        cache: "no-store",
      });
    } catch {
      // Upstream failure must not block local cookie cleanup.
    }
  }

  await clearSessionCookies();
  return new NextResponse(null, { status: 204 });
}
