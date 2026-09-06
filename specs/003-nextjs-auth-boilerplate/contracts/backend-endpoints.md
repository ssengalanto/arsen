# Upstream Go Backend Contract (as-built)

Verified against `apps/api/features/auth/handler.go`, `features/user/handler.go`, and `pkg/response/problem.go`. Base URL from env `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`). All error bodies are `application/problem+json`.

## Shared shapes

```jsonc
// sessionResponse (login, refresh)
{ "self": "string", "kind": "Session|TokenPair", "accessToken": "jwt",
  "refreshToken": "opaque", "tokenType": "Bearer", "expiresIn": 900 }

// userResponse (register, me, verify)
{ "self": "string", "kind": "User", "id": "uuid", "email": "string",
  "emailVerified": false, "createdAt": "RFC3339", "message": "string?" }

// ProblemDetail (all errors)
{ "type": "string", "title": "string", "status": 400, "detail": "string",
  "instance": "string", "errors": [{ "field": "email", "detail": "is required" }] }
```

## Endpoints

### POST /api/sessions — login
- Body: `{ "email": string, "password": string }`
- 200: `sessionResponse`
- 400 invalid body/validation · 401 invalid credentials · **403 email not verified** · 429 rate limited
- Rate limit: ~10/min, burst 20.

### POST /api/tokens — refresh
- Body: `{ "refreshToken": string }`
- 200: `sessionResponse` (`kind: "TokenPair"`)
- 401 invalid/expired/rotated · 429
- Note: refresh rotates the token family; reuse of a revoked token invalidates the family.

### DELETE /api/sessions/current — logout
- Auth: `Authorization: Bearer <access>`
- 204 no content (revokes all refresh tokens for the user) · 401

### POST /api/users — register
- Body: `{ "email": string, "password": string }`
- 201: `userResponse` (**no tokens**; `emailVerified:false`)
- 400 (validation OR normalized conflict — anti-enumeration, so a taken email also returns a generic 400, not a field error) · 429

### GET /api/users/me — profile
- Auth: `Authorization: Bearer <access>`
- 200: `userResponse` · 401

## Consequences for the BFF
- Tokens arrive in the **response body**; the BFF converts them to httpOnly cookies (never returned to the browser JS).
- Register never authenticates — see [research R-2](../research.md). Login is blocked (403) until the email is verified.
- Validation errors carry `errors[]` for field mapping; registration conflicts do **not** (generic 400).
