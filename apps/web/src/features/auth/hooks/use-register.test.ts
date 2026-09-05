import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { problem } from "@/test/msw/fixtures";
import { useAuthStore } from "../store";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

import { useRegister } from "./use-register";

beforeEach(() => {
  push.mockClear();
  localStorage.clear();
  sessionStorage.clear();
  useAuthStore.setState({ user: null, isAuthenticated: false });
});

async function submit(email: string, password: string) {
  const { result } = renderHook(() => useRegister());
  act(() => {
    result.current.form.setValue("email", email);
    result.current.form.setValue("password", password);
  });
  await act(async () => {
    await result.current.onSubmit();
  });
  return result;
}

describe("useRegister", () => {
  it("redirects to /login?registered=1 and never exposes a token", async () => {
    await submit("new@example.com", "supersecret");
    await waitFor(() => expect(push).toHaveBeenCalledWith("/login?registered=1"));

    const serialized =
      JSON.stringify(localStorage) +
      JSON.stringify(sessionStorage) +
      JSON.stringify(useAuthStore.getState());
    expect(serialized).not.toMatch(/access-token/i);
    expect(serialized).not.toMatch(/refresh-token/i);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("shows a generic message on conflict (anti-enumeration)", async () => {
    server.use(
      http.post("/api/auth/register", () =>
        HttpResponse.json(problem(400, "invalid request"), { status: 400 }),
      ),
    );
    const result = await submit("taken@example.com", "supersecret");
    await waitFor(() =>
      expect(result.current.form.formState.errors.root?.message).toBeTruthy(),
    );
    expect(push).not.toHaveBeenCalled();
  });
});
