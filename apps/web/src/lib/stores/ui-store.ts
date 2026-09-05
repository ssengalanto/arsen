import { create } from "zustand";

interface UiState {
  sidebarOpen: boolean;
  activeModal: string | null;
  modalPayload: unknown;
  toggleSidebar: () => void;
  openModal: (modal: string, payload?: unknown) => void;
  closeModal: () => void;
}

// UI-only state — never persisted, never holds server data.
export const useUiStore = create<UiState>((set) => ({
  sidebarOpen: false,
  activeModal: null,
  modalPayload: undefined,
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  openModal: (modal, payload) => set({ activeModal: modal, modalPayload: payload }),
  closeModal: () => set({ activeModal: null, modalPayload: undefined }),
}));
