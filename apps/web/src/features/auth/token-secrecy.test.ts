import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useAuthStore } from "./store";
import { sessionResponseFixture } from "@/test/msw/fixtures";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
  useSearchParams: () => new URLSearchParams(),
}));

import { useLogin } from "./hooks/use-login";

beforeEach(() => {
  push.mockClear();
  localStorage.clear();
  sessionStorage.clear();
  useAuthStore.setState({ user: null, isAuthenticated: false });
});

// Highest-priority guarantee (FR-007/FR-008, SC-002): the token pair the Go
// backend returns is converted to httpOnly cookies by the BFF and must never
// reach any client-readable surface.
describe("token secrecy", () => {
  it("keeps tokens out of every client-readable surface after login", async () => {
    const { result } = renderHook(() => useLogin());
    act(() => {
      result.current.form.setValue("email", "verified@example.com");
      result.current.form.setValue("password", "secret123");
    });
    await act(async () => {
      await result.current.onSubmit();
    });
    await waitFor(() => expect(push).toHaveBeenCalled());

    const surfaces = [
      JSON.stringify(localStorage),
      JSON.stringify(sessionStorage),
      JSON.stringify(useAuthStore.getState()),
    ].join("|");

    expect(surfaces).not.toContain(sessionResponseFixture.accessToken);
    expect(surfaces).not.toContain(sessionResponseFixture.refreshToken);
    expect(surfaces).not.toMatch(/accessToken|refreshToken/);

    // The non-sensitive profile IS present.
    expect(useAuthStore.getState().user?.email).toBe("verified@example.com");
  });
});
