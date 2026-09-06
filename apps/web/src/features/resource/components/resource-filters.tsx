"use client";

import { Button } from "@/components/ui/button";
import { useResourceStore } from "../store";
import type { ResourceStatus } from "../types";

const FILTERS: { value: ResourceStatus | "all"; label: string }[] = [
  { value: "all", label: "All" },
  { value: "draft", label: "Draft" },
  { value: "active", label: "Active" },
  { value: "archived", label: "Archived" },
];

// Status filter. The selection lives in the persisted UI store (not SWR) so it
// survives a reload; `useResources` reads it to scope what it renders.
export function ResourceFilters() {
  const status = useResourceStore((s) => s.filters.status);
  const setStatusFilter = useResourceStore((s) => s.setStatusFilter);

  return (
    <div className="flex flex-wrap gap-1.5" role="group" aria-label="Filter by status">
      {FILTERS.map((filter) => (
        <Button
          key={filter.value}
          size="sm"
          variant={status === filter.value ? "default" : "outline"}
          aria-pressed={status === filter.value}
          onClick={() => setStatusFilter(filter.value)}
        >
          {filter.label}
        </Button>
      ))}
    </div>
  );
}
