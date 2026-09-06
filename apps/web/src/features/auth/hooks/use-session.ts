"use client";

import { useEffect } from "react";
import useSWR from "swr";
import { meKey } from "@/lib/swr/keys";
import type { UserProfile } from "../types";
import { useAuthStore } from "../store";

// The single documented session accessor. Renders instantly from the persisted
// profile, then reconciles against the server truth at /api/me — a stale cached
// profile always yields to what the server returns.
export function useSession() {
  const cached = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const clearAuth = useAuthStore((s) => s.clearAuth);

  const { data, error, isLoading } = useSWR<UserProfile, Error>(meKey);

  useEffect(() => {
    if (data) setUser(data);
  }, [data, setUser]);

  useEffect(() => {
    if (error) clearAuth();
  }, [error, clearAuth]);

  return {
    user: data ?? cached,
    isLoading,
    error,
    isAuthenticated: Boolean(data ?? cached),
  };
}
