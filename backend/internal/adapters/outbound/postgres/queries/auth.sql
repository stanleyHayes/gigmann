-- name: FindAccountByEmail :one
SELECT u.id, u.name, u.role, u.facility_id, u.preferences,
       c.email, c.password_hash, c.mfa_secret
FROM credentials c
JOIN users u ON u.id = c.user_id
WHERE c.email = $1;

-- name: FindAccountByID :one
SELECT u.id, u.name, u.role, u.facility_id, u.preferences,
       c.email, c.password_hash, c.mfa_secret
FROM users u
JOIN credentials c ON c.user_id = u.id
WHERE u.id = $1;

-- name: UpsertUser :exec
INSERT INTO users (id, name, role, facility_id, preferences)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    role = EXCLUDED.role,
    facility_id = EXCLUDED.facility_id,
    preferences = EXCLUDED.preferences;

-- name: UpsertCredentials :exec
INSERT INTO credentials (user_id, email, password_hash, mfa_secret)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    mfa_secret = EXCLUDED.mfa_secret,
    updated_at = now();

-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (token_hash, user_id, name, role, facility_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ConsumeRefreshToken :one
DELETE FROM refresh_tokens
WHERE token_hash = $1
RETURNING user_id, name, role, facility_id, expires_at;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token_hash = $1;
