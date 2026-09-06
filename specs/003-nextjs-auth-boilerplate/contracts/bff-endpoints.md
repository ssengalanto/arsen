# BFF Route-Handler Contract (Next `app/api/*`)

The browser talks **only** to these handlers. Each proxies to the Go backend (or the in-memory resource store) server-side, manages httpOnly cookies, and never returns a token to the client. Errors are re-emitted as `application/problem+json` so the client `ApiError` and error boundary work uniformly.

## Auth

### POST /api/auth/login
- Body: `{ email, password }` → proxied to `POST /api/sessions`.
- On 200: set `arsen_access` (Max-Age = `expiresIn`) + `arsen_refresh` cookies; respond `200 { user: UserProfile }` (profile fetched via `/api/users/me`, or minimal from token) — **no tokens in body**.
- On 401/403/400/429: pass through the ProblemDetail (403 → surface "verify your email").

### POST /api/auth/register
- Body: `{ email, password }` → proxied to `POST /api/users`.
- On 201: respond `201 { message }` and DO NOT set cookies (register does not authenticate). Client redirects to `/login?registered=1`.
- On 400/429: pass through ProblemDetail (generic on conflict).

### POST /api/auth/refresh
- No body. Reads `arsen_refresh` cookie → `POST /api/tokens { refreshToken }`.
- On 200: rotate both cookies; respond `204`. On 401: clear cookies; respond `401`.
- Primarily called internally by the server fetch helper; also exposed for an explicit client-triggered refresh if needed (never from `useEffect`).

### DELETE /api/auth/logout
- Reads `arsen_access` → `DELETE /api/sessions/current` (Bearer). Always clears both cookies afterward, even if upstream fails. Responds `204`.

### GET /api/me
- Reads `arsen_access` → `GET /api/users/me` (with silent-refresh-on-401, one retry).
- On success: `200 UserProfile`. On terminal 401: clear cookies, `401`.

## Resource (in-memory store stand-in — see research R-7)

### GET /api/resources
- Optional query filters (e.g. `?status=active`). `200 Resource[]`.

### POST /api/resources
- Body: `CreateResourceInput` `{ title, description?, status }`. `201 Resource`. `400 ProblemDetail` on validation failure (shape mirrors backend: `errors:[{field,detail}]`).

### GET /api/resources/[id]
- `200 Resource` · `404 ProblemDetail`.

### PATCH /api/resources/[id]
- Body: `UpdateResourceInput`. `200 Resource` · `404` · `400`.

### DELETE /api/resources/[id]
- `204` · `404`.

## Cross-cutting
- All handlers run on the Node runtime (need cookies + server fetch).
- Cookie attributes: `httpOnly; SameSite=Lax; Path=/; Secure` (prod). Presence of the access/refresh cookie is what `middleware.ts` gates on.
- Client fetchers (`lib/api/client.ts`) call these relative URLs with `credentials: 'include'` and throw `ApiError` on non-2xx.
