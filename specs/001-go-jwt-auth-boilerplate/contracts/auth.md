# API Contract: Auth (Sessions & Tokens)

All endpoints use JSON request/response bodies. Errors conform to RFC 9457 (`application/problem+json`). All responses include `self` and `kind` properties. Property names use camelCase. Dates use ISO 8601 UTC.

---

## POST /api/sessions — Login

Authenticates a user with email and password. Returns access + refresh tokens. Requires verified email.

**Rate limit**: 10 req/min per IP

**Request**:
```json
{
  "email": "alice@example.com",
  "password": "S3cure!Pass"
}
```

**Response 200 OK**:
```json
{
  "self": "/api/sessions/current",
  "kind": "Session",
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "a1b2c3d4e5f6g7h8i9j0...",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

**Response 401 (wrong credentials or user not found)**: Generic authentication failure. Same response body and comparable response time for both "user not found" and "wrong password" cases.
```json
{
  "type": "about:blank",
  "title": "Authentication Failed",
  "status": 401,
  "detail": "Invalid email or password.",
  "instance": "/api/sessions"
}
```

**Response 403 (email not verified)**:
```json
{
  "type": "about:blank",
  "title": "Email Not Verified",
  "status": 403,
  "detail": "You must verify your email address before logging in. Check your inbox or request a new verification email.",
  "instance": "/api/sessions"
}
```

**Response 429**: Rate limit exceeded.

---

## POST /api/tokens — Refresh Token

Exchanges a valid refresh token for a new access token and new refresh token. Old refresh token is invalidated (rotated).

**Request**:
```json
{
  "refreshToken": "a1b2c3d4e5f6g7h8i9j0..."
}
```

**Response 200 OK**:
```json
{
  "self": "/api/tokens",
  "kind": "TokenPair",
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "k1l2m3n4o5p6q7r8s9t0...",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

**Response 401**: Refresh token expired, already rotated (reuse detection triggers family revocation), or invalid.
```json
{
  "type": "about:blank",
  "title": "Authentication Failed",
  "status": 401,
  "detail": "Invalid or expired refresh token.",
  "instance": "/api/tokens"
}
```

---

## DELETE /api/sessions/current — Logout

Revokes the entire refresh token family for the current session.

**Auth**: Required (Bearer access token)

**Response 204 No Content**: Success. Idempotent — calling again with a revoked token still returns 204.

**Response 401**: Missing, expired, or invalid access token.
