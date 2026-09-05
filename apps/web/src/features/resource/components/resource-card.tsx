"use client";

import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useUpdateResource } from "../hooks/use-update-resource";
import { useDeleteResource } from "../hooks/use-delete-resource";
import type { Resource, ResourceStatus } from "../types";

const STATUS_LABEL: Record<ResourceStatus, string> = {
  draft: "Draft",
  active: "Active",
  archived: "Archived",
};

// A single resource row. Demonstrates the write path: the status toggle calls
// the update hook, the delete button calls the delete hook — both reconcile the
// SWR caches themselves.
export function ResourceCard({ resource }: { resource: Resource }) {
  const { update, isUpdating } = useUpdateResource();
  const { remove, isDeleting } = useDeleteResource();

  const nextStatus: ResourceStatus = resource.status === "archived" ? "active" : "archived";

  async function toggleStatus() {
    try {
      await update(resource.id, { status: nextStatus });
    } catch {
      toast.error("Could not update the resource");
    }
  }

  async function onDelete() {
    try {
      await remove(resource.id);
      toast.success("Resource deleted");
    } catch {
      toast.error("Could not delete the resource");
    }
  }

  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle>{resource.title}</CardTitle>
        {resource.description && <CardDescription>{resource.description}</CardDescription>}
        <CardAction>
          <span className="rounded-md border px-2 py-0.5 text-xs text-muted-foreground">
            {STATUS_LABEL[resource.status]}
          </span>
        </CardAction>
      </CardHeader>
      <CardContent className="sr-only">{resource.id}</CardContent>
      <CardFooter className="gap-2">
        <Button size="sm" variant="outline" disabled={isUpdating} onClick={() => void toggleStatus()}>
          {resource.status === "archived" ? "Activate" : "Archive"}
        </Button>
        <Button
          size="sm"
          variant="destructive"
          disabled={isDeleting}
          onClick={() => void onDelete()}
        >
          Delete
        </Button>
      </CardFooter>
    </Card>
  );
}
