"use client";

import useSWR from "swr";
import { resourceKey } from "@/lib/swr/keys";
import type { Resource } from "../types";

// Single-resource detail. Passing `null` when there's no id disables the fetch
// (SWR's conditional-key convention). Hydrates from the RSC-provided fallback,
// then revalidates to server truth on mount.
export function useResource(id: string | null | undefined) {
  const { data, error, isLoading, mutate } = useSWR<Resource, Error>(resourceKey(id));

  return {
    resource: data,
    error,
    isLoading: isLoading && data === undefined,
    mutate,
  };
}
