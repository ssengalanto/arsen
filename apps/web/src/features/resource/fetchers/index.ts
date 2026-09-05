import { apiFetch, apiPost } from "@/lib/api/client";
import { resourcesKey } from "@/lib/swr/keys";
import type { CreateResourceInput, Resource, UpdateResourceInput } from "../types";

// Reads flow fetcher → hook → component. Every call hits a relative BFF URL;
// the browser never talks to the Go backend directly.

export function listResources(): Promise<Resource[]> {
  return apiFetch(resourcesKey);
}

export function getResource(id: string): Promise<Resource> {
  return apiFetch(`${resourcesKey}/${encodeURIComponent(id)}`);
}

export function createResource(input: CreateResourceInput): Promise<Resource> {
  return apiPost(resourcesKey, input);
}

export function updateResource(id: string, input: UpdateResourceInput): Promise<Resource> {
  return apiFetch(`${resourcesKey}/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteResource(id: string): Promise<void> {
  return apiFetch(`${resourcesKey}/${encodeURIComponent(id)}`, { method: "DELETE" });
}
