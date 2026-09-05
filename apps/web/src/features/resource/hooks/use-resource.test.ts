import { describe, it, expect, beforeEach, vi } from "vitest";
import type { ReactNode } from "react";
import { createElement } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { SWRConfig, unstable_serialize } from "swr";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { swrConfig } from "@/lib/swr/config";
import { resourceKey } from "@/lib/swr/keys";
import { useResource } from "./use-resource";
import type { Resource } from "../types";

const row: Resource = {
  id: "res-1",
  title: "Server-rendered detail",
  description: "",
  status: "active",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

function wrapperWithFallback(fallback: Record<string, unknown>) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(
      SWRConfig,
      { value: { ...swrConfig, provider: () => new Map(), dedupingInterval: 0, fallback } },
      children,
    );
  };
}

beforeEach(() => {
  server.resetHandlers();
});

describe("useResource", () => {
  it("returns undefined and never fetches when id is null", () => {
    const fetchSpy = vi.fn();
    server.use(
      http.get("/api/resources/:id", () => {
        fetchSpy();
        return HttpResponse.json(row);
      }),
    );

    const { result } = renderHook(() => useResource(null), {
      wrapper: wrapperWithFallback({}),
    });

    expect(result.current.resource).toBeUndefined();
    expect(result.current.isLoading).toBe(false);
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("hydrates instantly from the SWR fallback", () => {
    const { result } = renderHook(() => useResource("res-1"), {
      wrapper: wrapperWithFallback({ [unstable_serialize(resourceKey("res-1"))]: row }),
    });

    expect(result.current.isLoading).toBe(false);
    expect(result.current.resource).toEqual(row);
  });

  it("revalidates to server truth after first paint", async () => {
    server.use(
      http.get("/api/resources/res-1", () =>
        HttpResponse.json({ ...row, title: "Revalidated detail" }),
      ),
    );

    const { result } = renderHook(() => useResource("res-1"), {
      wrapper: wrapperWithFallback({ [unstable_serialize(resourceKey("res-1"))]: row }),
    });

    await waitFor(() => expect(result.current.resource?.title).toBe("Revalidated detail"));
  });
});
