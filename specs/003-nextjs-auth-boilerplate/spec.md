# Feature Specification: Next.js Feature-Sliced Frontend Boilerplate

**Feature Branch**: `003-nextjs-auth-boilerplate`

**Created**: 2026-09-05

**Status**: Draft

**Input**: User description: "Production-grade, reusable Next.js App Router boilerplate (feature-sliced; SWR for server data, Zustand for UI state, react-hook-form + zod for forms, shadcn/ui primitives). Frontend counterpart to the existing Go JWT backend that returns RFC 9457 problem+json errors. Deliver end-to-end authentication securely (access token never readable by JavaScript) plus one generic CRUD `resource` slice that demonstrates the full data-fetching chain. It must run, type-check, lint, and test clean."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Secure end-to-end authentication (Priority: P1)

An end user of an app forked from this boilerplate can register an account, log in, stay signed in across page reloads, be blocked from protected pages when unauthenticated, and log out. Throughout, the access credential is never exposed to client-side JavaScript.

**Why this priority**: Authentication is the core deliverable and the hardest thing to get right and secure. A boilerplate that gets auth wrong is worse than none. This slice alone — register, login, protected area, logout — is a viable MVP that every forked app immediately relies on.

**Independent Test**: With the Go backend reachable, exercise register → land in the protected app → reload (still signed in) → log out → attempt a protected route and get redirected to login. Separately, assert via automated test that the access token is absent from every client-readable surface (Zustand state, `localStorage`, `sessionStorage`, and any client-serialized output).

**Acceptance Scenarios**:

1. **Given** a visitor on the register page, **When** they submit valid unique account details, **Then** an account is created, a session is established, and they are redirected into the protected area.
2. **Given** a registered user on the login page, **When** they submit correct credentials, **Then** a session is established and they are redirected to the protected app (or to the `next` destination they were bounced from).
3. **Given** an authenticated user, **When** they reload the page or open a new tab, **Then** they remain signed in and their profile is shown without re-entering credentials.
4. **Given** an unauthenticated visitor, **When** they navigate directly to any protected route, **Then** they are redirected to the login page with the intended destination preserved.
5. **Given** an authenticated user whose access credential is near or at expiry, **When** they perform an action, **Then** the credential is rotated behind the scenes and the action completes without the user re-authenticating.
6. **Given** an authenticated user, **When** they log out, **Then** the server-side session is revoked, the client profile is cleared, and they are redirected to login; afterwards protected routes are inaccessible.
7. **Given** a user submitting invalid credentials or a duplicate registration, **When** the backend returns a problem+json error, **Then** the form surfaces a field-level or generic message without leaking raw backend text, tokens, or PII.

---

### User Story 2 - Add a new feature in one folder (Priority: P2)

A new engineer who forks the boilerplate can add a complete feature — its UI, hooks, data fetchers, validation schema, store, and types — inside a single feature directory, following a documented step-by-step guide, without editing unrelated code.

**Why this priority**: Maintainability is the boilerplate's primary reason to exist. The value compounds across every future feature and engineer. It depends on the architectural conventions being present and demonstrated, which the auth slice (P1) already begins to establish.

**Independent Test**: Following only the architecture guide, a developer creates a second feature slice that reads and writes data through the prescribed chain, and it works without touching other slices. Confirm cross-feature imports are prevented (a lint error is raised if one feature imports another's internals).

**Acceptance Scenarios**:

1. **Given** the architecture guide, **When** a developer adds a new slice per the steps, **Then** all of its concerns live in one directory and no unrelated files are modified.
2. **Given** the conventions, **When** a developer attempts to import one feature's internals from another feature, **Then** the tooling flags it as an error.
3. **Given** the conventions, **When** a developer tries to fetch data inside a component or place server data into UI state, **Then** the documented rules and (where enforceable) tooling steer them to the correct pattern.

---

### User Story 3 - Worked CRUD resource slice (Priority: P3)

An engineer can study one generic `resource` slice that demonstrates the complete data-fetching chain end to end: list, detail, create, update, delete — including an optimistic list update — using server-state and UI-state kept strictly separate.

**Why this priority**: The worked example turns conventions into a copyable pattern, accelerating every future slice. It is valuable but secondary to auth working and the architecture being navigable.

**Independent Test**: In the running app, view the resource list, open a detail, create a new resource and see it appear optimistically before server confirmation, edit it, and delete it — with loading, success, and error states visible for each operation.

**Acceptance Scenarios**:

1. **Given** the resources list, **When** the page loads, **Then** initial data is server-rendered for instant first paint and thereafter kept fresh by client-side revalidation without a duplicate fetch.
2. **Given** the create form, **When** a valid resource is submitted, **Then** it appears in the list immediately (optimistically) and is reconciled with the server result; on failure the optimistic entry is rolled back and an error is shown.
3. **Given** a resource detail, **When** it is edited or deleted, **Then** the change is reflected in the relevant views and the appropriate cache is refreshed.
4. **Given** a form with a draft in progress, **When** the form is closed and reopened, **Then** the draft is restored; on successful submit the draft is cleared.

---

### Edge Cases

- **Expired/invalid session mid-session**: a proxied data call returns unauthorized → the system attempts one credential rotation and retry; if that fails, the user is redirected to login.
- **Stale persisted profile**: the client-cached profile disagrees with the server's current session → the server session is treated as the source of truth and the cache is reconciled.
- **Backend unreachable or malformed error**: forms and pages show a generic, user-safe error and do not crash the error boundary or leak internals.
- **Direct navigation to a protected route with no session** vs. **with an expired session** → both redirect to login preserving the intended destination.
- **Double submit / in-flight mutation**: submitting a form while a mutation is in flight does not fire duplicate writes.
- **Reload during a multi-step draft**: transient selection and wizard step are not restored (avoid confusing state), while intentional draft content is.
- **Logout ordering**: if the server-side revoke fails, the client still clears local state and redirects, and does not leave the user appearing signed in.

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication & session

- **FR-001**: The system MUST let a visitor register a new account and, on success, establish an authenticated session and redirect into the protected area.
- **FR-002**: The system MUST let a registered user log in with correct credentials and establish an authenticated session.
- **FR-003**: The system MUST keep the user signed in across page reloads and new tabs until logout or session expiry.
- **FR-004**: The system MUST let a user log out, which revokes the server-side session, clears the client-held profile, and redirects to login — in that order, and completing the client cleanup even if the server revoke fails.
- **FR-005**: The system MUST redirect unauthenticated requests for protected routes to the login page while preserving the originally requested destination.
- **FR-006**: The system MUST rotate the access credential automatically before/at expiry (or upon an unauthorized proxied response), retrying the original action once, without user interaction.
- **FR-007**: The access credential MUST NOT be readable by client-side JavaScript at any point — it MUST NOT appear in browser web storage, client state, or any client-serialized output. Only a non-sensitive user profile may be cached client-side for UI.
- **FR-008**: The refresh credential MUST never be exposed to the browser; credential rotation MUST be performed server-side.
- **FR-009**: The system MUST expose the current user's profile to the app shell via a single documented session accessor, reconciling any client-cached profile against the authoritative server session.
- **FR-010**: Authentication forms MUST map backend problem-detail errors to field-level messages where applicable, with a safe generic fallback, and MUST NOT display raw backend text, tokens, or personal data, nor log them.

#### Architecture & maintainability

- **FR-011**: Code MUST be organized by feature, with each feature owning its UI, hooks, fetchers, validation schema, store, and types in a single directory; there MUST be no global dumping grounds for these concerns.
- **FR-012**: Feature slices MUST NOT import one another's internals; shared needs MUST be lifted to a shared/common layer, with the single documented exception of the session accessor consumed by the app shell.
- **FR-013**: Server (fetched) data and client/UI state MUST be kept in separate mechanisms and MUST NOT be mixed (no server data in UI state; no ephemeral UI state in the server-data cache).
- **FR-014**: Reads MUST flow fetcher → data hook → component, and writes MUST flow validation schema → resolver → form → mutation; components MUST NOT fetch data directly.
- **FR-015**: The project MUST document a step-by-step guide for adding a new feature slice that a new hire can follow, and MUST include architecture and decision records (each runtime dependency justified with the alternative that was rejected).
- **FR-016**: The tooling MUST enforce the key constraints where feasible (at minimum: block cross-feature internal imports; forbid untyped escape hatches; require the prescribed export style except where the framework mandates otherwise).

#### Data-fetching demonstration (resource slice)

- **FR-017**: The system MUST include exactly one generic CRUD `resource` slice demonstrating list, detail, create, update, and delete, with no placeholder/stub pages elsewhere.
- **FR-018**: Initial list/first-paint data for the demonstration MUST be server-rendered and handed to the client cache so the client hydrates instantly without a duplicate fetch, after which the client owns revalidation.
- **FR-019**: The create operation MUST perform an optimistic list update with rollback on error, using the cache bound to the relevant read.
- **FR-020**: Every read and write path MUST expose loading, success, and error states to the UI.
- **FR-021**: In-progress form drafts MUST survive closing and reopening the form; on successful submit the draft MUST be cleared. Transient selection and multi-step wizard position MUST NOT be restored after reload.

#### Error contract & integration

- **FR-022**: All non-success responses from the backend MUST be surfaced as a typed application error so that error states and the global error boundary function correctly.
- **FR-023**: All calls to the external backend MUST be proxied server-side (the browser MUST NOT call the backend directly with the credential), which also removes the need for cross-origin browser configuration.
- **FR-024**: Route protection MUST, at minimum, verify presence of a session on protected routes; the chosen depth of verification (presence vs. signature) MUST be documented with its rationale.

#### Quality gates

- **FR-025**: A fresh clone MUST install and run a working app against the backend, and MUST pass type-checking with zero errors, linting with zero warnings, and the automated test suite.
- **FR-026**: Automated tests MUST cover the logic-dense parts first (schemas, hooks, the auth flow, and mutation/optimistic paths), including an explicit test that the access credential is absent from all client-readable surfaces, and MUST mock the network rather than hand-stub it.
- **FR-027**: An end-to-end test MUST cover the happy path: register → land in app → create a resource → see it listed → log out → confirm a protected route redirects to login.
- **FR-028**: A README MUST get a developer from clone to a working login in under five minutes.

### Key Entities *(include if feature involves data)*

- **User Profile**: The non-sensitive representation of the signed-in user (identifier, email, display fields) safe to cache client-side for UI. Explicitly excludes any credential.
- **Session**: The authenticated state of a user, held server-side and represented to the browser only as an opaque, non-JavaScript-readable cookie. Source of truth for whether the user is signed in.
- **Access Credential / Refresh Credential**: Short-lived and long-lived proofs of authentication managed entirely server-side; never exposed to the browser.
- **Resource**: A generic domain record used to demonstrate the CRUD chain (list/detail/create/update/delete). Its concrete fields are placeholders to be renamed per real feature.
- **Draft**: A partially completed resource being authored, retained across form close/reopen and discarded on successful submit.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new engineer can add a complete, working feature slice by following the guide alone, in under 30 minutes, without modifying any unrelated slice.
- **SC-002**: In 100% of automated and manual checks, the access credential is absent from browser web storage, client state, and any client-serialized output.
- **SC-003**: An unauthenticated attempt to reach any protected route results in a redirect to login 100% of the time, with the intended destination preserved.
- **SC-004**: A user remains signed in across page reloads and credential expiry without any manual re-authentication for the lifetime of a valid session.
- **SC-005**: A fresh clone reaches a working login screen in under five minutes following the README, and the type-check, lint, and test gates all pass clean.
- **SC-006**: Creating a resource shows the new item in the list within ~100 ms (optimistically), and a failed create rolls the item back 100% of the time.
- **SC-007**: Automated coverage on feature logic (hooks, fetchers, schemas) is at least 80%.
- **SC-008**: The end-to-end happy path (register → create → logout → protected-route redirect) passes reliably.
- **SC-009**: Every runtime dependency has a one-line justification and a rejected alternative recorded.

## Assumptions

- The existing Go JWT backend is the counterpart API, reachable at a configurable base URL (default `http://localhost:8080`), and returns RFC 9457 `application/problem+json` errors with login/register/refresh/logout endpoints as already implemented in this repo's `apps/api`.
- This boilerplate targets the repo's existing `apps/web` Next.js workspace (App Router), using the latest stable Next.js/React/TypeScript resolved at install time rather than pinned from memory.
- The chosen approach proxies all backend calls through a server-side layer in the web app and stores the session in an httpOnly, Secure, SameSite cookie; the browser never holds a bearer token. (This is the default resolution; the alternative direct-browser-to-backend approach is explicitly not chosen.)
- Server-rendered first paint is used for list/detail data, handed to the client cache as fallback; interactive/mutable data is owned by the client cache thereafter.
- "Resource" is a deliberately generic stand-in intended to be renamed per real feature; no additional demo features or placeholder pages are in scope.
- Design/UI is intentionally raw and easily re-themeable (a boilerplate starting point), built on generated component primitives rather than a bespoke design system.
- Standard modern-browser support; no IE/legacy targets. Mobile-responsive is desirable but not a gating requirement for the boilerplate.
- Email verification, password reset, and other extended auth flows exist in the backend but are out of scope for this frontend boilerplate's initial slice unless later requested; the core is register/login/logout/session/refresh.
