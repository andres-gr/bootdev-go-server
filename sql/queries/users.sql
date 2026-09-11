-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
    VALUES (uuidv7 (), now(), now(), $1, $2)
RETURNING
    id, created_at, updated_at, email;

-- name: ResetUsers :exec
-- pgls-ignore lint/safety/banDeleteWithoutWhere: only allowed on dev env
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1
LIMIT 1;
