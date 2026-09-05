"use client";

import { useResource } from "../hooks/use-resource";
import { ResourceCard } from "./resource-card";

// Detail consumer. Hydrates instantly from the RSC-provided SWR fallback, then
// revalidates. Renders the same card used in the list for write actions.
export function ResourceDetail({ id }: { id: string }) {
  const { resource, error, isLoading } = useResource(id);

  if (isLoading) {
    return <div className="h-24 animate-pulse rounded-xl bg-muted" aria-busy="true" />;
  }
  if (error || !resource) {
    return (
      <p className="text-destructive text-sm" role="alert">
        Could not load this resource.
      </p>
    );
  }

  return <ResourceCard resource={resource} />;
}
