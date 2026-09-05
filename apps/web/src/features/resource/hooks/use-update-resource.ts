"use client";

import { useState } from "react";
import { useSWRConfig } from "swr";
import { resourceKey, resourcesKey } from "@/lib/swr/keys";
import { updateResource as sendUpdate } from "../fetchers";
import type { Resource, UpdateResourceInput } from "../types";

// Update flows through the BFF, then reconciles both caches: the detail key is
// populated with server truth (no extra roundtrip) and the list is revalidated.
export function useUpdateResource() {
  const { mutate } = useSWRConfig();
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  async function update(id: string, input: UpdateResourceInput): Promise<Resource> {
    setIsUpdating(true);
    setError(null);
    try {
      const saved = await sendUpdate(id, input);
      await Promise.all([
        mutate(resourceKey(id), saved, { revalidate: false }),
        mutate(resourcesKey),
      ]);
      return saved;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setIsUpdating(false);
    }
  }

  return { update, isUpdating, error };
}
