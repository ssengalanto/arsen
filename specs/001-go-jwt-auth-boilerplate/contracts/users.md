# API Contract: Users

All endpoints use JSON request/response bodies. Errors conform to RFC 9457 (`application/problem+json`). All responses include `self` and `kind` properties. Property names use camelCase. Dates use ISO 8601 UTC.

---

## POST /api/users — Register

Creates a new user account with unverified email. Sends verification email via Resend.

**Rate limit**: 10 req/min per IP

**Request**:
```json
{
  "email": "alice@example.com",
  "password": "S3cure!Pass"
}
```

**Response 201 Created**:
```
Location: /api/users/{id}
```
```json
{
  "self": "/api/users/{id}",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": false,
  "createdAt": "2026-09-04T12:30:00Z",
  "message": "Account created. Please check your email to verify your address."
}
```

**Response 400 (validation)**:
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 400,
  "detail": "One or more fields failed validation.",
  "instance": "/api/users",
  "errors": [
    {
      "field": "password",
      "detail": "Must be at least 8 characters with uppercase, lowercase, digit, and special character."
    }
  ]
}
```

**Response 409/400 (duplicate email)**: Returns a generic error indistinguishable from other registration failures (no user enumeration).

**Response 429**: Rate limit exceeded.

---

## GET /api/users/me — Get Profile

Returns the authenticated user's profile.

**Auth**: Required (Bearer access token)

**Response 200 OK**:
```json
{
  "self": "/api/users/me",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": true,
  "createdAt": "2026-09-04T12:30:00Z"
}
```

**Response 401**: Missing, expired, or invalid access token.

---

## POST /api/users/verify — Verify Email

Verifies a user's email address using a verification token. Triggers welcome email on success.

**Rate limit**: 10 req/min per IP

**Request**:
```json
{
  "token": "a1b2c3d4e5f6..."
}
```

**Response 200 OK**:
```json
{
  "self": "/api/users/{id}",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": true,
  "createdAt": "2026-09-04T12:30:00Z",
  "message": "Email verified successfully."
}
```

**Response 401**: Token expired, already used, or invalid. All cases return identical response body.

**Response 429**: Rate limit exceeded.

---

## POST /api/users/resend-verification — Resend Verification Email

Resends a verification email. Always returns success regardless of email existence (no user enumeration).

**Rate limit**: 5 req/min per IP

**Request**:
```json
{
  "email": "alice@example.com"
}
```

**Response 200 OK**:
```json
{
  "kind": "Acknowledgment",
  "message": "If an account exists with this email and is not yet verified, a new verification email has been sent."
}
```

**Response 429**: Rate limit exceeded.

---

## POST /api/users/forgot-password — Request Password Reset

Sends a password reset email. Always returns 200 regardless of email existence (no user enumeration). Does not send email if user's email is unverified.

**Rate limit**: 5 req/min per IP

**Request**:
```json
{
  "email": "alice@example.com"
}
```

**Response 200 OK**:
```json
{
  "kind": "Acknowledgment",
  "message": "If an account exists with this email, a password reset link has been sent."
}
```

**Response 429**: Rate limit exceeded.

---

## POST /api/users/reset-password — Reset Password

Resets the user's password using a reset token. Revokes all refresh token families for the user.

**Rate limit**: 10 req/min per IP

**Request**:
```json
{
  "token": "x9y8z7w6v5u4...",
  "password": "N3w!S3cure"
}
```

**Response 200 OK**:
```json
{
  "kind": "Acknowledgment",
  "message": "Password has been reset successfully. Please log in with your new password."
}
```

**Response 400 (validation)**:
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 400,
  "detail": "One or more fields failed validation.",
  "instance": "/api/users/reset-password",
  "errors": [
    {
      "field": "password",
      "detail": "Must be at least 8 characters with uppercase, lowercase, digit, and special character."
    }
  ]
}
```

**Response 401**: Token expired, already used, or invalid. All cases return identical response body.

**Response 429**: Rate limit exceeded.
