"use client";

import { useUiStore } from "@/lib/stores/ui-store";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useResources } from "../hooks/use-resources";
import { useResourceStore } from "../store";
import { ResourceCard } from "./resource-card";
import { ResourceFilters } from "./resource-filters";
import { ResourceForm } from "./resource-form";

const MODAL = "resource-create";

// Top-level read UI for the slice: filter bar + create dialog + the list with
// its loading / error / empty states. Server data comes from `useResources`
// (SWR); the status filter comes from the persisted UI store.
export function ResourceList() {
  const { resources, error, isLoading } = useResources();
  const statusFilter = useResourceStore((s) => s.filters.status);
  const activeModal = useUiStore((s) => s.activeModal);
  const openModal = useUiStore((s) => s.openModal);
  const closeModal = useUiStore((s) => s.closeModal);

  const visible =
    statusFilter === "all"
      ? resources
      : resources.filter((resource) => resource.status === statusFilter);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <ResourceFilters />
        <Button size="sm" onClick={() => openModal(MODAL)}>
          New resource
        </Button>
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-2" aria-busy="true">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-20 animate-pulse rounded-xl bg-muted" />
          ))}
        </div>
      ) : error ? (
        <p className="text-destructive text-sm" role="alert">
          Could not load resources. Try again.
        </p>
      ) : visible.length === 0 ? (
        <p className="text-muted-foreground rounded-xl border border-dashed p-6 text-center text-sm">
          {statusFilter === "all"
            ? "No resources yet. Create your first one."
            : `No ${statusFilter} resources. Try a different filter.`}
        </p>
      ) : (
        <div className="flex flex-col gap-3">
          {visible.map((resource) => (
            <ResourceCard key={resource.id} resource={resource} />
          ))}
        </div>
      )}

      <Dialog
        open={activeModal === MODAL}
        onOpenChange={(open) => (open ? openModal(MODAL) : closeModal())}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New resource</DialogTitle>
            <DialogDescription>
              Fill in the details below. Your draft is saved automatically.
            </DialogDescription>
          </DialogHeader>
          <ResourceForm onCreated={closeModal} />
        </DialogContent>
      </Dialog>
    </div>
  );
}
