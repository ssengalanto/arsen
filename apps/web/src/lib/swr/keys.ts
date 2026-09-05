// SWR keys. Strings for static resources, tuples for parameterized ones.
// Never use object keys (unstable serialization).
export const meKey = "/api/me";
export const resourcesKey = "/api/resources";

export function resourceKey(id: string | null | undefined) {
  return id ? ([resourcesKey, id] as const) : null;
}

export type ResourceKey = ReturnType<typeof resourceKey>;
