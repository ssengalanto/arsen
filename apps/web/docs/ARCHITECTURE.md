# Architecture & "add a feature slice" guide

This frontend is organized as **feature slices**. A slice owns everything it
needs — schemas, types, store, fetchers, hooks, and components — and exposes a
single public entry (`index.ts`). Cross-feature reach-ins are blocked by lint.
The goal: you can add a complete, working feature in **one folder** in under 30
minutes without editing any other slice.

## The two layers

```
src/
  features/<name>/     # self-contained slices — the app's real surface area
  app/                 # Next.js routes: pages (RSC) + BFF route handlers
  lib/                 # shared, feature-agnostic plumbing (api, swr, auth, stores)
  components/ui/       # shadcn/base-ui primitives (generated, shared)
  components/shared/   # cross-feature presentational components (if any)
```

- **A feature never imports another feature's internals.** Import only
  `@/features/<name>` (the public entry). This is enforced by
  `no-restricted-imports` — see [the boundary rule](#import-boundary).
- **`lib/` is shared and feature-agnostic.** It knows nothing about any specific
  feature. Features depend on `lib/`, never the reverse.

## Anatomy of a slice

```
features/auth/
  index.ts             # PUBLIC ENTRY — the only thing other code may import
  types.ts             # domain types (e.g. UserProfile) — never any token fields
  schemas/             # zod schemas + inferred input types (one source of truth)
  store.ts             # Zustand UI/client state (persisted profile only)
  fetchers/            # thin functions calling BFF relative URLs via lib/api/client
  hooks/               # the read/write chains that components consume
  components/          # the slice's UI
```

Everything inside the slice imports its siblings with **relative paths**
(`../store`, `./use-login`). Only code *outside* the slice goes through
`@/features/auth`.

## The read chain (server data → screen)

Server data always flows **fetcher → hook (SWR) → component**. Components never
call fetch directly and never store server data in Zustand.

```
component  ─uses─▶  hook (useSWR / useSWRMutation)  ─calls─▶  fetcher  ─▶  /api/... (BFF)  ─▶  Go backend
```

- The **fetcher** hits a *relative* BFF URL (`/api/me`), never the Go backend
  directly. `lib/api/client.ts` sends `credentials: "include"` so cookies ride
  along.
- The **hook** owns caching + revalidation via SWR. Example: `useSession`
  (`features/auth/hooks/use-session.ts`) reads the `/api/me` SWR key and
  reconciles it against the persisted profile.
- The **component** renders `{ data, isLoading, error }` states. It subscribes to
  server state through the hook and to UI state through the store — never mixes
  the two.

For first paint on server-rendered routes, an RSC fetches the initial data and
hands it to a client `<SWRConfig fallback={...}>` boundary so the client hook
hydrates without a second network round-trip (see R-6).

## The write chain (form → mutation → cache)

Writes flow **schema → resolver → form → mutation hook**, then reconcile the SWR
cache.

```
zod schema ─▶ zodResolver ─▶ react-hook-form ─▶ mutation hook (fetcher + bound mutate) ─▶ cache update
```

1. Define a **zod schema** in `schemas/`; export the schema and its inferred type
   (`export type LoginInput = z.infer<typeof loginSchema>`). This is the single
   source of truth for both validation and the TS type.
2. Wire it into `react-hook-form` via `zodResolver(schema)`.
3. On submit, call the **fetcher**; on success update cache with the **bound**
   `mutate` from `useSWRConfig()` or the mutation hook return — never the global
   `swr` `mutate` (also lint-blocked).
4. Map RFC 9457 problem+json field errors onto `form.setError(field, …)`; fall
   back to a generic `root` error. See `features/auth/hooks/use-login.ts` for the
   canonical pattern.

## Server state vs UI state — keep them separate

| Concern | Home | Example |
|---|---|---|
| **Server state** (anything that lives on the backend) | SWR | profile, resource list/detail |
| **UI/client state** (ephemeral or view-only) | Zustand | modal open, wizard step, selected rows, form drafts |

Rules:

- **Never** put server data into a Zustand store as the source of truth. The one
  deliberate exception is the auth store persisting the *non-sensitive profile*
  for instant hydration — and even then `/api/me` (SWR) is authoritative and
  overwrites it (`useSession`).
- **Persist only what should survive a reload.** Zustand `persist` +
  `partialize` — e.g. persist a form `draft` and `filters`, but not `selectedIds`
  or `wizardStep`.
- Tokens live **only** in httpOnly cookies set by the BFF. They must never appear
  in `localStorage`, `sessionStorage`, or any store — this is asserted by
  `features/auth/token-secrecy.test.ts`.

## Import boundary

`eslint.config.mjs` exports `restrictedImportsOptions`, used by both the project
lint config and `features/__tests__/import-boundary.test.ts`. It blocks:

- `@/features/*/*` and deeper — reaching into another slice's internals. Import
  `@/features/<name>` instead.
- the global `mutate` from `swr` — use the bound `mutate`.

**The one documented exception:** the app shell (`app/(app)/layout.tsx`) is
allowed to consume auth's session via the public entry `@/features/auth`
(`useSession`, `UserMenu`). This goes through the public entry, so it satisfies
the rule without a special case.

## Adding a new slice — checklist

1. `mkdir src/features/<name>` and add: `types.ts`, `schemas/`, `store.ts`,
   `fetchers/`, `hooks/`, `components/`.
2. Write the zod schema(s) first; export schema + inferred type.
3. Add fetchers hitting relative `/api/<name>` BFF URLs via `lib/api/client.ts`.
4. Add BFF route handlers under `app/api/<name>/` that proxy to your backend
   (convert any tokens to httpOnly cookies in the handler — never return them to
   the browser).
5. Build read hooks (`useSWR`) and write hooks (`useSWRMutation` + bound
   `mutate`).
6. Build components consuming the hooks; keep server state in SWR and UI state in
   the store.
7. Export the public surface from `src/features/<name>/index.ts`.
8. Add route pages under `app/`; RSC-fetch initial data into a `<SWRConfig
   fallback>` boundary for first paint.
9. Run `pnpm check` (typecheck + lint + test). The boundary lint proves you did
   not reach into another slice.

See `features/auth/` as the reference implementation and `features/resource/` for
the full CRUD + optimistic-update pattern.
