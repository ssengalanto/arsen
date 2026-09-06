import { NextResponse } from "next/server";
import { UPSTREAM_URL, relayError } from "@/lib/api/bff";

interface RegisterBody {
  email: string;
  password: string;
}

// POST /api/auth/register → Go POST /api/users. Registration never
// authenticates: no cookies are set and no tokens are returned. The client
// redirects to /login?registered=1 with a verify-your-email notice.
export async function POST(request: Request): Promise<NextResponse> {
  const body = (await request.json()) as RegisterBody;

  const res = await fetch(`${UPSTREAM_URL}/api/users`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  if (!res.ok) return relayError(res);

  return NextResponse.json(
    { message: "Check your email to verify your account." },
    { status: 201 },
  );
}
