# Data Model: Go Backend Boilerplate with JWT Authentication

**Phase 1 Output** | **Date**: 2026-09-04

## Entities

### User

Represents a registered account. Central entity that owns all token types.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | PK, generated (gen_random_uuid()) | Stable identifier across all references |
| email | citext | UNIQUE, NOT NULL | Case-insensitive via PostgreSQL citext extension |
| password_hash | text | NOT NULL | bcrypt hash, cost >= 12. Never exposed in API responses or logs |
| email_verified | boolean | NOT NULL, DEFAULT false | Gates login access. Transitions to true via verification flow |
| created_at | timestamptz | NOT NULL, DEFAULT now() | ISO 8601 in API responses |
| updated_at | timestamptz | NOT NULL, DEFAULT now() | Updated on password change, email verification |

**Indexes**:
- `users_pkey` on `id` (PK)
- `users_email_key` on `email` (UNIQUE)

**State transitions**:
- `email_verified`: `false` → `true` (via POST /api/users/verify, irreversible)
- `password_hash`: updated via POST /api/users/reset-password
- `updated_at`: auto-updated on any mutation

---

### Refresh Token

Represents a single token in a rotation chain. Tokens within the same family form a linked chain; revoking a family invalidates all tokens in it.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | PK, generated | |
| user_id | UUID | FK → users(id) ON DELETE CASCADE, NOT NULL | Owner |
| token_hash | bytea | NOT NULL | SHA-256 hash of the 256-bit random token value |
| family_id | UUID | NOT NULL | Groups tokens in a rotation chain |
| expires_at | timestamptz | NOT NULL | ~7 days from creation |
| revoked | boolean | NOT NULL, DEFAULT false | Set true on rotation, reuse detection, logout, or password reset |
| created_at | timestamptz | NOT NULL, DEFAULT now() | |

**Indexes**:
- `refresh_tokens_pkey` on `id` (PK)
- `refresh_tokens_token_hash_idx` on `token_hash` (for lookup during refresh)
- `refresh_tokens_family_id_idx` on `family_id` (for family revocation)
- `refresh_tokens_user_id_idx` on `user_id` (for user-level revocation on password reset)
- `refresh_tokens_expires_at_idx` on `expires_at` WHERE `revoked = false` (for cleanup query)

**State transitions**:
- `revoked`: `false` → `true` (on token rotation, reuse detection, logout, or password reset; irreversible)

**Family lifecycle**:
1. Login creates a new token with a new `family_id`
2. Refresh rotates: creates a new token with same `family_id`, revokes the old token
3. Reuse detection: revokes ALL tokens with the same `family_id`
4. Logout: revokes ALL tokens with the same `family_id`
5. Password reset: revokes ALL tokens for the `user_id` (all families)

---

### Verification Token

Represents a single-use token for email verification. Only the latest token per user is valid.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | PK, generated | |
| user_id | UUID | FK → users(id) ON DELETE CASCADE, NOT NULL | Owner |
| token_hash | bytea | NOT NULL | SHA-256 hash of the 256-bit random token value |
| expires_at | timestamptz | NOT NULL | 24 hours from creation |
| used_at | timestamptz | NULL | Set when token is consumed; NULL = unused |
| created_at | timestamptz | NOT NULL, DEFAULT now() | |

**Indexes**:
- `verification_tokens_pkey` on `id` (PK)
- `verification_tokens_token_hash_idx` on `token_hash` (for lookup during verification)
- `verification_tokens_user_id_idx` on `user_id` (for invalidating previous tokens on resend)
- `verification_tokens_expires_at_idx` on `expires_at` WHERE `used_at IS NULL` (for cleanup)

**State transitions**:
- `used_at`: `NULL` → timestamp (when token is consumed via POST /api/users/verify; irreversible)

**Invalidation rules**:
- When a new verification token is generated (resend), all previous unused tokens for that user are marked as used (set `used_at = now()`)
- Cleanup goroutine purges expired and used tokens periodically

---

### Password Reset Token

Represents a single-use token for password reset. Only the latest token per user is valid.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | PK, generated | |
| user_id | UUID | FK → users(id) ON DELETE CASCADE, NOT NULL | Owner |
| token_hash | bytea | NOT NULL | SHA-256 hash of the 256-bit random token value |
| expires_at | timestamptz | NOT NULL | 1 hour from creation |
| used_at | timestamptz | NULL | Set when token is consumed; NULL = unused |
| created_at | timestamptz | NOT NULL, DEFAULT now() | |

**Indexes**:
- `password_reset_tokens_pkey` on `id` (PK)
- `password_reset_tokens_token_hash_idx` on `token_hash` (for lookup during reset)
- `password_reset_tokens_user_id_idx` on `user_id` (for invalidating previous tokens)
- `password_reset_tokens_expires_at_idx` on `expires_at` WHERE `used_at IS NULL` (for cleanup)

**State transitions**:
- `used_at`: `NULL` → timestamp (when token is consumed via POST /api/users/reset-password; irreversible)

**Invalidation rules**:
- When a new reset token is generated (forgot-password), all previous unused tokens for that user are marked as used
- Cleanup goroutine purges expired and used tokens periodically

---

## Entity Relationships

```text
┌──────────┐       ┌─────────────────┐
│   User   │──1:N──│  Refresh Token  │
│          │       │  (family chain) │
│  id (PK) │       │  user_id (FK)   │
│  email   │       │  family_id      │
│  ...     │       └─────────────────┘
│          │
│          │       ┌─────────────────────┐
│          │──1:N──│ Verification Token  │
│          │       │  user_id (FK)       │
│          │       └─────────────────────┘
│          │
│          │       ┌──────────────────────┐
│          │──1:N──│ Password Reset Token │
│          │       │  user_id (FK)        │
└──────────┘       └──────────────────────┘
```

All token tables cascade on user deletion.

## Migration Strategy

Migrations use golang-migrate with versioned SQL file pairs (up/down):

1. **000001_create_users**: Creates `users` table with citext extension, UUID generation
2. **000002_create_refresh_tokens**: Creates `refresh_tokens` table with FK to users
3. **000003_create_verification_tokens**: Creates `verification_tokens` table with FK to users
4. **000004_create_password_reset_tokens**: Creates `password_reset_tokens` table with FK to users

Each down migration drops the table. Extension creation (`CREATE EXTENSION IF NOT EXISTS citext`) goes in migration 000001.

## Validation Rules (from spec)

### User Registration
- `email`: Required, valid email format (validator tag: `required,email`)
- `password`: Required, min 8 chars, at least 1 uppercase, 1 lowercase, 1 digit, 1 special char (custom validator)

### Password Reset
- `token`: Required, non-empty string
- `password`: Same rules as registration password

### Email Verification
- `token`: Required, non-empty string

### Resend Verification / Forgot Password
- `email`: Required, valid email format

## Access Token Claims (JWT)

| Claim | Type | Description |
|-------|------|-------------|
| sub | string (UUID) | User ID |
| iat | number | Issued at (Unix timestamp) |
| exp | number | Expiration (iat + ~15 min) |
| jti | string (UUID) | Unique token identifier |
| iss | string | Issuer (configurable, default: `{{PROJECT_NAME}}`) |
| aud | string | Audience (configurable, default: `{{PROJECT_NAME}}-api`) |
