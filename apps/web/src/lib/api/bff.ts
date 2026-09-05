import { NextResponse } from "next/server";
import { ApiError, type ProblemFieldError } from "./error";

export const UPSTREAM_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Build an RFC 9457 problem+json response from scratch. Used by BFF routes that
// answer directly (e.g. the resource stand-in) rather than relaying upstream.
export function problemResponse(
  status: number,
  title: string,
  detail?: string,
  errors?: ProblemFieldError[],
): NextResponse {
  return NextResponse.json(
    { title, status, ...(detail ? { detail } : {}), ...(errors?.length ? { errors } : {}) },
    { status, headers: { "content-type": "application/problem+json" } },
  );
}

// Re-emit an upstream error Response verbatim, preserving the problem+json body
// and status so the client `ApiError` and error boundary work uniformly.
export async function relayError(res: Response): Promise<NextResponse> {
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: {
      "content-type": res.headers.get("content-type") ?? "application/problem+json",
    },
  });
}

// Turn a thrown `ApiError` (from server.ts silent-refresh flow) back into a
// problem+json response. Re-throws anything that isn't an ApiError.
export function apiErrorResponse(err: unknown): NextResponse {
  if (err instanceof ApiError) {
    return NextResponse.json(
      {
        type: err.type,
        title: err.title,
        status: err.status,
        detail: err.detail,
        instance: err.instance,
        errors: err.errors,
      },
      {
        status: err.status,
        headers: { "content-type": "application/problem+json" },
      },
    );
  }
  throw err;
}
