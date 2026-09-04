# Quickstart Validation Guide

**Purpose**: Runnable validation scenarios to prove the feature works end-to-end.

## Prerequisites

- Docker and Docker Compose installed
- Task (go-task.dev) installed
- Go (latest stable) installed (for running tests on host)
- A Resend API key (optional for dev mode — no-op email sender logs tokens)

## Setup

```bash
# Clone and enter the project
git clone {{MODULE_PATH}}
cd {{PROJECT_NAME}}

# Copy environment template
cp .env.example .env
# Edit .env with your values (or use defaults for dev mode)

# Start the full stack (API + PostgreSQL + Redis)
task dev
# OR
docker compose up
```

The API should be available at `http://localhost:${PORT}` within seconds.

## Validation Scenarios

### Scenario 1: Health Checks

```bash
# Liveness — should return 200
curl -s http://localhost:8080/healthz | jq .

# Readiness — should return 200 (database connected)
curl -s http://localhost:8080/readyz | jq .
```

**Expected**: Both return 200 with `status: "ok"` / `status: "ready"`.

---

### Scenario 2: Register → Verify → Login → Profile → Refresh → Logout

This is the golden path covering the complete user lifecycle.

**Step 1: Register**
```bash
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test!234"}' | jq .
```
**Expected**: 201 Created. Response includes `emailVerified: false` and a message to check email. In dev mode, the verification token is logged to stdout.

**Step 2: Verify Email**
```bash
# Extract the verification token from dev logs or Resend dashboard
TOKEN="<token-from-logs>"

curl -s -X POST http://localhost:8080/api/users/verify \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | jq .
```
**Expected**: 200 OK. Response includes `emailVerified: true`. Welcome email is dispatched (logged in dev mode).

**Step 3: Login**
```bash
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test!234"}' | jq .
```
**Expected**: 200 OK. Response includes `accessToken`, `refreshToken`, `expiresIn: 900`.

**Step 4: Access Profile**
```bash
ACCESS_TOKEN="<access-token-from-step-3>"

curl -s http://localhost:8080/api/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```
**Expected**: 200 OK. Response includes `email`, `emailVerified: true`, `createdAt`.

**Step 5: Refresh Token**
```bash
REFRESH_TOKEN="<refresh-token-from-step-3>"

curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}" | jq .
```
**Expected**: 200 OK. New `accessToken` and `refreshToken` returned. Old refresh token is now invalid.

**Step 6: Logout**
```bash
NEW_ACCESS_TOKEN="<access-token-from-step-5>"

curl -s -X DELETE http://localhost:8080/api/sessions/current \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN"
```
**Expected**: 204 No Content.

**Step 7: Verify Refresh Token Revoked**
```bash
NEW_REFRESH_TOKEN="<refresh-token-from-step-5>"

curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$NEW_REFRESH_TOKEN\"}" | jq .
```
**Expected**: 401 Unauthorized. The entire token family was revoked on logout.

---

### Scenario 3: Login Blocked for Unverified Email

```bash
# Register a new user (don't verify)
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"unverified@example.com","password":"Test!234"}' | jq .

# Attempt login
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"unverified@example.com","password":"Test!234"}' | jq .
```

**Expected**: 403 Forbidden with message indicating email verification is required.

---

### Scenario 4: Password Reset Flow

```bash
# Use the verified user from Scenario 2
# Step 1: Request password reset
curl -s -X POST http://localhost:8080/api/users/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' | jq .
```
**Expected**: 200 OK. Generic acknowledgment. Reset token logged in dev mode.

```bash
# Step 2: Reset password
RESET_TOKEN="<reset-token-from-logs>"

curl -s -X POST http://localhost:8080/api/users/reset-password \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$RESET_TOKEN\",\"password\":\"N3w!Pass5\"}" | jq .
```
**Expected**: 200 OK. Password changed. All refresh token families revoked.

```bash
# Step 3: Login with new password
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"N3w!Pass5"}' | jq .
```
**Expected**: 200 OK. New tokens issued.

```bash
# Step 4: Login with old password fails
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test!234"}' | jq .
```
**Expected**: 401 Unauthorized.

---

### Scenario 5: Refresh Token Reuse Detection

```bash
# Login to get tokens
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"N3w!Pass5"}' | jq .

REFRESH_1="<refresh-token>"

# Rotate the refresh token
curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_1\"}" | jq .

REFRESH_2="<new-refresh-token>"

# Replay the OLD refresh token (reuse attack)
curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_1\"}" | jq .
```

**Expected**: 401 Unauthorized. The entire token family is revoked. Even `REFRESH_2` no longer works:

```bash
curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_2\"}" | jq .
```
**Expected**: 401 Unauthorized.

---

### Scenario 6: Anti-Enumeration Checks

```bash
# Forgot password for non-existent email
curl -s -w "\nHTTP Status: %{http_code}\nTime: %{time_total}s\n" \
  -X POST http://localhost:8080/api/users/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"nobody@example.com"}'

# Forgot password for existing email
curl -s -w "\nHTTP Status: %{http_code}\nTime: %{time_total}s\n" \
  -X POST http://localhost:8080/api/users/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

**Expected**: Both return 200 OK with identical response body. Response times should be within ~100ms of each other.

---

### Scenario 7: 405 Method Not Allowed

```bash
# GET on a POST-only endpoint
curl -s -X GET http://localhost:8080/api/sessions -D - | head -5
```

**Expected**: 405 Method Not Allowed with `Allow: POST` header.

---

## Automated Test Execution

```bash
# Run all tests
task test

# Run with verbose output
task test -- -v

# Run specific feature tests
go test ./features/auth/... -v
go test ./features/user/... -v
go test ./features/health/... -v

# Run CQRS package tests
go test ./pkg/cqrs/... -v
```

## Swagger UI

In development mode, visit `http://localhost:8080/swagger/` to browse the interactive API documentation.
