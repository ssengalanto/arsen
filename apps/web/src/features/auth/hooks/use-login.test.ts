import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { userProfileFixture, problem } from "@/test/msw/fixtures";
import { useAuthStore } from "../store";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
  useSearchParams: () => new URLSearchParams(),
}));

import { useLogin } from "./use-login";

beforeEach(() => {
  push.mockClear();
  localStorage.clear();
  useAuthStore.setState({ user: null, isAuthenticated: false });
});

async function submit(email: string, password: string) {
  const { result } = renderHook(() => useLogin());
  act(() => {
    result.current.form.setValue("email", email);
    result.current.form.setValue("password", password);
  });
  await act(async () => {
    await result.current.onSubmit();
  });
  return result;
}

describe("useLogin", () => {
  it("sets the profile and redirects on success", async () => {
    const result = await submit("verified@example.com", "secret123");
    await waitFor(() => expect(push).toHaveBeenCalledWith("/dashboard"));
    expect(useAuthStore.getState().user?.email).toBe(userProfileFixture.email);
    expect(result.current.form.formState.errors.root).toBeUndefined();
  });

  it("surfaces a verify-your-email message on 403", async () => {
    server.use(
      http.post("/api/auth/login", () =>
        HttpResponse.json(problem(403, "email not verified"), { status: 403 }),
      ),
    );
    const result = await submit("unverified@example.com", "secret123");
    await waitFor(() =>
      expect(result.current.form.formState.errors.root?.message).toMatch(/verify/i),
    );
    expect(push).not.toHaveBeenCalled();
  });

  it("surfaces a generic error on 401", async () => {
    server.use(
      http.post("/api/auth/login", () =>
        HttpResponse.json(problem(401, "invalid credentials"), { status: 401 }),
      ),
    );
    const result = await submit("verified@example.com", "wrong");
    await waitFor(() =>
      expect(result.current.form.formState.errors.root?.message).toBeTruthy(),
    );
    expect(push).not.toHaveBeenCalled();
    expect(useAuthStore.getState().user).toBeNull();
  });
});
