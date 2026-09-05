import { NextResponse } from "next/server";
import { serverFetchJson } from "@/lib/api/server";
import { apiErrorResponse } from "@/lib/api/bff";
import type { UserProfile } from "@/features/auth";

// GET /api/me → Go GET /api/users/me with silent-refresh-on-401 (one retry).
// On a terminal 401 the server helper clears the cookies and throws; we
// re-emit that as problem+json.
export async function GET(): Promise<NextResponse> {
  try {
    const profile = await serverFetchJson<UserProfile>("/api/users/me");
    return NextResponse.json({
      id: profile.id,
      email: profile.email,
      emailVerified: profile.emailVerified,
      createdAt: profile.createdAt,
    });
  } catch (err) {
    return apiErrorResponse(err);
  }
}
