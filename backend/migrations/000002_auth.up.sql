-- 000002_auth.up.sql — credentials + refresh tokens (GEC-12 auth persistence).
-- Credentials live apart from the users profile so password hashes and TOTP
-- secrets sit in their own table. Refresh tokens persist only a SHA-256 *hash*
-- of the raw token (never the token itself) and are single-use (rotated on
-- refresh). The principal snapshot (name/role/facility_id) is denormalised onto
-- the row so a refresh need not re-read the user.
CREATE TABLE credentials (
    user_id       text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    mfa_secret    text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    token_hash  text PRIMARY KEY,
    user_id     text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        text NOT NULL DEFAULT '',
    role        text NOT NULL,
    facility_id text NOT NULL DEFAULT '',
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
