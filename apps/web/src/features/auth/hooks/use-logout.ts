"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { logout } from "../fetchers";
import { useAuthStore } from "../store";

export function useLogout() {
  const router = useRouter();
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const [pending, setPending] = useState(false);

  const run = useCallback(async () => {
    setPending(true);
    try {
      await logout();
    } catch {
      // Client cleanup must complete even if the server revoke fails.
    } finally {
      clearAuth();
      setPending(false);
      router.push("/login");
    }
  }, [clearAuth, router]);

  return { logout: run, pending };
}
