CREATE TABLE verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX verification_tokens_token_hash_idx ON verification_tokens(token_hash);
CREATE INDEX verification_tokens_user_id_idx ON verification_tokens(user_id);
CREATE INDEX verification_tokens_expires_at_idx ON verification_tokens(expires_at) WHERE used_at IS NULL;
