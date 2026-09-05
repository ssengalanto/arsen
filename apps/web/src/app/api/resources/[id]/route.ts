import { NextResponse } from "next/server";
import { problemResponse } from "@/lib/api/bff";
import { deleteResource, getResource, updateResource } from "@/lib/resource-store";
import { resourceSchema } from "@/features/resource";
import type { ProblemFieldError } from "@/lib/api/error";

type Params = { params: Promise<{ id: string }> };

const notFound = () => problemResponse(404, "Not found", "No resource with that id exists.");

// GET /api/resources/:id
export async function GET(_request: Request, { params }: Params): Promise<NextResponse> {
  const { id } = await params;
  const resource = getResource(id);
  return resource ? NextResponse.json(resource) : notFound();
}

// PATCH /api/resources/:id — partial update, validated against the shared schema.
export async function PATCH(request: Request, { params }: Params): Promise<NextResponse> {
  const { id } = await params;
  const body: unknown = await request.json();
  const parsed = resourceSchema.partial().safeParse(body);
  if (!parsed.success) {
    const errors: ProblemFieldError[] = parsed.error.issues.map((issue) => ({
      field: issue.path.join("."),
      detail: issue.message,
    }));
    return problemResponse(400, "Validation failed", "The request body is invalid.", errors);
  }

  const updated = updateResource(id, parsed.data);
  return updated ? NextResponse.json(updated) : notFound();
}

// DELETE /api/resources/:id
export async function DELETE(_request: Request, { params }: Params): Promise<NextResponse> {
  const { id } = await params;
  return deleteResource(id) ? new NextResponse(null, { status: 204 }) : notFound();
}
