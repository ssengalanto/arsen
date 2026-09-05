import { NextResponse } from "next/server";
import { problemResponse } from "@/lib/api/bff";
import { createResource, listResources } from "@/lib/resource-store";
import { resourceSchema, type ResourceStatus } from "@/features/resource";
import type { ProblemFieldError } from "@/lib/api/error";

const STATUSES: readonly ResourceStatus[] = ["draft", "active", "archived"];

function parseStatus(value: string | null): ResourceStatus | undefined {
  return value && (STATUSES as readonly string[]).includes(value)
    ? (value as ResourceStatus)
    : undefined;
}

// GET /api/resources[?status=] — list from the stand-in store.
export function GET(request: Request): NextResponse {
  const status = parseStatus(new URL(request.url).searchParams.get("status"));
  return NextResponse.json(listResources(status));
}

// POST /api/resources — validate against the shared schema, then create.
// Validation failures mirror the Go backend's problem+json `errors[]` shape.
export async function POST(request: Request): Promise<NextResponse> {
  const body: unknown = await request.json();
  const parsed = resourceSchema.safeParse(body);
  if (!parsed.success) {
    const errors: ProblemFieldError[] = parsed.error.issues.map((issue) => ({
      field: issue.path.join("."),
      detail: issue.message,
    }));
    return problemResponse(400, "Validation failed", "The request body is invalid.", errors);
  }

  return NextResponse.json(createResource(parsed.data), { status: 201 });
}
