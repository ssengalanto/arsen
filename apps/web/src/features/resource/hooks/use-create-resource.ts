"use client";

import useSWR from "swr";
import { resourcesKey } from "@/lib/swr/keys";
import { createResource } from "../fetchers";
import type { CreateResourceInput, Resource } from "../types";

function tempId() {
  return `temp-${Math.random().toString(36).slice(2)}`;
}

// Optimistic create: the new row shows immediately, then reconciles against the
// server response. Passing the request as a promise lets SWR supersede the
// in-flight list revalidation (so it can't clobber our write). On error we
// restore the captured baseline — the fallback never lands in the cache, so
// SWR's own rollback would restore `undefined` instead of the prior list.
export function useCreateResource() {
  const { data, mutate } = useSWR<Resource[], Error>(resourcesKey);
  const current = data ?? [];

  async function create(input: CreateResourceInput): Promise<Resource> {
    const now = new Date().toISOString();
    const optimistic: Resource = {
      id: tempId(),
      title: input.title,
      description: input.description ?? "",
      status: input.status,
      createdAt: now,
      updatedAt: now,
    };

    const request = createResource(input);

    try {
      await mutate(
        request.then((saved) => [...current, saved]),
        {
          optimisticData: [...current, optimistic],
          rollbackOnError: true,
          revalidate: false,
          populateCache: true,
        },
      );
      return await request;
    } catch (err) {
      await mutate(current, { revalidate: false, populateCache: true });
      throw err;
    }
  }

  return { create };
}
