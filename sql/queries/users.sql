-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
    VALUES (uuidv7 (), now(), now(), $1, $2)
RETURNING
    id, created_at, updated_at, email, is_chirpy_red;

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

-- name: GetUserById :one
SELECT
    *
FROM
    users
WHERE
    id = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE
    users
SET
    updated_at = now(),
    email = $2,
    hashed_password = $3
WHERE
    id = $1
RETURNING
    id,
    created_at,
    updated_at,
    email,
    is_chirpy_red;

-- name: SetUserChirpyRed :one
UPDATE
    users
SET
    is_chirpy_red = TRUE
WHERE
    id = $1
RETURNING
    id,
    created_at,
    updated_at,
    email,
    is_chirpy_red;
