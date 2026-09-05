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
import { useCreateResource } from "./use-create-resource";
import type { Resource } from "../types";

const existing: Resource[] = [
  {
    id: "existing-1",
    title: "Already here",
    description: "",
    status: "active",
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
        // Isolate optimistic-create behavior from list revalidation, which is
        // covered separately in use-resources.test.ts.
        revalidateOnMount: false,
        fallback: { [resourcesKey]: existing },
      },
    },
    children,
  );
}

function useListAndCreate() {
  return { list: useResources(), create: useCreateResource() };
}

beforeEach(() => {
  server.resetHandlers();
});

describe("useCreateResource", () => {
  it("optimistically inserts the new row, then reconciles with the server", async () => {
    server.use(
      http.post("/api/resources", async ({ request }) => {
        const body = (await request.json()) as { title: string };
        return HttpResponse.json(
          {
            id: "server-1",
            title: body.title,
            description: "",
            status: "draft",
            createdAt: "2026-02-01T00:00:00Z",
            updatedAt: "2026-02-01T00:00:00Z",
          },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(useListAndCreate, { wrapper });

    let pending!: Promise<unknown>;
    act(() => {
      pending = result.current.create.create({ title: "Optimistic row", status: "draft" });
    });

    // The optimistic row shows immediately, before the server responds.
    expect(result.current.list.resources.some((r) => r.title === "Optimistic row")).toBe(true);

    await act(async () => {
      await pending;
    });

    // Reconciled against server truth: the real id replaces the temp one.
    await waitFor(() =>
      expect(result.current.list.resources.some((r) => r.id === "server-1")).toBe(true),
    );
    expect(result.current.list.resources.filter((r) => r.title === "Optimistic row")).toHaveLength(
      1,
    );
  });

  it("rolls back the optimistic row when the server rejects", async () => {
    server.use(
      http.post("/api/resources", () =>
        HttpResponse.json(problem(400, "invalid request"), { status: 400 }),
      ),
    );

    const { result } = renderHook(useListAndCreate, { wrapper });

    await act(async () => {
      await result.current.create
        .create({ title: "Doomed row", status: "draft" })
        .catch(() => undefined);
    });

    // The list is back to exactly what it was before the attempt.
    expect(result.current.list.resources).toHaveLength(existing.length);
    expect(result.current.list.resources.some((r) => r.title === "Doomed row")).toBe(false);
  });
});
