import { describe, it, expect, beforeEach, vi } from "vitest";
import type { ReactNode } from "react";
import { createElement } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { SWRConfig } from "swr";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { swrConfig } from "@/lib/swr/config";
import { resourcesKey } from "@/lib/swr/keys";
import { useResources } from "./use-resources";
import type { Resource } from "../types";

const seed: Resource[] = [
  {
    id: "seed-1",
    title: "Server-rendered row",
    description: "",
    status: "active",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
];

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

describe("useResources", () => {
  it("hydrates instantly from the SWR fallback (no loading flash, no double fetch)", () => {
    const fetchSpy = vi.fn();
    server.use(
      http.get("/api/resources", () => {
        fetchSpy();
        return HttpResponse.json(seed);
      }),
    );

    const { result } = renderHook(() => useResources(), {
      wrapper: wrapperWithFallback({ [resourcesKey]: seed }),
    });

    // First paint comes straight from the RSC-provided fallback.
    expect(result.current.isLoading).toBe(false);
    expect(result.current.resources).toEqual(seed);
  });

  it("revalidates to server truth after first paint", async () => {
    const fresh: Resource[] = [
      { ...seed[0]!, title: "Revalidated row" },
      {
        id: "seed-2",
        title: "Added on the server",
        description: "",
        status: "draft",
        createdAt: "2026-01-02T00:00:00Z",
        updatedAt: "2026-01-02T00:00:00Z",
      },
    ];
    server.use(http.get("/api/resources", () => HttpResponse.json(fresh)));

    const { result } = renderHook(() => useResources(), {
      wrapper: wrapperWithFallback({ [resourcesKey]: seed }),
    });

    await waitFor(() => expect(result.current.resources).toHaveLength(2));
    expect(result.current.resources[0]?.title).toBe("Revalidated row");
  });
});
