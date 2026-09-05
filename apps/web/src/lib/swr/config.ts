import type { SWRConfiguration } from "swr";
import { apiFetch } from "@/lib/api/client";

// Resolves a SWR key (string or `[base, param]` tuple) to a BFF URL and fetches
// it through the shared client (which throws `ApiError` on non-2xx).
function fetcher(key: string | readonly [string, string]): Promise<unknown> {
  const url =
    typeof key === "string" ? key : `${key[0]}/${encodeURIComponent(key[1])}`;
  return apiFetch(url);
}

export const swrConfig: SWRConfiguration = {
  fetcher,
  revalidateOnFocus: false,
  shouldRetryOnError: false,
};
