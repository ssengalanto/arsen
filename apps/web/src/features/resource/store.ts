import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { CreateResourceInput, ResourceStatus } from "./types";

type StatusFilter = ResourceStatus | "all";

interface ResourceFilters {
  status: StatusFilter;
}

interface ResourceState {
  // UI-only, ephemeral — never persisted.
  selectedIds: string[];
  wizardStep: number;
  // Persisted — should survive a reload.
  filters: ResourceFilters;
  draft: Partial<CreateResourceInput> | null;

  toggleSelected: (id: string) => void;
  clearSelection: () => void;
  setStatusFilter: (status: StatusFilter) => void;
  setDraft: (draft: Partial<CreateResourceInput>) => void;
  clearDraft: () => void;
  setWizardStep: (step: number) => void;
}

// UI/client state only. Server data lives in SWR, never here. `persist` keeps
// just `filters` + `draft` so a half-filled form survives a reload, while
// selection and wizard position reset each session.
export const useResourceStore = create<ResourceState>()(
  persist(
    (set) => ({
      selectedIds: [],
      wizardStep: 0,
      filters: { status: "all" },
      draft: null,

      toggleSelected: (id) =>
        set((state) => ({
          selectedIds: state.selectedIds.includes(id)
            ? state.selectedIds.filter((existing) => existing !== id)
            : [...state.selectedIds, id],
        })),
      clearSelection: () => set({ selectedIds: [] }),
      setStatusFilter: (status) => set({ filters: { status } }),
      setDraft: (draft) => set({ draft }),
      clearDraft: () => set({ draft: null }),
      setWizardStep: (wizardStep) => set({ wizardStep }),
    }),
    {
      name: "arsen-resource",
      partialize: (state) => ({ filters: state.filters, draft: state.draft }),
    },
  ),
);
