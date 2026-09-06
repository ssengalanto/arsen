import type { CreateResourceInput, Resource, ResourceStatus, UpdateResourceInput } from "@/features/resource";

// ---------------------------------------------------------------------------
// STAND-IN BACKEND — replace me.
//
// The Go API has no `resources` endpoint yet, so this module-level Map fakes
// one for the worked CRUD slice. It is process-local and resets on every server
// restart. When a real backend exists, delete this file and repoint the BFF
// routes in `app/api/resources/*` at it (proxy through `lib/api/server.ts`,
// exactly like the auth routes do).
// ---------------------------------------------------------------------------

const store = new Map<string, Resource>();

function seed() {
  if (store.size > 0) return;
  const now = new Date().toISOString();
  const rows: Resource[] = [
    {
      id: "seed-1",
      title: "Draft the onboarding guide",
      description: "Outline the steps a new engineer follows on day one.",
      status: "draft",
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "seed-2",
      title: "Ship the resource slice",
      description: "The worked CRUD example wiring SWR, Zustand, and the BFF.",
      status: "active",
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "seed-3",
      title: "Archive the legacy prototype",
      description: "Old spike kept only for reference.",
      status: "archived",
      createdAt: now,
      updatedAt: now,
    },
  ];
  for (const row of rows) store.set(row.id, row);
}

function newId(): string {
  return `res-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

export function listResources(status?: ResourceStatus): Resource[] {
  seed();
  const all = [...store.values()].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
  return status ? all.filter((r) => r.status === status) : all;
}

export function getResource(id: string): Resource | undefined {
  seed();
  return store.get(id);
}

export function createResource(input: CreateResourceInput): Resource {
  seed();
  const now = new Date().toISOString();
  const resource: Resource = {
    id: newId(),
    title: input.title,
    description: input.description ?? "",
    status: input.status,
    createdAt: now,
    updatedAt: now,
  };
  store.set(resource.id, resource);
  return resource;
}

export function updateResource(id: string, input: UpdateResourceInput): Resource | undefined {
  seed();
  const existing = store.get(id);
  if (!existing) return undefined;
  const updated: Resource = {
    ...existing,
    ...(input.title !== undefined ? { title: input.title } : {}),
    ...(input.description !== undefined ? { description: input.description } : {}),
    ...(input.status !== undefined ? { status: input.status } : {}),
    updatedAt: new Date().toISOString(),
  };
  store.set(id, updated);
  return updated;
}

export function deleteResource(id: string): boolean {
  seed();
  return store.delete(id);
}
