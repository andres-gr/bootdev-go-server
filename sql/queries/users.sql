-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
    VALUES (uuidv7(), now(), now(), $1)
RETURNING
    *;

-- name: ResetUsers :exec
-- pgls-ignore lint/safety/banDeleteWithoutWhere: only allowed on dev env
DELETE FROM users;
