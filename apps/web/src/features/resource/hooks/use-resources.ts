"use client";

import useSWR from "swr";
import { resourcesKey } from "@/lib/swr/keys";
import type { Resource } from "../types";

// Server cache lives in SWR. The list hydrates instantly from the RSC-provided
// fallback, then revalidates to server truth on mount. `isLoading` treats the
// fallback as loaded data so there's no loading flash on first paint.
export function useResources() {
  const { data, error, isLoading, mutate } = useSWR<Resource[], Error>(resourcesKey);

  return {
    resources: data ?? [],
    error,
    isLoading: isLoading && data === undefined,
    mutate,
  };
}
