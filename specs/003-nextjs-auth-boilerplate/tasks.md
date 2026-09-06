---
description: "Task list for Next.js Feature-Sliced Frontend Boilerplate"
---

# Tasks: Next.js Feature-Sliced Frontend Boilerplate

**Input**: Design documents from `/specs/003-nextjs-auth-boilerplate/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/)

**Tests**: INCLUDED — the spec explicitly requires them (FR-026 network-mocked unit/hook tests incl. token-secrecy assertion; FR-027 Playwright E2E happy path) and project memory mandates TDD (write failing tests first).

**Organization**: Tasks are grouped by user story (US1 auth = P1/MVP, US2 architecture = P2, US3 resource = P3) so each is independently implementable and testable.

**Base path**: All source paths are under `apps/web/` unless noted. `@/*` maps to `apps/web/src/*`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: US1 / US2 / US3 (setup, foundational, and polish tasks carry no story label)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install dependencies and configure the toolchain in the existing `apps/web` workspace (Next 16.3.4 / React 19.2.8 / Tailwind v4 / Vitest already present).

- [X] T001 Read `apps/web/node_modules/next/dist/docs/` (App Router, route handlers, cookies, middleware) before writing any Next 16 code, per `apps/web/AGENTS.md` — Next 16 has breaking changes vs training data
- [X] T002 Add runtime deps to `apps/web/package.json`: `swr`, `zustand`, `react-hook-form`, `zod`, `@hookform/resolvers`, `class-variance-authority`, `tailwind-merge`, `clsx`, `lucide-react` (via `pnpm add`); verify `@hookform/resolvers` supports the installed `zod` major and pin the compatible pair if needed (record in DECISIONS.md later)
- [X] T003 Add dev deps to `apps/web/package.json`: `msw`, `@playwright/test`, `prettier-plugin-tailwindcss`, `@testing-library/jest-dom`, `typescript-eslint` (via `pnpm add -D`); run `pnpm exec playwright install` for browsers
- [X] T004 Extend `apps/web/tsconfig.json` compiler options with `noUncheckedIndexedAccess`, `noImplicitOverride`, `verbatimModuleSyntax` (keep existing `strict`, `moduleResolution: bundler`, `@/*` alias)
- [X] T005 [P] Update `apps/web/eslint.config.mjs`: add `typescript-eslint` recommended-type-checked, ban `any`, `no-restricted-imports` to block cross-feature internal imports (`@/features/*/!(index)`) and global `mutate` misuse, require named exports outside Next reserved files, put `eslint-config-prettier` last
- [X] T006 [P] Add `prettier-plugin-tailwindcss` to `apps/web/prettier.config.mjs` (create if absent)
- [X] T007 [P] Create `apps/web/vitest.config.ts` with happy-dom env, `@/*` alias, `setupFiles: ['./src/test/setup.ts']`, and coverage config targeting `src/features/**` (≥80% threshold on hooks/fetchers/schemas)
- [X] T008 [P] Create `apps/web/playwright.config.ts` (baseURL `http://localhost:3000`, webServer runs `pnpm dev`, chromium project, `testDir: ./e2e`)
- [X] T009 Add scripts to `apps/web/package.json`: `e2e` (`playwright test`) and `check` (`pnpm typecheck && pnpm lint && pnpm test`); confirm `typecheck`, `lint`, `test`, `format` already exist
- [X] T010 Initialize shadcn/ui: run `pnpm dlx shadcn@latest init`, accept Tailwind v4 CSS-first output, generate `apps/web/components.json`; verify it emits `@theme` CSS (not a JS config) and follow whatever it generates
- [X] T011 Ensure `apps/web/.env.example` documents `NEXT_PUBLIC_API_URL=http://localhost:8080` and `SESSION_REFRESH_MAX_AGE` (optional, default 30d); copy to `apps/web/.env`

**Checkpoint**: `pnpm install` clean; `pnpm typecheck` and `pnpm lint` run (may report on not-yet-written files); toolchain ready.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared `lib/*` infrastructure, cookie/BFF plumbing, providers, error boundary, middleware, and the test harness that EVERY user story depends on.

**⚠️ CRITICAL**: No user-story work can begin until this phase is complete.

- [X] T012 [P] Create `src/lib/utils/cn.ts` — `cn()` helper (`clsx` + `tailwind-merge`), named export
- [X] T013 [P] Create `src/lib/api/error.ts` — `ApiError` class wrapping the RFC 9457 `application/problem+json` body (`{type,title,status,detail,instance,errors?:[{field,detail}]}`) per [contracts/backend-endpoints.md](./contracts/backend-endpoints.md) and research R-3
- [X] T014 [P] Write failing test `src/lib/api/error.test.ts` — asserts `ApiError` parses problem+json, exposes `errors[]`, and never retains tokens/PII (TDD: write before T013 passes)
- [X] T015 [P] Create `src/lib/auth/cookies.ts` (server-only) — cookie name constants (`arsen_access`, `arsen_refresh`) and set/clear helpers with `httpOnly; SameSite=Lax; Path=/; Secure(prod)`, access Max-Age = `expiresIn`, refresh Max-Age = `SESSION_REFRESH_MAX_AGE` (default 30d), per data-model Cookies table + research R-4
- [X] T016 Create `src/lib/api/server.ts` (server-only) — server fetch helper that attaches `Authorization: Bearer <access cookie>`, on 401 calls the refresh flow once, rewrites both cookies, retries once, then surfaces 401; throws `ApiError` on non-2xx (research R-5)
- [X] T017 Write failing test `src/lib/api/server.test.ts` — simulates 401-then-refresh-then-success and terminal double-401 (clears cookies, surfaces 401), network mocked via MSW (FR-006)
- [X] T018 [P] Create `src/lib/api/client.ts` — browser fetchers hitting relative BFF URLs with `credentials: 'include'`, throwing `ApiError` on non-2xx (never talks to the Go backend directly, FR-023)
- [X] T019 [P] Create `src/lib/swr/keys.ts` — `'/api/me'`, `'/api/resources'`, `['/api/resources', id]` (strings for static, arrays for parameterized; no object keys), per data-model SWR Keys
- [X] T020 [P] Create `src/lib/swr/config.ts` — default `SWRConfig` options (fetcher from client.ts, revalidation policy, onError → ApiError)
- [X] T021 [P] Create `src/lib/stores/ui-store.ts` — `uiStore` (no persist): `sidebarOpen`, `activeModal`, `modalPayload`; actions `toggleSidebar`, `openModal`, `closeModal`; granular selectors only (data-model Stores table)
- [X] T022 Create `src/app/providers.tsx` — client `<SWRConfig>` boundary + toaster
- [X] T023 Create `src/app/layout.tsx` (root: `<html>`, fonts, providers), `src/app/globals.css` (Tailwind v4 `@theme` from shadcn init), `src/app/error.tsx` (global error boundary reading `ApiError`), `src/app/not-found.tsx`
- [X] T024 Create `apps/web/middleware.ts` — presence-check `arsen_access`/`arsen_refresh` on the `(app)` group; redirect to `/login?next=…` when absent (no signature verification, FR-024/research R-4)
- [X] T025 Write failing test for middleware route protection (unauthenticated → redirect to `/login?next=…`) in `src/test/middleware.test.ts` (FR-005, SC-003)
- [X] T026 [P] Create `src/test/setup.ts` — RTL + `@testing-library/jest-dom` + MSW server start/reset/close
- [X] T027 [P] Create `src/test/msw/handlers.ts` and `src/test/msw/fixtures.ts` — MSW handlers for the BFF + upstream Go endpoints with realistic problem+json fixtures (FR-026: mock network, never hand-stub)
- [X] T028 [P] Generate baseline shadcn primitives into `src/components/ui/` (button, input, label, form, card, dialog, sonner/toast) via `pnpm dlx shadcn@latest add …`

**Checkpoint**: Shared infra + BFF plumbing + test harness ready. User stories can now begin.

---

## Phase 3: User Story 1 - Secure end-to-end authentication (Priority: P1) 🎯 MVP

**Goal**: Register, login, session persistence across reload, presence-gated protected routes, silent refresh, and logout — with the access/refresh tokens never readable by client JS (httpOnly cookies only).

**Independent Test**: With the Go backend reachable and a pre-verified account (research R-2), log in → land in `(app)` → reload (still signed in) → log out → hit a protected route → redirected to `/login`. Automated test asserts the token is absent from `localStorage`, `sessionStorage`, `authStore` state, and any serialized client output.

> ⚠️ **NOTE (spec deviation, research R-1/R-2)**: register does NOT log the user in — the backend returns no tokens and blocks login (403) until email is verified. Register success redirects to `/login?registered=1` with a verify notice; the authenticated flow uses a pre-verified account.

### Tests for User Story 1 (write first, ensure they FAIL) ⚠️

- [X] T029 [P] [US1] Write failing test `src/features/auth/schemas/schemas.test.ts` — `loginSchema` (email + non-empty password) and `registerSchema` (email + password min 8) accept/reject correctly (data-model Zod schemas)
- [X] T030 [P] [US1] Write failing test `src/features/auth/hooks/use-login.test.ts` (MSW) — successful login sets profile and redirects; 403 surfaces "verify your email"; 401 surfaces generic error
- [X] T031 [P] [US1] Write failing test `src/features/auth/hooks/use-register.test.ts` (MSW) — success redirects to `/login?registered=1` and asserts NO token appears anywhere (localStorage/sessionStorage/store); conflict → generic message (anti-enumeration)
- [X] T032 [P] [US1] Write failing **token-secrecy** test `src/features/auth/token-secrecy.test.ts` — after a full login, assert no token in `localStorage`, `sessionStorage`, `authStore` state, or serialized client output; only the profile is present (FR-007/FR-008, SC-002 — highest priority)
- [X] T033 [P] [US1] Write failing test `src/features/auth/hooks/use-session.test.ts` (MSW) — hydrates from persisted profile and reconciles against `/api/me`; stale cached profile yields to server truth (FR-009, edge case)

### BFF route handlers for User Story 1

- [X] T034 [P] [US1] Create `src/app/api/auth/login/route.ts` — POST → Go `POST /api/sessions`; on 200 set both cookies (access Max-Age=`expiresIn`), respond `200 { user: UserProfile }` with NO tokens in body; pass through 401/403/400/429 problem+json ([contracts/bff-endpoints.md](./contracts/bff-endpoints.md))
- [X] T035 [P] [US1] Create `src/app/api/auth/register/route.ts` — POST → Go `POST /api/users`; on 201 respond `201 { message }` and set NO cookies; pass through 400/429 (generic on conflict)
- [X] T036 [P] [US1] Create `src/app/api/auth/refresh/route.ts` — POST (no body) reads `arsen_refresh` → Go `POST /api/tokens`; on 200 rotate both cookies, respond 204; on 401 clear cookies, respond 401
- [X] T037 [P] [US1] Create `src/app/api/auth/logout/route.ts` — DELETE reads `arsen_access` → Go `DELETE /api/sessions/current` (Bearer); ALWAYS clear both cookies afterward even on upstream failure; respond 204 (FR-004 ordering)
- [X] T038 [P] [US1] Create `src/app/api/me/route.ts` — GET reads `arsen_access` → Go `GET /api/users/me` with silent-refresh-on-401 (one retry via server.ts); on success 200 `UserProfile`; terminal 401 clears cookies, 401

### Feature slice for User Story 1

- [X] T039 [P] [US1] Create `src/features/auth/schemas/login.ts` and `register.ts` — export schema + inferred type (`LoginInput`, `RegisterInput`) as named exports (makes T029 pass)
- [X] T040 [P] [US1] Create `src/features/auth/types.ts` — `UserProfile` (`id`, `email`, `emailVerified`, `createdAt`); explicitly no token fields
- [X] T041 [P] [US1] Create `src/features/auth/store.ts` — `authStore` (Zustand `persist`, partialize to profile only): `user`, `isAuthenticated`; actions `setUser`, `clearAuth`; granular selectors (makes T032 constraints enforceable)
- [X] T042 [P] [US1] Create `src/features/auth/fetchers/` — `login`, `register`, `logout`, `me` calling BFF relative URLs via `lib/api/client.ts` (reads flow fetcher→hook→component, FR-014)
- [X] T043 [US1] Create `src/features/auth/hooks/use-login.ts` — RHF + zodResolver + BFF login fetcher; on success `setUser` + redirect to `next` or `(app)`; maps problem+json `errors[]` to `setError`, generic fallback (makes T030 pass; FR-002/FR-010)
- [X] T044 [US1] Create `src/features/auth/hooks/use-register.ts` — RHF + zodResolver + register fetcher; on success redirect to `/login?registered=1`; no cookies/tokens touched (makes T031 pass; FR-001 adjusted per R-2)
- [X] T045 [US1] Create `src/features/auth/hooks/use-session.ts` — the single documented session accessor: SWR `'/api/me'` reconciled against persisted profile (makes T033 pass; FR-009)
- [X] T046 [US1] Create `src/features/auth/hooks/use-logout.ts` — calls logout fetcher, then `clearAuth`, then redirect to `/login`; completes client cleanup even if server revoke fails (FR-004)
- [X] T047 [P] [US1] Create `src/features/auth/components/login-form.tsx` (uses use-login, shows loading/success/error) and `register-form.tsx` (uses use-register, shows verify notice on success)
- [X] T048 [P] [US1] Create `src/features/auth/components/user-menu.tsx` — profile display + logout action (consumes use-session, use-logout)

### Routes/pages for User Story 1

- [X] T049 [P] [US1] Create `src/app/(auth)/login/page.tsx` (renders LoginForm; reads `?registered=1` notice and `?next=`) and `src/app/(auth)/register/page.tsx` (renders RegisterForm)
- [X] T050 [US1] Create `src/app/(app)/layout.tsx` — protected group layout: consumes `useSession`, renders app shell + `UserMenu`; middleware already gates entry (FR-005)
- [X] T051 [US1] Create `src/app/(app)/dashboard/page.tsx` — minimal authenticated landing (mostly static, no SWR) as the post-login destination

**Checkpoint**: US1 fully functional — MVP. Auth works end-to-end; token-secrecy test green. Deployable/demoable on its own.

---

## Phase 4: User Story 2 - Add a new feature in one folder (Priority: P2)

**Goal**: A forked engineer can add a complete slice in one directory following a documented guide, and tooling prevents cross-feature internal imports and server-data-in-UI-state mistakes.

**Independent Test**: Following only the guide, add a throwaway second slice that reads/writes through the prescribed chain without editing other slices; confirm a cross-feature internal import raises a lint error.

- [X] T052 [P] [US2] Write failing lint-rule test/fixture `src/features/__tests__/import-boundary.test.ts` (or an eslint fixture file) proving a cross-feature internal import (`@/features/auth/store` from `resource`) is flagged (FR-012/FR-016, US2 scenario 2)
- [X] T053 [US2] Harden `apps/web/eslint.config.mjs` `no-restricted-imports` patterns so only a feature's public entry is importable across features (and the documented `auth` → app-shell `useSession` exception is allowed); make T052 pass
- [X] T054 [P] [US2] Create `apps/web/docs/ARCHITECTURE.md` — the step-by-step "add a feature slice" guide: folder layout, read chain (fetcher→hook→component), write chain (schema→resolver→form→mutation), server-vs-UI-state rules, the import-boundary rule and its one exception (FR-011/FR-013/FR-014/FR-015; SC-001 target <30 min)
- [X] T055 [P] [US2] Create `apps/web/docs/DECISIONS.md` — one-line justification + rejected alternative for every runtime dependency (research R-8), the cookie/BFF strategy (R-4), the in-memory resource store as a labeled stand-in (R-7), and the two spec deviations (R-1/R-2) (FR-015; SC-009)
- [X] T056 [P] [US2] Create `apps/web/README.md` — clone → working login in under 5 minutes: prerequisites, env, backend up, pre-verified-account note, `pnpm dev`, quality gates (FR-028; SC-005)

**Checkpoint**: Conventions documented and lint-enforced; a new slice can be added in one folder with boundaries guarded.

---

## Phase 5: User Story 3 - Worked CRUD resource slice (Priority: P3)

**Goal**: One generic `resource` slice demonstrating list/detail/create/update/delete with RSC→SWR first paint, optimistic create + rollback, and persisted drafts — server-state (SWR) and UI-state (Zustand) strictly separate.

**Independent Test**: In the running app, view the list (server-rendered first paint, no double fetch), open a detail, create a resource (appears optimistically, reconciled on server response; rolls back on forced error), edit and delete it, with loading/success/error states throughout; close+reopen the create form to see the draft restored.

> ⚠️ **NOTE (research R-7)**: no upstream backend resources endpoint exists — the BFF backs onto an in-memory, per-server-process store, clearly labeled as a rename-me stand-in.

### Tests for User Story 3 (write first, ensure they FAIL) ⚠️

- [X] T057 [P] [US3] Write failing test `src/features/resource/schemas/schema.test.ts` — `resourceSchema` (title 1–120 required, description ≤2000 optional, status enum draft/active/archived) (data-model)
- [X] T058 [P] [US3] Write failing test `src/features/resource/hooks/use-create-resource.test.ts` (MSW) — optimistic insert via bound `mutate`; on forced error the optimistic row rolls back (FR-019, SC-006)
- [X] T059 [P] [US3] Write failing test `src/features/resource/hooks/use-resources.test.ts` (MSW) — list hook hydrates from SWR `fallback` and revalidates without a duplicate fetch (FR-018, US3 scenario 1)
- [X] T060 [P] [US3] Write failing test `src/features/resource/store.test.ts` — draft persists across close/reopen and clears on successful submit; `selectedIds`/`wizardStep` are NOT persisted (FR-021)

### In-memory store + BFF handlers for User Story 3

- [X] T061 [US3] Create `src/lib/resource-store.ts` (server-only, module-level `Map`) — seeded rows, CRUD functions; header comment labeling it a stand-in to repoint at a real backend (research R-7)
- [X] T062 [P] [US3] Create `src/app/api/resources/route.ts` — GET list (optional `?status=` filter) → `200 Resource[]`; POST `CreateResourceInput` → `201 Resource` / `400` problem+json mirroring backend `errors:[{field,detail}]` (contracts/bff-endpoints)
- [X] T063 [P] [US3] Create `src/app/api/resources/[id]/route.ts` — GET `200`/`404`; PATCH `UpdateResourceInput` `200`/`404`/`400`; DELETE `204`/`404`

### Feature slice for User Story 3

- [X] T064 [P] [US3] Create `src/features/resource/types.ts` — `Resource`, `CreateResourceInput`, `UpdateResourceInput = Partial<CreateResourceInput>` (data-model)
- [X] T065 [P] [US3] Create `src/features/resource/schemas/resource.ts` — `resourceSchema` + inferred type, named exports (makes T057 pass)
- [X] T066 [P] [US3] Create `src/features/resource/store.ts` — `resourceStore` (persist partialize: `filters` + `draft` only): `selectedIds`, `filters`, `draft`, `wizardStep`; selection/filter/draft/wizard actions (makes T060 pass)
- [X] T067 [P] [US3] Create `src/features/resource/fetchers/` — list, detail, create, update, delete against BFF relative URLs via `lib/api/client.ts`
- [X] T068 [US3] Create `src/features/resource/hooks/use-resources.ts` (list, consumes SWR `fallback`) and `use-resource.ts` (detail, `['/api/resources', id]`, null when `!id`) (makes T059 pass; FR-018/FR-020)
- [X] T069 [US3] Create `src/features/resource/hooks/use-create-resource.ts` — `useSWRMutation` with optimistic insert via bound `mutate` on the list key + rollback on error (makes T058 pass; FR-019)
- [X] T070 [P] [US3] Create `src/features/resource/hooks/use-update-resource.ts` and `use-delete-resource.ts` — mutate + revalidate relevant caches, loading/success/error states (FR-020, US3 scenario 3)
- [X] T071 [P] [US3] Create `src/features/resource/components/resource-list.tsx`, `resource-card.tsx`, `resource-filters.tsx` (list + filter UI, loading/empty/error states)
- [X] T072 [US3] Create `src/features/resource/components/resource-form.tsx` — RHF + zodResolver(resourceSchema); draft restore on mount; on successful submit order: `reset` → `clearDraft` → `closeModal`; guards double-submit while a mutation is in flight (FR-021, edge cases)

### Routes/pages for User Story 3

- [X] T073 [US3] Create `src/app/(app)/resources/page.tsx` (RSC) — server-fetch initial list via the shared server store fn, pass into `<SWRConfig fallback>` boundary for `useResources` (FR-018, research R-6)
- [X] T074 [US3] Create `src/app/(app)/resources/[id]/page.tsx` (RSC) — server-fetch detail as `fallback` for `useResource(id)`

**Checkpoint**: Full CRUD chain works with optimistic create, RSC→SWR first paint, and persisted drafts.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: E2E, coverage, and final validation across all stories.

- [X] T075 [US1] Create `apps/web/e2e/happy-path.spec.ts` (Playwright) — register (→ verify notice) → login with a pre-verified account → land in app → create a resource → see it listed → logout → confirm `/resources` redirects to `/login` (FR-027, SC-008; note R-2 pre-verified-account leg)
- [X] T076 [P] Verify coverage ≥80% on `src/features/**` hooks/fetchers/schemas; add any missing unit tests (SC-007)
- [X] T077 [P] Add `src/app/(app)/resources` filter/empty/error state polish and consistent loading skeletons across read/write paths (FR-020)
- [X] T078 Run `pnpm check` (typecheck + lint + test) and `pnpm e2e` — resolve to zero type errors, zero lint warnings, green tests (FR-025, SC-005)
- [ ] T079 Walk through all 11 scenarios in [quickstart.md](./quickstart.md) manually against the running backend; fix any gaps

---

## Phase 7: Component Documentation (Storybook)

**Purpose**: Stand up Storybook as the canonical, searchable catalog of reusable UI so engineers check for an existing component before building a new one (FR-029/FR-030/FR-031, SC-010). Depends on the reusable primitives existing (Phases 3–5).

- [X] T080 Install and configure Storybook for the web app: add dev deps (`storybook`, `@storybook/nextjs`, `@storybook/addon-docs`) via `pnpm add -D`, create `apps/web/.storybook/main.ts` (Next.js framework preset, stories glob `../src/components/**/*.stories.tsx`, docs addon) and `apps/web/.storybook/preview.ts` (import Tailwind `../src/app/globals.css`, enable `autodocs` tag); add `storybook` and `build-storybook` scripts to `apps/web/package.json`
- [X] T081 [P] Write co-located stories for all existing reusable components — one `*.stories.tsx` per file under `apps/web/src/components/ui/`: `button.stories.tsx` (all variants/sizes), `card.stories.tsx`, `dialog.stories.tsx`, `input.stories.tsx`, `label.stories.tsx`, `sonner.stories.tsx` — each covering primary states/variants
- [X] T082 Enable the docs addon for auto-generated per-component docs: apply the `autodocs` tag (globally in `.storybook/preview.ts` or per story) and wire `react-docgen-typescript` so props/variants tables generate from TypeScript; verify a Docs page renders for each component
- [X] T083 [P] Add a contribution note to `apps/web/docs/ARCHITECTURE.md` and `apps/web/README.md`: engineers MUST check Storybook (`pnpm storybook`) for an existing component before creating a new one, to avoid duplicating UI
- [X] T084 Wire `build-storybook` into CI as a build gate (fails on a broken/undocumented component) and upload `storybook-static/` as an artifact; optionally publish it to a static host so the catalog is browsable without a local checkout

**Checkpoint**: `pnpm build-storybook` succeeds; every reusable component has a searchable, auto-documented story; the "check before building" step is recorded in the contribution docs.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies — start immediately.
- **Foundational (Phase 2)**: depends on Setup — BLOCKS all user stories.
- **User Stories (Phase 3–5)**: all depend on Foundational. US1 is the MVP. US2 and US3 depend only on Foundational (US2 hardens the same eslint config touched in T005; US3 consumes shared `lib/*` + test harness). US3's E2E leg (T075) also needs US1's auth flow.
- **Polish (Phase 6)**: depends on the desired user stories being complete (T075 needs US1 + US3).
- **Component Documentation (Phase 7)**: depends on the reusable primitives existing (Phases 3–5); independent of Polish and can run in parallel with it.

### User Story Dependencies

- **US1 (P1)**: independent once Foundational is done. Delivers the MVP.
- **US2 (P2)**: independent — documentation + lint hardening. Best done after US1 exists so the guide/DECISIONS can reference real slices, but not code-blocked by it.
- **US3 (P3)**: independent slice; its E2E happy path (T075, Polish) requires US1's login to be working.

### Within Each User Story

- Tests are written FIRST and must FAIL before implementation (TDD, project memory).
- Schemas/types → store → fetchers → BFF handlers → hooks → components → pages.

### Parallel Opportunities

- Setup: T005–T008 in parallel; T010/T011 after deps installed.
- Foundational: T012/T013/T015/T018/T019/T020/T021/T026/T027/T028 are independent files — parallelizable; T016 depends on T013+T015; T024 before T025.
- US1: all test tasks T029–T033 in parallel; BFF handlers T034–T038 in parallel; schemas/types/store/fetchers T039–T042 in parallel; then hooks (T043–T046) which depend on their fetchers/store.
- US3: test tasks T057–T060 in parallel; BFF T062/T063 after T061; slice files T064–T067 in parallel.
- Across stories: once Foundational is done, US1 / US2 / US3 can be staffed to different developers in parallel (US3's E2E leg lands last).

---

## Parallel Example: User Story 1 tests

```bash
# Write all US1 tests together first (they must FAIL):
Task: "schemas.test.ts — login/register schema validation"
Task: "use-login.test.ts — success/403/401 paths (MSW)"
Task: "use-register.test.ts — redirect + no-token assertion (MSW)"
Task: "token-secrecy.test.ts — token absent from all client surfaces"
Task: "use-session.test.ts — hydrate + reconcile against /api/me"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1 Setup → 2. Phase 2 Foundational → 3. Phase 3 US1 → **STOP and VALIDATE**: auth end-to-end, token-secrecy test green. Deploy/demo — this is the boilerplate's core.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. US1 → test independently → demo (MVP: secure auth).
3. US2 → docs + lint boundaries → a new hire can add a slice.
4. US3 → worked CRUD slice with optimistic create.
5. Polish → E2E + coverage + quickstart validation.

### Parallel Team Strategy

After Foundational: Dev A → US1, Dev B → US2 (docs/lint), Dev C → US3; converge on Polish (T075 E2E needs US1 + US3).

---

## Notes

- [P] = different files, no dependency on incomplete tasks.
- Every task has an explicit file path; commit after each task or logical group.
- Verify tests FAIL before implementing (TDD required — project memory).
- Highest-priority single test: T032 token-secrecy (FR-007/FR-008, SC-002).
- Two spec deviations are load-bearing (research R-1/R-2 email verification; R-7 in-memory resource store) — surfaced in tasks T031/T034/T035/T044/T061 and documented in T055.
