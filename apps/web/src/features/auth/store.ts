import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { UserProfile } from "./types";

interface AuthState {
  user: UserProfile | null;
  isAuthenticated: boolean;
  setUser: (user: UserProfile) => void;
  clearAuth: () => void;
}

// Persisted to localStorage — profile only, never a token. `isAuthenticated`
// is recomputed from the rehydrated profile so the two never drift.
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      setUser: (user) => set({ user, isAuthenticated: true }),
      clearAuth: () => set({ user: null, isAuthenticated: false }),
    }),
    {
      name: "arsen-auth",
      partialize: (state) => ({ user: state.user }),
      onRehydrateStorage: () => (state) => {
        if (state?.user) state.isAuthenticated = true;
      },
    },
  ),
);
