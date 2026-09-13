-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
    VALUES ($1, now(), now(), $2, $3, NULL)
RETURNING
    token;

-- name: GetRefreshToken :one
SELECT
    *
FROM
    refresh_tokens
WHERE
    token = $1
LIMIT 1;

-- name: GetUserFromRefreshToken :one
SELECT
    token,
    expires_at,
    revoked_at,
    users.id AS user_id
FROM
    refresh_tokens
    INNER JOIN users ON users.id = refresh_tokens.user_id
WHERE
    token = $1
LIMIT 1;

-- name: RevokeRefreshToken :one
UPDATE
    refresh_tokens
SET
    updated_at = now(),
    revoked_at = now()
WHERE
    token = $1
RETURNING
    *;
