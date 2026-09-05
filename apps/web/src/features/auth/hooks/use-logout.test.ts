import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { useAuthStore } from "../store";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

import { useLogout } from "./use-logout";

beforeEach(() => {
  push.mockClear();
  server.resetHandlers();
  useAuthStore.setState({
    user: {
      id: "1",
      email: "verified@example.com",
      emailVerified: true,
      createdAt: "2026-01-01T00:00:00Z",
    },
    isAuthenticated: true,
  });
});

describe("useLogout", () => {
  it("revokes upstream, clears auth, and redirects to /login", async () => {
    const { result } = renderHook(() => useLogout());

    await act(async () => {
      await result.current.logout();
    });

    await waitFor(() => expect(push).toHaveBeenCalledWith("/login"));
    expect(useAuthStore.getState().user).toBeNull();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(result.current.pending).toBe(false);
  });

  it("still clears client state and redirects when the server revoke fails", async () => {
    server.use(
      http.delete("/api/auth/logout", () => new HttpResponse(null, { status: 500 })),
    );

    const { result } = renderHook(() => useLogout());

    await act(async () => {
      await result.current.logout();
    });

    await waitFor(() => expect(push).toHaveBeenCalledWith("/login"));
    expect(useAuthStore.getState().user).toBeNull();
  });
});
