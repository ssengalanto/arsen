import { describe, it, expect, beforeEach } from "vitest";
import type { ReactNode } from "react";
import { createElement } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { SWRConfig } from "swr";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { userProfileFixture } from "@/test/msw/fixtures";
import { swrConfig } from "@/lib/swr/config";
import { useAuthStore } from "../store";
import { useSession } from "./use-session";

function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    SWRConfig,
    { value: { ...swrConfig, provider: () => new Map(), dedupingInterval: 0 } },
    children,
  );
}

beforeEach(() => {
  localStorage.clear();
  useAuthStore.setState({ user: null, isAuthenticated: false });
});

describe("useSession", () => {
  it("renders the cached profile first, then reconciles against server truth", async () => {
    const stale = { ...userProfileFixture, email: "stale@example.com" };
    useAuthStore.setState({ user: stale, isAuthenticated: true });

    const { result } = renderHook(() => useSession(), { wrapper });

    // Instant hydration from the persisted profile.
    expect(result.current.user?.email).toBe("stale@example.com");

    // Server truth (verified@example.com) wins once /api/me resolves.
    await waitFor(() =>
      expect(result.current.user?.email).toBe(userProfileFixture.email),
    );
    expect(useAuthStore.getState().user?.email).toBe(userProfileFixture.email);
  });

  it("clears auth when /api/me returns 401", async () => {
    useAuthStore.setState({ user: userProfileFixture, isAuthenticated: true });
    server.use(
      http.get("/api/me", () =>
        HttpResponse.json({ title: "Unauthorized", status: 401 }, { status: 401 }),
      ),
    );

    renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(useAuthStore.getState().user).toBeNull());
  });
});
