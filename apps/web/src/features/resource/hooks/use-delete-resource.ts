"use client";

import { useState } from "react";
import { useSWRConfig } from "swr";
import { resourceKey, resourcesKey } from "@/lib/swr/keys";
import { deleteResource as sendDelete } from "../fetchers";

// Delete flows through the BFF, then drops the detail cache and revalidates the
// list so the removed row disappears.
export function useDeleteResource() {
  const { mutate } = useSWRConfig();
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  async function remove(id: string): Promise<void> {
    setIsDeleting(true);
    setError(null);
    try {
      await sendDelete(id);
      await Promise.all([
        mutate(resourceKey(id), undefined, { revalidate: false }),
        mutate(resourcesKey),
      ]);
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setIsDeleting(false);
    }
  }

  return { remove, isDeleting, error };
}
