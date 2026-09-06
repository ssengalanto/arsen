# Phase 1 Data Model: Next.js Feature-Sliced Frontend Boilerplate

Types are the client-facing shapes; validation rules come from the spec's Functional Requirements and the backend contract (see [contracts/](./contracts/)). Field names use the backend's JSON casing where they cross the wire.

## Entities

### UserProfile (client-cached, non-sensitive)

| Field | Type | Notes |
|---|---|---|
| `id` | `string` | Backend `id` |
| `email` | `string` | |
| `emailVerified` | `boolean` | From `GET /api/users/me` |
| `createdAt` | `string` (ISO 8601) | Display only |

- Cached in `authStore` (Zustand `persist`, localStorage) for instant UI. **Never** contains any token.
- Source of truth is the server session; `useSession` reconciles the cached profile against `/api/me`.

### Session (server-side only)

- Represented to the browser only as httpOnly cookies (`arsen_access`, `arsen_refresh`). Not a JS-readable object.
- See [Cookies](#cookies) and research R-4.

### Resource (demo CRUD stand-in)

| Field | Type | Validation | Notes |
|---|---|---|---|
| `id` | `string` | server-assigned | read-only |
| `title` | `string` | required, 1–120 chars | |
| `description` | `string \| undefined` | optional, ≤2000 chars | |
| `status` | `'draft' \| 'active' \| 'archived'` | required enum | defaults to `draft` on create |
| `createdAt` | `string` (ISO 8601) | server-assigned | read-only |
| `updatedAt` | `string` (ISO 8601) | server-assigned | read-only |

- `CreateResourceInput` = `{ title, description?, status }` (client-provided subset).
- `UpdateResourceInput` = `Partial<CreateResourceInput>`.
- Backed by an in-memory server store (research R-7); labeled as a rename-me stand-in.

### Draft (UI-only)

- `Partial<CreateResourceInput> | null` held in `resourceStore`; persisted (partialize: `filters` + `draft` only). Restored on form mount; cleared after successful submit. Never contains server-assigned fields.

## Zod Schemas (validation source of truth)

- `loginSchema`: `{ email: z.string().email(), password: z.string().min(1) }` → `LoginInput`.
- `registerSchema`: `{ email: z.string().email(), password: z.string().min(8) }` → `RegisterInput`. (Backend enforces its own password policy; mirror the minimum, surface backend field errors for the rest.)
- `resourceSchema`: `{ title: z.string().min(1).max(120), description: z.string().max(2000).optional(), status: z.enum(['draft','active','archived']) }` → `CreateResourceInput`.

Each schema file exports both the schema and its `z.infer<>` type (named exports).

## Stores (Zustand)

| Store | Location | Persist | Partialize | State | Actions |
|---|---|---|---|---|---|
| `uiStore` | `lib/stores/ui-store.ts` | ✗ | — | `sidebarOpen`, `activeModal: string \| null`, `modalPayload: unknown` | `toggleSidebar`, `openModal(modal, payload?)`, `closeModal` |
| `authStore` | `features/auth/store.ts` | ✓ | profile only | `user: UserProfile \| null`, `isAuthenticated: boolean` | `setUser`, `clearAuth` |
| `resourceStore` | `features/resource/store.ts` | ✓ | `filters` + `draft` | `selectedIds: string[]`, `filters`, `draft`, `wizardStep: number` | selection, `setFilter`/`resetFilters`, `setDraft`/`clearDraft`, `nextStep`/`prevStep`/`resetWizard` |

- Consumers use granular selectors `(s) => s.field` only.
- `selectedIds` and `wizardStep` are **never** persisted.

## SWR Keys (`lib/swr/keys.ts`)

| Key | Meaning |
|---|---|
| `'/api/me'` | session profile |
| `'/api/resources'` | list |
| `['/api/resources', id]` | detail (null when `!id`) |

Strings for static keys, arrays for parameterized. No object keys.

## Cookies

| Name | Contains | httpOnly | Secure | SameSite | Max-Age | Set/cleared by |
|---|---|---|---|---|---|---|
| `arsen_access` | JWT access token | yes | prod only | Lax | `expiresIn` (login/refresh response) | `POST /api/auth/login`, `/refresh`; cleared on `/logout` |
| `arsen_refresh` | opaque refresh token | yes | prod only | Lax | env `SESSION_REFRESH_MAX_AGE` (default 30d) | same as above |

Presence of `arsen_access`/`arsen_refresh` is what `middleware.ts` checks for the `(app)` group.

## State Transitions

**Auth session**: `anonymous → (login success) → authenticated → (access expired) → refreshing → authenticated | anonymous(redirect)`; `authenticated → (logout) → anonymous`. Register produces `unverified` (cannot log in until `emailVerified`), per research R-2.

**Resource**: `draft ↔ active → archived`. Create defaults to `draft`. Optimistic create inserts a temporary row (temp id) reconciled or rolled back on the server response.
