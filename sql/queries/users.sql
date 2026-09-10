-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
    VALUES (uuidv7(), now(), now(), $1)
RETURNING
    *;
