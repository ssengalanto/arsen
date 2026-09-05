import { describe, it, expect, beforeEach } from "vitest";
import type { ReactNode } from "react";
import { createElement } from "react";
import { renderHook, act, waitFor } from "@testing-library/react";
import { SWRConfig, unstable_serialize } from "swr";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { problem } from "@/test/msw/fixtures";
import { swrConfig } from "@/lib/swr/config";
import { resourceKey } from "@/lib/swr/keys";
import { useResource } from "./use-resource";
import { useUpdateResource } from "./use-update-resource";
import type { Resource } from "../types";

const row: Resource = {
  id: "res-1",
  title: "Before",
  description: "",
  status: "draft",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    SWRConfig,
    {
      value: {
        ...swrConfig,
        provider: () => new Map(),
        dedupingInterval: 0,
        revalidateOnMount: false,
        fallback: { [unstable_serialize(resourceKey("res-1"))]: row },
      },
    },
    children,
  );
}

function useDetailAndUpdate() {
  return { detail: useResource("res-1"), update: useUpdateResource() };
}

beforeEach(() => {
  server.resetHandlers();
});

describe("useUpdateResource", () => {
  it("writes server truth into the detail cache without an extra refetch", async () => {
    server.use(
      http.patch("/api/resources/res-1", async ({ request }) => {
        const body = (await request.json()) as { status: string };
        return HttpResponse.json({ ...row, status: body.status, title: "After" });
      }),
    );

    const { result } = renderHook(useDetailAndUpdate, { wrapper });

    await act(async () => {
      await result.current.update.update("res-1", { status: "archived" });
    });

    await waitFor(() => expect(result.current.detail.resource?.title).toBe("After"));
    expect(result.current.detail.resource?.status).toBe("archived");
    expect(result.current.update.isUpdating).toBe(false);
  });

  it("surfaces the error and leaves isUpdating false", async () => {
    server.use(
      http.patch("/api/resources/res-1", () =>
        HttpResponse.json(problem(400, "invalid"), { status: 400 }),
      ),
    );

    const { result } = renderHook(useDetailAndUpdate, { wrapper });

    await act(async () => {
      await result.current.update.update("res-1", { status: "archived" }).catch(() => undefined);
    });

    expect(result.current.update.error).toBeTruthy();
    expect(result.current.update.isUpdating).toBe(false);
  });
});
