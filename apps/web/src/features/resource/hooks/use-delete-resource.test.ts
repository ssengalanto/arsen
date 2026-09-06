import { describe, it, expect, beforeEach } from "vitest";
import type { ReactNode } from "react";
import { createElement } from "react";
import { renderHook, act, waitFor } from "@testing-library/react";
import { SWRConfig } from "swr";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { problem } from "@/test/msw/fixtures";
import { swrConfig } from "@/lib/swr/config";
import { resourcesKey } from "@/lib/swr/keys";
import { useResources } from "./use-resources";
import { useDeleteResource } from "./use-delete-resource";
import type { Resource } from "../types";

const existing: Resource[] = [
  {
    id: "res-1",
    title: "Doomed",
    description: "",
    status: "draft",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
];

function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    SWRConfig,
    {
      value: {
        ...swrConfig,
        provider: () => new Map(),
        dedupingInterval: 0,
        fallback: { [resourcesKey]: existing },
      },
    },
    children,
  );
}

function useListAndDelete() {
  return { list: useResources(), remove: useDeleteResource() };
}

beforeEach(() => {
  server.resetHandlers();
});

describe("useDeleteResource", () => {
  it("removes the row and revalidates the list to the server truth", async () => {
    server.use(
      http.delete("/api/resources/res-1", () => new HttpResponse(null, { status: 204 })),
      http.get("/api/resources", () => HttpResponse.json([])),
    );

    const { result } = renderHook(useListAndDelete, { wrapper });

    await act(async () => {
      await result.current.remove.remove("res-1");
    });

    await waitFor(() => expect(result.current.list.resources).toHaveLength(0));
    expect(result.current.remove.isDeleting).toBe(false);
  });

  it("surfaces the error and leaves isDeleting false", async () => {
    server.use(
      http.delete("/api/resources/res-1", () =>
        HttpResponse.json(problem(404, "not found"), { status: 404 }),
      ),
    );

    const { result } = renderHook(useListAndDelete, { wrapper });

    await act(async () => {
      await result.current.remove.remove("res-1").catch(() => undefined);
    });

    expect(result.current.remove.error).toBeTruthy();
    expect(result.current.remove.isDeleting).toBe(false);
  });
});
